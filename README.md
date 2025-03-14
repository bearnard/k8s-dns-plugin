# k8s-dns-plugin

## Overview

The k8s-dns-plugin is a CoreDNS plugin designed to watch Kubernetes services and manage DNS records based on service annotations and external IPs. It provides support for various DNS record types and implements an in-cluster configuration approach with an in-memory cache that respects TTLs.

## Features

- Watches Kubernetes services for changes.
- Supports DNS record annotations for dynamic DNS management.
- Uses EXTERNAL-IPs of services for A records.
- Implements an in-memory cache for DNS records with TTL support.
- Supports multiple DNS record types (A, CNAME, etc.).
- In-cluster configuration for seamless integration with Kubernetes.

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd k8s-dns-plugin
   ```

2. Build the plugin:
   ```
   make build
   ```

3. Deploy the plugin to your Kubernetes cluster:
   ```
   kubectl apply -f deploy/
   ```

## Configuration

The plugin can be configured using annotations on Kubernetes services. The following annotations are supported:

- `external-dns.kubernetes.io/hostname`: Specifies the DNS hostname for the service.
- `external-dns.kubernetes.io/ttl`: Specifies the TTL for the DNS record.

## Usage

Once deployed, the k8s-dns-plugin will automatically watch for changes in Kubernetes services and update the DNS records accordingly. You can query the DNS records using standard DNS queries.

## Metrics

The plugin includes metrics collection for monitoring DNS queries and cache hits/misses. Metrics can be accessed via the configured metrics endpoint.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.