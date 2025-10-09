package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

var ErrDataNotFound = fmt.Errorf("data not found")

// LabelWithNotify used to mark an instance as called with a notify-chanel.
const LabelWithNotify = "WithNotify"

type Engine struct {
	db       *gorm.DB
	dataBus  DataBus
	compiler core.Compiler
	js       nats.JetStreamContext
}

func (e *Engine) ListInstances(ctx context.Context, namespace string) ([]uuid.UUID, error) {
	// TODO implement me
	panic("implement me")
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

func (e *Engine) StartWorkflow(ctx context.Context, namespace string, workflowPath string, input string, metadata map[string]string) (*InstanceStatus, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return nil, fmt.Errorf("fetch script: %w", err)
	}

	return e.startScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, input, nil, metadata)
}

func (e *Engine) RunWorkflow(ctx context.Context, namespace string, workflowPath string, input string, metadata map[string]string) (<-chan *InstanceStatus, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return nil, fmt.Errorf("fetch script: %w", err)
	}

	notify := make(chan *InstanceStatus, 1)
	_, err = e.startScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, input, notify, metadata)
	if err != nil {
		return nil, err
	}

	return notify, nil
}

func (e *Engine) startScript(ctx context.Context, namespace string, script string, mappings string, fn string, input string, notify chan<- *InstanceStatus, metadata map[string]string) (*InstanceStatus, error) {
	if !json.Valid([]byte(input)) {
		return nil, fmt.Errorf("input is not a valid json string: %s", input)
	}
	instID := uuid.New()

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[LabelWithNotify] = "no"
	if notify != nil {
		metadata[LabelWithNotify] = "yes"
	}

	pEv := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  namespace,
		Type:       StateCodePending,
		Time:       time.Now(),
		Metadata:   metadata,

		Script:   script,
		Mappings: mappings,
		Fn:       fn,
		Input:    json.RawMessage(input),
	}
	err := e.dataBus.PublishInstanceHistoryEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push history stream: %w", err)
	}

	if notify != nil {
		e.dataBus.NotifyInstanceStatus(ctx, instID, notify)
	}

	err = e.dataBus.PublishInstanceQueueEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push queue stream: %w", err)
	}

	st := &InstanceStatus{}
	ApplyInstanceEvent(st, pEv)

	return st, nil
}

func (e *Engine) execInstance(ctx context.Context, inst *InstanceEvent) error {
	startEv := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
		Type:       StateCodeRunning,
		Fn:         inst.Fn,
		Input:      inst.Input,
		Time:       time.Now(),
	}

	err := e.dataBus.PublishInstanceHistoryEvent(ctx, startEv)
	if err != nil {
		return fmt.Errorf("push history start event, inst: %s: %w", inst.InstanceID, err)
	}

	sc := &runtime.Script{
		InstID:   inst.InstanceID,
		Text:     inst.Script,
		Mappings: inst.Mappings,
		Fn:       inst.Fn,
		Input:    string(inst.Input),
		Metadata: inst.Metadata,
	}

	onFinish := func(output []byte) error {
		endEv := &InstanceEvent{
			EventID:    uuid.New(),
			InstanceID: inst.InstanceID,
			Namespace:  inst.Namespace,
			Type:       StateCodeComplete,
			Output:     output,
			Time:       time.Now(),
		}

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}
	onTransition := func(memory []byte, fn string) error {
		endEv := &InstanceEvent{
			EventID:    uuid.New(),
			InstanceID: inst.InstanceID,
			Namespace:  inst.Namespace,
			Type:       StateCodeRunning,
			Fn:         fn,
			Memory:     memory,
			Time:       time.Now(),
		}

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}

	err = runtime.ExecScript(sc, onFinish, onTransition)
	if err == nil {
		return nil
	}
	endEv := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
		Type:       StateCodeFailed,
		Error:      err.Error(),
		Time:       time.Now(),
	}
	err = e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	if err != nil {
		return fmt.Errorf("push history end event, inst: %s: %w", inst.InstanceID, err)
	}

	return nil
}

func (e *Engine) ListInstanceStatuses(ctx context.Context, namespace string, limit int, offset int) ([]*InstanceStatus, int, error) {
	data, total := e.dataBus.ListInstanceStatuses(ctx, namespace, uuid.Nil, limit, offset)
	if len(data) == 0 {
		return nil, 0, ErrDataNotFound
	}

	return data, total, nil
}

func (e *Engine) GetInstanceStatus(ctx context.Context, namespace string, id uuid.UUID) (*InstanceStatus, error) {
	data, _ := e.dataBus.ListInstanceStatuses(ctx, namespace, id, 0, 0)
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
