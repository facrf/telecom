ALTER TABLE scan_hosts ADD COLUMN mac TEXT NOT NULL DEFAULT '';
ALTER TABLE scan_hosts ADD COLUMN discovery_methods TEXT NOT NULL DEFAULT '[]';

CREATE TABLE oui_vendors (
    prefix TEXT PRIMARY KEY,
    vendor TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'builtin',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO oui_vendors(prefix, vendor) VALUES
('BCAD28', 'Hikvision'),
('3C2DB7', 'Hikvision'),
('A0F3C1', 'Dahua'),
('9C8E99', 'Intelbras'),
('B0A737', 'TP-Link'),
('ACCC8E', 'Axis Communications'),
('000C29', 'VMware'),
('001B21', 'Intelbras');

CREATE INDEX idx_scan_hosts_mac ON scan_hosts(mac);
CREATE INDEX idx_oui_vendors_vendor ON oui_vendors(vendor);
