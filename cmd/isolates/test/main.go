package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/grpc"
)

func main() {

	log.Infof("run client")

	external := false
	name := "testme"
	ns := ""
	wf := ""
	img := "vorteil/request:v6"

	info := &igrpc.BaseInfo{
		Name:      &name,
		Namespace: &ns,
		Workflow:  &wf,
		Image:     &img,
	}

	var g int32 = 1
	c := &igrpc.Config{
		MinScale: &g,
	}

	// Info     *BaseInfo `protobuf:"bytes,1,opt,name=info,proto3,oneof" json:"info,omitempty"`
	// Config   *Config   `protobuf:"bytes,2,opt,name=config,proto3,oneof" json:"config,omitempty"`
	// External *bool     `protobuf:"varint,3,opt,name=external,proto3,oneof" json:"external,omitempty"`

	sr := &igrpc.CreateIsolateRequest{
		Info:     info,
		Config:   c,
		External: &external,
		// Name:      &name,
		// External:  &external,
		// Namespace: &ns,
		// Workflow:  &wf,
	}
	// // StoreIsolate(ctx context.Context, in *StoreIsolateRequest, opts ...grpc.CallOption)
	// Name      *string           `protobuf:"bytes,1,opt,name=name,proto3,oneof" json:"name,omitempty"`
	// Namespace *string           `protobuf:"bytes,2,opt,name=namespace,proto3,oneof" json:"namespace,omitempty"`
	// Workflow  *string           `protobuf:"bytes,3,opt,name=workflow,proto3,oneof" json:"workflow,omitempty"`
	// Config    map[string]string `protobuf:"bytes,4,rep,name=config,proto3" json:"config,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Input     *string           `protobuf:"bytes,5,opt,name=input,proto3,oneof" json:"input,omitempty"`
	// Output    *string           `protobuf:"bytes,6,opt,name=output,proto3,oneof" json:"output,omitempty"`
	// External  *bool             `protobuf:"varint,7,opt,name=external,proto3,oneof" json:"external,omitempty"`
	// Revision  *string           `protobuf:"bytes,8,opt,name=revision,proto3,oneof" json:"revision,omitempty"`

	conn, err := grpc.Dial("127.0.0.1:30234", grpc.WithInsecure())
	if err != nil {
		log.Errorf("ERR %v", err)
	}
	defer conn.Close()

	client := igrpc.NewIsolatesServiceClient(conn)

	log.Infof("new client %v", client)

	_, err = client.CreateIsolate(context.Background(), sr)
	// _, err = client.UpdateIsolate(context.Background(), sr)

	if err != nil {
		log.Errorf("ERR %v", err)
	}

}
