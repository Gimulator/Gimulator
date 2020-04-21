package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/auth"
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

	file, err := os.Create("log.txt")
	if err != nil {
		logrus.SetOutput(os.Stderr)
	}
	logrus.SetOutput(file)

	logrus.SetReportCaller(true)
	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableSorting:  false,
		SortingFunc:     sort,

		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	}
	logrus.SetFormatter(formatter)
}

func main() {
	ip := flag.String("ip", "localhost:3030", "ip is for listening and serving")
	configFile := flag.String("config-file", "", "this is a config file for auth")
	flag.Parse()

	if *configFile == "" {
		flag.PrintDefaults()
		return
	}

	storage := storage.NewMemory()
	simulator := simulator.NewSimulator(storage)
	auth, err := auth.NewAuth(*configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	api := api.NewManager(simulator, auth)
	api.ListenAndServe(*ip)
}
