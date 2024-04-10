package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	database2 "github.com/direktiv/direktiv/pkg/refactor/database"

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	protocol "github.com/cloudevents/sdk-go/v2/protocol/http"
	igrpc "github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
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
	slog.Debug("Processing incoming event for namespace.", "namespace", name, "method", r.Method, "uri", r.RequestURI)

	m := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(ctx, m)
	if err != nil {
		slog.Error("Failed to convert HTTP request to CloudEvent.", "namespace", name, "error", err)
		return err
	}
	var ns *nohome.Namespace
	err = rcv.flow.runSqlTx(ctx, func(tx *database2.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, name)
		return err
	})
	if err != nil {
		slog.Error("Failed to retrieve namespace from database.", "namespace", name, "error", err)
		return err
	}

	c := context.WithValue(ctx, EventingCtxKeySource, "eventing")
	slog.Info("Broadcasting CloudEvent to namespace.", "namespace", name, "eventID", ev.ID(), "eventType", ev.Type())

	return rcv.events.BroadcastCloudevent(c, ns, ev, 0)
}

func (rcv *eventReceiver) NamespaceHandler(w http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["ns"]
	slog.Debug("Received Knative event for namespace.", "namespace", ns, "path", r.URL.Path)

	err := rcv.sendToNamespace(ns, r)
	if err != nil {
		slog.Error("Failed to process event for namespace.", "namespace", ns, "error", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	slog.Info("Successfully processed and accepted Knative event for namespace.", "namespace", ns)
	w.WriteHeader(http.StatusAccepted)
}

func (rcv *eventReceiver) MultiNamespaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var nss []*nohome.Namespace
	var err error
	err = rcv.flow.runSqlTx(context.Background(), func(tx *database2.SQLStore) error {
		nss, err = tx.DataStore().Namespaces().GetAll(ctx)
		return err
	})
	if err != nil {
		slog.Error("Failed to fetch namespaces for event broadcasting.", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Debug("Starting to send events to all namespaces.", "count", len(nss))
	for i := range nss {
		err := rcv.sendToNamespace(nss[i].Name, r)
		if err != nil {
			slog.Error("Failed to send event to namespace.", "namespace", nss[i].Name, "error", err)
		}
	}
	slog.Info("Completed sending events to all namespaces.", "totalNamespaces", len(nss))
}

func PublishKnativeEvent(ce *cloudevents.Event) {
	var errorClients []string

	knativeClients.Range(func(k, v interface{}) bool {
		id, _ := k.(string)
		c, _ := v.(client)

		b, err := format.Protobuf.Marshal(ce)
		if err != nil {
			slog.Error("Failed to marshal CloudEvent.", "error", err)
			return false
		}

		ce := &igrpc.CloudEvent{
			Ce: b,
		}

		if err := c.stream.Send(ce); err != nil {
			slog.Error("Failed to send CloudEvent to client.", "client_id", id, "error", err)
			errorClients = append(errorClients, id)
		}
		return true
	})

	// error clients getting removed
	for _, id := range errorClients {
		slog.Info("Removing client due to errors.", "client_id", id)
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

	slog.Debug("Starting event receiver", "addr", ":1644")

	err := http.ListenAndServe(":1644", r)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start HTTP server.", "error", err, "addr", ":1644")
		}
	}

	slog.Debug("Event receiver started.", "addr", ":1644")
}
