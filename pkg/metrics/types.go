package metrics

import "time"

type NextEnum int

const (
	NextEnd        NextEnum = iota // State has ended
	NextTransition                 // State transitioned
	NextRetry                      // State retried
)

var NextEnums = []NextEnum{
	NextEnd, NextTransition, NextRetry,
}

type InvokerEnum int

const (
	InvokerUnknown InvokerEnum = iota
)

var InvokerEnumLabels = map[InvokerEnum]string{
	InvokerUnknown: "unknown",
}

// InsertRecordArgs ..
type InsertRecordArgs struct {
	Namespace            string
	Workflow             string
	Instance             string
	State                string
	WorkflowMilliSeconds int64
	IsolateMilliSeconds  int64
	ErrorCode            string
	Invoker              string
	Next                 NextEnum
	Transition           string
}

// GetMetricsArgs ..
type GetMetricsArgs struct {
	Namespace string
	Workflow  string
	Since     time.Time
}

// Dataset ..
type Dataset struct {
	TotalInstancesRun         int32 `json:"totalInstancesRun"`
	TotalInstanceMilliSeconds int32 `json:"totalInstanceMilliseconds"`
	SuccessfulExecutions      int32 `json:"successfulExecutions"`
	FailedExecutions          int32 `json:"failedExecutions"`

	ErrorCodes               map[string]int32   `json:"errorCodes"`
	ErrorCodesRepresentation map[string]float32 `json:"errorCodesRepresentation"`

	SampleSize               int32   `json:"sampleSize"`
	MeanInstanceMilliSeconds int32   `json:"avgInstanceMilliseconds"`
	SuccessRate              float32 `json:"successRate"`
	FailureRate              float32 `json:"failureRate"`

	States []StateData `json:"states" toml:"states"`
}

// StateData ..
type StateData struct {
	Name string `json:"name"`

	Invokers               map[string]int32 `json:"invokers"`
	InvokersRepresentation map[string]float32

	TotalExecutions   int32 `json:"totalExecutions"`
	TotalMilliSeconds int32 `json:"totalMilliseconds"`
	TotalSuccesses    int32 `json:"totalSuccesses"`
	TotalFailures     int32 `json:"totalFailures"`

	UnhandledErrors               map[string]int32   `json:"unhandledErrors"`
	UnhandledErrorsRepresentation map[string]float32 `json:"unhandledErrorsRepresentation"`

	TotalRetries int32 `json:"totalRetries"`
	Outcomes     struct {
		EndStates struct {
			Success int32 `json:"success"`
			Failure int32 `json:"failure"`
		} `json:"endStates"`
		Transitions map[string]int32 `json:"transitions"`
	} `json:"outcomes"`

	MeanExecutionsPerInstance   int32   `json:"avgExecutionsPerInstance"`
	MeanMilliSecondsPerInstance int32   `json:"avgSecondsPerInstance"`
	SuccessRate                 float32 `json:"successRate"`
	FailureRate                 float32 `json:"failureRate"`

	MeanRetries  float32 `json:"avgRetries"`
	MeanOutcomes struct {
		EndStates struct {
			Success float32 `json:"success"`
			Failure float32 `json:"failure"`
		} `json:"endStates"`
		Transitions map[string]float32 `json:"transitions"`
	} `json:"avgOutcomes"`

	totalUnhandledErrors int32
}
