package dummy

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/vorteil/direktiv/pkg/dlog"
)

type DummyLogger struct {
}

func NewLogger() (*DummyLogger, error) {
	l := new(DummyLogger)
	return l, nil
}

type dLogger struct {
	log15.Logger
}

func (dl *dLogger) Close() error {
	return nil
}

func (l *DummyLogger) NamespaceLogger(namespace string) (dlog.Logger, error) {
	logger := new(dLogger)
	logger.Logger = log15.New("namespace", namespace)
	return logger, nil
}
func (l *DummyLogger) LoggerFunc(namespace, instance string) (dlog.Logger, error) {
	logger := new(dLogger)
	logger.Logger = log15.New("namespace", namespace, "instance", instance)
	return logger, nil
}

// func (l *DummyLogger) QueryNamespaceLogs(ctx context.Context, instance string, limit, offset int) (dlog.QueryReponse, error) {
// 	dlg := dlog.QueryReponse{
// 		Logs: make([]dlog.LogEntry, 0),
// 	}

// 	return dlg, nil
// }

func (l *DummyLogger) QueryLogs(ctx context.Context, instance string, limit, offset int) (dlog.QueryReponse, error) {
	dlg := dlog.QueryReponse{
		Logs: make([]dlog.LogEntry, 0),
	}

	return dlg, nil
}

func (l *DummyLogger) QueryAllLogs(instance string) (dlog.QueryReponse, error) {
	dlg := dlog.QueryReponse{
		Logs: make([]dlog.LogEntry, 0),
	}

	return dlg, nil
}

func (l *DummyLogger) DeleteNamespaceLogs(namespace string) error {
	return nil
}

func (l *DummyLogger) DeleteInstanceLogs(instance string) error {
	return nil
}
