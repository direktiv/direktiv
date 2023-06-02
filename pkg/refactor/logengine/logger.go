package logengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BetterLogger interface {
	Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{})
	Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{})
	Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{})
}

type SugarBetterLogger struct {
	Sugar        *zap.SugaredLogger
	AddTraceFrom func(ctx context.Context, toTags map[string]interface{}) map[string]interface{}
}

func (s SugarBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID
	s.log(Debug, tags, msg)
}

func (s SugarBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID
	s.log(Info, tags, msg)
}

func (s SugarBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID
	s.log(Error, tags, msg)
}

func (s SugarBetterLogger) log(level LogLevel, tags map[string]interface{}, msg string) {
	logToSuggar := s.Sugar.Debugw
	switch level {
	case Debug:
	case Info:
		logToSuggar = s.Sugar.Infow
	case Error:
		logToSuggar = s.Sugar.Errorw
	}
	ar := make([]interface{}, len(tags)+len(tags))
	i := 0
	for k, v := range tags {
		ar[i] = k
		ar[i+1] = v
		i += 2
	}
	logToSuggar(msg, ar...)
}

type ChainedBetterLogger []BetterLogger

func (loggers ChainedBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Debugf(ctx, recipientID, tags, msg, a...)
	}
}

func (loggers ChainedBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Infof(ctx, recipientID, tags, msg, a...)
	}
}

func (loggers ChainedBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Errorf(ctx, recipientID, tags, msg, a...)
	}
}

// DataStoreBetterLogger records log information into the datastore so that UI frontend page can show log data about
// different objects.
type DataStoreBetterLogger struct {
	Store    LogStore
	LogError func(template string, args ...interface{})
}

func (s DataStoreBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	_ = recipientID
	err := s.Store.Append(context.Background(), time.Now(), Debug, fmt.Sprintf(msg, a...), constructKey(tags), tags)
	if err != nil {
		s.LogError("writing better-logs to the database", "error", err)
	}
}

func (s DataStoreBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	_ = recipientID
	err := s.Store.Append(context.Background(), time.Now(), Info, fmt.Sprintf(msg, a...), constructKey(tags), tags)
	if err != nil {
		s.LogError("writing better-logs to the database", "error", err)
	}
}

func (s DataStoreBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = ctx
	_ = recipientID
	err := s.Store.Append(context.Background(), time.Now(), Error, fmt.Sprintf(msg, a...), constructKey(tags), tags)
	if err != nil {
		s.LogError("writing better-logs to the database", "error", err)
	}
}

func constructKey(tags map[string]interface{}) string {
	key := "server"
	if v, ok := tags["namespace"]; ok {
		key = fmt.Sprintf("%v", v)
	}
	if v, ok := tags["workflow"]; ok {
		key += fmt.Sprintf("/%v", v)
	}
	if v, ok := tags["instance"]; ok {
		key += fmt.Sprintf("%v", v)
	}
	if v, ok := tags["mirror"]; ok {
		key += fmt.Sprintf("%v", v)
	}

	return key
}

// NotifierBetterLogger is a pseudo action logger that doesn't log any information, instead it calls a callback
// that reporting the object that was logged.
type NotifierBetterLogger struct {
	Callback func(objectID uuid.UUID, objectType string)
	LogError func(template string, args ...interface{})
}

func (n NotifierBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = msg
	_ = a
	_ = ctx
	n.log(recipientID, tags)
}

func (n NotifierBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = msg
	_ = a
	_ = ctx
	n.log(recipientID, tags)
}

func (n NotifierBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]interface{}, msg string, a ...interface{}) {
	_ = msg
	_ = a
	_ = ctx
	n.log(recipientID, tags)
}

func (n NotifierBetterLogger) log(recipientID uuid.UUID, tags map[string]interface{}) {
	senderType, ok := tags["sender_type"]
	if !ok {
		n.LogError("cannot find sender type in better-logs tags", "tags", tags)

		return
	}

	n.Callback(recipientID, fmt.Sprintf("%s", senderType))
}
