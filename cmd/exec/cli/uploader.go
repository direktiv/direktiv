package cli

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type uploader struct {
	matcher gitignore.Matcher
	profile profile
}

type fileObject struct {
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Typ      string `json:"type,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newUploader(projectRoot string, profile profile) (*uploader, error) {
	uploader := &uploader{
		profile: profile,
	}

	if projectRoot != "" {
		err := uploader.loadIgnoresMatcher(filepath.Join(projectRoot, ".direktivignore"))
		if err != nil {
			return nil, err
		}
	}

	return uploader, nil
}

func (u *uploader) createDirectory(path string) error {
	if path == "." {
		return nil
	}

	fmt.Printf("creating directory %s\n", path)

	dir := fileObject{
		Typ: "directory",
	}

	err := u.createFileItem(path, "POST", dir)
	if err != nil && err.Error() == "filesystem path already exists" {
		return nil
	}

	return err
}

func (u *uploader) createFile(path, filePath string) error {
	fmt.Printf("creating file %s\n", path)
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	b64 := base64.StdEncoding.EncodeToString(b)

	obj := fileObject{
		Data: b64,
	}

	if strings.HasSuffix(path, ".direktiv.ts") {
		obj.Typ = string(filestore.FileTypeWorkflow)
		obj.MimeType = "application/x-typescript"
	} else if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {
		obj.MimeType = "application/yaml"

		resource, err := model.LoadResource(b)
		if errors.Is(err, model.ErrNotDirektivAPIResource) {
			obj.Typ = string(filestore.FileTypeFile)
		}

		switch resource.(type) {
		case *model.Workflow:
			obj.Typ = string(filestore.FileTypeWorkflow)
		case *core.EndpointFile:
			obj.Typ = string(filestore.FileTypeEndpoint)
		case *core.ConsumerFile:
			obj.Typ = string(filestore.FileTypeConsumer)
		case *core.ServiceFile:
			obj.Typ = string(filestore.FileTypeService)
		default:
			obj.Typ = string(filestore.FileTypeFile)
		}
	} else {
		obj.Typ = string(filestore.FileTypeFile)
		mt := mime.TypeByExtension(filepath.Ext(path))
		if mt == "" {
			mt = "text/plain"
		}
		obj.MimeType = mt
	}

	err = u.createFileItem(path, "POST", obj)
	if err != nil && err.Error() == "filesystem path already exists" {
		return u.createFileItem(path, "PATCH", obj)
	}

	return err
}

func (u *uploader) loadIgnoresMatcher(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var ps []gitignore.Pattern
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	u.matcher = gitignore.NewMatcher(ps)
	return nil
}

func (u *uploader) createFileItem(path, method string, obj fileObject) error {
	parent := path

	if method == "POST" {
		base := filepath.Base(path)
		parent = filepath.Dir(path)
		if parent == "." {
			parent = ""
		}
		obj.Name = base
	}

	// generate url
	url := fmt.Sprintf("%s/api/v2/namespaces/%s/files/%s", u.profile.Address, u.profile.Namespace, parent)

	fObj, err := json.MarshalIndent(obj, "", "   ")
	if err != nil {
		return err
	}

	resp, err := u.sendRequest(method, url, fObj)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var errJson errorResponse
		err = json.Unmarshal(b, &errJson)
		if err != nil {
			return err
		}

		return fmt.Errorf(errJson.Error.Message)
	}

	return nil
}

func (u *uploader) sendRequest(method, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if u.profile.Token != "" {
		req.Header.Add("Direktiv-Token", u.profile.Token)
	}

	req.Header.Add("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: u.profile.Insecure},
	}
	client := &http.Client{Transport: tr}
	return client.Do(req)
}
