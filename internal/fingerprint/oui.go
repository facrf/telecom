package fingerprint

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

func ImportOUI(ctx context.Context, db *sql.DB, reader io.Reader) (int, error) {
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return 0, fmt.Errorf("CSV OUI inválido: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	count := 0
	for index, record := range records {
		if len(record) < 2 {
			return 0, fmt.Errorf("linha %d deve conter prefixo e fabricante", index+1)
		}
		prefix := normalizePrefix(record[0])
		vendor := strings.TrimSpace(record[1])
		if index == 0 && (strings.EqualFold(prefix, "PREFIX") || strings.EqualFold(strings.TrimSpace(record[0]), "prefix")) {
			continue
		}
		if len(prefix) != 6 || vendor == "" {
			return 0, fmt.Errorf("linha %d inválida", index+1)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO oui_vendors(prefix,vendor,source,updated_at)VALUES(?,?,'import',CURRENT_TIMESTAMP) ON CONFLICT(prefix) DO UPDATE SET vendor=excluded.vendor,source='import',updated_at=CURRENT_TIMESTAMP", prefix, vendor); err != nil {
			return 0, err
		}
		count++
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}
func normalizePrefix(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	replacer := strings.NewReplacer(":", "", "-", "", ".", "")
	return replacer.Replace(value)
}
