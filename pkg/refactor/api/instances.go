package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InstanceData struct {
	ID           uuid.UUID           `json:"id"`
	CreatedAt    time.Time           `json:"createdAt"`
	EndedAt      *time.Time          `json:"endedAt"`
	Status       string              `json:"status"`
	WorkflowPath string              `json:"path"`
	ErrorCode    string              `json:"errorCode"`
	Invoker      string              `json:"invoker"`
	Definition   []byte              `json:"definition,omitempty"`
	ErrorMessage []byte              `json:"errorMessage,omitempty"`
	Flow         []string            `json:"flow"`
	TraceID      string              `json:"traceId"`
	Lineage      []engine.ParentInfo `json:"lineage"`

	InputLength    *int   `json:"inputLength,omitempty"`
	Input          []byte `json:"input,omitempty"`
	OutputLength   *int   `json:"outputLength,omitempty"`
	Output         []byte `json:"output,omitempty"`
	MetadataLength *int   `json:"metadataLength,omitempty"`
	Metadata       []byte `json:"metadata,omitempty"`
}

func marshalForAPI(data *instancestore.InstanceData) *InstanceData {
	resp := &InstanceData{
		ID:           data.ID,
		CreatedAt:    data.CreatedAt,
		EndedAt:      data.EndedAt,
		Status:       data.Status.String(),
		WorkflowPath: data.WorkflowPath,
		ErrorCode:    data.ErrorCode,
		Invoker:      data.Invoker,
		Definition:   data.Definition,
		ErrorMessage: data.ErrorMessage,
	}

	x, err := engine.ParseInstanceData(data)
	if err == nil {
		resp.Flow = x.RuntimeInfo.Flow
		resp.TraceID = x.TelemetryInfo.TraceID
		resp.Lineage = x.DescentInfo.Descent
	}

	return resp
}

type instController struct {
	db      *database.SQLStore
	manager *instancestore.InstanceManager
}

func (e *instController) mountRouter(r chi.Router) {
	r.Get("/{instanceID}/subscribe", e.stream)

	r.Get("/{instanceID}/input", e.input)
	r.Get("/{instanceID}/output", e.output)
	r.Get("/{instanceID}/metadata", e.metadata)

	r.Get("/{instanceID}", e.get)
	r.Patch("/{instanceID}", e.update)

	r.Get("/", e.list)
	r.Post("/", e.create)
}

func (e *instController) blob(w http.ResponseWriter, r *http.Request) (*InstanceData, *instancestore.InstanceData) {
	ctx := r.Context()
	ns := extractContextNamespace(r)
	instanceID := chi.URLParam(r, "instanceID")

	id, err := uuid.Parse(instanceID)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: fmt.Errorf("unparsable instance UUID: %w", err).Error(),
		})

		return nil, nil
	}

	data, err := e.db.InstanceStore().ForInstanceID(id).GetSummaryWithInput(ctx)
	if err != nil {
		writeInstanceStoreError(w, err)

		return nil, nil
	}

	if data.Namespace != ns.Name {
		writeInstanceStoreError(w, instancestore.ErrNotFound)

		return nil, nil
	}

	// TODO: option to return the data raw

	resp := marshalForAPI(data)

	return resp, data
}

func (e *instController) input(w http.ResponseWriter, r *http.Request) {
	resp, data := e.blob(w, r)
	if resp != nil {
		resp.Input = data.Input

		l := len(data.Input)
		resp.InputLength = &l

		writeJSON(w, resp)
	}
}

func (e *instController) output(w http.ResponseWriter, r *http.Request) {
	resp, data := e.blob(w, r)
	if resp != nil {
		resp.Output = data.Output

		l := len(data.Output)
		resp.OutputLength = &l

		writeJSON(w, resp)
	}
}

func (e *instController) metadata(w http.ResponseWriter, r *http.Request) {
	resp, data := e.blob(w, r)
	if resp != nil {
		resp.Metadata = data.Metadata

		l := len(data.Metadata)
		resp.MetadataLength = &l

		writeJSON(w, resp)
	}
}

func (e *instController) getOnce(r *http.Request, instanceID uuid.UUID) (*instancestore.InstanceData, error) {
	ctx := r.Context()
	ns := extractContextNamespace(r)

	data, err := e.db.InstanceStore().ForInstanceID(instanceID).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	if data.Namespace != ns.Name {
		return nil, instancestore.ErrNotFound
	}

	return data, nil
}

func (e *instController) get(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceID")

	id, err := uuid.Parse(instanceID)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: fmt.Errorf("unparsable instance UUID: %w", err).Error(),
		})

		return
	}

	data, err := e.getOnce(r, id)
	if err != nil {
		writeInstanceStoreError(w, err)

		return
	}

	resp := marshalForAPI(data)
	resp.InputLength = &data.InputLength
	resp.OutputLength = &data.OutputLength
	resp.MetadataLength = &data.MetadataLength

	writeJSON(w, resp)
}

