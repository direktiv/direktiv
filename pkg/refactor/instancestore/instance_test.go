package instancestore_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore/instancestoresql"
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

	if idata.WorkflowPath != args.WorkflowPath {
		t.Errorf("unexpected idata.WorkflowPath, got: >%s<, want: >%s<", idata.WorkflowPath, args.WorkflowPath)

		return
	}

	if idata.RootInstanceID != args.RootInstanceID {
		t.Errorf("unexpected idata.RootInstanceID, got: >%s<, want: >%s<", idata.RootInstanceID, args.RootInstanceID)

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

	// NOTE: disabled these tests because supporting them would cost us performance for little benefit.
	//
	// tf := time.Now()
	// t0 := tf.Add(time.Second * -1)
	//
	// if !idata.CreatedAt.Before(tf) || !idata.CreatedAt.After(t0) {
	// 	t.Errorf("unexpected idata.CreatedAt, got: >%s<, want something closer to: >%s<", idata.CreatedAt.String(), tf.String())
	//
	// 	return
	// }
	//
	// if !idata.UpdatedAt.Before(tf) || !idata.UpdatedAt.After(t0) {
	// 	t.Errorf("unexpected idata.UpdatedAt, got: >%s<, want something closer to: >%s<", idata.UpdatedAt.String(), tf.String())
	//
	// 	return
	// }

	if idata.EndedAt != nil {
		t.Errorf("unexpected idata.EndedAt, got: >%s<, want: >%s<", idata.EndedAt.String(), "nil")

		return
	}

	if idata.Status != instancestore.InstanceStatusPending {
		t.Errorf("unexpected idata.Status, got: >%s<, want: >%v<", idata.Status, instancestore.InstanceStatusPending)

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

	// expect = []byte(``)
	// if !isIdenticalBytes(expect, idata.StateMemory) {
	// 	t.Errorf("unexpected idata.StateMemory, got: >%v<, want: >%v<", idata.StateMemory, expect)
	//
	// 	return
	// }

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

// nolint
func Test_sqlInstanceStore_CreateInstanceData(t *testing.T) {
	server := uuid.New()

	db, err := database.NewMockGorm()
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
			assertInstanceStoreCorrectInstanceDataCreation(t, instances, tt.args)
		})
	}
}

