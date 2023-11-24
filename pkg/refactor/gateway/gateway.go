package gateway

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/go-chi/chi/v5"

	// This triggers the init function within for auth plugins to register them.

	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/endpoints"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"

	// This triggers the init function within for inbound plugins to register them.
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/target"
)

type namespaceGateway struct {
	EndpointList *endpoints.EndpointList
	ConsumerList *consumer.ConsumerList
}

type gatewayManager struct {
	db         *database.DB
	nsGateways map[string]*namespaceGateway
	lock       sync.RWMutex
}

const anonymousUsername = "Anonymous"

func NewGatewayManager(db *database.DB) core.GatewayManager {
	return &gatewayManager{
		db:         db,
		nsGateways: make(map[string]*namespaceGateway),
	}
}

func (ep *gatewayManager) DeleteNamespace(ns string) {
	slog.Info("deleting namespace from gateway", "namespace", ns)
	delete(ep.nsGateways, ns)
}

func (ep *gatewayManager) UpdateNamespace(ns string) {
	slog.Info("updating namespace gateway", slog.String("namespace", ns))

	ep.lock.Lock()
	defer ep.lock.Unlock()

	gw, ok := ep.nsGateways[ns]
	if !ok {
		gw = &namespaceGateway{
			EndpointList: endpoints.NewEndpointList(),
			ConsumerList: consumer.NewConsumerList(),
		}
		ep.nsGateways[ns] = gw
	}

	fStore := ep.db.FileStore()
	ctx := context.Background()

	files, err := fStore.ForNamespace(ns).ListDirektivFiles(ctx)
	if err != nil {
		slog.Error("error listing files", slog.String("error", err.Error()))

		return
	}

	endpoints := make([]*core.Endpoint, 0)
	consumers := make([]*spec.ConsumerFile, 0)

	for _, file := range files {
		if file.Typ != filestore.FileTypeConsumer &&
			file.Typ != filestore.FileTypeEndpoint {
			continue
		}

		data, err := fStore.ForFile(file).GetData(ctx)
		if err != nil {
			slog.Error("read file data", slog.String("error", err.Error()))

			continue
		}

		if file.Typ == filestore.FileTypeConsumer {
			item, err := spec.ParseConsumerFile(data)
			if err != nil {
				slog.Error("parse endpoint file", slog.String("error", err.Error()))

				continue
			}

			// username can not be empty or contain a colon for basic auth
			if item.Username == "" ||
				strings.Contains(item.Username, ":") {
				slog.Warn("username invalid", slog.String("user", item.Username))

				continue
			}

			consumers = append(consumers, item)
		} else {
			ep := &core.Endpoint{
				Namespace: ns,
				FilePath:  file.Path,
				Errors:    make([]string, 0),
				Warnings:  make([]string, 0),
			}
			item, err := spec.ParseEndpointFile(data)
			// if parsing fails, the endpoint is still getting added to report
			// an error in the API
			if err != nil {
				slog.Error("parse endpoint file", slog.String("error", err.Error()))
				ep.Errors = append(ep.Errors, err.Error())
			}
			ep.EndpointFile = item

			endpoints = append(endpoints, ep)
		}
	}

	gw.EndpointList.SetEndpoints(endpoints)
	gw.ConsumerList.SetConsumers(consumers)
}

func (ep *gatewayManager) UpdateAll() {
	_, dStore := ep.db.FileStore(), ep.db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		slog.Error("listing namespaces", slog.String("error", err.Error()))

		return
	}

	for _, ns := range nsList {
		ep.UpdateNamespace(ns.Name)
	}
}

type DummyWriter struct {
	HeaderMap http.Header
	Body      *bytes.Buffer
	Code      int
}

func NewDummyWriter() *DummyWriter {
	return &DummyWriter{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      http.StatusOK,
	}
}

func (dr *DummyWriter) Header() http.Header {
	return dr.HeaderMap
}

func (dr *DummyWriter) Write(buf []byte) (int, error) {
	return dr.Body.Write(buf)
}

func (dr *DummyWriter) WriteHeader(statusCode int) {
	dr.Code = statusCode
}

