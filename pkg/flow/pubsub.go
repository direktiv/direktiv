package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	pubsubNotifyFunction               = "notify"
	pubsubDisconnectFunction           = "disconnect"
	pubsubDeleteTimerFunction          = "deleteTimer"
	pubsubDeleteInstanceTimersFunction = "deleteInstanceTimers"
	pubsubDeleteActivityTimersFunction = "deleteActivityTimers"
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
			log.Errorf("pubsub error: %v %v\n", ev, err)
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

	go func(l *pq.Listener) {

		defer func() {
			err := pubsub.Close()
			if err != nil {
				if !errors.Is(err, os.ErrClosed) {
					log.Errorf("Error closing pubsub: %v.", err)
				}
			}
		}()

		defer func() {
			err := l.UnlistenAll()
			if err != nil {
				log.Errorf("Error deregistering listeners: %v.", err)
			}
		}()

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

			if req.Sender == pubsub.id.String() {
				continue
			}

			handler, exists := pubsub.handlers[req.Handler]
			if !exists {
				log.Errorf("unexpected notification type on database listener: %v\n", err)
				continue
			}

			go handler(req)

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

		b, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}

		handler, exists := pubsub.handlers[req.Handler]
		if !exists {
			pubsub.log.Errorf("unexpected notification type on database listener: %v\n", err)
		} else {
			go handler(req)
		}

		if req.Hostname == "" {
			x := *req
			x.Sender = ""
			go pubsub.Notify(&x)
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

	if pubsub.id.String() == req.Sender {
		return
	}

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
	// last   time.Time
}

func (s *subscription) Wait(ctx context.Context) bool {

	/*
		t := time.Now()
		dt := t.Sub(s.last)
		if dt < (time.Millisecond * 150) {
			time.Sleep(dt)
		}

		defer func() {
			s.last = time.Now()
		}()
	*/

	select {
	case <-ctx.Done():
		return false
	case _, more := <-s.ch:
		return more
	case <-time.After(time.Second * 30):
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

func (pubsub *pubsub) SubscribeNamespace(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String())

}

func (pubsub *pubsub) NotifyNamespace(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(ns.ID.String()))

}

func (pubsub *pubsub) CloseNamespace(ns *database.Namespace) {

	pubsub.publish(pubsubDisconnect(ns.ID.String()))

}

func (pubsub *pubsub) namespaceLogs(ns *database.Namespace) string {

	return fmt.Sprintf("nslog:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceLogs(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceLogs(ns))

}

func (pubsub *pubsub) NotifyNamespaceLogs(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceLogs(ns)))

}

func (pubsub *pubsub) namespaceEventListeners(ns *database.Namespace) string {

	return fmt.Sprintf("nsel:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeEventListeners(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceEventListeners(ns))

}

func (pubsub *pubsub) NotifyEventListeners(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceEventListeners(ns)))

}

func (pubsub *pubsub) namespaceEvents(ns *database.Namespace) string {

	return fmt.Sprintf("nsev:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeEvents(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceEvents(ns))

}

func (pubsub *pubsub) NotifyEvents(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceEvents(ns)))

}

func (pubsub *pubsub) walkInodeKeys(cached *database.CacheData) []string {

	array := make([]string, 0)

	for i := len(cached.Inodes) - 1; i >= 0; i-- {
		x := cached.Inodes[i]
		array = append(array, x.ID.String())
	}

	array = append(array, cached.Namespace.ID.String())

	var keys = make([]string, 0)
	for i := len(array) - 1; i >= 0; i-- {
		keys = append(keys, array[i])
	}

	return keys

}

func (pubsub *pubsub) SubscribeInode(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyInode(ino *database.Inode) {

	pubsub.log.Debugf("PS Notify Inode: %s", ino.ID.String())

	pubsub.publish(pubsubNotify(ino.ID.String()))

}

func (pubsub *pubsub) CloseInode(ino *database.Inode) {

	pubsub.publish(pubsubDisconnect(ino.ID.String()))

}

func (pubsub *pubsub) inodeAnnotations(ino *database.Inode) string {

	return fmt.Sprintf("inonotes:%s", ino.ID.String())

}

func (pubsub *pubsub) SubscribeInodeAnnotations(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	ino := cached.Inodes[len(cached.Inodes)-1]
	keys = append(keys, pubsub.inodeAnnotations(ino))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) mirror(ino *database.Inode) string {

	return fmt.Sprintf("mirror:%s", ino.ID.String())

}

func (pubsub *pubsub) NotifyInodeAnnotations(ino *database.Inode) {

	pubsub.publish(pubsubNotify(pubsub.inodeAnnotations(ino)))

}

func (pubsub *pubsub) SubscribeMirror(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	ino := cached.Inodes[len(cached.Inodes)-1]
	keys = append(keys, pubsub.mirror(ino))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyMirror(ino *database.Inode) {

	pubsub.publish(pubsubNotify(pubsub.mirror(ino)))

}

func (pubsub *pubsub) CloseMirror(ino *database.Inode) {

	pubsub.publish(pubsubDisconnect(pubsub.mirror(ino)))

}

func (pubsub *pubsub) workflowVars(wf *database.Workflow) string {

	return fmt.Sprintf("wfvars:%s", wf.ID.String())

}

func (pubsub *pubsub) SubscribeWorkflowVariables(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowVars(cached.Workflow))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyWorkflowVariables(wf *database.Workflow) {

	pubsub.publish(pubsubNotify(pubsub.workflowVars(wf)))

}

func (pubsub *pubsub) workflowAnnotations(wf *database.Workflow) string {

	return fmt.Sprintf("wfnotes:%s", wf.ID.String())

}

func (pubsub *pubsub) SubscribeWorkflowAnnotations(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowAnnotations(cached.Workflow))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyWorkflowAnnotations(wf *database.Workflow) {

	pubsub.publish(pubsubNotify(pubsub.workflowAnnotations(wf)))

}

func (pubsub *pubsub) workflowLogs(wf *database.Workflow) string {

	return fmt.Sprintf("wflogs:%s", wf.ID.String())

}

func (pubsub *pubsub) NotifyWorkflowLogs(wf *database.Workflow) {

	pubsub.publish(pubsubNotify(pubsub.workflowLogs(wf)))

}

func (pubsub *pubsub) SubscribeWorkflowLogs(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowLogs(cached.Workflow))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) SubscribeWorkflow(cached *database.CacheData) *subscription {

	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String())

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyWorkflow(wf *database.Workflow) {

	pubsub.publish(pubsubNotify(wf.ID.String()))

}

func (pubsub *pubsub) CloseWorkflow(wf *database.Workflow) {

	pubsub.publish(pubsubDisconnect(wf.ID.String()))

}

func (pubsub *pubsub) namespaceVars(ns *database.Namespace) string {

	return fmt.Sprintf("nsvar:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceVariables(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceVars(ns))

}

func (pubsub *pubsub) NotifyNamespaceVariables(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceVars(ns)))

}

