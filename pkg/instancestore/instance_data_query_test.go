package instancestore_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/google/uuid"
)

func assertInstanceStoreCorrectGetMost(t *testing.T, is instancestore.Store, args *instancestore.CreateInstanceDataArgs) {
	t.Helper()

	assertInstanceStoreCorrectInstanceDataCreation(t, is, args)
	if t.Failed() {
		return
	}

	idata, err := is.ForInstanceID(args.ID).GetMost(context.Background())
	if err != nil {
		t.Errorf("unexpected GetSummary() error: %v", err)

		return
	}
	if idata == nil {
		t.Errorf("unexpected nil idata GetSummary()")

		return
	}

	// validation
	assertInstanceDataIsMost(t, idata)
}

// nolint
func TestInstanceDataQuery_sqlInstanceStore_GetMost(t *testing.T) {
	server := uuid.New()

	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	instances := db.InstanceStore()

	var tests []assertInstanceStoreCorrectInstanceDataCreationTest

	id := uuid.New()
	tests = append(tests, assertInstanceStoreCorrectInstanceDataCreationTest{
		name: "validCase",
		args: &instancestore.CreateInstanceDataArgs{
			ID:             id,
			NamespaceID:    ns.ID,
			RootInstanceID: id,
			Server:         server,
			Invoker:        "api",
			WorkflowPath:   "/test.yaml",
			Definition: []byte(`
states:
- id: test
  type: noop
`),
			Input:         []byte(`{}`),
			TelemetryInfo: []byte(`{}`),
			DescentInfo:   []byte(`{}`),
			RuntimeInfo:   []byte(`{}`),
			ChildrenInfo:  []byte(`{}`),
			LiveData:      []byte(`{}`),
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertInstanceStoreCorrectGetMost(t, instances, tt.args)
		})
	}
}

func assertInstanceStoreCorrectGetSummary(t *testing.T, is instancestore.Store, args *instancestore.CreateInstanceDataArgs) {
	t.Helper()

	assertInstanceStoreCorrectInstanceDataCreation(t, is, args)
	if t.Failed() {
		return
	}

	idata, err := is.ForInstanceID(args.ID).GetSummary(context.Background())
	if err != nil {
		t.Errorf("unexpected GetSummary() error: %v", err)

		return
	}
	if idata == nil {
		t.Errorf("unexpected nil idata GetSummary()")

		return
	}

	if idata.InputLength != len(args.Input) {
		t.Errorf("incorrect input length field")
	}

	// validation
	assertInstanceDataIsSummary(t, idata)
}

// nolint
func TestInstanceDataQuery_sqlInstanceStore_GetSummary(t *testing.T) {
	server := uuid.New()

	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	instances := db.InstanceStore()

	var tests []assertInstanceStoreCorrectInstanceDataCreationTest

	id := uuid.New()
	tests = append(tests, assertInstanceStoreCorrectInstanceDataCreationTest{
		name: "validCase",
		args: &instancestore.CreateInstanceDataArgs{
			ID:             id,
			NamespaceID:    ns.ID,
			RootInstanceID: id,
			Server:         server,
			Invoker:        "api",
			WorkflowPath:   "/test.yaml",
			Definition: []byte(`
states:
- id: test
  type: noop
`),
			Input:         []byte(`{}`),
			TelemetryInfo: []byte(`{}`),
			DescentInfo:   []byte(`{}`),
			RuntimeInfo:   []byte(`{}`),
			ChildrenInfo:  []byte(`{}`),
			LiveData:      []byte(`{}`),
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertInstanceStoreCorrectGetSummary(t, instances, tt.args)
		})
	}
}
