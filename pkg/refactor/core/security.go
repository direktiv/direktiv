package core

import (
	"crypto/hmac"
	"crypto/sha256"
)

func ComputeHMAC(data string, key []byte) []byte {
	hmacSha256 := hmac.New(sha256.New, key)
	hmacSha256.Write([]byte(data))

	return hmacSha256.Sum(nil)
}

// CompareHMAC compares the expected HMAC with the actual HMAC in a constant time manner.
func CompareHMAC(expected, actual []byte) bool {
	return hmac.Equal(expected, actual)
}
