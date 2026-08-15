package backup

import (
	"encoding/json"
	"fmt"
	"time"
)

const Format = "telecom-backup"
const Version = 1

type Manifest struct {
	Format    string    `json:"format"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

func NewManifest() Manifest {
	return Manifest{Format: Format, Version: Version, CreatedAt: time.Now().UTC()}
}
func ValidateManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("manifest inválido: %w", err)
	}
	if manifest.Format != Format || manifest.Version != Version {
		return Manifest{}, fmt.Errorf("backup incompatível")
	}
	return manifest, nil
}
