package datasql_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/internal/database"
	"github.com/google/uuid"

	"github.com/direktiv/direktiv/internal/datastore"
)

func Test_Secrets(t *testing.T) {
	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	ds := db.DataStore()
	err = ds.Secrets().Set(context.Background(), &datastore.Secret{
		Name:      "test",
		Namespace: ns.Name,
		Data:      []byte("value"),
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	res, err := ds.Secrets().Get(context.Background(), ns.Name, "test")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if string(res.Data) != "value" {
		t.Errorf("value does not match, was %v should %v", string(res.Data), "value")
	}

	if err != nil {
		t.Errorf("error: %v", err)
	}
	l, err := ds.Secrets().GetAll(context.Background(), "ns")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	for _, s := range l {
		if string(s.Data) != "value" {
			t.Errorf("value does not match, was %v should %v", string(s.Data), "value")
		}
	}
}
