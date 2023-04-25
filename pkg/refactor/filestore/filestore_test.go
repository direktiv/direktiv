package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func TestSha256CalculateChecksum(t *testing.T) {
	got := string(filestore.Sha256CalculateChecksum([]byte("some_string")))
	want := "539a374ff43dce2e894fd4061aa545e6f7f5972d40ee9a1676901fb92125ffee"
	if got != want {
		t.Errorf("unexpected Sha256CalculateChecksum() result, got: %s, want: %s", got, want)
	}
}
