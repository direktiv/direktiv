package core

import (
	"reflect"
	"testing"
)

func TestFileAnnotationsData_AppendFileUserAttributes(t *testing.T) {
	type args struct {
		newAttributes []string
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
			if got := tt.data.AppendFileUserAttributes(tt.args.newAttributes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendFileUserAttributes() = %v, want %v", got, tt.want)
			}
		})
	}
}