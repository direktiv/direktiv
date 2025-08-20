package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/google/uuid"
)

type engine struct {
	db *database.DB
}

func NewEngine(db *database.DB) (core.JSEngine, error) {
	return &engine{
		db: db,
	}, nil
}

func (e *engine) Run(circuit *core.Circuit) error {
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
		return uuid.Nil, err
	}

	ret, err := e.execJSScript(fileData, input)
	patch := map[string]any{
		"ended_at": true,
	}
	if err != nil {
		patch["status"] = 2
		patch["error"] = []byte(err.Error())
	} else {
		retBytes, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		patch["status"] = 3
		patch["output"] = retBytes
	}

	db, err := e.db.BeginTx(ctx)
	if err != nil {
		return id, err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	err = dStore.JSInstances().Patch(ctx, id, patch)
	if err != nil {
		return id, err
	}
	err = db.Commit(ctx)

	return id, err
}

func (e *engine) createWorkflowInstance(ctx context.Context, namespace string, path string, input string) (uuid.UUID, []byte, error) {
	db, err := e.db.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	defer db.Rollback()
	dStore := db.DataStore()
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
	err = dStore.JSInstances().Create(ctx, &datastore.JSInstance{
		ID:           id,
		Namespace:    namespace,
		WorkflowPath: path,
		WorkflowData: string(fileData),
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
