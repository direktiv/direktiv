package logengine

// // This logger SHOULD only be used to log relevant monitoring information for
// // direktiv workflow developers and workflow administrators.
// //
// // 1. BetterLogger is an event-style-logger.
// // This means that an log-entry MUST have an "type".
// // The type is passed via tags with the key recipientType. The sender via recipientID.
// //
// // BetterLogger will "smartly" publish a notification on any waiting
// // listeners about new logs, depending on the source and type values.
// //
// // 2. Additional contextual information for the log-entry SHOULD be passed via tags.
// // 3. !!! Tags will be parsed and !modified to enable "smart" logs.
// // Therefor BetterLogger makes assumptions about some entries in the tags.
// // BetterLogger assumes:
// // - tags["callpath"] is present when tags["recipientType"] is instance.
// // - tags["callpath"] has a "special" structure "/uuid/uuid/".
// // - tags["instance-id"] is present and is a uuid when tags["recipientType"] is instance.
// // - tags["loop-index"] is present when the log-message originated inside a loop execution in a workflow.
// // - tags["state-type"] is present when tags["loop-index"] is present.
// //
// // Also BetterLogger will extract from the passed context ctx the trace-id and add it to the tags.
// type BetterLogger interface {
// 	Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
// 	Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
// 	Warnf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
// 	Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{})
// }

// type SugarBetterJSONLogger struct {
// 	Sugar        *zap.SugaredLogger
// 	AddTraceFrom func(ctx context.Context, toTags map[string]string) map[string]string
// }

// func (s SugarBetterJSONLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	tags["source"] = recipientID.String()
// 	s.log(Debug, tags, msg)
// }

// func (s SugarBetterJSONLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	tags["source"] = recipientID.String()
// 	s.log(Info, tags, msg)
// }

// func (s SugarBetterJSONLogger) Warnf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	tags["source"] = recipientID.String()
// 	s.log(Warn, tags, msg)
// }

// func (s SugarBetterJSONLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	tags["source"] = recipientID.String()
// 	s.log(Error, tags, msg)
// }

// func (s SugarBetterJSONLogger) log(level LogLevel, tags map[string]string, msg string) {
// 	logToSuggar := s.Sugar.Debugw
// 	switch level {
// 	case Debug:
// 	case Info:
// 		logToSuggar = s.Sugar.Infow
// 	case Warn:
// 		logToSuggar = s.Sugar.Warnw
// 	case Error:
// 		logToSuggar = s.Sugar.Errorw
// 	}
// 	ar := make([]interface{}, len(tags)+len(tags))
// 	i := 0
// 	for k, v := range tags {
// 		ar[i] = k
// 		ar[i+1] = v
// 		i += 2
// 	}
// 	logToSuggar(msg, ar...)
// }

// type SugarBetterConsoleLogger struct {
// 	Sugar        *zap.SugaredLogger
// 	AddTraceFrom func(ctx context.Context, toTags map[string]string) map[string]string
// 	RetainTags   []string
// }

// func (s SugarBetterConsoleLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	_ = recipientID
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	s.log(Debug, tags, msg)
// }

// func (s SugarBetterConsoleLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	_ = recipientID
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	s.log(Info, tags, msg)
// }

// func (s SugarBetterConsoleLogger) Warnf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	_ = recipientID
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	s.log(Warn, tags, msg)
// }

// func (s SugarBetterConsoleLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	_ = recipientID
// 	msg = fmt.Sprintf(msg, a...)
// 	tags = s.AddTraceFrom(ctx, tags)
// 	s.log(Error, tags, msg)
// }

