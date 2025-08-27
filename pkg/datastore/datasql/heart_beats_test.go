package datasql_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/direktiv/direktiv/internal/testutils"
	"github.com/direktiv/direktiv/pkg/datastore"
)

func Test_HeartBeats(t *testing.T) {
	db, ns, err := testutils.NewTestDBWithNamespace(t, uuid.NewString())
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

	res, err := ds.HeartBeats().Since(context.Background(), "some_group", 0)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("unepxected result: %v", res)
	}

	err = ds.HeartBeats().Set(context.Background(), &datastore.HeartBeat{
		Group: "some_group",
		Key:   "some_key",
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	err = ds.HeartBeats().Set(context.Background(), &datastore.HeartBeat{
		Group: "some_group",
		Key:   "some_key",
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	time.Sleep(400 * time.Millisecond)

	err = ds.HeartBeats().Set(context.Background(), &datastore.HeartBeat{
		Group: "some_group",
		Key:   "some_key_2",
	})
	if err != nil {
		t.Errorf("error: %v", err)
	}

	res, err = ds.HeartBeats().Since(context.Background(), "some_group", 1)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("unepxected result: %v", res)
	}

	time.Sleep(610 * time.Millisecond)

	res, err = ds.HeartBeats().Since(context.Background(), "some_group", 1)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("unepxected result: %v", res)
	}

}
