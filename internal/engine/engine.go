package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

var ErrDataNotFound = fmt.Errorf("data not found")

// LabelWithNotify used to mark an instance as called with a notify-chanel.
const (
	LabelWithNotify   = "WithNotify"
	LabelWithSyncExec = "WithSyncExec"
)

type Engine struct {
	db       *gorm.DB
	dataBus  DataBus
	compiler core.Compiler
	js       nats.JetStreamContext
}

func NewEngine(bus DataBus, compiler core.Compiler, js nats.JetStreamContext) (*Engine, error) {
	return &Engine{
		dataBus:  bus,
		compiler: compiler,
		js:       js,
	}, nil
}

func (e *Engine) Start(lc *lifecycle.Manager) error {
	err := e.dataBus.Start(lc)
	if err != nil {
		return fmt.Errorf("start databus: %w", err)
	}

	err = e.startQueueWorkers(lc)
	if err != nil {
		return fmt.Errorf("start queue workers: %w", err)
	}

	return nil
}

func (e *Engine) StartWorkflow(ctx context.Context, namespace string, workflowPath string, input string, metadata map[string]string) (*InstanceEvent, <-chan *InstanceEvent, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch script: %w", err)
	}

	notify := make(chan *InstanceEvent, 1)
	st, err := e.startScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, input, notify, metadata)
	if err != nil {
		return nil, nil, err
	}

	return st, notify, nil
}

var (
	notifyMap  = map[string]chan<- *InstanceEvent{}
	notifyLock = &sync.Mutex{}
)

func (e *Engine) startScript(ctx context.Context, namespace string, script string, mappings string, fn string, input string, notify chan<- *InstanceEvent, metadata map[string]string) (*InstanceEvent, error) {
	if !json.Valid([]byte(input)) {
		return nil, fmt.Errorf("input is not a valid json string: %s", input)
	}
	instID := uuid.New()

	if metadata == nil {
		metadata = make(map[string]string)
	}

	pEv := &InstanceEvent{
		State: StateCodePending,

		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  namespace,
		Metadata:   metadata,
		Script:     script,
		Fn:         fn,
		Mappings:   mappings,

		Input:  json.RawMessage(input),
		Output: nil,
		Error:  "",

		CreatedAt: time.Now(),
		StartedAt: time.Time{},
		EndedAt:   time.Time{},
	}
	err := e.dataBus.PublishInstanceHistoryEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push history stream: %w", err)
	}

	if notify != nil {
		notifyLock.Lock()
		notifyMap[instID.String()] = notify
		notifyLock.Unlock()
	}

	if metadata[LabelWithSyncExec] == "true" {
		err = e.execInstance(ctx, pEv)
		if err != nil {
			return nil, fmt.Errorf("exec instance: %w", err)
		}

		return pEv, nil
	}

	err = e.dataBus.PublishInstanceQueueEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push queue stream: %w", err)
	}

	return pEv, nil
}

func (e *Engine) execInstance(ctx context.Context, inst *InstanceEvent) error {
	startEv := inst.Clone()
	startEv.EventID = uuid.New()
	startEv.State = StateCodeRunning
	startEv.StartedAt = time.Now()

	err := e.dataBus.PublishInstanceHistoryEvent(ctx, startEv)
	if err != nil {
		return fmt.Errorf("push history start event, inst: %s: %w", inst.InstanceID, err)
	}

	sc := &runtime.Script{
		InstID:   startEv.InstanceID,
		Text:     startEv.Script,
		Mappings: startEv.Mappings,
		Fn:       startEv.Fn,
		Input:    string(startEv.Input),
		Metadata: startEv.Metadata,
	}

	onFinish := func(output []byte) error {
		endEv := startEv.Clone()
		endEv.EventID = uuid.New()
		endEv.State = StateCodeComplete
		endEv.Output = output
		endEv.EndedAt = time.Now()
		endEv.Fn = ""

		if endEv.Metadata[LabelWithNotify] == "true" {
			notifyLock.Lock()
			notify, ok := notifyMap[endEv.InstanceID.String()]
			notifyLock.Unlock()
			if ok {
				notify <- endEv
			}
		}

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}
	onTransition := func(memory []byte, fn string) error {
		endEv := startEv.Clone()
		endEv.EventID = uuid.New()
		endEv.State = StateCodeRunning
		endEv.Output = memory
		endEv.Fn = fn

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}

	err = runtime.ExecScript(sc, onFinish, onTransition)
	if err == nil {
		return nil
	}
	endEv := startEv.Clone()
	endEv.EventID = uuid.New()
	endEv.State = StateCodeFailed
	endEv.Error = err.Error()
	endEv.EndedAt = time.Now()

	if inst.Metadata[LabelWithNotify] == "true" {
		notifyLock.Lock()
		notify, ok := notifyMap[endEv.InstanceID.String()]
		notifyLock.Unlock()
		if ok {
			notify <- endEv
		}
	}
	err = e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	if err != nil {
		return fmt.Errorf("push history end event, inst: %s: %w", inst.InstanceID, err)
	}

	return nil
}

func (e *Engine) ListInstanceStatuses(ctx context.Context, limit int, offset int, filters filter.Values) ([]*InstanceEvent, int, error) {
	data, total := e.dataBus.ListInstanceStatuses(ctx, limit, offset, filters)

	return data, total, nil
}

func (e *Engine) GetInstanceStatus(ctx context.Context, namespace string, id uuid.UUID) (*InstanceEvent, error) {
	data, _ := e.dataBus.ListInstanceStatuses(ctx, 0, 0, filter.With(nil,
		filter.FieldEQ("namespace", namespace),
		filter.FieldEQ("instanceID", id.String()),
	))
	if len(data) == 0 {
		return nil, ErrDataNotFound
	}

	return data[0], nil
}

func (e *Engine) GetInstanceHistory(ctx context.Context, namespace string, id uuid.UUID) ([]*InstanceEvent, error) {
	list := e.dataBus.GetInstanceHistory(ctx, namespace, id)
	if len(list) == 0 {
		return nil, ErrDataNotFound
	}

	return list, nil
}

func (e *Engine) DeleteNamespace(ctx context.Context, name string) error {
	return e.dataBus.DeleteNamespace(ctx, name)
}
