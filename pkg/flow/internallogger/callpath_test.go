package internallogger

import (
	"fmt"
	"testing"
)

func TestAppendInstanceID(t *testing.T) {
	callpath := "/c1d87df6-56fb-4b03-a9e9-00e5122e4884"
	instanceID := "105cbf37-76b9-452a-b67d-5c9a8cd54ecc"
	prefix := AppendInstanceID(callpath, instanceID)
	expected := callpath + "/" + instanceID
	if prefix != expected {
		t.Errorf("got %s; want %s", prefix, expected)
	}
	callpath = ""
	instanceID = ""
	prefix = AppendInstanceID(callpath, instanceID)
	expected = "/"
	if prefix != expected {
		t.Errorf("got %s; want %s", prefix, expected)
	}
	callpath = "/"
	instanceID = ""
	prefix = AppendInstanceID(callpath, instanceID)
	expected = "/"
	if prefix != expected {
		t.Errorf("got %s; want %s", prefix, expected)
	}
	callpath = "/"
	instanceID = "105cbf37-76b9-452a-b67d-5c9a8cd54ecc"
	prefix = AppendInstanceID(callpath, instanceID)
	expected = "/" + instanceID
	if prefix != expected {
		t.Errorf("got %s; want %s", prefix, expected)
	}
}

func TestGetRootinstanceID(t *testing.T) {
	expected := "c1d87df6-56fb-4b03-a9e9-00e5122e4884"
	root, err := getRootinstanceID(fmt.Sprintf("/%s/105cbf37-76b9-452a-b67d-5c9a8cd54ecc", expected))
	if root != expected {
		t.Errorf("got %s; want %s", root, expected)
	}

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}
	_, err = getRootinstanceID("/api")
	if err == nil {
		t.Error("expected an error")
	}
	out, _ := getRootinstanceID(fmt.Sprintf("/%s", expected))
	if out != expected {
		t.Errorf("got %s; want %s", out, expected)
	}
}
