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
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/util"
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
		log.Debugf("Cancelling worker %d.", worker.id)
		worker.cancel()
	}

	worker.lock.Unlock()
}

func (worker *inboundWorker) run() {
	log.Debugf("Starting worker %d.", worker.id)

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
		log.Debugf("Worker %d picked up request '%s'.", worker.id, id)

		worker.handleFunctionRequest(req)
	}

	log.Debugf("Worker %d shut down.", worker.id)
}

func (worker *inboundWorker) fileReader(ctx context.Context, ir *functionRequest, f *functionFiles, pw *io.PipeWriter) error {
	err := worker.srv.getVar(ctx, ir, pw, nil, f.Scope, f.Key)
	if err != nil {
		return err
	}

	return nil
}

type outcome struct {
	data    []byte
	errCode string
	errMsg  string
}

func (worker *inboundWorker) doFunctionRequest(ctx context.Context, ir *functionRequest) (*outcome, error) {
	log.Debugf("Forwarding request '%s' to service.", ir.actionId)

	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(ir.input))
	if err != nil {
		return nil, err
	}

	req.Header.Set(actionIDHeader, ir.actionId)
	req.Header.Set(IteratorHeader, fmt.Sprintf("%d", ir.iterator))
	req.Header.Set("Direktiv-TempDir", worker.functionDir(ir))
	req.Header.Set("Content-Type", "application/json")

	cleanup := util.TraceHTTPRequest(ctx, req)
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

	cap := int64(134217728) // 128 MiB (changed to same value as API)
	if resp.ContentLength > cap {
		return nil, errors.New("service response is too large")
	}
	r := io.LimitReader(resp.Body, cap)

	out.data, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (worker *inboundWorker) prepOneFunctionFiles(ctx context.Context, ir *functionRequest, f *functionFiles) error {
	pr, pw := io.Pipe()

	go func() {
		err := worker.fileReader(ctx, ir, f, pw)
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
	}()

	err := worker.fileWriter(ctx, ir, f, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return err
	}

	_ = pr.Close()

	return nil
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
		log.Error(err)
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

	for i, f := range ir.files {
		err = worker.prepOneFunctionFiles(ctx, ir, f)
		if err != nil {
			return fmt.Errorf("failed to prepare function files %d: %w", i, err)
		}
	}

	subDirs := []string{util.VarScopeFileSystem, util.VarScopeNamespace, util.VarScopeWorkflow, util.VarScopeInstance}
	for _, d := range subDirs {
		err := os.MkdirAll(path.Join(dir, fmt.Sprintf("out/%s", d)), 0o777)
		if err != nil {
			return fmt.Errorf("failed to prepare function output dirs: %w", err)
		}
	}

	return nil
}

func (worker *inboundWorker) handleFunctionRequest(req *inboundRequest) {
	defer func() {
		close(req.end)
	}()

	ir := worker.validateFunctionRequest(req)
	if ir == nil {
		return
	}

	ctx := req.r.Context()
	ctx, cancel := context.WithDeadline(ctx, ir.deadline)
	defer cancel()

	defer worker.cleanupFunctionRequest(ir)

	err := worker.prepFunctionRequest(ctx, ir)
	if err != nil {
		worker.reportSidecarError(ir, err)
		return
	}

	// NOTE: rctx exists because we don't want to immediately cancel the function request if our context is cancelled
	rctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rctx = util.TransplantTelemetryContextInformation(ctx, rctx)

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
		worker.reportSidecarError(ir, err)
		return
	}

	// fetch output variables
	err = worker.setOutVariables(rctx, ir)
	if err != nil {
		worker.reportSidecarError(ir, err)
		return
	}

	worker.respondToFlow(rctx, ir, out)
}

func (worker *inboundWorker) setOutVariables(ctx context.Context, ir *functionRequest) error {
	subDirs := []string{util.VarScopeFileSystem, util.VarScopeNamespace, util.VarScopeWorkflow, util.VarScopeInstance}
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

func (worker *inboundWorker) respondToFlow(ctx context.Context, ir *functionRequest, out *outcome) {
	step := int32(ir.step)

	_, err := worker.srv.flow.ReportActionResults(ctx, &grpc.ReportActionResultsRequest{
		InstanceId:   ir.instanceId,
		Step:         step,
		ActionId:     ir.actionId,
		Iterator:     int32(ir.iterator),
		Output:       out.data,
		ErrorCode:    out.errCode,
		ErrorMessage: out.errMsg,
	})
	if err != nil {
		log.Errorf("Failed to report results for request '%s': %v.", ir.actionId, err)
		return
	}

	if out.errCode != "" {
		log.Errorf("Request '%s' failed with catchable error '%s': %s.", ir.actionId, out.errCode, out.errMsg)
	} else if out.errMsg != "" {
		log.Errorf("Request '%s' failed with uncatchable service error: %s.", ir.actionId, out.errMsg)
	} else {
		log.Infof("Request '%s' completed successfully.", ir.actionId)
	}
}

func (worker *inboundWorker) reportSidecarError(ir *functionRequest, err error) {
	ctx := context.Background()

	worker.respondToFlow(ctx, ir, &outcome{
		errMsg: err.Error(),
	})
}

func (worker *inboundWorker) reportValidationError(req *inboundRequest, code int, err error) {
	id := req.r.Header.Get(actionIDHeader)

	msg := err.Error()

	http.Error(req.w, msg, code)

	log.Warnf("Request '%s' returned %v due to failed validation: %v.", id, code, err)
}

func (worker *inboundWorker) getRequiredStringHeader(req *inboundRequest, x *string, hdr string) bool {
	s := req.r.Header.Get(hdr)
	*x = s
	if s == "" {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("missing %s", hdr))
		return false
	}

	return true
}

