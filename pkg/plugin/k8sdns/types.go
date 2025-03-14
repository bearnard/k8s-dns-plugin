package k8sdns

import (
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// K8sDNS is the main plugin struct
type K8sDNS struct {
	Next             plugin.Handler
	Controller       *Controller
	Cache            *Cache
	Zones            []string
	DefaultTTL       int
	AnnotationPrefix string
	Namespaces       []string
}

// DNSRecord represents a single DNS record
type DNSRecord struct {
	Name       string
	Type       string
	TTL        int
	ExternalIP string
	Data       string // Added Data field that was missing
}

// ServiceAnnotation stores annotation data from k8s services
type ServiceAnnotation struct {
	RecordType string
	TTL        int
}

// Config for plugin configuration
type Config struct {
	Namespace string
	Labels    map[string]string
}

// CacheEntry represents a cached DNS record set
type CacheEntry struct {
	Records []dns.RR
	Expiry  int64
}

// Cache stores DNS records
type Cache struct {
	entries map[string]map[uint16]CacheEntry
	mutex   sync.RWMutex
}
