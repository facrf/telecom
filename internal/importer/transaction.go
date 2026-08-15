package importer

import (
	"context"
	"database/sql"
	telecomexport "github.com/local/telecom/internal/export"
)

type ApplyFunc func(context.Context, *sql.Tx, telecomexport.Document) error

func Apply(ctx context.Context, db *sql.DB, document telecomexport.Document, apply ApplyFunc) error {
	if err := telecomexport.Validate(document); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err = apply(ctx, tx, document); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
