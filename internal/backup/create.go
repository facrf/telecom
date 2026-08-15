package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Service struct{ db *sql.DB }

func New(db *sql.DB) *Service { return &Service{db: db} }
func (s *Service) Create(ctx context.Context, output, attachmentsDir string) error {
	temporary, err := os.CreateTemp(filepath.Dir(output), "telecom-backup-*.sqlite")
	if err != nil {
		return err
	}
	databasePath := temporary.Name()
	temporary.Close()
	defer os.Remove(databasePath)
	quoted := strings.ReplaceAll(databasePath, "'", "''")
	if _, err = s.db.ExecContext(ctx, "VACUUM INTO '"+quoted+"'"); err != nil {
		return fmt.Errorf("backup SQLite: %w", err)
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()
	manifest, err := json.Marshal(NewManifest())
	if err != nil {
		return err
	}
	if err = writeEntry(writer, "manifest.json", strings.NewReader(string(manifest))); err != nil {
		return err
	}
	source, err := os.Open(databasePath)
	if err != nil {
		return err
	}
	defer source.Close()
	if err = writeEntry(writer, "telecom.sqlite", source); err != nil {
		return err
	}
	if attachmentsDir == "" {
		return nil
	}
	return filepath.Walk(attachmentsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(attachmentsDir, path)
		if err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		return writeEntry(writer, filepath.ToSlash(filepath.Join("attachments", relative)), input)
	})
}
func writeEntry(writer *zip.Writer, name string, reader io.Reader) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, reader)
	return err
}
