package core_test

import (
	"reflect"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

func TestFileAnnotationsData_AppendFileUserAttributes(t *testing.T) {
	data := core.FileAnnotationsData{}

	data = data.AppendFileUserAttributes([]string{
		"str1 ",
		" str1",
		" str2 ",
		"  ",
		"",
		"str3",
		"",
		"str1 ",
		" str1",
		" str2 ",
	})

	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"user_attributes": "str1,str2,str3",
	}) {
		t.Errorf("AppendFileUserAttributes() failed")
	}

	data = data.ReduceFileUserAttributes([]string{
		"str1 ",
		" str1",
		" str2 ",
		"  ",
		"",
		"",
		"str1 ",
		" str1",
		" str2 ",
	})

	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"user_attributes": "str3",
	}) {
		t.Errorf("ReduceFileUserAttributes() failed")
	}

	data = data.ReduceFileUserAttributes([]string{
		"str1 ",
		" str1",
		" str2 ",
		"  ",
		"",
		"str3",
		"",
		"str1 ",
		" str1",
		" str2 ",
	})

	if !reflect.DeepEqual(data, core.FileAnnotationsData{
		"user_attributes": "",
	}) {
		t.Errorf("ReduceFileUserAttributes() failed")
	}
}
