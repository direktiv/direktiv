package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vorteil/direktiv/pkg/isolate"
	"google.golang.org/grpc"
)

func main() {

	var err error
	connString := "127.0.0.1:8888"

	if len(os.Args) > 1 {
		connString = os.Args[1]
	}

	fmt.Printf("connecting to %s", connString)

	conn, err := grpc.Dial(connString, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to grpc: %v", err)
	}

	// create client for grpc
	client := isolate.NewDirektivIsolateClient(conn)
	ctx := context.Background()
	namespace := "example"

	var s int32
	img := "vorteil/request:v1"
	id := "randomActionId"

	data := `{
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/todos/1"
    }`

	_, err = client.RunIsolate(ctx, &isolate.RunIsolateRequest{
		ActionId:   &id,
		Namespace:  &namespace,
		InstanceId: &id,
		Image:      &img,
		Size:       &s,
		Data:       []byte(data),
	})

	fmt.Printf("error: %v", err)

}