// func (s SugarBetterConsoleLogger) log(level LogLevel, tags map[string]string, msg string) {
// 	logToSuggar := s.Sugar.Debugw
// 	switch level {
// 	case Debug:
// 	case Info:
// 		logToSuggar = s.Sugar.Infow
// 	case Warn:
// 		logToSuggar = s.Sugar.Warnw
// 	case Error:
// 		logToSuggar = s.Sugar.Errorw
// 	}
// 	ar := make([]interface{}, len(tags)+len(tags))
// 	i := 0
// 	for k, v := range tags {
// 		if s.RetainTags != nil {
// 			for _, consoleTag := range s.RetainTags {
// 				if v != consoleTag {
// 					continue
// 				}
// 			}
// 		}
// 		ar[i] = k
// 		ar[i+1] = v
// 		i += 2
// 	}
// 	logToSuggar(msg, ar...)
// }

// type ChainedBetterLogger []BetterLogger

// func (loggers ChainedBetterLogger) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	tagsCopy := map[string]string{}
// 	for k, v := range tags {
// 		tagsCopy[k] = v
// 	}
// 	for i := range loggers {
// 		loggers[i].Debugf(ctx, recipientID, tagsCopy, msg, a...)
// 	}
// }

// func (loggers ChainedBetterLogger) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	tagsCopy := map[string]string{}
// 	for k, v := range tags {
// 		tagsCopy[k] = v
// 	}
// 	for i := range loggers {
// 		loggers[i].Infof(ctx, recipientID, tagsCopy, msg, a...)
// 	}
// }

// func (loggers ChainedBetterLogger) Warnf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	tagsCopy := map[string]string{}
// 	for k, v := range tags {
// 		tagsCopy[k] = v
// 	}
// 	for i := range loggers {
// 		loggers[i].Warnf(ctx, recipientID, tagsCopy, msg, a...)
// 	}
// }

// func (loggers ChainedBetterLogger) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	tagsCopy := map[string]string{}
// 	for k, v := range tags {
// 		tagsCopy[k] = v
// 	}
// 	for i := range loggers {
// 		loggers[i].Errorf(ctx, recipientID, tagsCopy, msg, a...)
// 	}
// }

// type CachedSQLLogStore struct {
// 	logQueue chan *logMessage
// 	storeAdd func(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error
// 	callback func(objectID uuid.UUID, objectType string)
// 	logError func(template string, args ...interface{})
// }

// type logMessage struct {
// 	recipientID   uuid.UUID
// 	recipientType string
// 	time          time.Time
// 	tags          map[string]string
// 	msg           string
// 	level         LogLevel
// }

// func (cls *CachedSQLLogStore) logWorker() {
// 	for {
// 		l, more := <-cls.logQueue
// 		if !more {
// 			return
// 		}
// 		attributes := make(map[string]string)
// 		attributes["recipientType"] = "type"
// 		attributes["root-instance-id"] = "root_instance_id"
// 		attributes["callpath"] = "log_instance_call_path"
// 		for k, v := range attributes {
// 			if e, ok := l.tags[k]; ok {
// 				l.tags[v] = e
// 			}
// 		}
// 		convertedTags := make(map[string]interface{})
// 		for k, v := range l.tags {
// 			convertedTags[k] = v
// 		}
// 		convertedTags["source"] = l.recipientID
// 		err := cls.storeAdd(context.Background(), l.time, l.level, l.msg, convertedTags)
// 		if err != nil {
// 			cls.logError("cachedSQLLogStore error storing logs, %v", err)
// 		}
// 		cls.callback(l.recipientID, l.recipientType)
// 	}
// }

// func NewCachedLogger(
// 	queueSize int,
// 	storeAdd func(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error,
// 	pub func(objectID uuid.UUID, objectType string),
// 	logError func(template string, args ...interface{}),
// ) (BetterLogger, func(), func()) {
// 	cls := CachedSQLLogStore{storeAdd: storeAdd, callback: pub, logError: logError, logQueue: make(chan *logMessage, queueSize)}

// 	return &cls, cls.logWorker, cls.closeLogWorkers
// }

// func (cls *CachedSQLLogStore) closeLogWorkers() {
// 	close(cls.logQueue)
// }

