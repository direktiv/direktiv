package sidecar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/direktiv/direktiv/pkg/utils"
)

type inboundWorker struct {
	id     int
	cancel func()
	lock   sync.Mutex
	srv    *LocalServer
}

func (worker *inboundWorker) Cancel() {
	worker.lock.Lock()

	if worker.cancel != nil {
		slog.Debug("cancelling worker", "worker", worker.id)
		worker.cancel()
	}

	worker.lock.Unlock()
}

func (worker *inboundWorker) run() {
	slog.Debug("starting worker", "worker", worker.id)

	for {
		worker.lock.Lock()

		req, more := <-worker.srv.queue
		if req == nil || !more {
			worker.cancel = nil
			worker.lock.Unlock()
			break
		}

		ctx, cancel := context.WithCancel(req.r.Context())
		worker.cancel = cancel
		req.r = req.r.WithContext(ctx)

		worker.lock.Unlock()

		id := req.r.Header.Get(actionIDHeader)
		slog.Debug("worker picked up request", "worker", worker.id, "action-id", id)

		worker.handleFunctionRequest(req)
	}

	slog.Debug("worker shut down", "worker", worker.id)
}

type outcome struct {
	data    []byte
	errCode string
	errMsg  string
}

// nolint:canonicalheader
func (worker *inboundWorker) doFunctionRequest(ctx context.Context, ir *functionRequest) (*outcome, error) {
	// ctx, spanEnd, err := tracing.NewSpan(ctx, "execting function request: "+ir.actionId+", workflow: "+ir.Workflow)
	// if err != nil {
	// 	slog.Debug("doFunctionRequest failed", "error", err)
	// }
	// defer spanEnd()

	slog.Debug("forwarding request to service", "action-id", ir.actionId)

	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(ir.input))
	if err != nil {
		return nil, err
	}

	req.Header.Set(actionIDHeader, ir.actionId)
	req.Header.Set(IteratorHeader, fmt.Sprintf("%d", ir.Branch))
	req.Header.Set("Direktiv-TempDir", worker.functionDir(ir))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := new(outcome)

	out.errCode = resp.Header.Get("Direktiv-ErrorCode")
	out.errMsg = resp.Header.Get("Direktiv-ErrorMessage")

	if out.errCode != "" {
		return out, nil
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		out.errCode = "container failed"
		out.errMsg = string(out.data)
	}

	capa := int64(134217728) // 128 MiB (changed to same value as API)
	if resp.ContentLength > capa {
		return nil, errors.New("service response is too large")
	}
	r := io.LimitReader(resp.Body, capa)

	out.data, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func untarFile(tr *tar.Reader, perms string, path string) error {
	pdir, _ := filepath.Split(path)
	err := os.MkdirAll(pdir, 0o750)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	/* #nosec */
	defer f.Close()

	_, err = io.Copy(f, tr)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	if perms != "" {
		p, err := strconv.ParseUint(perms, 8, 32)
		if err != nil {
			return fmt.Errorf("failed to parse file permissions: %w", err)
		}

		err = os.Chmod(f.Name(), os.FileMode(uint32(p)))
		if err != nil {
			return fmt.Errorf("failed to apply file permissions: %w", err)
		}
	}

	return nil
}

func untar(dst string, perms string, r io.Reader) error {
	err := os.MkdirAll(dst, 0o750)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)

	for {
		/* #nosec */
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		hdr.Name = filepath.Clean(hdr.Name)
		if strings.Contains(hdr.Name, "..") {
			return errors.New("zip-slip")
		}

		/* #nosec */
		path := filepath.Join(dst, hdr.Name)

		if hdr.Typeflag == tar.TypeReg {
			err = untarFile(tr, perms, path)
			if err != nil {
				return err
			}
		} else if hdr.Typeflag == tar.TypeDir {
			err = os.MkdirAll(path, 0o750)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unsupported tar archive contents")
		}
	}

	return nil
}

func writeFile(dst, perms string, r io.Reader) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	/* #nosec */
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	if perms != "" {
		p, err := strconv.ParseUint(perms, 8, 32)
		if err != nil {
			return fmt.Errorf("failed to parse file permissions: %w", err)
		}

		err = os.Chmod(f.Name(), os.FileMode(uint32(p)))
		if err != nil {
			return fmt.Errorf("failed to apply file permissions: %w", err)
		}
	}

	return nil
}

type WrapperReader struct {
	rd     io.Reader
	offset int
}

func (wr *WrapperReader) Read(p []byte) (n int, err error) {
	o, err := wr.rd.Read(p)
	wr.offset += o
	return o, err
}

