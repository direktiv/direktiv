package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Contents of the "helloworld" workflow
const helloworldWorkflowUpdate = `
id: helloworld 
description: "Updated Description"
states:
- id: hello
  type: noop
  transform: '{ result: "Hello World!" }'`

// This example updates the workflow "helloworld" in the namespace "example"
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

	// get workflow uid
	uid, err := getWorkflowUIDFromID(client, workflowID, namespace)
	if err != nil {
		log.Fatalf("could not get workflow uid of 'example/helloworld', error: %v", err)
	}
	active := true

	// prepare request
	request := ingress.UpdateWorkflowRequest{
		Uid:      uid,
		Active:   &active,
		Workflow: []byte(helloworldWorkflowUpdate),
	}

	// send grpc request
	resp, err := client.UpdateWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)

		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	log.Printf("Updated workflow '%s' in '%s' namespace", resp.GetId(), namespace)
	return
}

// getWorkflowUIDFromID - Gets the uid from a workflow in a namespace
func getWorkflowUIDFromID(client ingress.DirektivIngressClient, wID string, ns string) (*string, error) {
	ctx := context.Background()

	// prepare request
	request := ingress.GetWorkflowByIdRequest{
		Namespace: &ns,
		Id:        &wID,
	}

	// send grpc request
	resp, err := client.GetWorkflowById(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		return nil, fmt.Errorf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	return resp.Uid, nil
}
