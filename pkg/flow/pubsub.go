package flow

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	"go.uber.org/zap"
)

const (
	pubsubNotifyFunction               = "notify"
	pubsubDisconnectFunction           = "disconnect"
	pubsubDeleteTimerFunction          = "deleteTimer"
	pubsubDeleteInstanceTimersFunction = "deleteInstanceTimers"
	pubsubCancelWorkflowFunction       = "cancelWorkflow"
	pubsubConfigureRouterFunction      = "configureRouter"
)

type pubsub struct {
	id       uuid.UUID
	notifier notifier
	log      *zap.SugaredLogger

	handlers map[string]func(*PubsubUpdate)

	closed   bool
	closer   chan bool
	hostname string
	queue    chan *PubsubUpdate
	mtx      sync.RWMutex
	channels map[string]map[*subscription]bool
}

func (pubsub *pubsub) Close() error {

	if !pubsub.closed {
		close(pubsub.closer)
	}

	pubsub.closed = true

	return nil

}

type PubsubUpdate struct {
	Handler  string
	Sender   string
	Key      string
	Hostname string
}

const flowSync = "flowsync"

type notifier interface {
	notifyCluster(string) error
	notifyHostname(string, string) error
}

func initPubSub(log *zap.SugaredLogger, notifier notifier, database string) (*pubsub, error) {

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Errorf("pubsub error: %v\n", err)
			os.Exit(1)
		}
	}

	listener := pq.NewListener(database, 10*time.Second,
		time.Minute, reportProblem)
	err := listener.Listen(flowSync)
	if err != nil {
		return nil, err
	}

	pubsub := new(pubsub)
	pubsub.id = uuid.New()
	pubsub.log = log

	pubsub.hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	pubsub.notifier = notifier
	pubsub.closer = make(chan bool)
	pubsub.queue = make(chan *PubsubUpdate, 1024)
	pubsub.channels = make(map[string]map[*subscription]bool)

	go pubsub.dispatcher()

	pubsub.handlers = make(map[string]func(*PubsubUpdate))

	// pool := notifier.redisPool()
	//
	// conn := pool.Get()
	//
	// _, err = conn.Do("PING")
	// if err != nil {
	// 	return nil, fmt.Errorf("can't connect to redis, got error:\n%v", err)
	// }
	//
	// go func() {
	//
	// 	rc := pool.Get()
	//
	// 	psc := redis.PubSubConn{Conn: rc}
	// 	if err := psc.PSubscribe(flowSync); err != nil {
	// 		log.Error(err.Error())
	// 	}
	//
	// 	for {
	// 		switch v := psc.Receive().(type) {
	// 		default:
	// 			data, _ := json.Marshal(v)
	// 			log.Debug(string(data))
	// 		case redis.Message:
	// 			req := new(PubsubUpdate)
	// 			err = json.Unmarshal(v.Data, req)
	// 			if err != nil {
	// 				log.Error(fmt.Sprintf("Unexpected notification on database listener: %v", err))
	// 			} else {
	// 				handler, exists := pubsub.handlers[req.Handler]
	// 				if !exists {
	// 					log.Errorf("unexpected notification type on database listener: %v\n", err)
	// 					continue
	// 				}
	// 				handler(req)
	// 			}
	// 		}
	// 	}
	//
	// }()

	go func(l *pq.Listener) {

		defer pubsub.Close()
		defer l.UnlistenAll()

		for {

			var more bool
			var notification *pq.Notification

			select {
			case <-pubsub.closer:
			case notification, more = <-l.Notify:
				if !more {
					log.Errorf("database listener closed\n")
					return
				}
			}

			if notification == nil {
				continue
			}

			req := new(PubsubUpdate)
			err = json.Unmarshal([]byte(notification.Extra), req)
			if err != nil {
				log.Errorf("unexpected notification on database listener: %v\n", err)
				continue
			}

			handler, exists := pubsub.handlers[req.Handler]
			if !exists {
				log.Errorf("unexpected notification type on database listener: %v\n", err)
				continue
			}

			handler(req)

		}

	}(listener)

	return pubsub, nil

}

func (pubsub *pubsub) registerFunction(name string, fn func(*PubsubUpdate)) {

	if _, ok := pubsub.handlers[name]; ok {
		panic(fmt.Errorf("function already exists"))
	}

	pubsub.handlers[name] = fn

}

func (pubsub *pubsub) dispatcher() {

	for {

		req, more := <-pubsub.queue
		if !more {
			return
		}

		b, _ := json.Marshal(req)

		var err error

		if req.Hostname == "" {
			err = pubsub.notifier.notifyCluster(string(b))
		} else {
			err = pubsub.notifier.notifyHostname(req.Hostname, string(b))
		}

		if err != nil {
			pubsub.log.Errorf("pubsub error: %v\n", err)
			os.Exit(1)
		}

	}

}

func (pubsub *pubsub) Notify(req *PubsubUpdate) {

	pubsub.mtx.RLock()
	defer pubsub.mtx.RUnlock()

	channel, exists := pubsub.channels[req.Key]
	if !exists {
		return
	}

	for sub, listening := range channel {
		if listening {
			select {
			case sub.ch <- true:
			default:
			}
		}
	}

}

