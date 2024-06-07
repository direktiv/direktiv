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

	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/google/uuid"
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
		slog.Debug("Cancelling worker.", "worker_id", worker.id)
		worker.cancel()
	}

	worker.lock.Unlock()
}

func (worker *inboundWorker) run() {
	slog.Debug("Starting worker", "worker_id", worker.id)

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
		slog.Debug("Worker picked up request.", "worker_id", worker.id, "action_id", id)

		worker.handleFunctionRequest(req)
	}

	slog.Debug("Worker shut down.", "worker_id", worker.id)
}

type outcome struct {
	data    []byte
	errCode string
	errMsg  string
}

// nolint:canonicalheader
func (worker *inboundWorker) doFunctionRequest(ctx context.Context, ir *functionRequest) (*outcome, error) {
	slog.Debug("Forwarding request to service.", "action_id", ir.actionId)

	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(ir.input))
	if err != nil {
		return nil, err
	}

	req.Header.Set(actionIDHeader, ir.actionId)
	req.Header.Set(IteratorHeader, fmt.Sprintf("%d", ir.iterator))
	req.Header.Set("Direktiv-TempDir", worker.functionDir(ir))
	req.Header.Set("Content-Type", "application/json")

	cleanup := utils.TraceHTTPRequest(ctx, req)
	defer cleanup()

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

func (worker *inboundWorker) prepFunctionRequest(ctx context.Context, ir *functionRequest) error {
	err := worker.prepFunctionFiles(ctx, ir)
	if err != nil {
		return fmt.Errorf("failed to prepare functions files: %w", err)
	}

	return nil
}

func (worker *inboundWorker) prepFunctionFiles(ctx context.Context, ir *functionRequest) error {
	dir := worker.functionDir(ir)

	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return err
	}
	err = worker.fetchFunctionFiles(ctx, ir)
	if err != nil {
		return err
	}
	subDirs := []string{utils.VarScopeFileSystem, utils.VarScopeNamespace, utils.VarScopeWorkflow, utils.VarScopeInstance}
	for _, d := range subDirs {
		err := os.MkdirAll(path.Join(dir, fmt.Sprintf("out/%s", d)), 0o777)
		if err != nil {
			return fmt.Errorf("failed to prepare function output dirs: %w", err)
		}
	}

	return nil
}

