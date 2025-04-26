package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wallet_service_requests_total",
			Help: "Total number of requests to wallet service",
		},
		[]string{"handler", "method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wallet_service_request_duration_seconds",
			Help:    "Histogram of request processing durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method"},
	)
)

func Register() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
