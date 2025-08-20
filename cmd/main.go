package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/ChannelMeter/iso8601duration"
	_ "github.com/coreos/go-oidc/v3/oidc"
	"github.com/direktiv/direktiv/cmd/cli"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	_ "github.com/hashicorp/golang-lru/v2"
	"github.com/nats-io/nats.go"
)

func main() {
	testNatsConnection()
	cli.Run()
}

// TODO: remove this devCode.
func testNatsConnection() {
	var nc *nats.Conn
	var err error
	for i := 0; i < 50; i++ {
		nc, err = nats.Connect("nats://nats:4222")
		if err != nil {
			fmt.Printf(">>>> Error connecting to nats server: %v\n", err)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		fmt.Printf(">>>> Error connecting to nats server: %v\n", err)
		panic(err)
	}
	defer nc.Drain()

	_, err = nc.Subscribe("foo", func(m *nats.Msg) {
		fmt.Printf("Received message: %s\n", string(m.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("hello %d", i)
		fmt.Printf("Publishing: %s\n", msg)
		if err := nc.Publish("foo", []byte(msg)); err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	}

	// Give subscriber time to process
	time.Sleep(2 * time.Second)
}
