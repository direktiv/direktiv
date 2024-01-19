package flow

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/metrics"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsWf = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "workflows",
			Help:      "Total number of workflows.",
		},
		[]string{"direktiv_namespace", "direktiv_tenant"},
	)

	metricsWfUpdated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "updated_total",
			Help:      "Total number of workflows updated.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsCloudEventsReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "cloudevents_received",
			Help:      "Total number of cloudevents received.",
		},
		[]string{"direktiv_namespace", "ce_type", "ce_source", "direktiv_tenant"},
	)

	metricsCloudEventsCaptured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "cloudevents_captured",
			Help:      "Total number of cloudevents captured.",
		},
		[]string{"direktiv_namespace", "ce_type", "ce_source", "direktiv_tenant"},
	)

	metricsWfInvoked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "invoked_total",
			Help:      "Total number of workflows invoked.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "success_total",
			Help:      "Total number of workflows successfully finished.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfFail = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "failed_total",
			Help:      "Total number of workflows failed.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfPending = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "pending_total",
			Help:      "Total number of workflows pending.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "total_milliseconds",
			Help:      "Total time workflow has been actively executing.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfStateDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "states",
			Name:      "milliseconds",
			Help:      "Average time each state spends in execution.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "state", "direktiv_tenant"},
	)

	metricsWfOutcome = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "outcomes",
			Help:      "Results of each workflow instance.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant", "direktiv_instance_status", "direktiv_errcode"},
	)
)

func reportStateEnd(namespace, workflow, state string, t time.Time) {
	ms := time.Since(t).Milliseconds()
	metricsWfStateDuration.WithLabelValues(namespace, GetInodePath(workflow), state, namespace).Observe(float64(ms))
}

