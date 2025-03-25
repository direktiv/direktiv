package instancestoresql_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/google/uuid"
)

func Test_NewSQLInstanceStore(t *testing.T) {
	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	server := uuid.New()

	store := db.InstanceStore()

	telemetryInfo := &engine.InstanceTelemetryInfo{
		Version:     "v1",
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	}

	telemetryInfoBytes, err := json.Marshal(telemetryInfo)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateInstanceData(context.Background(), &instancestore.CreateInstanceDataArgs{
		ID:             uuid.New(),
		NamespaceID:    ns.ID,
		RootInstanceID: uuid.New(),
		Server:         server,
		Invoker:        "api",
		WorkflowPath:   "someRandomWfPath",
		Definition:     []byte{},
		DescentInfo:    []byte{},
		TelemetryInfo:  telemetryInfoBytes,
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
	res, err := store.GetNamespaceInstances(context.Background(), ns.ID, opts)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Total != 1 {
		t.Errorf("expected one instance as result but got %v", res.Total)
	}
	if len(res.Results) != res.Total {
		t.Errorf("results entries differ from res total")
	}

	if len(res.Results) > 0 {
		var storedTelemetry engine.InstanceTelemetryInfo
		err := json.Unmarshal(res.Results[0].TelemetryInfo, &storedTelemetry)
		if err != nil {
			t.Errorf("failed to unmarshal telemetry info: %v", err)
		}
		if !reflect.DeepEqual(storedTelemetry, telemetryInfo) {
			t.Errorf("telemetry info mismatch: got %+v, want %+v", storedTelemetry, telemetryInfo)
		}
	}
}
