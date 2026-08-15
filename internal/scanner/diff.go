package scanner

import (
	"fmt"
	"sort"
)

type SnapshotHost struct {
	IP, MAC, Hostname string
	Ports             map[int]string
}

type Change struct {
	Type     string `json:"type"`
	Subject  string `json:"subject"`
	Previous string `json:"previous"`
	Current  string `json:"current"`
}

func Diff(previous, current []SnapshotHost) []Change {
	oldByKey := map[string]SnapshotHost{}
	newByKey := map[string]SnapshotHost{}
	key := func(h SnapshotHost) string {
		if h.MAC != "" {
			return "mac:" + h.MAC
		}
		return "ip:" + h.IP
	}
	for _, h := range previous {
		oldByKey[key(h)] = h
	}
	for _, h := range current {
		newByKey[key(h)] = h
	}
	var changes []Change
	for k, next := range newByKey {
		old, exists := oldByKey[k]
		if !exists {
			changes = append(changes, Change{"host_new", next.IP, "", next.IP})
			continue
		}
		if old.IP != next.IP {
			changes = append(changes, Change{"ip_changed", next.IP, old.IP, next.IP})
		}
		if old.Hostname != next.Hostname {
			changes = append(changes, Change{"hostname_changed", next.IP, old.Hostname, next.Hostname})
		}
		changes = append(changes, portChanges(next.IP, old.Ports, next.Ports)...)
	}
	for k, old := range oldByKey {
		if _, exists := newByKey[k]; !exists {
			changes = append(changes, Change{"host_missing", old.IP, old.IP, ""})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Type+changes[i].Subject < changes[j].Type+changes[j].Subject })
	return changes
}

func portChanges(subject string, old, current map[int]string) []Change {
	var changes []Change
	for port, service := range current {
		oldService, exists := old[port]
		target := fmt.Sprintf("%s:%d", subject, port)
		if !exists {
			changes = append(changes, Change{"port_opened", target, "", service})
		} else if oldService != service {
			changes = append(changes, Change{"service_changed", target, oldService, service})
		}
	}
	for port, service := range old {
		if _, exists := current[port]; !exists {
			target := fmt.Sprintf("%s:%d", subject, port)
			changes = append(changes, Change{"port_closed", target, service, ""})
		}
	}
	return changes
}

