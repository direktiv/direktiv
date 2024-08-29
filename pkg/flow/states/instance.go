package states

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
)

type Instance interface { //nolint:interfacebloat
	GetInstanceID() uuid.UUID
	GetInstanceData() interface{}
	GetMemory() interface{}
	UnmarshalMemory(x interface{}) error
	GetModel() (*model.Workflow, error)
	PrimeDelayedEvent(event cloudevents.Event)
	SetMemory(ctx context.Context, x interface{}) error
	StoreData(key string, val interface{}) error
	GetVariables(ctx context.Context, vars []VariableSelector) ([]Variable, error)
	Sleep(ctx context.Context, d time.Duration, x interface{}) error
	Raise(ctx context.Context, err *derrors.CatchableError) error
	Log(ctx context.Context, level log.Level, a string, x ...interface{})
	AddAttribute(tag, value string)
	SetVariables(ctx context.Context, vars []VariableSetter) error
	BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error
	ListenForEvents(ctx context.Context, events []*model.ConsumeEventDefinition, all bool) error
	RetrieveSecret(ctx context.Context, secret string) (string, error)
	CreateChild(ctx context.Context, args CreateChildArgs) (Child, error)
	Iterator() (int, bool)
	Deadline(ctx context.Context) time.Time
	LivingChildren(ctx context.Context) []*ChildInfo
	GetTraceID(ctx context.Context) string
}

type Child interface {
	Run(ctx context.Context)
	Info() ChildInfo
}

type CreateChildArgs struct {
	Definition model.FunctionDefinition
	Input      []byte
	Timeout    int
	Async      bool
	Files      []model.FunctionFileDefinition
	Iterator   int
}

type ChildInfo struct {
	ID          string
	Complete    bool
	Type        string
	Attempts    int
	Results     interface{}
	ServiceName string
}

type VariableSelector struct {
	Scope string
	Key   string
}

type Variable struct {
	Scope string
	Key   string
	Data  []byte
}

type VariableSetter struct {
	Scope    string
	Key      string
	MIMEType string
	Data     []byte
}
