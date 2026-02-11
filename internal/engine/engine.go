package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var ErrDataNotFound = fmt.Errorf("data not found")

// LabelWithNotify used to mark an instance as called with a notify-chanel.
const (
	LabelWithNotify   = "WithNotify"
	LabelWithSyncExec = "WithSyncExec"
	LabelInvokerType  = "InvokerType"

	// LabelWithScope used to mark a workflow instance if it is:
	//	1- a main execution or
	//  2- a subflow execution.
	// For main execution, use string "main"
	// For subflow execution, use string uuid that uniquely identifies the subflow.
	LabelWithScope = "WithScope"
)

type Engine struct {
	dataBus  DataBus
	compiler core.Compiler
	js       nats.JetStreamContext
	store    datastore.Store
}

func NewEngine(dataBus DataBus, compiler core.Compiler, js nats.JetStreamContext, store datastore.Store) (*Engine, error) {
	e := &Engine{
		dataBus:  dataBus,
		compiler: compiler,
		js:       js,
		store:    store,
	}
	return e, nil
}

func (e *Engine) Start(lc *lifecycle.Manager) error {
	err := e.dataBus.Start(lc)
	if err != nil {
		return fmt.Errorf("start databus: %w", err)
	}

	err = e.startQueueWorkers(lc)
	if err != nil {
		return fmt.Errorf("start queue workers: %w", err)
	}

	return nil
}

func (e *Engine) StartWorkflow(ctx context.Context, instID uuid.UUID, namespace string, workflowPath string, input string, metadata map[string]string) (*InstanceEvent, <-chan *InstanceEvent, error) {
	flowDetails, err := e.compiler.FetchScript(ctx, namespace, workflowPath, true)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch script: %w", err)
	}

	// fetch all the secrets here
	metadata[core.EngineMappingSecrets] = flowDetails.Secrets
	metadata[core.EngineMappingNamespace] = namespace
	metadata[core.EngineMappingPath] = workflowPath

	notify := make(chan *InstanceEvent, 1)
	st, err := e.startScript(ctx, instID, namespace, flowDetails.Script, flowDetails.Mapping, flowDetails.Config.State, input, notify, metadata)
	if err != nil {
		return nil, nil, err
	}

	return st, notify, nil
}

var (
	notifyMap  = map[string]chan<- *InstanceEvent{}
	notifyLock = &sync.Mutex{}
)

func (e *Engine) startScript(ctx context.Context, instID uuid.UUID, namespace string, script string, mappings string, fn string, input string, notify chan<- *InstanceEvent, metadata map[string]string) (*InstanceEvent, error) {
	if !json.Valid([]byte(input)) {
		return nil, fmt.Errorf("input is not a valid json string: %s", input)
	}

	if metadata == nil {
		metadata = map[string]string{
			LabelWithScope: "main",
		}
	}

	pEv := &InstanceEvent{
		State: StateCodePending,

		EventID:    uuid.New(),
		InstanceID: instID,
		Namespace:  namespace,
		Metadata:   metadata,
		Script:     script,
		Fn:         fn,
		Mappings:   mappings,

		Input:  json.RawMessage(input),
		Output: nil,
		Error:  "",

		CreatedAt: time.Now(),
		StartedAt: time.Time{},
		EndedAt:   time.Time{},
	}
	err := e.dataBus.PublishInstanceHistoryEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push history stream: %w", err)
	}

	if notify != nil {
		notifyLock.Lock()
		notifyMap[pEv.FullID()] = notify
		notifyLock.Unlock()
	}

	if metadata[LabelWithSyncExec] == "true" {
		err = e.execInstance(ctx, pEv)
		if err != nil {
			return nil, fmt.Errorf("exec instance: %w", err)
		}

		return pEv, nil
	}

	err = e.dataBus.PublishInstanceQueueEvent(ctx, pEv)
	if err != nil {
		return nil, fmt.Errorf("push queue stream: %w", err)
	}

	return pEv, nil
}

