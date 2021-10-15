package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/pkg/flow"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func testStartWorkflow(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(simpleWorkflow),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if iresp.Instance.As != "/testwf" {
		return errors.New("instance returned incorrect 'As'")
	}

	if len(iresp.Flow) != 1 || iresp.Flow[0] != "a" {
		return errors.New("incorrect flow array")
	}

	input, err := c.InstanceInput(ctx, &grpc.InstanceInputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if len(input.Data) != 0 {
		return errors.New("instance input should have been zero length")
	}

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if string(output.Data) != `{"input":""}` {
		return errors.New("unexpected instance output")
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if len(logs.Edges) == 0 {
		return errors.New("missing instance logs")
	}

	return nil

}

func testStateLogSimple(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    log: "Hello, world!"
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Hello, world!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing simple instance log")
	}

	return nil

}

func testStateLogJQ(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    log: 'jq(.name)!'
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Direktiv!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testStateLogJQNested(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    log: 'Hello, jq(.name)!'
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Hello, Direktiv!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testStateLogJQObject(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    log: 
      Name: 'jq(.name)'
      Constant: 5
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `{
  "Constant": 5,
  "Name": "Direktiv"
}` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testInstanceEventAnd(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
- id: eventstart
  type: eventAnd
  events:
  - type: a-checked
  - type: b-checked
  transition: a
- id: a
  type: noop
  transform:
    execute: a
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	// sleep for workflow to start
	time.Sleep(time.Second * 1)

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "a", uuid.New().String(), "a")),
	})
	if err != nil {
		return err
	}

	// go is to fast sleep for event broadcast
	time.Sleep(time.Second * 1)

	// check for instance see if its still pending after one event
	instance, err := c.Instance(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if instance.GetInstance().GetStatus() != "pending" {
		return errors.New("eventAnd ended before the second event was sent")
	}

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "b", uuid.New().String(), "b")),
	})
	if err != nil {
		return err
	}

	// go is to fast sleep for event broadcast
	time.Sleep(time.Second * 1)

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	data := output.GetData()

	var test map[string]string

	err = json.Unmarshal(data, &test)
	if err != nil {
		return err
	}

	if test["execute"] == "a" {
		return nil
	}

	return errors.New("eventAnd state did not end properly for expected output")
}

func testInstanceParallel(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	// create workflow that runs two actions one that will fail via post request using and mode but successful on or mode
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwffailed",
		Source: []byte(`
functions:
- id: get
  image: vorteil/request:v10
  type: reusable
states:
- id: runpara
  type: parallel
  actions:
  - function: get
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
  - function: get
    input:
      method: "GET"
      url: "jsonplaceholder.typicode.com/todos/1"
  mode: and
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwfsuccess",
		Source: []byte(`
functions:
- id: get
  image: vorteil/request:v10
  type: reusable
states:
- id: runpara
  type: parallel
  actions:
  - function: get
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
  - function: get
    input:
      method: "GET"
      url: "jsonplaceholder.typicode.com/todos/1"
  mode: or
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwfsuccessand",
		Source: []byte(`
functions:
- id: get
  image: vorteil/request:v10
  type: reusable
states:
- id: runpara
  type: parallel
  actions:
  - function: get
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
  - function: get
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
  mode: and
`),
	})
	if err != nil {
		return err
	}

	respFailed, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwffailed",
	})
	if err != nil {
		return err
	}

	respSuccessAnd, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwfsuccessand",
	})
	if err != nil {
		return err
	}

	respSuccess, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwfsuccess",
	})
	if err != nil {
		return err
	}

	client, err := c.InstanceStream(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  respSuccessAnd.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	// if workflow thats meant to be successful fails
	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("parallel instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	client, err = c.InstanceStream(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  respSuccess.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	// if workflow thats meant to be successful fails
	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("parallel instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	client, err = c.InstanceStream(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  respFailed.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	// if workflow thats meant to fail fails
	if iresp.Instance.Status == flow.StatusFailed {
		return nil
	}

	return errors.New("'and' or 'or' mode didnt not finish properly")
}

func testInstanceForeach(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
functions:
- id: get
  image: vorteil/request:v10
  type: reusable
states:
- id: fe
  type: foreach
  array: 'jq(.x[] | { xp: . })'
  action:
    function: get
    input:
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"x":["0", "1", "2"]}`),
	})
	if err != nil {
		return err
	}

	client, err := c.InstanceStream(ctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("foreach instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil
}

func testInstanceValidate(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
- id: validate-email
  type: validate
  subject: jq(.)
  schema:
    type: object
    properties:
      email:
        type: string
        format: email
  catch:
  - error: direktiv.schema.*
    transition: email-not-valid
  transition: email-valid
- id: email-not-valid
  type: noop
  transform:
    result: "Email is not valid."
- id: email-valid
  type: noop
  transform:
    result: "Email is valid."
`),
	})
	if err != nil {
		return err
	}

	// this should be valid
	respValid, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"email": "trent.hilliam@vorteil.io"}`),
	})
	if err != nil {
		return err
	}

	respNotValid, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"email": "trent.hilliamvorteil.io"}`),
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 1)

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  respValid.Instance,
	})
	if err != nil {
		return err
	}
	data := output.GetData()

	var testA map[string]string

	err = json.Unmarshal(data, &testA)
	if err != nil {
		return err
	}

	output, err = c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  respNotValid.Instance,
	})
	if err != nil {
		return err
	}
	data = output.GetData()

	var testB map[string]string

	err = json.Unmarshal(data, &testB)
	if err != nil {
		return err
	}

	if testA["result"] == "Email is valid." && testB["result"] == "Email is not valid." {
		return nil
	}

	return errors.New("validate state failed could not verify an email")
}