func (pubsub *pubsub) namespaceAnnotations(ns *database.Namespace) string {

	return fmt.Sprintf("nsnote:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceAnnotations(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceAnnotations(ns))

}

func (pubsub *pubsub) NotifyNamespaceAnnotations(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceAnnotations(ns)))

}

func (pubsub *pubsub) namespaceSecrets(ns *database.Namespace) string {

	return fmt.Sprintf("secrets:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceSecrets(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceSecrets(ns))

}

func (pubsub *pubsub) NotifyNamespaceSecrets(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceSecrets(ns)))

}

func (pubsub *pubsub) namespaceRegistries(ns *database.Namespace) string {

	return fmt.Sprintf("registries:%s", ns.ID.String())

}

func (pubsub *pubsub) SubscribeNamespaceRegistries(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceRegistries(ns))

}

func (pubsub *pubsub) NotifyNamespaceRegistries(ns *database.Namespace) {

	pubsub.publish(pubsubNotify(pubsub.namespaceRegistries(ns)))

}

func (pubsub *pubsub) instanceLogs(in *database.Instance) string {

	return fmt.Sprintf("instlogs:%s", in.ID.String())

}

func (pubsub *pubsub) SubscribeInstanceLogs(cached *database.CacheData) *subscription {

	keys := []string{}

	keys = append(keys, cached.Namespace.ID.String(), pubsub.instanceLogs(cached.Instance))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyInstanceLogs(in *database.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instanceLogs(in)))

}

func (pubsub *pubsub) activityLogs(act *database.MirrorActivity) string {
	return fmt.Sprintf("mactlogs:%s", act.ID.String())
}

func (pubsub *pubsub) SubscribeMirrorActivityLogs(ns *database.Namespace, act *database.MirrorActivity) *subscription {

	keys := []string{}

	keys = append(keys, ns.ID.String(), pubsub.activityLogs(act))

	return pubsub.Subscribe(keys...)

}

func (pubsub *pubsub) NotifyMirrorActivityLogs(act *database.MirrorActivity) {

	pubsub.publish(pubsubNotify(pubsub.activityLogs(act)))

}

func (pubsub *pubsub) instanceVars(in *database.Instance) string {

	return fmt.Sprintf("instvar:%s", in.ID.String())

}

func (pubsub *pubsub) SubscribeInstanceVariables(cached *database.CacheData) *subscription {

	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instanceVars(cached.Instance))

}

func (pubsub *pubsub) NotifyInstanceVariables(in *database.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instanceVars(in)))

}

func (pubsub *pubsub) instanceAnnotations(in *database.Instance) string {

	return fmt.Sprintf("instnote:%s", in.ID.String())

}

func (pubsub *pubsub) SubscribeInstanceAnnotations(cached *database.CacheData) *subscription {

	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instanceAnnotations(cached.Instance))

}

func (pubsub *pubsub) NotifyInstanceAnnotations(in *database.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instanceAnnotations(in)))

}

func (pubsub *pubsub) instances(ns *database.Namespace) string {

	return fmt.Sprintf("instances:%s", ns.ID.String())

}

func (pubsub *pubsub) NotifyInstances(ns *database.Namespace) {

	// pubsub.publish(pubsubNotify(pubsub.instances(ns)))

}

func (pubsub *pubsub) SubscribeInstances(ns *database.Namespace) *subscription {

	return pubsub.Subscribe(ns.ID.String(), pubsub.instances(ns))

}

func (pubsub *pubsub) instance(in *database.Instance) string {

	return fmt.Sprintf("instance:%s", in.ID.String())

}

func (pubsub *pubsub) NotifyInstance(in *database.Instance) {

	pubsub.publish(pubsubNotify(pubsub.instance(in)))

}

func (pubsub *pubsub) SubscribeInstance(cached *database.CacheData) *subscription {

	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instance(cached.Instance))

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

func (pubsub *pubsub) ClusterDeleteActivityTimers(name string) {

	pubsub.publish(&PubsubUpdate{
		Handler: pubsubDeleteActivityTimersFunction,
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
