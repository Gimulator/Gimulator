package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/Gimulator/Gimulator/config"
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
	rolesPath := os.Getenv("GIMULATOR_ROLES_PATH")
	roles, err := config.NewRoles(rolesPath)
	if err != nil {
		panic(err)
	}

	credsPath := os.Getenv("GIMULATOR_CREDENTIALS_PATH")
	creds, err := config.NewCredentials(credsPath)
	if err != nil {
		panic(err)
	}
}