func (worker *inboundWorker) validateUintHeader(req *inboundRequest, x *int, hdr, s string) bool {
	var err error

	*x, err = strconv.Atoi(s)
	if err != nil {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s: %w", hdr, err))
		return false
	}
	if *x < 0 {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s value: %v", hdr, s))
		return false
	}

	return true
}

func (worker *inboundWorker) validateTimeHeader(req *inboundRequest, x *time.Time, hdr, s string) bool {
	var err error

	*x, err = time.Parse(time.RFC3339, s)
	if err != nil {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s: %w", hdr, err))
		return false
	}

	return true
}

func (worker *inboundWorker) loadBody(req *inboundRequest, data *[]byte) bool {
	cap := int64(134217728) // 4 MiB (cahnged to API value)
	if req.r.ContentLength == 0 {
		code := http.StatusLengthRequired
		worker.reportValidationError(req, code, errors.New(http.StatusText(code)))
		return false
	}
	if req.r.ContentLength > cap {
		worker.reportValidationError(req, http.StatusRequestEntityTooLarge, fmt.Errorf("size limit: %d bytes", cap))
		return false
	}
	r := io.LimitReader(req.r.Body, cap)

	var err error
	*data, err = io.ReadAll(r)
	if err != nil {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("failed to read request body: %w", err))
		return false
	}
	if int64(len(*data)) != req.r.ContentLength {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("request body doesn't match Content-Length"))
		return false
	}

	return true
}

func (worker *inboundWorker) validateFilesHeaders(req *inboundRequest, ifiles *[]*functionFiles) bool {
	hdr := "Direktiv-Files"
	strs := req.r.Header.Values(hdr)
	for i, s := range strs {
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %w", hdr, i, err))
			return false
		}

		files := new(functionFiles)
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		err = dec.Decode(files)
		if err != nil {
			worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %w", hdr, i, err))
			return false
		}

		*ifiles = append(*ifiles, files)
	}

	return true
}

func (worker *inboundWorker) validateFnCtxHeaders(req *inboundRequest, fnCtxStr string) (enginerefactor.FunctionContext, bool) {
	hdr := "Direktiv-Function-Context"
	data, err := base64.StdEncoding.DecodeString(fnCtxStr)
	if err != nil {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %w", hdr, 0, err))
		return enginerefactor.FunctionContext{}, false
	}

	fnCtx := new(enginerefactor.FunctionContext)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	err = dec.Decode(fnCtx)
	if err != nil {
		worker.reportValidationError(req, http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %w", hdr, 0, err))
		return enginerefactor.FunctionContext{}, false
	}

	return *fnCtx, true
}

func (worker *inboundWorker) validateFunctionRequest(req *inboundRequest) *functionRequest {
	ir := new(functionRequest)

	var step string
	var deadline string
	var it string
	var fnCtx string

	headers := []string{actionIDHeader, "Direktiv-InstanceID", "Direktiv-Namespace", "Direktiv-Step", "Direktiv-Iterator", "Direktiv-Deadline", "Direktiv-Function-Context"}
	ptrs := []*string{&ir.actionId, &ir.instanceId, &ir.namespace, &step, &it, &deadline, &fnCtx}

	for i := 0; i < len(headers); i++ {
		if !worker.getRequiredStringHeader(req, ptrs[i], headers[i]) {
			return nil
		}
	}

	if !worker.validateUintHeader(req, &ir.step, "Direktiv-Step", step) {
		return nil
	}

	if !worker.validateUintHeader(req, &ir.iterator, "Direktiv-Iterator", it) {
		return nil
	}

	if !worker.validateTimeHeader(req, &ir.deadline, "Direktiv-Deadline", deadline) {
		return nil
	}

	if !worker.loadBody(req, &ir.input) {
		return nil
	}

	if !worker.validateFilesHeaders(req, &ir.files) {
		return nil
	}
	var ok bool
	if ir.fnCtx, ok = worker.validateFnCtxHeaders(req, fnCtx); !ok {
		return nil
	}

	return ir
}
