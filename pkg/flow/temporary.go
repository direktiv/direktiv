package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
	"github.com/direktiv/direktiv/pkg/service"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/google/uuid"
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

			switch selector.Scope {
			case utils.VarScopeInstance:
				telemetry.LogInstanceInfo(ctx, fmt.Sprintf("fetching instance variable %s", selector.Key))
				item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, selector.Key)
			case utils.VarScopeWorkflow:
				telemetry.LogInstanceInfo(ctx, fmt.Sprintf("fetching workflow variable %s", selector.Key))
				item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, selector.Key)
			case utils.VarScopeNamespace:
				telemetry.LogInstanceInfo(ctx, fmt.Sprintf("fetching namespace variable %s", selector.Key))
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
			telemetry.LogInstanceInfo(ctx, fmt.Sprintf("fetching file %s", selector.Key))
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
	// ctx = tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
	// 	Namespace:    im.Namespace().Name,
	// 	InstanceID:   im.GetInstanceID().String(),
	// 	Invoker:      im.instance.Instance.Invoker,
	// 	Callpath:     tracing.CreateCallpath(im.instance),
	// 	WorkflowPath: im.instance.Instance.WorkflowPath,
	// 	Status:       core.LogUnknownStatus,
	// }, im.GetState())
	// ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrack(im.instance))
	switch level {
	case log.Info:
		telemetry.LogInstanceInfo(ctx, fmt.Sprintf(a, x...))
	case log.Debug:
		telemetry.LogInstanceDebug(ctx, fmt.Sprintf(a, x...))
	case log.Error:
		telemetry.LogInstanceError(ctx, fmt.Sprintf(a, x...), fmt.Errorf(a, x...))
	case log.Panic:
		telemetry.LogInstanceError(ctx, fmt.Sprintf(a, x...), fmt.Errorf(a, x...))
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
			telemetry.LogInstanceInfo(ctx, fmt.Sprintf("setting instance variable %s (%s)", v.Key, action))
			item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, v.Key)
		case utils.VarScopeWorkflow:
			telemetry.LogInstanceInfo(ctx, fmt.Sprintf("setting workflow variable %s (%s)", v.Key, action))
			item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, v.Key)
		case utils.VarScopeNamespace:
			telemetry.LogInstanceInfo(ctx, fmt.Sprintf("setting namespace variable %s (%s)", v.Key, action))
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
	return trace.SpanFromContext(ctx).SpanContext().TraceID().String()
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
	if ar.Timeout == 0 {
		ar.Timeout = 5 * 60 // 5 mins default, knative's default
	}
	arReq := enginerefactor.ActionRequest{
		Async:     async,
		UserInput: inputData,
		Deadline:  time.Now().UTC().Add(time.Duration(timeout) * time.Second), // TODO?
	}
	callpath := ""
	if len(im.instance.DescentInfo.Descent) == 0 {
		callpath = im.GetInstanceID().String()
	}
	for _, v := range im.instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	arCtx := enginerefactor.ActionContext{
		TraceParent: im.instance.TelemetryInfo.TraceParent,
		State:       stateID,
		Branch:      iterator,
		Namespace:   im.Namespace().Name,
		Workflow:    im.instance.Instance.WorkflowPath,
		Instance:    im.ID().String(),
		Callpath:    callpath,
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
		ar.Container.Service = service.GetServiceURL(arCtx.Namespace, core.ServiceTypeWorkflow, arCtx.Workflow, con.ID)
	case model.NamespacedKnativeFunctionType:
		con := fn.(*model.NamespacedFunctionDefinition) //nolint:forcetypeassert
		ar.Container.ID = con.ID
		ar.Container.Service = service.GetServiceURL(arCtx.Namespace, core.ServiceTypeNamespace, con.Path, "")
	case model.SystemKnativeFunctionType:
		con := fn.(*model.SystemFunctionDefinition) //nolint:forcetypeassert
		ar.Container.ID = con.ID
		ar.Container.Service = service.GetServiceURL(core.SystemNamespace, core.ServiceTypeSystem, con.Path, "")
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
	// Log warning if timeout exceeds max allowed timeout.
	// ctx = tracing.AddInstanceAttr(ctx, tracing.InstanceAttributes{
	// 	Namespace:    arReq.Namespace,
	// 	InstanceID:   arReq.Instance,
	// 	Callpath:     arReq.Callpath,
	// 	WorkflowPath: arReq.Workflow,
	// 	Status:       core.LogUnknownStatus,
	// })

	// ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrackViaCallpath(arReq.Callpath))

	if actionTimeout := time.Duration(ar.Timeout) * time.Second; actionTimeout > engine.server.config.GetFunctionsTimeout() {
		telemetry.LogInstanceWarn(ctx,
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
	ctx, spanEnd, err := tracing.NewSpan(ctx, "executing knative request to action")
	if err != nil {
		slog.Debug("failed in doKnativeHTTPRequest", "error", err)
	}
	defer spanEnd()
	telemetry.LogInstanceDebug(ctx, "starting function request")
	tr := engine.createTransport()
	addr := ar.Container.Service

	slog.Debug("function request for image", "name", ar.Container.Image, "addr", addr, "image_id", ar.Container.ID)

	rctx, cancel := context.WithDeadline(context.Background(), arReq.Deadline)
	defer cancel()

	telemetry.LogInstanceDebug(ctx, fmt.Sprintf("deadline for request is %s", time.Until(arReq.Deadline)))

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

	client := &http.Client{
		Transport: tr,
	}

	var resp *http.Response

	// potentially dns error for a brand new service
	// we just loop and see if we can recreate the service
	// one minute wait max

	//nolint:intrange
	for i := 0; i < 300; i++ { // 5 minutes max retry
		telemetry.LogInstanceDebug(ctx, fmt.Sprintf("attempting function request %d, %s", i, addr))

		resp, err = client.Do(req)
		if err != nil {
			if ctxErr := rctx.Err(); ctxErr != nil {
				telemetry.LogInstanceError(ctx, "request canceled or deadline exceeded", ctxErr)
				engine.reportError(ctx, &arReq.ActionContext, fmt.Errorf("request timed out or was canceled: %w", ctxErr))

				return
			}

			if i%10 == 0 {
				slog.Debug(fmt.Sprintf("retrying function request for container '%s' (attempt %d)", ar.Container.ID, i), err)
			} else {
				slog.Debug("retrying function request", "image", ar.Container.Image, "image_id", ar.Container.ID, "error", err)
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

	if err != nil {
		err := fmt.Errorf("failed creating function with image %s name %s with error: %w", ar.Container.Image, ar.Container.ID, err)
		engine.reportError(ctx, &arReq.ActionContext, err)

		return
	}

	if resp.StatusCode != http.StatusOK {
		engine.reportError(ctx, &arReq.ActionContext, fmt.Errorf("action error status: %d", resp.StatusCode))
	}
	telemetry.LogInstanceDebug(ctx, "function request done")
}

func (engine *engine) reportError(ctx context.Context, ar *enginerefactor.ActionContext, err error) {
	// ctx = tracing.AddNamespace(ctx, ar.Namespace)
	// tracing.AddInstanceAttr(ctx, tracing.InstanceAttributes{
	// 	Namespace:    ar.Namespace,
	// 	InstanceID:   ar.Instance,
	// 	Callpath:     ar.Callpath,
	// 	WorkflowPath: ar.Workflow,
	// 	Status:       core.LogUnknownStatus,
	// })
	// ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrackViaCallpath(ar.Callpath))
	// slog.ErrorContext(
	// 	ctx,
	// 	"action failed",
	// 	"error",
	// 	err,
	// )

	telemetry.LogInstanceError(ctx, "action failed", err)
}
