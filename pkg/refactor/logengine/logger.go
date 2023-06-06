package logengine

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BetterLogger interface {
	Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
	Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
	Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
}

type SugarBetterLogger struct {
	Sugar        *zap.SugaredLogger
	AddTraceFrom func(ctx context.Context, toTags map[string]string) map[string]string
}

func (s SugarBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID.String()
	s.log(Debug, tags, msg)
}

func (s SugarBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID.String()
	s.log(Info, tags, msg)
}

func (s SugarBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	msg = fmt.Sprintf(msg, a...)
	tags = s.AddTraceFrom(ctx, tags)
	tags["sender"] = recipientID.String()
	s.log(Error, tags, msg)
}

func (s SugarBetterLogger) log(level LogLevel, tags map[string]string, msg string) {
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

func (loggers ChainedBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Debugf(ctx, recipientID, tags, msg, a...)
	}
}

func (loggers ChainedBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Infof(ctx, recipientID, tags, msg, a...)
	}
}

func (loggers ChainedBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	for i := range loggers {
		loggers[i].Errorf(ctx, recipientID, tags, msg, a...)
	}
}

type CachedSQLLogStore struct {
	logQueue chan *logMessage
	storeAdd func(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error
	callback func(objectID uuid.UUID, objectType string)
	logError func(template string, args ...interface{})
}

type logMessage struct {
	recipientID   uuid.UUID
	recipientType string
	time          time.Time
	tags          map[string]string
	msg           string
	level         LogLevel
}

func (cls *CachedSQLLogStore) logWorker() {
	for {
		l, more := <-cls.logQueue
		if !more {
			return
		}
		if v, ok := l.tags["callpath"]; ok {
			if l.tags["callpath"] == "/" {
				l.tags["root-instance-id"] = l.tags["instance-id"]
			}
			l.tags["callpath"] = internallogger.AppendInstanceID(v, l.tags["instance-id"])
			res, err := internallogger.GetRootinstanceID(v)
			if err != nil {
				l.tags["root-instance-id"] = l.tags["instance-id"]
			} else {
				l.tags["root-instance-id"] = res
			}
		}
		attributes := make(map[string]string)
		attributes["recipientType"] = "sender_type"
		attributes["root-instance-id"] = "root_instance_id"
		attributes["callpath"] = "log_instance_call_path"
		for k, v := range attributes {
			if e, ok := l.tags[k]; ok {
				l.tags[v] = e
			}
		}
		convertedTags := make(map[string]interface{})
		for k, v := range l.tags {
			convertedTags[k] = v
		}
		err := cls.storeAdd(context.Background(), l.time, l.level, l.msg, convertedTags)
		if err != nil {
			cls.logError("cachedSQLLogStore error storing logs, %v", err)
		}
		cls.callback(l.recipientID, l.recipientType)
	}
}

func NewCachedLogger(
	queueSize int,
	storeAdd func(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error,
	pub func(objectID uuid.UUID, objectType string),
	logError func(template string, args ...interface{}),
) (BetterLogger, func(), func()) {
	cls := CachedSQLLogStore{storeAdd: storeAdd, callback: pub, logError: logError, logQueue: make(chan *logMessage, queueSize)}

	return &cls, cls.logWorker, cls.closeLogWorkers
}

func (cls *CachedSQLLogStore) closeLogWorkers() {
	close(cls.logQueue)
}

func (cls *CachedSQLLogStore) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	select {
	case cls.logQueue <- &logMessage{
		time:          time.Now(),
		recipientID:   recipientID,
		tags:          tags,
		msg:           fmt.Sprintf(msg, a...),
		recipientType: fmt.Sprintf("%v", tags["recipientType"]),
		level:         Debug,
	}:
	default:
		cls.logError("!! Log-buffer is/was full.")
	}
}

func (cls *CachedSQLLogStore) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	select {
	case cls.logQueue <- &logMessage{
		time:          time.Now(),
		recipientID:   recipientID,
		tags:          tags,
		msg:           fmt.Sprintf(msg, a...),
		recipientType: fmt.Sprintf("%v", tags["recipientType"]),
		level:         Error,
	}:
	default:
		cls.logError("!! Log-buffer is/was full.")
	}
}

func (cls *CachedSQLLogStore) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
	_ = ctx
	select {
	case cls.logQueue <- &logMessage{
		time:          time.Now(),
		recipientID:   recipientID,
		tags:          tags,
		msg:           fmt.Sprintf(msg, a...),
		recipientType: fmt.Sprintf("%v", tags["recipientType"]),
		level:         Info,
	}:
	default:
		cls.logError("!! Log-buffer is/was full.")
	}
}
