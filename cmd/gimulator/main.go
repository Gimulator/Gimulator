package main

import (
	"fmt"
	"net"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/cmd"
	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/manager"
	"github.com/Gimulator/Gimulator/epilogues"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/storage"
	proto "github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func sort(str []string) {
	str[0] = "time"
	str[1] = "level"
	str[2] = "file"
	str[3] = "msg"
}

func init() {
	cmd.ParseFlags()

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
	log := logrus.WithField("component", "main")

	log.WithField("config-dir", cmd.ConfigDir).Info("starting to setup configs")
	config, err := config.NewConfig(cmd.ConfigDir)
	if err != nil {
		log.WithField("config-dir", cmd.ConfigDir).WithError(err).Fatal("could not setup configs")
		panic(err)
	}

	log.Info("starting to setup sqlite")
	// Using In-Memory Database with Shared Cache (Instad of private cache)
	sqlite, err := storage.NewSqlite("file::memory:?cache=shared", config)
	if err != nil {
		log.WithError(err).Fatal("could not setup sqlite")
		panic(err)
	}

	log.Info("starting to setup simulator")
	simulator, err := simulator.NewSimulator(sqlite)
	if err != nil {
		log.WithError(err).Fatal("could not setup simulator")
		panic(err)
	}

	log.Info("starting to setup manager")
	manager, err := manager.NewManager(sqlite, sqlite)
	if err != nil {
		log.WithError(err).Fatal("could not setup manager")
		panic(err)
	}

	var epilogue epilogues.Epilogue

	switch cmd.EpilogueType {
	case "console":
		epilogue, err = epilogues.NewConsole()
		if err != nil {
			log.WithError(err).Fatal("could not setup console")
			panic(err)
		}
	case "rabbitmq":
		log.Info("starting to setup rabbit")  // FIXME optionalize
		epilogue, err = epilogues.NewRabbitMQ(cmd.RabbitHost, cmd.RabbitUsername, cmd.RabbitPassword, cmd.RabbitQueue)
		if err != nil {
			log.WithError(err).Fatal("could not setup rabbit")
			panic(err)
		}
	}

	log.Info("starting to setup server")
	server, err := api.NewServer(manager, simulator, epilogue)
	if err != nil {
		log.WithError(err).Fatal("could not setup server")
		panic(err)
	}

	log.WithField("host", cmd.Host).Info("starting to setup listener")
	listener, err := net.Listen("tcp", cmd.Host)
	if err != nil {
		log.WithError(err).Fatal("could not setup listener")
		panic(err)
	}

	log.Info("starting to serve")
	s := grpc.NewServer()
	proto.RegisterMessageAPIServer(s, server)
	proto.RegisterOperatorAPIServer(s, server)
	proto.RegisterDirectorAPIServer(s, server)
	proto.RegisterUserAPIServer(s, server)
	if err := s.Serve(listener); err != nil {
		log.WithError(err).Fatal("could not serve")
		panic(err)
	}
}
