package sidecar

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/mholt/archives"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type externalServer struct {
	server *http.Server

	rm *requestMap
}

type contextSpanKey string

const spanKey contextSpanKey = "sidecar-call"

const sharedDir = "/mnt/shared"

func newExternalServer(rm *requestMap) (*externalServer, error) {
	// we can ignore the error here
	addr, _ := url.Parse("http://localhost:8080")
	proxy := httputil.NewSingleHostReverseProxy(addr)

	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("DB"),
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
	}), gormConf)
	if err != nil {
		return nil, err
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

		fmt.Println("-------------")
		fmt.Println(req.Header)
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

		// create temp directory
		setupDirectories(actionID)

		// prepare files
		err = handleInFiles(req.Context(), lo.Namespace, lo.Path, path.Join(sharedDir, actionID), db, req.Header)
		if err != nil {
			telemetry.LogInstance(ctx, telemetry.LogLevelError, fmt.Sprintf("could not prepare in files: %s", err.Error()))
		}

		req.Header.Set(core.EngineHeaderTempDir, path.Join(sharedDir, actionID))

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

		defer func() {
			slog.Info("deleting tmp directory")
			os.RemoveAll(filepath.Join(sharedDir, lo.ID))
		}()

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

		// handle out files
		err := handleOutFiles(resp.Request.Context(), db, lo.Namespace, lo.InstanceInfo.Path, lo.ID)
		if err != nil {
			slog.Error("cannot handle out files", slog.Any("error", err))
			telemetry.LogInstance(resp.Request.Context(), telemetry.LogLevelError,
				fmt.Sprintf("cannot handle variables: %s", err.Error()))
			resp.StatusCode = http.StatusInternalServerError
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

	return s, nil
}

func handleInFiles(ctx context.Context, namespace, workflow, dirs string, db *gorm.DB, header http.Header) error {
	files := header.Values(core.EngineHeaderFile)

	dStore := datasql.NewStore(db)
	fStore := filesql.NewStore(db)

	slog.Info("handle in files")
	for a := range files {
		f := files[a]

		slog.Info("handling in file", slog.String("value", f))
		split := strings.Split(f, ";")
		if len(split) != 2 {
			return fmt.Errorf("file value not valid: %s", f)
		}

		writePath := filepath.Join(dirs, filepath.Base(split[1]))

		switch split[0] {
		case "workflow":
			variable, err := dStore.RuntimeVariables().GetForWorkflow(ctx, namespace, workflow, split[1])
			if err != nil {
				return err
			}

			b, err := dStore.RuntimeVariables().LoadData(ctx, variable.ID)
			if err != nil {
				return err
			}

			err = os.WriteFile(writePath, b, 0755)
			if err != nil {
				return err
			}
		case "namespace":
			variable, err := dStore.RuntimeVariables().GetForNamespace(ctx, namespace, split[1])
			if err != nil {
				return err
			}

			b, err := dStore.RuntimeVariables().LoadData(ctx, variable.ID)
			if err != nil {
				return err
			}

			err = os.WriteFile(writePath, b, 0755)
			if err != nil {
				return err
			}
		default:

			fmt.Println("-------------------------------------------------")
			fmt.Println(split[1])

			// calculate absolute path, if relative is given
			if !filepath.IsAbs(split[1]) {
				absFile := filepath.Join(filepath.Dir(workflow), split[1])
				fmt.Println(absFile)

				// remove the temporary filesystem path
				direktivPath := strings.Replace(absFile, dirs, "", 1)
				fmt.Println(direktivPath)

				split[1] = direktivPath
			}

			f, err := fStore.ForRoot(namespace).GetFile(ctx, split[1])
			if err != nil {
				return err
			}
			b, err := fStore.ForFile(f).GetData(ctx)
			if err != nil {
				return err
			}

			err = os.WriteFile(writePath, b, 0755)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handleOutFiles(ctx context.Context, db *gorm.DB, namespace, filePath, id string) error {
	subDirs := []string{"instance", "workflow", "namespace"}
	for i := range subDirs {
		tmpDir := path.Join(sharedDir, id, fmt.Sprintf("out/%s", subDirs[i]))
		slog.Info("checking out folder", slog.String("dir", tmpDir))

		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			slog.Error("cannot read out dir", slog.Any("error", err))
			return err
		}

		dStore := datasql.NewStore(db)

		for a := range entries {
			entry := entries[a]

			data, mimeType, err := prepareFile(path.Join(sharedDir, id), filepath.Join(tmpDir, entry.Name()), entry.IsDir())
			if err != nil {
				return err
			}

			variable := &datastore.RuntimeVariable{
				Namespace: namespace,
				Name:      entry.Name(),
				MimeType:  mimeType,
				Data:      data,
			}

			var existingVar *datastore.RuntimeVariable

			switch subDirs[i] {
			case "workflow":
				variable.WorkflowPath = filePath
				existingVar, err = dStore.RuntimeVariables().GetForWorkflow(ctx, namespace, filePath, entry.Name())
			case "instance":
				variable.InstanceID = uuid.MustParse(id)
				existingVar, err = dStore.RuntimeVariables().GetForInstance(ctx, variable.ID, entry.Name())
			default:
				existingVar, err = dStore.RuntimeVariables().GetForNamespace(ctx, namespace, entry.Name())
			}

			update := true
			if err != nil {
				if errors.Is(err, datastore.ErrNotFound) {
					update = false
				} else {
					return err
				}
			}

			if update {
				telemetry.LogInstance(ctx, telemetry.LogLevelInfo, fmt.Sprintf("updating variable %s in scope %s", entry.Name(), subDirs[i]))
				_, err = dStore.RuntimeVariables().Patch(ctx, existingVar.ID, &datastore.RuntimeVariablePatch{
					Name:     &variable.Name,
					MimeType: &variable.MimeType,
					Data:     data,
				})
				if err != nil {
					return err
				}
			} else {
				telemetry.LogInstance(ctx, telemetry.LogLevelInfo, fmt.Sprintf("creating variable %s in scope %s", entry.Name(), subDirs[i]))
				_, err = dStore.RuntimeVariables().Create(ctx, variable)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func prepareFile(tmpDir, p string, isDir bool) ([]byte, string, error) {
	if isDir {
		tarFile := fmt.Sprintf("%s.tar.gz", filepath.Base(p))
		out, err := os.Create(filepath.Join(tmpDir, tarFile))
		if err != nil {
			return nil, "", err
		}
		defer out.Close()

		format := archives.CompressedArchive{
			Compression: archives.Gz{},
			Archival:    archives.Tar{},
		}

		files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
			p: "",
		})
		if err != nil {
			return nil, "", err
		}

		err = format.Archive(context.Background(), out, files)

		data, err := os.ReadFile(filepath.Join(tmpDir, tarFile))

		return data, mimetype.Lookup("application/gzip").String(), err
	}

	mtype, err := mimetype.DetectFile(p)
	if err != nil {
		mtype = mimetype.Lookup("application/text")
	}

	data, err := os.ReadFile(p)

	return data, mtype.String(), err
}

func setupDirectories(id string) error {
	subDirs := []string{"instance", "workflow", "namespace", "system"}
	for i := range subDirs {
		tmpDir := path.Join(sharedDir, id, fmt.Sprintf("out/%s", subDirs[i]))
		slog.Info("creating tmp folder", slog.String("dir", tmpDir))

		err := os.MkdirAll(path.Join(sharedDir, id), 0777)
		if err != nil {
			slog.Error("cannot create tmp dir", slog.Any("error", err))
			return err
		}
		err = os.MkdirAll(tmpDir, 0o777)
		if err != nil {
			slog.Error("cannot create tmp dirs", slog.Any("error", err))
			return err
		}
	}

	return nil
}
