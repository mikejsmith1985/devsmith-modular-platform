// Package search provides advanced filtering and search functionality for log entries.
// This package implements a full query parsing engine with support for:
// - Field-specific searches (e.g., message:error, service:portal)
// - Boolean operators (AND, OR, NOT)
// - Regex patterns with safety validation
// - Field aliases for user convenience
// - SQL WHERE clause generation for PostgreSQL
// - Query optimization and validation
package search

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryParser handles parsing of search queries with boolean operators and field filters.
// It validates query syntax and generates optimized SQL WHERE clauses.
type QueryParser struct {
	supportedFields map[string]string // field aliases mapping (e.g., msg→message)
}

// NewQueryParser creates a new query parser instance with field aliases pre-configured.
// Supported field aliases:
//   - message, msg → message
//   - service, svc → service
//   - level, lvl → level
//   - tags, tag → tags
func NewQueryParser() *QueryParser {
	return &QueryParser{
		supportedFields: map[string]string{
			"message": "message",
			"msg":     "message",
			"service": "service",
			"svc":     "service",
			"level":   "level",
			"lvl":     "level",
			"tags":    "tags",
			"tag":     "tags",
		},
	}
}

// Parse parses a query string into a Query structure without validation.
// This method handles:
// - Empty queries
// - Regex patterns (e.g., /pattern/)
// - Boolean operators (AND, OR, NOT)
// - Field:value syntax
func (p *QueryParser) Parse(queryString string) *Query {
	if queryString == "" {
		return &Query{
			Text:   "",
			Fields: make(map[string]string),
		}
	}

	query := &Query{
		Text:   queryString,
		Fields: make(map[string]string),
	}

	// Check if it's a regex pattern (enclosed in /)
	if strings.HasPrefix(queryString, "/") && strings.LastIndex(queryString, "/") > 0 {
		lastSlash := strings.LastIndex(queryString, "/")
		query.IsRegex = true
		query.RegexPattern = queryString[1:lastSlash]
		return query
	}

	// Parse field:value pairs and boolean operators
	p.parseFieldsAndOperators(queryString, query)

	return query
}

// parseFieldsAndOperators parses field:value pairs and boolean operators from query string.
// Handles NOT operator prefix and boolean operator combinations.
func (p *QueryParser) parseFieldsAndOperators(queryString string, query *Query) {
	// Handle NOT operator prefix
	if strings.HasPrefix(strings.TrimSpace(queryString), "NOT ") {
		query.IsNegated = true
		queryString = strings.TrimSpace(queryString[4:])
	}

	// Check for boolean operators (AND/OR)
	if strings.Contains(queryString, " AND ") || strings.Contains(queryString, " OR ") {
		query.BooleanOp = p.parseBooleanExpression(queryString)
		return
	}

	// Parse field:value pairs
	p.parseFields(queryString, query)
}

// parseBooleanExpression parses AND/OR operators with proper precedence (OR before AND).
func (p *QueryParser) parseBooleanExpression(queryString string) *BooleanOp {
	// Try OR first (lower precedence)
	if strings.Contains(queryString, " OR ") {
		parts := strings.Split(queryString, " OR ")
		conditions := make([]interface{}, 0)
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			q := p.Parse(trimmed)
			conditions = append(conditions, q)
		}
		return &BooleanOp{
			Operator:   "OR",
			Conditions: conditions,
		}
	}

	// Then try AND (higher precedence)
	if strings.Contains(queryString, " AND ") {
		parts := strings.Split(queryString, " AND ")
		conditions := make([]interface{}, 0)
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			q := p.Parse(trimmed)
			conditions = append(conditions, q)
		}
		return &BooleanOp{
			Operator:   "AND",
			Conditions: conditions,
		}
	}

	return nil
}

