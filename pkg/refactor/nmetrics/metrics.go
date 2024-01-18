package nmetrics

import (
	"errors"
	"net/http"
	"time"

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
			Name:      "cloudevents_received_total",
			Help:      "Total number of cloudevents received.",
		},
		[]string{"direktiv_namespace", "ce_type", "ce_source", "direktiv_tenant"},
	)

	metricsCloudEventsCaptured = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "cloudevents_captured_total",
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
			Name:      "pending",
			Help:      "Total number of workflows pending.",
		},
		[]string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "total_seconds",
			Help:      "Total time workflow has been actively executing.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant"},
	)

	metricsWfStateDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "direktiv",
			Subsystem: "states",
			Name:      "seconds",
			Help:      "Average time each state spends in execution.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "state", "direktiv_tenant"},
	)

	metricsWfOutcome = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "workflows",
			Name:      "outcomes_total",
			Help:      "Results of each workflow instance.",
		}, []string{"direktiv_namespace", "direktiv_workflow", "direktiv_tenant", "direktiv_instance_status", "direktiv_errcode"},
	)

	metricsGwRouteRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "gateway",
			Name:      "request_total",
			Help:      "Counter for the amount of requests per route",
		}, []string{"direktiv_route"},
	)

	metricsGwRouteErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "direktiv",
			Subsystem: "gateway",
			Name:      "errors_total",
			Help:      "Counter for the amount of errors per requests per route",
		}, []string{"direktiv_route_errors"},
	)
)

func SetupPrometheusEndpoint() error {
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
	prometheus.MustRegister(metricsGwRouteRequest)
	prometheus.MustRegister(metricsGwRouteErrors)

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":2112",
		ReadHeaderTimeout: time.Minute,
	}
	err := server.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func GwRouteRequest(route string) {
	metricsGwRouteRequest.WithLabelValues(route).Inc()
}

func GwRouteRequestError(route string) {
	metricsGwRouteRequest.WithLabelValues(route).Inc()
}

func WfFail(namespace string, workflow string) {
	metricsWfFail.WithLabelValues(namespace, workflow, namespace).Inc()
}

func WfSuccess(namespace string, workflow string) {
	metricsWfInvoked.WithLabelValues(namespace, workflow, namespace).Inc()
}

func WfInvoked(namespace string, workflow string) {
	metricsWfSuccess.WithLabelValues(namespace, workflow, namespace).Inc()
}

func WfPending(namespace string, workflow string) {
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Inc()
}

func WfOutcome(namespace string, workflow string, status string, errorCode string) {
	metricsWfOutcome.WithLabelValues(namespace, workflow, namespace, status, errorCode).Inc()
}

func WfPendingDec(namespace string, workflow string) {
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Dec()
}

func WfPendingInc(namespace string, workflow string) {
	metricsWfPending.WithLabelValues(namespace, workflow, namespace).Inc()
}

func WfObserveStateDuration(namespace string, workflow string, state string, ms float64) {
	metricsWfStateDuration.WithLabelValues(namespace, workflow, state, namespace).Observe(ms)
}

func WfInc(namspace string) {
	metricsWf.WithLabelValues(namspace, namspace).Inc()
}

func WfDec(namspace string) {
	metricsWf.WithLabelValues(namspace, namspace).Dec()
}

func WfUpdated(namspace string, filePath string) {
	metricsWf.WithLabelValues(namspace, filePath).Inc()
}

func CloudEventsCaptured(namespace string, eventType string, eventSource string) {
	metricsCloudEventsCaptured.WithLabelValues(namespace, eventType, eventSource, namespace).Inc()
}

func CloudEventsReceived(namespace string, eventType string, eventSource string) {
	metricsCloudEventsReceived.WithLabelValues(namespace, eventType, eventSource, namespace).Inc()
}
