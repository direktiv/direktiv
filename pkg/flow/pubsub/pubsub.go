package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	PubsubNotifyFunction               = "notify"
	PubsubDisconnectFunction           = "disconnect"
	PubsubDeleteTimerFunction          = "deleteTimer"
	PubsubDeleteInstanceTimersFunction = "deleteInstanceTimers"
	PubsubDeleteActivityTimersFunction = "deleteActivityTimers"
	PubsubCancelWorkflowFunction       = "cancelWorkflow"
	PubsubConfigureRouterFunction      = "configureRouter"
	PubsubUpdateEventDelays            = "updateEventDelays"
	FlowSync                           = "flowsync"
)

type Pubsub struct {
	id       uuid.UUID
	notifier Notifier
	Log      *zap.SugaredLogger

	handlers map[string]func(*PubsubUpdate)

	closed   bool
	closer   chan bool
	Hostname string
	queue    chan *PubsubUpdate
	mtx      sync.RWMutex
	channels map[string]map[*Subscription]bool

	bufferIdx int
	buffer    []*PubsubUpdate
	bufferMtx sync.Mutex
}

func (pubsub *Pubsub) Close() error {
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

type Notifier interface {
	NotifyCluster(string) error
	NotifyHostname(string, string) error
}

func InitPubSub(log *zap.SugaredLogger, notifier Notifier, database string) (*Pubsub, error) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Errorf("pubsub error: %v %v\n", ev, err)
			os.Exit(1)
		}
	}

	listener := pq.NewListener(database, 10*time.Second,
		time.Minute, reportProblem)
	err := listener.Listen(FlowSync)
	if err != nil {
		return nil, err
	}

	pubsub := new(Pubsub)
	pubsub.id = uuid.New()
	pubsub.Log = log
	pubsub.buffer = make([]*PubsubUpdate, 1024)

	pubsub.Hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	pubsub.notifier = notifier
	pubsub.closer = make(chan bool)
	pubsub.queue = make(chan *PubsubUpdate, 1024)
	pubsub.channels = make(map[string]map[*Subscription]bool)

	go pubsub.periodicFlush()
	// go pubsub.dispatcher()

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

			reqs := make([]*PubsubUpdate, 0)
			err = json.Unmarshal([]byte(notification.Extra), &reqs)
			if err != nil {
				log.Errorf("unexpected notification on database listener: %v\n", err)
				continue
			}

			if len(reqs) == 0 {
				continue
			}

			for _, req := range reqs {
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

		}
	}(listener)

	return pubsub, nil
}

func (pubsub *Pubsub) RegisterFunction(name string, fn func(*PubsubUpdate)) {
	if _, ok := pubsub.handlers[name]; ok {
		panic(fmt.Errorf("function already exists"))
	}

	pubsub.handlers[name] = fn
}

