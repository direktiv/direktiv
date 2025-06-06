package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/filestore"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TEMPORARY EVERYTHING

func (im *instanceMemory) BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error {
	return im.engine.events.BroadcastCloudevent(ctx, im.Namespace(), event, dd)
}

//nolint:gocognit
func (im *instanceMemory) GetVariables(ctx context.Context, vars []states.VariableSelector) ([]states.Variable, error) {
	x := make([]states.Variable, 0)

	tx, err := im.engine.flow.beginSQLTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, selector := range vars {
		//nolint:nestif
		if selector.Scope == "" || selector.Scope == utils.VarScopeInstance || selector.Scope == utils.VarScopeWorkflow || selector.Scope == utils.VarScopeNamespace {
			if selector.Scope == "" {
				selector.Scope = utils.VarScopeNamespace
			}

			var item *datastore.RuntimeVariable

			telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("fetching %s variable %s", selector.Scope, selector.Key))
			switch selector.Scope {
			case utils.VarScopeInstance:
				item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, selector.Key)
			case utils.VarScopeWorkflow:
				item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, selector.Key)
			case utils.VarScopeNamespace:
				item, err = tx.DataStore().RuntimeVariables().GetForNamespace(ctx, im.instance.Instance.Namespace, selector.Key)
			default:
				return nil, derrors.NewInternalError(errors.New("invalid scope"))
			}
			if errors.Is(err, datastore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  []byte{},
				})
			} else if err != nil {
				return nil, derrors.NewInternalError(err)
			} else {
				data, err := tx.DataStore().RuntimeVariables().LoadData(ctx, item.ID)
				if err != nil {
					return nil, derrors.NewInternalError(err)
				}
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  data,
				})
			}

			continue
		}

		if selector.Scope == utils.VarScopeFileSystem { //nolint:nestif
			telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("fetching file %s", selector.Key))
			file, err := tx.FileStore().ForNamespace(im.instance.Instance.Namespace).GetFile(ctx, selector.Key)
			if errors.Is(err, filestore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  make([]byte, 0),
				})
			} else if err != nil {
				return nil, err
			} else {
				// TODO: alan, maybe need to enhance the GetData function to also return us some information like mime type, checksum, and size
				if file.Typ == filestore.FileTypeDirectory {
					return nil, model.ErrVarNotFile
				}
				data, err := tx.FileStore().ForFile(file).GetData(ctx)
				if err != nil {
					return nil, err
				}
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  data,
				})
			}

			continue
		}
	}

	return x, nil
}

func (im *instanceMemory) ListenForEvents(ctx context.Context, events []*model.ConsumeEventDefinition, all bool) error {
	err := im.engine.events.deleteInstanceEventListeners(ctx, im)
	if err != nil {
		return err
	}

	err = im.engine.events.listenForEvents(ctx, im, events, all)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) Log(ctx context.Context, level log.Level, a string, x ...interface{}) {
	switch level {
	case log.Info:
		telemetry.LogInstance(ctx, telemetry.LogLevelInfo, fmt.Sprintf(a, x...))
	case log.Debug:
		telemetry.LogInstance(ctx, telemetry.LogLevelDebug, fmt.Sprintf(a, x...))
	case log.Error:
		telemetry.LogInstance(ctx, telemetry.LogLevelError, fmt.Sprintf(a, x...))
	case log.Panic:
		telemetry.LogInstance(ctx, telemetry.LogLevelError, fmt.Sprintf(a, x...))
	}
}

func (im *instanceMemory) AddAttribute(tag, value string) {
	if im.tags == nil {
		im.tags = make(map[string]string)
	}
	im.tags[tag] = value
}

func (im *instanceMemory) Iterator() (int, bool) {
	if im.tags == nil {
		return 0, false
	}
	val, ok := im.tags["loop-index"]
	iterator, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return iterator, ok
}

func (im *instanceMemory) Raise(ctx context.Context, err *derrors.CatchableError) error {
	return im.engine.InstanceRaise(ctx, im, err)
}

func (im *instanceMemory) RetrieveSecret(ctx context.Context, secret string) (string, error) {
	tx, err := im.engine.flow.beginSQLTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	secretData, err := tx.DataStore().Secrets().Get(ctx, im.instance.Instance.Namespace, secret)
	if err != nil {
		return "", err
	}

	return string(secretData.Data), nil
}