func (e *Engine) execInstance(ctx context.Context, inst *InstanceEvent) error {
	startEv := inst.Clone()
	startEv.EventID = uuid.New()
	startEv.State = StateCodeRunning
	startEv.StartedAt = time.Now()

	err := e.dataBus.PublishInstanceHistoryEvent(ctx, startEv)
	if err != nil {
		return fmt.Errorf("push history start event, inst: %s: %w", inst.InstanceID, err)
	}

	sc := &runtime.Script{
		InstID:   startEv.InstanceID,
		Text:     startEv.Script,
		Mappings: startEv.Mappings,
		Fn:       startEv.Fn,
		Input:    string(startEv.Input),
		Metadata: startEv.Metadata,
	}

	var onAction runtime.OnActionHook = func(svcID string) error {
		// return e.dataBus.PublishIgniteAction(ctx, config,
		// 	inst.Metadata[core.EngineMappingNamespace], inst.Metadata[core.EngineMappingPath])
		return e.dataBus.PublishIgniteAction(ctx, svcID)
	}
	var onFinish runtime.OnFinishHook = func(output []byte) error {
		endEv := startEv.Clone()
		endEv.EventID = uuid.New()
		endEv.State = StateCodeComplete
		endEv.Output = output
		endEv.EndedAt = time.Now()
		endEv.Fn = ""

		if endEv.Metadata[LabelWithNotify] == "true" {
			notifyLock.Lock()
			notify, ok := notifyMap[endEv.FullID()]
			notifyLock.Unlock()
			if ok {
				notify <- endEv
			}
		}

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}
	var onTransition runtime.OnTransitionHook = func(memory []byte, fn string) error {
		endEv := startEv.Clone()
		endEv.EventID = uuid.New()
		endEv.State = StateCodeRunning
		endEv.Output = memory
		endEv.Fn = fn

		return e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	}
	var onSetVariable runtime.OnSetVariableHook = func(ctx context.Context, scope string, name string, data []byte) error {
		if name == "" {
			return datastore.ErrInvalidRuntimeVariableName
		}

		rv := &datastore.RuntimeVariable{
			Namespace: inst.Namespace,
			Name:      name,
			Data:      data,
			MimeType:  "application/octet-stream",
		}

		switch core.VariableScope(scope) {
		case core.VariableScopeNamespace:
		case core.VariableScopeWorkflow:
			wfPath := inst.Metadata[core.EngineMappingPath]
			if wfPath == "" {
				return fmt.Errorf("missing workflow path in instance metadata for workflow-scoped variable")
			}
			rv.WorkflowPath = wfPath
		case core.VariableScopeInstance:
			rv.InstanceID = inst.InstanceID
		default:
			return fmt.Errorf("invalid variable scope %q", scope)
		}

		if _, err := e.store.RuntimeVariables().Create(ctx, rv); err != nil {
			return err
		}

		// publish simple event
		return e.dataBus.PublishRuntimeVariableSet(ctx, rv.Namespace, rv.Name, string(rv.Data))
	}
	var onGetVariable runtime.OnGetVariableHook = func(ctx context.Context, scope string, name string) ([]byte, error) {
		var (
			v   *datastore.RuntimeVariable
			err error
		)

		switch core.VariableScope(scope) {
		case core.VariableScopeNamespace:
			v, err = e.store.RuntimeVariables().GetForNamespace(ctx, inst.Namespace, name)
		case core.VariableScopeWorkflow:
			wfPath := inst.Metadata[core.EngineMappingPath]
			if wfPath == "" {
				return nil, fmt.Errorf("missing workflow path in instance metadata for workflow-scoped variable")
			}
			v, err = e.store.RuntimeVariables().GetForWorkflow(ctx, inst.Namespace, wfPath, name)
		case core.VariableScopeInstance:
			v, err = e.store.RuntimeVariables().GetForInstance(ctx, inst.InstanceID, name)
		default:
			return nil, fmt.Errorf("invalid variable scope %q", scope)
		}

		if err != nil {
			if errors.Is(err, datastore.ErrNotFound) {
				// not found -> represented as null in the runtime
				return nil, nil
			}

			return nil, err
		}

		data, err := e.store.RuntimeVariables().LoadData(ctx, v.ID)
		if err != nil {
			if errors.Is(err, datastore.ErrNotFound) {
				return nil, nil
			}

			return nil, err
		}

		return data, nil
	}
	var onSubflow runtime.OnSubflowHook = func(ctx context.Context, path string, input []byte) ([]byte, error) {
		_, notify, err := e.StartWorkflow(ctx, inst.InstanceID, inst.Namespace, path, string(input), map[string]string{
			LabelWithNotify:   strconv.FormatBool(true),
			LabelWithSyncExec: strconv.FormatBool(true),
			LabelInvokerType:  inst.Metadata[LabelInvokerType],
			LabelWithScope:    uuid.New().String(),
		})
		if err != nil {
			return nil, err
		}
		st := <-notify
		if st.State != StateCodeComplete {
			return nil, fmt.Errorf("subflow did not complete: %s", st.Error)
		}

		return st.Output, nil
	}

	err = runtime.ExecScript(ctx, sc, onFinish, onTransition, onAction, onSubflow, onSetVariable, onGetVariable)
	if err == nil {
		return nil
	}
	telemetry.LogInstance(ctx, telemetry.LogLevelError, fmt.Sprintf("flow execution failed: %s", err.Error()))

	endEv := startEv.Clone()
	endEv.EventID = uuid.New()
	endEv.State = StateCodeFailed
	endEv.Fn = ""
	endEv.Error = err.Error()
	endEv.EndedAt = time.Now()

	if inst.Metadata[LabelWithNotify] == "true" {
		notifyLock.Lock()
		notify, ok := notifyMap[endEv.FullID()]
		notifyLock.Unlock()
		if ok {
			notify <- endEv
		}
	}
	err = e.dataBus.PublishInstanceHistoryEvent(ctx, endEv)
	if err != nil {
		return fmt.Errorf("push history end event, inst: %s: %w", inst.InstanceID, err)
	}

	return nil
}

func (e *Engine) ListInstanceStatuses(ctx context.Context, limit int, offset int, filters filter.Values) ([]*InstanceEvent, int, error) {
	data, total := e.dataBus.ListInstanceStatuses(ctx, limit, offset, filters)

	return data, total, nil
}

func (e *Engine) GetInstanceStatus(ctx context.Context, namespace string, id uuid.UUID) (*InstanceEvent, error) {
	data, _ := e.dataBus.ListInstanceStatuses(ctx, 0, 0, filter.With(nil,
		filter.FieldEQ("namespace", namespace),
		filter.FieldEQ("instanceID", id.String()),
	))
	if len(data) == 0 {
		return nil, ErrDataNotFound
	}

	return data[0], nil
}

func (e *Engine) GetInstanceHistory(ctx context.Context, namespace string, id uuid.UUID) ([]*InstanceEvent, error) {
	list := e.dataBus.GetInstanceHistory(ctx, namespace, id)
	if len(list) == 0 {
		return nil, ErrDataNotFound
	}

	return list, nil
}

func (e *Engine) DeleteNamespace(ctx context.Context, name string) error {
	return e.dataBus.DeleteNamespace(ctx, name)
}
