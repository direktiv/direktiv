package mirror

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Callbacks interface {
	// ConfigureWorkflowFunc is a hookup function the gets called for every new or updated workflow file.
	ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error

	ProcessLogger() ProcessLogger

	SysLogCrit(msg string)

	Store() Store
	FileStore() filestore.FileStore
	VarStore() core.RuntimeVariablesStore
	FileAnnotationsStore() core.FileAnnotationsStore
}

type ProcessLogger interface {
	Error(pid uuid.UUID, msg string, keysAndValues ...interface{})
	Warn(pid uuid.UUID, msg string, keysAndValues ...interface{})
	Info(pid uuid.UUID, msg string, keysAndValues ...interface{})
	Debug(pid uuid.UUID, msg string, keysAndValues ...interface{})
}

type FormatLogger interface {
	Errorf(msg string, a ...interface{})
	Warnf(msg string, a ...interface{})
	Infof(msg string, a ...interface{})
	Debugf(msg string, a ...interface{})
}

type pidFormatLogger struct {
	logger ProcessLogger
	pid    uuid.UUID
	attr   []interface{}
}

func newPIDFormatLogger(logger ProcessLogger, pid uuid.UUID, attr ...interface{}) *pidFormatLogger {
	return &pidFormatLogger{
		logger: logger,
		pid:    pid,
		attr:   attr,
	}
}

func (l *pidFormatLogger) Errorf(msg string, a ...interface{}) {
	l.logger.Error(l.pid, fmt.Sprintf(msg, a...), l.attr...)
}

func (l *pidFormatLogger) Warnf(msg string, a ...interface{}) {
	l.logger.Warn(l.pid, fmt.Sprintf(msg, a...), l.attr...)
}

func (l *pidFormatLogger) Infof(msg string, a ...interface{}) {
	l.logger.Info(l.pid, fmt.Sprintf(msg, a...), l.attr...)
}

func (l *pidFormatLogger) Debugf(msg string, a ...interface{}) {
	l.logger.Debug(l.pid, fmt.Sprintf(msg, a...), l.attr...)
}
