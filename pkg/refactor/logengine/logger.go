package logengine

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type LogNotify interface {
	NotifyLogs(recipientID string, recipientType string)
}

type loggerw struct {
	sugar *zap.SugaredLogger
	store LogStore
	pub   LogNotify
}

type Log func(tags map[string]interface{}, level string, msg string, a ...interface{}) error

func Logger(ls LogStore, sug zap.SugaredLogger, pub LogNotify) Log {
	return loggerw{sugar: &sug, store: ls, pub: pub}.log
}

func (logger loggerw) log(tags map[string]interface{}, level string, msg string, a ...interface{}) error {
	msg = fmt.Sprintf(msg, a...)

	if len(tags) == 0 {
		logger.sugar.Infow(msg)
	} else {
		ar := make([]interface{}, len(tags)+len(tags))
		i := 0
		for k, v := range tags {
			ar[i] = k
			ar[i+1] = v
			i += 2
		}
		switch level {
		case "info":
			logger.sugar.Infow(msg, ar...)
		case "debug":
			logger.sugar.Debugw(msg, ar...)
		case "error":
			logger.sugar.Errorw(msg, ar...)
		case "panic":
			logger.sugar.Panicw(msg, ar...)
		default:
			logger.sugar.Debugw(msg, ar...) // this should never happen
		}
	}

	tags["level"] = level
	err := logger.store.Append(context.Background(), time.Now(), msg, tags)
	if err != nil {
		return err
	}
	logger.pub.NotifyLogs(fmt.Sprintf("%s", tags["sender"]), fmt.Sprintf("%s", tags["senderType"]))

	return nil
}
