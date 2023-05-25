package instancestore_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore/instancestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

func checksum(data []byte) string {
	hasher := sha256.New()
	x := hasher.Sum(data)

	return base64.StdEncoding.EncodeToString(x)
}

func isIdenticalBytes(a, b []byte) bool {
	return checksum(a) == checksum(b)
}

type assertInstanceStoreCorrectInstanceDataCreationTest struct {
	name string
	args *instancestore.CreateInstanceDataArgs
}

func assertInstanceStoreCorrectInstanceDataCreation(t *testing.T, is instancestore.Store, args *instancestore.CreateInstanceDataArgs) {
	t.Helper()

	idata, err := is.CreateInstanceData(context.Background(), args)
	if err != nil {
		t.Errorf("unexpected CreateInstanceData() error: %v", err)

		return
	}
	if idata == nil {
		t.Errorf("unexpected nil idata CreateInstanceData()")

		return
	}

	// validation
	if idata.ID != args.ID {
		t.Errorf("unexpected idata.ID, got: >%s<, want: >%s<", idata.ID, args.ID)

		return
	}

	if idata.NamespaceID != args.NamespaceID {
		t.Errorf("unexpected idata.NamespaceID, got: >%s<, want: >%s<", idata.NamespaceID, args.NamespaceID)

		return
	}

	if idata.WorkflowID != args.WorkflowID {
		t.Errorf("unexpected idata.WorkflowID, got: >%s<, want: >%s<", idata.WorkflowID, args.WorkflowID)

		return
	}

	if idata.RevisionID != args.RevisionID {
		t.Errorf("unexpected idata.RevisionID, got: >%s<, want: >%s<", idata.RevisionID, args.RevisionID)

		return
	}

	if idata.RootInstanceID != args.RootInstanceID {
		t.Errorf("unexpected idata.RootInstanceID, got: >%s<, want: >%s<", idata.RootInstanceID, args.RootInstanceID)

		return
	}

	if idata.CalledAs != args.CalledAs {
		t.Errorf("unexpected idata.CalledAs, got: >%s<, want: >%s<", idata.CalledAs, args.CalledAs)

		return
	}

	if idata.Invoker != args.Invoker {
		t.Errorf("unexpected idata.Invoker, got: >%s<, want: >%s<", idata.Invoker, args.Invoker)

		return
	}

	if !isIdenticalBytes(args.Definition, idata.Definition) {
		t.Errorf("unexpected idata.Definition, got: >%v<, want: >%v<", idata.Definition, args.Definition)

		return
	}

	if !isIdenticalBytes(args.Settings, idata.Settings) {
		t.Errorf("unexpected idata.Settings, got: >%v<, want: >%v<", idata.Settings, args.Settings)

		return
	}

	if !isIdenticalBytes(args.DescentInfo, idata.DescentInfo) {
		t.Errorf("unexpected idata.DescentInfo, got: >%v<, want: >%v<", idata.DescentInfo, args.DescentInfo)

		return
	}

	if !isIdenticalBytes(args.TelemetryInfo, idata.TelemetryInfo) {
		t.Errorf("unexpected idata.TelemetryInfo, got: >%v<, want: >%v<", idata.TelemetryInfo, args.TelemetryInfo)

		return
	}

	if !isIdenticalBytes(args.Input, idata.Input) {
		t.Errorf("unexpected idata.Input, got: >%v<, want: >%v<", idata.Input, args.Input)

		return
	}

	tf := time.Now()
	t0 := tf.Add(time.Second * -1)

	if !idata.CreatedAt.Before(tf) || !idata.CreatedAt.After(t0) {
		t.Errorf("unexpected idata.CreatedAt, got: >%s<, want something closer to: >%s<", idata.CreatedAt.String(), tf.String())

		return
	}

	if !idata.UpdatedAt.Before(tf) || !idata.UpdatedAt.After(t0) {
		t.Errorf("unexpected idata.UpdatedAt, got: >%s<, want something closer to: >%s<", idata.UpdatedAt.String(), tf.String())

		return
	}

	if idata.EndedAt != nil {
		t.Errorf("unexpected idata.EndedAt, got: >%s<, want: >%s<", idata.EndedAt.String(), "nil")

		return
	}

	if idata.Status != util.InstanceStatusPending {
		t.Errorf("unexpected idata.Status, got: >%s<, want: >%s<", idata.Status, util.InstanceStatusPending)

		return
	}

	if idata.ErrorCode != "" {
		t.Errorf("unexpected idata.ErrorCode, got: >%s<, want: >%s<", idata.ErrorCode, "")

		return
	}

	if idata.Deadline != nil {
		t.Errorf("unexpected idata.Deadline, got: >%s<, want: >%s<", idata.Deadline.String(), "nil")

		return
	}

	expect := []byte(`{}`)
	if !isIdenticalBytes(expect, idata.RuntimeInfo) {
		t.Errorf("unexpected idata.RuntimeInfo, got: >%v<, want: >%v<", idata.RuntimeInfo, expect)

		return
	}

	if !isIdenticalBytes(expect, idata.ChildrenInfo) {
		t.Errorf("unexpected idata.ChildrenInfo, got: >%v<, want: >%v<", idata.ChildrenInfo, expect)

		return
	}

	if !isIdenticalBytes(expect, idata.LiveData) {
		t.Errorf("unexpected idata.LiveData, got: >%v<, want: >%v<", idata.LiveData, expect)

		return
	}

	expect = []byte(``)
	if !isIdenticalBytes(expect, idata.StateMemory) {
		t.Errorf("unexpected idata.StateMemory, got: >%v<, want: >%v<", idata.StateMemory, expect)

		return
	}

	if idata.ErrorMessage != nil {
		t.Errorf("unexpected idata.ErrorMessage, got: >%s<, want: >%s<", string(idata.ErrorMessage), "nil")

		return
	}

	if idata.Output != nil {
		t.Errorf("unexpected idata.Output, got: >%v<, want: >%s<", string(idata.Output), "nil")

		return
	}

	if idata.Metadata != nil {
		t.Errorf("unexpected idata.Metadata, got: >%v<, want: >%s<", string(idata.Metadata), "nil")

		return
	}
}

