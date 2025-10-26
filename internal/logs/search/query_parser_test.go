// Package search provides advanced filtering and search functionality for log entries.
// RED Phase: Test-driven development with comprehensive failing tests.
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryParser_ParseSimpleQuery tests parsing of simple text queries
func TestQueryParser_ParseSimpleQuery(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("error")

	require.NotNil(t, query)
	assert.Equal(t, "error", query.Text)
	assert.False(t, query.IsRegex)
	assert.Empty(t, query.Fields)
}

// TestQueryParser_ParseFieldSpecificQuery tests parsing queries with field selectors
// Example: message:error service:portal
func TestQueryParser_ParseFieldSpecificQuery(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("message:error service:portal")

	require.NotNil(t, query)
	assert.Equal(t, "error", query.Fields["message"])
	assert.Equal(t, "portal", query.Fields["service"])
}

// TestQueryParser_ParseBooleanAND tests parsing AND operator
// Example: message:error AND service:portal
func TestQueryParser_ParseBooleanAND(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("message:error AND service:portal")

	require.NotNil(t, query)
	assert.NotNil(t, query.BooleanOp)
	assert.Equal(t, "AND", query.BooleanOp.Operator)
	assert.Len(t, query.BooleanOp.Conditions, 2)
}

// TestQueryParser_ParseBooleanOR tests parsing OR operator
// Example: level:error OR level:warn
func TestQueryParser_ParseBooleanOR(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("level:error OR level:warn")

	require.NotNil(t, query)
	assert.NotNil(t, query.BooleanOp)
	assert.Equal(t, "OR", query.BooleanOp.Operator)
	assert.Len(t, query.BooleanOp.Conditions, 2)
}

// TestQueryParser_ParseBooleanNOT tests parsing NOT operator
// Example: NOT level:debug
func TestQueryParser_ParseBooleanNOT(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("NOT level:debug")

	require.NotNil(t, query)
	assert.True(t, query.IsNegated)
	assert.Equal(t, "debug", query.Fields["level"])
}

// TestQueryParser_ParseComplexBoolean tests complex boolean expressions
// Example: (message:error AND service:portal) OR message:timeout
func TestQueryParser_ParseComplexBoolean(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("(message:error AND service:portal) OR message:timeout")

	require.NotNil(t, query)
	assert.NotNil(t, query.BooleanOp)
	// Should have 2 top-level conditions joined by OR
	assert.Equal(t, "OR", query.BooleanOp.Operator)
	assert.Len(t, query.BooleanOp.Conditions, 2)
}

// TestQueryParser_ParseRegexPattern tests regex pattern parsing
// Example: /error: \d+/
func TestQueryParser_ParseRegexPattern(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("/error: \\d+/")

	require.NotNil(t, query)
	assert.True(t, query.IsRegex)
	assert.NotEmpty(t, query.RegexPattern)
}

// TestQueryParser_ValidateRegex tests regex validation for safety
// Should reject regex patterns that could cause catastrophic backtracking
func TestQueryParser_ValidateRegex(t *testing.T) {
	parser := NewQueryParser()

	// Simple safe regex should pass
	query := parser.Parse("/error/")
	require.NotNil(t, query)
	assert.NoError(t, parser.ValidateRegex(query.RegexPattern))

	// Catastrophic backtracking pattern should fail
	backtrackQuery := parser.Parse("/(a+)+/")
	err := parser.ValidateRegex(backtrackQuery.RegexPattern)
	assert.Error(t, err, "Should reject regex with catastrophic backtracking")
}

// TestQueryParser_ParseEmptyQuery tests handling of empty queries
func TestQueryParser_ParseEmptyQuery(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("")

	require.NotNil(t, query)
	assert.Empty(t, query.Text)
	assert.Empty(t, query.Fields)
}

// TestQueryParser_ParseQuotedStrings tests parsing quoted strings with spaces
// Example: message:"database connection failed"
func TestQueryParser_ParseQuotedStrings(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse(`message:"database connection failed"`)

	require.NotNil(t, query)
	assert.Equal(t, "database connection failed", query.Fields["message"])
}

// TestQueryParser_EscapeSpecialChars tests handling of escaped special characters
// Example: level:error\:info
func TestQueryParser_EscapeSpecialChars(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse(`message:error\:info`)

	require.NotNil(t, query)
	assert.Equal(t, "error:info", query.Fields["message"])
}

