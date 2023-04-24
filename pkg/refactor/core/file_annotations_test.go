package core_test

import (
	"reflect"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

func TestFileAnnotationsData(t *testing.T) {
	data := core.FileAnnotationsData{}
	data = data.SetEntry("foo", "bar")

	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"foo": "bar",
	}) {
		t.Errorf("SetEntry() failed")
	}

	data = data.SetEntry("foo2", "bar2")

	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"foo":  "bar",
		"foo2": "bar2",
	}) {
		t.Errorf("SetEntry() failed")
	}

	if !reflect.DeepEqual(data.GetEntry("foo"), "bar") {
		t.Errorf("GetEntry() failed")
	}
	if !reflect.DeepEqual(data.GetEntry("foo2"), "bar2") {
		t.Errorf("GetEntry() failed")
	}

	data = data.RemoveEntry("foo")
	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"foo2": "bar2",
	}) {
		t.Errorf("RemoveEntry() failed")
	}
	data = data.RemoveEntry("foo2")
	if !reflect.DeepEqual(data, core.FileAnnotationsData{}) {
		t.Errorf("RemoveEntry() failed")
	}
}
