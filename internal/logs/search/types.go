// Package search provides advanced filtering and search functionality for log entries.
//
//nolint:revive // Type names SearchHistory, SearchMetadata, and Search* are intentional for public API clarity
package search

import (
	"time"
)

// Query represents a parsed search query with filters and operators.
// nolint:govet // field alignment optimization not worth restructuring
type Query struct {
	Text         string
	IsRegex      bool
	RegexPattern string
	Fields       map[string]string
	BooleanOp    *BooleanOp
	IsNegated    bool
}

// BooleanOp represents a boolean operation (AND, OR).
type BooleanOp struct {
	Operator   string
	Conditions []interface{}
}

// SavedSearch represents a saved search query for a user.
// nolint:govet // field alignment optimization not worth restructuring
type SavedSearch struct {
	ID          int64
	UserID      int64
	Name        string
	QueryString string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SearchHistory represents a search history entry.
//nolint:govet // field alignment optimization not worth restructuring
type SearchHistory struct {
	ID          int64
	UserID      int64
	QueryString string
	SearchedAt  time.Time
}

// SearchMetadata contains metadata about a search.
//nolint:govet // field alignment optimization not worth restructuring
type SearchMetadata struct {
	ID          int64
	QueryString string
	CreatedAt   time.Time
}
