package export

import "time"

const Format = "telecom"
const SchemaVersion = 1

type Document struct {
	Format        string    `json:"format"`
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	Client        any       `json:"client"`
	Project       any       `json:"project"`
	Devices       []any     `json:"devices"`
	Diagrams      []any     `json:"diagrams"`
	Documents     []any     `json:"documents"`
}

func NewDocument() Document {
	return Document{Format: Format, SchemaVersion: SchemaVersion, GeneratedAt: time.Now().UTC(), Devices: []any{}, Diagrams: []any{}, Documents: []any{}}
}
func Validate(document Document) error {
	if document.Format != Format {
		return ErrInvalidFormat
	}
	if document.SchemaVersion != SchemaVersion {
		return ErrUnsupportedVersion
	}
	return nil
}

type formatError string

func (e formatError) Error() string { return string(e) }

const ErrInvalidFormat = formatError("formato de exportação inválido")
const ErrUnsupportedVersion = formatError("versão de schema não suportada")
