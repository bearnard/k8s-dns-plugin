package k8sdns

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

// NewCache creates a new DNS record cache
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]map[uint16]CacheEntry),
	}
}

// Add adds DNS records to the cache
func (c *Cache) Add(name string, qtype uint16, records []dns.RR, ttl int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiry := time.Now().Add(time.Duration(ttl) * time.Second).Unix()

	if _, ok := c.entries[name]; !ok {
		c.entries[name] = make(map[uint16]CacheEntry)
	}

	c.entries[name][qtype] = CacheEntry{
		Records: records,
		Expiry:  expiry,
	}

	// Update metrics
	recordCount.WithLabelValues(dns.Type(qtype).String()).Inc()
}

// Get retrieves DNS records from the cache
func (c *Cache) Get(name string, qtype uint16) []dns.RR {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	now := time.Now().Unix()

	if qm, ok := c.entries[name]; ok {
		if entry, ok := qm[qtype]; ok {
			if entry.Expiry > now {
				return entry.Records
			}

			// Expired - remove from cache
			c.mutex.RUnlock()
			c.Remove(name, qtype)
			c.mutex.RLock()
		}
	}

	return nil
}

// Remove deletes DNS records from the cache
func (c *Cache) Remove(name string, qtype uint16) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if qm, ok := c.entries[name]; ok {
		if _, ok := qm[qtype]; ok {
			delete(qm, qtype)
			recordCount.WithLabelValues(dns.Type(qtype).String()).Dec()
		}

		if len(qm) == 0 {
			delete(c.entries, name)
		}
	}
}

// Delete removes all records for a given name
func (c *Cache) Delete(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if qm, ok := c.entries[name]; ok {
		for qtype := range qm {
			recordCount.WithLabelValues(dns.Type(qtype).String()).Dec()
		}
		delete(c.entries, name)
	}
}

// AddRecord adds a DNSRecord to cache
func (c *Cache) AddRecord(record *DNSRecord) {
	// Convert DNSRecord to dns.RR
	var rr dns.RR
	switch record.Type {
	case "A":
		rr = &dns.A{
			Hdr: dns.RR_Header{
				Name:   record.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(record.TTL),
			},
			A: net.ParseIP(record.Data),
		}
	}

	if rr != nil {
		c.Add(record.Name, rr.Header().Rrtype, []dns.RR{rr}, record.TTL)
	}
}
