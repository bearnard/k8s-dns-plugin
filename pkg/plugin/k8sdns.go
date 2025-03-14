package k8sdns

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sDNS struct {
	Next       plugin.Handler
	Cache      *Cache
	Controller *Controller
	Zones      []string
}

func (k *K8sDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: r}

	// Check if we should handle this query
	zone := plugin.Zones(k.Zones).Matches(req.Name())
	if zone == "" {
		return plugin.NextOrFailure(k.Name(), k.Next, ctx, w, r)
	}

	// Increment metrics
	requestCount.WithLabelValues(zone, dns.Type(r.Question[0].Qtype).String()).Inc()

	// Handle DNS request and return appropriate response
	return k.handleRequest(req)
}

func (k *K8sDNS) handleRequest(req request.Request) (int, error) {
	// Check cache first
	if answer := k.Cache.Get(req.Name(), req.QType()); answer != nil {
		// Cache hit - increment metrics
		cacheHitCount.WithLabelValues(req.Zone, dns.Type(req.QType()).String()).Inc()

		// Build response from cache
		resp := new(dns.Msg)
		resp.SetReply(req.Req)
		resp.Answer = append(resp.Answer, answer...)
		req.W.WriteMsg(resp)
		return dns.RcodeSuccess, nil
	}

	// Cache miss - query Kubernetes data
	// This would be implemented in a real plugin

	return dns.RcodeSuccess, nil
}

func (k *K8sDNS) Name() string {
	return "k8sdns"
}

// Ready implements the Ready interface
func (k *K8sDNS) Ready() bool {
	return k.Controller != nil && k.Controller.HasSynced()
}

// Health implements the Health interface
func (k *K8sDNS) Health() bool {
	return k.Ready()
}

func (k *K8sDNS) OnStartup() error {
	// Initialize Kubernetes client and start the controller
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	k.Controller = NewController(clientset, k.Cache)
	go k.Controller.Run(context.Background())

	return nil
}

func (k *K8sDNS) OnShutdown() error {
	// Logic to stop the controller and cleanup resources
	if k.Controller != nil {
		k.Controller.Stop()
	}
	return nil
}

func New() *K8sDNS {
	return &K8sDNS{
		Cache: NewCache(),
		Zones: []string{"."}, // Default to all zones
	}
}
