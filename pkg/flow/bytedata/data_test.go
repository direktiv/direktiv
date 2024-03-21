package bytedata

import (
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/google/uuid"
)

func TestConvertMirrorProcessesToGrpcMirrorActivityInfoList(t *testing.T) {
	// Test data
	process1 := &datastore.MirrorProcess{
		ID:        uuid.New(),
		UpdatedAt: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	process2 := &datastore.MirrorProcess{
		ID:        uuid.New(),
		UpdatedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	list := []*datastore.MirrorProcess{process2, process1} // Intentionally out of order

	// Invoke the function
	result := ConvertMirrorProcessesToGrpcMirrorActivityInfoList(list)

	// Assertions
	if len(result) != 2 {
		t.Fatalf("Expected result length to be 2, but got %d", len(result))
	}

	if result[0].Id != process1.ID.String() {
		t.Errorf("Expected first result ID to be %s, but got %s", process1.ID.String(), result[0].Id)
	}

	if result[0].UpdatedAt.AsTime() != process1.UpdatedAt {
		t.Errorf("Expected first result UpdatedAt to be %v, but got %v", process1.UpdatedAt, result[0].UpdatedAt.AsTime())
	}

	if result[1].Id != process2.ID.String() {
		t.Errorf("Expected second result ID to be %s, but got %s", process2.ID.String(), result[1].Id)
	}

	if result[1].UpdatedAt.AsTime() != process2.UpdatedAt {
		t.Errorf("Expected second result UpdatedAt to be %v, but got %v", process2.UpdatedAt, result[1].UpdatedAt.AsTime())
	}
}
