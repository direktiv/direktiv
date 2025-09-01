package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
)

type store interface {
	PushInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error)
	PullInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]core.EngineMessage, error)
}

type Engine struct {
	db    *database.DB
	store store
}

func (e *Engine) ListInstances(ctx context.Context, namespace string) ([]uuid.UUID, error) {
	// TODO implement me
	panic("implement me")
}

func NewEngine(db *database.DB, store store) (*Engine, error) {
	return &Engine{
		db:    db,
		store: store,
	}, nil
}

func (e *Engine) Start(circuit *core.Circuit) error {
	cycleTime := time.Second
	for {
		if circuit.IsDone() {
			return nil
		}
		// TODO: implement async engine exec of workflows and retries.
		time.Sleep(cycleTime)
	}
}

func (e *Engine) ExecWorkflow(ctx context.Context, namespace string, script string, fn string, args any, labels map[string]string) (uuid.UUID, error) {
	input, ok := args.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid input")
	}
	if input == "" {
		input = "{}"
	}

	id := uuid.New()

	_, err := e.store.PushInstanceMessage(ctx, namespace, id, "init", core.InstanceMessage{
		InstanceID: id,
		Namespace:  namespace,
		Script:     script,
		Labels:     labels,
		Status:     0,
		Input:      json.RawMessage(input),
		Memory:     nil,
		Output:     nil,
		Error:      "",
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create workflow instance: %w", err)
	}

	ret, err := e.execJSScript([]byte(script), fn, input)
	endMsg := core.InstanceMessage{
		InstanceID: id,
		Namespace:  namespace,
		Script:     script,
		Labels:     labels,
		Status:     0,
		EndedAt:    time.Now(),
		Memory:     nil,
		Output:     nil,
		Error:      "",
	}
	if err != nil {
		endMsg.Status = 2
		endMsg.Error = err.Error()
		endMsg.EndedAt = time.Now()
	} else {
		retBytes, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		endMsg.Status = 3
		endMsg.Output = retBytes
		endMsg.EndedAt = time.Now()
	}

	_, err = e.store.PushInstanceMessage(ctx, namespace, id, "end", endMsg)
	if err != nil {
		return id, fmt.Errorf("put end instance message: %w", err)
	}

	return id, err
}

func (e *Engine) GetInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID) ([]core.EngineMessage, error) {
	return e.store.PullInstanceMessages(ctx, namespace, instanceID, "*")
}
