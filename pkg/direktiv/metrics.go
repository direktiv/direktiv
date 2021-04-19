package direktiv

import (
	"context"

	"github.com/vorteil/direktiv/pkg/metrics"

	"github.com/vorteil/direktiv/pkg/ingress"
)

func (is *ingressServer) WorkflowMetrics(ctx context.Context, in *ingress.WorkflowMetricsRequest) (*ingress.WorkflowMetricsResponse, error) {

	resp, err := is.wfServer.engine.metricsClient.GetMetrics(&metrics.GetMetricsArgs{
		Namespace: *in.Namespace,
		Workflow:  *in.Workflow,
		Since:     in.SinceTimestamp.AsTime(),
	})
	if err != nil {
		return nil, err
	}

	out := new(ingress.WorkflowMetricsResponse)
	out.TotalInstancesRun = &resp.TotalInstancesRun
	out.TotalInstanceMilliseconds = &resp.TotalInstanceMilliSeconds
	out.SuccessfulExecutions = &resp.SuccessfulExecutions
	out.FailedExecutions = &resp.FailedExecutions
	out.SampleSize = &resp.TotalInstancesRun
	out.MeanInstanceMilliseconds = &resp.MeanInstanceMilliSeconds

	out.ErrorCodes = resp.ErrorCodes
	out.ErrorCodesRepresentation = resp.ErrorCodesRepresentation

	var sr, fr float32
	sr = float32(resp.SuccessRate)
	fr = float32(resp.FailureRate)

	out.SuccessRate = &sr
	out.FailureRate = &fr

	states := make([]*ingress.State, 0)
	for _, s := range resp.States {

		thisState := s

		is := new(ingress.State)
		x := thisState.Name
		is.Name = &x

		is.Invokers = thisState.Invokers
		is.InvokersRepresentation = thisState.InvokersRepresentation

		is.TotalExecutions = &thisState.TotalExecutions
		is.TotalMilliseconds = &thisState.TotalMilliSeconds
		is.TotalSuccesses = &thisState.TotalSuccesses
		is.TotalFailures = &thisState.TotalFailures
		is.TotalRetries = &thisState.TotalRetries
		is.Outcomes = &ingress.Outcomes{
			Success:     &thisState.Outcomes.EndStates.Success,
			Failure:     &thisState.Outcomes.EndStates.Failure,
			Transitions: s.Outcomes.Transitions,
		}

		var fr, sr float32
		sr = float32(thisState.MeanOutcomes.EndStates.Success)
		fr = float32(thisState.MeanOutcomes.EndStates.Failure)

		is.MeanOutcomes = &ingress.MeanOutcomes{
			Success:     &sr,
			Failure:     &fr,
			Transitions: s.MeanOutcomes.Transitions,
		}
		is.MeanExecutionsPerInstance = &thisState.MeanExecutionsPerInstance
		is.MeanMillisecondsPerInstance = &thisState.MeanMilliSecondsPerInstance

		sr2 := float32(thisState.SuccessRate)
		fr2 := float32(thisState.FailureRate)
		ar := float32(thisState.MeanRetries)

		is.SuccessRate = &sr2
		is.FailureRate = &fr2
		is.MeanRetries = &ar

		is.UnhandledErrors = thisState.UnhandledErrors
		is.UnhandledErrorsRepresentation = thisState.UnhandledErrorsRepresentation

		states = append(states, is)
	}

	out.States = states

	return out, nil
}
