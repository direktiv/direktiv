package localsecrets_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/direktiv/direktiv/pkg/secrets/localsecrets"
	"github.com/google/uuid"
)

var db *database.DB
var ns *datastore.Namespace

func TestSuper(t *testing.T) {
	var err error

	// initialize secrets store
	db, ns, err = database.NewTestDBWithNamespace(t, uuid.NewString())
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

	t.Run("TestSource", testSource)
	t.Run("TestMissingNamespace", testMissingNamespace)
	t.Run("TestBadConfig", testBadConfig)
}

func testSource(t *testing.T) {
	// initialize driver
	driver := &localsecrets.Driver{
		SecretsStore: db.DataStore().Secrets(),
	}

	// marshal config
	config := localsecrets.Config{
		DriverName: localsecrets.DriverName,
		Namespace:  ns.Name,
	}

	configData, _ := json.Marshal(config)

	// construct source
	src := driver.ConstructSource(configData)

	// query undefined secret
	data, err := src.Get(context.Background(), "x")
	if !errors.Is(err, secrets.ErrSecretNotFound) {
		t.Errorf("expected secret not found, but got: %v", err)
	}
	if data != nil {
		t.Errorf("expected nil data, but got: %v", data)
	}

	// query defined secret
	data, err = src.Get(context.Background(), "test")
	if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}
	if data == nil || string(data) != "value" {
		t.Errorf("expected data 'value', but got: %s", data)
	}
}

func testMissingNamespace(t *testing.T) {
	// initialize driver
	driver := &localsecrets.Driver{
		SecretsStore: db.DataStore().Secrets(),
	}

	// marshal config
	config := localsecrets.Config{
		DriverName: localsecrets.DriverName,
		Namespace:  "x" + ns.Name,
	}

	configData, _ := json.Marshal(config)

	// construct source
	src := driver.ConstructSource(configData)

	// query
	_, err := src.Get(context.Background(), "x")
	if err == nil {
		t.Errorf("expected an error, but got none")

		// TODO: this logic currently cannot recognize the difference between a missing secret and a missing namespace
	}
}

func testBadConfig(t *testing.T) {
	// initialize driver
	driver := &localsecrets.Driver{
		SecretsStore: db.DataStore().Secrets(),
	}

	// first bad config: incorrect driver name

	configData, _ := json.Marshal(localsecrets.Config{
		DriverName: "x" + localsecrets.DriverName,
		Namespace:  ns.Name,
	})

	src := driver.ConstructSource(configData)

	_, err := src.Get(context.Background(), "x")
	if err == nil {
		t.Errorf("expected an error, but got none")
	}

	expect := "invalid config for 'database' driver: invalid driver name: 'xdatabase'"
	if err.Error() != expect {
		t.Errorf("expected error '%s', but got '%s'", expect, err.Error())
	}

	// second bad config: missing namespace name

	configData, _ = json.Marshal(localsecrets.Config{
		DriverName: localsecrets.DriverName,
		Namespace:  "",
	})

	src = driver.ConstructSource(configData)

	_, err = src.Get(context.Background(), "x")
	if err == nil {
		t.Errorf("expected an error, but got none")
	}

	expect = "invalid config for 'database' driver: missing namespace"
	if err.Error() != expect {
		t.Errorf("expected error '%s', but got '%s'", expect, err.Error())
	}

	// third bad config: unexpected fields

	configData, _ = json.Marshal(map[string]interface{}{
		"DriverName":       localsecrets.DriverName,
		"Namespace":        ns.Name,
		"ConnectionString": "blah",
	})

	src = driver.ConstructSource(configData)

	_, err = src.Get(context.Background(), "x")
	if err == nil {
		t.Errorf("expected an error, but got none")
	}

	expect = "invalid config for 'database' driver: json: unknown field \"ConnectionString\""
	if err.Error() != expect {
		t.Errorf("expected error '%s', but got '%s'", expect, err.Error())
	}
}
