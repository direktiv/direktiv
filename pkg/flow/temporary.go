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
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
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
		if selector.Scope == "" || selector.Scope == util.VarScopeInstance || selector.Scope == util.VarScopeWorkflow || selector.Scope == util.VarScopeNamespace {
			if selector.Scope == "" {
				selector.Scope = util.VarScopeNamespace
			}

			var item *datastore.RuntimeVariable

			switch selector.Scope {
			case util.VarScopeInstance:
				item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, selector.Key)
			case util.VarScopeWorkflow:
				item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, selector.Key)
			case util.VarScopeNamespace:
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

		if selector.Scope == util.VarScopeFileSystem { //nolint:nestif
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
	ctx = im.WithTags(ctx)
	slog.With(enginerefactor.GetSlogAttributesWithStatus(ctx, core.LogRunningStatus)...)
	switch level {
	case log.Info:
		slog.Info(fmt.Sprintf(a, x...))
	case log.Debug:
		slog.Debug(fmt.Sprintf(a, x...))
	case log.Error:
		slog.Error(fmt.Sprintf(a, x...))
	case log.Panic:
		slog.Error(fmt.Sprintf("Panic: "+a, x...))
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

		switch v.Scope {
		case util.VarScopeInstance:
			item, err = tx.DataStore().RuntimeVariables().GetForInstance(ctx, im.instance.Instance.ID, v.Key)
		case util.VarScopeWorkflow:
			item, err = tx.DataStore().RuntimeVariables().GetForWorkflow(ctx, im.instance.Instance.Namespace, im.instance.Instance.WorkflowPath, v.Key)
		case util.VarScopeNamespace:
			item, err = tx.DataStore().RuntimeVariables().GetForNamespace(ctx, im.instance.Instance.Namespace, v.Key)
		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		if err != nil && !errors.Is(err, datastore.ErrNotFound) {
			return err
		}

		d := string(v.Data)

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
			case util.VarScopeInstance:
				newVar.InstanceID = im.instance.Instance.ID
			case util.VarScopeWorkflow:
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

	arCtx := enginerefactor.ActionContext{
		Trace:     im.instance.TelemetryInfo.TraceID,
		Span:      im.instance.TelemetryInfo.SpanID,
		State:     stateID,
		Branch:    iterator,
		Namespace: im.Namespace().Name,
		Workflow:  im.instance.Instance.WorkflowPath,
		Instance:  im.ID().String(),
		Callpath:  im.instance.TelemetryInfo.CallPath,
		Action:    uid.String(),
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
	go child.engine.doActionRequest(ctx, child.ar, child.arReq)
}

func (child *knativeHandle) Info() states.ChildInfo {
	return child.info
}

func (engine *engine) doActionRequest(ctx context.Context, ar *functionRequest, arReq *enginerefactor.ActionRequest) {
	// Log warning if timeout exceeds max allowed timeout.
	if actionTimeout := time.Duration(ar.Timeout) * time.Second; actionTimeout > engine.server.config.GetFunctionsTimeout() {
		_, err := engine.internal.ActionLog(context.Background(), &grpc.ActionLogRequest{ //nolint:contextcheck
			InstanceId: arReq.ActionContext.Instance, Msg: []string{fmt.Sprintf("Warning: Action timeout '%v' is longer than max allowed duariton '%v'", actionTimeout, engine.server.config.GetFunctionsTimeout())},
		})
		if err != nil {
			slog.Error("failed to write action log", "error", err)
		}
	}

	switch ar.Container.Type { //nolint:exhaustive
	case model.DefaultFunctionType:
		fallthrough
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.SystemKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:
		go engine.doKnativeHTTPRequest(ctx, ar, arReq)
	default:
		panic(fmt.Errorf("unexpected type: %+v", ar.Container.Type))
	}
}

func (engine *engine) doKnativeHTTPRequest(ctx context.Context,
	ar *functionRequest, arReq *enginerefactor.ActionRequest,
) {
	var err error

	tr := engine.createTransport()

	addr := ar.Container.Service

	slog.Debug("function request for image", "name", ar.Container.Image, "addr", addr, "image_id", ar.Container.ID)

	rctx, cancel := context.WithDeadline(context.Background(), arReq.Deadline)
	defer cancel()

	slog.Debug("deadline for request", "deadline", time.Until(arReq.Deadline))
	reader, err := enginerefactor.EncodeActionRequest(*arReq)
	if err != nil {
		engine.reportError(&arReq.ActionContext, err) //nolint:contextcheck

		return
	}
	req, err := http.NewRequestWithContext(rctx, http.MethodPost, addr, reader)
	if err != nil {
		engine.reportError(&arReq.ActionContext, err) //nolint:contextcheck

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
	cleanup := util.TraceHTTPRequest(ctx, req)
	defer cleanup()

	//nolint:intrange
	for i := 0; i < 300; i++ { // 5 minutes retries.
		slog.Debug("functions request", "i", i, "addr", addr)
		resp, err = client.Do(req)
		if err != nil {
			if ctxErr := rctx.Err(); ctxErr != nil {
				slog.Debug("context error in knative call", "error", rctx.Err())
				return
			}
			slog.Debug("function request", "image", ar.Container.Image, "image_id", ar.Container.ID, "error", err)

			time.Sleep(time.Second)
		} else {
			defer resp.Body.Close()
			slog.Debug("successfully created function", "image", ar.Container.Image, "image_id", ar.Container.ID)
			aid := resp.Header.Get(DirektivActionIDHeader)
			if len(aid) == 0 {
				slog.Debug("action id was empty", "this", this())

				return
			}
			var respBody enginerefactor.ActionResponse
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&respBody); err != nil {
				slog.Debug("failed to decode the response body", "error", err)

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
				slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

				return
			}

			err = engine.enqueueInstanceMessage(ctx, uid, "action", payload)
			if err != nil {
				slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

				return
			}

			break
		}
	}

	if err != nil {
		err := fmt.Errorf("failed creating function with image %s name %s with error: %w", ar.Container.Image, ar.Container.ID, err)
		engine.reportError(&arReq.ActionContext, err) //nolint:contextcheck

		return
	}

	if resp.StatusCode != http.StatusOK {
		engine.reportError(&arReq.ActionContext, fmt.Errorf("action error status: %d", resp.StatusCode)) //nolint:contextcheck
	}

	slog.Debug("function request done")
}

func (engine *engine) reportError(ar *enginerefactor.ActionContext, err error) {
	ec := ""
	em := err.Error()
	step := int32(ar.Step)
	r := &grpc.ReportActionResultsRequest{
		InstanceId:   ar.Instance,
		Step:         step,
		ActionId:     ar.Action,
		ErrorCode:    ec,
		ErrorMessage: em,
		Iterator:     int32(ar.Branch),
	}

	_, err = engine.internal.ReportActionResults(context.Background(), r)
	if err != nil {
		slog.Error("failed to respond to flow", "error", err)
	}
}
