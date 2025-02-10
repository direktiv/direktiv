package secrets_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/secrets"
)

func TestMemoryCache(t *testing.T) {
	c := &secrets.MemoryCache{}

	l, _ := c.List(nil)

	if len(l) != 0 {
		t.Fail()
	}

	if err := c.Insert(context.Background(), secrets.Secret{
		Path:   "mysecret",
		Source: "local",
	}); err != nil {
		t.Error(err)
	}

	l, _ = c.List(nil)

	if len(l) != 1 {
		t.FailNow()
	}

	if l[0].Path != "mysecret" || l[0].Source != "local" {
		t.FailNow()
	}

	if err := c.Insert(context.Background(), secrets.Secret{
		Path:   "mysecret",
		Source: "local",
	}); err == nil {
		t.FailNow()
	}

	if err := c.Insert(context.Background(), secrets.Secret{
		Path:   "mysecret2",
		Source: "local",
	}); err != nil {
		t.Error(err)
	}

	l, _ = c.List(nil)

	if len(l) != 2 {
		t.FailNow()
	}
}
