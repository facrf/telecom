CREATE TABLE network_scans (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    network TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued','running','completed','cancelled','failed')),
    hosts_scanned INTEGER NOT NULL DEFAULT 0,
    hosts_found INTEGER NOT NULL DEFAULT 0,
    started_at TEXT,
    finished_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE scan_hosts (
    id TEXT PRIMARY KEY,
    scan_id TEXT NOT NULL REFERENCES network_scans(id) ON DELETE CASCADE,
    ip TEXT NOT NULL,
    hostname TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK(status IN ('online','offline','unknown')),
    discovery_method TEXT NOT NULL DEFAULT 'tcp_connect',
    confidence REAL NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scan_id, ip)
);
CREATE INDEX idx_network_scans_project ON network_scans(project_id, created_at DESC);
CREATE INDEX idx_scan_hosts_scan ON scan_hosts(scan_id);
CREATE INDEX idx_scan_hosts_ip ON scan_hosts(ip);