type cancelPayload struct {
	Status string
}

func (e *instController) update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ns := extractContextNamespace(r)
	instanceID := chi.URLParam(r, "instanceID")

	// TODO: parse input and confirm that this is clearly an attempt to cancel the instance

	pl := new(cancelPayload)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&pl)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: err.Error(),
		})

		return
	}

	if pl.Status != instancestore.InstanceStatusCancelled.String() {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "unsupported patch payload",
		})

		return
	}

	err = e.manager.Cancel(ctx, ns.Name, instanceID)
	if err != nil {
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}
}

type paginationOrderOptions struct {
	Field     string
	Direction string
}

type paginationFilterOptions struct {
	Field string
	Type  string
	Val   string
}

type paginationOptions struct {
	Limit  int
	Offset int
	Order  []*paginationOrderOptions
	Filter []*paginationFilterOptions
}

func (e *instController) getPagination(r *http.Request) (*paginationOptions, error) {
	opts := new(paginationOptions)

	x := r.URL.Query().Get("limit")
	if x != "" {
		k, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad 'limit' query parameter: %w", err)
		}
		opts.Limit = int(k)
	}

	x = r.URL.Query().Get("offset")
	if x != "" {
		k, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad 'offset' query parameter: %w", err)
		}
		opts.Offset = int(k)
	}

	var l1, l2, l3 int
	var ok bool
	var orderfields []string
	var orderdirection []string
	var orderings []*paginationOrderOptions
	if orderfields, ok = r.URL.Query()["order.field"]; ok {
		l1 = len(orderfields)
	}
	if orderdirection, ok = r.URL.Query()["order.direction"]; ok {
		l2 = len(orderdirection)
	}
	if l1 == 1 && l2 == 0 {
		ofield := orderfields[0]
		orderings = append(orderings, &paginationOrderOptions{
			Field: ofield,
		})
	} else {
		if l1 != l2 {
			return nil, errors.New("bad ordering arguments")
		}
		if l1 > 0 {
			for i := range orderfields {
				ofield := orderfields[i]
				direction := orderdirection[i]
				orderings = append(orderings, &paginationOrderOptions{
					Field:     ofield,
					Direction: direction,
				})
			}
		}
	}

	opts.Order = orderings

	l1 = 0
	l2 = 0
	var filterfields []string
	var filtertypes []string
	var filtervals []string
	var filters []*paginationFilterOptions
	if filterfields, ok = r.URL.Query()["filter.field"]; ok {
		l1 = len(filterfields)
	}
	if filtertypes, ok = r.URL.Query()["filter.type"]; ok {
		l2 = len(filtertypes)
	}
	if filtervals, ok = r.URL.Query()["filter.val"]; ok {
		l3 = len(filtervals)
	}
	if l1 != l2 || l1 != l3 {
		return nil, errors.New("bad filtering arguments")
	}
	if l1 > 0 {
		for i := range filterfields {
			ffield := filterfields[i]
			ftype := filtertypes[i]
			fval := filtervals[i]
			filters = append(filters, &paginationFilterOptions{
				Field: ffield,
				Type:  ftype,
				Val:   fval,
			})
		}
	}

	opts.Filter = filters

	return opts, nil
}

func (e *instController) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ns := extractContextNamespace(r)

	pagination, err := e.getPagination(r)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: fmt.Errorf("pagination options error: %w", err).Error(),
		})
	}

	opts := new(instancestore.ListOpts)
	if pagination != nil {
		opts.Limit = pagination.Limit
		opts.Offset = pagination.Offset

		for idx := range pagination.Order {
			x := pagination.Order[idx]
			var order instancestore.Order
			switch x.Direction {
			case "":
				fallthrough
			case "DESC":
				order.Descending = true
			case "ASC":
			default:
				writeError(w, &Error{
					Code:    "request_data_invalid",
					Message: fmt.Errorf("bad pagination direction: %s", x.Direction).Error(),
				})

				return
			}

			switch x.Field {
			case "CREATED":
				order.Field = instancestore.FieldCreatedAt
			default:
				order.Field = x.Field
			}

			opts.Orders = append(opts.Orders, order)
		}

		var err error

		for idx := range pagination.Filter {
			x := pagination.Filter[idx]
			var filter instancestore.Filter

			switch x.Type {
			case "CONTAINS":
				filter.Kind = instancestore.FilterKindContains
			case "WORKFLOW":
				filter.Kind = instancestore.FilterKindMatch
			case "PREFIX":
				filter.Kind = instancestore.FilterKindPrefix
			case "MATCH":
				filter.Kind = instancestore.FilterKindMatch
			case "AFTER":
				filter.Kind = instancestore.FilterKindAfter
			case "BEFORE":
				filter.Kind = instancestore.FilterKindBefore
			default:
				filter.Kind = x.Type
			}

			switch x.Field {
			case "AS":
				filter.Field = instancestore.FieldWorkflowPath
				filter.Value = x.Val
			case "CREATED":
				filter.Field = instancestore.FieldCreatedAt
				t, err := time.Parse(time.RFC3339, x.Val)
				if err != nil {
					writeError(w, &Error{
						Code:    "request_data_invalid",
						Message: fmt.Errorf("invalid filter value: %w", err).Error(),
					})

					return
				}
				filter.Value = t.UTC()
			case "STATUS":
				filter.Field = instancestore.FieldStatus
				filter.Value, err = instancestore.InstanceStatusFromString(x.Val)
				if err != nil {
					writeError(w, &Error{
						Code:    "request_data_invalid",
						Message: fmt.Errorf("invalid filter value: %w", err).Error(),
					})

					return
				}
			case "TRIGGER":
				filter.Field = instancestore.FieldInvoker
				filter.Value = x.Val
			default:
				filter.Field = x.Field
				filter.Value = x.Val
			}

			opts.Filters = append(opts.Filters, filter)
		}
	}

	data, err := e.db.InstanceStore().GetNamespaceInstances(ctx, ns.ID, opts)
	if err != nil {
		writeInstanceStoreError(w, err)

		return
	}

	writeJSON(w, data)
}

