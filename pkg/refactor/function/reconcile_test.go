package function

import (
	"fmt"
	"testing"
)

type mockedObject struct {
	idStr   string
	hashStr string
}

func (m *mockedObject) id() string {
	return m.idStr
}

func (m *mockedObject) hash() string {
	return m.hashStr
}

func Test_reconcile_case1(t *testing.T) {
	src := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc5", hashStr: "hash5"},
	}

	result := reconcile(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.deletes, result.creates, result.updates)
	want := "delete:[svc5], create:[svc3], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case2(t *testing.T) {
	src := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []reconcileObject{}

	result := reconcile(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.deletes, result.creates, result.updates)
	want := "delete:[], create:[svc1 svc2 svc3], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case3(t *testing.T) {
	src := []reconcileObject{}

	target := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	result := reconcile(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.deletes, result.creates, result.updates)
	want := "delete:[svc1 svc2 svc3], create:[], update:[]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}

func Test_reconcile_case4(t *testing.T) {
	src := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash2"},
		&mockedObject{idStr: "svc3", hashStr: "hash3"},
	}

	target := []reconcileObject{
		&mockedObject{idStr: "svc1", hashStr: "hash1"},
		&mockedObject{idStr: "svc2", hashStr: "hash4"},
		&mockedObject{idStr: "svc5", hashStr: "hash5"},
	}

	result := reconcile(src, target)

	got := fmt.Sprintf("delete:%s, create:%s, update:%s", result.deletes, result.creates, result.updates)
	want := "delete:[svc5], create:[svc3], update:[svc2]"

	if got != want {
		t.Errorf("want reconsile result: %v, go: %v", want, got)
	}
}
