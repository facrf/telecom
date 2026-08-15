package scanner

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/local/telecom/internal/discovery"
	"github.com/local/telecom/internal/fingerprint"
)

type Scan struct {
	ID           string `json:"id"`
	ProjectID    string `json:"projectId"`
	Network      string `json:"network"`
	Status       string `json:"status"`
	HostsScanned int    `json:"hostsScanned"`
	HostsFound   int    `json:"hostsFound"`
	StartedAt    string `json:"startedAt"`
	FinishedAt   string `json:"finishedAt"`
}
type Host struct {
	IP               string                 `json:"ip"`
	MAC              string                 `json:"mac"`
	Hostname         string                 `json:"hostname"`
	Status           string                 `json:"status"`
	DiscoveryMethod  string                 `json:"discoveryMethod"`
	DiscoveryMethods []string               `json:"discoveryMethods"`
	Manufacturer     string                 `json:"manufacturer"`
	DeviceType       string                 `json:"deviceType"`
	Confidence       float64                `json:"confidence"`
	Evidence         []fingerprint.Evidence `json:"evidence"`
	OpenPorts        []int                  `json:"openPorts,omitempty"`
}
type Progress struct {
	ScanID       string `json:"scanId"`
	Status       string `json:"status"`
	HostsScanned int    `json:"hostsScanned"`
	HostsFound   int    `json:"hostsFound"`
	Total        int    `json:"total"`
	Host         *Host  `json:"host,omitempty"`
}
type Manager struct {
	db      *sql.DB
	workers int
	timeout time.Duration
	mutex   sync.RWMutex
	jobs    map[string]*job
}
type job struct {
	cancel      context.CancelFunc
	total       int
	workers     int
	timeout     time.Duration
	ports       []int
	subscribers map[chan Progress]struct{}
	mutex       sync.Mutex
}