func (pubsub *Pubsub) Notify(req *PubsubUpdate) {
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

func (pubsub *Pubsub) Disconnect(req *PubsubUpdate) {
	pubsub.mtx.RLock()
	defer pubsub.mtx.RUnlock()

	channel, exists := pubsub.channels[req.Key]
	if !exists {
		return
	}

	for sub := range channel {
		go func(sub *Subscription) {
			_ = sub.Close()
		}(sub)
	}
}

type Subscription struct {
	keys   []string
	ch     chan bool
	closed bool
	pubsub *Pubsub
	// last   time.Time
}

func (s *Subscription) Wait(ctx context.Context) bool {
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

func (s *Subscription) Close() error {
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

func (pubsub *Pubsub) periodicFlush() {
	for {
		time.Sleep(time.Millisecond)
		pubsub.timeFlush()
	}
}

func (pubsub *Pubsub) flush() {
	slice := pubsub.buffer[:pubsub.bufferIdx]
	clusterMessages := make([]string, pubsub.bufferIdx)
	messageIndex := 0
	pubsub.bufferIdx = 0

	set := make(map[string]bool)

	for idx := range slice {
		req := slice[idx]

		b, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}

		if _, exists := set[string(b)]; exists {
			continue
		}
		set[string(b)] = true

		handler, exists := pubsub.handlers[req.Handler]
		if !exists {
			pubsub.Log.Errorf("unexpected notification type on database listener: %v\n", req.Handler)
		} else {
			go handler(req)
		}

		if req.Hostname == "" {
			x := *req
			x.Sender = ""
			go pubsub.Notify(&x)
			clusterMessages[messageIndex] = string(b)
			messageIndex++
		} else {
			err = pubsub.notifier.NotifyHostname(req.Hostname, "["+string(b)+"]")
		}

		if err != nil {
			pubsub.Log.Errorf("pubsub error: %v\n", err)
			os.Exit(1)
		}
	}

	clusterSlice := clusterMessages[:messageIndex]
	msg := "["
	l := 3
	comma := false

	for idx := range clusterSlice {

		s := clusterSlice[idx]

		if l+len(s) >= 8000 {
			msg += "]"
			err := pubsub.notifier.NotifyCluster(msg)
			if err != nil {
				pubsub.Log.Errorf("pubsub error: %v\n", err)
				os.Exit(1)
			}

			msg = "["
			comma = false
			l = 3
		} else {
			if comma {
				msg += ","
				l++
			}
			msg += s
			comma = true
			l += len(s)
		}

	}

	if l > 3 {
		msg += "]"
		err := pubsub.notifier.NotifyCluster(msg)
		if err != nil {
			pubsub.Log.Errorf("pubsub error: %v\n", err)
			os.Exit(1)
		}
	}

	pubsub.bufferMtx.Unlock()
}

func (pubsub *Pubsub) timeFlush() {
	pubsub.bufferMtx.Lock()
	go pubsub.flush()
}

func (pubsub *Pubsub) Publish(req *PubsubUpdate) {
	req.Sender = pubsub.id.String()

	pubsub.bufferMtx.Lock()

	pubsub.buffer[pubsub.bufferIdx] = req
	pubsub.bufferIdx++

	if pubsub.bufferIdx >= 1024 {
		go pubsub.flush()
	} else {
		pubsub.bufferMtx.Unlock()
	}

	/*
		select {
		case pubsub.queue <- req:
		default:
		}
	*/
}

func (pubsub *Pubsub) Subscribe(id ...string) *Subscription {
	s := new(Subscription)

	s.ch = make(chan bool, 1)
	s.pubsub = pubsub

	pubsub.mtx.Lock()

	key := ""
	s.keys = append(s.keys, key)
	channel, exists := pubsub.channels[key]
	if !exists {
		channel = make(map[*Subscription]bool)
		pubsub.channels[key] = channel
	}
	channel[s] = len(id) == 0

	for idx, x := range id {
		key := x
		s.keys = append(s.keys, key)
		channel, exists := pubsub.channels[key]
		if !exists {
			channel = make(map[*Subscription]bool)
			pubsub.channels[key] = channel
		}
		channel[s] = idx == len(id)-1
	}

	pubsub.mtx.Unlock()

	return s
}

func pubsubNotify(key string) *PubsubUpdate {
	return &PubsubUpdate{
		Handler: PubsubNotifyFunction,
		Key:     key,
	}
}

func pubsubDisconnect(key string) *PubsubUpdate {
	return &PubsubUpdate{
		Handler: PubsubDisconnectFunction,
		Key:     key,
	}
}

func (pubsub *Pubsub) NotifyLogs(recipientID uuid.UUID, recipientType recipient.RecipientType) {
	switch recipientType {
	case recipient.Server:
		pubsub.Publish(pubsubNotify(""))
	case recipient.Instance:
		pubsub.Publish(pubsubNotify(pubsub.instanceLogs(&recipientID)))
	case recipient.Workflow:
		pubsub.Publish(pubsubNotify(pubsub.workflowLogs(&recipientID)))
	case recipient.Namespace:
		pubsub.Publish(pubsubNotify(pubsub.namespaceLogs(&recipientID)))
	case recipient.Mirror:
		pubsub.Publish(pubsubNotify(pubsub.activityLogs(&recipientID)))
	default:
		panic("how?")
	}
}

func (pubsub *Pubsub) SubscribeServerLogs() *Subscription {
	return pubsub.Subscribe()
}

func (pubsub *Pubsub) SubscribeNamespaces() *Subscription {
	return pubsub.Subscribe("namespaces")
}

func (pubsub *Pubsub) NotifyNamespaces() {
	pubsub.Publish(pubsubNotify("namespaces"))
}

func (pubsub *Pubsub) SubscribeNamespace(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String())
}

func (pubsub *Pubsub) NotifyNamespace(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(ns.ID.String()))
}

func (pubsub *Pubsub) CloseNamespace(ns *database.Namespace) {
	pubsub.Publish(pubsubDisconnect(ns.ID.String()))
}

func (pubsub *Pubsub) namespaceLogs(ns *uuid.UUID) string {
	return fmt.Sprintf("nslog:%s", ns.String())
}

func (pubsub *Pubsub) SubscribeNamespaceLogs(ns *uuid.UUID) *Subscription {
	return pubsub.Subscribe(ns.String(), pubsub.namespaceLogs(ns))
}

func (pubsub *Pubsub) namespaceEventListeners(ns *database.Namespace) string {
	return fmt.Sprintf("nsel:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeEventListeners(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceEventListeners(ns))
}

func (pubsub *Pubsub) NotifyEventListeners(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceEventListeners(ns)))
}

