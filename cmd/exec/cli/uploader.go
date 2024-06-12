package cli

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type uploader struct {
	matcher gitignore.Matcher
	profile profile
}

type fileObject struct {
	Name     string `json:"name"`
	Data     string `json:"data,omitempty"`
	Typ      string `json:"type"`
	MimeType string `json:"mimeType,omitempty"`
	// {
	// 	"name": "sss.yaml",
	// 	"data": "ZGlyZWt0aXZfYXBpOiB3b3JrZmxvdy92MQpkZXNjcmlwdGlvbjogQSBzaW1wbGUgJ25vLW9wJyBzdGF0ZSB0aGF0IHJldHVybnMgJ0hlbGxvIHdvcmxkIScKc3RhdGVzOgotIGlkOiBoZWxsb3dvcmxkCiAgdHlwZTogbm9vcAogIHRyYW5zZm9ybToKICAgIHJlc3VsdDogSGVsbG8gd29ybGQhCg==",
	// 	"type": "workflow",
	// 	"mimeType": "application/yaml"
	//   }
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// {
// 	"error": {
// 		"code": "resource_already_exists",
// 		"message": "filesystem path already exists"
// 	}
// }

func newUploader(projectRoot string, profile profile) (*uploader, error) {

	uploader := &uploader{
		profile: profile,
	}

	err := uploader.loadIgnoresMatcher(filepath.Join(projectRoot, ".direktivignore"))
	if err != nil {
		return nil, err
	}

	return uploader, nil
}

func (u *uploader) createDirectory(path string) error {

	if path == "." || path == ".." {
		return nil
	}

	fmt.Printf("creating directory %s\n", path)

	base := filepath.Base(path)
	parent := filepath.Dir(path)

	if parent == "." {
		parent = ""
	}

	// generate url
	url := fmt.Sprintf("%s/api/v2/namespaces/%s/files/%s", u.profile.Address, u.profile.Namespace, parent)

	dir := fileObject{
		Typ:  "directory",
		Name: base,
	}

	dirObj, err := json.MarshalIndent(dir, "", "   ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dirObj))
	if err != nil {
		return err
	}

	if u.profile.Token != "" {
		req.Header.Add("Direktiv-Token", u.profile.Token)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: u.profile.Insecure},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
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

		// that is fine, if the dir exists
		if errJson.Error.Code != "resource_already_exists" {
			return fmt.Errorf(errJson.Error.Message)
		}
	}

	return nil

}

func (u *uploader) createFile(path, filePath string) error {

	fmt.Printf("creating file %s\n", path)

	base := filepath.Base(path)
	parent := filepath.Dir(path)

	if parent == "." {
		parent = ""
	}

	fmt.Printf("BASE %v\n", base)
	fmt.Printf("PARENT %v\n", parent)

	if strings.HasSuffix(path, ".direktiv.ts") {

	} else if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {


		// check if workflow
	} else {

	}

	return nil
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
