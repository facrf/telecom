package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Stage(path, destination string) (Inspection, string, error) {
	inspection, err := Inspect(path)
	if err != nil {
		return Inspection{}, "", err
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		return Inspection{}, "", err
	}
	defer archive.Close()
	staging, err := os.MkdirTemp(destination, "telecom-restore-")
	if err != nil {
		return Inspection{}, "", err
	}
	for _, entry := range archive.File {
		clean := filepath.Clean(entry.Name)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			os.RemoveAll(staging)
			return Inspection{}, "", fmt.Errorf("caminho inválido")
		}
		target := filepath.Join(staging, clean)
		if !strings.HasPrefix(target, staging+string(os.PathSeparator)) {
			os.RemoveAll(staging)
			return Inspection{}, "", fmt.Errorf("caminho inválido")
		}
		if entry.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0o750); err != nil {
				os.RemoveAll(staging)
				return Inspection{}, "", err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			os.RemoveAll(staging)
			return Inspection{}, "", err
		}
		in, err := entry.Open()
		if err != nil {
			os.RemoveAll(staging)
			return Inspection{}, "", err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
		if err == nil {
			_, err = io.Copy(out, in)
		}
		in.Close()
		out.Close()
		if err != nil {
			os.RemoveAll(staging)
			return Inspection{}, "", err
		}
	}
	if _, err = os.Stat(filepath.Join(staging, "telecom.sqlite")); err != nil {
		os.RemoveAll(staging)
		return Inspection{}, "", fmt.Errorf("banco não extraído")
	}
	return inspection, staging, nil
}
