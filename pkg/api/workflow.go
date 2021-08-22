package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/itchyny/gojq"
	"github.com/rung/go-safecast"
	"gopkg.in/yaml.v2"

	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/util"
)

func (h *Handler) getUIDforName(ctx context.Context, ns, name string) (string, error) {

	ctx, cancel := CtxDeadline(ctx)
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowByName(ctx, &ingress.GetWorkflowByNameRequest{
		Namespace: &ns,
		Name:      &name,
	})

	if err != nil {
		return "", err
	}

	return resp.GetUid(), nil

}

func (h *Handler) workflows(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflows(ctx, &ingress.GetWorkflowsRequest{
		Namespace: &n,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	// wip uids
	for _, t := range resp.Workflows {
		t.Uid = nil
	}

	writeData(resp, w)

}

func (h *Handler) getWorkflow(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	getRefs := false
	if r.URL.Query().Get("get_refs") != "" {
		getRefs = true
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowByName(ctx, &ingress.GetWorkflowByNameRequest{
		Namespace:     &n,
		Name:          &name,
		GetReferences: &getRefs,
	})

	if err != nil {
		ErrResponse(w, err)
		return
	}

	// wipe uid
	resp.Uid = nil

	writeData(resp, w)

}

func (h *Handler) updateWorkflow(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var useRevision bool
	rev, err := strconv.Atoi(r.URL.Query().Get("revision"))
	if err == nil {
		useRevision = true
	}

	var active bool
	var useActive bool
	if val, ok := r.URL.Query()["active"]; ok {
		if val[0] == "true" {
			active = true
			useActive = true
		} else if val[0] == "false" {
			useActive = true
		}
	}

	var logEvent string
	var useLogEvent bool
	if val, ok := r.URL.Query()["logEvent"]; ok {
		logEvent = val[0]
		useLogEvent = true
	}

	// revision := int32(rev)
	revision, err := safecast.Int32(rev)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	// Read Body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var contentType string
	if typeMap, ok := r.Header["Content-Type"]; ok {
		contentType = typeMap[0]
	}

	// check name of workflow in updated workflow yaml
	m := make(map[string]interface{})

	switch contentType {
	case "text/yaml":
		err = yaml.Unmarshal(b, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to marshal yaml body"))
		}
	default:
		ErrResponse(w, fmt.Errorf("content type '%s' is not supported. supported media types: 'text/yaml'", contentType))
		return
	}

	if x, ok := m["id"]; ok {
		if str, ok := x.(string); ok {
			if ok = util.MatchesRegex(str); !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(errWorkflowRegex.Error()))
				return
			}
		}
	}

	// Construct direktiv GRPC Request
	request := ingress.UpdateWorkflowRequest{
		Uid:      &uid,
		Workflow: b,
	}

	if useActive {
		request.Active = &active
	}

	if useLogEvent {
		request.LogToEvents = &logEvent
	}

	if useRevision {
		request.Revision = &revision
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.UpdateWorkflow(ctx, &request)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)
}

func (h *Handler) toggleWorkflow(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowByUid(ctx, &ingress.GetWorkflowByUidRequest{
		Uid: &uid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	active := !*resp.Active

	resp2, err := h.s.direktiv.UpdateWorkflow(ctx, &ingress.UpdateWorkflowRequest{
		Uid:      &uid,
		Active:   &active,
		Workflow: resp.Workflow,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp2, w)

}

func (h *Handler) createWorkflow(w http.ResponseWriter, r *http.Request) {

	n := mux.Vars(r)["namespace"]
	active := true

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	var contentType string
	if typeMap, ok := r.Header["Content-Type"]; ok {
		contentType = typeMap[0]
	}

	if contentType != "text/yaml" {
		ErrResponse(w, fmt.Errorf("content type '%s' is not supported. supported media types: 'text/yaml'", contentType))
		return
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to marshal yaml body"))
		return
	}

	if x, ok := m["id"]; ok {
		if str, ok := x.(string); ok {
			if ok = util.MatchesRegex(str); !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(errWorkflowRegex.Error()))
				return
			}
		}
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.AddWorkflow(ctx, &ingress.AddWorkflowRequest{
		Namespace: &n,
		Active:    &active,
		Workflow:  b,
	})

	if err != nil {
		ErrResponse(w, err)
		return
	}

	// empty UID which omits it
	resp.Uid = nil

	retData, err := json.Marshal(resp)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	/* #nosec */
	_, _ = w.Write(retData)

}

func (h *Handler) deleteWorkflow(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.DeleteWorkflow(ctx, &ingress.DeleteWorkflowRequest{
		Uid: &uid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}

func (h *Handler) downloadWorkflow(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	uid, err := h.getUIDforName(r.Context(), ns, name)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetWorkflowByUid(ctx, &ingress.GetWorkflowByUidRequest{
		Uid: &uid,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+*resp.Id+".yaml")
	w.Header().Set("Content-Type", "application/x-yaml")

	writeData(resp, w)

}

func sendContent(w http.ResponseWriter, r *http.Request, data []byte) error {

	var in map[string]interface{}
	json.Unmarshal(data, &in)

	query, err := gojq.Parse(r.URL.Query().Get("field"))
	if err != nil {
		return err
	}

	// jq we get the first result
	iter := query.Run(in)
	v, ok := iter.Next()

	if ok {

		s, ok := v.(string)

		// if it is not a string, it is json
		if !ok {
			w.Header().Set("Content-Type", "application/json")

			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("can not return field: %v", err)
			}

			// return json
			_, err = w.Write(b)
			return err

		}

		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(s))

		buf := make([]byte, 16384)
		for {
			n, err := decoder.Read(buf)

			// check for error or non-base64 retData
			if err != nil && err != io.EOF {
				if strings.Contains(err.Error(), "illegal base64 data") {
					mimeType := http.DetectContentType([]byte(s))
					w.Header().Set("Content-Type", mimeType)
					w.Write([]byte(s))
					err = nil
				}
				return err
			}

			// here we can guess content type and set if in the first loop
			if w.Header().Get("Content-Type") == "" {
				mimeType := http.DetectContentType(buf[:n])
				w.Header().Set("Content-Type", mimeType)
			}

			w.Write(buf[:n])

			if err == io.EOF {
				err = nil
				break
			}
		}

	}

	return nil

}

func (h *Handler) executeWorkflow(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	name := mux.Vars(r)["workflowTarget"]

	// check query parameter 'wait'. if set it waits for workflow response
	// instead of the instance id only. this will still timeout after 30 seconds
	// which means it is for shorter workflows only
	wait := false
	if r.URL.Query().Get("wait") != "" {
		wait = true
	}

	field := ""
	if r.URL.Query().Get("field") != "" {
		field = r.URL.Query().Get("field")
	}

	var b []byte
	var err error
	if r.Method == http.MethodPost {
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			ErrResponse(w, err)
			return
		}
	} else {
		// read all query fields ignore wait and field parameters
		values := r.URL.Query()
		values.Del("wait")
		values.Del("field")

		body := make(map[string]interface{})
		for k := range values {
			if len(values[k]) > 1 {
				body[k] = values[k]
			} else {
				body[k] = values.Get(k)
			}
		}
		b, err = json.Marshal(&body)
		if err != nil {
			ErrResponse(w, err)
			return
		}

	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.InvokeWorkflow(ctx, &ingress.InvokeWorkflowRequest{
		Namespace: &ns,
		Name:      &name,
		Input:     b,
		Wait:      &wait,
	})

	if err != nil {
		ErrResponse(w, err)
		return
	}

	if resp.ErrorCode != nil {
		w.Header().Set("Direktiv_ErrorCode", *resp.ErrorCode)
	}

	if resp.ErrorMsg != nil {
		w.Header().Set("Direktiv_ErrorMsg", *resp.ErrorMsg)
	}

	// for wait there is special handling
	if wait && field != "" {

		err := sendContent(w, r, resp.Output)
		if err != nil {
			ErrResponse(w, err)
		}

	} else if wait {

		// simple wait without displaying data
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(direktiv.DirektivInstanceIDHeader, *resp.InstanceId)
		w.Write(resp.Output)

	} else {

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			ErrResponse(w, err)
		}

	}

}

func (h *Handler) workflowInstances(w http.ResponseWriter, r *http.Request) {

	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflowTarget"]

	o, l := paginationParams(r)
	offset, err := safecast.Int32(o)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	limit, err := safecast.Int32(l)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	ctx, cancel := CtxDeadline(r.Context())
	defer cancel()

	resp, err := h.s.direktiv.GetInstancesByWorkflow(ctx, &ingress.GetInstancesByWorkflowRequest{
		Offset:    &offset,
		Limit:     &limit,
		Namespace: &ns,
		Workflow:  &wf,
	})
	if err != nil {
		ErrResponse(w, err)
		return
	}

	writeData(resp, w)

}