// func (cls *CachedSQLLogStore) Debugf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	select {
// 	case cls.logQueue <- &logMessage{
// 		time:          time.Now().UTC(),
// 		recipientID:   recipientID,
// 		tags:          tags,
// 		msg:           fmt.Sprintf(msg, a...),
// 		recipientType: tags["recipientType"],
// 		level:         Debug,
// 	}:
// 	default:
// 		cls.logError("!! Log-buffer is/was full.")
// 	}
// }

// func (cls *CachedSQLLogStore) Errorf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	select {
// 	case cls.logQueue <- &logMessage{
// 		time:          time.Now().UTC(),
// 		recipientID:   recipientID,
// 		tags:          tags,
// 		msg:           fmt.Sprintf(msg, a...),
// 		recipientType: tags["recipientType"],
// 		level:         Error,
// 	}:
// 	default:
// 		cls.logError("!! Log-buffer is/was full.")
// 	}
// }

// func (cls *CachedSQLLogStore) Warnf(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	select {
// 	case cls.logQueue <- &logMessage{
// 		time:          time.Now().UTC(),
// 		recipientID:   recipientID,
// 		tags:          tags,
// 		msg:           fmt.Sprintf(msg, a...),
// 		recipientType: tags["recipientType"],
// 		level:         Warn,
// 	}:
// 	default:
// 		cls.logError("!! Log-buffer is/was full.")
// 	}
// }

// func (cls *CachedSQLLogStore) Infof(ctx context.Context, recipientID uuid.UUID, tags map[string]string, msg string, a ...interface{}) {
// 	_ = ctx
// 	appendInstanceInheritanceInfo(tags)
// 	select {
// 	case cls.logQueue <- &logMessage{
// 		time:          time.Now().UTC(),
// 		recipientID:   recipientID,
// 		tags:          tags,
// 		msg:           fmt.Sprintf(msg, a...),
// 		recipientType: tags["recipientType"],
// 		level:         Info,
// 	}:
// 	default:
// 		cls.logError("!! Log-buffer is/was full.")
// 	}
// }

// // constructing the callpath and setting the root-instance-id
// // function assumes the callpath misses the creators id at the end.
// // WHY: currently we expect the callpath to miss the uuid of the instance where the logmsg originated (instance-id in the tags) from.
// // The reason for this was: the id for the instance was set only after instance was inserted into the database by ent.
// // TODO: It would be better to have the uuid of the instance to be already in the callpath.
// // Example for current callpath structure:
// // the log message was created by instance-id: "75d8b87a"
// // the parent of "75d8b87a" was instance-id: "1dd92e"
// // the parent of "1dd92e" was instance-id: "124279"
// // the callpath for this example would be: "/124279/1dd92e/"
// // the final callpath after applying the function should look be: "/124279/1dd92e/75d8b87a"
// // other example the log message was created by instance-id: "75d8b87a"
// // "75d8b87b" has no parent instance, therefor is the root-instance
// // for this case we expect the callpath to be "/"
// // the final callpath after applying the function should look be: "/75d8b87b"
// //
// // to make querying the logs more connivent and efficient we append the missing
// // instance-id before to the callpath tag of the log-entry
// // and add the root-instance-id tag fro the constructed final callpath.
// func appendInstanceInheritanceInfo(tags map[string]string) {
// 	if _, ok := tags["callpath"]; ok {
// 		if tags["callpath"] == "/" {
// 			tags["root-instance-id"] = tags["instance-id"]
// 		}
// 		if strings.Contains(tags["callpath"], tags["instance-id"]) {
// 			return
// 		}
// 		tags["callpath"] = internallogger.AppendInstanceID(tags["callpath"], tags["instance-id"])
// 		res, err := internallogger.GetRootinstanceID(tags["callpath"])
// 		if err != nil {
// 			tags["root-instance-id"] = tags["instance-id"]
// 		} else {
// 			tags["root-instance-id"] = res
// 		}
// 	}
// }
