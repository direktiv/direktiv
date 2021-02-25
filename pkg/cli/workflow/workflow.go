package workflow

import (
	"fmt"
	"io/ioutil"

	"github.com/vorteil/direktiv/pkg/cli/util"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Toggle enables or disables the workflow
func Toggle(conn *grpc.ClientConn, namespace, workflow string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	request := ingress.GetWorkflowByIdRequest{
		Namespace: &namespace,
		Id:        &workflow,
	}

	resp, err := client.GetWorkflowById(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	toggle := !*resp.Active

	uRequest := ingress.UpdateWorkflowRequest{
		Uid:      resp.Uid,
		Workflow: resp.Workflow,
		Active:   &toggle,
	}

	_, err = client.UpdateWorkflow(ctx, &uRequest)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	if toggle {
		return fmt.Sprintf("Enabled workflow '%s'", workflow), nil
	}

	return fmt.Sprintf("Disabled workflow '%s'", workflow), nil
}

// List returns an array of workflows for a given namespace
func List(conn *grpc.ClientConn, namespace string) ([]*ingress.GetWorkflowsResponse_Workflow, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	// prepare request
	request := ingress.GetWorkflowsRequest{
		Namespace: &namespace,
	}

	// send grpc request
	resp, err := client.GetWorkflows(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return resp.Workflows, nil
}

// Execute a workflow using the yaml provided
func Execute(conn *grpc.ClientConn, namespace string, id string, input string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	var err error
	var b []byte
	if input != "" {
		b, err = ioutil.ReadFile(input)
		if err != nil {
			return "", err
		}
	}

	// prepare request
	request := ingress.InvokeWorkflowRequest{
		Namespace:  &namespace,
		Input:      b,
		WorkflowId: &id,
	}

	// send grpc request
	resp, err := client.InvokeWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return fmt.Sprintf("Successfully invoked, Instance ID: %s", resp.GetInstanceId()), nil
}

// getWorkflowUID returns the UID of a workflow
func getWorkflowUID(conn *grpc.ClientConn, namespace, id string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()
	// prepare request
	request := ingress.GetWorkflowByIdRequest{
		Namespace: &namespace,
		Id:        &id,
	}

	// send grpc request
	resp, err := client.GetWorkflowById(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}
	return resp.GetUid(), nil
}

// Get returns a workflow definition in YAML format
func Get(conn *grpc.ClientConn, namespace string, id string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	// prepare request
	request := ingress.GetWorkflowByIdRequest{
		Namespace: &namespace,
		Id:        &id,
	}

	// send grpc request
	resp, err := client.GetWorkflowById(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return string(resp.GetWorkflow()), nil
}

// Update a workflow specified by ID.
func Update(conn *grpc.ClientConn, namespace string, id string, filepath string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	uid, err := getWorkflowUID(conn, namespace, id)
	if err != nil {
		return "", err
	}

	// prepare request
	request := ingress.UpdateWorkflowRequest{
		Uid:      &uid,
		Workflow: b,
	}

	// send grpc request
	resp, err := client.UpdateWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return fmt.Sprintf("Successfully updated '%s'", resp.GetId()), nil
}

// Delete an existing workflow.
func Delete(conn *grpc.ClientConn, namespace, id string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	uid, err := getWorkflowUID(conn, namespace, id)
	if err != nil {
		return "", err
	}

	// prepare request
	request := ingress.DeleteWorkflowRequest{
		Uid: &uid,
	}

	// send grpc request
	_, err = client.DeleteWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return fmt.Sprintf("Deleted workflow '%v'", id), nil
}

// Add creates a new workflow
func Add(conn *grpc.ClientConn, namespace string, filepath string) (string, error) {
	client, ctx, cancel := util.CreateClient(conn)
	defer cancel()

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// prepare request
	request := ingress.AddWorkflowRequest{
		Namespace: &namespace,
		Workflow:  b,
	}

	// send grpc request
	resp, err := client.AddWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
	}

	return fmt.Sprintf("Created workflow '%s'", resp.GetId()), nil
}
