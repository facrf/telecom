ALTER TABLE scan_hosts ADD COLUMN manufacturer TEXT NOT NULL DEFAULT '';
ALTER TABLE scan_hosts ADD COLUMN device_type TEXT NOT NULL DEFAULT '';
ALTER TABLE scan_hosts ADD COLUMN fingerprint_json TEXT NOT NULL DEFAULT '{}';
