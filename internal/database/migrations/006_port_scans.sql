CREATE TABLE port_scans (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    mode TEXT NOT NULL CHECK(mode IN ('quick','standard','custom')),
    ports TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued','running','completed','cancelled','failed')),
    started_at TEXT,
    finished_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE scan_ports (
    id TEXT PRIMARY KEY,
    port_scan_id TEXT NOT NULL REFERENCES port_scans(id) ON DELETE CASCADE,
    port INTEGER NOT NULL CHECK(port BETWEEN 1 AND 65535),
    protocol TEXT NOT NULL DEFAULT 'tcp',
    state TEXT NOT NULL CHECK(state IN ('open','closed','filtered')),
    service TEXT NOT NULL DEFAULT '',
    product TEXT NOT NULL DEFAULT '',
    version TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL DEFAULT 0,
    banner TEXT NOT NULL DEFAULT '',
    detected_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(port_scan_id, port, protocol)
);
CREATE INDEX idx_port_scans_device ON port_scans(device_id, created_at DESC);
CREATE INDEX idx_scan_ports_scan ON scan_ports(port_scan_id);
