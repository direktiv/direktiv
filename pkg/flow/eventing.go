package flow

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	protocol "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/dlog"
	igrpc "github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type eventReceiver struct {
	logger *zap.SugaredLogger
	events *events
	flow   *flow

	clients sync.Map

	igrpc.UnimplementedEventingServer
}

type client struct {
	stream     igrpc.Eventing_RequestEventsServer
	disconnect chan<- bool
}

func newEventReceiver(events *events, flow *flow) (*eventReceiver, error) {

	logger, err := dlog.ApplicationLogger("eventing")
	if err != nil {
		return nil, err
	}

	logger.Infof("creating event receiver")

	return &eventReceiver{
		logger: logger,
		events: events,
		flow:   flow,
	}, nil

}

func requestToEvent(r *http.Request) (*cloudevents.Event, error) {
	m := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(context.Background(), m)
	if err != nil {
		return nil, err
	}

	return ev, ev.Validate()
}

func (rcv *eventReceiver) sendToNamespace(ns string, r *http.Request) error {

	rcv.logger.Debugf("event for namespace %s", ns)

	m := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(context.Background(), m)
	if err != nil {
		return err
	}

	namespace, err := rcv.flow.getNamespace(context.Background(), rcv.flow.db.Namespace, ns)
	if err != nil {
		rcv.logger.Errorf("error getting namespace: %s", err.Error())
		return err
	}

	c := context.WithValue(context.Background(), "source", "eventing")

	return rcv.events.BroadcastCloudevent(c, namespace, ev, 0)

}

func (rcv *eventReceiver) NamespaceHandler(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["ns"]

	err := rcv.sendToNamespace(ns, r)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusAccepted)

}

func (rcv *eventReceiver) MultiNamespaceHandler(w http.ResponseWriter, r *http.Request) {

	nsc := rcv.flow.db.Namespace
	nss, err := nsc.Query().All(context.Background())
	if err != nil {
		rcv.logger.Errorf("can not get namespaces: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := range nss {
		err := rcv.sendToNamespace(nss[i].Name, r)
		if err != nil {
			rcv.logger.Errorf("error sending event: %s", err.Error())
		}
	}

}

func PublishKnativeEvent(ce *cloudevents.Event) {
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!EVENT %+v\n", ce)
}

// func (flow *flow) InstanceInput(ctx context.Context, req *grpc.InstanceInputRequest) (*grpc.InstanceInputResponse, error) {
func (rcv *eventReceiver) RequestEvents(req *igrpc.EventingRequest, stream igrpc.Eventing_RequestEventsServer) error {

	rcv.logger.Infof("client connected: %v", req.GetUuid())

	disconnect := make(chan bool)
	rcv.clients.Store(req.GetUuid(), client{stream: stream, disconnect: disconnect})

	ctx := stream.Context()

	for {
		select {
		case <-disconnect:
			rcv.logger.Infof("closing stream for client: %d", req.GetUuid())
			return nil
		case <-ctx.Done():
			rcv.logger.Infof("client %d has disconnected", req.GetUuid())
			return nil
		}
	}

}

func (rcv *eventReceiver) startGRPC() {

	rcv.logger.Infof("starting eventing grpc")

	var grpcServer *grpc.Server
	util.GrpcStart(&grpcServer, "eventing",
		fmt.Sprintf(":%d", 3333), func(srv *grpc.Server) {
			igrpc.RegisterEventingServer(srv, rcv)
			reflection.Register(srv)
		})

}

func (rcv *eventReceiver) Start() {

	r := mux.NewRouter()
	r.HandleFunc("/{ns}", rcv.NamespaceHandler).Methods(http.MethodPost)
	r.HandleFunc("/", rcv.MultiNamespaceHandler).Methods(http.MethodPost)

	go rcv.startGRPC()

	rcv.logger.Infof("starting event receiver")
	http.ListenAndServe(":8080", r)

}
