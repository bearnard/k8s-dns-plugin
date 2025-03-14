package main

import (

	// Import the plugin correctly
	_ "github.com/bearnard/k8s-dns-plugin/pkg/plugin/k8sdns" // Register the plugin
	"github.com/coredns/coredns/coremain"
)

func main() {
	// This is the recommended way to integrate with CoreDNS
	// Register your plugin in the init() function of your plugin package
	// And then use coremain.Run()
	coremain.Run()
}
