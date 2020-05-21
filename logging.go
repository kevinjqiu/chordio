package chordio

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

func SetLogLevel(loglevel string) {
	fmt.Printf("Setting loglevel to: %s\n", loglevel)
	switch loglevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.Fatal("unrecognized loglevel: ", loglevel)
	}
}
