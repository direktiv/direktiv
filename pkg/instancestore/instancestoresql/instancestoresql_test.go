package instancestoresql_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/instancestore/instancestoresql"
	"github.com/google/uuid"
)

func Test_NewSQLInstanceStore(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatal(err)
	}
	ns := uuid.New()
	server := uuid.New()

	store := instancestoresql.NewSQLInstanceStore(db)
	_, err = store.CreateInstanceData(context.Background(), &instancestore.CreateInstanceDataArgs{
		ID:             uuid.New(),
		NamespaceID:    ns,
		RootInstanceID: uuid.New(),
		Server:         server,
		Invoker:        "api",
		WorkflowPath:   "someRandomWfPath",
		Definition:     []byte{},
		DescentInfo:    []byte{},
		TelemetryInfo:  []byte{},
		RuntimeInfo:    []byte{},
		ChildrenInfo:   []byte{},
		Input:          []byte{},
		LiveData:       []byte{},
	})
	if err != nil {
		t.Error(err)
	}
	_, err = store.CreateInstanceData(context.Background(), &instancestore.CreateInstanceDataArgs{
		ID:             uuid.New(),
		NamespaceID:    ns,
		RootInstanceID: uuid.New(),
		Server:         server,
		Invoker:        "api",
		WorkflowPath:   "someRandomWfPathPlus",
		Definition:     []byte{},
		DescentInfo:    []byte{},
		TelemetryInfo:  []byte{},
		RuntimeInfo:    []byte{},
		ChildrenInfo:   []byte{},
		Input:          []byte{},
		LiveData:       []byte{},
	})
	if err != nil {
		t.Error(err)
	}
	_, err = store.CreateInstanceData(context.Background(), &instancestore.CreateInstanceDataArgs{
		ID:             uuid.New(),
		NamespaceID:    ns,
		RootInstanceID: uuid.New(),
		Server:         server,
		Invoker:        "api",
		WorkflowPath:   "-someRandomWfPath",
		Definition:     []byte{},
		DescentInfo:    []byte{},
		TelemetryInfo:  []byte{},
		RuntimeInfo:    []byte{},
		ChildrenInfo:   []byte{},
		Input:          []byte{},
		LiveData:       []byte{},
	})
	if err != nil {
		t.Error(err)
	}
	opts := &instancestore.ListOpts{
		Limit:  4,
		Offset: 0,
		Filters: []instancestore.Filter{
			{
				Field: instancestore.FieldWorkflowPath,
				Kind:  instancestore.FilterKindMatch,
				Value: "someRandomWfPath",
			},
		},
	}
	res, err := store.GetNamespaceInstances(context.Background(), ns, opts)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Total != 1 {
		t.Errorf("expected one instance as result but got %v", res.Total)
	}
	if len(res.Results) != res.Total {
		t.Errorf("results entries differs from res total")
	}
}
