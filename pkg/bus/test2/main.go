package main

import (
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
)

func main() {
	fmt.Println("start")

	// if port1 == "4280" {
	// publish(c.NSQDListen, port1)

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
