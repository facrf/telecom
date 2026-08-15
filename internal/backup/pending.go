package backup

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const pendingName = "restore.pending.zip"

func QueueRestore(dataDir string, reader io.Reader) (Inspection, error) {
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return Inspection{}, err
	}
	temporary, err := os.CreateTemp(dataDir, "restore-upload-*.zip")
	if err != nil {
		return Inspection{}, err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = io.Copy(temporary, io.LimitReader(reader, 2<<30)); err != nil {
		temporary.Close()
		return Inspection{}, err
	}
	if err = temporary.Close(); err != nil {
		return Inspection{}, err
	}
	inspection, err := Inspect(temporaryPath)
	if err != nil {
		return Inspection{}, err
	}
	pending := filepath.Join(dataDir, pendingName)
	if err = os.Rename(temporaryPath, pending); err != nil {
		return Inspection{}, err
	}
	return inspection, nil
}

func ApplyPending(dataDir string) (bool, error) {
	pending := filepath.Join(dataDir, pendingName)
	if _, err := os.Stat(pending); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	_, staging, err := Stage(pending, dataDir)
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(staging)
	stagedDatabase := filepath.Join(staging, "telecom.sqlite")
	db, err := sql.Open("sqlite", stagedDatabase)
	if err != nil {
		return false, err
	}
	var integrity string
	err = db.QueryRow("PRAGMA integrity_check").Scan(&integrity)
	db.Close()
	if err != nil || integrity != "ok" {
		return false, fmt.Errorf("SQLite do backup inválido")
	}
	suffix := time.Now().UTC().Format("20060102T150405.000000000Z")
	currentDatabase := filepath.Join(dataDir, "telecom.sqlite")
	previousDatabase := filepath.Join(dataDir, "telecom.sqlite.pre-restore-"+suffix)
	currentAttachments := filepath.Join(dataDir, "attachments")
	previousAttachments := filepath.Join(dataDir, "attachments.pre-restore-"+suffix)
	stagedAttachments := filepath.Join(staging, "attachments")
	if err = os.MkdirAll(stagedAttachments, 0o750); err != nil {
		return false, err
	}
	databaseMoved := false
	attachmentsMoved := false
	if _, err = os.Stat(currentDatabase); err == nil {
		if err = os.Rename(currentDatabase, previousDatabase); err != nil {
			return false, err
		}
		databaseMoved = true
	}
	for _, extension := range []string{"-wal", "-shm"} {
		sidecar := currentDatabase + extension
		if _, sideErr := os.Stat(sidecar); sideErr == nil {
			if sideErr = os.Rename(sidecar, previousDatabase+extension); sideErr != nil {
				if databaseMoved {
					_ = os.Rename(previousDatabase, currentDatabase)
				}
				return false, sideErr
			}
		}
	}
	if _, currentErr := os.Stat(currentAttachments); currentErr == nil {
		if err = os.Rename(currentAttachments, previousAttachments); err != nil {
			if databaseMoved {
				_ = os.Rename(previousDatabase, currentDatabase)
			}
			return false, err
		}
		attachmentsMoved = true
	}
	rollback := func() {
		_ = os.Remove(currentDatabase)
		_ = os.RemoveAll(currentAttachments)
		if databaseMoved {
			_ = os.Rename(previousDatabase, currentDatabase)
		}
		if attachmentsMoved {
			_ = os.Rename(previousAttachments, currentAttachments)
		}
		for _, extension := range []string{"-wal", "-shm"} {
			_ = os.Rename(previousDatabase+extension, currentDatabase+extension)
		}
	}
	if err = os.Rename(stagedDatabase, currentDatabase); err != nil {
		rollback()
		return false, err
	}
	if err = os.Rename(stagedAttachments, currentAttachments); err != nil {
		rollback()
		return false, err
	}
	if err = os.Remove(pending); err != nil {
		return false, err
	}
	return true, nil
}