func (pubsub *Pubsub) namespaceEvents(ns *database.Namespace) string {
	return fmt.Sprintf("nsev:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeEvents(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceEvents(ns))
}

func (pubsub *Pubsub) NotifyEvents(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceEvents(ns)))
}

func (pubsub *Pubsub) walkInodeKeys(cached *database.CacheData) []string {
	array := make([]string, 0)

	for i := len(cached.Inodes) - 1; i >= 0; i-- {
		x := cached.Inodes[i]
		array = append(array, x.ID.String())
	}

	array = append(array, cached.Namespace.ID.String())

	keys := make([]string, 0)
	for i := len(array) - 1; i >= 0; i-- {
		keys = append(keys, array[i])
	}

	return keys
}

func (pubsub *Pubsub) SubscribeInode(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyInode(ino *database.Inode) {
	// pubsub.log.Debugf("PS Notify Inode: %s", ino.ID.String())

	pubsub.Publish(pubsubNotify(ino.ID.String()))
}

func (pubsub *Pubsub) CloseInode(ino *database.Inode) {
	pubsub.Publish(pubsubDisconnect(ino.ID.String()))
}

func (pubsub *Pubsub) inodeAnnotations(ino *database.Inode) string {
	return fmt.Sprintf("inonotes:%s", ino.ID.String())
}

func (pubsub *Pubsub) SubscribeInodeAnnotations(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	ino := cached.Inodes[len(cached.Inodes)-1]
	keys = append(keys, pubsub.inodeAnnotations(ino))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) mirror(ino *database.Inode) string {
	return fmt.Sprintf("mirror:%s", ino.ID.String())
}

func (pubsub *Pubsub) NotifyInodeAnnotations(ino *database.Inode) {
	pubsub.Publish(pubsubNotify(pubsub.inodeAnnotations(ino)))
}

func (pubsub *Pubsub) SubscribeMirror(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	ino := cached.Inodes[len(cached.Inodes)-1]
	keys = append(keys, pubsub.mirror(ino))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyMirror(ino *database.Inode) {
	pubsub.Publish(pubsubNotify(pubsub.mirror(ino)))
}

func (pubsub *Pubsub) CloseMirror(ino *database.Inode) {
	pubsub.Publish(pubsubDisconnect(pubsub.mirror(ino)))
}

func (pubsub *Pubsub) workflowVars(wf *database.Workflow) string {
	return fmt.Sprintf("wfvars:%s", wf.ID.String())
}

func (pubsub *Pubsub) SubscribeWorkflowVariables(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowVars(cached.Workflow))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyWorkflowVariables(wf *database.Workflow) {
	pubsub.Publish(pubsubNotify(pubsub.workflowVars(wf)))
}

func (pubsub *Pubsub) workflowAnnotations(wf *database.Workflow) string {
	return fmt.Sprintf("wfnotes:%s", wf.ID.String())
}

func (pubsub *Pubsub) SubscribeWorkflowAnnotations(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowAnnotations(cached.Workflow))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyWorkflowAnnotations(wf *database.Workflow) {
	pubsub.Publish(pubsubNotify(pubsub.workflowAnnotations(wf)))
}

func (pubsub *Pubsub) workflowLogs(wf *uuid.UUID) string {
	return fmt.Sprintf("wflogs:%s", wf.String())
}

func (pubsub *Pubsub) SubscribeWorkflowLogs(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String(), pubsub.workflowLogs(&cached.Workflow.ID))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) SubscribeWorkflow(cached *database.CacheData) *Subscription {
	keys := pubsub.walkInodeKeys(cached)

	keys = append(keys, cached.Workflow.ID.String())

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyWorkflow(wf *database.Workflow) {
	pubsub.Publish(pubsubNotify(wf.ID.String()))
}

func (pubsub *Pubsub) CloseWorkflow(wf *database.Workflow) {
	pubsub.Publish(pubsubDisconnect(wf.ID.String()))
}

func (pubsub *Pubsub) namespaceVars(ns *database.Namespace) string {
	return fmt.Sprintf("nsvar:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeNamespaceVariables(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceVars(ns))
}

func (pubsub *Pubsub) NotifyNamespaceVariables(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceVars(ns)))
}

func (pubsub *Pubsub) namespaceAnnotations(ns *database.Namespace) string {
	return fmt.Sprintf("nsnote:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeNamespaceAnnotations(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceAnnotations(ns))
}

func (pubsub *Pubsub) NotifyNamespaceAnnotations(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceAnnotations(ns)))
}

func (pubsub *Pubsub) namespaceSecrets(ns *database.Namespace) string {
	return fmt.Sprintf("secrets:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeNamespaceSecrets(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceSecrets(ns))
}

func (pubsub *Pubsub) NotifyNamespaceSecrets(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceSecrets(ns)))
}