func (worker *inboundWorker) writeFile(ftype, dst, perms string, pr io.Reader) error {
	var err error

	// wrap reader to detect empty tar/gz and DON'T error if the variable
	// does not exist. It counts offset

	switch ftype {
	case "":
		fallthrough

	case "plain":
		err = writeFile(dst, perms, pr)
		if err != nil {
			return err
		}

	case "base64":
		r := base64.NewDecoder(base64.StdEncoding, pr)
		err = writeFile(dst, perms, r)
		if err != nil {
			return err
		}

	case "tar":

		wr := &WrapperReader{
			rd: pr,
		}

		err = untar(dst, perms, wr)
		if err != nil && wr.offset > 0 {
			return err
		}

	case "tar.gz":

		wr := &WrapperReader{
			rd: pr,
		}

		gr, err := gzip.NewReader(pr)
		if err != nil {
			if wr.offset == 0 {
				err = os.MkdirAll(dst, 0o750)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						return err
					}
				}
				return nil
			}
			return err
		}

		err = untar(dst, perms, gr)
		if err != nil && wr.offset > 0 {
			return err
		}

		err = gr.Close()
		if err != nil {
			return err
		}

	default:
		panic(ftype)
	}

	return nil
}

func (worker *inboundWorker) fileWriter(ctx context.Context, ir *functionRequest, f *functionFiles, pr *io.PipeReader) error {
	slog.Info("starting writer", "f", f)

	dir := worker.functionDir(ir)
	dst := f.Key
	if f.As != "" {
		dst = f.As
	}
	dst = filepath.Join(dir, dst)
	dir, _ = filepath.Split(dst)

	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return err
	}

	err = worker.writeFile(f.Type, dst, f.Permissions, pr)
	if err != nil {
		return err
	}

	return nil
}

func (worker *inboundWorker) functionDir(ir *functionRequest) string {
	return filepath.Join(sharedDir, ir.actionId)
}

func (worker *inboundWorker) cleanupFunctionRequest(ir *functionRequest) {
	dir := worker.functionDir(ir)
	err := os.RemoveAll(dir)
	if err != nil {
		slog.Error("cleanup function", "error", err)
	}
}

func (worker *inboundWorker) prepFunctionRequest(ctx context.Context, ir *functionRequest) (int, error) {
	statusCode, err := worker.prepFunctionFiles(ctx, ir)
	if err != nil {
		return statusCode, fmt.Errorf("failed to prepare functions files: %w", err)
	}

	return statusCode, nil
}

func (worker *inboundWorker) prepFunctionFiles(ctx context.Context, ir *functionRequest) (int, error) {
	dir := worker.functionDir(ir)

	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	statusCode, err := fetchFunctionFiles(ctx, worker.srv.flowToken, worker.srv.flowAddr, ir, worker.fileWriter)
	if err != nil {
		return statusCode, err
	}
	subDirs := []string{utils.VarScopeFileSystem, utils.VarScopeNamespace, utils.VarScopeWorkflow, utils.VarScopeInstance}
	for _, d := range subDirs {
		err := os.MkdirAll(path.Join(dir, fmt.Sprintf("out/%s", d)), 0o777)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to prepare function output dirs: %w", err)
		}
	}

	return statusCode, nil
}

func fetchFunctionFiles(ctx context.Context, flowToken string, flowAddr string, ir *functionRequest, fileWriter func(context.Context, *functionRequest, *functionFiles, *io.PipeReader) error) (int, error) {
	namespaceVariables, statusCode, err := getNamespaceVariables(ctx, flowToken, flowAddr, ir)
	if err != nil {
		return statusCode, fmt.Errorf("failed to get namespace variables: %w", err)
	}

	workflowVariables, statusCode, err := getWorkflowVariables(ctx, flowToken, flowAddr, ir)
	if err != nil {
		return statusCode, fmt.Errorf("failed to get workflow variables: %w", err)
	}

	instanceVariables, statusCode, err := getInstanceVariables(ctx, flowToken, flowAddr, ir)
	if err != nil {
		return statusCode, fmt.Errorf("failed to get instance variables: %w", err)
	}
	vars := append(namespaceVariables.Data, workflowVariables.Data...)
	vars = append(vars, instanceVariables.Data...)
	slog.Info("variables for processing", "data", fmt.Sprintf("%v", vars))

	for i := range ir.files {
		file := ir.files[i]
		typ, err := determineVarType(file.Scope)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to determine variable type: %w", err)
		}

		idx := slices.IndexFunc(vars, func(e variable) bool { return e.Typ == typ && e.Name == file.Key })
		pr, pw := io.Pipe()

		go func(flowToken string, flowAddr string, namespace string, file *functionFiles, idx int) {
			var data []byte
			var err error

			if typ == "file" {
				dataLocal, statusCode, err := getReferencedFile(ctx, flowToken, flowAddr, namespace, file.Key)
				if err != nil {
					slog.Info("Ok error, failed fetching file", "error", err, "statusCode", statusCode)
				}
				data = dataLocal
			}

			if idx > -1 {
				slog.Info("starting request api routine", "file", file, "idx", idx, "id", vars[idx].ID)
				addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables/%v", flowAddr, namespace, vars[idx].ID)

				client := &http.Client{}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to create new request: %w", err))
					return
				}
				req.Header.Set("Direktiv-Api-Key", flowToken)
				resp, err := client.Do(req)
				if err != nil {
					pw.CloseWithError(fmt.Errorf("failed to execute request: %w", err))
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					pw.CloseWithError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
					return
				}

				var variable variableResponse
				decoder := json.NewDecoder(resp.Body)
				if err = decoder.Decode(&variable); err != nil {
					pw.CloseWithError(fmt.Errorf("failed to decode response body: %w", err))
					return
				}
				data = variable.Data.Data
			}

			if _, err = io.Copy(pw, bytes.NewReader(data)); err != nil {
				pw.CloseWithError(fmt.Errorf("failed to copy data to pipe writer: %w", err))
				return
			}

			pw.Close()
		}(flowToken, flowAddr, ir.Namespace, file, idx)

		if err = fileWriter(ctx, ir, file, pr); err != nil {
			pr.CloseWithError(fmt.Errorf("failed to write file: %w", err))
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}