//nolint:gocognit
func (im *instanceMemory) SetVariables(ctx context.Context, vars []states.VariableSetter) error {
	tx, err := im.engine.flow.beginSQLTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for idx := range vars {
		v := vars[idx]

		var item *datastore.RuntimeVariable
		d := string(v.Data)

		action := "create"
		if len(d) == 0 || "null" == d {
			action = "delete"
		}

		switch v.Scope {
		case utils.VarScopeInstance:
			telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("setting instance variable %s (%s)", v.Key, action))
			item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, v.Key)
		case utils.VarScopeWorkflow:
			telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("setting workflow variable %s (%s)", v.Key, action))
			item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, v.Key)
		case utils.VarScopeNamespace:
			telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("setting namespace variable %s (%s)", v.Key, action))
			telemetry.LogNamespace(telemetry.LogLevelInfo, im.instance.Instance.Namespace,
				fmt.Sprintf("setting namespace variable %s (%s)", v.Key, action))
			item, err = tx.DataStore().RuntimeVariables().GetForNamespace(ctx, im.instance.Instance.Namespace, v.Key)
		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		if err != nil && !errors.Is(err, datastore.ErrNotFound) {
			return err
		}

		if len(d) == 0 {
			err = tx.DataStore().RuntimeVariables().Delete(ctx, item.ID)
			if err != nil && !errors.Is(err, datastore.ErrNotFound) {
				return err
			}

			continue
		}

		//nolint:nestif
		if !(v.MIMEType == "text/plain; charset=utf-8" || v.MIMEType == "text/plain" || v.MIMEType == "application/octet-stream") && (d == "{}" || d == "[]" || d == "0" || d == `""` || d == "null") {
			if item != nil {
				err = tx.DataStore().RuntimeVariables().Delete(ctx, item.ID)
				if err != nil && !errors.Is(err, datastore.ErrNotFound) {
					return err
				}
			}
		} else {
			newVar := &datastore.RuntimeVariable{
				Name:      v.Key,
				MimeType:  v.MIMEType,
				Data:      v.Data,
				Namespace: im.instance.Instance.Namespace,
			}

			switch v.Scope {
			case utils.VarScopeInstance:
				newVar.InstanceID = im.instance.Instance.ID
			case utils.VarScopeWorkflow:
				newVar.WorkflowPath = im.instance.Instance.WorkflowPath
			}

			_, err = tx.DataStore().RuntimeVariables().Set(ctx, newVar)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) Sleep(ctx context.Context, d time.Duration, x interface{}) error {
	return im.ScheduleRetry(ctx, d, im.logic.GetID(), x)
}

func (im *instanceMemory) GetInstanceData() interface{} {
	return im.data
}

func (im *instanceMemory) GetModel() (*model.Workflow, error) {
	return im.Model()
}

func (im *instanceMemory) GetInstanceID() uuid.UUID {
	return im.instance.Instance.ID
}

func (im *instanceMemory) GetTraceID(ctx context.Context) string {
	if trace.SpanFromContext(ctx).SpanContext().IsValid() {
		return trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	}

	return ""
}

func (im *instanceMemory) PrimeDelayedEvent(event cloudevents.Event) {
	im.eventQueue = append(im.eventQueue, event.ID())
}

func (im *instanceMemory) SetMemory(ctx context.Context, x interface{}) error {
	return im.engine.SetMemory(ctx, im, x)
}

func (im *instanceMemory) Deadline(ctx context.Context) time.Time {
	return time.Now().UTC().Add(states.DefaultShortDeadline)
}

func (im *instanceMemory) LivingChildren(ctx context.Context) []*states.ChildInfo {
	return nil
}

