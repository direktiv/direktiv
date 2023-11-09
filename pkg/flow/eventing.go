package flow

import (
	"sync"

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	igrpc "github.com/direktiv/direktiv/pkg/flow/grpc"
	"go.uber.org/zap"
)

var knativeClients sync.Map

type EventingCtxKey string

const EventingCtxKeySource EventingCtxKey = "source"

type client struct {
	stream igrpc.Eventing_RequestEventsServer
}

var publishLogger *zap.SugaredLogger

func PublishKnativeEvent(ce *cloudevents.Event) {
	var errorClients []string

	knativeClients.Range(func(k, v interface{}) bool {
		id, _ := k.(string)
		c, _ := v.(client)

		b, err := format.Protobuf.Marshal(ce)
		if err != nil {
			publishLogger.Errorf("can not marshal cloud event: %v", err)
			return false
		}

		ce := &igrpc.CloudEvent{
			Ce: b,
		}

		if err := c.stream.Send(ce); err != nil {
			publishLogger.Errorf("can not send event for client %s: %v", id, err)
			errorClients = append(errorClients, id)
		}
		return true
	})

	// error clients getting removed
	for _, id := range errorClients {
		knativeClients.Delete(id)
	}
}
