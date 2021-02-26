package log

import (
	"github.com/sirupsen/logrus"
	"github.com/vorteil/vorteil/pkg/elog"
)

var log elog.View

func init() {
	logger := &elog.CLI{}
	logrus.SetFormatter(logger)
	logrus.SetLevel(logrus.TraceLevel)
	log = logger
}

// GetLogger returns the global 'log' variable
func GetLogger() elog.View {
	return log
}
