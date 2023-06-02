package datastore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
)

func Test_sqlRuntimeVariablesStore_SetAndGet(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)
	file := createFile(t, fs)

	testVar := &core.RuntimeVariable{
		WorkflowID: file.ID,
		Name:       "myvar",
		MimeType:   "text/json",
		Data:       []byte("some data"),
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
	if variable.WorkflowID != testVar.WorkflowID ||
		variable.Name != testVar.Name {
		t.Errorf("unexpected GetByID() result: %v", variable)

		return
	}

	return
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
		WorkflowID: file.ID,
		Name:       "myvar$$",
		MimeType:   "text/json",
		Data:       []byte("some data"),
	}
	_, err = ds.RuntimeVariables().Set(context.Background(), testVar)
	if err == nil {
		t.Errorf("unexpected Set() nil error")

		return
	}

	return
}

func Test_sqlRuntimeVariablesStore_CrudOnList(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)
	file := createFile(t, fs)

	// Test Set().
	for _, i := range []int{0, 1, 2, 3} {
		v := &core.RuntimeVariable{
			WorkflowID: file.ID,
			Name:       fmt.Sprintf("var_%d", i),
			MimeType:   "text/json",
			Data:       []byte(fmt.Sprintf("data_%d", i)),
		}
		_, err = ds.RuntimeVariables().Set(context.Background(), v)
		if err != nil {
			t.Errorf("unexpected Set() error: %v", err)

			return
		}
		if v == nil {
			t.Errorf("unexpected Set() nil result")

			return
		}
	}

	// Test ListByWorkflowID().
	vars, err := ds.RuntimeVariables().ListByWorkflowID(context.Background(), file.ID)

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

		if v.WorkflowID != file.ID ||
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

		if v.WorkflowID != file.ID ||
			v.Name != fmt.Sprintf("var_%d", i) ||
			string(data) != fmt.Sprintf("data_%d", i) {
			t.Errorf("unexpected ListByWorkflowID()[%d] result: %v", i, v)

			return
		}
	}

	return
}
