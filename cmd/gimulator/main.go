package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/manager"
	"github.com/Gimulator/Gimulator/mq"
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

	configDir := os.Getenv("GIMULATOR_CONFIG_DIR")
	log.WithField("config-dir", configDir).Info("starting to setup configs")
	config, err := config.NewConfig(configDir)
	if err != nil {
		log.WithField("config-dir", configDir).WithError(err).Fatal("could not setup configs")
		panic(err)
	}

	log.Info("starting to setup sqlite")
	sqlite, err := storage.NewSqlite(":memory:", config)
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

	log.Info("starting to setup auther")
	manager, err := manager.NewManager(sqlite, sqlite)
	if err != nil {
		log.WithError(err).Fatal("could not setup auther")
		panic(err)
	}

	log.Info("starting to setup RabbitMQ")
	rabbitURL := os.Getenv("GIMULATOR_RABBIT_URL")
	if rabbitURL == "" {
		panic("set th 'GIMULATOR_RABBIT_URL' environment variable for sending result to RabbitMQ")
	}
	rabbitQueue := os.Getenv("GIMULATOR_RABBIT_QUEUE")
	if rabbitQueue == "" {
		panic("set th 'GIMULATOR_RABBIT_QUEUE' environment variable for sending result to RabbitMQ")
	}

	rabbit, err := mq.NewRabbit(rabbitURL, rabbitQueue)
	if err != nil {
		panic(err)
	}

	log.Info("starting to setup server")
	server, err := api.NewServer(manager, simulator, rabbit)
	if err != nil {
		log.WithError(err).Fatal("could not setup server")
		panic(err)
	}

	port := os.Getenv("GIMULATOR_SERVICE_PORT")
	if port == "" {
		port = "23579"
	}
	host := "0.0.0.0:" + port

	log.WithField("host", host).Info("starting to setup listener")
	listener, err := net.Listen("tcp", host)
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
