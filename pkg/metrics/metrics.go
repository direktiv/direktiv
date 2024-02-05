package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Client ..
type Client struct {
	db *gorm.DB
}

type Metrics struct {
	ID         int       `json:"id"`
	Namespace  string    `json:"namespace"`
	Workflow   string    `json:"workflow"`
	Instance   string    `json:"instance"`
	State      string    `json:"state"`
	Timestamp  time.Time `json:"timestamp"`
	WorkflowMS int       `json:"workflow_ms"`
	IsolateMS  int       `json:"isolate_ms"`
	ErrorCode  string    `json:"error_code"`
	Invoker    string    `json:"invoker"`
	Next       int       `json:"next"`
	Transition string    `json:"transition"`
}

func (r *Metrics) didSucceed() bool {
	if r.ErrorCode == "" {
		// state finished without error
		return true
	}

	if NextEnums[r.Next] == NextTransition {
		// error occurred but was caught
		return true
	}

	// uncaught error
	return false
}

// NewClient ..
func NewClient(db *gorm.DB) *Client {
	return &Client{
		db: db,
	}
}

// InsertRecord inserts a metric record into the database.
func (c *Client) InsertRecord(args *InsertRecordArgs) error {
	wf := strings.Split(args.Workflow, ":")[0]

	res := c.db.WithContext(context.Background()).Exec(
		`
					INSERT INTO metrics(
									namespace,workflow,instance,state,
									workflow_ms,isolate_ms,error_code,
									invoker,next,transition) 
					VALUES(?,?,?,?,?,?,?,?,?,?)`,
		args.Namespace,
		wf,
		args.Instance,
		args.State,

		args.WorkflowMilliSeconds,
		args.IsolateMilliSeconds,
		args.ErrorCode,

		args.Invoker,
		args.Next,
		args.Transition,
	)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected metrics insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

// GetMetrics returns the metrics from the database.
func (c *Client) GetMetrics(args *GetMetricsArgs) (*Dataset, error) {
	ctx := context.Background()

	var metricsList []*Metrics
	res := c.db.WithContext(ctx).Raw(`
					SELECT * FROM services
					WHERE namespace=? AND workflow=? AND timestamp=?`,
		args.Namespace,
		args.Workflow,
		args.Since,
	).
		Find(&metricsList)
	if res.Error != nil {
		return nil, res.Error
	}

	return generateDataset(metricsList)
}

func generateDataset(records []*Metrics) (*Dataset, error) {
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

func sortRecord(m map[string]*StateData, instances map[string]int, v *Metrics) {
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

	if v.Invoker == "" {
		v.Invoker = "unknown"
	}

	if _, ok := s.Invokers[v.Invoker]; !ok {
		s.Invokers[v.Invoker] = 0
	}
	s.Invokers[v.Invoker]++

	if _, ok := s.InvokersRepresentation[v.Invoker]; !ok {
		s.InvokersRepresentation[v.Invoker] = 0
	}

	s.TotalExecutions++
	s.TotalMilliSeconds += int32(v.WorkflowMS)

	handleSuccessRecord(v, s)
	handleFailRecord(v, s)

	m[v.State] = s
}

func handleSuccessRecord(r *Metrics, s *StateData) {
	if !r.didSucceed() {
		return
	}

	if NextEnums[r.Next] == NextEnd {
		s.TotalSuccesses++
		s.Outcomes.EndStates.Success++
	} else {
		if _, ok := s.Outcomes.Transitions[r.Transition]; !ok {
			s.Outcomes.Transitions[r.Transition] = 1
			s.MeanOutcomes.Transitions[r.Transition] = 0
		} else {
			s.Outcomes.Transitions[r.Transition]++
		}
	}
}

func handleFailRecord(r *Metrics, s *StateData) {
	if r.didSucceed() {
		return
	}

	s.Outcomes.EndStates.Failure++
	s.totalUnhandledErrors++

	if _, ok := s.UnhandledErrors[r.ErrorCode]; !ok {
		s.UnhandledErrors[r.ErrorCode] = 0
		s.UnhandledErrorsRepresentation[r.ErrorCode] = 0
	}
	s.UnhandledErrors[r.ErrorCode] = s.UnhandledErrors[r.ErrorCode] + 1

	if NextEnums[r.Next] == NextRetry {
		s.TotalRetries++
	} else {
		s.TotalFailures++
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