func assertInstanceDataIsMost(t *testing.T, idata *instancestore.InstanceData) {
	t.Helper()
	if idata.Definition == nil {
		t.Errorf("missing idata.Definition")

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
}

func assertInstanceDataIsSummary(t *testing.T, idata *instancestore.InstanceData) {
	t.Helper()
	// if idata.Definition != nil {
	// 	t.Errorf("unexpected idata.Definition")
	//
	// 	return
	// }

	// if idata.Settings != nil {
	// 	t.Errorf("unexpected idata.Settings")
	//
	// 	return
	// }

	// if idata.DescentInfo != nil {
	// 	t.Errorf("unexpected idata.DescentInfo")
	//
	// 	return
	// }

	// if idata.TelemetryInfo != nil {
	// 	t.Errorf("unexpected idata.TelemetryInfo")
	//
	// 	return
	// }

	if idata.Input != nil {
		t.Errorf("unexpected idata.Input")

		return
	}

	// if idata.RuntimeInfo != nil {
	// 	t.Errorf("unexpected idata.RuntimeInfo")
	//
	// 	return
	// }

	// if idata.ChildrenInfo != nil {
	// 	t.Errorf("unexpected idata.ChildrenInfo")
	//
	// 	return
	// }

	if idata.LiveData != nil {
		t.Errorf("unexpected idata.LiveData")

		return
	}

	if idata.StateMemory != nil {
		t.Errorf("unexpected idata.StateMemory")

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

	results, err := is.GetNamespaceInstances(context.Background(), nsID, nil)
	if err != nil {
		t.Errorf("unexpected GetNamespaceInstances() error: %v", err)

		return
	}
	if results.Results == nil {
		results.Results = make([]instancestore.InstanceData, 0)
	}

	// validation
	if results.Total < len(results.Results) {
		t.Errorf("illogical rowsAffected value, got: %d, want: %d", results.Total, len(results.Results))

		return
	}

	if len(results.Results) != len(ids) {
		t.Errorf("unexpected results count, got: %d, want: %d", len(results.Results), len(ids))

		return
	}

	for idx, idata := range results.Results {
		if idata.ID != ids[idx] {
			t.Errorf("unexpected idata.ID, got: >%s<, want: >%s<", idata.ID, ids[idx])

			return
		}

		assertInstanceDataIsSummary(t, &results.Results[idx])
		if t.Failed() {
			return
		}
	}
}

// nolint
func Test_sqlInstanceStore_GetNamespaceInstances(t *testing.T) {
	server := uuid.New()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	var tests []assertInstanceStoreCorrectGetNamespaceInstancesTest

	args := &instancestore.CreateInstanceDataArgs{
		ID:           uuid.New(),
		Server:       server,
		Invoker:      "api",
		WorkflowPath: "/test.yaml",
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

	nsID := uuid.New()
	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: nsID,
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

	res, err := instances.GetNamespaceInstances(context.Background(), nsID, nil)
	if err != nil {
		t.Errorf("unexpected GetNamespaceInstances() error: %v", err)

		return
	}

	if len(res.Results) != res.Total || res.Total != 3 {
		t.Errorf("unexpected GetNamespaceInstances() results: %+v", err)

		return
	}

	t0 := res.Results[0].CreatedAt
	for i := 1; i < res.Total; i++ {
		tx := res.Results[i].CreatedAt
		if tx.After(t0) {
			t.Errorf("GetNamespaceInstances() results are improperly sorted: %+v", err)

			return
		}
		t0 = tx
	}

	idata := res.Results[1]

	res, err = instances.GetNamespaceInstances(context.Background(), nsID, &instancestore.ListOpts{
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Errorf("unexpected GetNamespaceInstances() error: %v", err)

		return
	}

	if len(res.Results) != 1 || res.Total != 3 || idata.ID != res.Results[0].ID {
		t.Errorf("unexpected GetNamespaceInstances() results: %+v", err)

		return
	}
}

// nolint
func Test_sqlInstanceStore_GetHangingInstances(t *testing.T) {
	server := uuid.New()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	idatas, err := instances.GetHangingInstances(context.Background())
	if err != nil {
		t.Errorf("unexpected GetHangingInstances() error: %v", err)

		return
	}

	if len(idatas) > 0 {
		t.Errorf("unexpected hanging instances: got >%+v<", idatas)

		return
	}

	id := uuid.New()

	args := &instancestore.CreateInstanceDataArgs{
		ID:           id,
		Server:       server,
		Invoker:      instancestore.InvokerCron,
		WorkflowPath: "/test.yaml",
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
	}

	assertInstanceStoreCorrectInstanceDataCreation(t, instances, args)

	idatas, err = instances.GetHangingInstances(context.Background())
	if err != nil {
		t.Errorf("unexpected GetHangingInstances() error: %v", err)

		return
	}

	if len(idatas) > 0 {
		t.Errorf("unexpected hanging instances: got >%+v<", idatas)

		return
	}

	tf := time.Now().Add(30 * time.Second)
	err = instances.ForInstanceID(id).UpdateInstanceData(context.Background(), &instancestore.UpdateInstanceDataArgs{
		Server:   server,
		Deadline: &tf,
	})
	if err != nil {
		t.Errorf("unexpected UpdateInstanceData() error: %v", err)

		return
	}

	idatas, err = instances.GetHangingInstances(context.Background())
	if err != nil {
		t.Errorf("unexpected GetHangingInstances() error: %v", err)

		return
	}

	if len(idatas) > 0 {
		t.Errorf("unexpected hanging instances: got >%+v<", idatas)

		return
	}

	tf = time.Now().Add(-30 * time.Second)
	status := instancestore.InstanceStatusComplete
	err = instances.ForInstanceID(id).UpdateInstanceData(context.Background(), &instancestore.UpdateInstanceDataArgs{
		Server:   server,
		Deadline: &tf,
		Status:   &status,
	})
	if err != nil {
		t.Errorf("unexpected UpdateInstanceData() error: %v", err)

		return
	}

	idatas, err = instances.GetHangingInstances(context.Background())
	if err != nil {
		t.Errorf("unexpected GetHangingInstances() error: %v", err)

		return
	}

	if len(idatas) > 0 {
		t.Errorf("unexpected hanging instances: got >%+v<", idatas)

		return
	}

	tf = time.Now().Add(-30 * time.Second)
	status = instancestore.InstanceStatusPending
	err = instances.ForInstanceID(id).UpdateInstanceData(context.Background(), &instancestore.UpdateInstanceDataArgs{
		Server:   server,
		Deadline: &tf,
		Status:   &status,
	})
	if err != nil {
		t.Errorf("unexpected UpdateInstanceData() error: %v", err)

		return
	}

	idatas, err = instances.GetHangingInstances(context.Background())
	if err != nil {
		t.Errorf("unexpected GetHangingInstances() error: %v", err)

		return
	}

	if len(idatas) == 0 {
		t.Errorf("expected results but got none")

		return
	}
}

// nolint
func Test_sqlInstanceStore_DeleteOldInstances(t *testing.T) {
	server := uuid.New()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	err = instances.DeleteOldInstances(context.Background(), time.Now())
	if err != nil {
		t.Errorf("unexpected DeleteOldInstances() error: %v", err)

		return
	}

	id := uuid.New()

	args := &instancestore.CreateInstanceDataArgs{
		ID:           id,
		Server:       server,
		Invoker:      instancestore.InvokerCron,
		WorkflowPath: "/test.yaml",
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
	}

	assertInstanceStoreCorrectInstanceDataCreation(t, instances, args)

	err = instances.DeleteOldInstances(context.Background(), time.Now())
	if err != nil {
		t.Errorf("unexpected DeleteOldInstances() error: %v", err)

		return
	}

	_, err = instances.ForInstanceID(id).GetSummary(context.Background())
	if err != nil {
		t.Errorf("unexpected GetSummary() error: %v", err)

		return
	}

	status := instancestore.InstanceStatusComplete
	tf := time.Now().Add(-5 * time.Second)
	err = instances.ForInstanceID(id).UpdateInstanceData(context.Background(), &instancestore.UpdateInstanceDataArgs{
		Server:  server,
		Status:  &status,
		EndedAt: &tf,
	})
	if err != nil {
		t.Errorf("unexpected UpdateInstanceData() error: %v", err)

		return
	}

	err = instances.DeleteOldInstances(context.Background(), time.Now().Add(-30*time.Second))
	if err != nil {
		t.Errorf("unexpected DeleteOldInstances() error: %v", err)

		return
	}

	_, err = instances.ForInstanceID(id).GetSummary(context.Background())
	if err != nil {
		t.Errorf("unexpected GetSummary() error: %v", err)

		return
	}

	err = instances.DeleteOldInstances(context.Background(), time.Now())
	if err != nil {
		t.Errorf("unexpected DeleteOldInstances() error: %v", err)

		return
	}

	_, err = instances.ForInstanceID(id).GetSummary(context.Background())
	if !errors.Is(err, instancestore.ErrNotFound) {
		t.Errorf("unexpected GetSummary() error: expect is '%v' but got '%v'", instancestore.ErrNotFound, err)

		return
	}
}

// nolint
func Test_sqlInstanceStore_GetNamespaceInstanceCounts(t *testing.T) {
	server := uuid.New()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	instances := instancestoresql.NewSQLInstanceStore(db)

	var tests []assertInstanceStoreCorrectGetNamespaceInstancesTest

	wfPath := "/test.yaml"

	args := &instancestore.CreateInstanceDataArgs{
		ID:           uuid.New(),
		Server:       server,
		Invoker:      "api",
		WorkflowPath: wfPath,
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

	nsID := uuid.New()
	tests = append(tests, assertInstanceStoreCorrectGetNamespaceInstancesTest{
		name: "validCase",
		args: args,
		nsID: nsID,
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

	tf := time.Now().Add(-30 * time.Second)
	status := instancestore.InstanceStatusComplete

	err = instances.ForInstanceID(tests[0].args.ID).UpdateInstanceData(context.Background(), &instancestore.UpdateInstanceDataArgs{
		Server:   server,
		Deadline: &tf,
		Status:   &status,
	})
	if err != nil {
		t.Errorf("unexpected UpdateInstanceData() error: %v", err)

		return
	}

	res, err := instances.GetNamespaceInstanceCounts(context.Background(), nsID, wfPath)
	if err != nil {
		t.Errorf("unexpected GetNamespaceInstances() error: %v", err)

		return
	}

	if res.Complete != 1 || res.Pending != 2 || res.Cancelled != 0 || res.Crashed != 0 || res.Failed != 0 || res.Total != 3 {
		t.Errorf("unexpected GetNamespaceInstanceCounts() error: got '%v' ", res)

		return
	}
}
