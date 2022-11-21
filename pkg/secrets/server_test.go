package secrets

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

var ip string
var namespace string

type GetSecretsBody struct {
	Namespace string `json:"namespace"`
	Secrets   struct {
		PageInfo struct {
			Order  []interface{} `json:"order"`
			Filter []interface{} `json:"filter"`
			Limit  int           `json:"limit"`
			Offset int           `json:"offset"`
			Total  int           `json:"total"`
		} `json:"pageInfo"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	} `json:"secrets"`
}

func init() {
	ip = "192.168.2.81"
	namespace = "test999"
	//DELETE NAMESPACE
	DeleteNamespace(ip, namespace)
	//CREATE NAMESPACE
	CreateNamespace(ip, namespace)
}

// StoreSecret stores secrets in backends
func TestStoreSecretAndGetSecrets(t *testing.T) {
	CreateNamespace(ip, namespace)
	secretsList := []string{"test1", "test2", "test3", "test4", "a/b/c/d", "a/b/c/e", "a/b/c", "a/b/c/z", "a/b/c/d/e"}
	FillDatabase(ip, namespace, secretsList)
	//check if StoreSecret stores secrets and folder correct
	//check no root level
	body := GetSecrets(ip, namespace, "")
	bodyNames := BodyToList(body)
	expected := []string{"a/", "test1", "test2", "test3", "test4"}
	assert.Equal(t, bodyNames, expected)
	//check empty Folder
	body = GetSecrets(ip, namespace, "a/")
	bodyNames = BodyToList(body)
	expected = []string{"a/", "a/b/"}
	assert.Equal(t, bodyNames, expected)
	//check secrets inside folderS
	body = GetSecrets(ip, namespace, "a/b/c/")
	bodyNames = BodyToList(body)
	expected = []string{"a/b/c/", "a/b/c/d", "a/b/c/d/", "a/b/c/e", "a/b/c/z"}
	assert.Equal(t, bodyNames, expected)
	//check for calling createSecret with suffix slash -> this case will not come cause the http handler in flow will recognise if
	//request have slash or not and forward directly to right function

	//clean database
	DeleteNamespace(ip, namespace)

}

// AddFolder stores folders and create all missing folders in the path
func TestAddFolder(t *testing.T) {
	CreateNamespace(ip, namespace)
	secretsList := []string{"a/", "b/", "c/", "a/b/c/d/e"}
	FillDatabase(ip, namespace, secretsList)
	//check if folders created in root level
	body := GetSecrets(ip, namespace, "")
	bodyNames := BodyToList(body)
	expected := []string{"a/", "b/", "c/"}
	assert.Equal(t, bodyNames, expected)

	//check if folders are autoCreated
	body = GetSecrets(ip, namespace, "a/")
	bodyNames = BodyToList(body)
	expected = []string{"a/", "a/b/"}
	assert.Equal(t, bodyNames, expected)

	body = GetSecrets(ip, namespace, "a/b/")
	bodyNames = BodyToList(body)
	expected = []string{"a/b/", "a/b/c/"}
	assert.Equal(t, bodyNames, expected)

	body = GetSecrets(ip, namespace, "a/b/c/")
	bodyNames = BodyToList(body)
	expected = []string{"a/b/c/", "a/b/c/d/"}
	assert.Equal(t, bodyNames, expected)

	//clean database
	DeleteNamespace(ip, namespace)
}

// DeleteFolder deletes folder from backend
func TestDeleteFolder(t *testing.T) {
	CreateNamespace(ip, namespace)

	//check delete highest level folder
	folderAndSecretsList := []string{"a/b/c/d/", "a/b/a", "a/b/b", "a/b/c"}
	FillDatabase(ip, namespace, folderAndSecretsList)

	folder := "a/b/c/d/"
	DeleteSecretOrFolder(ip, namespace, folder)
	body := GetSecrets(ip, namespace, "a/b/c/")
	bodyNames := BodyToList(body)
	expected := []string{"a/b/c/"}
	assert.Equal(t, bodyNames, expected)

	//check delete folder which includes secrets
	folder = "a/b/"
	DeleteSecretOrFolder(ip, namespace, folder)
	body = GetSecrets(ip, namespace, "a/")
	bodyNames = BodyToList(body)
	expected = []string{"a/"}
	assert.Equal(t, bodyNames, expected)

	//clean database
	DeleteNamespace(ip, namespace)
}

// DeleteSecret deletes secret from backend
func TestDeleteSecret(t *testing.T) {
	CreateNamespace(ip, namespace)

	//check delete not existing secret
	respStatusCode, _ := DeleteSecretOrFolder(ip, namespace, "notExistingSecret")
	assert.Equal(t, 404, respStatusCode)

	//check
	folderAndSecretsList := []string{"a/b/c/d", "a/b/c/e", "a/b/c/d/", "secret"}
	FillDatabase(ip, namespace, folderAndSecretsList)

	//check if secret deleted in root level
	secret := "secret"
	DeleteSecretOrFolder(ip, namespace, secret)
	body := GetSecrets(ip, namespace, "")
	bodyNames := BodyToList(body)
	expected := []string{"a/"}
	assert.Equal(t, bodyNames, expected)

	// check if secret is deleted in folder
	secret = "a/b/c/e"
	DeleteSecretOrFolder(ip, namespace, secret)
	body = GetSecrets(ip, namespace, "a/b/c/")
	bodyNames = BodyToList(body)
	expected = []string{"a/b/c/", "a/b/c/d", "a/b/c/d/"}
	assert.Equal(t, bodyNames, expected)

	//clean database
	DeleteNamespace(ip, namespace)

}

func TestOverwriteSecret(t *testing.T) {

	//check overwrite not existing secret
	respStatusCode, _ := OverwriteSecret(ip, namespace, "hallo")
	assert.Equal(t, 404, respStatusCode)

}

// SearchSecret search for secrets anf folder per name
func TestSearchSecret(t *testing.T) {
	CreateNamespace(ip, namespace)
	secretsList := []string{"test1", "test2", "test3", "test4", "a/b/c/test", "a/test/c/e", "a/b/test/", "a/b/c/z", "test/b/c/d/e", "secret", "a/b/secret", "secret/"}
	FillDatabase(ip, namespace, secretsList)

	body := SearchSecret(ip, namespace, "test")
	bodyNames := BodyToList(body)
	expected := []string{"a/b/c/test", "a/b/test/", "a/test/", "a/test/c/", "a/test/c/e", "test/", "test/b/", "test/b/c/", "test/b/c/d/", "test/b/c/d/e", "test1", "test2", "test3", "test4"}
	assert.Equal(t, bodyNames, expected)

	//clean database
	DeleteNamespace(ip, namespace)

}

//////////////////HELPER FUNCTIONS//////////////////////////////////
func FillDatabase(ip string, namespace string, list []string) {
	for _, secretName := range list {
		AddSecretOrFolder(ip, namespace, secretName)
	}

}

func CreateNamespace(ip string, name string) error {
	url := "http://" + ip + "/api/namespaces/" + name
	req, _ := http.NewRequest("PUT", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error ", resp.StatusCode)
		return err
	}
	return nil
}

func AddSecretOrFolder(ip string, namespace string, name string) error {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/secrets/" + name
	bodyS := ""
	if !strings.HasSuffix(name, "/") {
		bodyS = "SOMETHING"
	}
	body := strings.NewReader(bodyS)
	req, _ := http.NewRequest("PUT", url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error ", resp.StatusCode)
		return err
	}
	return nil
}

func DeleteSecretOrFolder(ip string, namespace string, name string) (int, error) {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/secrets/" + name
	bodyS := ""
	if !strings.HasSuffix(name, "/") {
		bodyS = "SOMETHING"
	}
	body := strings.NewReader(bodyS)
	req, _ := http.NewRequest("DELETE", url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error ", resp.StatusCode)
		return -1, err
	}
	return resp.StatusCode, err
}

func GetSecrets(ip string, namespace string, folder string) GetSecretsBody {
	path := folder
	if path != "" {
		path = "/" + path
	}
	url := "http://" + ip + "/api/namespaces/" + namespace + "/secrets" + path
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	b, _ := ioutil.ReadAll(resp.Body)
	var body GetSecretsBody
	_ = json.Unmarshal(b, &body)
	return body
}

func DeleteNamespace(ip string, namespace string) error {
	url := "http://" + ip + "/api/namespaces/" + namespace
	req, _ := http.NewRequest("DELETE", url, nil)
	_, err := http.DefaultClient.Do(req)

	return err
}

func OverwriteSecret(ip string, namespace string, secret string) (int, error) {
	url := "http://" + ip + "/api/namespaces/overwrite" + namespace + "/secrets/" + secret
	req, _ := http.NewRequest("PUT", url, nil)
	resp, err := http.DefaultClient.Do(req)
	b, _ := ioutil.ReadAll(resp.Body)
	var body GetSecretsBody
	_ = json.Unmarshal(b, &body)
	return resp.StatusCode, err
}

func BodyToList(body GetSecretsBody) []string {
	var bodySecrets []string
	for _, name := range body.Secrets.Results {
		bodySecrets = append(bodySecrets, name.Name)
	}
	return bodySecrets
}

func SearchSecret(ip string, namespace string, name string) GetSecretsBody {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/search/secrets/" + name
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	b, _ := ioutil.ReadAll(resp.Body)
	var body GetSecretsBody
	_ = json.Unmarshal(b, &body)
	return body
}
