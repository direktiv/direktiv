package logengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BetterLogger interface {
	// Logs a log message with contextual information, which are passed via tags.
	// Remember to pass the trace-id for the logentry via the tags with the key trace.
	Log(tags map[string]interface{}, level string, msg string, a ...interface{})
}

type SugarBetterLogger struct {
	Sugar *zap.SugaredLogger
}

func (s SugarBetterLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	msg = fmt.Sprintf(msg, a...)
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

type ChainedBetterLogger []BetterLogger

func (loggers ChainedBetterLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Log(tags, level, msg, a...)
	}
}

// DataStoreBetterLogger records log information into the datastore so that UI frontend page can show log data about
// different objects.
type DataStoreBetterLogger struct {
	Store    LogStore
	LogError func(template string, args ...interface{})
}

func (s DataStoreBetterLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	err := s.Store.Append(context.Background(), time.Now(), level, fmt.Sprintf(msg, a...), tags)
	if err != nil {
		s.LogError("writing action log to the database", "error", err)
	}
}

// NotifierBetterLogger is a pseudo action logger that doesn't log any information, instead it calls a callback
// that reporting the object that was logged.
type NotifierBetterLogger struct {
	Callback func(objectID uuid.UUID, objectType string)
	LogError func(template string, args ...interface{})
}

func (n NotifierBetterLogger) Log(tags map[string]interface{}, level string, msg string, a ...interface{}) {
	tags["level"] = level
	_ = msg
	_ = a
	senderID, ok := tags["sender"]
	if !ok {
		n.LogError("cannot find sender id in action log tags", "tags", tags)

		return
	}
	senderType, ok := tags["senderType"]
	if !ok {
		n.LogError("cannot find sender type in action log tags", "tags", tags)

		return
	}
	id, err := uuid.Parse(fmt.Sprintf("%s", senderID))
	if err != nil {
		n.LogError("cannot parse sender id in action log tags", "tags", tags, "error", err)

		return
	}

	n.Callback(id, fmt.Sprintf("%s", senderType))
}
