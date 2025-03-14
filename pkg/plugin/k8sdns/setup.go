package k8sdns

import (
	"context"
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	caddy.RegisterPlugin("k8sdns", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	k, err := parseConfig(c)
	if err != nil {
		return plugin.Error("k8sdns", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		k.Next = next
		return k
	})

	return nil
}

func parseConfig(c *caddy.Controller) (*K8sDNS, error) {
	k8s := &K8sDNS{
		Zones:            []string{"."},
		Cache:            NewCache(),
		DefaultTTL:       3600,
		AnnotationPrefix: "dns.coredns.io",
	}

	for c.Next() {
		// Parse zones if any
		args := c.RemainingArgs()
		if len(args) > 0 {
			k8s.Zones = args
		}

		for c.NextBlock() {
			switch c.Val() {
			case "ttl":
				args := c.RemainingArgs()
				if len(args) == 1 {
					ttl, err := strconv.Atoi(args[0])
					if err != nil {
						return nil, c.Errf("invalid TTL value: %s", args[0])
					}
					k8s.DefaultTTL = ttl
				}
			case "annotation-prefix":
				args := c.RemainingArgs()
				if len(args) == 1 {
					k8s.AnnotationPrefix = args[0]
				}
			case "namespace":
				args := c.RemainingArgs()
				if len(args) >= 1 {
					k8s.Namespaces = args
				}
			}
		}
	}

	// Set up in-cluster Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	cache := NewCache()
	controller := NewController(clientset, cache)
	k8s.Controller = controller
	k8s.Cache = cache

	return k8s, nil
}

func (k *K8sDNS) OnStartup() error {
	go k.Controller.Run(context.Background())
	return nil
}

func (k *K8sDNS) Name() string {
	return "k8sdns"
}

// Ready implements the Ready method from plugin.Handler
func (k *K8sDNS) Ready() bool {
	return k.Controller != nil && k.Controller.HasSynced()
}

// Health implements the Health method from plugin.Handler
func (k *K8sDNS) Health() bool {
	return k.Ready()
}