func determineVarType(fileScope string) (string, error) {
	switch fileScope {
	case utils.VarScopeFileSystem:
		return "file", nil
	case utils.VarScopeInstance:
		return "instance-variable", nil
	case utils.VarScopeWorkflow:
		return "workflow-variable", nil
	case utils.VarScopeNamespace:
		return "namespace-variable", nil
	case utils.VarScopeSystem:
	case utils.VarScopeThread:
	}

	return "", fmt.Errorf("Unknown scope")
}

func (worker *inboundWorker) handleFunctionRequest(req *inboundRequest) {
	defer func() {
		close(req.end)
	}()
	aid := req.r.Header.Get(actionIDHeader)
	maxCap := int64(134217728) // 4 MiB (cahnged to API value)
	if req.r.ContentLength == 0 {
		code := http.StatusLengthRequired
		worker.reportValidationError(aid, req.w, code, errors.New(http.StatusText(code)))
		return
	}
	if req.r.ContentLength > maxCap {
		worker.reportValidationError(aid, req.w, http.StatusRequestEntityTooLarge, fmt.Errorf("size limit: %d bytes", maxCap))
		return
	}

	action, err := enginerefactor.DecodeActionRequest(req.r)
	if err != nil {
		slog.Error("failed to construct action-data from request", "error", err)
		return
	}

	files := make([]*functionFiles, len(action.Files))
	for i := range action.Files {
		f := action.Files[i]
		files[i] = &functionFiles{
			Key:         f.Key,
			As:          f.As,
			Scope:       f.Scope,
			Type:        f.Type,
			Permissions: f.Permissions,
		}
	}
	ir := &functionRequest{
		actionId:      aid,
		deadline:      action.Deadline,
		input:         action.UserInput,
		files:         files,
		ActionContext: action.ActionContext,
	}
	if ir.deadline.IsZero() {
		ir.deadline = time.Now().Add(3 * time.Second)
	}

	ctx := req.r.Context()
	ctx, cancel := context.WithDeadline(ctx, ir.deadline)
	defer cancel()

	defer worker.cleanupFunctionRequest(ir)

	statusCode, err := worker.prepFunctionRequest(ctx, ir)
	if err != nil {
		worker.reportSidecarError(req.w, ir, fmt.Errorf("failed to prepare function request with status %v: %w", statusCode, err))
		return
	}

	logObject := telemetry.LogObject{
		Namespace: ir.Namespace,
		ID:        aid,
		Scope:     telemetry.LogScopeInstance,
		InstanceInfo: telemetry.InstanceInfo{
			Invoker:  ir.Invoker,
			Callpath: ir.Callpath,
			Path:     ir.Workflow,
			State:    ir.State,
			Status:   core.LogRunningStatus,
		},
	}

	rctx := telemetry.LogInitCtx(context.Background(), logObject)
	// rctx = tracing.AddNamespace(rctx, ir.Namespace)
	// rctx = tracing.AddInstanceMemoryAttr(rctx, tracing.InstanceAttributes{
	// 	Namespace:    ir.Namespace,
	// 	InstanceID:   ir.Instance,
	// 	Status:       core.LogUnknownStatus,
	// 	WorkflowPath: ir.Workflow,
	// 	Callpath:     ir.Callpath,
	// }, ir.State)
	// rctx = tracing.WithTrack(rctx, tracing.BuildInstanceTrackViaCallpath(ir.Callpath))
	// rctx = tracing.AddActionID(rctx, aid)
	// rctx = tracing.AddNamespace(rctx, ir.Namespace)
	// rctx = tracing.AddStateAttr(rctx, ir.State)
	// rctx, end, err2 := tracing.NewSpan(rctx, "handle function request")
	// if err2 != nil {
	// 	slog.Debug("failed while doFunctionRequest", "error", err2)
	// }
	// defer end()
	// rctx, span, err2 := tracing.InjectTraceParent(rctx, ir.ActionContext.TraceParent, "action registered for execution: "+ir.actionId+", workflow: "+ir.Workflow)
	// if err2 != nil {
	// 	slog.Warn("failed while doFunctionRequest", "error", err2)
	// }
	// defer span.End()

	worker.srv.registerActiveRequest(ir, rctx, cancel)
	defer worker.srv.deregisterActiveRequest(ir.actionId)
	go func() {
		select {
		case <-rctx.Done():
		case <-ctx.Done():
			worker.srv.cancelActiveRequest(rctx, ir.actionId)
		}
	}()

	out, err := worker.doFunctionRequest(rctx, ir)
	if err != nil {
		slog.Error("failed while doFunctionRequest", "error", err)
		worker.reportSidecarError(req.w, ir, err)
		return
	}

	// fetch output variables
	statusCode, err = worker.setOutVariables(rctx, ir)
	if err != nil {
		slog.Error("failed while setOutVariables", "error", err, "statusCode", statusCode)
		worker.reportSidecarError(req.w, ir, err)
		return
	}

	worker.respondToFlow(req.w, ir.actionId, out)
}

