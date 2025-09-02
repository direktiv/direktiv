package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
)

// Checksum is a shortcut to calculate a hash for any given input by first marshalling it to json.
func Checksum(x interface{}) string {
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	hash, err := ComputeHash(data)
	if err != nil {
		panic(err)
	}

	return hash
}

// ComputeHash is a shortcut to calculate a hash for a byte slice.
func ComputeHash(data []byte) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Marshal is a shortcut to marshal any given input to json with our preferred indentation settings.
func Marshal(x interface{}) string {
	data, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data)
}
