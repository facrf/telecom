package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/local/telecom/internal/fingerprint"
)

type PortResult struct {
	Port       int     `json:"port"`
	Protocol   string  `json:"protocol"`
	State      string  `json:"state"`
	Service    string  `json:"service"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	Banner     string  `json:"banner"`
	Confidence float64 `json:"confidence"`
}
type PortScan struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	Mode       string `json:"mode"`
	Ports      string `json:"ports"`
	Status     string `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}
type PortManager struct {
	db      *sql.DB
	workers int
	timeout time.Duration
}

func (m *PortManager) Results(ctx context.Context, id string) ([]PortResult, error) {
	rows, e := m.db.QueryContext(ctx, "SELECT port,protocol,state,service,product,version,banner,confidence FROM scan_ports WHERE port_scan_id=? ORDER BY port", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []PortResult
	for rows.Next() {
		var value PortResult
		if e = rows.Scan(&value.Port, &value.Protocol, &value.State, &value.Service, &value.Product, &value.Version, &value.Banner, &value.Confidence); e != nil {
			return nil, e
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func (m *PortManager) List(ctx context.Context, deviceID string) ([]PortScan, error) {
	rows, e := m.db.QueryContext(ctx, "SELECT id,device_id,mode,ports,status,COALESCE(started_at,''),COALESCE(finished_at,'') FROM port_scans WHERE ?='' OR device_id=? ORDER BY created_at DESC", deviceID, deviceID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []PortScan
	for rows.Next() {
		var value PortScan
		if e = rows.Scan(&value.ID, &value.DeviceID, &value.Mode, &value.Ports, &value.Status, &value.StartedAt, &value.FinishedAt); e != nil {
			return nil, e
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func NewPortManager(db *sql.DB, workers int, timeout time.Duration) *PortManager {
	return &PortManager{db: db, workers: workers, timeout: timeout}
}
func (m *PortManager) Start(ctx context.Context, id, deviceID, mode, custom string) error {
	ports, e := PortsForMode(mode, custom)
	if e != nil {
		return e
	}
	if mode == "quick" {
		var configured string
		if m.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key='default_ports'").Scan(&configured) == nil {
			if parsed, parseErr := ParsePorts(configured, 4096); parseErr == nil {
				ports = parsed
			}
		}
	}
	var address string
	e = m.db.QueryRowContext(ctx, "SELECT address FROM device_addresses WHERE device_id=? AND type='ipv4' ORDER BY is_primary DESC LIMIT 1", deviceID).Scan(&address)
	if e != nil {
		return fmt.Errorf("equipamento sem endereço IPv4")
	}
	portText := custom
	if mode != "custom" {
		values := make([]string, len(ports))
		for index, port := range ports {
			values[index] = strconv.Itoa(port)
		}
		portText = strings.Join(values, ",")
	}
	_, e = m.db.ExecContext(ctx, "INSERT INTO port_scans(id,device_id,mode,ports,status,started_at)VALUES(?,?,?,?, 'running',CURRENT_TIMESTAMP)", id, deviceID, mode, portText)
	if e != nil {
		return e
	}
	workers := integerSetting(m.db, "scan_workers", m.workers, 1, 512)
	timeout := time.Duration(integerSetting(m.db, "scan_timeout_ms", int(m.timeout/time.Millisecond), 50, 30000)) * time.Millisecond
	go m.run(context.Background(), id, address, ports, workers, timeout)
	return nil
}
func (m *PortManager) run(ctx context.Context, id, address string, ports []int, workers int, timeout time.Duration) {
	tasks := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range tasks {
				if result := m.probe(ctx, address, port, timeout); result.State == "open" {
					_, _ = m.db.Exec("INSERT INTO scan_ports(id,port_scan_id,port,protocol,state,service,product,version,confidence,banner)VALUES(?,?,?,?,?,?,?,?,?,?)", id+"-"+strconv.Itoa(port), id, result.Port, result.Protocol, result.State, result.Service, result.Product, result.Version, result.Confidence, result.Banner)
				}
			}
		}()
	}
	for _, port := range ports {
		tasks <- port
	}
	close(tasks)
	wg.Wait()
	_, _ = m.db.Exec("UPDATE port_scans SET status='completed',finished_at=CURRENT_TIMESTAMP WHERE id=?", id)
	m.applyFingerprint(id)
}
func (m *PortManager) probe(ctx context.Context, address string, port int, timeout time.Duration) PortResult {
	r := PortResult{Port: port, Protocol: "tcp", State: "closed"}
	d := net.Dialer{Timeout: timeout}
	connection, e := d.DialContext(ctx, "tcp", net.JoinHostPort(address, strconv.Itoa(port)))
	if e != nil {
		return r
	}
	defer connection.Close()
	r.State = "open"
	r.Service = serviceHint(port)
	r.Confidence = .7
	_ = connection.SetDeadline(time.Now().Add(300 * time.Millisecond))
	if port == 80 || port == 8000 || port == 8080 || port == 8081 {
		_, _ = io.WriteString(connection, "HEAD / HTTP/1.0\r\nHost: "+address+"\r\n\r\n")
	}
	buffer := make([]byte, 1024)
	count, _ := connection.Read(buffer)
	if count > 0 {
		r.Banner = strings.TrimSpace(string(buffer[:count]))
		first := strings.Split(r.Banner, "\n")[0]
		if strings.HasPrefix(first, "SSH-") {
			r.Product = "SSH"
			r.Version = strings.TrimSpace(first)
		} else if strings.HasPrefix(first, "HTTP/") {
			r.Product = "HTTP"
			r.Version = strings.Fields(first)[0]
		}
	}
	return r
}

func (m *PortManager) applyFingerprint(scanID string) {
	var deviceID, hostname string
	if m.db.QueryRow("SELECT ps.device_id,d.hostname FROM port_scans ps JOIN devices d ON d.id=ps.device_id WHERE ps.id=?", scanID).Scan(&deviceID, &hostname) != nil {
		return
	}
	rows, err := m.db.Query("SELECT port FROM scan_ports WHERE port_scan_id=? AND state='open'", scanID)
	if err != nil {
		return
	}
	var ports []int
	for rows.Next() {
		var port int
		if rows.Scan(&port) == nil {
			ports = append(ports, port)
		}
	}
	rows.Close()
	var mac string
	_ = m.db.QueryRow("SELECT address FROM device_addresses WHERE device_id=? AND type='mac' ORDER BY is_primary DESC LIMIT 1", deviceID).Scan(&mac)
	identified := fingerprint.Identify(fingerprint.Input{MAC: mac, Hostname: hostname, Ports: ports})
	category := map[string]string{"Câmera IP": "ip-camera", "Equipamento de rede": "other", "Servidor ou equipamento de rede": "server"}[identified.DeviceType]
	_, _ = m.db.Exec("UPDATE devices SET manufacturer=CASE WHEN manufacturer='' THEN ? ELSE manufacturer END,category_id=CASE WHEN ?<>'' THEN ? ELSE category_id END,updated_at=CURRENT_TIMESTAMP WHERE id=?", identified.Manufacturer, category, category, deviceID)
}
func serviceHint(port int) string {
	hints := map[int]string{20: "FTP-data", 21: "FTP", 22: "SSH", 23: "Telnet", 53: "DNS", 80: "HTTP", 110: "POP3", 123: "NTP", 161: "SNMP", 389: "LDAP", 443: "HTTPS", 445: "SMB", 554: "RTSP", 1433: "MSSQL", 1883: "MQTT", 3306: "MySQL", 3389: "RDP", 5432: "PostgreSQL", 5900: "VNC", 8000: "HTTP-alt", 8080: "HTTP-alt", 8081: "HTTP-alt", 8443: "HTTPS-alt", 8554: "RTSP", 8883: "MQTTS", 9000: "HTTP-alt", 37777: "Dahua", 34567: "DVR"}
	return hints[port]
}