// parseFields extracts field:value pairs from query string.
// Handles both quoted values (with spaces) and unquoted values.
func (p *QueryParser) parseFields(queryString string, query *Query) {
	// Handle quoted strings: field:"value with spaces"
	quotePattern := regexp.MustCompile(`(\w+):"([^"]*)"`)
	matches := quotePattern.FindAllStringSubmatchIndex(queryString, -1)

	for _, match := range quotePattern.FindAllStringSubmatch(queryString, -1) {
		field := p.resolveFieldAlias(match[1])
		if field != "" {
			value := match[2]
			query.Fields[field] = value
		}
	}

	// Remove quoted sections from remaining string to avoid matching them again
	remaining := queryString
	for _, match := range matches {
		start, end := match[0], match[1]
		remaining = remaining[:start] + " " + remaining[end:]
	}

	// Handle unquoted: field:value
	unquotedPattern := regexp.MustCompile(`(\w+):(\S+)`)
	for _, match := range unquotedPattern.FindAllStringSubmatch(remaining, -1) {
		field := p.resolveFieldAlias(match[1])
		if field != "" {
			value := match[2]
			// Unescape colons
			value = strings.ReplaceAll(value, `\:`, ":")
			// Skip if value is empty or just a colon
			if value != "" && value != ":" {
				query.Fields[field] = value
			}
		}
	}

	// If no fields found, treat as text search on message
	if len(query.Fields) == 0 {
		text := strings.TrimSpace(queryString)
		if text != "" && !strings.HasPrefix(text, "/") {
			query.Fields["message"] = text
		}
	}
}

// resolveFieldAlias resolves field name aliases to canonical names.
// For example: "msg" → "message", "svc" → "service"
func (p *QueryParser) resolveFieldAlias(alias string) string {
	if canonical, ok := p.supportedFields[strings.ToLower(alias)]; ok {
		return canonical
	}
	return ""
}

// ParseAndValidate parses and validates a query string.
// Returns error if query exceeds size limits or has syntax errors.
func (p *QueryParser) ParseAndValidate(queryString string) (*Query, error) {
	if queryString == "" {
		return &Query{Fields: make(map[string]string)}, nil
	}

	// Check for performance limits
	if len(queryString) > 10000 {
		return nil, fmt.Errorf("query string too long: %d > 10000 characters", len(queryString))
	}

	query := p.Parse(queryString)

	// Validate regex if present
	if query.IsRegex {
		if err := p.ValidateRegex(query.RegexPattern); err != nil {
			return nil, err
		}
	}

	// Validate syntax
	if err := p.validateSyntax(queryString); err != nil {
		return nil, err
	}

	return query, nil
}

// validateSyntax performs basic syntax validation on the query string.
// Checks for: unmatched quotes, parentheses, unclosed regex, dangling operators, missing field values.
func (p *QueryParser) validateSyntax(queryString string) error {
	trimmed := strings.TrimSpace(queryString)

	// Check for unmatched quotes
	if strings.Count(queryString, `"`)%2 != 0 {
		return fmt.Errorf("unmatched quotes in query")
	}

	// Check for unmatched parentheses
	openCount := strings.Count(queryString, "(")
	closeCount := strings.Count(queryString, ")")
	if openCount != closeCount {
		return fmt.Errorf("unmatched parentheses in query")
	}

	// Check for unclosed regex
	if strings.HasPrefix(queryString, "/") && !strings.Contains(queryString[1:], "/") {
		return fmt.Errorf("unclosed regex pattern")
	}

	// Check for dangling operators at end
	if strings.HasSuffix(trimmed, " AND") || strings.HasSuffix(trimmed, " OR") {
		return fmt.Errorf("query ends with operator")
	}

	// Check for operators at start (except NOT)
	if strings.HasPrefix(trimmed, "AND ") || strings.HasPrefix(trimmed, "OR ") {
		return fmt.Errorf("query starts with AND/OR operator")
	}

	// Check for field: with no value (e.g., "message:" or "service:")
	fieldNoValuePattern := regexp.MustCompile(`(\w+):\s*($|\s)`)
	if fieldNoValuePattern.MatchString(queryString) {
		return fmt.Errorf("field has no value")
	}

	return nil
}

