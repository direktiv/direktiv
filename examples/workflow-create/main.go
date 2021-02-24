package main

import (
	"context"
	"errors"
	"log"

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

// This example creates the workflow "helloworld" in the namespace "example"
//	If the namespace "example" does not exist, it will b created
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

	// workflows require a namespace, so create example namespace if not exists
	log.Printf("Creating namespace '%s'", namespace)
	if code, err := createNamespace(client, namespace); err != nil {
		if code != codes.AlreadyExists {
			log.Fatalf("Failed to create example namespace: %v", err)
		}
		log.Printf("Skipped creating namespace '%s' since it already exists", namespace)
	}

	// prepare request
	request := ingress.AddWorkflowRequest{
		Namespace: &namespace,
		Workflow:  []byte(helloworldWorkflow),
	}

	// send grpc request
	log.Printf("Creating Workflow '%s'", "helloworld")
	resp, err := client.AddWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)

		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	log.Printf("Created workflow '%s' in '%s' namespace", resp.GetId(), namespace)
	return
}

// createNamespace - creates the namespace 'ns' using a grpc request
func createNamespace(client ingress.DirektivIngressClient, ns string) (codes.Code, error) {
	ctx := context.Background()

	// prepare request
	request := ingress.AddNamespaceRequest{
		Name: &ns,
	}

	// send grpc request
	_, err := client.AddNamespace(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return s.Code(), errors.New(s.Message())
	}

	return 0, nil

}
