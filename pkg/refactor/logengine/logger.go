package logengine

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ActionLogger interface {
	Log(tags map[string]interface{}, level string, msg string, a ...interface{})
}

type SugarActionLogger struct {
	Sugar *zap.SugaredLogger
}

func (s SugarActionLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	if len(tags) == 0 {
		s.Sugar.Infow(msg)
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
			s.Sugar.Infow(msg, ar...)
		case "debug":
			s.Sugar.Debugw(msg, ar...)
		case "error":
			s.Sugar.Errorw(msg, ar...)
		case "panic":
			s.Sugar.Panicw(msg, ar...)
		default:
			s.Sugar.Debugw(msg, ar...) // this should never happen
		}
	}
}

type ChainedActionLogger []ActionLogger

func (loggers ChainedActionLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Log(tags, level, msg, a...)
	}
}

// DataStoreActionLogger records log information into the datastore so that UI frontend page can show log data about
// different objects.
type DataStoreActionLogger struct {
	Store       LogStore
	ErrorLogger *zap.SugaredLogger
}

func (s DataStoreActionLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	err := s.Store.Append(context.Background(), level, msg, tags)
	if err != nil {
		s.ErrorLogger.Error("writing action log to the database", "error", err)
	}
}

// NotifierActionLogger is a pseudo action logger that doesn't log any information, instead it calls a callback
// that reporting the object that was logged.
type NotifierActionLogger struct {
	Callback    func(objectID uuid.UUID, objectType string)
	ErrorLogger *zap.SugaredLogger
}

func (n NotifierActionLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	tags["level"] = level
	senderID, ok := tags["sender"]
	if !ok {
		n.ErrorLogger.Error("cannot find sender id in action log tags", "tags", tags)
		return
	}
	senderType, ok := tags["senderType"]
	if !ok {
		n.ErrorLogger.Error("cannot find sender type in action log tags", "tags", tags)
		return
	}
	id, err := uuid.Parse(fmt.Sprintf("%s", senderID))
	if err != nil {
		n.ErrorLogger.Error("cannot parse sender id in action log tags", "tags", tags, "error", err)
		return
	}

	n.Callback(id, fmt.Sprintf("%s", senderType))
}
