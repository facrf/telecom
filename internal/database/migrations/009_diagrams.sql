CREATE TABLE diagrams (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE diagram_nodes (
    id TEXT PRIMARY KEY,
    diagram_id TEXT NOT NULL REFERENCES diagrams(id) ON DELETE CASCADE,
    device_id TEXT REFERENCES devices(id) ON DELETE SET NULL,
    label TEXT NOT NULL DEFAULT '',
    x REAL NOT NULL DEFAULT 0,
    y REAL NOT NULL DEFAULT 0,
    width REAL NOT NULL DEFAULT 180,
    height REAL NOT NULL DEFAULT 100,
    style_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE diagram_edges (
    id TEXT PRIMARY KEY,
    diagram_id TEXT NOT NULL REFERENCES diagrams(id) ON DELETE CASCADE,
    source_node_id TEXT NOT NULL REFERENCES diagram_nodes(id) ON DELETE CASCADE,
    target_node_id TEXT NOT NULL REFERENCES diagram_nodes(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'Ethernet',
    source_interface TEXT NOT NULL DEFAULT '',
    target_interface TEXT NOT NULL DEFAULT '',
    speed TEXT NOT NULL DEFAULT '',
    vlan TEXT NOT NULL DEFAULT '',
    technology TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '#64748b',
    line_style TEXT NOT NULL DEFAULT 'solid',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_diagrams_project ON diagrams(project_id);
CREATE INDEX idx_diagram_nodes_diagram ON diagram_nodes(diagram_id);
CREATE INDEX idx_diagram_edges_diagram ON diagram_edges(diagram_id);
