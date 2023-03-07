package flow

import (
	"fmt"
	"testing"
)

func TestExtractCaller(t *testing.T) {
	caller, err := extractCaller("/api/instance:c1d87df6-56fb-4b03-a9e9-00e5122e4884/instance:105cbf37-76b9-452a-b67d-5c9a8cd54ecc")
	if caller != "api" {
		t.Errorf("got %s; want api", caller)
	}
	if err != nil {
		t.Errorf("%s; got an unexpected error", err)
	}
	caller, err = extractCaller("/cron/instance:c1d87df6-56fb-4b03-a9e9-00e5122e4884/instance:105cbf37-76b9-452a-b67d-5c9a8cd54ecc")
	if caller != "cron" {
		t.Errorf("got %s; want api", caller)
	}
	if err != nil {
		t.Errorf("%s; got an unexpected error", err)
	}
	caller, err = extractCaller("/instance:c1d87df6-56fb-4b03-a9e9-00e5122e4884/instance:105cbf37-76b9-452a-b67d-5c9a8cd54ecc")
	if caller != "instance:c1d87df6-56fb-4b03-a9e9-00e5122e4884" {
		t.Errorf("got %s; want instance:c1d87df6-56fb-4b03-a9e9-00e5122e4884", caller)
	}
	if err != nil {
		t.Errorf("%s; got an unexpected error", err)
	}
	caller, err = extractCaller("/")
	if err == nil {
		t.Errorf("expected an error, got result a back %s", caller)
	}
}

func TestExtractRoot(t *testing.T) {
	expected := "c1d87df6-56fb-4b03-a9e9-00e5122e4884"
	root, err := extractRoot(fmt.Sprintf("/api/instance:%s/instance:105cbf37-76b9-452a-b67d-5c9a8cd54ecc", expected))
	if root != expected {
		t.Errorf("got %s; want %s", root, expected)
	}

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}
	_, err = extractRoot("/api")
	if err == nil {
		t.Error("expected an error")
	}
	out, _ := extractRoot(fmt.Sprintf("/cron/instance:%s", expected))
	if out != expected {
		t.Errorf("got %s; want %s", out, expected)
	}
}
