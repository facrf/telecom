CREATE TABLE technical_visits (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    protocol TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL CHECK(length(trim(title)) > 0),
    visit_type TEXT NOT NULL CHECK(length(trim(visit_type)) > 0),
    status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','scheduled','in_progress','completed','cancelled')),
    result TEXT NOT NULL DEFAULT '' CHECK(result IN ('','resolved','partially_resolved','not_resolved','waiting_material','waiting_customer','requires_return','no_fault_found')),
    scheduled_at TEXT NOT NULL,
    started_at TEXT NOT NULL DEFAULT '',
    finished_at TEXT NOT NULL DEFAULT '',
    responsible_technician TEXT NOT NULL DEFAULT '',
    requester TEXT NOT NULL DEFAULT '',
    local_contact TEXT NOT NULL DEFAULT '',
    request_description TEXT NOT NULL DEFAULT '',
    initial_situation TEXT NOT NULL DEFAULT '',
    diagnosis TEXT NOT NULL DEFAULT '',
    work_summary TEXT NOT NULL DEFAULT '',
    recommendations TEXT NOT NULL DEFAULT '',
    pending_summary TEXT NOT NULL DEFAULT '',
    customer_notes TEXT NOT NULL DEFAULT '',
    internal_notes TEXT NOT NULL DEFAULT '',
    requires_return INTEGER NOT NULL DEFAULT 0 CHECK(requires_return IN (0,1)),
    return_reason TEXT NOT NULL DEFAULT '',
    suggested_return_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_technical_visits_project_id ON technical_visits(project_id);
CREATE INDEX idx_technical_visits_protocol ON technical_visits(protocol);
CREATE INDEX idx_technical_visits_scheduled_at ON technical_visits(scheduled_at);
CREATE INDEX idx_technical_visits_status ON technical_visits(status);
CREATE INDEX idx_technical_visits_result ON technical_visits(result);
