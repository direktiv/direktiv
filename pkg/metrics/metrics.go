package metrics

import (
	"context"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/metrics/ent/metrics"
	"github.com/direktiv/direktiv/pkg/util"

	"github.com/direktiv/direktiv/pkg/metrics/ent"
)

// Client ..
type Client struct {
	db *ent.Client
}

// NewClient ..
func NewClient() (*Client, error) {

	db, err := ent.Open("postgres", os.Getenv(util.DBConn))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Run the auto migration tool.
	if err := db.Schema.Create(ctx); err != nil {
		return nil, err
	}

	out := new(Client)
	out.db = db

	return out, nil
}

// InsertRecord inserts a metric record into the database
func (c *Client) InsertRecord(args *InsertRecordArgs) error {

	r := c.db.Metrics.Create()
	r = r.SetNamespace(args.Namespace)
	r = r.SetWorkflow(args.Workflow)
	r = r.SetRevision(args.Revision)
	r = r.SetInstance(args.Instance)
	r = r.SetState(args.State)
	r = r.SetTimestamp(time.Now())
	r = r.SetWorkflowMs(args.WorkflowMilliSeconds)
	r = r.SetIsolateMs(args.IsolateMilliSeconds)
	r = r.SetErrorCode(args.ErrorCode)
	r = r.SetInvoker(args.Invoker)
	r = r.SetNext(int8(args.Next))
	r = r.SetTransition(args.Transition)

	_, err := r.Save(context.Background())
	return err
}

// GetMetrics returns the metrics from the database
func (c *Client) GetMetrics(args *GetMetricsArgs) (*Dataset, error) {

	ctx := context.Background()

	records, err := c.db.Metrics.Query().Where(
		metrics.And(
			metrics.NamespaceEQ(args.Namespace),
			metrics.WorkflowEQ(args.Workflow),
			metrics.RevisionEQ(args.Revision),
			metrics.TimestampGT(args.Since),
		),
	).All(ctx)
	if err != nil {
		return nil, err
	}

	return generateDataset(records)
}

func generateDataset(records []*ent.Metrics) (*Dataset, error) {

	out := new(Dataset)

	// range through all records and sort by state
	instances := make(map[string]int)
	states := make(map[string]*StateData)

	for _, v := range records {
		sortRecord(states, instances, v)
	}

	out.SampleSize = out.TotalInstancesRun
	out.TotalInstancesRun = int32(len(instances))
	out.States = make([]StateData, 0)

	// range over states, with total numbers now finalised
	// and perform further calculations
	var totalErrors int32
	allErrors := make(map[string]int32)

	finaliseStateRecordValues(&finaliseStateRecordValuesArgs{
		states:      states,
		out:         out,
		totalErrors: &totalErrors,
		allErrors:   allErrors,
		instances:   instances,
	})

	out.SuccessRate = float32(out.SuccessfulExecutions) / float32(out.TotalInstancesRun)
	out.FailureRate = float32(out.FailedExecutions) / float32(out.TotalInstancesRun)

	out.ErrorCodes = allErrors
	out.ErrorCodesRepresentation = make(map[string]float32)
	for k, v := range allErrors {
		out.ErrorCodesRepresentation[k] = float32(v) / float32(totalErrors)
	}

	return out, nil
}

func sortRecord(m map[string]*StateData, instances map[string]int, v *ent.Metrics) {

	if _, ok := m[v.State]; !ok {
		m[v.State] = &StateData{
			Name: v.State,
		}
	}
	s := m[v.State]

	if s.UnhandledErrors == nil {
		s.UnhandledErrors = make(map[string]int32)
	}
	if s.UnhandledErrorsRepresentation == nil {
		s.UnhandledErrorsRepresentation = make(map[string]float32)
	}
	if s.Outcomes.Transitions == nil {
		s.Outcomes.Transitions = make(map[string]int32)
	}
	if s.MeanOutcomes.Transitions == nil {
		s.MeanOutcomes.Transitions = make(map[string]float32)
	}
	if s.Invokers == nil {
		s.Invokers = make(map[string]int32)
	}
	if s.InvokersRepresentation == nil {
		s.InvokersRepresentation = make(map[string]float32)
	}

	if _, ok := instances[v.Instance]; !ok {
		instances[v.Instance] = 1
	} else {
		x := instances[v.Instance]
		instances[v.Instance] = x + 1
	}

	r := record{
		r: v,
	}

	if r.r.Invoker == "" {
		r.r.Invoker = "unknown"
	}

	if _, ok := s.Invokers[r.r.Invoker]; !ok {
		s.Invokers[r.r.Invoker] = 0
	}
	s.Invokers[r.r.Invoker] += 1

	if _, ok := s.InvokersRepresentation[r.r.Invoker]; !ok {
		s.InvokersRepresentation[r.r.Invoker] = 0
	}

	s.TotalExecutions += 1
	s.TotalMilliSeconds += int32(v.WorkflowMs)

	handleSuccessRecord(&r, s)
	handleFailRecord(&r, s)

	m[v.State] = s

}

