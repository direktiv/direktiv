package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/direktiv/direktiv-ui/server/backend/server"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestReadConfigAndPrepareFileNotExists(t *testing.T) {

	_, err := server.ReadConfigAndPrepare("", true)
	assert.Error(t, err)

}

func writeConfig(c *server.Config) (string, error) {

	file, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp(os.TempDir(), "config")
	if err != nil {
		return "", err
	}
	err = os.WriteFile(f.Name(), file, 0644)

	return f.Name(), err
}

func TestReadConfigAndPrepareJSON(t *testing.T) {

	c := &server.Config{}

	path, err := writeConfig(c)
	assert.NoError(t, err)
	defer os.Remove(path)

	// switch stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stderr = w

	_, err = server.ReadConfigAndPrepare(path, true)
	assert.NoError(t, err)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	var log map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &log)
	assert.NoError(t, err)

	// switch stdout and reset buffer
	buf.Reset()
	r, w, _ = os.Pipe()
	os.Stderr = w

	_, err = server.ReadConfigAndPrepare(path, false)
	assert.NoError(t, err)
	w.Close()
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)
	os.Stderr = old

	err = json.Unmarshal(buf.Bytes(), &log)
	assert.Error(t, err)

}

func TestReadConfigLevelDebug(t *testing.T) {

	c := &server.Config{
		Log: server.Log{
			API: "debug",
		},
	}

	path, err := writeConfig(c)
	assert.NoError(t, err)
	defer os.Remove(path)

	_, err = server.ReadConfigAndPrepare(path, true)
	assert.NoError(t, err)

	e := log.Debug()
	assert.True(t, e.Enabled())

}

func TestReadConfigLevelInfo(t *testing.T) {

	c := &server.Config{}

	path, err := writeConfig(c)
	assert.NoError(t, err)
	defer os.Remove(path)

	_, err = server.ReadConfigAndPrepare(path, true)
	assert.NoError(t, err)

	e := log.Debug()
	assert.False(t, e.Enabled())

}
