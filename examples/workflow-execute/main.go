package main

import (
	"context"
	"log"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// This example executes the workflow "helloworld" from the namespace "example"
//	This is the namespace/workflow that is created in the workflow-create example
//	A instance ID is logged at the end of this example, which can be used for getting logs later
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
	request := ingress.InvokeWorkflowRequest{
		Namespace:  &namespace,
		WorkflowId: &workflowID,
	}

	// send grpc request
	resp, err := client.InvokeWorkflow(ctx, &request)
	if err != nil {
		s := status.Convert(err)
		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	log.Printf("Invoked workflow, instance ID: %s'", *resp.InstanceId)
	return
}
