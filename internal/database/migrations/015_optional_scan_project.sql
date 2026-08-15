-- Permite que scans de rede sejam executados sem vinculo inicial obrigatorio com projeto
CREATE TABLE IF NOT EXISTS network_scans_v2 (
    id TEXT PRIMARY KEY,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    network TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('queued','running','completed','cancelled','failed')),
    hosts_scanned INTEGER NOT NULL DEFAULT 0,
    hosts_found INTEGER NOT NULL DEFAULT 0,
    started_at TEXT,
    finished_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO network_scans_v2 (id, project_id, network, status, hosts_scanned, hosts_found, started_at, finished_at, created_at)
SELECT id, CASE WHEN project_id = '' THEN NULL ELSE project_id END, network, status, hosts_scanned, hosts_found, started_at, finished_at, created_at
FROM network_scans;

DROP TABLE IF EXISTS network_scans;

ALTER TABLE network_scans_v2 RENAME TO network_scans;

CREATE INDEX IF NOT EXISTS idx_network_scans_project ON network_scans(project_id, created_at DESC);
