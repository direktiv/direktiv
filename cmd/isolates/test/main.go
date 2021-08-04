package main

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/grpc"
)

func main() {

	log.Infof("run client")

	img := "vorteil/request:v5"
	var sz int32 = 2
	info := &igrpc.BaseInfo{
		Image: &img,
		Size:  &sz,
	}

	conn, err := grpc.Dial("127.0.0.1:30234", grpc.WithInsecure())
	if err != nil {
		log.Errorf("ERR %v", err)
	}
	defer conn.Close()

	client := igrpc.NewIsolatesServiceClient(conn)

	svn := "w-8829097305702293016"
	sr := igrpc.UpdateIsolateRequest{
		Info:        info,
		ServiceName: &svn,
	}

	_, err = client.UpdateIsolate(context.Background(), &sr)
	if err != nil {
		log.Errorf("ERR %v", err)
	}

	// a := make(map[string]string)
	// a["direktiv.io/workflow"] = "dsdsdsssd"
	//
	// g2 := igrpc.ListIsolatesRequest{
	// 	Annotations: a,
	// }
	//
	// _, err = client.ListIsolates(context.Background(), &g2)
	// log.Infof("new client %v", client)

	g2 := igrpc.GetIsolateRequest{
		ServiceName: &svn,
	}
	items, err := client.GetIsolate(context.Background(), &g2)
	if err != nil {
		log.Errorf(">> %v", err)
	}

	b1, err := json.MarshalIndent(items, "", "    ")
	if err != nil {
		log.Errorf("error marshalling new services: %v", err)
	}
	fmt.Printf("%s", string(b1))

	// log.Infof("new client %v", client)

	// _, err = client.CreateIsolate(context.Background(), sr)
	// _, err = client.UpdateIsolate(context.Background(), sr)

	// log.Infof("%v", items)

}