func (ep *gatewayManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chiCtx := chi.RouteContext(r.Context())
	namespace := core.MagicalGatewayNamespace
	routePath := chi.URLParam(r, "*")

	// get namespace from URL or use magical one
	if chiCtx.RoutePattern() == "/ns/{namespace}/*" {
		namespace = chi.URLParam(r, "namespace")
	}

	gw, ok := ep.nsGateways[namespace]
	if !ok {
		plugins.ReportNotFound(w)

		return
	}

	endpointEntry, urlParams := gw.EndpointList.FindRoute(routePath, r.Method)
	if endpointEntry == nil {
		plugins.ReportNotFound(w)

		return
	}

	// add url params e.g. /{id}
	ctx := context.WithValue(r.Context(), plugins.URLParamCtxKey, urlParams)
	ctx = context.WithValue(ctx, plugins.ConsumersParamCtxKey, gw.ConsumerList)

	// timeout
	t := endpointEntry.EndpointFile.Timeout

	// timeout is 30 secs if not set
	if t == 0 {
		t = 30
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(t))
	defer cancel()
	r = r.WithContext(ctx)

	c := &spec.ConsumerFile{}
	for i := range endpointEntry.AuthPluginInstances {
		authPlugin := endpointEntry.AuthPluginInstances[i]

		// auth plugins always return true to check all of them
		authPlugin.ExecutePlugin(c, w, r)

		// check and exit if consumer is set in plugin
		if c.Username != "" {
			slog.Info("user authenticated", "user", c.Username)

			break
		}
	}

	// if user not authenticated and anonymous access not enabled
	if c.Username == "" && !endpointEntry.EndpointFile.AllowAnonymous {
		plugins.ReportError(w, http.StatusUnauthorized, "no permission",
			fmt.Errorf("request not authorized"))

		return
	}

	// set username Anonymous if allowed and not set via auth plugin
	if c.Username == "" {
		c.Username = anonymousUsername
	}

	// run inbound
	for i := range endpointEntry.InboundPluginInstances {
		inboundPlugin := endpointEntry.InboundPluginInstances[i]
		proceed := inboundPlugin.ExecutePlugin(c, w, r)
		if !proceed {
			return
		}
	}

	// if there are outbound plugins the reponsewrite is getting swapped out
	// because target plugins can do io.copy and set headers which would go
	// on the wire immediatley.
	targetWriter := w
	if len(endpointEntry.OutboundPluginInstances) > 0 {
		targetWriter = NewDummyWriter()
	}

	// run target if it exists
	if endpointEntry.TargetPluginInstance != nil &&
		!endpointEntry.TargetPluginInstance.ExecutePlugin(c, targetWriter, r) {
		return
	}

	for i := range endpointEntry.OutboundPluginInstances {
		outboundPlugin := endpointEntry.OutboundPluginInstances[i]

		// nolint
		tw := targetWriter.(*DummyWriter)

		rin, err := swapRequestResponse(r, tw)
		if err != nil {
			plugins.ReportError(w, http.StatusUnauthorized, "output plugin failed",
				err)

			return
		}

		proceed := executePlugin(c, tw, rin,
			outboundPlugin.ExecutePlugin)

		// in outbound we need to break and not return
		// to write the actual output
		if !proceed {
			break
		}
	}

	// response already written, except if there are outbound plugins
	if len(endpointEntry.OutboundPluginInstances) > 0 {
		// nolint
		tw := targetWriter.(*DummyWriter)

		for k, v := range tw.HeaderMap {
			for a := range v {
				w.Header().Add(k, v[a])
			}
		}
		w.WriteHeader(tw.Code)
		_, err := w.Write(tw.Body.Bytes())
		if err != nil {
			slog.Error("can not write api repsonse", slog.Any("error", err.Error()))
		}
	}
}

func executePlugin(c *spec.ConsumerFile, w http.ResponseWriter, r *http.Request,
	fn func(*spec.ConsumerFile, http.ResponseWriter, *http.Request) bool) bool {

	select {
	case <-r.Context().Done():
		w.WriteHeader(http.StatusRequestTimeout)
		//nolint
		w.Write([]byte("request timed out"))

		return false
	default:
	}

	return fn(c, w, r)
}

func swapRequestResponse(rin *http.Request, w *DummyWriter) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodGet, "/writer", w.Body)
	if err != nil {
		return nil, err
	}
	r.Response = &http.Response{
		StatusCode: w.Code,
	}

	return r.WithContext(rin.Context()), nil
}