type variableForAPI struct {
	ID        uuid.UUID `json:"id"`
	Typ       string    `json:"type"`
	Reference string    `json:"reference"`
	Name      string    `json:"name"`
	Data      []byte    `json:"data"`
	Size      int       `json:"size"`
	MimeType  string    `json:"mimeType"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type varResponseData struct {
	Data  []variableForAPI `json:"data"`
	Error *any             `json:"error"`
}

type varInduvidualResponseData struct {
	Data  variableForAPI `json:"data"`
	Error *any           `json:"error"`
}

func (worker *inboundWorker) getNamespaceVariables(ctx context.Context, ir *functionRequest) (*varResponseData, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables", worker.srv.flowAddr, ir.namespace)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Direktiv-Token", worker.srv.flowToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	namespaceVariables := varResponseData{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&namespaceVariables); err != nil {
		return nil, err
	}

	return &namespaceVariables, nil
}

func (worker *inboundWorker) getWorkflowVariables(ctx context.Context, ir *functionRequest) (*varResponseData, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables?workflowPath=%v", worker.srv.flowAddr, ir.namespace, ir.workflowPath)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Direktiv-Token", worker.srv.flowToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	workflowVariables := varResponseData{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&workflowVariables); err != nil {
		return nil, err
	}

	return &workflowVariables, nil
}

func (worker *inboundWorker) getInstanceVariables(ctx context.Context, ir *functionRequest) (*varResponseData, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables?instanceId=%v", worker.srv.flowAddr, ir.namespace, ir.instanceId)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Direktiv-Token", worker.srv.flowToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	instanceVariables := varResponseData{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&instanceVariables); err != nil {
		return nil, err
	}

	return &instanceVariables, nil
}

func (worker *inboundWorker) fetchFunctionFiles(ctx context.Context, ir *functionRequest) error {
	namespaceVariables, err := worker.getNamespaceVariables(ctx, ir)
	if err != nil {
		return err
	}
	slog.Info("ns vars for processing", "data", fmt.Sprintf("%v", namespaceVariables))

	wfVar, err := worker.getWorkflowVariables(ctx, ir)
	if err != nil {
		return err
	}
	slog.Info("wf for processing", "data", fmt.Sprintf("%v", wfVar))

	insVars, err := worker.getInstanceVariables(ctx, ir)
	if err != nil {
		return err
	}
	slog.Info("ins for processing", "data", fmt.Sprintf("%v", insVars))

	vars := append(namespaceVariables.Data, wfVar.Data...)
	vars = append(vars, insVars.Data...)
	slog.Info("variables for processing", "data", fmt.Sprintf("%v", vars))
	for i := range ir.files {
		file := ir.files[i]
		typ, err := determineVarType(file.Scope)
		if err != nil {
			return err
		}
		idx := slices.IndexFunc(vars, func(e variableForAPI) bool { return e.Typ == typ && e.Name == file.Key })
		pr, pw := io.Pipe()
		go func(flowToken string, flowAddr string, namespace string, file *functionFiles, idx int) {
			var data []byte = nil
			var err error
			slog.Info("starting creating variables", "file", file, "idx", idx)

			if idx > -1 {
				slog.Info("starting request api rotine", "file", file, "idx", idx, "id", vars[idx].ID)
				addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables/%v", flowAddr, namespace, vars[idx].ID)
				client := &http.Client{}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
				if err != nil {
					sysErr := pw.CloseWithError(err)
					if sysErr != nil {
						slog.Error("failed fetching files with", "error", sysErr)
					}
				}
				req.Header.Set("Direktiv-Token", flowToken)
				resp, err := client.Do(req)
				if err != nil {
					sysErr := pw.CloseWithError(err)
					if sysErr != nil {
						slog.Error("failed fetching files with", "error", sysErr)
					}
				}
				defer resp.Body.Close()

				variable := varInduvidualResponseData{}
				decoder := json.NewDecoder(resp.Body)
				if err = decoder.Decode(&variable); err != nil {
					sysErr := pw.CloseWithError(err)
					if sysErr != nil {
						slog.Error("failed fetching files with", "error", sysErr)
					}
				}
				data = variable.Data.Data
			}
			slog.Info("copy data to for writer", "file", file, "idx", idx, "data", data)

			_, err = io.Copy(pw, bytes.NewReader(data))
			if err != nil {
				sysErr := pw.CloseWithError(err)
				if sysErr != nil {
					slog.Error("failed fetching files with", "error", sysErr)
				}
			}

			pw.Close()
		}(worker.srv.flowToken, worker.srv.flowAddr, ir.namespace, file, idx)

		err = worker.fileWriter(ctx, ir, file, pr)
		if err != nil {
			sysErr := pr.CloseWithError(err)
			if sysErr != nil {
				slog.Error("failed fetching files with", "error", sysErr)
			}

			return err
		}
	}

	return nil
}

func determineVarType(fileScope string) (string, error) {
	switch fileScope {
	case utils.VarScopeFileSystem:
		return "instance-variable", nil
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
		actionId:     aid,
		instanceId:   action.Instance,
		namespace:    action.Namespace,
		step:         action.Step,
		deadline:     action.Deadline,
		input:        action.UserInput,
		iterator:     action.Branch,
		files:        files,
		workflowPath: action.Workflow,
	}

	ctx := req.r.Context()
	ctx, cancel := context.WithDeadline(ctx, ir.deadline)
	defer cancel()

	defer worker.cleanupFunctionRequest(ir)

	err = worker.prepFunctionRequest(ctx, ir)
	if err != nil {
		worker.reportSidecarError(req.w, ir, err)
		return
	}

	// NOTE: rctx exists because we don't want to immediately cancel the function request if our context is cancelled
	rctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rctx = utils.TransplantTelemetryContextInformation(ctx, rctx)

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
		worker.reportSidecarError(req.w, ir, err)
		return
	}

	// fetch output variables
	err = worker.setOutVariables(rctx, ir)
	if err != nil {
		worker.reportSidecarError(req.w, ir, err)
		return
	}

	worker.respondToFlow(req.w, ir.actionId, out)
}

func (worker *inboundWorker) setOutVariables(ctx context.Context, ir *functionRequest) error {
	subDirs := []string{utils.VarScopeFileSystem, utils.VarScopeNamespace, utils.VarScopeWorkflow, utils.VarScopeInstance}
	for _, d := range subDirs {
		out := path.Join(worker.functionDir(ir), "out", d)

		files, err := os.ReadDir(out)
		if err != nil {
			return fmt.Errorf("can not read out folder: %w", err)
		}

		for _, f := range files {
			fp := path.Join(worker.functionDir(ir), "out", d, f.Name())

			fi, err := f.Info()
			if err != nil {
				return err
			}

			switch mode := fi.Mode(); {
			case mode.IsDir():

				tf, err := os.CreateTemp("", "outtar")
				if err != nil {
					return err
				}

				err = tarGzDir(fp, tf)
				if err != nil {
					return err
				}
				defer os.Remove(tf.Name())

				var end int64
				end, err = tf.Seek(0, io.SeekEnd)
				if err != nil {
					return err
				}

				_, err = tf.Seek(0, io.SeekStart)
				if err != nil {
					return err
				}

				err = worker.srv.setVar(ctx, ir, end, tf, d, f.Name(), "")
				if err != nil {
					return err
				}
			case mode.IsRegular():

				/* #nosec */
				v, err := os.Open(fp)
				if err != nil {
					return err
				}

				err = worker.srv.setVar(ctx, ir, fi.Size(), v, d, f.Name(), "")
				if err != nil {
					_ = v.Close()
					return err
				}

				err = v.Close()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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
		slog.Error("Failed to report results for request.", "action_id", actionId, "error", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		slog.Error("Failed to write results for request.", "action_id", actionId, "error", err)
		return
	}
	if out.errCode != "" {
		slog.Error("Request failed with catchable", "action_id", actionId, "action_err_code", out.errCode, "error", out.errMsg)
	} else if out.errMsg != "" {
		slog.Error("Request failed with uncatchable service error.", "action_id", actionId, "error", out.errMsg)
	} else {
		slog.Info("Request completed successfully.", "action_id", actionId)
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
	slog.Warn("Request returned due to failed validation.", "action_id", id, "action_err_code", code, "error", err)
}
