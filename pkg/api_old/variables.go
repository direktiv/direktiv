package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vorteil/direktiv/pkg/ingress"
)

const grpcChunkSize = 2 * 1024 * 1024

func (h *Handler) workflowVariables(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.ListWorkflowVariables(ctx, &ingress.ListWorkflowVariablesRequest{
		WorkflowUid: &uid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)
}

func (h *Handler) setWorkflowVariable(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]
	wfVar := mux.Vars(r)["variable"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	var contentLength string
	if typeMap, ok := r.Header["Content-Length"]; ok {
		contentLength = typeMap[0]
	}

	l, err := strconv.Atoi(contentLength)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	totalSize := int64(l)
	chunkSize := int64(grpcChunkSize)

	client, err := h.s.direktiv.SetWorkflowVariable(ctx)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var totalRead int64
	var chunks int

	for {
		buf := new(bytes.Buffer)
		rdr := io.LimitReader(r.Body, chunkSize)
		var k int64
		k, err = io.Copy(buf, rdr)
		totalRead += k
		if err != nil {
			ErrResponse(w, err)
			return
		}

		if k == 0 && chunks > 0 {
			break
		}

		data := buf.Bytes()

		req := new(ingress.SetWorkflowVariableRequest)
		req.WorkflowUid = &uid
		req.Key = &wfVar
		req.Value = data
		req.TotalSize = &totalSize
		req.ChunkSize = &chunkSize
		err = client.Send(req)
		if err != nil {
			ErrResponse(w, err)
			return
		}

		chunks++
		if totalRead >= totalSize {
			break
		}
	}

	if err != nil {
		ErrResponse(w, err)
		return
	}

	// Wait for Server To Return
	for {
		err = client.RecvMsg(nil)
		if err == io.EOF {
			// Server has completed operations, break out of wait
			break
		} else if err != nil {
			ErrResponse(w, err)
			return
		}
	}

	writeData(map[string]interface{}{}, w)
}

func (h *Handler) getWorkflowVariable(w http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]
	wfVar := mux.Vars(r)["variable"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	client, err := h.s.direktiv.GetWorkflowVariable(ctx, &ingress.GetWorkflowVariableRequest{
		WorkflowUid: &uid,
		Key:         &wfVar,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	in, err := client.Recv()
	if err != nil {
		ErrResponse(w, err)
		return
	}

	// TODO: Check this works with large variables
	k, err := io.Copy(w, bytes.NewReader(in.GetValue()))
	if err != nil {
		ErrResponse(w, err)
		return
	}

	if k == 0 {
		ErrResponse(w, status.Errorf(codes.NotFound, "variable %s does not exist", wfVar))
		return
	}
}

func (h *Handler) namespaceVariables(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.ListNamespaceVariables(ctx, &ingress.ListNamespaceVariablesRequest{
		Namespace: &ns,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)
}

func (h *Handler) setNamespaceVariable(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	wfVar := mux.Vars(r)["variable"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	var contentLength string
	if typeMap, ok := r.Header["Content-Length"]; ok {
		contentLength = typeMap[0]
	}

	l, err := strconv.Atoi(contentLength)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	totalSize := int64(l)
	chunkSize := int64(grpcChunkSize)

	client, err := h.s.direktiv.SetNamespaceVariable(ctx)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var totalRead int64
	var chunks int

	for {
		buf := new(bytes.Buffer)
		rdr := io.LimitReader(r.Body, chunkSize)
		var k int64
		k, err = io.Copy(buf, rdr)
		totalRead += k
		if err != nil {
			ErrResponse(w, err)
			return
		}

		if k == 0 && chunks > 0 {
			break
		}

		data := buf.Bytes()

		req := new(ingress.SetNamespaceVariableRequest)
		req.Namespace = &ns
		req.Key = &wfVar
		req.Value = data
		req.TotalSize = &totalSize
		req.ChunkSize = &chunkSize
		err = client.Send(req)
		if err != nil {
			ErrResponse(w, err)
			return
		}

		chunks++
		if totalRead >= totalSize {
			break
		}
	}

	if err != nil {
		ErrResponse(w, err)
		return
	}

	// Wait for Server To Return
	for {
		err = client.RecvMsg(nil)
		if err == io.EOF {
			// Server has completed operations, break out of wait
			break
		} else if err != nil {
			ErrResponse(w, err)
			return
		}
	}

	writeData(map[string]interface{}{}, w)
}

func (h *Handler) getNamespaceVariable(w http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	wfVar := mux.Vars(r)["variable"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	client, err := h.s.direktiv.GetNamespaceVariable(ctx, &ingress.GetNamespaceVariableRequest{
		Namespace: &ns,
		Key:       &wfVar,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	in, err := client.Recv()
	if err != nil {
		ErrResponse(w, err)
		return
	}

	// TODO: Check this works with large variables
	k, err := io.Copy(w, bytes.NewReader(in.GetValue()))
	if err != nil {
		ErrResponse(w, err)
		return
	}

	if k == 0 {
		ErrResponse(w, status.Errorf(codes.NotFound, "variable %s does not exist", wfVar))
		return
	}
}
