package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type Inspection struct {
	Manifest        Manifest
	HasDatabase     bool
	AttachmentCount int
}

func Inspect(path string) (Inspection, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return Inspection{}, fmt.Errorf("abrir backup: %w", err)
	}
	defer archive.Close()
	var result Inspection
	for _, entry := range archive.File {
		if filepath.IsAbs(entry.Name) || strings.Contains(filepath.Clean(entry.Name), "..") {
			return Inspection{}, fmt.Errorf("backup contém caminho inválido")
		}
		switch entry.Name {
		case "manifest.json":
			reader, e := entry.Open()
			if e != nil {
				return Inspection{}, e
			}
			data, e := io.ReadAll(io.LimitReader(reader, 1<<20))
			reader.Close()
			if e != nil {
				return Inspection{}, e
			}
			if e = json.Unmarshal(data, &result.Manifest); e != nil {
				return Inspection{}, fmt.Errorf("manifest inválido")
			}
		case "telecom.sqlite":
			result.HasDatabase = true
		default:
			if strings.HasPrefix(entry.Name, "attachments/") {
				result.AttachmentCount++
			}
		}
	}
	if result.Manifest.Format != Format || result.Manifest.Version != Version {
		return Inspection{}, fmt.Errorf("backup incompatível")
	}
	if !result.HasDatabase {
		return Inspection{}, fmt.Errorf("backup sem banco SQLite")
	}
	return result, nil
}
