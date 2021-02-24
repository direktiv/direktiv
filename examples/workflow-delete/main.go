package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// This example deletes the workflow "helloworld" from the namespace "example"
//	This is the namespace/workflow that is created in the workflow-create example
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

	// get workflow uid
	uid, err := getWorkflowUIDFromID(client, "helloworld", "example")
	if err != nil {
		log.Fatalf("could not get workflow uid of 'example/helloworld', error: %v", err)
	}

	// prepare request
	request := ingress.DeleteWorkflowRequest{
		Uid: uid,
	}

	// send grpc request
	_, err = client.DeleteWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	log.Printf("Deleted workflow - ID: %s, UID: %s ", "helloworld", *uid)
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