func setupPrometheusEndpoint() error {
	prometheus.MustRegister(metricsWfInvoked)
	prometheus.MustRegister(metricsWfSuccess)
	prometheus.MustRegister(metricsWfFail)
	prometheus.MustRegister(metricsWfDuration)
	prometheus.MustRegister(metricsWfStateDuration)
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	prometheus.MustRegister(metricsWf)
	prometheus.MustRegister(metricsWfUpdated)
	prometheus.MustRegister(metricsWfPending)
	prometheus.MustRegister(metricsWfOutcome)
	prometheus.MustRegister(metricsCloudEventsReceived)
	prometheus.MustRegister(metricsCloudEventsCaptured)

	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

// WorkflowMetrics - Gets the Workflow metrics of a given Workflow Revision Ref
// if ref is not set in the request, it will be automatically be set to latest.
func (flow *flow) WorkflowMetrics(ctx context.Context, req *grpc.WorkflowMetricsRequest) (*grpc.WorkflowMetricsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	var rev *filestore.Revision

	rev, err = tx.FileStore().ForFile(file).GetRevision(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := flow.metrics.GetMetrics(&metrics.GetMetricsArgs{
		Namespace: ns.Name,
		Workflow:  file.Path,
		Revision:  rev.ID.String(),
		Since:     req.SinceTimestamp.AsTime(),
	})
	if err != nil {
		return nil, err
	}

	out := new(grpc.WorkflowMetricsResponse)
	out.TotalInstancesRun = resp.TotalInstancesRun
	out.TotalInstanceMilliseconds = resp.TotalInstanceMilliSeconds
	out.SuccessfulExecutions = resp.SuccessfulExecutions
	out.FailedExecutions = resp.FailedExecutions
	out.SampleSize = resp.TotalInstancesRun
	out.MeanInstanceMilliseconds = resp.MeanInstanceMilliSeconds

	out.ErrorCodes = resp.ErrorCodes
	out.ErrorCodesRepresentation = resp.ErrorCodesRepresentation

	var sr, fr float32
	sr = resp.SuccessRate
	fr = resp.FailureRate

	out.SuccessRate = sr
	out.FailureRate = fr

	states := make([]*grpc.State, 0)
	for _, s := range resp.States {
		thisState := s

		is := new(grpc.State)
		x := thisState.Name
		is.Name = x

		is.Invokers = thisState.Invokers
		is.InvokersRepresentation = thisState.InvokersRepresentation

		is.TotalExecutions = thisState.TotalExecutions
		is.TotalMilliseconds = thisState.TotalMilliSeconds
		is.TotalSuccesses = thisState.TotalSuccesses
		is.TotalFailures = thisState.TotalFailures
		is.TotalRetries = thisState.TotalRetries
		is.Outcomes = &grpc.Outcomes{
			Success:     thisState.Outcomes.EndStates.Success,
			Failure:     thisState.Outcomes.EndStates.Failure,
			Transitions: s.Outcomes.Transitions,
		}

		var fr, sr float32
		sr = thisState.MeanOutcomes.EndStates.Success
		fr = thisState.MeanOutcomes.EndStates.Failure

		is.MeanOutcomes = &grpc.MeanOutcomes{
			Success:     sr,
			Failure:     fr,
			Transitions: s.MeanOutcomes.Transitions,
		}
		is.MeanExecutionsPerInstance = thisState.MeanExecutionsPerInstance
		is.MeanMillisecondsPerInstance = thisState.MeanMilliSecondsPerInstance

		sr2 := thisState.SuccessRate
		fr2 := thisState.FailureRate
		ar := thisState.MeanRetries

		is.SuccessRate = sr2
		is.FailureRate = fr2
		is.MeanRetries = ar

		is.UnhandledErrors = thisState.UnhandledErrors
		is.UnhandledErrorsRepresentation = thisState.UnhandledErrorsRepresentation

		states = append(states, is)
	}

	out.States = states

	return out, nil
}

func (engine *engine) metricsCompleteState(ctx context.Context, im *instanceMemory, nextState, errCode string, retrying bool) {
	workflow := GetInodePath(im.instance.Instance.WorkflowPath)

	reportStateEnd(im.instance.TelemetryInfo.NamespaceName, workflow, im.logic.GetID(), im.instance.RuntimeInfo.StateBeginTime)

	if im.Step() == 0 {
		return
	}

	args := new(metrics.InsertRecordArgs)

	args.Namespace = im.instance.TelemetryInfo.NamespaceName
	args.Workflow = workflow
	args.Instance = im.instance.Instance.ID.String()

	caller := engine.InstanceCaller(ctx, im)
	if caller != nil {
		args.Invoker = caller.ID.String()
	}

	flow := im.Flow()
	args.State = flow[len(flow)-1]

	d := time.Since(im.StateBeginTime())
	args.WorkflowMilliSeconds = d.Milliseconds()

	args.ErrorCode = errCode
	args.Transition = nextState
	args.Next = metrics.NextTransition
	if nextState == "" {
		args.Next = metrics.NextEnd
	} else if retrying {
		args.Next = metrics.NextRetry
	}

	if im.Step() == 1 {
		args.Invoker = "start"
	}

	err := engine.metrics.InsertRecord(args)
	if err != nil {
		engine.sugar.Error(err)
	}
}

func (engine *engine) metricsCompleteInstance(ctx context.Context, im *instanceMemory) {
	t := im.StateBeginTime()
	namespace := im.instance.TelemetryInfo.NamespaceName
	workflow := GetInodePath(im.instance.Instance.WorkflowPath)

	// Trim workflow revision until revisions are fully implemented.
	if divider := strings.LastIndex(workflow, ":"); divider > 0 {
		workflow = workflow[0:divider]
	}

	now := time.Now().UTC()
	empty := time.Time{}

	if im.Status() == util.InstanceStatusFailed || im.Status() == util.InstanceStatusCrashed {
		metricsWfFail.WithLabelValues(namespace, workflow, namespace).Inc()
	} else {
		metricsWfSuccess.WithLabelValues(namespace, workflow, namespace).Inc()
	}

	metricsWfOutcome.WithLabelValues(namespace, workflow, namespace, im.instance.Instance.Status.String(), im.instance.Instance.ErrorCode).Inc()
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Dec()

	if t != empty {
		ms := now.Sub(t).Milliseconds()
		metricsWfDuration.WithLabelValues(namespace, workflow, namespace).Observe(float64(ms))
	}
}