func testInstanceEventXor(ctx context.Context, c grpc.FlowClient, namespace string) error {
	a := false
	b := false

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
- id: eventstart
  type: eventXor
  events:
  - event:
      type: a-checked
    transition: a
  - event:
      type: b-checked
    transition: b
- id: a
  type: noop
  transform:
    execute: a
- id: b
  type: noop
  transform:
    execute: b
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	// sleep for workflow to start
	time.Sleep(time.Second * 1)

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "a", uuid.New().String(), "a")),
	})
	if err != nil {
		return err
	}

	// go is to fast sleep for event broadcast
	time.Sleep(time.Second * 1)

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	data := output.GetData()

	var testA map[string]string

	err = json.Unmarshal(data, &testA)
	if err != nil {
		return err
	}

	// a ran fine
	if testA["execute"] == "a" {
		a = true
	}

	resp, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	// go is to fast sleep for event broadcast
	time.Sleep(time.Second * 1)

	_, err = c.BroadcastCloudevent(ctx, &grpc.BroadcastCloudeventRequest{
		Namespace:  namespace,
		Cloudevent: []byte(fmt.Sprintf(aCheckedCloudEvent, "b", uuid.New().String(), "b")),
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 1)

	output, err = c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	data = output.GetData()

	var testB map[string]string

	err = json.Unmarshal(data, &testB)
	if err != nil {
		return err
	}

	if testB["execute"] == "b" {
		b = true
	}

	if a && b {
		return nil
	}

	return errors.New("eventXor state was not handled properly during execution")
}

