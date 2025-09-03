package datasql_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/internal/datastore/datasql"
	database2 "github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"

	"github.com/direktiv/direktiv/internal/datastore"
)

func Test_Secrets(t *testing.T) {
	ns := uuid.NewString()
	conn, err := database2.NewTestDBWithNamespace(t, ns)
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	ds := datasql.NewStore(conn)
	err = ds.Secrets().Set(context.Background(), &datastore.Secret{
		Name:      "test",
		Namespace: ns,
		Data:      []byte("value"),
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	res, err := ds.Secrets().Get(context.Background(), ns, "test")
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
