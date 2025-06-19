package models

import (
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_requests",
			Help: "Number of active (queued) requests by computation type (LLM / CV / KReth).",
		},
		[]string{"type"},
	)
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Number of total computation requests.",
		},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration",
			Help:    "HTTP request duration in seconds by computation type (LLM / CV / KReth).",
			Buckets: []float64{0.01, 0.1, 1, 2, 5, 10, 25, 30, 45, 60, 120, 240, 480},
		},
		[]string{"type"},
	)
)

func MetricsMiddleware(c fiber.Ctx) error {
	requestType := string(c.Locals("query").(RelyingPartyQuery).Computation)

	// Log number of active requests
	activeRequests.WithLabelValues(requestType).Inc()
	defer activeRequests.WithLabelValues(requestType).Dec()
	// Increase total log
	totalRequests.Add(1)
	// The next middleware should also keep track of duration:
	err := c.Next()
	duration := c.Locals("duration").(float64)
	requestDuration.WithLabelValues(requestType).Observe(duration)

	return err

}

func MakePromRegistry() *prometheus.Registry {
	PromRegistry := prometheus.NewRegistry()
	PromRegistry.MustRegister(activeRequests)
	PromRegistry.MustRegister(totalRequests)
	PromRegistry.MustRegister(requestDuration)
	return PromRegistry
}
