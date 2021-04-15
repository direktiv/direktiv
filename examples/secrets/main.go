package main

import (
	"context"
	"fmt"
	"log"
	"os"

	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"google.golang.org/grpc"
)

// This example creates the workflow "helloworld" in the namespace "example"
//	If the namespace "example" does not exist, it will b created
func main() {

	ns := os.Args[1]
	name := os.Args[2]
	data := os.Args[3]

	var err error
	connString := "127.0.0.1:2610"

	conn, err := grpc.Dial(connString, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to grpc: %v", err)
	}

	// create client for grpc
	client := secretsgrpc.NewSecretsServiceClient(conn)

	ssr := &secretsgrpc.SecretsStoreRequest{
		Namespace: &ns,
		Name:      &name,
		Data:      []byte(data),
	}

	_, err = client.StoreSecret(context.Background(), ssr)
	fmt.Printf("Add Secret: %v\n", err)

	srr := &secretsgrpc.SecretsRetrieveRequest{
		Namespace: &ns,
		Name:      &name,
	}

	sec, err := client.RetrieveSecret(context.Background(), srr)

	if err != nil {
		fmt.Printf("Get Secret: %v\n", err)
	} else {
		fmt.Printf("Get Secret: %v %v\n", string(sec.Data), err)
	}

	getSecrets(client, ns)

	sdr := &secretsgrpc.SecretsDeleteRequest{
		Namespace: &ns,
		Name:      &name,
	}

	_, err = client.DeleteSecret(context.Background(), sdr)
	fmt.Printf("Delete Secret: %v\n", err)

	for i := 1; i <= 5; i++ {

		name := fmt.Sprintf("name%d", i)
		ssr.Name = &name

		_, err = client.StoreSecret(context.Background(), ssr)
		fmt.Printf("Add Secret: %v\n", err)

	}

	getSecrets(client, ns)

	dr := &secretsgrpc.DeleteSecretsRequest{
		Namespace: &ns,
	}

	_, err = client.DeleteSecrets(context.Background(), dr)
	fmt.Printf("Delete Secrets: %v\n", err)

}

func getSecrets(client secretsgrpc.SecretsServiceClient, ns string) {

	srrs := &secretsgrpc.GetSecretsRequest{
		Namespace: &ns,
	}

	names, err := client.GetSecrets(context.Background(), srrs)

	if err != nil {
		fmt.Printf("Get Secret: %v\n", err)
	} else {
		fmt.Printf("Get Secret: %v %v\n", names, err)
	}

}
