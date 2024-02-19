package service

import (
	"fmt"
	"testing"
)

type mockedObject struct {
	idStr   string
	hashStr string
}

func (m *mockedObject) GetID() string {
	return m.idStr
}

func (m *mockedObject) GetValueHash() string {
	return m.hashStr
}

func Test_reconcile_case1(t *testing.T) {
	src := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc5", hashStr: "hash5"},
	}

	result := Run(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.Deletes, result.Creates, result.Updates)
	want := "delete:[svc5], create:[svc3], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case2(t *testing.T) {
	src := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []Item{}

	result := Run(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.Deletes, result.Creates, result.Updates)
	want := "delete:[], create:[svc1 svc2 svc3], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case3(t *testing.T) {
	src := []Item{}

	target := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	result := Run(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.Deletes, result.Creates, result.Updates)
	want := "delete:[svc1 svc2 svc3], create:[], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case4(t *testing.T) {
	src := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []Item{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash4"},
		&mockedObject{idStr: "svc5", hashStr: "hash5"},
	}

	result := Run(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.Deletes, result.Creates, result.Updates)
	want := "delete:[svc5], create:[svc3], update:[svc2]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}
