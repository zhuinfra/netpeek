package dns

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSServer represents a DNS server to test
type DNSServer struct {
	Name string
	IP   string
}

// DNSResult represents the result of a DNS lookup
type DNSResult struct {
	Server       string
	IPv4s        []string
	IPv6s        []string
	CNAMEChain   []string
	ResponseTime time.Duration
	Error        error
}

// CheckHijacking checks for DNS hijacking by comparing results from multiple DNS servers
func CheckHijacking(domain string, queryType string) {
	// DNS servers to test
	dnsServers := []DNSServer{
		{Name: "Local DNS", IP: ""}, // Empty means use local resolver
		{Name: "Aliyun (223.5.5.5)", IP: "223.5.5.5"},
		{Name: "Google (8.8.8.8)", IP: "8.8.8.8"},
	}

	// Validate query type
	if queryType != "A" && queryType != "AAAA" {
		queryType = "A" // Default to IPv4
	}

	// Print header
	fmt.Println("DNS Hijack Check Result")
	fmt.Println("========================================================================")
	fmt.Printf("Domain:         %s\n\n", domain)

	// Results slice
	results := make([]DNSResult, 0, len(dnsServers))

	// Perform DNS lookups
	for _, server := range dnsServers {
		result := lookupDomain(domain, server, queryType)
		results = append(results, result)
	}

	// Display CNAME chains
	displayCNAMEChains(results, queryType)

	// Display DNS server results in table
	displayDNSResultsTable(results, queryType)

	// Analyze results for hijacking
	analysis := analyzeResults(results, queryType)

	// Display final status
	displayFinalStatus(analysis)
}

// lookupDomain performs a DNS lookup using the specified server, including CNAME chain
func lookupDomain(domain string, server DNSServer, queryType string) DNSResult {
	result := DNSResult{
		Server:     server.Name,
		CNAMEChain: []string{domain},
	}

	start := time.Now()

	var ipv4s []string
	var ipv6s []string
	var err error
	var cnameChain []string

	// Set timeout for DNS query
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if server.IP == "" {
		// Use local resolver with default behavior (will use /etc/hosts)
		cnameChain, ipv4s, ipv6s, err = resolveUsingLocalResolver(domain, queryType)
	} else {
		// Use direct DNS query to bypass /etc/hosts
		if queryType == "A" {
			ipv4s, cnameChain, err = directDNSQueryA(timeoutCtx, domain, server.IP)
		} else if queryType == "AAAA" {
			ipv6s, cnameChain, err = directDNSQueryAAAA(timeoutCtx, domain, server.IP)
		}
	}

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}

	if len(cnameChain) > 0 {
		result.CNAMEChain = cnameChain
	}

	// Set IPv4 and IPv6 addresses
	result.IPv4s = ipv4s
	result.IPv6s = ipv6s

	return result
}

// resolveUsingLocalResolver resolves using the local system resolver (will use /etc/hosts)
func resolveUsingLocalResolver(domain string, queryType string) ([]string, []string, []string, error) {
	var cnameChain []string
	var ipv4s []string
	var ipv6s []string
	var err error

	// Resolve CNAME first
	cname, err := net.LookupCNAME(domain)
	if err == nil {
		cname = strings.TrimSuffix(cname, ".")
		if cname != domain {
			cnameChain = append(cnameChain, domain)
			cnameChain = append(cnameChain, cname)
		} else {
			cnameChain = append(cnameChain, domain)
		}
	} else {
		cnameChain = append(cnameChain, domain)
	}

	// Resolve IPs based on query type
	if queryType == "A" || queryType == "AAAA" {
		ips, err := net.LookupIP(domain)
		if err == nil {
			for _, ip := range ips {
				if ip.To4() != nil {
					if queryType == "A" {
						ipv4s = append(ipv4s, ip.String())
					}
				} else {
					if queryType == "AAAA" {
						ipv6s = append(ipv6s, ip.String())
					}
				}
			}
		}
	}

	return cnameChain, ipv4s, ipv6s, nil
}

