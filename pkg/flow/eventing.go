package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	protocol "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/direktiv/direktiv/pkg/flow/database"
	igrpc "github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var knativeClients sync.Map

type EventingCtxKey string

const EventingCtxKeySource EventingCtxKey = "source"

type eventReceiver struct {
	events *events
	flow   *flow

	igrpc.UnimplementedEventingServer
}

type client struct {
	stream igrpc.Eventing_RequestEventsServer
}

func newEventReceiver(events *events, flow *flow) (*eventReceiver, error) {
	slog.Info("creating event receiver")

	return &eventReceiver{
		events: events,
		flow:   flow,
	}, nil
}

func (rcv *eventReceiver) sendToNamespace(name string, r *http.Request) error {
	ctx := r.Context()
	ctx, end := startIncomingEvent(ctx, "http")
	defer end()
	slog.Debug("event for namespace", "namespace", name)

	m := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(ctx, m)
	if err != nil {
		return err
	}
	var ns *database.Namespace
	err = rcv.flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, name)
		return err
	})
	if err != nil {
		slog.Error("error getting namespace:", "error", err)
		return err
	}

	c := context.WithValue(ctx, EventingCtxKeySource, "eventing")

	return rcv.events.BroadcastCloudevent(c, ns, ev, 0)
}

func (rcv *eventReceiver) NamespaceHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("namespace knative event")

	ns := mux.Vars(r)["ns"]

	err := rcv.sendToNamespace(ns, r)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (rcv *eventReceiver) MultiNamespaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var nss []*database.Namespace
	var err error
	err = rcv.flow.runSqlTx(context.Background(), func(tx *sqlTx) error {
		nss, err = tx.DataStore().Namespaces().GetAll(ctx)
		return err
	})
	if err != nil {
		slog.Error("can not get namespace", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i := range nss {
		err := rcv.sendToNamespace(nss[i].Name, r)
		if err != nil {
			slog.Error("error sending event", "error", err.Error())
		}
	}
}

func PublishKnativeEvent(ce *cloudevents.Event) {
	var errorClients []string

	knativeClients.Range(func(k, v interface{}) bool {
		id, _ := k.(string)
		c, _ := v.(client)

		b, err := format.Protobuf.Marshal(ce)
		if err != nil {
			slog.Error("can not marshal cloud event", "error", err)
			return false
		}

		ce := &igrpc.CloudEvent{
			Ce: b,
		}

		if err := c.stream.Send(ce); err != nil {
			slog.Error("can not send event for client", "id", id, "error", err)
			errorClients = append(errorClients, id)
		}
		return true
	})

	// error clients getting removed
	for _, id := range errorClients {
		knativeClients.Delete(id)
	}
}

func (rcv *eventReceiver) RequestEvents(req *igrpc.EventingRequest, stream igrpc.Eventing_RequestEventsServer) error {
	slog.Debug("Client connected to event stream.", "client", req.GetUuid())

	knativeClients.Store(req.GetUuid(), client{stream: stream})

	ctx := stream.Context()

	<-ctx.Done()

	slog.Debug("Client has disconnected from event stream.", "client", req.GetUuid())
	knativeClients.Delete(req.GetUuid())
	return nil
}

func (rcv *eventReceiver) startGRPC() {
	slog.Debug("Starting the eventing gRPC server.", "port", 3333)

	var grpcServer *grpc.Server

	err := util.GrpcStart(&grpcServer, "eventing",
		fmt.Sprintf(":%d", 3333), func(srv *grpc.Server) {
			igrpc.RegisterEventingServer(srv, rcv)
			reflection.Register(srv)
		})
	if err != nil {
		slog.Error("Failed to start the eventing gRPC server.", "error", err, "port", 3333)
	}
}

func (rcv *eventReceiver) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/{ns}", rcv.NamespaceHandler).Methods(http.MethodPost)
	r.HandleFunc("/", rcv.MultiNamespaceHandler).Methods(http.MethodPost)

	go rcv.startGRPC()

	slog.Debug("starting event receiver")

	err := http.ListenAndServe(":1644", r)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start HTTP server.", "error", err)
		}
	}
}
