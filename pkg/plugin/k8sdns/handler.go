package k8sdns

import (
	"context"
	"net"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// ServeDNS implements the plugin.Handler interface
func (k *K8sDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: r}

	// Check if we should handle this query
	zone := plugin.Zones(k.Zones).Matches(req.Name())
	if zone == "" {
		return plugin.NextOrFailure(k.Name(), k.Next, ctx, w, r)
	}

	// Increment metrics
	requestCount.WithLabelValues(zone, dns.Type(r.Question[0].Qtype).String()).Inc()

	// Handle DNS request
	return k.handleRequest(ctx, req)
}

func (k *K8sDNS) handleRequest(ctx context.Context, req request.Request) (int, error) {
	// Lookup in cache
	if answer := k.Cache.Get(req.Name(), req.QType()); answer != nil {
		// Cache hit
		cacheHitCount.WithLabelValues(req.Zone, dns.Type(req.QType()).String()).Inc()

		resp := new(dns.Msg)
		resp.SetReply(req.Req)
		resp.Answer = append(resp.Answer, answer...)
		req.W.WriteMsg(resp)
		return dns.RcodeSuccess, nil
	}

	// Cache miss - lookup from k8s data
	return dns.RcodeServerFailure, nil
}

func (k *K8sDNS) UpdateRecords(serviceName string, serviceNamespace string, externalIPs []string, ttl time.Duration) {
	for _, ip := range externalIPs {
		if net.ParseIP(ip) != nil {
			record := &DNSRecord{
				Name:       serviceName,
				Type:       "A",
				TTL:        int(ttl.Seconds()),
				ExternalIP: ip,
				Data:       ip,
			}
			k.Cache.AddRecord(record)
		}
	}
}

func (k *K8sDNS) DeleteRecords(serviceName string) {
	k.Cache.Delete(serviceName)
}