// directDNSQueryA performs a direct DNS query for A records, bypassing /etc/hosts
func directDNSQueryA(ctx context.Context, domain string, dnsServer string) ([]string, []string, error) {
	// This is a simplified DNS query implementation
	// For production use, consider using a proper DNS library

	// Create a UDP connection to the DNS server
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(dnsServer),
		Port: 53,
	})
	if err != nil {
		return nil, []string{domain}, err
	}
	defer conn.Close()

	// Set deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// Prepare DNS query
	query := buildDNSQuery(domain, 1) // 1 = A record type

	// Send query
	_, err = conn.Write(query)
	if err != nil {
		return nil, []string{domain}, err
	}

	// Receive response
	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		return nil, []string{domain}, err
	}

	// Parse response
	ips, cname, err := parseDNSResponse(response[:n], domain, 1)
	if err != nil {
		return nil, []string{domain}, err
	}

	cnameChain := []string{domain}
	if cname != "" && cname != domain {
		cnameChain = append(cnameChain, cname)
	}

	return ips, cnameChain, nil
}

// directDNSQueryAAAA performs a direct DNS query for AAAA records, bypassing /etc/hosts
func directDNSQueryAAAA(ctx context.Context, domain string, dnsServer string) ([]string, []string, error) {
	// Create a UDP connection to the DNS server
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(dnsServer),
		Port: 53,
	})
	if err != nil {
		return nil, []string{domain}, err
	}
	defer conn.Close()

	// Set deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// Prepare DNS query
	query := buildDNSQuery(domain, 28) // 28 = AAAA record type

	// Send query
	_, err = conn.Write(query)
	if err != nil {
		return nil, []string{domain}, err
	}

	// Receive response
	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		return nil, []string{domain}, err
	}

	// Parse response
	ips, cname, err := parseDNSResponse(response[:n], domain, 28)
	if err != nil {
		return nil, []string{domain}, err
	}

	cnameChain := []string{domain}
	if cname != "" && cname != domain {
		cnameChain = append(cnameChain, cname)
	}

	return ips, cnameChain, nil
}

// buildDNSQuery builds a simple DNS query
func buildDNSQuery(domain string, qType uint16) []byte {
	var query bytes.Buffer

	// Transaction ID (random)
	query.Write([]byte{0x12, 0x34})

	// Flags (standard query)
	query.Write([]byte{0x01, 0x00})

	// Questions count
	query.Write([]byte{0x00, 0x01})

	// Answer records count
	query.Write([]byte{0x00, 0x00})

	// Authority records count
	query.Write([]byte{0x00, 0x00})

	// Additional records count
	query.Write([]byte{0x00, 0x00})

	// Question section
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		query.WriteByte(byte(len(label)))
		query.Write([]byte(label))
	}
	query.WriteByte(0x00) // End of domain

	// Query type
	binary.Write(&query, binary.BigEndian, qType)

	// Query class (IN)
	query.Write([]byte{0x00, 0x01})

	return query.Bytes()
}

// parseDNSResponse parses a DNS response
func parseDNSResponse(response []byte, domain string, qType uint16) ([]string, string, error) {
	var ips []string
	var cname string

	// Check response is long enough
	if len(response) < 12 {
		return nil, "", errors.New("invalid DNS response")
	}

	// Header section
	// Transaction ID: bytes 0-1
	// Flags: bytes 2-3
	// Questions: bytes 4-5
	// Answers: bytes 6-7
	answerCount := binary.BigEndian.Uint16(response[6:8])
	// Authority: bytes 8-9
	// Additional: bytes 10-11

	// Question section
	// Skip to answer section
	ptr := 12
	for ptr < len(response) {
		length := int(response[ptr])
		if length == 0 {
			ptr++
			break // End of domain
		}
		ptr += length + 1
	}
	// Skip query type and class
	ptr += 4

	// Answer section
	for i := 0; i < int(answerCount); i++ {
		if ptr >= len(response) {
			break
		}

		// Skip name (compressed or uncompressed)
		if (response[ptr] & 0xC0) == 0xC0 {
			// Compressed name
			ptr += 2
		} else {
			// Uncompressed name
			for ptr < len(response) {
				length := int(response[ptr])
				if length == 0 {
					ptr++
					break
				}
				ptr += length + 1
			}
		}

		if ptr+8 > len(response) {
			break
		}

		// Record type
		recordType := binary.BigEndian.Uint16(response[ptr : ptr+2])
		ptr += 2

		// Record class
		ptr += 2

		// TTL
		ptr += 4

		// Data length
		dataLength := binary.BigEndian.Uint16(response[ptr : ptr+2])
		ptr += 2

		if ptr+int(dataLength) > len(response) {
			break
		}

		// Record data
		data := response[ptr : ptr+int(dataLength)]
		ptr += int(dataLength)

		if recordType == qType {
			if qType == 1 {
				// A record
				if len(data) == 4 {
					ip := net.IPv4(data[0], data[1], data[2], data[3])
					ips = append(ips, ip.String())
				}
			} else if qType == 28 {
				// AAAA record
				if len(data) == 16 {
					ip := net.IP(data)
					ips = append(ips, ip.String())
				}
			}
		} else if recordType == 5 && cname == "" {
			// CNAME record (only capture first one)
			cname = parseDomainName(data, response)
		}
	}

	return ips, cname, nil
}

