package secrets_test

import (
	"encoding/json"
	"testing"

	"github.com/direktiv/direktiv/pkg/secrets"
)

func TestSecretRefUnmarshalling(t *testing.T) {
	var ref secrets.SecretRef

	// test string form
	data := []byte(`"mysecret"`)
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Error(err)
	}

	expect := `{"name":"mysecret","path":"","source":""}`
	got, _ := json.Marshal(&ref)
	if string(got) != expect {
		t.Errorf("expect '%s'; got '%s'", expect, string(got))
	}

	if err := ref.Validate(); err != nil {
		t.Error(err)
	}

	// test struct form
	expect = `{"name":"mysecret","path":"folder/mysecret","source":"vault"}`
	data = []byte(expect)
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Error(err)
	}

	got, _ = json.Marshal(&ref)
	if string(got) != expect {
		t.Errorf("expect '%s'; got '%s'", expect, string(got))
	}

	if err := ref.Validate(); err != nil {
		t.Error(err)
	}
}
