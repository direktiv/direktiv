package datastore_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/datastore"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
)

func Test_sqlMirrorStore_Process_SetAndGet(t *testing.T) {
	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDB() error = %v", err)
	}
	ds := db.DataStore()

	newProcess := &datastore.MirrorProcess{
		ID:        uuid.New(),
		Namespace: ns.Name,
		Status:    "new",
	}

	// test create.
	process, err := ds.Mirror().CreateProcess(context.Background(), newProcess)
	if err != nil {
		t.Errorf("unexpected CreateProcess() error: %v", err)
	}

	if newProcess.ID != process.ID {
		t.Errorf("unexpected CreateProcess().ID, want: %v, got %v", newProcess.ID, process.ID)
	}
	if newProcess.Status != process.Status {
		t.Errorf("unexpected CreateProcess().Status, want: %v, got %v", newProcess.Status, process.Status)
	}

	// test update.
	newProcess.Status = "other"
	process, err = ds.Mirror().UpdateProcess(context.Background(), newProcess)
	if err != nil {
		t.Errorf("unexpected UpdateProcess() error: %v", err)
	}
	if newProcess.ID != process.ID {
		t.Errorf("unexpected UpdateProcess().ID, want: %v, got %v", newProcess.ID, process.ID)
	}
	if newProcess.Status != process.Status {
		t.Errorf("unexpected UpdateProcess().Status, want: %v, got %v", newProcess.Status, process.Status)
	}

	// test get.
	process, err = ds.Mirror().GetProcess(context.Background(), newProcess.ID)
	if err != nil {
		t.Errorf("unexpected GetProcess() error: %v", err)
	}
	if newProcess.ID != process.ID {
		t.Errorf("unexpected GetProcess().ID, want: %v, got %v", newProcess.ID, process.ID)
	}
	if newProcess.Status != process.Status {
		t.Errorf("unexpected GetProcess().Status, want: %v, got %v", newProcess.Status, process.Status)
	}

	secondProcess := &datastore.MirrorProcess{
		ID:        uuid.New(),
		Namespace: newProcess.Namespace,
		Status:    "new",
	}

	// test get by config_id.
	_, err = ds.Mirror().CreateProcess(context.Background(), secondProcess)
	if err != nil {
		t.Errorf("unexpected CreateProcess() error: %v", err)
	}

	list, err := ds.Mirror().GetProcessesByNamespace(context.Background(), newProcess.Namespace)
	if err != nil {
		t.Errorf("unexpected GetProcessesByConfig() error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("unexpected GetProcessesByConfig() length: got: %v, want %v", len(list), 2)
	}
	if list[0].ID != newProcess.ID {
		t.Errorf("unexpected GetProcess().ID, want: %v, got %v", list[0].ID, newProcess.ID)
	}
	if list[1].ID != secondProcess.ID {
		t.Errorf("unexpected GetProcess().ID, want: %v, got %v", list[1].ID, secondProcess.ID)
	}

}

func Test_sqlMirrorStore_Config_SetAndGet(t *testing.T) {
	db, ns, err := database.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDB() error = %v", err)
	}
	ds := db.DataStore()

	// test create.

	newConfig := &datastore.MirrorConfig{
		Namespace: ns.Name,
		URL:       "some_url",
		GitRef:    "123",
		AuthType:  "public",
	}
	config, err := ds.Mirror().CreateConfig(context.Background(), newConfig)
	if err != nil {
		t.Fatalf("unexpected CreateConfig() error: %v", err)
	}
	if newConfig.Namespace != config.Namespace {
		t.Errorf("unexpected CreateConfig().Namespace, want: %v, got %v", newConfig.Namespace, config.Namespace)
	}
	if newConfig.URL != config.URL {
		t.Errorf("unexpected CreateConfig().Status, want: %v, got %v", newConfig.URL, config.URL)
	}

	ns2, err := ds.Namespaces().Create(context.Background(), &datastore.Namespace{
		Name: uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("unexpected Create() error: %v", err)
	}

	secondConfig := &datastore.MirrorConfig{
		Namespace: ns2.Name,
		URL:       "some_url",
		GitRef:    "123",
		AuthType:  "public",
	}
	_, err = ds.Mirror().CreateConfig(context.Background(), secondConfig)
	if err != nil {
		t.Errorf("unexpected CreateConfig() error: %v", err)
	}

	// test update.
	newConfig.URL = "other"
	config, err = ds.Mirror().UpdateConfig(context.Background(), newConfig)
	if err != nil {
		t.Errorf("unexpected UpdateConfig() error: %v", err)
	}
	if newConfig.Namespace != config.Namespace {
		t.Errorf("unexpected UpdateConfig().Namespace, want: %v, got %v", newConfig.Namespace, config.Namespace)
	}
	if newConfig.URL != config.URL {
		t.Errorf("unexpected UpdateConfig().Status, want: %v, got %v", newConfig.URL, config.URL)
	}

	// test get.
	config, err = ds.Mirror().GetConfig(context.Background(), newConfig.Namespace)
	if err != nil {
		t.Errorf("unexpected GetConfig() error: %v", err)
	}
	if newConfig.Namespace != config.Namespace {
		t.Errorf("unexpected GetConfig().Namespace, want: %v, got %v", newConfig.Namespace, config.Namespace)
	}
	if newConfig.URL != config.URL {
		t.Errorf("unexpected GetConfig().Status, want: %v, got %v", newConfig.URL, config.URL)
	}
}