// parseDomainName parses a domain name from DNS response (handling compression)
func parseDomainName(data []byte, response []byte) string {
	var domain strings.Builder
	ptr := 0

	for ptr < len(data) {
		length := int(data[ptr])
		if length == 0 {
			break
		}

		if (length & 0xC0) == 0xC0 {
			// Compressed name
			if ptr+1 >= len(data) {
				break
			}
			compressedPtr := int(binary.BigEndian.Uint16(data[ptr:ptr+2]) & 0x3FFF)
			if compressedPtr < len(response) {
				if domain.Len() > 0 {
					domain.WriteString(".")
				}
				domain.WriteString(parseDomainName(response[compressedPtr:], response))
			}
			break
		} else {
			// Uncompressed label
			if domain.Len() > 0 {
				domain.WriteString(".")
			}
			if ptr+1+length > len(data) {
				break
			}
			domain.Write(data[ptr+1 : ptr+1+length])
			ptr += 1 + length
		}
	}

	return domain.String()
}

// displayCNAMEChains displays the CNAME chains for each DNS server
func displayCNAMEChains(results []DNSResult, queryType string) {
	fmt.Println("CNAME Chain:")

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		// Display server name
		serverLabel := ""
		switch result.Server {
		case "Local DNS":
			serverLabel = "Local DNS:"
		case "Aliyun (223.5.5.5)":
			serverLabel = "Public DNS (Ali):"
		case "Google (8.8.8.8)":
			serverLabel = "Public DNS (Google):"
		default:
			serverLabel = fmt.Sprintf("%s:", result.Server)
		}

		fmt.Printf("  %s\n", serverLabel)

		// Display CNAME chain
		for i, name := range result.CNAMEChain {
			indent := strings.Repeat("    ", i+1)

			if i == len(result.CNAMEChain)-1 {
				// Last entry, display IPs
				fmt.Printf("%s-> %s\n", indent, name)

				if queryType == "A" || queryType == "" {
					// Display IPv4 addresses
					if len(result.IPv4s) > 0 {
						for _, ip := range result.IPv4s {
							fmt.Printf("%s  -> A %s\n", indent, ip)
						}
					} else {
						fmt.Printf("%s  -> No A records\n", indent)
					}
				} else if queryType == "AAAA" {
					// Display IPv6 addresses
					if len(result.IPv6s) > 0 {
						for _, ip := range result.IPv6s {
							fmt.Printf("%s  -> AAAA %s\n", indent, ip)
						}
					} else {
						fmt.Printf("%s  -> No AAAA records\n", indent)
					}
				}
			} else {
				// Intermediate CNAME record
				fmt.Printf("%s-> %s (CNAME)\n", indent, name)
			}
		}

		fmt.Println()
	}
}

