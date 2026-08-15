package attachments

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

type Metadata struct {
	OriginalFilename string
	StoredFilename   string
	MIMEType         string
	Size             int64
}

func Validate(filename string, content []byte, maxSize int64) (Metadata, error) {
	if filename == "" || filepath.Base(filename) != filename || strings.ContainsAny(filename, "\\/") {
		return Metadata{}, fmt.Errorf("nome de arquivo inválido")
	}
	if int64(len(content)) > maxSize {
		return Metadata{}, fmt.Errorf("arquivo excede o limite permitido")
	}
	mime := http.DetectContentType(content)
	allowed := map[string]bool{"application/pdf": true, "image/png": true, "image/jpeg": true, "image/webp": true, "text/plain; charset=utf-8": true, "text/csv; charset=utf-8": true, "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true}
	if !allowed[mime] {
		return Metadata{}, fmt.Errorf("tipo de arquivo não permitido: %s", mime)
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return Metadata{}, err
	}
	return Metadata{OriginalFilename: filename, StoredFilename: hex.EncodeToString(token) + filepath.Ext(filename), MIMEType: mime, Size: int64(len(content))}, nil
}
