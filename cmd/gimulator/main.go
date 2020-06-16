package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/storage"
	"github.com/sirupsen/logrus"
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
	host := os.Getenv("GIMULATOR_HOST")
	if host == "" {
		panic("set 'GIMULATOR_HOST' to listen and serve")
	}

	configPath := os.Getenv("GIMULATOR_ROLES_FILE_PATH")
	if configPath == "" {
		panic("set 'GIMULATOR_ROLES_FILE_PATH' to setup auth")
	}

	storage := storage.NewMemory()
	simulator := simulator.NewSimulator(storage)

	config, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	auth, err := auth.NewAuth(config)
	if err != nil {
		panic(err)
	}

	api := api.NewManager(simulator, auth)
	if err := api.ListenAndServe(host); err != nil {
		panic(err)
	}
}
