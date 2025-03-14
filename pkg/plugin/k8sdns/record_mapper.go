package k8sdns

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// No need for a custom RecordType, just use string constants
const (
	RecordTypeA     = "A"
	RecordTypeCNAME = "CNAME"
)

type RecordMapper struct {
	records map[string]DNSRecord
}

func NewRecordMapper() *RecordMapper {
	return &RecordMapper{
		records: make(map[string]DNSRecord),
	}
}

func (rm *RecordMapper) AddService(service *v1.Service) error {
	if service.Spec.ExternalIPs == nil || len(service.Spec.ExternalIPs) == 0 {
		return fmt.Errorf("service %s has no external IPs", service.Name)
	}

	// Use a safe TTL extraction
	ttl := getServiceTTL(service, 300)

	for _, ip := range service.Spec.ExternalIPs {
		record := DNSRecord{
			Name:       service.Name,
			Type:       RecordTypeA, // Use string constant
			TTL:        ttl,         // Use int instead of time.Duration
			ExternalIP: ip,          // Use string not net.IP
			Data:       ip,
		}
		rm.records[service.Name] = record
	}

	return nil
}

func (rm *RecordMapper) RemoveService(service *v1.Service) {
	delete(rm.records, service.Name)
}

func (rm *RecordMapper) GetRecord(name string) (DNSRecord, bool) {
	record, exists := rm.records[name]
	return record, exists
}

func (rm *RecordMapper) GetAllRecords() []DNSRecord {
	var allRecords []DNSRecord
	for _, record := range rm.records {
		allRecords = append(allRecords, record)
	}
	return allRecords
}

func (rm *RecordMapper) MapAnnotationsToRecords(service *v1.Service) error {
	// Use a safe TTL extraction
	ttl := getServiceTTL(service, 300)

	for key, value := range service.Annotations {
		if strings.HasPrefix(key, "dns.k8s.io/") {
			switch key {
			case "dns.k8s.io/cname":
				rm.records[service.Name] = DNSRecord{
					Name:       value,
					Type:       RecordTypeCNAME, // Use string constant
					TTL:        ttl,             // Use int instead of time.Duration
					ExternalIP: "",
					Data:       value,
				}
			}
		}
	}
	return nil
}

func createARecord(hostName string, ip string, ttl int) *DNSRecord {
	return &DNSRecord{
		Name:       hostName,
		Type:       RecordTypeA,
		TTL:        ttl,
		ExternalIP: ip,
		Data:       ip,
	}
}

func createCNAMERecord(hostName string, target string, ttl int) *DNSRecord {
	return &DNSRecord{
		Name:       hostName,
		Type:       RecordTypeCNAME,
		TTL:        ttl,
		ExternalIP: "",
		Data:       target,
	}
}

// Extract TTL safely from Service
func getServiceTTL(service *v1.Service, defaultTTL int) int {
	if service.Spec.SessionAffinityConfig != nil &&
		service.Spec.SessionAffinityConfig.ClientIP != nil &&
		service.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds != nil {
		// Convert *int32 to int safely
		return int(*service.Spec.SessionAffinityConfig.ClientIP.TimeoutSeconds)
	}
	return defaultTTL
}
