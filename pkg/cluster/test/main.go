package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/cluster"
	"github.com/nsqio/go-nsq"
)

func main() {
	fmt.Println("start")

	port1 := os.Getenv("PORT1")
	port1a := os.Getenv("PORT1a")
	port2 := os.Getenv("PORT2")
	port3 := os.Getenv("PORT3")

	c := cluster.DefaultConfig()
	c.NSQDListen = fmt.Sprintf("127.0.0.1:%s", port1)
	c.NSQDListenHTTP = fmt.Sprintf("127.0.0.1:%s", port1a)
	c.LookupListen = fmt.Sprintf("127.0.0.1:%s", port2)
	c.LookupListenHTTP = fmt.Sprintf("127.0.0.1:%s", port3)
	c.PREFIX = "[!!!!!!!!!!!!!!!!!JENS] "

	nfs := cluster.NewNodefinderStatic(
		[]string{
			"127.0.0.1:4310",
			"127.0.0.1:4270",
		},
	)

	// c.ID = 1
	// if port1a == "4300" {
	// 	c.ID = 2
	// 	nfs = bus.NewNodefinderStatic(
	// 		[]string{
	// 			"127.0.0.1:4310",
	// 			"127.0.0.1:4270",
	// 		},
	// 	)
	// }

	c.Nodefinder = nfs

	bus1, err := startBus(c)

	if err != nil {
		panic(err)
	}
	defer bus1.Stop()

	time.Sleep(5 * time.Second)

	//
	// if port1a != "4300" {
	// 	url := fmt.Sprintf("http://%s/topic/create?topic=%s", c.NSQDListenHTTP, "topic1")
	// 	resp, err := http.Post(url, "application/json", nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	r, err := io.ReadAll(resp.Body)
	// 	fmt.Printf(">> %v %v\n", string(r), err)

	// 	url = fmt.Sprintf("http://%s/topic/create?topic=%s", c.NSQDListenHTTP, "topic2")
	// 	resp, err = http.Post(url, "application/json", nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	r, err = io.ReadAll(resp.Body)
	// 	fmt.Printf(">> %v %v\n", string(r), err)
	// }

	for i := 1; i < 20; i++ {

		url1 := fmt.Sprintf("http://%s/nodes", fmt.Sprintf("127.0.0.1:%s", port3))
		resp2, err := http.Get(url1)
		if err != nil {
			panic(err)
		}

		r2, err := io.ReadAll(resp2.Body)
		fmt.Printf(">>INFO %v %v\n", string(r2), err)

		time.Sleep(1 * time.Second)
	}

	// config := nsq.NewConfig()

	// cons, err := nsq.NewConsumer("topic1", port3, config)
	// if err != nil {
	// 	panic(err)
	// }
	// cons.AddHandler(&myMessageHandler{
	// 	topic: "PORT1",
	// })
	// err = cons.ConnectToNSQLookupd("127.0.0.1:4320")

	// cons2, err := nsq.NewConsumer("topic1", port2, config)
	// if err != nil {
	// 	panic(err)
	// }
	// cons2.AddHandler(&myMessageHandler{
	// 	topic: "PORT2",
	// })
	// err = cons2.ConnectToNSQLookupd("127.0.0.1:4320")

	// if err != nil {
	// 	panic(err)
	// }

	// for i := 1; i < 20; i++ {
	// 	fmt.Println(bus1.Status())
	// 	bus1.JJ()
	// 	time.Sleep(1 * time.Second)
	// }
	time.Sleep(60 * time.Second)

	fmt.Println(bus1.Status())
}

type myMessageHandler struct {
	topic string
}

var c1 int

func (h *myMessageHandler) HandleMessage(m *nsq.Message) error {
	c1 += 1
	fmt.Printf("?????????????????????????????????????????????????????????????????MESSAGE %v: %d, %v\n", h.topic, c1, string(m.Body))
	return nil
}

func startBus(config *cluster.Config) (*cluster.Bus, error) {

	d, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("%v", time.Now().UnixMilli()))
	config.DataPath = d

	bus, err := cluster.NewBus(config)
	if err != nil {
		return nil, err
	}

	go bus.Start()
	// ch := make(chan bool, 1)
	// bus.WaitTillConnected(ch)
	// ret := <-ch

	// if !ret {
	// 	return nil, fmt.Errorf("bus did not start in time")
	// }

	return bus, nil
}

// func publish(listen, topic string) {

// 	for {
// 		config := nsq.NewConfig()
// 		prod, err := nsq.NewProducer(listen, config)
// 		if err != nil {
// 			panic(err)
// 		}

// 		d := fmt.Sprintf("%v", time.Now().UnixMicro())

// 		err = prod.Publish(topic, []byte(d))
// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Printf("PUBLISHED!!!!! %v\n", topic)
// 		time.Sleep(2 * time.Second)
// 	}

// }
