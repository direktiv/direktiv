// nolint
package gateway

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
	"github.com/direktiv/direktiv/pkg/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/gateway/endpoints"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type namespaceGateway struct {
	EndpointList *endpoints.EndpointList
	ConsumerList *consumer.List
}

type gatewayManager struct {
	db         *database.SQLStore
	nsGateways map[string]*namespaceGateway
	lock       sync.RWMutex
}

const anonymousUsername = "Anonymous"

func NewGatewayManager(db *database.SQLStore) core.GatewayManager {
	return &gatewayManager{
		db:         db,
		nsGateways: make(map[string]*namespaceGateway),
	}
}

func (ep *gatewayManager) updateNamespace(ns string) {
	slog.Debug("updating namespace gateway", slog.String("namespace", ns), "track", recipient.Namespace.String()+"."+ns)

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

	files, err := fStore.ForNamespace(ns).ListDirektivFilesWithData(ctx)
	if err != nil {
		slog.Error("list files", "err", err, "track", recipient.Namespace.String()+"."+ns)

		return
	}

	eps := make([]*core.Endpoint, 0)
	consumers := make([]*core.ConsumerFile, 0)

	for _, file := range files {
		if file.Typ != filestore.FileTypeConsumer &&
			file.Typ != filestore.FileTypeEndpoint {
			continue
		}

		if file.Typ == filestore.FileTypeConsumer {
			item, err := core.ParseConsumerFile(file.Data)
			if err != nil {
				slog.Error("parse endpoint file", "err", err, "track", recipient.Namespace.String()+"."+ns)

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
				Methods:                 make([]string, 0),
				Errors:                  make([]string, 0),
				Warnings:                make([]string, 0),
				AuthPluginInstances:     make([]core.PluginInstance, 0),
				InboundPluginInstances:  make([]core.PluginInstance, 0),
				OutboundPluginInstances: make([]core.PluginInstance, 0),
				TargetPluginInstance:    nil,
				FilePath:                file.Path,
				Namespace:               ns,
			}

			item, err := core.ParseEndpointFile(file.Data)
			// if parsing fails, the endpoint is still getting added to report
			// an error in the API
			if err != nil {
				slog.Error("parse endpoint file", "err", err)
				ep.Errors = append(ep.Errors, err.Error())
				eps = append(eps, ep)

				continue
			}

			ep.ServerPath = filepath.Join("/ns", ns, item.Path)

			ep.AllowAnonymous = item.AllowAnonymous
			ep.Timeout = item.Timeout
			ep.Methods = item.Methods
			ep.Path = item.Path
			ep.Plugins = item.Plugins

			endpoints.MakeEndpointPluginChain(ep, &item.Plugins)

			eps = append(eps, ep)
		}
	}

	gw.EndpointList.SetEndpoints(eps)
	gw.ConsumerList.SetConsumers(consumers)
}

func (ep *gatewayManager) UpdateAll() {
	_, dStore := ep.db.FileStore(), ep.db.DataStore()

	ep.nsGateways = map[string]*namespaceGateway{}

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		slog.Error("listing namespaces", "err", err)

		return
	}

	for _, ns := range nsList {
		ep.updateNamespace(ns.Name)
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
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()
	spanContext := span.SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID()
	slog := slog.With("trace", traceID, "span", spanID, "component", "gateway")
	chiCtx := chi.RouteContext(r.Context())
	namespace := core.SystemNamespace
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

	// if there are configuration errors, return it
	if len(endpointEntry.Errors) > 0 {
		plugins.ReportError(ctx, w, http.StatusInternalServerError, "plugin has errors",
			fmt.Errorf(strings.Join(endpointEntry.Errors, ", ")))

		return
	}

	slogRoute := slog.With("trace", traceID,
		"span", spanID,
		"track", recipient.Route.String()+"."+namespace+"."+endpointEntry.Path,
		"namespace", namespace, "endpoint", endpointEntry.Path,
		"route", routePath)

	// add url params e.g. /{id}
	ctx = context.WithValue(ctx, plugins.URLParamCtxKey, urlParams)
	ctx = context.WithValue(ctx, plugins.ConsumersParamCtxKey, gw.ConsumerList)

	// timeout
	t := endpointEntry.Timeout

	// timeout is 30 secs if not set
	if t == 0 {
		t = 30
	}

	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("direktiv/flow")
	ctx, childSpan := tracer.Start(ctx, "plugins-processing")
	defer childSpan.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(t))
	defer cancel()
	ctx = context.WithValue(ctx, plugins.NamespaceCtxKey, namespace)
	ctx = context.WithValue(ctx, plugins.EndpointCtxKey, endpointEntry.Path)
	ctx = context.WithValue(ctx, plugins.RouteCtxKey, routePath)
	r = r.WithContext(ctx)
	c := &core.ConsumerFile{}
	for i := range endpointEntry.AuthPluginInstances {
		authPlugin := endpointEntry.AuthPluginInstances[i]

		// all authplugins succeed, the setting of the consumer is the success message
		authPlugin.ExecutePlugin(c, w, r)

		// check and exit if consumer is set in plugin
		if c.Username != "" {
			slog.Info("user authenticated", "user", c.Username)

			break
		}
	}

	// if user not authenticated and anonymous access not enabled
	if c.Username == "" && !endpointEntry.AllowAnonymous {
		plugins.ReportError(ctx, w, http.StatusUnauthorized, "no permission",
			fmt.Errorf("request not authorized"))
		slogRoute.Debug("user authenticated", "user", c.Username)

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
	// on the wire immediately.
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
			plugins.ReportError(ctx, w, http.StatusUnauthorized, "output plugin failed",
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
			slogRoute.Error("Failed to write api response", "err", err)
		}
	}
}

func executePlugin(c *core.ConsumerFile, w http.ResponseWriter, r *http.Request,
	fn func(*core.ConsumerFile, http.ResponseWriter, *http.Request) bool,
) bool {
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

// API functions.
func (ep *gatewayManager) GetConsumers(namespace string) ([]*core.ConsumerFile, error) {
	g, ok := ep.nsGateways[namespace]
	if !ok {
		return nil, fmt.Errorf("no consumers for namespace %s", namespace)
	}

	return g.ConsumerList.GetConsumers(), nil
}

func (ep *gatewayManager) GetRoutes(namespace, filteredPath string) ([]*core.Endpoint, error) {
	g, ok := ep.nsGateways[namespace]
	if !ok {
		return nil, fmt.Errorf("no routes for namespace %s", namespace)
	}

	result := make([]*core.Endpoint, 0)
	for _, v := range g.EndpointList.GetEndpoints() {
		if v.Path == filteredPath || filteredPath == "" {
			result = append(result, v)
		}
	}

	return result, nil
}
