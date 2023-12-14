package core_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

// Example of a timing-safe comparison of two HMACs with a constant-time function
func TestCompareHMAC(t *testing.T) {
	// In your actual code, use this function like this:
	expectedHMAC := core.ComputeHMAC("data", []byte("some-secret-key"))
	actualHMAC := core.ComputeHMAC("data", []byte("some-secret-key"))

	// Compare HMACs in a constant time manner
	if !core.CompareHMAC(expectedHMAC, actualHMAC) {
		t.Errorf("HMACs do not match")
	}

	expectedHMAC = core.ComputeHMAC("data", []byte("your-secret-key"))
	actualHMAC = core.ComputeHMAC("data", []byte("wrong-secret-key"))

	// Compare HMACs in a constant time manner
	if core.CompareHMAC(expectedHMAC, actualHMAC) {
		t.Errorf("HMACs should not match")
	}
}