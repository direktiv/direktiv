package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Contents of the "helloworld" workflow
const helloworldWorkflow = `
id: helloworld 
states:
- id: hello
  type: noop
  transform: '{ result: "Hello World!" }'`

// This example gets the workflow "helloworld" in the namespace "example"
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

	// prepare request
	request := ingress.GetWorkflowByIdRequest{
		Namespace: &namespace,
		Id:        &workflowID,
	}

	// send grpc request
	log.Printf("Getting Workflow '%s/%s'", namespace, workflowID)
	resp, err := client.GetWorkflowById(ctx, &request)
	if err != nil {
		s := status.Convert(err)

		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal workflow response: %v", err)
	}

	log.Printf("Got Workflow Details: ")
	fmt.Println(string(b) + "\n")

	log.Printf("Workflow Contents: ")
	fmt.Println(strings.TrimSpace(string(resp.Workflow)))

	return
}
