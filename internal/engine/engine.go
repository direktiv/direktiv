package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Engine struct {
	db        *gorm.DB
	projector Projector
	dataBus   DataBus
}

func (e *Engine) ListInstances(ctx context.Context, namespace string) ([]uuid.UUID, error) {
	// TODO implement me
	panic("implement me")
}

func NewEngine(db *gorm.DB, proj Projector, bus DataBus) (*Engine, error) {
	return &Engine{
		db:        db,
		projector: proj,
		dataBus:   bus,
	}, nil
}

func (e *Engine) Start(lc *lifecycle.Manager) error {
	err := e.projector.Start(lc)
	if err != nil {
		return fmt.Errorf("start projector: %w", err)
	}
	err = e.dataBus.Start(lc)
	if err != nil {
		return fmt.Errorf("start databus: %w", err)
	}

	cycleTime := time.Second

	lc.Go(func() error {
		for {
			select {
			case <-lc.Done():
				return nil
			case <-time.Tick(cycleTime):
				//TODO: implement me
			}
		}
	})

	return nil
}

func (e *Engine) ExecWorkflow(ctx context.Context, namespace string, script string, fn string, args any, labels map[string]string) (uuid.UUID, error) {
	input, ok := args.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid input")
	}
	if input == "" {
		input = "{}"
	}

	instID := uuid.New()

	err := e.dataBus.PushInstanceEvent(ctx, &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Type:       "init",
		Time:       time.Now(),

		Script: script,
		Input:  json.RawMessage(input),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create workflow instance: %w", err)
	}

	endMsg := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Time:       time.Now(),
	}
	ret, err := e.execJSScript([]byte(script), fn, input)
	if err != nil {
		endMsg.Type = "fail"
		endMsg.Error = err.Error()

	} else {
		retBytes, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		endMsg.Type = "success"
		endMsg.Output = retBytes
	}

	err = e.dataBus.PushInstanceEvent(ctx, endMsg)
	if err != nil {
		return instID, fmt.Errorf("put end instance message: %w", err)
	}

	return instID, nil
}

func (e *Engine) GetInstances(ctx context.Context, namespace string) []InstanceStatus {
	return e.dataBus.QueryInstanceStatus(ctx, namespace, uuid.Nil)
}

func (e *Engine) GetInstanceByID(ctx context.Context, namespace string, id uuid.UUID) []InstanceStatus {
	return e.dataBus.QueryInstanceStatus(ctx, namespace, id)
}