// displayDNSResultsTable displays the DNS lookup results in a table
func displayDNSResultsTable(results []DNSResult, queryType string) {
	// Print table header line
	fmt.Println("DNS Server              IP Addresses                          Time    Status")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, result := range results {
		var status, ipsStr string

		if result.Error != nil {
			status = "ERROR"
			ipsStr = result.Error.Error()
		} else {
			status = "OK"
			var ips []string
			if queryType == "A" || queryType == "" {
				ips = result.IPv4s
			} else if queryType == "AAAA" {
				ips = result.IPv6s
			}

			ipsStr = strings.Join(ips, ", ")
			if ipsStr == "" {
				ipsStr = "No records found"
			}
		}

		// Format output line with proper spacing
		fmt.Printf("%-22s %-35s %-8s %s\n", result.Server, truncateString(ipsStr, 35), formatDuration(result.ResponseTime), status)
	}

	fmt.Println()
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// analyzeResults analyzes the results for potential DNS hijacking
func analyzeResults(results []DNSResult, queryType string) map[string]bool {
	analysis := map[string]bool{
		"differentCNAME": false,
		"differentIPs":   false,
		"abnormalLocal":  false,
		"suspicious":     false,
	}

	// Group results by IP sets (ignoring order)
	ipGroups := make(map[string][]string)
	cnameGroups := make(map[string][]string)
	validResults := 0

	for _, result := range results {
		var ips []string
		if queryType == "A" || queryType == "" {
			ips = result.IPv4s
		} else if queryType == "AAAA" {
			ips = result.IPv6s
		}

		if result.Error == nil && len(ips) > 0 {
			validResults++
			// Create a sorted string of IPs for grouping (to ignore order)
			sortedIPs := make([]string, len(ips))
			copy(sortedIPs, ips)
			for i := 0; i < len(sortedIPs); i++ {
				for j := i + 1; j < len(sortedIPs); j++ {
					if sortedIPs[i] > sortedIPs[j] {
						sortedIPs[i], sortedIPs[j] = sortedIPs[j], sortedIPs[i]
					}
				}
			}
			ipKey := strings.Join(sortedIPs, ",")
			ipGroups[ipKey] = append(ipGroups[ipKey], result.Server)

			// Create a string of CNAME chain for grouping
			cnameKey := strings.Join(result.CNAMEChain, "->")
			cnameGroups[cnameKey] = append(cnameGroups[cnameKey], result.Server)
		}
	}

	if validResults == 0 {
		return analysis
	}

	// Check if there are different IP groups
	if len(ipGroups) > 1 {
		analysis["differentIPs"] = true
		analysis["suspicious"] = true
	}

	// Check if there are different CNAME chains
	if len(cnameGroups) > 1 {
		analysis["differentCNAME"] = true
		analysis["suspicious"] = true
	}

	// Check if local DNS has different results
	for _, result := range results {
		if result.Server == "Local DNS" && result.Error == nil {
			var localIPs []string
			if queryType == "A" || queryType == "" {
				localIPs = result.IPv4s
			} else if queryType == "AAAA" {
				localIPs = result.IPv6s
			}

			if len(localIPs) > 0 {
				// Create sorted IP key for local result
				sortedLocalIPs := make([]string, len(localIPs))
				copy(sortedLocalIPs, localIPs)
				for i := 0; i < len(sortedLocalIPs); i++ {
					for j := i + 1; j < len(sortedLocalIPs); j++ {
						if sortedLocalIPs[i] > sortedLocalIPs[j] {
							sortedLocalIPs[i], sortedLocalIPs[j] = sortedLocalIPs[j], sortedLocalIPs[i]
						}
					}
				}
				localIPKey := strings.Join(sortedLocalIPs, ",")

				// Check if local IPs match any public DNS IPs (ignoring order)
				localHasMatch := false
				for ipKey, servers := range ipGroups {
					if ipKey == localIPKey {
						// Check if this group contains public DNS servers
						for _, server := range servers {
							if server != "Local DNS" {
								localHasMatch = true
								break
							}
						}
						if localHasMatch {
							break
						}
					}
				}
				if !localHasMatch {
					analysis["abnormalLocal"] = true
					analysis["suspicious"] = true
				}
			}
		}
	}

	return analysis
}

// displayFinalStatus displays the final analysis and status
func displayFinalStatus(analysis map[string]bool) {
	fmt.Println("Analysis:")

	if analysis["differentCNAME"] {
		fmt.Println("  - Local DNS returned different CNAME chain")
	}

	if analysis["differentIPs"] {
		fmt.Println("  - Local IP does not match public DNS results")
	}

	if analysis["suspicious"] {
		fmt.Println("  - Resolution result is suspicious")
	}

	if !analysis["differentCNAME"] && !analysis["differentIPs"] && !analysis["abnormalLocal"] && !analysis["suspicious"] {
		fmt.Println("  - No DNS hijacking detected")
		fmt.Println("  - All DNS servers returned consistent results")
		fmt.Println()
		fmt.Println("Status: LOW RISK - NO DNS HIJACKING DETECTED")
		return
	}

	fmt.Println()

	if analysis["suspicious"] {
		fmt.Println("Status: HIGH RISK - POSSIBLE DNS HIJACKING")
	} else {
		fmt.Println("Status: MEDIUM RISK - POTENTIAL ISSUES DETECTED")
	}
}

// formatDuration formats the duration for display
func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
