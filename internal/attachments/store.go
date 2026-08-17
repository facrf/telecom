package attachments

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	root    string
	maxSize int64
}

func NewStore(dataDir string, maxSize int64) *Store {
	return &Store{root: filepath.Join(dataDir, "attachments"), maxSize: maxSize}
}
func (s *Store) Save(entityType, filename string, content []byte) (Metadata, string, error) {
	if entityType != "client" && entityType != "project" && entityType != "device" && entityType != "edge" && entityType != "document" && entityType != "technical_visit" {
		return Metadata{}, "", fmt.Errorf("tipo de entidade inválido")
	}
	metadata, err := Validate(filename, content, s.maxSize)
	if err != nil {
		return Metadata{}, "", err
	}
	directory := filepath.Join(s.root, entityType)
	if err = os.MkdirAll(directory, 0o750); err != nil {
		return Metadata{}, "", err
	}
	destination := filepath.Join(directory, metadata.StoredFilename)
	if err = os.WriteFile(destination, content, 0o640); err != nil {
		return Metadata{}, "", err
	}
	hash := sha256.Sum256(content)
	return metadata, fmt.Sprintf("%x", hash[:]), nil
}
