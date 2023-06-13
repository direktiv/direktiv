package main

import (
	"fmt"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/bus"
	"github.com/nsqio/go-nsq"
)

func main() {
	fmt.Println("start")

	port1 := os.Getenv("PORT1")
	port2 := os.Getenv("PORT2")
	port3 := os.Getenv("PORT3")

	c := bus.DefaultConfig()
	c.NSQDListen = fmt.Sprintf("127.0.0.1:%s", port1)
	c.LookupListen = fmt.Sprintf("127.0.0.1:%s", port2)
	c.LookupListenHTTP = fmt.Sprintf("127.0.0.1:%s", port3)
	c.PREFIX = "[JENS] "

	nfs := bus.NewNodefinderStatic(
		[]string{
			// "127.0.0.1:5551",
			"127.0.0.1:4260",
		},
	)
	c.Nodefinder = nfs

	bus1, err := startBus(c)

	if err != nil {
		panic(err)
	}
	defer bus1.Stop()

	if port1 == "4250" {
		time.Sleep(5 * time.Second)
	}

	config := nsq.NewConfig()

	// if port1 == "4280" {
	go publish(c.NSQDListen, port1)
	// }
	time.Sleep(1 * time.Second)
	// prod, err := nsq.NewProducer(c.NSQDListen, config)
	// if err != nil {
	// 	panic(err)
	// }
	// err = prod.Publish("jens", []byte("GERKE"))
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("PUBLISHED!!!!!")
	// time.Sleep(5 * time.Second)

	cons, err := nsq.NewConsumer("4250", port3, config)
	if err != nil {
		panic(err)
	}
	cons.AddHandler(&myMessageHandler{
		topic: "PORT1",
	})
	err = cons.ConnectToNSQLookupd("127.0.0.1:4270")

	cons2, err := nsq.NewConsumer("4280", port2, config)
	if err != nil {
		panic(err)
	}
	cons2.AddHandler(&myMessageHandler{
		topic: "PORT2",
	})
	err = cons2.ConnectToNSQLookupd("127.0.0.1:4270")

	if err != nil {
		panic(err)
	}

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

func startBus(config *bus.Config) (*bus.Bus, error) {

	d, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("%v", time.Now().UnixMilli()))
	config.DataPath = d

	bus, err := bus.NewBus(config)
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

func publish(listen, topic string) {

	for {
		config := nsq.NewConfig()
		prod, err := nsq.NewProducer(listen, config)
		if err != nil {
			panic(err)
		}

		d := fmt.Sprintf("%v", time.Now().UnixMicro())

		err = prod.Publish(topic, []byte(d))
		if err != nil {
			panic(err)
		}
		fmt.Printf("PUBLISHED!!!!! %v\n", topic)
		time.Sleep(2 * time.Second)
	}

}
