package direktiv

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
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