func (im *instanceMemory) ScheduleRetry(ctx context.Context, d time.Duration, stateID string, x interface{}) error {
	data, err := json.Marshal(x)
	if err != nil {
		return err
	}

	t := time.Now().UTC().Add(d)

	err = im.engine.scheduleRetry(im.ID().String(), t, data) //nolint:contextcheck
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) CreateChild(ctx context.Context, args states.CreateChildArgs) (states.Child, error) {
	var ci states.ChildInfo

	if args.Definition.GetType() == model.SubflowFunctionType {
		pi := &enginerefactor.ParentInfo{
			ID:     im.ID(),
			State:  im.logic.GetID(),
			Step:   im.Step(),
			Branch: args.Iterator,
		}

		sfim, err := im.engine.subflowInvoke(ctx, pi, im.instance, args.Definition.(*model.SubflowFunctionDefinition).Workflow, args.Input)
		if err != nil {
			return nil, err
		}

		ci.ID = sfim.ID().String()
		ci.Type = "subflow"
		// ci.Attempts: this is ignored here. Must be handled elsewhere.

		return &subflowHandle{
			im:     sfim,
			info:   ci,
			engine: im.engine,
		}, nil
	}

	switch args.Definition.GetType() { //nolint:exhaustive
	case model.SystemKnativeFunctionType:
	case model.NamespacedKnativeFunctionType:
	case model.ReusableContainerFunctionType:
	default:
		return nil, derrors.NewInternalError(fmt.Errorf("unsupported function type: %v", args.Definition.GetType()))
	}

	uid := uuid.New()

	ar, arReq, err := im.engine.newIsolateRequest(im, im.logic.GetID(), args.Timeout, args.Definition, args.Input, uid, args.Async, args.Files, args.Iterator)
	if err != nil {
		return nil, err
	}

	ci.ID = ar.ActionID
	ci.ServiceName = ar.Container.Service
	ci.Type = "isolate"

	return &knativeHandle{
		im:     im,
		info:   ci,
		engine: im.engine,
		ar:     ar,
		arReq:  arReq,
	}, nil
}

type subflowHandle struct {
	im     *instanceMemory
	info   states.ChildInfo
	engine *engine
}

func (child *subflowHandle) Run(ctx context.Context) {
	go child.engine.start(child.im) //nolint:contextcheck
}

func (child *subflowHandle) Info() states.ChildInfo {
	return child.info
}

func (engine *engine) newIsolateRequest(im *instanceMemory, stateID string, timeout int,
	fn model.FunctionDefinition, inputData []byte,
	uid uuid.UUID, async bool, files []model.FunctionFileDefinition, iterator int,
) (*functionRequest, *enginerefactor.ActionRequest, error) {
	ar := new(functionRequest)
	ar.Timeout = timeout
	ar.ActionID = uid.String()
	ar.CallPath = im.instance.TelemetryInfo.CallPath

	if ar.Timeout == 0 {
		ar.Timeout = 5 * 60 // 5 mins default, knative's default
	}
	arReq := enginerefactor.ActionRequest{
		Async:     async,
		UserInput: inputData,
		Deadline:  time.Now().UTC().Add(time.Duration(timeout) * time.Second), // TODO?
	}

	arCtx := enginerefactor.ActionContext{
		TraceParent: im.instance.TelemetryInfo.TraceParent,
		State:       stateID,
		Branch:      iterator,
		Namespace:   im.Namespace().Name,
		Workflow:    im.instance.Instance.WorkflowPath,
		Instance:    im.ID().String(),
		Action:      uid.String(),
		Path:        im.instance.Instance.WorkflowPath,
		Invoker:     im.instance.Instance.Invoker,
	}
	arReq.ActionContext = arCtx

	ar.Container.Type = fn.GetType()

	switch ar.Container.Type { //nolint:exhaustive
	case model.ReusableContainerFunctionType:
		con := fn.(*model.ReusableFunctionDefinition) //nolint:forcetypeassert
		scale := int32(0)
		ar.Container.Image = con.Image
		ar.Container.Cmd = con.Cmd
		ar.Container.Size = con.Size
		ar.Container.Scale = int(scale)
		ar.Container.ID = con.ID
		ar.Container.Service = engine.ServiceManager.GetServiceURL(arCtx.Namespace, core.ServiceTypeWorkflow, arCtx.Workflow, con.ID)
	case model.NamespacedKnativeFunctionType:
		con := fn.(*model.NamespacedFunctionDefinition) //nolint:forcetypeassert
		ar.Container.ID = con.ID
		ar.Container.Service = engine.ServiceManager.GetServiceURL(arCtx.Namespace, core.ServiceTypeNamespace, con.Path, "")
	case model.SystemKnativeFunctionType:
		con := fn.(*model.SystemFunctionDefinition) //nolint:forcetypeassert
		ar.Container.ID = con.ID
		ar.Container.Service = engine.ServiceManager.GetServiceURL(core.SystemNamespace, core.ServiceTypeSystem, con.Path, "")
	default:
		return nil, nil, fmt.Errorf("unexpected function type: %v", fn)
	}

	// check for duplicate file names
	m := make(map[string]*model.FunctionFileDefinition)
	for i := range files {
		f := &files[i]
		k := f.As
		if k == "" {
			k = f.Key
		}
		if _, exists := m[k]; exists {
			return nil, nil, fmt.Errorf("multiple files with same name: %s", k)
		}
		m[k] = f
	}
	files2 := make([]enginerefactor.FunctionFileDefinition, len(files))
	for i := range files {
		files2[i] = enginerefactor.FunctionFileDefinition{
			Key:         files[i].Key,
			As:          files[i].As,
			Scope:       files[i].Scope,
			Type:        files[i].Type,
			Permissions: files[i].Permissions,
			// Content:    TODO: evaluate if we should inject the content here?
		}
	}
	arReq.Files = files2

	return ar, &arReq, nil
}

