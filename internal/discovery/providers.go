package discovery

import (
	"bufio"
	"context"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Result struct {
	IP       string
	MAC      string
	Hostname string
	Methods  []string
	Hint     string
}

func ProbeICMP(ctx context.Context, ip string) bool {
	if _, err := netip.ParseAddr(ip); err != nil {
		return false
	}
	ping, err := exec.LookPath("ping")
	if err != nil {
		return false
	}
	command := exec.CommandContext(ctx, ping, "-c", "1", "-W", "1", ip)
	return command.Run() == nil
}

func ARPTable() map[string]string {
	result := map[string]string{}
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return result
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if scanner.Scan() { /* header */
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 && fields[3] != "00:00:00:00:00:00" {
			result[fields[0]] = strings.ToUpper(fields[3])
		}
	}
	return result
}

func Multicast(ctx context.Context) map[string]Result {
	merged := map[string]Result{}
	var mutex sync.Mutex
	var wait sync.WaitGroup
	providers := []func(context.Context) []Result{discoverSSDP, discoverONVIF, discoverMDNS}
	for _, provider := range providers {
		wait.Add(1)
		go func(provider func(context.Context) []Result) {
			defer wait.Done()
			for _, item := range provider(ctx) {
				mutex.Lock()
				current := merged[item.IP]
				current.IP = item.IP
				current.Methods = appendUnique(current.Methods, item.Methods...)
				if len(item.Hint) > len(current.Hint) {
					current.Hint = item.Hint
				}
				merged[item.IP] = current
				mutex.Unlock()
			}
		}(provider)
	}
	wait.Wait()
	return merged
}

func discoverSSDP(ctx context.Context) []Result {
	payload := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: ssdp:all\r\n\r\n"
	return udpDiscover(ctx, "239.255.255.250:1900", []byte(payload), "ssdp")
}

func discoverONVIF(ctx context.Context) []Result {
	payload := `<?xml version="1.0" encoding="UTF-8"?><e:Envelope xmlns:e="http://www.w3.org/2003/05/soap-envelope" xmlns:w="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"><e:Header><w:MessageID>uuid:telecom-discovery</w:MessageID><w:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To><w:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</w:Action></e:Header><e:Body><d:Probe/></e:Body></e:Envelope>`
	return udpDiscover(ctx, "239.255.255.250:3702", []byte(payload), "onvif")
}

func discoverMDNS(ctx context.Context) []Result {
	// DNS query for _services._dns-sd._udp.local PTR.
	query := []byte{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 9, '_', 's', 'e', 'r', 'v', 'i', 'c', 'e', 's', 7, '_', 'd', 'n', 's', '-', 's', 'd', 4, '_', 'u', 'd', 'p', 5, 'l', 'o', 'c', 'a', 'l', 0, 0, 12, 0, 1}
	return udpDiscover(ctx, "224.0.0.251:5353", query, "mdns")
}

func udpDiscover(ctx context.Context, destination string, payload []byte, method string) []Result {
	address, err := net.ResolveUDPAddr("udp4", destination)
	if err != nil {
		return nil
	}
	connection, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(1500 * time.Millisecond))
	if _, err = connection.WriteToUDP(payload, address); err != nil {
		return nil
	}
	var results []Result
	seen := map[string]bool{}
	buffer := make([]byte, 65535)
	for {
		if ctx.Err() != nil {
			break
		}
		length, source, readErr := connection.ReadFromUDP(buffer)
		if readErr != nil {
			break
		}
		ip := source.IP.String()
		if !seen[ip] {
			seen[ip] = true
			hint := string(buffer[:minimum(length, 4096)])
			results = append(results, Result{IP: ip, Methods: []string{method}, Hint: hint})
		}
	}
	return results
}

func minimum(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}
