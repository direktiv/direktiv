package main

import (
	"context"
	"log"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// This example deletes the namespace "example"
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

	// prepare request
	request := ingress.DeleteNamespaceRequest{
		Name: &namespace,
	}

	// send grpc request
	log.Printf("Deleting namespace '%s'", namespace)
	resp, err := client.DeleteNamespace(ctx, &request)
	if err != nil {
		s := status.Convert(err)

		log.Fatalf("GRPC request failed, status: '%s', message: '%s'", s.Code(), s.Message())
	}

	log.Printf("Deleted namespace '%s'", *resp.Name)
	return
}