func (pubsub *pubsub) Disconnect(req *PubsubUpdate) {

	pubsub.mtx.RLock()
	defer pubsub.mtx.RUnlock()

	channel, exists := pubsub.channels[req.Key]
	if !exists {
		return
	}

	for sub := range channel {
		go func(sub *subscription) {
			_ = sub.Close()
		}(sub)
	}

}

type subscription struct {
	keys   []string
	ch     chan bool
	closed bool
	pubsub *pubsub
	last   time.Time
}

func (s *subscription) Wait() bool {

	t := time.Now()
	dt := t.Sub(s.last)
	if dt < (time.Millisecond * 150) {
		time.Sleep(dt)
	}

	defer func() {
		s.last = time.Now()
	}()

	select {
	case _, more := <-s.ch:
		return more
	case <-time.After(time.Minute):
		return true
	}

}

func (s *subscription) Close() error {

	if !s.closed {

		defer func() {
			r := recover()
			if r != nil {
				fmt.Println("PANIC", r)
			}
		}()

		s.closed = true

		s.pubsub.mtx.Lock()

		for _, key := range s.keys {
			channel := s.pubsub.channels[key]
			delete(channel, s)
		}

		s.pubsub.mtx.Unlock()

		close(s.ch)

	}

	return nil

}

func (pubsub *pubsub) publish(req *PubsubUpdate) {

	req.Sender = pubsub.id.String()

	select {
	case pubsub.queue <- req:
	default:
	}

}

func (pubsub *pubsub) Subscribe(id ...string) *subscription {

	s := new(subscription)

	s.ch = make(chan bool, 1)
	s.pubsub = pubsub

	pubsub.mtx.Lock()

	key := ""
	s.keys = append(s.keys, key)
	channel, exists := pubsub.channels[key]
	if !exists {
		channel = make(map[*subscription]bool)
		pubsub.channels[key] = channel
	}
	channel[s] = len(id) == 0

	for idx, x := range id {
		key := x
		s.keys = append(s.keys, key)
		channel, exists := pubsub.channels[key]
		if !exists {
			channel = make(map[*subscription]bool)
			pubsub.channels[key] = channel
		}
		channel[s] = idx == len(id)-1
	}

	pubsub.mtx.Unlock()

	return s

}

func pubsubNotify(key string) *PubsubUpdate {
	return &PubsubUpdate{
		Handler: pubsubNotifyFunction,
		Key:     key,
	}
}

func pubsubDisconnect(key string) *PubsubUpdate {
	return &PubsubUpdate{
		Handler: pubsubDisconnectFunction,
		Key:     key,
	}
}

func (pubsub *pubsub) NotifyServerLogs() {

	pubsub.publish(pubsubNotify(""))

}

func (pubsub *pubsub) SubscribeServerLogs() *subscription {

	return pubsub.Subscribe()

}

func (pubsub *pubsub) SubscribeNamespaces() *subscription {

	return pubsub.Subscribe("namespaces")

}

func (pubsub *pubsub) NotifyNamespaces() {

	pubsub.publish(pubsubNotify("namespaces"))

}

func (pubsub *pubsub) SubscribeNamespace(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String())

}

func (pubsub *pubsub) NotifyNamespace(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(ns.ID.String()))

}

func (pubsub *pubsub) CloseNamespace(ns *ent.Namespace) {

	pubsub.publish(pubsubDisconnect(ns.ID.String()))

}

func (pubsub *pubsub) namespaceLogs(ns *ent.Namespace) string {

	return fmt.Sprintf("nslog:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceLogs(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceLogs(ns))

}

func (pubsub *pubsub) NotifyNamespaceLogs(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceLogs(ns)))

}

func (pubsub *pubsub) walkInodeKeys(ino *ent.Inode) []string {

	array := make([]string, 0)

	x := ino
	array = append(array, x.ID.String())

	for x.Edges.Parent != nil {
		x = x.Edges.Parent
		array = append(array, x.ID.String())
	}

	ns := ino.Edges.Namespace
	array = append(array, ns.ID.String())

	var keys = make([]string, 0)
	for i := len(array) - 1; i >= 0; i-- {
		keys = append(keys, array[i])
	}

	return keys

}

