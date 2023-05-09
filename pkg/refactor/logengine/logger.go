package logengine

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LogNotify interface {
	NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType)
}

type Loggerw struct {
	Sugar *zap.SugaredLogger
	Store LogStore
	Pub   LogNotify
}

func (logger *Loggerw) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) error {
	msg = fmt.Sprintf(msg, a...)

	if len(tags) == 0 {
		logger.Sugar.Infow(msg)
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
			logger.Sugar.Infow(msg, ar...)
		case "debug":
			logger.Sugar.Debugw(msg, ar...)
		case "error":
			logger.Sugar.Errorw(msg, ar...)
		case "panic":
			logger.Sugar.Panicw(msg, ar...)
		default:
			logger.Sugar.Debugw(msg, ar...) // this should never happen
		}
	}

	tags["level"] = level
	err := logger.Store.Append(context.Background(), time.Now(), msg, tags)
	if err != nil {
		return err
	}
	id, _ := uuid.Parse(fmt.Sprintf("%s", tags["sender"]))
	logger.Pub.NotifyLogs(id, recipient.RecipientType(fmt.Sprintf("%s", tags["senderType"])))

	return nil
}
