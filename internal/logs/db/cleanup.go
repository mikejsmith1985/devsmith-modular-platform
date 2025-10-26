package db

import (
	"database/sql"
	"log"
)

// closeRows safely closes SQL rows and logs any errors.
// This helper prevents nolint directives from cluttering code.
// Errors from row closure are non-critical but logged for observability.
func closeRows(rows *sql.Rows) {
	if rows == nil {
		return
	}
	if err := rows.Close(); err != nil {
		log.Printf("warning: failed to close database rows: %v", err)
	}
}