func (worker *inboundWorker) setOutVariables(ctx context.Context, ir *functionRequest) (int, error) {
	subDirs := []string{utils.VarScopeFileSystem, utils.VarScopeNamespace, utils.VarScopeWorkflow, utils.VarScopeInstance}
	var statusCode int
	for _, d := range subDirs {
		out := path.Join(worker.functionDir(ir), "out", d)

		files, err := os.ReadDir(out)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("can not read out folder: %w", err)
		}

		for _, f := range files {
			fp := path.Join(worker.functionDir(ir), "out", d, f.Name())

			fi, err := f.Info()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			switch mode := fi.Mode(); {
			case mode.IsDir():

				tf, err := os.CreateTemp("", "outtar")
				if err != nil {
					return http.StatusInternalServerError, err
				}

				err = tarGzDir(fp, tf)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				defer os.Remove(tf.Name())

				_, err = tf.Seek(0, io.SeekStart)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				statusCode, err = worker.srv.setVar(ctx, ir, tf, d, f.Name(), "")
				if err != nil {
					slog.Error("failed to set variable", "error", err)
					return statusCode, err
				}
			case mode.IsRegular():

				/* #nosec */
				v, err := os.Open(fp)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				statusCode, err = worker.srv.setVar(ctx, ir, v, d, f.Name(), "")
				if err != nil {
					_ = v.Close()
					return statusCode, err
				}

				err = v.Close()
				if err != nil {
					return http.StatusInternalServerError, err
				}
			}
		}
	}

	return statusCode, nil
}

func tarGzDir(src string, buf io.Writer) error {
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsDir() && !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// use "subpath"
		header.Name = filepath.ToSlash(file[len(src):])

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			/* #nosec */
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}

func (worker *inboundWorker) respondToFlow(w http.ResponseWriter, actionId string, out *outcome) {
	ar := enginerefactor.ActionResponse{
		Output:  out.data,
		ErrMsg:  out.errMsg,
		ErrCode: out.errCode,
	}
	w.Header().Add(flow.DirektivActionIDHeader, actionId)
	b, err := json.Marshal(ar)
	if err != nil {
		slog.Error("failed to report results for request", "action_id", actionId, "error", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		slog.Error("failed to write results for request", "action_id", actionId, "error", err)
		return
	}
	if out.errCode != "" {
		slog.Error("request failed with catchable", "action_id", actionId, "action_err_code", out.errCode, "error", out.errMsg)
	} else if out.errMsg != "" {
		slog.Error("request failed with uncatchable service error", "action_id", actionId, "error", out.errMsg)
	} else {
		slog.Info("request completed successfully", "action_id", actionId)
	}
}

func (worker *inboundWorker) reportSidecarError(w http.ResponseWriter, ir *functionRequest, err error) {
	worker.respondToFlow(w, ir.actionId, &outcome{
		errMsg: err.Error(),
	})
}

func (worker *inboundWorker) reportValidationError(id string, w http.ResponseWriter, code int, err error) {
	msg := err.Error()
	http.Error(w, msg, code)
	slog.Warn("request returned due to failed validation", "action_id", id, "action_err_code", code, "error", err)
}
