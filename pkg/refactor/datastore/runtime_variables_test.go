package datastore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/google/uuid"
)

func Test_sqlRuntimeVariablesStore_SetAndGet(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)
	file := createFile(t, fs)

	ns := uuid.New()

	expect := []byte("some data")

	testVar := &core.RuntimeVariable{
		NamespaceID:  ns,
		WorkflowPath: file.Path,
		Name:         "myvar",
		MimeType:     "text/json",
		Data:         expect,
	}
	variable, err := ds.RuntimeVariables().Set(context.Background(), testVar)
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)

		return
	}
	if variable == nil {
		t.Errorf("unexpected Set() nil result")

		return
	}

	variable, err = ds.RuntimeVariables().GetByID(context.Background(), variable.ID)
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)

		return
	}
	if variable == nil {
		t.Errorf("unexpected Set() nil result")

		return
	}
	if variable.WorkflowPath != testVar.WorkflowPath ||
		variable.Name != testVar.Name {
		t.Errorf("unexpected GetByID() result: %v", variable)

		return
	}

	variable, err = ds.RuntimeVariables().GetByWorkflowAndName(context.Background(), ns, file.Path, "myvar")
	if err != nil {
		t.Errorf("unexpected GetByNamespaceAndName() error: %v", err)

		return
	}
	data, err := ds.RuntimeVariables().LoadData(context.Background(), variable.ID)
	if err != nil {
		t.Errorf("unexpected LoadData() error: %v", err)

		return
	}
	if string(data) != string(expect) {
		t.Errorf("unexpected GetByNamespaceAndName() result: %s", variable.Data)

		return
	}

	list, err := ds.RuntimeVariables().ListByWorkflowPath(context.Background(), ns, file.Path)
	if err != nil {
		t.Errorf("unexpected ListByWorkflowPath() error: %v", err)

		return
	}

	if len(list) != 1 {
		t.Errorf("unexpected ListByWorkflowPath() result: %v", list)

		return
	}
}

func Test_sqlRuntimeVariablesStore_Overwrite(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")

	ns := uuid.New()

	testVar := &core.RuntimeVariable{
		NamespaceID: ns,
		Name:        "myvar",
		MimeType:    "text/json",
		Data:        []byte("some data"),
	}
	variable, err := ds.RuntimeVariables().Set(context.Background(), testVar)
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)

		return
	}
	if variable == nil {
		t.Errorf("unexpected Set() nil result")

		return
	}

	expect := []byte("some data 2")
	testVar = &core.RuntimeVariable{
		NamespaceID: ns,
		Name:        "myvar",
		MimeType:    "text/json",
		Data:        expect,
	}
	variable, err = ds.RuntimeVariables().Set(context.Background(), testVar)
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)

		return
	}
	if variable == nil {
		t.Errorf("unexpected Set() nil result")

		return
	}

	variable, err = ds.RuntimeVariables().GetByNamespaceAndName(context.Background(), ns, "myvar")
	if err != nil {
		t.Errorf("unexpected GetByNamespaceAndName() error: %v", err)

		return
	}
	data, err := ds.RuntimeVariables().LoadData(context.Background(), variable.ID)
	if err != nil {
		t.Errorf("unexpected LoadData() error: %v", err)

		return
	}
	if string(data) != string(expect) {
		t.Errorf("unexpected GetByNamespaceAndName() result: %s", variable.Data)

		return
	}
}

func Test_sqlRuntimeVariablesStore_InvalidName(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)
	file := createFile(t, fs)

	testVar := &core.RuntimeVariable{
		NamespaceID:  uuid.New(),
		WorkflowPath: file.Path,
		Name:         "myvar$$",
		MimeType:     "text/json",
		Data:         []byte("some data"),
	}
	_, err = ds.RuntimeVariables().Set(context.Background(), testVar)
	if err == nil {
		t.Errorf("unexpected Set() nil error")

		return
	}
}

func Test_sqlRuntimeVariablesStore_CrudOnList(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)
	file := createFile(t, fs)

	ns := uuid.New()

	for _, i := range []int{0, 1, 2, 3} {
		v := &core.RuntimeVariable{
			NamespaceID:  ns,
			WorkflowPath: file.Path,
			Name:         fmt.Sprintf("var_%d", i),
			MimeType:     "text/json",
			Data:         []byte(fmt.Sprintf("data_%d", i)),
		}
		_, err = ds.RuntimeVariables().Set(context.Background(), v)
		if err != nil {
			t.Errorf("unexpected Set() error: %v", err)

			return
		}
	}

	// Test ListByWorkflowID().
	vars, err := ds.RuntimeVariables().ListByWorkflowPath(context.Background(), ns, file.Path)
	if err != nil {
		t.Errorf("unexpected ListByWorkflowID() error: %v", err)

		return
	}

	if len(vars) != 4 {
		t.Errorf("unexpected ListByWorkflowID() length, got:%d want:%d", len(vars), 3)

		return
	}

	// Assert correct list after insert.
	for _, i := range []int{0, 1, 2, 3} {
		v := vars[i]

		data, err := ds.RuntimeVariables().LoadData(context.Background(), v.ID)
		if err != nil {
			t.Errorf("unexpected LoadData() error: %v", err)

			return
		}

		if v.WorkflowPath != file.Path ||
			v.Name != fmt.Sprintf("var_%d", i) ||
			string(data) != fmt.Sprintf("data_%d", i) {
			t.Errorf("unexpected ListByWorkflowID()[%d] result: %v", i, v)

			return
		}
	}

	// Test Delete().
	err = ds.RuntimeVariables().Delete(context.Background(), vars[1].ID)
	if err != nil {
		t.Errorf("unexpected Delete() error: %v", err)

		return
	}
	err = ds.RuntimeVariables().Delete(context.Background(), vars[2].ID)
	if err != nil {
		t.Errorf("unexpected Delete() error: %v", err)

		return
	}

	// Assert correct list after Delete().
	for _, i := range []int{0, 3} {
		v := vars[i]

		data, err := ds.RuntimeVariables().LoadData(context.Background(), v.ID)
		if err != nil {
			t.Errorf("unexpected LoadData() error: %v", err)

			return
		}

		if v.WorkflowPath != file.Path ||
			v.Name != fmt.Sprintf("var_%d", i) ||
			string(data) != fmt.Sprintf("data_%d", i) {
			t.Errorf("unexpected ListByWorkflowID()[%d] result: %v", i, v)

			return
		}
	}
}