func (pubsub *Pubsub) namespaceRegistries(ns *database.Namespace) string {
	return fmt.Sprintf("registries:%s", ns.ID.String())
}

func (pubsub *Pubsub) SubscribeNamespaceRegistries(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceRegistries(ns))
}

func (pubsub *Pubsub) NotifyNamespaceRegistries(ns *database.Namespace) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceRegistries(ns)))
}

func (pubsub *Pubsub) instanceLogs(in *uuid.UUID) string {
	return fmt.Sprintf("instlogs:%s", in.String())
}

func (pubsub *Pubsub) SubscribeInstanceLogs(cached *database.CacheData) *Subscription {
	keys := []string{}

	keys = append(keys, cached.Namespace.ID.String(), pubsub.instanceLogs(&cached.Instance.ID))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) activityLogs(act *uuid.UUID) string {
	return fmt.Sprintf("mactlogs:%s", act.String())
}

func (pubsub *Pubsub) SubscribeMirrorActivityLogs(ns *database.Namespace, act *database.MirrorActivity) *Subscription {
	keys := []string{}

	keys = append(keys, ns.ID.String(), pubsub.activityLogs(&act.ID))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) instanceVars(in *database.Instance) string {
	return fmt.Sprintf("instvar:%s", in.ID.String())
}

func (pubsub *Pubsub) SubscribeInstanceVariables(cached *database.CacheData) *Subscription {
	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instanceVars(cached.Instance))
}

func (pubsub *Pubsub) NotifyInstanceVariables(in *database.Instance) {
	pubsub.Publish(pubsubNotify(pubsub.instanceVars(in)))
}

func (pubsub *Pubsub) instanceAnnotations(in *database.Instance) string {
	return fmt.Sprintf("instnote:%s", in.ID.String())
}

func (pubsub *Pubsub) SubscribeInstanceAnnotations(cached *database.CacheData) *Subscription {
	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instanceAnnotations(cached.Instance))
}

func (pubsub *Pubsub) NotifyInstanceAnnotations(in *database.Instance) {
	pubsub.Publish(pubsubNotify(pubsub.instanceAnnotations(in)))
}

func (pubsub *Pubsub) instances(ns *database.Namespace) string {
	return fmt.Sprintf("instances:%s", ns.ID.String())
}

func (pubsub *Pubsub) NotifyInstances(ns *database.Namespace) {
	// pubsub.publish(pubsubNotify(pubsub.instances(ns)))
}

func (pubsub *Pubsub) SubscribeInstances(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.instances(ns))
}

func (pubsub *Pubsub) instance(in *database.Instance) string {
	return fmt.Sprintf("instance:%s", in.ID.String())
}

func (pubsub *Pubsub) NotifyInstance(in *database.Instance) {
	pubsub.Publish(pubsubNotify(pubsub.instance(in)))
}

func (pubsub *Pubsub) SubscribeInstance(cached *database.CacheData) *Subscription {
	return pubsub.Subscribe(cached.Namespace.ID.String(), pubsub.instance(cached.Instance))
}

func (pubsub *Pubsub) ClusterDeleteTimer(name string) {
	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubDeleteTimerFunction,
		Key:     name,
	})
}

func (pubsub *Pubsub) ClusterDeleteInstanceTimers(name string) {
	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubDeleteInstanceTimersFunction,
		Key:     name,
	})
}

func (pubsub *Pubsub) ClusterDeleteActivityTimers(name string) {
	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubDeleteActivityTimersFunction,
		Key:     name,
	})
}

func (pubsub *Pubsub) HostnameDeleteTimer(hostname, name string) {
	pubsub.Publish(&PubsubUpdate{
		Handler:  PubsubDeleteTimerFunction,
		Key:      name,
		Hostname: hostname,
	})
}

type ConfigureRouterMessage struct {
	ID      string
	Cron    string
	Enabled bool
}

func (pubsub *Pubsub) ConfigureRouterCron(id, cron string, enabled bool) {
	msg := &ConfigureRouterMessage{
		ID:      id,
		Cron:    cron,
		Enabled: enabled,
	}

	key := bytedata.Marshal(msg)

	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubConfigureRouterFunction,
		Key:     key,
	})
}

func (pubsub *Pubsub) CancelWorkflow(id, code, message string, soft bool) {
	m := []interface{}{id, code, message, soft}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubCancelWorkflowFunction,
		Key:     string(data),
	})
}

func (pubsub *Pubsub) UpdateEventDelays() {
	pubsub.Publish(&PubsubUpdate{
		Handler: PubsubUpdateEventDelays,
	})
}
