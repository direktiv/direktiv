package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrDataNotFound = fmt.Errorf("data not found")

type Engine struct {
	db        *gorm.DB
	projector Projector
	dataBus   DataBus
}

type fetchScript func(ctx context.Context, namespace, path string)

func (e *Engine) ListInstances(ctx context.Context, namespace string) ([]uuid.UUID, error) {
	// TODO implement me
	panic("implement me")
}

func NewEngine(db *gorm.DB, proj Projector, bus DataBus, compiler core.Compiler) (*Engine, error) {
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
				// TODO: implement me
			}
		}
	})

	return nil
}

func (e *Engine) ExecWorkflow(ctx context.Context, namespace string, script string, fn string, args any, metadata map[string]string) (uuid.UUID, error) {
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
		Namespace:  namespace,
		Type:       "started",
		Time:       time.Now(),
		Metadata:   metadata,

		Script: script,
		Input:  json.RawMessage(input),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create workflow instance: %w", err)
	}

	endMsg := &InstanceEvent{
		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  namespace,
		Time:       time.Now(),
	}
	ret, err := e.execJSScript([]byte(script), nil, fn, input)
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
		return instID, fmt.Errorf("put end instance message: %w", err)
	}

	return instID, nil
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
