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

	err = e.startWorkers(lc)
	if err != nil {
		return fmt.Errorf("start workers: %w", err)
	}

	return nil
}

func (e *Engine) ExecWorkflow(ctx context.Context, namespace string, workflowPath string, args any, metadata map[string]string) (uuid.UUID, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath)
	if err != nil {
		return uuid.Nil, fmt.Errorf("fetch script: %w", err)
	}

	return e.ExecScript(ctx, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, args, metadata)
}

func (e *Engine) ExecScript(ctx context.Context, namespace string, script string, mappings string, fn string, args any, metadata map[string]string) (uuid.UUID, error) {
	input, ok := args.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid input")
	}
	if input == "" {
		input = "{}"
	}

	instID := uuid.New()

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
	err := e.dataBus.PushInstanceEvent(ctx, ev)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create workflow instance: %w", err)
	}
	err = e.dataBus.PushInstanceFoo(ctx, ev)
	if err != nil {
		return instID, fmt.Errorf("create workflow instance: %w", err)
	}

	return instID, nil
}

func (e *Engine) ExecInstance(ctx context.Context, inst *InstanceEvent) error {
	err := e.dataBus.PushInstanceEvent(ctx, &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
		Type:       "started",
		Time:       time.Now(),
	})
	if err != nil {
		return fmt.Errorf("put started instance event: %w", err)
	}

	endMsg := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: inst.InstanceID,
		Namespace:  inst.Namespace,
		Time:       time.Now(),
	}
	ret, err := e.execJSScript(inst.InstanceID, inst.Script, inst.Mappings, inst.Fn, string(inst.Input))
	if err != nil {
		endMsg.Type = "failed"
		endMsg.Error = err.Error()
	} else {
		retBytes, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		endMsg.Type = "succeeded"
		endMsg.Output = retBytes
	}

	err = e.dataBus.PushInstanceEvent(ctx, endMsg)
	if err != nil {
		return fmt.Errorf("put end instance event: %w", err)
	}

	return nil
}

func (e *Engine) GetInstances(ctx context.Context, namespace string) ([]*InstanceStatus, error) {
	data := e.dataBus.QueryInstanceStatus(ctx, namespace, uuid.Nil)
	if len(data) == 0 {
		return nil, ErrDataNotFound
	}

	return data, nil
}

func (e *Engine) GetInstanceByID(ctx context.Context, namespace string, id uuid.UUID) (*InstanceStatus, error) {
	data := e.dataBus.QueryInstanceStatus(ctx, namespace, id)
	if len(data) == 0 {
		return nil, ErrDataNotFound
	}

	return data[0], nil
}
