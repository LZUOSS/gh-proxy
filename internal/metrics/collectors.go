package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts total HTTP requests with labels for method, path, and status
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "github_proxy_requests_total",
			Help: "Total number of HTTP requests processed by the GitHub proxy",
		},
		[]string{"method", "path", "status"},
	)

	// CacheHitsTotal counts cache hits by type (memory, disk)
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "github_proxy_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"type"},
	)

	// CacheMissesTotal counts cache misses by type (memory, disk)
	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "github_proxy_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"type"},
	)

	// RequestDuration measures HTTP request duration in seconds
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "github_proxy_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "path"},
	)

	// ResponseSize measures HTTP response size in bytes
	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "github_proxy_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"path"},
	)

	// CacheSize tracks current cache size in bytes by type (memory, disk)
	CacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "github_proxy_cache_size_bytes",
			Help: "Current cache size in bytes",
		},
		[]string{"type"},
	)

	// ActiveConnections tracks the number of active connections
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "github_proxy_active_connections",
			Help: "Number of active connections",
		},
	)
)

// RecordRequest records an HTTP request with its method, path, and status
func RecordRequest(method, path, status string) {
	RequestsTotal.WithLabelValues(method, path, status).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordRequestDuration records the duration of an HTTP request
func RecordRequestDuration(method, path string, duration float64) {
	RequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordResponseSize records the size of an HTTP response
func RecordResponseSize(path string, size float64) {
	ResponseSize.WithLabelValues(path).Observe(size)
}

// SetCacheSize sets the current cache size
func SetCacheSize(cacheType string, size float64) {
	CacheSize.WithLabelValues(cacheType).Set(size)
}

// IncrementActiveConnections increments the active connections counter
func IncrementActiveConnections() {
	ActiveConnections.Inc()
}

// DecrementActiveConnections decrements the active connections counter
func DecrementActiveConnections() {
	ActiveConnections.Dec()
}

// SetActiveConnections sets the active connections gauge to a specific value
func SetActiveConnections(count float64) {
	ActiveConnections.Set(count)
}
