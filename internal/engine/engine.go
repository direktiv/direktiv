package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/internal/core"
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

func NewEngine(db *gorm.DB, bus DataBus, compiler core.Compiler, js nats.JetStreamContext) (*Engine, error) {
	return &Engine{
		db:       db,
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

func (e *Engine) StartWorkflow(ctx context.Context, namespace string, workflowPath string, args any, metadata map[string]string) (uuid.UUID, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return uuid.Nil, fmt.Errorf("fetch script: %w", err)
	}

	return e.startScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, args, nil, metadata)
}

func (e *Engine) RunWorkflow(ctx context.Context, namespace string, workflowPath string, args any, metadata map[string]string) (uuid.UUID, <-chan *InstanceStatus, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("fetch script: %w", err)
	}

	fmt.Printf("ACTIONS IN FLOW: %v\n", flowDetails.Config.Actions)

	notify := make(chan *InstanceStatus, 1)
	id, err := e.startScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, args, notify, metadata)

	return id, notify, err
}

func (e *Engine) startScript(ctx context.Context, namespace string, script string, mappings string, fn string, args any, notify chan<- *InstanceStatus, metadata map[string]string) (uuid.UUID, error) {
	input, ok := args.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid input")
	}
	if input == "" {
		input = "{}"
	}

	instID := uuid.New()

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[LabelWithNotify] = "no"
	if notify != nil {
		metadata[LabelWithNotify] = "yes"
	}

	ev := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  namespace,
		Type:       "pending",
		Time:       time.Now(),
		Metadata:   metadata,

		Script:   script,
		Mappings: mappings,
		Fn:       fn,
		Input:    json.RawMessage(input),
	}
	err := e.dataBus.PushHistoryStream(ctx, ev)
	if err != nil {
		return uuid.Nil, fmt.Errorf("push history stream: %w", err)
	}

	if notify != nil {
		e.dataBus.NotifyInstanceStatus(ctx, instID, notify)
	}

	err = e.dataBus.PushQueueStream(ctx, ev)
	if err != nil {
		return instID, fmt.Errorf("push queue stream: %w", err)
	}

	return instID, nil
}

func (e *Engine) execInstance(ctx context.Context, inst *InstanceEvent) error {
	startEv := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
		Type:       "started",
		Time:       time.Now(),
	}

	err := e.dataBus.PushHistoryStream(ctx, startEv)
	if err != nil {
		return fmt.Errorf("push history start event, inst: %s: %w", inst.InstanceID, err)
	}

	endEv := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
	}
	ret, err := e.execJSScript(inst.InstanceID, inst.Script, inst.Mappings, inst.Fn, string(inst.Input))
	if err != nil {
		endEv.Type = "failed"
		endEv.Error = err.Error()
	} else {
		retBytes, mErr := json.Marshal(ret)
		if mErr != nil {
			endEv.Type = "failed"
			endEv.Error = fmt.Errorf("marshal result: %w", mErr).Error()
		} else {
			endEv.Type = "succeeded"
			endEv.Output = retBytes
		}
	}

	// simulate a job that takes some long time
	// time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	endEv.Time = time.Now()
	err = e.dataBus.PushHistoryStream(ctx, endEv)
	if err != nil {
		return fmt.Errorf("push history end event, inst: %s: %w", inst.InstanceID, err)
	}

	return nil
}

func (e *Engine) GetInstances(ctx context.Context, namespace string, limit int, offset int) ([]*InstanceStatus, int, error) {
	data, total := e.dataBus.FetchInstanceStatus(ctx, namespace, uuid.Nil, limit, offset)
	if len(data) == 0 {
		return nil, 0, ErrDataNotFound
	}

	return data, total, nil
}

func (e *Engine) GetInstanceByID(ctx context.Context, namespace string, id uuid.UUID) (*InstanceStatus, error) {
	data, _ := e.dataBus.FetchInstanceStatus(ctx, namespace, id, 0, 0)
	if len(data) == 0 {
		return nil, ErrDataNotFound
	}

	return data[0], nil
}

func (e *Engine) DeleteNamespace(ctx context.Context, name string) error {
	return e.dataBus.DeleteNamespace(ctx, name)
}
