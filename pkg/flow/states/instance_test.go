package states

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
)

type testerInstance struct {
	tbuf []string

	instanceID     uuid.UUID
	traceID        string
	instanceData   map[string]interface{}
	instanceMemory interface{}
	instanceModel  *model.Workflow

	t0       time.Time
	delays   time.Duration
	wakedata []byte
}

func newTesterInstance() *testerInstance {
	instance := new(testerInstance)

	instance.tbuf = make([]string, 0)
	instance.instanceData = make(map[string]interface{})

	instance.t0 = time.Now().UTC()
	instance.delays = time.Millisecond * 0

	return instance
}

func (instance *testerInstance) dt() time.Duration {
	return time.Now().UTC().Add(instance.delays).Sub(instance.t0)
}

func (instance *testerInstance) dtCPU() time.Duration {
	return time.Since(instance.t0)
}

func (instance *testerInstance) getTrace() []string {
	return instance.tbuf
}

func (instance *testerInstance) getWakedata() []byte {
	data := instance.wakedata

	instance.wakedata = nil

	return data
}

func (instance *testerInstance) resetTrace() {
	instance.tbuf = make([]string, 0)
}

func (instance *testerInstance) trace() {
	var fn string

	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fn = runtime.FuncForPC(pc).Name()
		idx := strings.LastIndex(fn, ".")
		if idx >= 0 {
			fn = fn[idx+1:]
		}
	}

	instance.tbuf = append(instance.tbuf, fn)
}

func (instance *testerInstance) getTraceExclude(excl ...string) []string {
	x := instance.getTrace()

	var trace []string

	for _, s := range x {
		exclude := false

		for _, m := range excl {
			if m == s {
				exclude = true
			}
		}

		if !exclude {
			trace = append(trace, s)
		}
	}

	return trace
}

func (instance *testerInstance) BroadcastCloudevent(ctx context.Context, evetn *cloudevents.Event, dd int64) error {
	instance.trace()
	return nil
}

func (instance *testerInstance) CreateChild(ctx context.Context, args CreateChildArgs) (Child, error) {
	instance.trace()
	//nolint:nilnil
	return nil, nil
}

func (instance *testerInstance) Deadline(ctx context.Context) time.Time {
	return time.Now().UTC().Add(DefaultShortDeadline)
}

func (instance *testerInstance) GetInstanceData() interface{} {
	return instance.instanceData
}

func (instance *testerInstance) GetInstanceID() uuid.UUID {
	return instance.instanceID
}

func (instance *testerInstance) GetTraceID(ctx context.Context) string {
	return instance.traceID
}

func (instance *testerInstance) GetMemory() interface{} {
	return instance.instanceMemory
}

func (instance *testerInstance) GetModel() (*model.Workflow, error) {
	return instance.instanceModel, nil
}

func (instance *testerInstance) GetVariables(ctx context.Context, vars []VariableSelector) ([]Variable, error) {
	instance.trace()

	variables := make([]Variable, len(vars))

	return variables, nil
}

func (instance *testerInstance) ListenForEvents(ctx context.Context, events []*model.ConsumeEventDefinition, all bool) error {
	instance.trace()
	return nil
}

func (instance *testerInstance) LivingChildren(ctx context.Context) []*ChildInfo {
	return nil
}

func (instance *testerInstance) Log(ctx context.Context, level log.Level, a string, x ...interface{}) {
}

func (instance *testerInstance) AddAttribute(tag, value string) {
}

func (instance *testerInstance) Iterator() (int, bool) {
	return 0, false
}

func (instance *testerInstance) PrimeDelayedEvent(events cloudevents.Event) {
	instance.trace()
}

func (instance *testerInstance) Raise(ctx context.Context, err *derrors.CatchableError) error {
	instance.trace()
	return nil
}

func (instance *testerInstance) RetrieveSecret(ctx context.Context, secret string) (string, error) {
	instance.trace()
	return "", nil
}

func (instance *testerInstance) SetMemory(ctx context.Context, x interface{}) error {
	instance.trace()
	instance.instanceMemory = x
	return nil
}

func (instance *testerInstance) SetVariables(ctx context.Context, vars []VariableSetter) error {
	instance.trace()
	return nil
}

func (instance *testerInstance) Sleep(ctx context.Context, d time.Duration, x interface{}) error {
	instance.trace()

	instance.delays += d

	if instance.wakedata != nil {
		panic("sleep twice")
	}

	instance.wakedata = marshal(x)

	return nil
}

func (instance *testerInstance) StoreData(key string, val interface{}) error {
	instance.trace()
	instance.instanceData[key] = val
	return nil
}

func (instance *testerInstance) UnmarshalMemory(x interface{}) error {
	instance.trace()

	data, err := json.Marshal(instance.instanceMemory)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, x)
	if err != nil {
		return err
	}

	return nil
}

func marshal(x interface{}) []byte {
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return data
}
