package direktiv

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/metrics"

	"github.com/vorteil/direktiv/pkg/ingress"
)

var (
	metricsWfInvoked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "invoked_total",
			Help:      "Total number of workflows invoked.",
		},
		[]string{"namespace", "workflow", "tenant"},
	)

	metricsWfSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "success_total",
			Help:      "Total number of workflows sucessfully finished.",
		},
		[]string{"namespace", "workflow", "tenant"},
	)

	metricsWfFail = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "failed_total",
			Help:      "Total number of workflows failed.",
		},
		[]string{"namespace", "workflow", "tenant"},
	)

	metricsWfDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "total_milliseconds",
			Help:      "Total time workflow has been actively executing.",
		}, []string{"namespace", "workflow", "tenant"},
	)

	metricsWfStateDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "states",
			Name:      "milliseconds",
			Help:      "Average time each state spends in execution.",
		}, []string{"namespace", "workflow", "state", "tenant"},
	)
)

func reportMetricEnd(namespace, workflow, status string, t time.Time) {

	now := time.Now()
	empty := time.Time{}

	log.Debugf("reporting workflow %v/%v: %v", namespace, workflow, status)
	if status == "failed" {
		metricsWfFail.WithLabelValues(namespace, workflow, namespace).Inc()
	} else {
		metricsWfSuccess.WithLabelValues(namespace, workflow, namespace).Inc()
	}

	if t != empty {
		ms := now.Sub(t).Milliseconds()
		metricsWfDuration.WithLabelValues(namespace, workflow, namespace).Observe(float64(ms))
	}
}

func reportStateEnd(namespace, workflow, state string, t time.Time) {

	ms := time.Now().Sub(t).Milliseconds()
	metricsWfStateDuration.WithLabelValues(namespace, workflow, state, namespace).Observe(float64(ms))

}

func setupPrometheusEndpoint() {

	log.Infof("starting prometheus endpoint")

	prometheus.MustRegister(metricsWfInvoked)
	prometheus.MustRegister(metricsWfSuccess)
	prometheus.MustRegister(metricsWfFail)
	prometheus.MustRegister(metricsWfDuration)
	prometheus.MustRegister(metricsWfStateDuration)
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

}

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
