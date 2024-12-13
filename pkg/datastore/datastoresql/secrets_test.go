package datastoresql_test

import (
	"context"
	"github.com/google/uuid"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
)

func Test_Secrets(t *testing.T) {
	db, ns, err := database.NewTestDataStoreWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDataStoreWithNamespace() error = %v", err)
	}
	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
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