func testInstanceSimpleChain(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transition: b
  - id: b
    type: noop 
    transition: c
  - id: c
    type: noop 
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 3 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "c" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceSwitchLoop(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transform: 
      k: 5
    transition: b
  - id: b
    type: switch 
    conditions:
      - condition: 'jq(.k > 0)'
        transition: b
        transform: 'jq(.k -= 1)'
    defaultTransition: c
  - id: c
    type: noop 
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 8 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "b" ||
		iresp.Flow[3] != "b" || iresp.Flow[4] != "b" || iresp.Flow[5] != "b" ||
		iresp.Flow[6] != "b" || iresp.Flow[7] != "c" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceDelayLoop(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop
    transform:
      currentTime: jq(now)
    transition: b
  - id: b
    type: delay
    duration: PT8S
    transition: c
  - id: c
    type: noop
    transform:
      difference: jq(now - .currentTime)
    transition: d
  - id: d
    type: switch
    conditions:
      - condition: jq(.difference < 8)
        transition: fail
    defaultTransition: e
  - id: e
    type: noop 
  - id: fail
    type: noop 
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 5 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "c" ||
		iresp.Flow[3] != "d" || iresp.Flow[4] != "e" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceGenerateConsumeEvent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-consume",
		Source: []byte(`
states:
- id: a
  type: consumeEvent
  transition: b
  timeout: PT10S
  event:
    type: testcloudevent
- id: b
  type: noop
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-generate",
		Source: []byte(`
states:
- id: a
  type: generateEvent
  event:
    type: testcloudevent
    source: Direktiv
    data: 
      message: "helloworld"
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-consume",
	})
	if err != nil {
		return err
	}

	_, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-generate",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceTimeoutKill(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout-subflow",
		Source: []byte(`
timeouts: 
  kill: PT2S
states:
  - id: a
    type: delay
    duration: PT50S
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
		Source: []byte(fmt.Sprintf(`
functions:
- id: sub
  type: subflow
  workflow: testwf-timeout-subflow
states:
- id: a 
  type: action
  action:
    function: sub
  catch:
    - error: "%s"
      transition: b
- id: b
  type: noop
`, flow.ErrCodeHardTimeout)),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	instances, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	inst := new(grpc.Instance)
	for _, edge := range instances.GetInstances().GetEdges() {
		if strings.TrimPrefix(edge.GetNode().GetAs(), "/") == "testwf-timeout-subflow" {
			inst = edge.GetNode()
		}
	}

	if inst == nil {
		return errors.New("testwf-timeout-subflow instance not found")
	}

	if inst.Status != flow.StatusFailed {
		return fmt.Errorf("instance expected to be %s but was %s", flow.StatusFailed, inst.Status)
	}

	if inst.ErrorCode != flow.ErrCodeHardTimeout {
		return fmt.Errorf("instance error code expected to be %s but was %s", flow.ErrCodeHardTimeout, inst.ErrorCode)
	}

	return nil

}

func testInstanceTimeoutKillLong(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout-subflow",
		Source: []byte(`
timeouts: 
  kill: PT60S
states:
  - id: a
    type: delay
    duration: PT80S
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
		Source: []byte(fmt.Sprintf(`
functions:
- id: sub
  type: subflow
  workflow: testwf-timeout-subflow
states:
- id: a 
  type: action
  action:
    function: sub
  catch:
    - error: "%s"
      transition: b
- id: b
  type: noop
`, flow.ErrCodeHardTimeout)),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	instances, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	inst := new(grpc.Instance)
	for _, edge := range instances.GetInstances().GetEdges() {
		if strings.TrimPrefix(edge.GetNode().GetAs(), "/") == "testwf-timeout-subflow" {
			inst = edge.GetNode()
		}
	}

	if inst == nil {
		return errors.New("testwf-timeout-subflow instance not found")
	}

	if inst.Status != flow.StatusFailed {
		return fmt.Errorf("instance expected to be %s but was %s", flow.StatusFailed, inst.Status)
	}

	if inst.ErrorCode != flow.ErrCodeHardTimeout {
		return fmt.Errorf("instance error code expected to be %s but was %s", flow.ErrCodeHardTimeout, inst.ErrorCode)
	}

	return nil

}

func testInstanceTimeoutInterrupt(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout-subflow",
		Source: []byte(`
timeouts: 
  interrupt: PT2S
states:
  - id: a
    type: delay
    duration: PT50S
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
		Source: []byte(fmt.Sprintf(`
functions:
- id: sub
  type: subflow
  workflow: testwf-timeout-subflow
states:
- id: a 
  type: action
  action:
    function: sub
  catch:
    - error: "%s"
      transition: b
- id: b
  type: noop
`, flow.ErrCodeSoftTimeout)),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	instances, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	inst := new(grpc.Instance)
	for _, edge := range instances.GetInstances().GetEdges() {
		if strings.TrimPrefix(edge.GetNode().GetAs(), "/") == "testwf-timeout-subflow" {
			inst = edge.GetNode()
		}
	}

	if inst == nil {
		return errors.New("testwf-timeout-subflow instance not found")
	}

	if inst.Status != flow.StatusFailed {
		return fmt.Errorf("instance expected to be %s but was %s", flow.StatusFailed, inst.Status)
	}

	if inst.ErrorCode != flow.ErrCodeSoftTimeout {
		return fmt.Errorf("instance error code expected to be %s but was %s", flow.ErrCodeSoftTimeout, inst.ErrorCode)
	}

	return nil

}

func testInstanceTimeoutInterruptLong(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout-subflow",
		Source: []byte(`
timeouts: 
  interrupt: PT60S
states:
  - id: a
    type: delay
    duration: PT80S
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
		Source: []byte(fmt.Sprintf(`
functions:
- id: sub
  type: subflow
  workflow: testwf-timeout-subflow
states:
- id: a 
  type: action
  action:
    function: sub
  catch:
    - error: "%s"
      transition: b
- id: b
  type: noop
`, flow.ErrCodeSoftTimeout)),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-timeout",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	instances, err := c.Instances(ctx, &grpc.InstancesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	inst := new(grpc.Instance)
	for _, edge := range instances.GetInstances().GetEdges() {
		if strings.TrimPrefix(edge.GetNode().GetAs(), "/") == "testwf-timeout-subflow" {
			inst = edge.GetNode()
		}
	}

	if inst == nil {
		return errors.New("testwf-timeout-subflow instance not found")
	}

	if inst.Status != flow.StatusFailed {
		return fmt.Errorf("instance expected to be %s but was %s", flow.StatusFailed, inst.Status)
	}

	if inst.ErrorCode != flow.ErrCodeSoftTimeout {
		return fmt.Errorf("instance error code expected to be %s but was %s", flow.ErrCodeSoftTimeout, inst.ErrorCode)
	}

	return nil

}
