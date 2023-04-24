package core

import (
	"reflect"
	"testing"
)

func TestFileAnnotationsData_SetEntry(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		data FileAnnotationsData
		args args
		want FileAnnotationsData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.SetEntry(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}