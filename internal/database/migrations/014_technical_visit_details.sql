CREATE TABLE technical_visit_devices (
    id TEXT PRIMARY KEY,
    technical_visit_id TEXT NOT NULL REFERENCES technical_visits(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    role TEXT NOT NULL CHECK(role IN ('inspected','configured','installed','removed','replaced','defective','tested','updated')),
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    UNIQUE(technical_visit_id,device_id,role)
);

CREATE TABLE technical_visit_services (
    id TEXT PRIMARY KEY,
    technical_visit_id TEXT NOT NULL REFERENCES technical_visits(id) ON DELETE CASCADE,
    description TEXT NOT NULL CHECK(length(trim(description)) > 0),
    category TEXT NOT NULL DEFAULT '',
    device_id TEXT REFERENCES devices(id) ON DELETE RESTRICT,
    performed_at TEXT NOT NULL DEFAULT '',
    technician TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE technical_visit_checklist (
    id TEXT PRIMARY KEY,
    technical_visit_id TEXT NOT NULL REFERENCES technical_visits(id) ON DELETE CASCADE,
    text TEXT NOT NULL CHECK(length(trim(text)) > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','completed','not_applicable')),
    notes TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE technical_visit_materials (
    id TEXT PRIMARY KEY,
    technical_visit_id TEXT NOT NULL REFERENCES technical_visits(id) ON DELETE CASCADE,
    quantity REAL NOT NULL CHECK(quantity > 0),
    unit TEXT NOT NULL CHECK(length(trim(unit)) > 0),
    description TEXT NOT NULL CHECK(length(trim(description)) > 0),
    brand TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE technical_visit_pending_items (
    id TEXT PRIMARY KEY,
    technical_visit_id TEXT NOT NULL REFERENCES technical_visits(id) ON DELETE CASCADE,
    description TEXT NOT NULL CHECK(length(trim(description)) > 0),
    priority TEXT NOT NULL DEFAULT 'normal' CHECK(priority IN ('low','normal','high','critical')),
    responsible TEXT NOT NULL DEFAULT '',
    due_at TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','in_progress','resolved','cancelled')),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_technical_visit_devices_visit ON technical_visit_devices(technical_visit_id);
CREATE INDEX idx_technical_visit_devices_device ON technical_visit_devices(device_id);
CREATE INDEX idx_technical_visit_services_visit ON technical_visit_services(technical_visit_id,sort_order);
CREATE INDEX idx_technical_visit_checklist_visit ON technical_visit_checklist(technical_visit_id,sort_order);
CREATE INDEX idx_technical_visit_materials_visit ON technical_visit_materials(technical_visit_id);
CREATE INDEX idx_technical_visit_pending_visit ON technical_visit_pending_items(technical_visit_id,status,due_at);

CREATE TRIGGER technical_visit_devices_same_project_insert
BEFORE INSERT ON technical_visit_devices
WHEN NOT EXISTS (
    SELECT 1 FROM technical_visits v JOIN devices d ON d.id=NEW.device_id
    WHERE v.id=NEW.technical_visit_id AND v.project_id=d.project_id
)
BEGIN SELECT RAISE(ABORT,'technical visit and device must belong to the same project'); END;

CREATE TRIGGER technical_visit_devices_same_project_update
BEFORE UPDATE OF technical_visit_id,device_id ON technical_visit_devices
WHEN NOT EXISTS (
    SELECT 1 FROM technical_visits v JOIN devices d ON d.id=NEW.device_id
    WHERE v.id=NEW.technical_visit_id AND v.project_id=d.project_id
)
BEGIN SELECT RAISE(ABORT,'technical visit and device must belong to the same project'); END;

CREATE TRIGGER technical_visit_services_same_project_insert
BEFORE INSERT ON technical_visit_services
WHEN NEW.device_id IS NOT NULL AND NOT EXISTS (
    SELECT 1 FROM technical_visits v JOIN devices d ON d.id=NEW.device_id
    WHERE v.id=NEW.technical_visit_id AND v.project_id=d.project_id
)
BEGIN SELECT RAISE(ABORT,'technical visit and device must belong to the same project'); END;

CREATE TRIGGER technical_visit_services_same_project_update
BEFORE UPDATE OF technical_visit_id,device_id ON technical_visit_services
WHEN NEW.device_id IS NOT NULL AND NOT EXISTS (
    SELECT 1 FROM technical_visits v JOIN devices d ON d.id=NEW.device_id
    WHERE v.id=NEW.technical_visit_id AND v.project_id=d.project_id
)
BEGIN SELECT RAISE(ABORT,'technical visit and device must belong to the same project'); END;