func NewManager(db *sql.DB, workers int, timeout time.Duration) *Manager {
	return &Manager{db: db, workers: workers, timeout: timeout, jobs: map[string]*job{}}
}
func (m *Manager) Start(ctx context.Context, id, projectID, network string) (Scan, error) {
	target, err := ParseTarget(network, DefaultHostLimit, false)
	if err != nil {
		return Scan{}, err
	}
	var projectVal *string
	if strings.TrimSpace(projectID) != "" {
		p := strings.TrimSpace(projectID)
		projectVal = &p
	}
	scan := Scan{ID: id, ProjectID: projectID, Network: network, Status: "queued"}
	_, err = m.db.ExecContext(ctx, "INSERT INTO network_scans(id,project_id,network,status) VALUES(?,?,?,?)", id, projectVal, network, "queued")
	if err != nil {
		return Scan{}, fmt.Errorf("create scan: %w", err)
	}
	_, _ = m.db.ExecContext(ctx, "INSERT INTO audit_logs(id,action,entity_type,entity_id,details_json)VALUES(?,?,?,?,?)", randomID(), "scan_started", "network_scan", id, `{"network":`+strconv.Quote(network)+`}`)
	jobContext, cancel := context.WithCancel(context.Background())
	workers := integerSetting(m.db, "scan_workers", m.workers, 1, 512)
	timeout := time.Duration(integerSetting(m.db, "scan_timeout_ms", int(m.timeout/time.Millisecond), 50, 30000)) * time.Millisecond
	ports := append([]int(nil), QuickPorts...)
	var configuredPorts string
	if m.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key='default_ports'").Scan(&configuredPorts) == nil {
		if parsed, parseErr := ParsePorts(configuredPorts, 256); parseErr == nil {
			ports = parsed
		}
	}
	j := &job{cancel: cancel, total: int(target.Count), workers: workers, timeout: timeout, ports: ports, subscribers: map[chan Progress]struct{}{}}
	m.mutex.Lock()
	m.jobs[id] = j
	m.mutex.Unlock()
	go m.run(jobContext, scan, target, j)
	return scan, nil
}
func (m *Manager) Cancel(ctx context.Context, id string) error {
	m.mutex.RLock()
	j := m.jobs[id]
	m.mutex.RUnlock()
	if j == nil {
		return fmt.Errorf("scan não está em execução")
	}
	j.cancel()
	_, _ = m.db.ExecContext(ctx, "INSERT INTO audit_logs(id,action,entity_type,entity_id)VALUES(?,?,?,?)", randomID(), "scan_cancel_requested", "network_scan", id)
	return nil
}
func (m *Manager) Subscribe(id string) (<-chan Progress, func(), error) {
	m.mutex.RLock()
	j := m.jobs[id]
	m.mutex.RUnlock()
	if j == nil {
		return nil, nil, fmt.Errorf("scan não está em execução")
	}
	ch := make(chan Progress, 32)
	j.mutex.Lock()
	j.subscribers[ch] = struct{}{}
	j.mutex.Unlock()
	return ch, func() {
		j.mutex.Lock()
		if _, registered := j.subscribers[ch]; registered {
			delete(j.subscribers, ch)
			close(ch)
		}
		j.mutex.Unlock()
	}, nil
}
func (m *Manager) Get(ctx context.Context, id string) (Scan, error) {
	var s Scan
	err := m.db.QueryRowContext(ctx, "SELECT id,COALESCE(project_id,''),network,status,hosts_scanned,hosts_found,COALESCE(started_at,''),COALESCE(finished_at,'') FROM network_scans WHERE id=?", id).Scan(&s.ID, &s.ProjectID, &s.Network, &s.Status, &s.HostsScanned, &s.HostsFound, &s.StartedAt, &s.FinishedAt)
	return s, err
}
func (m *Manager) UpdateProject(ctx context.Context, id, projectID string) error {
	var projectVal *string
	if strings.TrimSpace(projectID) != "" {
		p := strings.TrimSpace(projectID)
		projectVal = &p
	}
	res, err := m.db.ExecContext(ctx, "UPDATE network_scans SET project_id=? WHERE id=?", projectVal, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	_, _ = m.db.ExecContext(ctx, "INSERT INTO audit_logs(id,action,entity_type,entity_id,details_json)VALUES(?,?,?,?,?)", randomID(), "scan_project_updated", "network_scan", id, `{"projectId":`+strconv.Quote(projectID)+`}`)
	return nil
}
func (m *Manager) List(ctx context.Context, projectID string) ([]Scan, error) {
	var rows *sql.Rows
	var err error
	if strings.TrimSpace(projectID) == "" {
		rows, err = m.db.QueryContext(ctx, "SELECT id,COALESCE(project_id,''),network,status,hosts_scanned,hosts_found,COALESCE(started_at,''),COALESCE(finished_at,'') FROM network_scans ORDER BY created_at DESC")
	} else {
		rows, err = m.db.QueryContext(ctx, "SELECT id,COALESCE(project_id,''),network,status,hosts_scanned,hosts_found,COALESCE(started_at,''),COALESCE(finished_at,'') FROM network_scans WHERE project_id=? ORDER BY created_at DESC", projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scans []Scan
	for rows.Next() {
		var scan Scan
		if err = rows.Scan(&scan.ID, &scan.ProjectID, &scan.Network, &scan.Status, &scan.HostsScanned, &scan.HostsFound, &scan.StartedAt, &scan.FinishedAt); err != nil {
			return nil, err
		}
		scans = append(scans, scan)
	}
	return scans, rows.Err()
}
func (m *Manager) Hosts(ctx context.Context, id string) ([]Host, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT ip,mac,hostname,status,discovery_method,discovery_methods,manufacturer,device_type,confidence,fingerprint_json,open_ports FROM scan_hosts WHERE scan_id=? ORDER BY ip", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hosts []Host
	for rows.Next() {
		var host Host
		var methodsJSON, fingerprintJSON, portsJSON string
		if err = rows.Scan(&host.IP, &host.MAC, &host.Hostname, &host.Status, &host.DiscoveryMethod, &methodsJSON, &host.Manufacturer, &host.DeviceType, &host.Confidence, &fingerprintJSON, &portsJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(methodsJSON), &host.DiscoveryMethods)
		_ = json.Unmarshal([]byte(portsJSON), &host.OpenPorts)
		var saved fingerprint.DeviceFingerprint
		if json.Unmarshal([]byte(fingerprintJSON), &saved) == nil {
			host.Evidence = saved.Evidence
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}
func (m *Manager) run(ctx context.Context, scan Scan, target TargetRange, j *job) {
	_, _ = m.db.Exec("UPDATE network_scans SET status='running',started_at=CURRENT_TIMESTAMP WHERE id=?", scan.ID)
	m.publish(j, Progress{ScanID: scan.ID, Status: "running", Total: j.total})
	jobs := make(chan string)
	var wg sync.WaitGroup
	var state struct {
		sync.Mutex
		scanned, found int
	}
	discoveryContext, cancelDiscovery := context.WithTimeout(ctx, 2*time.Second)
	advanced := discovery.Multicast(discoveryContext)
	cancelDiscovery()
	for range j.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				host := m.probe(ctx, ip, advanced[ip], j.ports, j.timeout)
				state.Lock()
				state.scanned++
				if host.Status == "online" {
					state.found++
					methodsJSON, _ := json.Marshal(host.DiscoveryMethods)
					portsJSON, _ := json.Marshal(host.OpenPorts)
					fingerprintJSON, _ := json.Marshal(fingerprint.DeviceFingerprint{Manufacturer: host.Manufacturer, DeviceType: host.DeviceType, Confidence: host.Confidence, Evidence: host.Evidence})
					_, _ = m.db.Exec("INSERT INTO scan_hosts(id,scan_id,ip,mac,hostname,status,discovery_method,discovery_methods,manufacturer,device_type,confidence,fingerprint_json,open_ports) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)", scan.ID+"-"+ip, scan.ID, host.IP, host.MAC, host.Hostname, host.Status, host.DiscoveryMethod, string(methodsJSON), host.Manufacturer, host.DeviceType, host.Confidence, string(fingerprintJSON), string(portsJSON))
				}
				p := Progress{ScanID: scan.ID, Status: "running", HostsScanned: state.scanned, HostsFound: state.found, Total: j.total, Host: &host}
				state.Unlock()
				m.publish(j, p)
			}
		}()
	}
	for ip := target.Start; ; ip = ip.Next() {
		select {
		case jobs <- ip.String():
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			m.finish(scan.ID, j, state.scanned, state.found, "cancelled")
			return
		}
		if ip == target.End {
			break
		}
	}
	close(jobs)
	wg.Wait()
	m.finish(scan.ID, j, state.scanned, state.found, "completed")
}
func (m *Manager) probe(ctx context.Context, ip string, advanced discovery.Result, ports []int, timeout time.Duration) Host {
	host := Host{IP: ip, Status: "offline", DiscoveryMethod: "tcp_connect", DiscoveryMethods: []string{}, Confidence: 0}
	for _, port := range ports {
		dialer := net.Dialer{Timeout: timeout}
		connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
		if err == nil {
			connection.Close()
			host.Status = "online"
			host.OpenPorts = append(host.OpenPorts, port)
			host.DiscoveryMethods = appendUnique(host.DiscoveryMethods, "tcp_connect")
			host.Confidence = 0.75
		}
	}
	if discovery.ProbeICMP(ctx, ip) {
		host.Status = "online"
		host.DiscoveryMethods = appendUnique(host.DiscoveryMethods, "icmp")
		if host.Confidence < .65 {
			host.Confidence = .65
		}
	}
	if advanced.IP != "" {
		host.Status = "online"
		host.DiscoveryMethods = appendUnique(host.DiscoveryMethods, advanced.Methods...)
		if host.Confidence < .8 {
			host.Confidence = .8
		}
	}
	if host.Status == "online" {
		names, _ := net.LookupAddr(ip)
		if len(names) > 0 {
			host.Hostname = names[0]
			host.Confidence = .85
		}
		if mac := discovery.ARPTable()[ip]; mac != "" {
			host.MAC = mac
			host.DiscoveryMethods = appendUnique(host.DiscoveryMethods, "arp")
		}
		identified := fingerprint.Identify(fingerprint.Input{MAC: host.MAC, Hostname: host.Hostname, Ports: host.OpenPorts, Methods: host.DiscoveryMethods, Banner: advanced.Hint})
		if identified.Manufacturer == "" && host.MAC != "" {
			prefix := strings.ToUpper(strings.ReplaceAll(host.MAC, ":", ""))
			if len(prefix) >= 6 {
				_ = m.db.QueryRow("SELECT vendor FROM oui_vendors WHERE prefix=?", prefix[:6]).Scan(&identified.Manufacturer)
			}
		}
		host.Manufacturer, host.DeviceType, host.Evidence = identified.Manufacturer, identified.DeviceType, identified.Evidence
		if identified.Confidence > host.Confidence {
			host.Confidence = identified.Confidence
		}
	}
	if len(host.DiscoveryMethods) > 0 {
		host.DiscoveryMethod = host.DiscoveryMethods[0]
	}
	return host
}
func (m *Manager) finish(id string, j *job, scanned, found int, status string) {
	_, _ = m.db.Exec("UPDATE network_scans SET status=?,hosts_scanned=?,hosts_found=?,finished_at=CURRENT_TIMESTAMP WHERE id=?", status, scanned, found, id)
	_, _ = m.db.Exec("INSERT INTO audit_logs(id,action,entity_type,entity_id,details_json)VALUES(?,?,?,?,?)", randomID(), "scan_"+status, "network_scan", id, fmt.Sprintf(`{"hostsScanned":%d,"hostsFound":%d}`, scanned, found))
	if status == "completed" {
		_ = m.persistDiff(id)
	}
	m.publish(j, Progress{ScanID: id, Status: status, HostsScanned: scanned, HostsFound: found, Total: j.total})
	m.mutex.Lock()
	delete(m.jobs, id)
	m.mutex.Unlock()
	j.mutex.Lock()
	for ch := range j.subscribers {
		close(ch)
	}
	j.subscribers = map[chan Progress]struct{}{}
	j.mutex.Unlock()
}

func (m *Manager) persistDiff(scanID string) error {
	var projectID sql.NullString
	var network string
	if err := m.db.QueryRow("SELECT project_id,network FROM network_scans WHERE id=?", scanID).Scan(&projectID, &network); err != nil {
		return err
	}
	var previousID string
	var queryErr error
	if projectID.Valid && strings.TrimSpace(projectID.String) != "" {
		queryErr = m.db.QueryRow("SELECT id FROM network_scans WHERE project_id=? AND network=? AND id<>? AND status='completed' ORDER BY finished_at DESC LIMIT 1", projectID.String, network, scanID).Scan(&previousID)
	} else {
		queryErr = m.db.QueryRow("SELECT id FROM network_scans WHERE (project_id IS NULL OR project_id='') AND network=? AND id<>? AND status='completed' ORDER BY finished_at DESC LIMIT 1", network, scanID).Scan(&previousID)
	}
	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			return nil
		}
		return queryErr
	}
	load := func(id string) ([]SnapshotHost, error) {
		rows, err := m.db.Query("SELECT ip,mac,hostname,open_ports FROM scan_hosts WHERE scan_id=?", id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var values []SnapshotHost
		for rows.Next() {
			var value SnapshotHost
			var portsJSON string
			if err = rows.Scan(&value.IP, &value.MAC, &value.Hostname, &portsJSON); err != nil {
				return nil, err
			}
			value.Ports = map[int]string{}
			var ports []int
			_ = json.Unmarshal([]byte(portsJSON), &ports)
			for _, port := range ports {
				value.Ports[port] = serviceHint(port)
			}
			values = append(values, value)
		}
		return values, rows.Err()
	}
	previous, err := load(previousID)
	if err != nil {
		return err
	}
	current, err := load(scanID)
	if err != nil {
		return err
	}
	for _, change := range Diff(previous, current) {
		_, err = m.db.Exec("INSERT INTO scan_events(id,scan_id,event_type,subject,previous_value,current_value)VALUES(?,?,?,?,?,?)", randomID(), scanID, change.Type, change.Subject, change.Previous, change.Current)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Events(ctx context.Context, scanID string) ([]Change, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT event_type,subject,previous_value,current_value FROM scan_events WHERE scan_id=? ORDER BY created_at", scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []Change
	for rows.Next() {
		var value Change
		if err = rows.Scan(&value.Type, &value.Subject, &value.Previous, &value.Current); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func integerSetting(db *sql.DB, key string, fallback, minimum, maximum int) int {
	var raw string
	if db.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&raw) != nil {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum {
		return fallback
	}
	return value
}
func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, v := range values {
		seen[v] = true
	}
	for _, v := range additions {
		if v != "" && !seen[v] {
			values = append(values, v)
			seen[v] = true
		}
	}
	return values
}
func randomID() string {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err == nil {
		return hex.EncodeToString(data)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
func (m *Manager) publish(j *job, p Progress) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	for ch := range j.subscribers {
		select {
		case ch <- p:
		default:
		}
	}
}
