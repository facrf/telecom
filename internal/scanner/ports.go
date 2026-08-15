package scanner

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var QuickPorts = []int{22, 53, 80, 443, 554, 161, 5000, 5060, 8000, 8006, 8080, 8291, 8443, 8554, 9100, 10001, 37777, 34567}
var StandardPorts = []int{20, 21, 22, 23, 25, 53, 67, 68, 80, 110, 123, 135, 137, 138, 139, 161, 389, 443, 445, 500, 515, 554, 631, 1433, 1883, 2049, 3306, 3389, 4444, 5000, 5001, 5060, 5061, 5432, 5900, 8000, 8006, 8080, 8081, 8291, 8443, 8554, 8883, 9000, 9100, 10001, 10443, 37777, 34567}


func ParsePorts(input string, limit int) ([]int, error) {
	if limit == 0 {
		limit = 4096
	}
	seen := map[int]struct{}{}
	for _, part := range strings.Split(strings.TrimSpace(input), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("lista de portas inválida")
		}
		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("faixa de portas inválida")
			}
			start, e := parsePort(bounds[0])
			if e != nil {
				return nil, e
			}
			end, e := parsePort(bounds[1])
			if e != nil || end < start {
				return nil, fmt.Errorf("faixa de portas inválida")
			}
			for p := start; p <= end; p++ {
				seen[p] = struct{}{}
			}
		} else {
			p, e := parsePort(part)
			if e != nil {
				return nil, e
			}
			seen[p] = struct{}{}
		}
		if len(seen) > limit {
			return nil, fmt.Errorf("limite de %d portas excedido", limit)
		}
	}
	ports := make([]int, 0, len(seen))
	for p := range seen {
		ports = append(ports, p)
	}
	sort.Ints(ports)
	return ports, nil
}
func parsePort(value string) (int, error) {
	p, e := strconv.Atoi(strings.TrimSpace(value))
	if e != nil || p < 1 || p > 65535 {
		return 0, fmt.Errorf("porta deve estar entre 1 e 65535")
	}
	return p, nil
}
func PortsForMode(mode, custom string) ([]int, error) {
	switch mode {
	case "quick":
		return append([]int(nil), QuickPorts...), nil
	case "standard":
		return append([]int(nil), StandardPorts...), nil
	case "custom":
		return ParsePorts(custom, 4096)
	default:
		return nil, fmt.Errorf("modo de scan inválido")
	}
}