// TestQueryParser_MultipleOperators tests operator precedence
// Example: message:error AND level:warn OR service:portal
// Should respect AND before OR precedence
func TestQueryParser_MultipleOperators(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("message:error AND level:warn OR service:portal")

	require.NotNil(t, query)
	// Should create nested boolean structure respecting precedence
	assert.NotNil(t, query.BooleanOp)
}

// TestQueryParser_InvalidSyntax tests rejection of invalid syntax
func TestQueryParser_InvalidSyntax(t *testing.T) {
	parser := NewQueryParser()

	invalidQueries := []string{
		"message:",        // Missing value
		"AND error",       // AND without left operand
		"(message:error",  // Unmatched parenthesis
		"/unclosed regex", // Unclosed regex
	}

	for _, q := range invalidQueries {
		query, err := parser.ParseAndValidate(q)
		assert.Error(t, err, "Should reject invalid query: %s", q)
		assert.Nil(t, query)
	}
}

// TestQueryParser_ParseAndValidate tests end-to-end query validation
func TestQueryParser_ParseAndValidate(t *testing.T) {
	parser := NewQueryParser()
	query, err := parser.ParseAndValidate("message:error AND level:warn")

	require.NoError(t, err)
	require.NotNil(t, query)
	assert.NotNil(t, query.BooleanOp)
}

// TestQueryParser_GetSQLCondition tests conversion to SQL WHERE clause
// Should generate parameterized SQL from parsed query
func TestQueryParser_GetSQLCondition(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("message:error")

	sqlCond, params, err := parser.GetSQLCondition(query)

	require.NoError(t, err)
	assert.NotEmpty(t, sqlCond)
	assert.NotEmpty(t, params)
	// SQL should not contain user input directly (parameterized)
	assert.NotContains(t, sqlCond, "error")
}

// TestQueryParser_GetSQLConditionComplex tests SQL generation for complex queries
func TestQueryParser_GetSQLConditionComplex(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("(message:error AND service:portal) OR level:critical")

	sqlCond, params, err := parser.GetSQLCondition(query)

	require.NoError(t, err)
	assert.NotEmpty(t, sqlCond)
	assert.Greater(t, len(params), 1, "Complex query should have multiple parameters")
	// Should contain SQL operators like OR/AND
	assert.Regexp(t, "(OR|AND)", sqlCond)
}

// TestQueryParser_OptimizeQuery tests query optimization
// Should remove redundant conditions or combine similar ones
func TestQueryParser_OptimizeQuery(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("message:error OR message:error")

	optimized := parser.Optimize(query)
	require.NotNil(t, optimized)
	// Optimized query should have eliminated redundancy
}

// TestQueryParser_CaseSensitivity tests case-insensitive search
func TestQueryParser_CaseSensitivity(t *testing.T) {
	parser := NewQueryParser()
	query1 := parser.Parse("MESSAGE:ERROR")
	query2 := parser.Parse("message:error")

	// Both should parse successfully (case-insensitive)
	require.NotNil(t, query1)
	require.NotNil(t, query2)
	// Should match same logs regardless of case
}

// TestQueryParser_SupportedFields tests that all valid fields are recognized
func TestQueryParser_SupportedFields(t *testing.T) {
	parser := NewQueryParser()
	supportedFields := parser.GetSupportedFields()

	assert.Contains(t, supportedFields, "message")
	assert.Contains(t, supportedFields, "service")
	assert.Contains(t, supportedFields, "level")
	assert.Contains(t, supportedFields, "tags")
}

// TestQueryParser_FieldAliases tests field name aliases
// Example: msg should resolve to message
func TestQueryParser_FieldAliases(t *testing.T) {
	parser := NewQueryParser()
	query := parser.Parse("msg:error")

	require.NotNil(t, query)
	// Should resolve msg to message field
	assert.Equal(t, "error", query.Fields["message"])
}

// TestQueryParser_PerformanceLimit tests query has reasonable size
// Should reject extremely long queries to prevent DoS
func TestQueryParser_PerformanceLimit(t *testing.T) {
	parser := NewQueryParser()

	// Create a very long query (>10KB)
	longQuery := "message:"
	for i := 0; i < 10000; i++ {
		longQuery += "a"
	}

	query, err := parser.ParseAndValidate(longQuery)
	assert.Error(t, err, "Should reject excessively long queries")
	assert.Nil(t, query)
}