func Test_sqlInstanceStore_CreateInstanceData(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	var tests []assertInstanceStoreCorrectInstanceDataCreationTest

	id := uuid.New()
	tests = append(tests, assertInstanceStoreCorrectInstanceDataCreationTest{
		name: "validCase",
		args: &instancestore.CreateInstanceDataArgs{
			ID:             id,
			NamespaceID:    uuid.New(),
			WorkflowID:     uuid.New(),
			RevisionID:     uuid.New(),
			RootInstanceID: id,
			Invoker:        "api",
			CalledAs:       "/test.yaml",
			Definition: []byte(`
states:
- id: test
  type: noop
`),
			Input:         []byte(`{}`),
			TelemetryInfo: []byte(`{}`),
			Settings:      []byte(`{}`),
			DescentInfo:   []byte(`{}`),
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertInstanceStoreCorrectInstanceDataCreation(t, instances, tt.args)
		})
	}
}

func assertInstanceDataIsEverything(t *testing.T, idata *instancestore.InstanceData) {
	if idata.Definition == nil {
		t.Errorf("missing idata.Definition")

		return
	}

	if idata.Settings == nil {
		t.Errorf("missing idata.Settings")

		return
	}

	if idata.DescentInfo == nil {
		t.Errorf("missing idata.DescentInfo")

		return
	}

	if idata.TelemetryInfo == nil {
		t.Errorf("missing idata.TelemetryInfo")

		return
	}

	if idata.Input == nil {
		t.Errorf("missing idata.Input")

		return
	}
}

func assertInstanceDataIsSummary(t *testing.T, idata *instancestore.InstanceData) {
	if idata.Definition != nil {
		t.Errorf("unexpected idata.Definition")

		return
	}

	if idata.Settings != nil {
		t.Errorf("unexpected idata.Settings")

		return
	}

	if idata.DescentInfo != nil {
		t.Errorf("unexpected idata.DescentInfo")

		return
	}

	if idata.TelemetryInfo != nil {
		t.Errorf("unexpected idata.TelemetryInfo")

		return
	}

	if idata.Input != nil {
		t.Errorf("unexpected idata.Input")

		return
	}

	if idata.RuntimeInfo != nil {
		t.Errorf("unexpected idata.RuntimeInfo")

		return
	}

	if idata.ChildrenInfo != nil {
		t.Errorf("unexpected idata.ChildrenInfo")

		return
	}

	if idata.LiveData != nil {
		t.Errorf("unexpected idata.LiveData")

		return
	}

	if idata.StateMemory != nil {
		t.Errorf("unexpected idata.StateMemory")

		return
	}

	if idata.ErrorMessage != nil {
		t.Errorf("unexpected idata.ErrorMessage")

		return
	}

	if idata.Output != nil {
		t.Errorf("unexpected idata.Output")

		return
	}

	if idata.Metadata != nil {
		t.Errorf("unexpected idata.Metadata")

		return
	}
}

type assertInstanceStoreCorrectGetNamespaceInstancesTest struct {
	name string
	args *instancestore.CreateInstanceDataArgs
	nsID uuid.UUID
	ids  []uuid.UUID
}

func assertInstanceStoreCorrectGetNamespaceInstances(t *testing.T, is instancestore.Store, args *instancestore.CreateInstanceDataArgs, nsID uuid.UUID, ids []uuid.UUID) {
	t.Helper()

	args.NamespaceID = nsID
	for _, id := range ids {
		args.ID = id
		assertInstanceStoreCorrectInstanceDataCreation(t, is, args)
		if t.Failed() {
			return
		}
	}

	idatas, err := is.GetNamespaceInstances(context.Background(), nsID)
	if err != nil {
		t.Errorf("unexpected GetNamespaceInstances() error: %v", err)

		return
	}
	if idatas == nil {
		idatas = make([]*instancestore.InstanceData, 0)
	}

	// validation
	if len(idatas) != len(ids) {
		t.Errorf("unexpected results count, got: %d, want: %d", len(idatas), len(ids))

		return
	}

	for idx, idata := range idatas {
		if idata.ID != ids[idx] {
			t.Errorf("unexpected idata.ID, got: >%s<, want: >%s<", idata.ID, ids[idx])

			return
		}

		assertInstanceDataIsSummary(t, idata)
		if t.Failed() {
			return
		}
	}

}

func Test_sqlInstanceStore_GetNamespaceInstances(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	var tests []assertInstanceStoreCorrectGetNamespaceInstancesTest

	args := &instancestore.CreateInstanceDataArgs{
		WorkflowID: uuid.New(),
		RevisionID: uuid.New(),
		Invoker:    "api",
		CalledAs:   "/test.yaml",
		Definition: []byte(`
states:
- id: test
type: noop
`),
		Input:         []byte(`{}`),
		TelemetryInfo: []byte(`{}`),
		Settings:      []byte(`{}`),
		DescentInfo:   []byte(`{}`),
	}

	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: uuid.New(),
		ids:  []uuid.UUID{},
	})

	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: uuid.New(),
		ids:  []uuid.UUID{uuid.New()},
	})

	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: uuid.New(),
		ids:  []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
	})

	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: uuid.New(),
		ids:  []uuid.UUID{},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertInstanceStoreCorrectGetNamespaceInstances(t, instances, tt.args, tt.nsID, tt.ids)
		})
	}
}
