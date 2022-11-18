package flow

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

var ip string
var namespace string

var script string

func init() {
	ip = "192.168.0.197"
	namespace = "test999"
	//DELETE NAMESPACE
	DeleteNamespace(ip, namespace)
	//CREATE NAMESPACE
	CreateNamespace(ip, namespace)

	//jsCode for froping event just return null for modifying return event
	script = `
	if (event["source"] == "drop"){
		return null
	}else{
		event["source"] = "newSource"
		return event
	}
	`
}

// CreateCloudEventFilter stores cloudEventFilter in backend
func TestCreateCloudEventFilterAndDeleteCloudEventfilter(t *testing.T) {

	//create new eventCloudFilter
	statusCode, _ := CreateCloudEventFilter(ip, namespace, "filter", script)
	assert.Equal(t, statusCode, http.StatusOK)

	//delete eventCloudFilter
	statusCode, _ = DeleteCloudEventFilter(ip, namespace, "filter")
	assert.Equal(t, statusCode, http.StatusOK)

	//clean database
	DeleteNamespace(ip, namespace)
}

// UpdateCloudEventFilter update cloudEventFilter in backend
func TestUpdateCloudEventFilter(t *testing.T) {

	//create new eventCloudFilter
	statusCode, _ := CreateCloudEventFilter(ip, namespace, "filter", script)
	assert.Equal(t, statusCode, http.StatusOK)

	//update existing eventCloudFilter
	newScript := `
	event["source"] = "newSource"
	return event
	`
	statusCode, _ = UpdateCloudEventFilter(ip, namespace, "filter", newScript)
	assert.Equal(t, statusCode, http.StatusOK)

	//clean database
	DeleteNamespace(ip, namespace)

}

// ApplyCloudEventFilter aplly filter on given cloudevent and drop or modified it
func TestApplyCloudEventFilter(t *testing.T) {

	//create new eventCloudFilter
	statusCode, _ := CreateCloudEventFilter(ip, namespace, "filter", script)
	assert.Equal(t, statusCode, http.StatusOK)

	//apply eventcloudFilter and broadcast it -> modfiying event
	var event = []byte(`{"specversion":"1.0", "type":"test", "source":"notDrop", "id":"12345678910"}`)
	statusCode, _ = ApplyCloudEventFilter(ip, namespace, "filter", event)
	assert.Equal(t, statusCode, http.StatusOK)

	//apply eventcloudFilter and broadcast it -> drop event
	event = []byte(`{"specversion":"1.0", "type":"test", "source":"drop", "id":"12345678910"}`)
	statusCode, _ = ApplyCloudEventFilter(ip, namespace, "filter", event)
	assert.Equal(t, statusCode, http.StatusOK)

	//DeleteNamespace(ip, namespace)

}

func CreateCloudEventFilter(ip string, namespace string, filterName string, jsCode string) (int, error) {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/eventfilter/" + filterName
	bodyS := jsCode
	body := strings.NewReader(bodyS)
	req, _ := http.NewRequest("PUT", url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, err
}

func DeleteCloudEventFilter(ip string, namespace string, filterName string) (int, error) {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/eventfilter/" + filterName
	bodyS := ""
	body := strings.NewReader(bodyS)
	req, _ := http.NewRequest("DELETE", url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, err
}

func UpdateCloudEventFilter(ip string, namespace string, filterName string, jsCode string) (int, error) {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/eventfilter/" + filterName
	bodyS := jsCode
	body := strings.NewReader(bodyS)
	req, _ := http.NewRequest("PATCH", url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, err
}

func ApplyCloudEventFilter(ip string, namespace string, filterName string, event []byte) (int, error) {
	url := "http://" + ip + "/api/namespaces/" + namespace + "/broadcast/" + filterName
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(event))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, err
}

func DeleteNamespace(ip string, namespace string) error {
	url := "http://" + ip + "/api/namespaces/" + namespace
	req, _ := http.NewRequest("DELETE", url, nil)
	_, err := http.DefaultClient.Do(req)

	return err
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
