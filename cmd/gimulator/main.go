package main

import (
	"fmt"
	"path"
	"runtime"

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
}
