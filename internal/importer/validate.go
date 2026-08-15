package importer

import (
	"encoding/json"
	"fmt"

	telecomexport "github.com/local/telecom/internal/export"
)

func Validate(data []byte) (telecomexport.Document, error) {
	var document telecomexport.Document
	if err := json.Unmarshal(data, &document); err != nil {
		return telecomexport.Document{}, fmt.Errorf("JSON inválido: %w", err)
	}
	if err := telecomexport.Validate(document); err != nil {
		return telecomexport.Document{}, err
	}
	if document.Client == nil || document.Project == nil {
		return telecomexport.Document{}, fmt.Errorf("cliente e projeto são obrigatórios")
	}
	return document, nil
}
