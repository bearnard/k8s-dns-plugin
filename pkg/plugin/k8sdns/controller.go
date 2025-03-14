package k8sdns

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kcache "k8s.io/client-go/tools/cache" // Renamed to kcache to avoid name conflicts
	"k8s.io/klog/v2"
)

type Controller struct {
	client          kubernetes.Interface
	cache           *Cache
	informerFactory informers.SharedInformerFactory
	serviceInformer kcache.SharedIndexInformer // Use kcache here
	hasSyncedFunc   func() bool
	mu              sync.RWMutex
	stopCh          chan struct{}
}

func NewController(client kubernetes.Interface, cache *Cache) *Controller {
	factory := informers.NewSharedInformerFactory(client, time.Hour)
	serviceInformer := factory.Core().V1().Services().Informer()

	controller := &Controller{
		client:          client,
		cache:           cache,
		informerFactory: factory,
		serviceInformer: serviceInformer,
		stopCh:          make(chan struct{}),
	}

	// Use kcache instead of cache
	serviceInformer.AddEventHandler(kcache.ResourceEventHandlerFuncs{
		AddFunc:    controller.handleServiceAdd,
		UpdateFunc: controller.handleServiceUpdate,
		DeleteFunc: controller.handleServiceDelete,
	})

	controller.hasSyncedFunc = serviceInformer.HasSynced

	return controller
}

func (c *Controller) Run(ctx context.Context) {
	defer close(c.stopCh)

	c.informerFactory.Start(c.stopCh)

	// Use kcache here too
	if !kcache.WaitForCacheSync(c.stopCh, c.hasSyncedFunc) {
		klog.Error("Failed to sync informer caches")
		return
	}

	klog.Info("K8sDNS controller synced and ready")

	<-ctx.Done()
}

func (c *Controller) HasSynced() bool {
	return c.hasSyncedFunc()
}

func (c *Controller) Stop() {
	close(c.stopCh)
}

// Handler function implementations
func (c *Controller) handleServiceAdd(obj interface{}) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return
	}

	c.processService(service)
}

func (c *Controller) handleServiceUpdate(oldObj, newObj interface{}) {
	oldService, ok := oldObj.(*corev1.Service)
	if !ok {
		return
	}

	newService, ok := newObj.(*corev1.Service)
	if !ok {
		return
	}

	// If ExternalIPs haven't changed, skip processing
	if sliceEqual(oldService.Spec.ExternalIPs, newService.Spec.ExternalIPs) {
		return
	}

	c.processService(newService)
}

func (c *Controller) handleServiceDelete(obj interface{}) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return
	}

	// Remove all records for this service
	for _, hostname := range getHostnamesFromAnnotations(service, "dns.coredns.io/hostname") {
		c.cache.Delete(hostname)
	}
}

func (c *Controller) processService(service *corev1.Service) {
	// Extract records from service
	if len(service.Spec.ExternalIPs) == 0 {
		return
	}

	// Get hostnames from annotations
	hostnames := getHostnamesFromAnnotations(service, "dns.coredns.io/hostname")
	if len(hostnames) == 0 {
		return
	}

	// Get TTL from annotations or use default
	ttl := 300 // Default TTL
	if ttlStr, ok := service.Annotations["dns.coredns.io/ttl"]; ok {
		if val, err := strconv.Atoi(ttlStr); err == nil && val > 0 {
			ttl = val
		}
	}

	// Create DNS records
	for _, hostname := range hostnames {
		for _, externalIP := range service.Spec.ExternalIPs {
			// Create A records for each external IP
			rrs := []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   hostname,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    uint32(ttl),
					},
					A: net.ParseIP(externalIP),
				},
			}
			c.cache.Add(hostname, dns.TypeA, rrs, ttl)
		}
	}
}

// Helper function to check if two string slices are equal
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to extract hostnames from service annotations
func getHostnamesFromAnnotations(service *corev1.Service, annotationKey string) []string {
	if service.Annotations == nil {
		return nil
	}

	if hostname, ok := service.Annotations[annotationKey]; ok {
		return strings.Split(hostname, ",")
	}

	return nil
}