func (e *instController) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ns := extractContextNamespace(r)
	path := r.URL.Query().Get("path")
	input, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	data, err := e.manager.Start(ctx, ns.Name, path, input)
	if err != nil {
		writeError(w, &Error{
			Code:    err.Error(),
			Message: err.Error(),
		})

		return
	}

	if r.URL.Query().Get("wait") == "true" {
		e.handleWait(ctx, w, r, data)

		return
	}

	writeJSON(w, marshalForAPI(data))
}

func (e *instController) handleWait(ctx context.Context, w http.ResponseWriter, r *http.Request, data *instancestore.InstanceData) {
	var err error

	id := data.ID
	dt := time.Millisecond * 100
	ddt := dt
	dtMax := time.Second

recheck:

	time.Sleep(dt)
	dt += ddt
	if dt > dtMax {
		dt = dtMax
	}

	data, err = e.db.InstanceStore().ForInstanceID(id).GetSummaryWithOutput(ctx)
	if err != nil {
		writeInstanceStoreError(w, err)

		return
	}

	if data.Status == instancestore.InstanceStatusPending {
		goto recheck
	}

	if data.Status > instancestore.InstanceStatusComplete {
		w.Header().Set("Direktiv-Instance-Error-Code", data.ErrorCode)
		w.Header().Set("Direktiv-Instance-Error-Message", string(data.ErrorMessage))

		writeError(w, &Error{
			Code:    data.ErrorCode,
			Message: string(data.ErrorMessage),
		})

		return
	}

	raw := data.Output

	field := r.URL.Query().Get("field")
	if field != "" {
		m := make(map[string]interface{})
		_ = json.Unmarshal(raw, &m)

		x, exists := m[field]
		if exists {
			raw, _ = json.Marshal(x)
		} else {
			raw, _ = json.Marshal(nil)
		}
	}

	var x interface{}

	_ = json.Unmarshal(raw, &x)

	rawo, _ := strconv.ParseBool(r.URL.Query().Get("raw"))

	if rawo {
		if x == nil {
			raw = make([]byte, 0)
		} else if str, ok := x.(string); ok {
			raw = []byte(str)
			b64, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				raw = b64
			}
		}
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(raw)))

	ctype := r.URL.Query().Get("ctype")
	if ctype == "" {
		mtype := mimetype.Detect(raw)
		ctype = mtype.String()
	}

	w.Header().Set("Content-Type", ctype)

	_, _ = io.Copy(w, bytes.NewReader(raw))
}

func (e *instController) stream(w http.ResponseWriter, r *http.Request) {
	// Set the appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(time.Second)

	// TODO: do we need to deduplicate events?

	instanceID := chi.URLParam(r, "instanceID")

	id, err := uuid.Parse(instanceID)
	if err != nil {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: fmt.Errorf("unparsable instance UUID: %w", err).Error(),
		})

		return
	}

	for {
		data, err := e.getOnce(r, id)
		if err != nil {
			return // TODO: how are we supposed to report errors in SSE?
		}

		resp := marshalForAPI(data)
		resp.InputLength = &data.InputLength
		resp.OutputLength = &data.OutputLength
		resp.MetadataLength = &data.MetadataLength

		raw, _ := json.Marshal(resp)

		dst := &bytes.Buffer{}
		_ = json.Compact(dst, raw)

		_, _ = io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: message\ndata: %v\n\n", uuid.New(), dst.String())))

		//nolint:forcetypeassert
		w.(http.Flusher).Flush()

		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func writeInstanceStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, instancestore.ErrNotFound) {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: err.Error(),
		})

		return
	}

	if errors.Is(err, instancestore.ErrBadListOpts) {
		writeError(w, &Error{
			Code:    "request_invalid_list_options",
			Message: err.Error(),
		})

		return
	}

	writeInternalError(w, err)
}
