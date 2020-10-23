package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/storage"
	proto "github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	//_ "github.com/mattn/go-sqlite3"
)

func sort(str []string) {
	str[0] = "time"
	str[1] = "level"
	str[2] = "file"
	str[3] = "msg"
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)

	formatter := &logrus.TextFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceQuote:       false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d\t", path.Base(f.File), f.Line)
		},
	}
	logrus.SetFormatter(formatter)
}

func main() {
	configDir := os.Getenv("GIMULATOR_CONFIG_Dir")
	config, err := config.NewConfig(configDir)
	if err != nil {
		panic(err)
	}

	storage, err := storage.NewSqlite(":memory:", config)
	if err != nil {
		panic(err)
	}

	simulator, err := simulator.NewSimulator(storage)
	if err != nil {
		panic(err)
	}

	auther, err := auth.NewAuther(storage)
	if err != nil {
		panic(err)
	}

	server, err := api.NewServer(auther, simulator)
	if err != nil {
		panic(err)
	}

	port := os.Getenv("GIMULATOR_SERVICE_PORT")
	if port == "" {
		port = "23579"
	}

	listen, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	proto.RegisterAPIServer(s, server)
	if err := s.Serve(listen); err != nil {
		panic(err)
	}
}