// ValidateRegex checks if a regex pattern is safe from catastrophic backtracking.
// Also verifies the pattern compiles successfully.
func (p *QueryParser) ValidateRegex(pattern string) error {
	// Check for catastrophic backtracking patterns
	dangerousPatterns := []string{
		`(a+)+`,
		`(a*)*`,
		`(a+)*`,
		`(a|a)+`,
		`(a|ab)+`,
	}

	for _, dangerous := range dangerousPatterns {
		if strings.Contains(pattern, dangerous) {
			return fmt.Errorf("regex pattern contains catastrophic backtracking risk: %s", dangerous)
		}
	}

	// Try to compile to ensure it's valid Go regex
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

// GetSQLCondition generates a SQL WHERE clause and parameters from a parsed query.
// Returns the WHERE clause string, parameters for parameterized queries, and any error.
//
//nolint:gocritic // return values are self-explanatory (sql, params, error)
func (p *QueryParser) GetSQLCondition(query *Query) (string, []interface{}, error) {
	if query == nil {
		return "1=1", []interface{}{}, nil
	}

	var conditions []string
	var params []interface{}

	// Handle boolean operators
	if query.BooleanOp != nil {
		sql, p, err := p.buildBooleanSQL(query.BooleanOp)
		if err != nil {
			return "", nil, err
		}
		if query.IsNegated {
			sql = "NOT (" + sql + ")"
		}
		return sql, p, nil
	}

	// Handle field conditions
	for field, value := range query.Fields {
		cond := ""
		switch field {
		case "message":
			cond = "message ILIKE $" + fmt.Sprintf("%d", len(params)+1)
			params = append(params, "%"+value+"%")
		case "service":
			cond = "service = $" + fmt.Sprintf("%d", len(params)+1)
			params = append(params, value)
		case "level":
			cond = "level = $" + fmt.Sprintf("%d", len(params)+1)
			params = append(params, value)
		case "tags":
			cond = "$" + fmt.Sprintf("%d", len(params)+1) + " = ANY(tags)"
			params = append(params, value)
		}
		if cond != "" {
			conditions = append(conditions, cond)
		}
	}

	if len(conditions) == 0 {
		return "1=1", []interface{}{}, nil
	}

	sql := strings.Join(conditions, " AND ")
	if query.IsNegated {
		sql = "NOT (" + sql + ")"
	}

	return sql, params, nil
}

// buildBooleanSQL recursively builds SQL for nested boolean operations.
//
//nolint:gocritic // return values are self-explanatory (sql, params, error)
func (p *QueryParser) buildBooleanSQL(boolOp *BooleanOp) (string, []interface{}, error) {
	var parts []string
	var allParams []interface{}

	for _, cond := range boolOp.Conditions {
		if q, ok := cond.(*Query); ok {
			sql, params, err := p.GetSQLCondition(q)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, sql)
			allParams = append(allParams, params...)
		}
	}

	op := " " + boolOp.Operator + " "
	return "(" + strings.Join(parts, op) + ")", allParams, nil
}

// Optimize optimizes a parsed query by removing redundancies.
// Currently removes duplicate conditions in boolean operations.
func (p *QueryParser) Optimize(query *Query) *Query {
	if query == nil {
		return nil
	}

	// Remove duplicate conditions in boolean operations
	if query.BooleanOp != nil && len(query.BooleanOp.Conditions) > 1 {
		// Simple deduplication for identical conditions
		seen := make(map[string]bool)
		unique := make([]interface{}, 0)

		for _, cond := range query.BooleanOp.Conditions {
			key := fmt.Sprintf("%v", cond)
			if !seen[key] {
				unique = append(unique, cond)
				seen[key] = true
			}
		}

		query.BooleanOp.Conditions = unique
	}

	return query
}

// GetSupportedFields returns the list of supported field names for queries.
func (p *QueryParser) GetSupportedFields() []string {
	return []string{"message", "service", "level", "tags"}
}
