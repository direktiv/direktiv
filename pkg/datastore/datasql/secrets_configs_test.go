package datasql_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
)

func Test_SecretsConfigs(t *testing.T) {
	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}
	ds := db.DataStore()
	err = ds.SecretsConfigs().Set(context.Background(), &datastore.SecretsConfigs{
		Namespace:     ns.Name,
		Configuration: []byte("value"),
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	res, err := ds.SecretsConfigs().Get(context.Background(), ns.Name)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if string(res.Configuration) != "value" {
		t.Errorf("value does not match, was %v should %v", string(res.Configuration), "value")
	}

	if err != nil {
		t.Errorf("error: %v", err)
	}
}
