package secrets_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/secrets"
)

func TestControllerRegistry(t *testing.T) {
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

	// register controller
	err := secrets.RegisterController("user", ctrl)
	if err != nil {
		t.Error(err)
	}

	// register duplicate
	err = secrets.RegisterController("user", ctrl)
	if err == nil {
		t.Error(errors.New("expected failure due to duplicate controller"))
	}

	// get non-existing controller
	_, err = secrets.GetController("otheruser")
	if err == nil {
		t.Error(errors.New("expected failure due to missing controller"))
	}

	// get existing controller
	_, err = secrets.GetController("user")
	if err == nil {
		t.Error(errors.New("expected failure due to missing controller"))
	}

	secrets.DeleteController("user")

	// get non-existing controller
	_, err = secrets.GetController("otheruser")
	if err == nil {
		t.Error(errors.New("expected failure due to missing controller"))
	}
}
