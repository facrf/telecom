CREATE TABLE scan_events (
    id TEXT PRIMARY KEY,
    scan_id TEXT NOT NULL REFERENCES network_scans(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK(event_type IN ('host_new','host_missing','host_returned','ip_changed','hostname_changed','port_opened','port_closed','service_changed')),
    subject TEXT NOT NULL,
    previous_value TEXT NOT NULL DEFAULT '',
    current_value TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_scan_events_scan ON scan_events(scan_id, created_at DESC);
