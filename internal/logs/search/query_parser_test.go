// Package search provides advanced filtering and search capabilities for logs.
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// QueryToken represents a parsed token from a search query.
type QueryToken struct {
	Type  string // "field", "operator", "value", "paren"
	Value string
}

// ParsedQuery represents a fully parsed search query.
type ParsedQuery struct {
	Tokens      []QueryToken
	RootNode    *QueryNode
	IsValid     bool
	ErrorMsg    string
	HasRegex    bool
	SearchTerms []string
}

// QueryNode represents a node in the query tree.
type QueryNode struct {
	Type      string      // "AND", "OR", "NOT", "FIELD", "REGEX"
	Field     string      // field name (e.g., "message", "service", "level")
	Value     string      // value or pattern
	Left      *QueryNode  // left operand
	Right     *QueryNode  // right operand
	IsNegated bool
}

// TestQueryParser_SimpleFieldValue tests basic "field:value" queries
func TestQueryParser_SimpleFieldValue(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
		wantField string
		wantValue string
	}{
		{
			name:      "simple field:value",
			query:     "service:portal",
			wantValid: true,
			wantField: "service",
			wantValue: "portal",
		},
		{
			name:      "level filter",
			query:     "level:error",
			wantValid: true,
			wantField: "level",
			wantValue: "error",
		},
		{
			name:      "message contains",
			query:     "message:\"database connection failed\"",
			wantValid: true,
			wantField: "message",
			wantValue: "database connection failed",
		},
		{
			name:      "invalid: missing value",
			query:     "service:",
			wantValid: false,
		},
		{
			name:      "invalid: unknown field",
			query:     "unknown:value",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.Equal(t, tt.wantField, result.RootNode.Field)
				assert.Equal(t, tt.wantValue, result.RootNode.Value)
			}
		})
	}
}

// TestQueryParser_AND_Operator tests AND boolean operator
func TestQueryParser_AND_Operator(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
		wantOp    string
	}{
		{
			name:      "two conditions with AND",
			query:     "service:portal AND level:error",
			wantValid: true,
			wantOp:    "AND",
		},
		{
			name:      "three conditions with AND",
			query:     "service:portal AND level:error AND message:\"failed\"",
			wantValid: true,
			wantOp:    "AND",
		},
		{
			name:      "AND operator case insensitive",
			query:     "service:portal and level:error",
			wantValid: true,
			wantOp:    "AND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.Equal(t, tt.wantOp, result.RootNode.Type)
			}
		})
	}
}

// TestQueryParser_OR_Operator tests OR boolean operator
func TestQueryParser_OR_Operator(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
		wantOp    string
	}{
		{
			name:      "two conditions with OR",
			query:     "level:error OR level:warn",
			wantValid: true,
			wantOp:    "OR",
		},
		{
			name:      "three conditions with OR",
			query:     "service:portal OR service:review OR service:analytics",
			wantValid: true,
			wantOp:    "OR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.Equal(t, tt.wantOp, result.RootNode.Type)
			}
		})
	}
}

// TestQueryParser_NOT_Operator tests NOT boolean operator
func TestQueryParser_NOT_Operator(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name       string
		query      string
		wantValid  bool
		wantNegated bool
	}{
		{
			name:        "NOT with single condition",
			query:       "NOT service:analytics",
			wantValid:   true,
			wantNegated: true,
		},
		{
			name:        "NOT with compound condition",
			query:       "NOT (level:debug AND service:portal)",
			wantValid:   true,
			wantNegated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.Equal(t, tt.wantNegated, result.RootNode.IsNegated)
			}
		})
	}
}

// TestQueryParser_Regex tests regex pattern support
func TestQueryParser_Regex(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
		wantRegex bool
	}{
		{
			name:      "regex in message field",
			query:     "message:/connection.*failed/",
			wantValid: true,
			wantRegex: true,
		},
		{
			name:      "regex with flags",
			query:     "message:/ERROR|WARNING/i",
			wantValid: true,
			wantRegex: true,
		},
		{
			name:      "invalid regex",
			query:     "message:/[invalid(/",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.Equal(t, tt.wantRegex, result.HasRegex)
			}
		})
	}
}

// TestQueryParser_ComplexQueries tests complex boolean expressions
func TestQueryParser_ComplexQueries(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
	}{
		{
			name:      "complex: (A AND B) OR (C AND NOT D)",
			query:     "(service:portal AND level:error) OR (service:review AND NOT level:debug)",
			wantValid: true,
		},
		{
			name:      "complex: A AND (B OR C) AND NOT D",
			query:     "service:portal AND (level:error OR level:warn) AND NOT message:\"test\"",
			wantValid: true,
		},
		{
			name:      "invalid: mismatched parentheses",
			query:     "(service:portal AND level:error",
			wantValid: false,
		},
		{
			name:      "invalid: consecutive operators",
			query:     "service:portal AND AND level:error",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
		})
	}
}

// TestQueryParser_ExtractSearchTerms extracts searchable terms from query
func TestQueryParser_ExtractSearchTerms(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name        string
		query       string
		wantValid   bool
		wantTerms   []string
		minTerms    int
	}{
		{
			name:       "single field extract",
			query:      "service:portal",
			wantValid:  true,
			minTerms:   1,
		},
		{
			name:       "multiple fields extract",
			query:      "service:portal AND level:error",
			wantValid:  true,
			minTerms:   2,
		},
		{
			name:       "message with quotes extract",
			query:      "message:\"database connection\"",
			wantValid:  true,
			minTerms:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid)
			if tt.wantValid {
				assert.GreaterOrEqual(t, len(result.SearchTerms), tt.minTerms)
			}
		})
	}
}

// TestQueryParser_EdgeCases tests edge cases
func TestQueryParser_EdgeCases(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		query     string
		wantValid bool
	}{
		{
			name:      "empty query",
			query:     "",
			wantValid: false,
		},
		{
			name:      "whitespace only",
			query:     "   ",
			wantValid: false,
		},
		{
			name:      "very long value",
			query:     "message:\"" + string(make([]byte, 10000)) + "\"",
			wantValid: false, // should have max length
		},
		{
			name:      "special characters in quoted value",
			query:     "message:\"!@#$%^&*()\"",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.wantValid, result.IsValid, tt.query)
		})
	}
}

// TestQueryParser_ValidFields tests allowed field names
func TestQueryParser_ValidFields(t *testing.T) {
	parser := NewQueryParser()

	validFields := []string{"service", "level", "message", "created_at"}
	for _, field := range validFields {
		t.Run("valid field: "+field, func(t *testing.T) {
			query := field + ":test"
			result := parser.Parse(query)
			assert.True(t, result.IsValid, "field %s should be valid", field)
		})
	}

	invalidFields := []string{"invalid", "xyz", "unknown"}
	for _, field := range invalidFields {
		t.Run("invalid field: "+field, func(t *testing.T) {
			query := field + ":test"
			result := parser.Parse(query)
			assert.False(t, result.IsValid, "field %s should be invalid", field)
		})
	}
}
