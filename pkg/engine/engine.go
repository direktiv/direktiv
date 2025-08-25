package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
)

type engine struct {
	db    *database.DB
	store Store
}

func NewEngine(db *database.DB, store Store) (core.Engine, error) {
	return &engine{
		db:    db,
		store: store,
	}, nil
}

func (e *engine) Start(circuit *core.Circuit) error {
	cycleTime := time.Second
	for {
		if circuit.IsDone() {
			return nil
		}
		// TODO: implement async engine exec of workflows and retries.
		time.Sleep(cycleTime)
	}
}

func (e *engine) ExecWorkflow(ctx context.Context, namespace string, path string, input string) (uuid.UUID, error) {
	if input == "" {
		input = "{}"
	}
	id, fileData, err := e.createWorkflowInstance(ctx, namespace, path, input)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create workflow instance: %s", err)
	}

	ret, err := e.execJSScript(fileData, input)
	endMsg := InstanceMessage{
		InstanceID:   id,
		Namespace:    namespace,
		WorkflowPath: path,
		Status:       0,
		EndedAt:      time.Now(),
		Memory:       sql.NullString{},
		Output:       sql.NullString{},
		Error:        sql.NullString{},
	}
	if err != nil {
		endMsg.Status = 2
		endMsg.Error = sql.NullString{Valid: true, String: err.Error()}
		endMsg.EndedAt = time.Now()
	} else {
		retBytes, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		endMsg.Status = 3
		endMsg.Output = sql.NullString{Valid: true, String: string(retBytes)}
		endMsg.EndedAt = time.Now()
	}

	_, err = e.store.PutInstanceMessage(ctx, namespace, id, "stateChange", endMsg)
	if err != nil {
		return id, fmt.Errorf("put end instance message: %s", err)
	}

	return id, err
}

// TODO: fix this api
func (e *engine) GetInstance(ctx context.Context, namespace string, instanceID uuid.UUID) (any, error) {
	return e.store.GetInstanceMessages(ctx, namespace, instanceID, "stateChange")
}

func (e *engine) createWorkflowInstance(ctx context.Context, namespace string, path string, input string) (uuid.UUID, []byte, error) {
	db, err := e.db.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	defer db.Rollback()
	fStore := db.FileStore()

	file, err := fStore.ForNamespace(namespace).GetFile(ctx, path)
	if err != nil {
		return uuid.Nil, nil, err
	}
	fileData, err := fStore.ForFile(file).GetData(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	id := uuid.New()

	_, err = e.store.PutInstanceMessage(ctx, namespace, id, "stateChange", InstanceMessage{
		InstanceID:   id,
		Namespace:    namespace,
		WorkflowPath: path,
		WorkflowText: string(fileData),
		Status:       0,
		Input:        sql.NullString{String: input, Valid: true},
		Memory:       sql.NullString{},
		Output:       sql.NullString{},
		Error:        sql.NullString{},
	})
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = db.Commit(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return id, fileData, nil
}
