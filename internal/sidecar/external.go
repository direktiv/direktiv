package sidecar

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/gabriel-vasile/mimetype"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const sharedDir = "/mnt/shared"

type externalServer struct {
	server    *http.Server
	variables datastore.RuntimeVariablesStore
	files     filestore.FileStore

	rm *requestMap
}

type contextSpanKey string

const spanKey contextSpanKey = "sidecar-call"

func newExternalServer(rm *requestMap, variables datastore.RuntimeVariablesStore, files filestore.FileStore) *externalServer {
	// we can ignore the error here
	addr, _ := url.Parse("http://localhost:8080")
	proxy := httputil.NewSingleHostReverseProxy(addr)

	es := &externalServer{
		rm:        rm,
		variables: variables,
		files:     files,
	}

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

		// create temp directory and handle input files
		tmpDir := filepath.Join(sharedDir, actionID)
		if err := os.MkdirAll(tmpDir, 0o755); err != nil {
			slog.Error("could not create temp directory", slog.Any("error", err))
		}

		namespace := req.Header.Get(core.EngineHeaderNamespace)
		workflowPath := req.Header.Get(core.EngineHeaderPath)

		if err := es.handleInFiles(ctx, req.Header, namespace, workflowPath, tmpDir); err != nil {
			slog.Error("error handling input files", slog.Any("error", err))
			telemetry.LogInstance(ctx, telemetry.LogLevelError,
				fmt.Sprintf("error handling input files: %v", err))
		}

		*req = *req.WithContext(ctx)
		originalDirector(req)
		otel.GetTextMapPropagator().Inject(
			ctx,
			propagation.HeaderCarrier(req.Header),
		)

		// set headers
		lo.ToHeader(&req.Header)

		req.Header.Set(core.EngineHeaderTempDir, tmpDir)

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

		// handle output files and cleanup temp directory
		actionID := resp.Request.Header.Get(core.EngineHeaderActionID)
		if actionID != "" {
			tmpDir := filepath.Join(sharedDir, actionID)
			namespace := resp.Request.Header.Get(core.EngineHeaderNamespace)
			workflowPath := resp.Request.Header.Get(core.EngineHeaderPath)

			if err := es.handleOutFiles(resp.Request.Context(), tmpDir, namespace, workflowPath, actionID); err != nil {
				slog.Error("error handling output files", slog.Any("error", err))
				telemetry.LogInstance(resp.Request.Context(), telemetry.LogLevelError,
					fmt.Sprintf("error handling output files: %v", err))
			}

			if err := os.RemoveAll(tmpDir); err != nil {
				slog.Error("could not remove temp directory", slog.String("path", tmpDir), slog.Any("error", err))
			}
		}

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

	es.server = &http.Server{
		Addr:    "0.0.0.0:8890",
		Handler: proxy,
	}

	return es
}

// handleInFiles parses Direktiv-File headers (format "scope:name"), fetches
// the data via the appropriate store, and writes the files into tmpDir.
func (es *externalServer) handleInFiles(ctx context.Context, headers http.Header, namespace, workflowPath, tmpDir string) error {
	fileHeaders := headers.Values(core.EngineHeaderFile)
	if len(fileHeaders) == 0 {
		return nil
	}

	for _, fh := range fileHeaders {
		parts := strings.SplitN(fh, ":", 3)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			slog.Warn("invalid file header format, expected scope:name[:permission]", slog.String("header", fh))
			continue
		}
		scope := parts[0]
		name := parts[1]
		permission := ""
		if len(parts) == 3 {
			permission = parts[2]
		}

		var data []byte

		switch scope {
		case "namespace", "workflow":
			if es.variables == nil {
				return fmt.Errorf("variable store not configured, cannot fetch %s variable '%s'", scope, name)
			}

			var (
				v   *datastore.RuntimeVariable
				err error
			)
			if scope == "namespace" {
				v, err = es.variables.GetForNamespace(ctx, namespace, name)
			} else {
				v, err = es.variables.GetForWorkflow(ctx, namespace, workflowPath, name)
			}
			if err != nil {
				return fmt.Errorf("fetching %s variable '%s': %w", scope, name, err)
			}

			data, err = es.variables.LoadData(ctx, v.ID)
			if err != nil {
				return fmt.Errorf("loading data for %s variable '%s': %w", scope, name, err)
			}

		case "file":
			if es.files == nil {
				return fmt.Errorf("file store not configured, cannot fetch file '%s'", name)
			}

			f, err := es.files.ForRoot(namespace).GetFile(ctx, name)
			if err != nil {
				return fmt.Errorf("fetching file '%s': %w", name, err)
			}

			data, err = es.files.ForFile(f).GetData(ctx)
			if err != nil {
				return fmt.Errorf("reading file data '%s': %w", name, err)
			}

		default:
			slog.Warn("unknown file scope", slog.String("scope", scope), slog.String("name", name))
			continue
		}

		perm := fs.FileMode(0o644)
		if permission != "" {
			p, err := strconv.ParseUint(permission, 8, 32)
			if err != nil {
				return fmt.Errorf("parsing permission '%s' for file '%s': %w", permission, name, err)
			}
			perm = fs.FileMode(p)
		}

		filePath := filepath.Join(tmpDir, name)
		if err := os.WriteFile(filePath, data, perm); err != nil {
			return fmt.Errorf("writing file '%s': %w", name, err)
		}

		slog.Info("wrote input file", slog.String("scope", scope),
			slog.String("name", name), slog.String("path", filePath))
	}

	return nil
}

// handleOutFiles reads output files from scope subdirectories in tmpDir
// (namespace/, workflow/, instance/) and writes them back as runtime variables.
func (es *externalServer) handleOutFiles(ctx context.Context, tmpDir, namespace, workflowPath, instanceID string) error {
	if es.variables == nil {
		return nil
	}

	scopes := []struct {
		dir   string
		scope string
	}{
		{"namespace", "namespace"},
		{"workflow", "workflow"},
		{"instance", "instance"},
	}

	for _, s := range scopes {
		scopeDir := filepath.Join(tmpDir, s.dir)
		entries, err := os.ReadDir(scopeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading %s output dir: %w", s.scope, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			data, err := os.ReadFile(filepath.Join(scopeDir, name))
			if err != nil {
				return fmt.Errorf("reading output file '%s' in scope '%s': %w", name, s.scope, err)
			}

			mtype := mimetype.Detect(data).String()

			rv := &datastore.RuntimeVariable{
				Namespace: namespace,
				Name:      name,
				MimeType:  mtype,
				Data:      data,
			}

			switch s.scope {
			case "namespace":
				// namespace-scoped: no workflow path or instance ID
			case "workflow":
				rv.WorkflowPath = workflowPath
			case "instance":
				instID, err := uuid.Parse(instanceID)
				if err != nil {
					return fmt.Errorf("parsing instance ID '%s': %w", instanceID, err)
				}
				rv.InstanceID = instID
			}

			_, err = es.variables.Set(ctx, rv)
			if err != nil {
				return fmt.Errorf("setting %s variable '%s': %w", s.scope, name, err)
			}

			slog.Info("wrote output variable", slog.String("scope", s.scope),
				slog.String("name", name))
		}
	}

	return nil
}
