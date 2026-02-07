package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Registry is the Prometheus registry for all metrics
	Registry *prometheus.Registry
)

// InitPrometheus initializes the Prometheus registry
func InitPrometheus() {
	// Create a new registry
	Registry = prometheus.NewRegistry()

	// Register all collectors with the registry
	Registry.MustRegister(RequestsTotal)
	Registry.MustRegister(CacheHitsTotal)
	Registry.MustRegister(CacheMissesTotal)
	Registry.MustRegister(RequestDuration)
	Registry.MustRegister(ResponseSize)
	Registry.MustRegister(CacheSize)
	Registry.MustRegister(ActiveConnections)

	// Optionally register default Go metrics and process collectors
	Registry.MustRegister(prometheus.NewGoCollector())
	Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// Handler returns an HTTP handler for the /metrics endpoint
func Handler() http.Handler {
	if Registry == nil {
		InitPrometheus()
	}
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// GinHandler wraps the Prometheus handler for use with Gin framework
func GinHandler() http.HandlerFunc {
	handler := Handler()
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