func (pubsub *pubsub) SubscribeInode(ino *ent.Inode) *subscription {

	keys := pubsub.walkInodeKeys(ino)

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyInode(ino *ent.Inode) {

	pubsub.log.Debugf("PS Notify Inode: %s", ino.ID.String())

	pubsub.publish(pubsubNotify(ino.ID.String()))

}

func (pubsub *pubsub) CloseInode(ino *ent.Inode) {

	pubsub.publish(pubsubDisconnect(ino.ID.String()))

}

func (pubsub *pubsub) workflowVars(wf *ent.Workflow) string {

	return fmt.Sprintf("wfvars:%s", wf.ID.String())

}

func (pubsub *pubsub) SubscribeWorkflowVariables(wf *ent.Workflow) *subscription {

	keys := pubsub.walkInodeKeys(wf.Edges.Inode)

	keys = append(keys, wf.ID.String(), pubsub.workflowVars(wf))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyWorkflowVariables(wf *ent.Workflow) {

	pubsub.publish(pubsubNotify(pubsub.workflowVars(wf)))

}

func (pubsub *pubsub) workflowLogs(wf *ent.Workflow) string {

	return fmt.Sprintf("wflogs:%s", wf.ID.String())

}

func (pubsub *pubsub) NotifyWorkflowLogs(wf *ent.Workflow) {

	pubsub.publish(pubsubNotify(pubsub.workflowLogs(wf)))

}

func (pubsub *pubsub) SubscribeWorkflowLogs(wf *ent.Workflow) *subscription {

	keys := pubsub.walkInodeKeys(wf.Edges.Inode)

	keys = append(keys, wf.ID.String(), pubsub.workflowLogs(wf))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) SubscribeWorkflow(wf *ent.Workflow) *subscription {

	keys := pubsub.walkInodeKeys(wf.Edges.Inode)

	keys = append(keys, wf.ID.String())

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyWorkflow(wf *ent.Workflow) {

	pubsub.publish(pubsubNotify(wf.ID.String()))

}

func (pubsub *pubsub) CloseWorkflow(wf *ent.Workflow) {

	pubsub.publish(pubsubDisconnect(wf.ID.String()))

}

func (pubsub *pubsub) namespaceVars(ns *ent.Namespace) string {

	return fmt.Sprintf("nsvar:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceVariables(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceVars(ns))

}

func (pubsub *pubsub) NotifyNamespaceVariables(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceVars(ns)))

}

func (pubsub *pubsub) namespaceSecrets(ns *ent.Namespace) string {

	return fmt.Sprintf("secrets:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceSecrets(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceSecrets(ns))

}

func (pubsub *pubsub) NotifyNamespaceSecrets(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceSecrets(ns)))

}

func (pubsub *pubsub) namespaceRegistries(ns *ent.Namespace) string {

	return fmt.Sprintf("registries:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceRegistries(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceRegistries(ns))

}

func (pubsub *pubsub) NotifyNamespaceRegistries(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceRegistries(ns)))

}

func (pubsub *pubsub) instanceLogs(in *ent.Instance) string {

	return fmt.Sprintf("instlogs:%s", in.ID.String())

}

func (pubsub *pubsub) SubscribeInstanceLogs(in *ent.Instance) *subscription {

	keys := []string{}

	keys = append(keys, in.Edges.Namespace.ID.String(), pubsub.instanceLogs(in))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyInstanceLogs(in *ent.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instanceLogs(in)))

}

func (pubsub *pubsub) instanceVars(in *ent.Instance) string {

	return fmt.Sprintf("instvar:%s", in.ID.String())

}

func (pubsub *pubsub) SubscribeInstanceVariables(in *ent.Instance) *subscription {

	return pubsub.Subscribe(in.Edges.Namespace.ID.String(), pubsub.instanceVars(in))

}

func (pubsub *pubsub) NotifyInstanceVariables(in *ent.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instanceVars(in)))

}

func (pubsub *pubsub) instances(ns *ent.Namespace) string {

	return fmt.Sprintf("instances:%s", ns.ID.String())

}

func (pubsub *pubsub) NotifyInstances(ns *ent.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.instances(ns)))

}

func (pubsub *pubsub) SubscribeInstances(ns *ent.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.instances(ns))

}

func (pubsub *pubsub) instance(in *ent.Instance) string {

	return fmt.Sprintf("instance:%s", in.ID.String())

}

func (pubsub *pubsub) NotifyInstance(in *ent.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instance(in)))

}

func (pubsub *pubsub) SubscribeInstance(in *ent.Instance) *subscription {

	return pubsub.Subscribe(in.Edges.Namespace.ID.String(), pubsub.instance(in))

}

func (pubsub *pubsub) ClusterDeleteTimer(name string) {

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubDeleteTimerFunction,
		Key:     name,
	})

}

func (pubsub *pubsub) ClusterDeleteInstanceTimers(name string) {

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubDeleteInstanceTimersFunction,
		Key:     name,
	})

}

func (pubsub *pubsub) HostnameDeleteTimer(hostname, name string) {

	pubsub.publish(&PubsubUpdate{
		Handler:  pubsubDeleteTimerFunction,
		Key:      name,
		Hostname: hostname,
	})

}

type configureRouterMessage struct {
	ID      string
	Cron    string
	Enabled bool
}

func (pubsub *pubsub) ConfigureRouterCron(id, cron string, enabled bool) {

	msg := &configureRouterMessage{
		ID:      id,
		Cron:    cron,
		Enabled: enabled,
	}

	key := marshal(msg)

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubConfigureRouterFunction,
		Key:     key,
	})

}

func (pubsub *pubsub) CancelWorkflow(id, code, message string, soft bool) {

	m := []interface{}{id, code, message, soft}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubCancelWorkflowFunction,
		Key:     string(data),
	})

}

func (pubsub *pubsub) UpdateEventDelays() {

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubUpdateEventDelays,
	})

}