type knativeHandle struct {
	im     *instanceMemory
	info   states.ChildInfo
	engine *engine
	ar     *functionRequest
	arReq  *enginerefactor.ActionRequest
}

func (child *knativeHandle) Run(ctx context.Context) {
	go child.engine.doActionRequest(ctx, child.ar, child.arReq) // using a go routine may be unsafe (caused panics) where but why?
}

func (child *knativeHandle) Info() states.ChildInfo {
	return child.info
}

func (engine *engine) doActionRequest(ctx context.Context, ar *functionRequest, arReq *enginerefactor.ActionRequest) {
	if actionTimeout := time.Duration(ar.Timeout) * time.Second; actionTimeout > engine.server.config.GetFunctionsTimeout() {
		telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
			fmt.Sprintf("action timeout '%v' is longer than max allowed duariton '%v'", actionTimeout, engine.server.config.GetFunctionsTimeout()))
	}

	switch ar.Container.Type { //nolint:exhaustive
	case model.DefaultFunctionType:
		fallthrough
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.SystemKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:
		go engine.doKnativeHTTPRequest(ctx, ar, arReq) // go routine causes panic
	default:
		panic(fmt.Errorf("unexpected type: %+v", ar.Container.Type))
	}
}

func (engine *engine) doKnativeHTTPRequest(ctx context.Context,
	ar *functionRequest, arReq *enginerefactor.ActionRequest,
) {
	ctx, span := telemetry.Tracer.Start(ctx, "call-action")
	defer span.End()

	// create a new indepenedent context with traceparent
	traceparent := telemetry.TraceParent(ctx)
	traceparentCtx := telemetry.FromTraceParent(context.Background(), traceparent)

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "image",
			Value: attribute.StringValue(ar.Container.Image),
		},
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(ar.Container.ID),
		},
		attribute.KeyValue{
			Key:   "service",
			Value: attribute.StringValue(ar.Container.Service),
		},
	)

	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"starting function request")
	tr := engine.createTransport()
	addr := ar.Container.Service

	slog.Debug("function request for image", "name", ar.Container.Image, "addr", addr, "image_id", ar.Container.ID)

	rctx, cancel := context.WithDeadline(traceparentCtx, arReq.Deadline)
	defer cancel()

	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		fmt.Sprintf("deadline for request is %s", time.Until(arReq.Deadline)))

	var resp *http.Response

	// potentially dns error for a brand new service
	// we just loop and see if we can recreate the service
	// one minute wait max

	//nolint:intrange

	serviceIgnited := false
	for i := range 300 { // 5 minutes max retry
		reader, err := enginerefactor.EncodeActionRequest(*arReq)
		if err != nil {
			engine.reportError(ctx, &arReq.ActionContext, err)

			return
		}
		req, err := http.NewRequestWithContext(rctx, http.MethodPost, addr, reader)
		if err != nil {
			engine.reportError(ctx, &arReq.ActionContext, err)

			return
		}
		req.Header.Add(DirektivActionIDHeader, ar.ActionID)
		req.Header.Add(DirektivCallPathHeader, ar.CallPath)
		client := http.Client{Transport: otelhttp.NewTransport(tr)}
		telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("attempting service request %d, %s", i, addr))

		err = engine.db.DataStore().HeartBeats().Set(context.Background(), &datastore.HeartBeat{
			Group: "life_services",
			Key:   ar.Container.Service,
		})
		if err != nil {
			engine.reportError(ctx, &arReq.ActionContext, err)

			return
		}
		resp, err = client.Do(req)
		isBadGatewayCode := resp != nil && resp.StatusCode >= 502 && resp.StatusCode <= 504

		if isBadGatewayCode {
			err = fmt.Errorf("bad gateway status code (%d)", resp.StatusCode)
		}
		if err != nil {
			isServiceDown := strings.Contains(err.Error(), "bad gateway status code") ||
				strings.Contains(err.Error(), "no such host") ||
				strings.Contains(err.Error(), "connection refused")

			if isServiceDown {
				telemetry.LogInstanceError(ctx, "service is scaled to zero", err)
			}

			// Try to ignite (once) the service if it was down.
			if isServiceDown && !serviceIgnited {
				igErr := engine.ServiceManager.IgniteService(ar.Container.Service)
				if igErr != nil {
					engine.reportError(ctx, &arReq.ActionContext, igErr)
				} else {
					serviceIgnited = true
					telemetry.LogInstance(ctx, telemetry.LogLevelInfo, "service ignition triggered")
				}
			}

			if ctxErr := rctx.Err(); ctxErr != nil {
				telemetry.LogInstanceError(ctx, "request canceled or deadline exceeded", ctxErr)
				engine.reportError(ctx, &arReq.ActionContext, fmt.Errorf("request timed out or was canceled: %w", ctxErr))

				return
			}

			if i%10 == 0 {
				slog.Debug(fmt.Sprintf("retrying function request for container '%s' (attempt %d)", ar.Container.ID, i), slog.Any("error", err))
			} else {
				slog.Debug("retrying function request", "image", ar.Container.Image, "image_id", ar.Container.ID, "error", err)
			}

			if resp != nil {
				resp.Body.Close()
				resp = nil
			}

			time.Sleep(time.Second)
		} else {
			defer resp.Body.Close()
			slog.Debug("function request successful", "image", ar.Container.Image, "image_id", ar.Container.ID)
			aid := resp.Header.Get(DirektivActionIDHeader)
			if len(aid) == 0 {
				slog.Debug("action ID missing from response", "this", this())
				engine.reportError(ctx, &arReq.ActionContext, fmt.Errorf("missing action ID in response"))

				return
			}
			var respBody enginerefactor.ActionResponse
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&respBody); err != nil {
				slog.Debug("failed to decode response body", "error", err)
				engine.reportError(ctx, &arReq.ActionContext, err)

				return
			}
			payload := &actionResultPayload{
				ActionID:     aid,
				ErrorCode:    respBody.ErrCode,
				ErrorMessage: respBody.ErrMsg,
				Output:       respBody.Output,
			}

			uid, err := uuid.Parse(arReq.Instance)
			if err != nil {
				slog.Debug("failed to parse instance UUID", "error", err)
				engine.reportError(ctx, &arReq.ActionContext, err)

				return
			}

			err = engine.enqueueInstanceMessage(ctx, uid, "action", payload)
			if err != nil {
				slog.Debug("failed to enqueue instance message", "error", err)
				engine.reportError(ctx, &arReq.ActionContext, err)

				return
			}

			break
		}
	}

	if resp.StatusCode != http.StatusOK {
		engine.reportError(ctx, &arReq.ActionContext, fmt.Errorf("action error status: %d", resp.StatusCode))
	}
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"function request done")
}

func (engine *engine) reportError(ctx context.Context, ar *enginerefactor.ActionContext, err error) {
	telemetry.LogInstanceError(ctx, "action failed", err)
}