func handleSuccessRecord(r *record, s *StateData) {
	if !r.didSucceed() {
		return
	}

	if NextEnums[r.r.Next] == NextEnd {
		s.TotalSuccesses += 1
		s.Outcomes.EndStates.Success += 1
	} else {
		if _, ok := s.Outcomes.Transitions[r.r.Transition]; !ok {
			s.Outcomes.Transitions[r.r.Transition] = 1
			s.MeanOutcomes.Transitions[r.r.Transition] = 0
		} else {
			s.Outcomes.Transitions[r.r.Transition] += 1
		}
	}
}

func handleFailRecord(r *record, s *StateData) {

	if r.didSucceed() {
		return
	}

	s.Outcomes.EndStates.Failure += 1
	s.totalUnhandledErrors += 1

	if _, ok := s.UnhandledErrors[r.r.ErrorCode]; !ok {
		s.UnhandledErrors[r.r.ErrorCode] = 0
		s.UnhandledErrorsRepresentation[r.r.ErrorCode] = 0
	}
	s.UnhandledErrors[r.r.ErrorCode] = s.UnhandledErrors[r.r.ErrorCode] + 1

	if NextEnums[r.r.Next] == NextRetry {
		s.TotalRetries += 1
	} else {
		s.TotalFailures += 1
	}

}

type finaliseStateRecordValuesArgs struct {
	states      map[string]*StateData
	out         *Dataset
	totalErrors *int32
	allErrors   map[string]int32
	instances   map[string]int
}

func finaliseStateRecordValues(args *finaliseStateRecordValuesArgs) {

	for k, s := range args.states {

		thisState := s

		args.out.SuccessfulExecutions += thisState.TotalSuccesses
		args.out.FailedExecutions += thisState.TotalFailures
		args.out.TotalInstanceMilliSeconds += thisState.TotalMilliSeconds

		thisState.MeanExecutionsPerInstance = (thisState.TotalExecutions - thisState.TotalRetries) / int32(len(args.instances))
		thisState.MeanMilliSecondsPerInstance = thisState.TotalMilliSeconds / int32(len(args.instances))
		thisState.SuccessRate = float32(thisState.TotalSuccesses) / float32(thisState.TotalExecutions)
		thisState.FailureRate = float32(thisState.TotalFailures) / float32(thisState.TotalExecutions)

		thisState.MeanRetries = float32(thisState.TotalRetries) / float32(thisState.TotalExecutions)
		thisState.MeanOutcomes.EndStates.Success = float32(thisState.Outcomes.EndStates.Success) / float32(thisState.TotalExecutions)
		thisState.MeanOutcomes.EndStates.Failure = float32(thisState.Outcomes.EndStates.Failure) / float32(thisState.TotalExecutions)

		for k, t := range thisState.Outcomes.Transitions {
			thisState.MeanOutcomes.Transitions[k] = float32(t) / float32(thisState.TotalExecutions)
		}
		for k, v := range thisState.UnhandledErrors {
			*args.totalErrors += v
			if _, ok := args.allErrors[k]; !ok {
				args.allErrors[k] = v
			} else {
				args.allErrors[k] += v
			}
			thisState.UnhandledErrorsRepresentation[k] = float32(v) / float32(thisState.totalUnhandledErrors)
		}
		for k, v := range thisState.Invokers {
			thisState.InvokersRepresentation[k] = float32(v) / float32(thisState.TotalExecutions)
		}

		args.states[k] = thisState
		args.out.States = append(args.out.States, *thisState)
	}

}
