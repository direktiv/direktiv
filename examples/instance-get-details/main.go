package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Contents of the "helloworld" workflow
const helloworldWorkflow = `
id: helloworld 
states:
- id: hello
  type: noop
  transform: '{ result: "Hello World!" }'`

const workflowInputSample = `
{
	"sample-data": "example"
}`

// creates and invokes the workflow 'helloworld' in the namespace 'example' with input data.
//	A get logs requests is then sent to the instance and the logs are printed.
func main() {
	var err error
	connString := "127.0.0.1:6666"

	conn, err := grpc.Dial(connString, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to grpc: %v", err)
	}

	// create client for grpc
	client := ingress.NewDirektivIngressClient(conn)
	ctx := context.Background()
	namespace := "example"
	workflowID := "helloworld"

	// prepare namespace, workflow, and instance

	// create namespace if not exist
	log.Printf("Creating namespace '%s'", namespace)
	if code, err := createNamespace(client, namespace); err != nil {
		if code != codes.AlreadyExists {
			log.Fatalf("Failed to create example namespace: %v", err)
		}
		log.Printf("Skipped creating namespace '%s' since it already exists", namespace)
	}

	// create workflow if not exist
	log.Printf("Creating workflow '%s'", workflowID)
	if code, err := createWorkflow(client, namespace, workflowID, []byte(helloworldWorkflow)); err != nil {
		if code != codes.AlreadyExists {
			log.Fatalf("Failed to create example workflow: %v", err)
		}
		log.Printf("Skipped creating workflow '%s' since it already exists", workflowID)
	}

	// invoke workflow
	log.Printf("Invoking workflow '%s/%s'", namespace, workflowID)
	instanceID, code, err := invokeWorkflow(client, namespace, workflowID, []byte(workflowInputSample))
	if err != nil {
		log.Fatalf("Failed to invoke instance on '%s/%s': %v %v", namespace, workflowID, code, err)
	}

	// Wait 5 seconds for logs
	log.Printf("Waiting 5 seconds for workflow to run")
	time.Sleep(5 * time.Second)

	// prepare request to get logs
	request := ingress.GetWorkflowInstanceRequest{
		Id: &instanceID,
	}

	// send grpc request
	log.Printf("Getting details from '%s'", instanceID)
	resp, err := client.GetWorkflowInstance(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal instance details response: %v", err)
	}

	log.Printf("Got Instance Details: ")
	fmt.Println(string(b))
	return
}

// util ...

// createWorkflow - creates the workflow 'workflow' in the namespace 'namespace' with the definition 'b' using a grpc request
func createWorkflow(client ingress.DirektivIngressClient, namespace string, workflow string, b []byte) (codes.Code, error) {
	ctx := context.Background()

	// prepare request
	request := ingress.AddWorkflowRequest{
		Namespace: &namespace,
		Workflow:  b,
	}

	// send grpc request
	_, err := client.AddWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return s.Code(), errors.New(s.Message())
	}

	return 0, nil
}

// createNamespace - creates the namespace 'namespace' using a grpc request
func createNamespace(client ingress.DirektivIngressClient, namespace string) (codes.Code, error) {
	ctx := context.Background()

	// prepare request
	request := ingress.AddNamespaceRequest{
		Name: &namespace,
	}

	// send grpc request
	_, err := client.AddNamespace(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return s.Code(), errors.New(s.Message())
	}

	return 0, nil

}

// invokeWorkflow - invokes the workflow 'workflow' in the namespace 'namespace' with the input 'b' using a grpc request
func invokeWorkflow(client ingress.DirektivIngressClient, namespace string, workflowID string, b []byte) (string, codes.Code, error) {
	ctx := context.Background()

	// prepare request
	request := ingress.InvokeWorkflowRequest{
		Namespace:  &namespace,
		WorkflowId: &workflowID,
		Input:      b,
	}

	// send grpc request
	resp, err := client.InvokeWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return "", s.Code(), errors.New(s.Message())
	}

	return *resp.InstanceId, 0, nil
}
