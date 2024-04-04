package sidecar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFiles_Success(t *testing.T) {
	// Create temporary directory for testing
	dir := os.TempDir()
	defer os.RemoveAll(dir)

	// Prepare test data
	data := []byte("This is test content")
	encodedData := base64.StdEncoding.EncodeToString(data)

	// Test cases for different file types
	testCases := []struct {
		desc      string
		fileDefs  []FunctionFileDefinition
		expectErr bool
	}{
		{
			desc: "Single plain text file",
			fileDefs: []FunctionFileDefinition{
				{Key: "test_file.txt", Content: encodedData, Type: "plain"},
			},
			expectErr: false,
		},
		{
			desc: "Valid TAR.GZ archive",
			fileDefs: []FunctionFileDefinition{
				{Key: "archive.tar.gz", Content: encodeTestTarGz(data), Type: "tar.gz"},
			},
			expectErr: false,
		},
		{
			desc: "Valid TAR archive",
			fileDefs: []FunctionFileDefinition{
				{Key: "archive.tar", Content: encodeTestTar(data), Type: "tar"},
			},
			expectErr: false,
		},
		{
			desc: "Multiple files of different types",
			fileDefs: []FunctionFileDefinition{
				{Key: "text_file.txt", Content: encodedData, Type: "plain"},
				{Key: "encoded_file.bin", Content: encodedData, Type: "base64"},
				// Add more cases for "tar" and "tar.gz" as needed
			},
			expectErr: false,
		},
		{
			desc: "Invalid file type",
			fileDefs: []FunctionFileDefinition{
				{Key: "invalid.file", Content: encodedData, Type: "unknown"},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := writeFiles(filepath.Join(dir, "action-123"), tc.fileDefs)
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for test case: %s", tc.desc)
			} else if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectErr {
				// Verify file content for each file
				for _, fileDef := range tc.fileDefs {
					fileKey := fileDef.Key
					if fileDef.Type == "tar" || fileDef.Type == "tar.gz" {
						fileKey = "test_file.txt"
					}
					b, err := os.ReadFile(filepath.Join(dir, "action-123", fileKey))
					if err != nil {
						t.Errorf("Error reading file %s: %v", fileDef.Key, err)
					}
					if !bytes.Equal(b, data) {
						t.Errorf("Expected content not found in file %s", fileDef.Key)
					}

				}
			}
		})
	}
}

func encodeTestTar(data []byte) string {
	// Create a test TAR archive with a single file containing the data
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	header := new(tar.Header)
	header.Name = "test_file.txt"
	header.Size = int64(len(data))
	if err := tw.WriteHeader(header); err != nil {
		return ""
	}
	_, err := tw.Write(data)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func encodeTestTarGz(data []byte) string {
	// Create a test TAR.GZ archive with the same content as the TAR
	buf := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	defer buf.Reset()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	header := new(tar.Header)
	header.Name = "test_file.txt"
	header.Size = int64(len(data))
	if err := tw.WriteHeader(header); err != nil {
		return ""
	}
	if _, err := tw.Write(data); err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
