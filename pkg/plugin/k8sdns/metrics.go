package k8sdns

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// requestCount counts the number of requests serviced by this plugin
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "k8sdns",
		Name:      "requests_total",
		Help:      "Counter of DNS requests processed by the k8sdns plugin.",
	}, []string{"zone", "type"})

	// cacheHitCount counts cache hits by zone and record type
	cacheHitCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "k8sdns",
		Name:      "cache_hits_total",
		Help:      "Counter of cache hits.",
	}, []string{"zone", "type"})

	// recordCount exports the number of records we have in our cache
	recordCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "k8sdns",
		Name:      "records_count",
		Help:      "Number of DNS records in the cache.",
	}, []string{"type"})
)

// REMOVE THIS INIT FUNCTION - promauto already registers metrics
// func init() {
//     prometheus.MustRegister(requestCount)
//     prometheus.MustRegister(cacheHitCount)
//     prometheus.MustRegister(recordCount)
// }

func RecordDNSQuery(recordType string) {
	requestCount.WithLabelValues("default", recordType).Inc()
}

func RecordCacheHit(zone, recordType string) {
	cacheHitCount.WithLabelValues(zone, recordType).Inc()
}

func RecordCacheMiss() {
	// This function is no longer needed and can be removed
}
