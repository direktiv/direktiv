package secrets_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/secrets"
)

func TestController(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	x := MockSource{
		Secrets: map[string][]byte{
			"x": []byte("5"),
		},
	}
	memoryData, _ := json.Marshal(x)

	ctrl := secrets.NewController(&secrets.Config{
		DefaultSource: "memory",
		RetryTime:     time.Millisecond * 10,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "memory",
				Driver: "mock",
				Data:   memoryData,
			},
		},
	}, &secrets.MemoryCache{})

	// first list should have zero entries because nothing has been fetched from a source and loaded into the cache
	l, _ := ctrl.List(context.Background())
	if len(l) != 0 {
		t.Errorf("unexpected list data")
	}

	// lookup some secrets to start populating the cache
	l, _ = ctrl.Lookup(context.Background(), []secrets.SecretRef{
		{
			Name:   "mysecret",
			Path:   "x",
			Source: "memory",
		},
	})
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectations
	expect := `[{"path":"x","source":"memory","data":"NQ==","error":null}]`
	resultData, _ := json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}

	// check that the values we looked up earlier are still in the cache
	l, _ = ctrl.List(context.Background())
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectations
	expect = `[{"path":"x","source":"memory","data":"NQ==","error":null}]`
	resultData, _ = json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}
}

func TestListSort(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	srcA := MockSource{
		Secrets: map[string][]byte{
			"x": []byte("5"),
			"z": []byte("6"),
			"y": []byte("7"),
		},
	}
	srcAMemData, _ := json.Marshal(srcA)

	srcB := MockSource{
		Secrets: map[string][]byte{
			"a": []byte("1"),
			"c": []byte("2"),
			"b": []byte("3"),
		},
	}
	srcBMemData, _ := json.Marshal(srcB)

	ctrl := secrets.NewController(&secrets.Config{
		DefaultSource: "memory",
		RetryTime:     time.Millisecond * 10,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "a",
				Driver: "mock",
				Data:   srcAMemData,
			},
			{
				Name:   "b",
				Driver: "mock",
				Data:   srcBMemData,
			},
		},
	}, &secrets.MemoryCache{})

	// lookup some secrets to start populating the cache
	l, _ := ctrl.Lookup(context.Background(), []secrets.SecretRef{
		{
			Name:   "X",
			Path:   "x",
			Source: "a",
		},
		{
			Name:   "Z",
			Path:   "z",
			Source: "a",
		},
		{
			Name:   "Y",
			Path:   "y",
			Source: "a",
		},
		{
			Name:   "A",
			Path:   "a",
			Source: "b",
		},
		{
			Name:   "C",
			Path:   "c",
			Source: "b",
		},
		{
			Name:   "B",
			Path:   "b",
			Source: "b",
		},
	})
	if len(l) != 6 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectations, including NOT sorting (should be returned in reference order)
	expect := `[{"path":"x","source":"a","data":"NQ==","error":null},{"path":"z","source":"a","data":"Ng==","error":null},{"path":"y","source":"a","data":"Nw==","error":null},{"path":"a","source":"b","data":"MQ==","error":null},{"path":"c","source":"b","data":"Mg==","error":null},{"path":"b","source":"b","data":"Mw==","error":null}]`
	resultData, _ := json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}

	// check that the values we looked up earlier are still in the cache
	l, _ = ctrl.List(context.Background())
	if len(l) != 6 {
		t.Errorf("unexpected list data")
	}

	// check that this second function also returns the correct (and sorted) data
	expect = `[{"path":"x","source":"a","data":"NQ==","error":null},{"path":"y","source":"a","data":"Nw==","error":null},{"path":"z","source":"a","data":"Ng==","error":null},{"path":"a","source":"b","data":"MQ==","error":null},{"path":"b","source":"b","data":"Mw==","error":null},{"path":"c","source":"b","data":"Mg==","error":null}]`
	resultData, _ = json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}
}

func TestSecretNotFound(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	x := MockSource{
		Secrets: map[string][]byte{},
	}
	memoryData, _ := json.Marshal(x)

	ctrl := secrets.NewController(&secrets.Config{
		DefaultSource: "memory",
		RetryTime:     time.Millisecond * 10,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "memory",
				Driver: "mock",
				Data:   memoryData,
			},
		},
	}, &secrets.MemoryCache{})

	// lookup some secrets to start populating the cache
	l, _ := ctrl.Lookup(context.Background(), []secrets.SecretRef{
		{
			Name:   "X",
			Path:   "x",
			Source: "memory",
		},
	})
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing secret error
	expect := `[{"path":"x","source":"memory","data":null,"error":"secret not found"}]`
	resultData, _ := json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}

	// check that the values we looked up earlier are still in the cache
	l, _ = ctrl.List(context.Background())
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing secret error
	expect = `[{"path":"x","source":"memory","data":null,"error":"secret not found"}]`
	resultData, _ = json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}
}

func TestSourceNotFound(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	x := MockSource{
		Secrets: map[string][]byte{},
	}
	memoryData, _ := json.Marshal(x)

	ctrl := secrets.NewController(&secrets.Config{
		DefaultSource: "memory",
		RetryTime:     time.Millisecond * 10,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "memory",
				Driver: "mock",
				Data:   memoryData,
			},
		},
	}, &secrets.MemoryCache{})

	// lookup some secrets to start populating the cache
	l, _ := ctrl.Lookup(context.Background(), []secrets.SecretRef{
		{
			Name:   "X",
			Path:   "x",
			Source: "mysource",
		},
	})
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing source error
	expect := `[{"path":"x","source":"mysource","data":null,"error":"source not found: mysource"}]`
	resultData, _ := json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}

	// check that the values we looked up earlier are still in the cache
	l, _ = ctrl.List(context.Background())
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing source error
	expect = `[{"path":"x","source":"mysource","data":null,"error":"source not found: mysource"}]`
	resultData, _ = json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}
}

func TestDriverNotFound(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	x := MockSource{
		Secrets: map[string][]byte{},
	}
	memoryData, _ := json.Marshal(x)

	ctrl := secrets.NewController(&secrets.Config{
		DefaultSource: "memory",
		RetryTime:     time.Millisecond * 10,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "memory",
				Driver: "mydriver",
				Data:   memoryData,
			},
		},
	}, &secrets.MemoryCache{})

	// lookup some secrets to start populating the cache
	l, _ := ctrl.Lookup(context.Background(), []secrets.SecretRef{
		{
			Name:   "X",
			Path:   "x",
			Source: "memory",
		},
	})
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing driver error
	expect := `[{"path":"x","source":"memory","data":null,"error":"driver not found: mydriver"}]`
	resultData, _ := json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}

	// check that the values we looked up earlier are still in the cache
	l, _ = ctrl.List(context.Background())
	if len(l) != 1 {
		t.Errorf("unexpected list data")
	}

	// check that the values returned match expectation: missing driver error
	expect = `[{"path":"x","source":"memory","data":null,"error":"driver not found: mydriver"}]`
	resultData, _ = json.Marshal(l)
	if string(resultData) != expect {
		t.Errorf("incorrect list data\n\texpect: %s\n\tgot: %s", expect, resultData)
	}
}
