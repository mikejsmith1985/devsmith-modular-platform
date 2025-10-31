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
	Fields       map[string]string
	BooleanOp    *BooleanOp
	Text         string
	RegexPattern string
	IsRegex      bool
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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	QueryString string
	Description string
	ID          int64
	UserID      int64
}

// SearchHistory represents a search history entry.
//
//nolint:govet // field alignment optimization not worth restructuring
type SearchHistory struct {
	SearchedAt  time.Time
	QueryString string
	ID          int64
	UserID      int64
}

// SearchMetadata contains metadata about a search.
//
//nolint:govet // field alignment optimization not worth restructuring
type SearchMetadata struct {
	CreatedAt   time.Time
	QueryString string
	ID          int64
}
