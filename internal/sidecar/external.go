package sidecar

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type externalServer struct {
	server *http.Server

	rm *requestMap
}

type contextSpanKey string

const spanKey contextSpanKey = "sidecar-call"

func newExternalServer(rm *requestMap) *externalServer {
	// we can ignore the error here
	addr, _ := url.Parse("http://localhost:8080")
	proxy := httputil.NewSingleHostReverseProxy(addr)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if req.URL.Path == "/up" {
			// status request to avoid retry
			// this makes the proxy fail
			req.URL = nil
			return
		}

		// add action header
		actionID := req.Header.Get(core.EngineHeaderActionID)

		ctx := otel.GetTextMapPropagator().Extract(
			req.Context(),
			propagation.HeaderCarrier(req.Header),
		)

		tracer := otel.Tracer("action-call")
		ctx, span := tracer.Start(ctx, "action-call")
		span.SetAttributes(attribute.KeyValue{
			Key:   "instance",
			Value: attribute.StringValue(actionID),
		},
			attribute.KeyValue{
				Key:   "image",
				Value: attribute.StringValue(os.Getenv("DIREKTIV_IMAGE")),
			},
		)
		ctx = context.WithValue(ctx, spanKey, span)

		// init logging
		lo := telemetry.LogObjectFromHeader(ctx, req.Header)
		ctx = telemetry.LogInitCtx(req.Context(), lo)

		// add log object for internal server
		rm.Add(actionID, lo)

		telemetry.LogInstance(ctx, telemetry.LogLevelInfo, "action request received")

		*req = *req.WithContext(ctx)
		originalDirector(req)
		otel.GetTextMapPropagator().Inject(
			ctx,
			propagation.HeaderCarrier(req.Header),
		)

		// set headers
		lo.ToHeader(&req.Header)

		// TODO: create temp directory
		req.Header.Set(core.EngineHeaderTempDir, "/tmp")

		// Log for debugging
		slog.Info("forwarding request to user container", slog.String("actionID", actionID))
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if r.URL == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if span, ok := resp.Request.Context().Value(spanKey).(trace.Span); ok {
			span.SetAttributes(
				attribute.Int("http.status_code", resp.StatusCode),
			)
			span.AddEvent("response received")

			if resp.StatusCode >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			}
			span.End()
		}

		telemetry.LogInstance(resp.Request.Context(), telemetry.LogLevelInfo,
			"action request finished")

		// remove from the request map
		lo, ok := resp.Request.Context().Value(telemetry.DirektivLogCtx(telemetry.LogObjectIdentifier)).(telemetry.LogObject)
		if ok {
			rm.Remove(lo.ID)
		}

		// if it is not ok, we return 502 to trigger the retry
		if resp.StatusCode != http.StatusOK {
			code := resp.Header.Get(core.EngineHeaderErrorCode)
			msg := resp.Header.Get(core.EngineHeaderErrorMessage)

			telemetry.LogInstance(resp.Request.Context(), telemetry.LogLevelError,
				fmt.Sprintf("action request failed with status code %d", resp.StatusCode))

			if code != "" {
				msg += fmt.Sprintf(" (%s)", code)
			}

			if msg != "" {
				telemetry.LogInstance(resp.Request.Context(), telemetry.LogLevelError,
					msg)
			}

			resp.StatusCode = 502
		}

		return nil
	}

	slog.Info("starting external proxy")

	s := &externalServer{
		server: &http.Server{
			Addr:    "0.0.0.0:8890",
			Handler: proxy,
		},
		rm: rm,
	}

	return s
}
