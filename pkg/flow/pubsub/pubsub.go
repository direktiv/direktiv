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
		pubsub.Publish(pubsubNotify(pubsub.instanceLogs(recipientID)))
	case recipient.Workflow:
		pubsub.Publish(pubsubNotify(pubsub.workflowLogs(recipientID)))
	case recipient.Namespace:
		pubsub.Publish(pubsubNotify(pubsub.namespaceLogs(recipientID)))
	case recipient.Mirror:
		pubsub.Publish(pubsubNotify(pubsub.activityLogs(recipientID)))
	default:
		pubsub.Publish(pubsubNotify(""))
		// panic("how?")
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

func (pubsub *Pubsub) CloseNamespace(ns *database.Namespace) {
	pubsub.Publish(pubsubDisconnect(ns.ID.String()))
}

func (pubsub *Pubsub) namespaceLogs(ns uuid.UUID) string {
	return fmt.Sprintf("nslog:%s", ns.String())
}

func (pubsub *Pubsub) SubscribeNamespaceLogs(ns uuid.UUID) *Subscription {
	return pubsub.Subscribe(ns.String(), pubsub.namespaceLogs(ns))
}

func (pubsub *Pubsub) namespaceEventListeners(id uuid.UUID) string {
	return fmt.Sprintf("nsel:%s", id.String())
}

func (pubsub *Pubsub) SubscribeEventListeners(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceEventListeners(ns.ID))
}

func (pubsub *Pubsub) NotifyEventListeners(id uuid.UUID) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceEventListeners(id)))
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

func (pubsub *Pubsub) workflowVars(id uuid.UUID) string {
	return fmt.Sprintf("wfvars:%s", id.String())
}

func (pubsub *Pubsub) SubscribeWorkflowVariables(id uuid.UUID) *Subscription {
	keys := []string{pubsub.workflowVars(id)}
	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) NotifyWorkflowVariables(id uuid.UUID) {
	pubsub.Publish(pubsubNotify(pubsub.workflowVars(id)))
}

func (pubsub *Pubsub) workflowLogs(wf uuid.UUID) string {
	return fmt.Sprintf("wflogs:%s", wf.String())
}

func (pubsub *Pubsub) SubscribeWorkflowLogs(id uuid.UUID) *Subscription {
	keys := []string{id.String()}
	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) namespaceVars(nsID uuid.UUID) string {
	return fmt.Sprintf("nsvar:%s", nsID.String())
}

func (pubsub *Pubsub) SubscribeNamespaceVariables(ns *database.Namespace) *Subscription {
	return pubsub.Subscribe(ns.ID.String(), pubsub.namespaceVars(ns.ID))
}

func (pubsub *Pubsub) NotifyNamespaceVariables(nsID uuid.UUID) {
	pubsub.Publish(pubsubNotify(pubsub.namespaceVars(nsID)))
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

func (pubsub *Pubsub) instanceLogs(instID uuid.UUID) string {
	return fmt.Sprintf("instlogs:%s", instID.String())
}

func (pubsub *Pubsub) SubscribeInstanceLogs(instID uuid.UUID) *Subscription {
	keys := []string{}

	keys = append(keys, pubsub.instanceLogs(instID))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) activityLogs(act uuid.UUID) string {
	return fmt.Sprintf("mactlogs:%s", act.String())
}

func (pubsub *Pubsub) SubscribeMirrorActivityLogs(namespaceID uuid.UUID, mirrorProcessID uuid.UUID) *Subscription {
	keys := []string{}

	keys = append(keys, namespaceID.String(), pubsub.activityLogs(mirrorProcessID))

	return pubsub.Subscribe(keys...)
}

func (pubsub *Pubsub) instanceVars(id uuid.UUID) string {
	return fmt.Sprintf("instvar:%s", id.String())
}

func (pubsub *Pubsub) SubscribeInstanceVariables(instID uuid.UUID) *Subscription {
	return pubsub.Subscribe(pubsub.instanceVars(instID))
}

func (pubsub *Pubsub) NotifyInstanceVariables(id uuid.UUID) {
	pubsub.Publish(pubsubNotify(pubsub.instanceVars(id)))
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

func (pubsub *Pubsub) instance(id uuid.UUID) string {
	return fmt.Sprintf("instance:%s", id.String())
}

func (pubsub *Pubsub) NotifyInstance(id uuid.UUID) {
	pubsub.Publish(pubsubNotify(pubsub.instance(id)))
}

func (pubsub *Pubsub) SubscribeInstance(instID uuid.UUID) *Subscription {
	return pubsub.Subscribe(pubsub.instance(instID))
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
