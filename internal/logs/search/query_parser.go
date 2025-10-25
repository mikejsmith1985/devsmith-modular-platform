package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	maxQueryLength = 5000
	maxFieldLength = 100
	maxValueLength = 1000
)

var validFields = map[string]bool{
	"service":    true,
	"level":      true,
	"message":    true,
	"created_at": true,
	"id":         true,
}

var validLevels = map[string]bool{
	"debug":   true,
	"info":    true,
	"warn":    true,
	"error":   true,
	"fatal":   true,
}

// Lexer tokenizes the input query string.
type Lexer struct {
	input  string
	pos    int
	tokens []QueryToken
	err    string
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		tokens: []QueryToken{},
	}
}

// Tokenize lexes the input and returns tokens.
func (l *Lexer) Tokenize() ([]QueryToken, error) {
	if len(l.input) == 0 {
		return nil, fmt.Errorf("empty query")
	}

	if len(l.input) > maxQueryLength {
		return nil, fmt.Errorf("query exceeds max length (%d chars)", maxQueryLength)
	}

	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		// Handle parentheses
		if ch == '(' || ch == ')' {
			l.tokens = append(l.tokens, QueryToken{Type: "paren", Value: string(ch)})
			l.pos++
			continue
		}

		// Handle operators and fields
		if isAlpha(ch) {
			l.lexKeywordOrField()
		} else if ch == '"' {
			l.lexQuotedString()
		} else if ch == '/' {
			l.lexRegex()
		} else {
			return nil, fmt.Errorf("unexpected character: %c", ch)
		}
	}

	return l.tokens, nil
}

// lexKeywordOrField lexes keywords (AND, OR, NOT) or field names.
func (l *Lexer) lexKeywordOrField() {
	start := l.pos
	for l.pos < len(l.input) && (isAlphaNumeric(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}

	word := l.input[start:l.pos]
	upper := strings.ToUpper(word)

	// Check for operators
	if upper == "AND" || upper == "OR" || upper == "NOT" {
		l.tokens = append(l.tokens, QueryToken{Type: "operator", Value: upper})
		return
	}

	// Must be a field, check for colon
	l.skipWhitespace()
	if l.pos < len(l.input) && l.input[l.pos] == ':' {
		l.pos++ // consume colon
		l.tokens = append(l.tokens, QueryToken{Type: "field", Value: word})
	} else {
		l.tokens = append(l.tokens, QueryToken{Type: "field", Value: word})
	}
}

// lexQuotedString lexes a quoted string value.
func (l *Lexer) lexQuotedString() {
	l.pos++ // skip opening quote
	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos++ // skip escape char
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		l.tokens = append(l.tokens, QueryToken{Type: "error", Value: "unterminated string"})
		return
	}

	value := l.input[start : l.pos]
	value = strings.ReplaceAll(value, `\"`, `"`)
	l.pos++ // skip closing quote

	if len(value) > maxValueLength {
		l.tokens = append(l.tokens, QueryToken{Type: "error", Value: "value too long"})
		return
	}

	l.tokens = append(l.tokens, QueryToken{Type: "value", Value: value})
}

// lexRegex lexes a regex pattern /.../
func (l *Lexer) lexRegex() {
	l.pos++ // skip opening /
	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '/' {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos++ // skip escape
		}
		l.pos++
	}

	if l.pos >= len(l.input) {
		l.tokens = append(l.tokens, QueryToken{Type: "error", Value: "unterminated regex"})
		return
	}

	pattern := l.input[start : l.pos]
	l.pos++ // skip closing /

	// Check for flags (i, g, m, s)
	for l.pos < len(l.input) && isAlpha(l.input[l.pos]) {
		l.pos++
	}

	// Validate regex
	if _, err := regexp.Compile(pattern); err != nil {
		l.tokens = append(l.tokens, QueryToken{Type: "error", Value: "invalid regex: " + err.Error()})
		return
	}

	l.tokens = append(l.tokens, QueryToken{Type: "regex", Value: pattern})
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

// Parser builds an AST from tokens using recursive descent parsing.
type Parser struct {
	tokens []QueryToken
	pos    int
	nodes  []string // search terms extracted
}

// NewParser creates a new parser for the given tokens.
func NewParser(tokens []QueryToken) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		nodes:  []string{},
	}
}

// Parse parses tokens into an AST.
func (p *Parser) Parse() (*QueryNode, error) {
	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("empty token list")
	}

	// Check for errors in token list
	for _, tok := range p.tokens {
		if tok.Type == "error" {
			return nil, fmt.Errorf("token error: %s", tok.Value)
		}
	}

	node, err := p.parseOR()
	if err != nil {
		return nil, err
	}

	if p.pos < len(p.tokens) {
		return nil, fmt.Errorf("unexpected token: %s", p.tokens[p.pos].Value)
	}

	return node, nil
}

// parseOR parses OR expressions (lowest precedence).
func (p *Parser) parseOR() (*QueryNode, error) {
	left, err := p.parseAND()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) && p.peekOperator() == "OR" {
		p.pos++ // consume OR
		right, err := p.parseAND()
		if err != nil {
			return nil, err
		}
		left = &QueryNode{Type: "OR", Left: left, Right: right}
	}

	return left, nil
}

// parseAND parses AND expressions (medium precedence).
func (p *Parser) parseAND() (*QueryNode, error) {
	left, err := p.parseNOT()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) && p.peekOperator() == "AND" {
		p.pos++ // consume AND
		right, err := p.parseNOT()
		if err != nil {
			return nil, err
		}
		left = &QueryNode{Type: "AND", Left: left, Right: right}
	}

	return left, nil
}

// parseNOT parses NOT expressions (high precedence).
func (p *Parser) parseNOT() (*QueryNode, error) {
	if p.peekOperator() == "NOT" {
		p.pos++ // consume NOT
		node, err := p.parseNOT()
		if err != nil {
			return nil, err
		}
		node.IsNegated = true
		return node, nil
	}

	return p.parsePrimary()
}

// parsePrimary parses primary expressions (field:value, parentheses).
func (p *Parser) parsePrimary() (*QueryNode, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of query")
	}

	tok := p.tokens[p.pos]

	// Handle parenthesized expressions
	if tok.Type == "paren" && tok.Value == "(" {
		p.pos++ // consume (
		node, err := p.parseOR()
		if err != nil {
			return nil, err
		}

		if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != "paren" || p.tokens[p.pos].Value != ")" {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++ // consume )

		return node, nil
	}

	// Handle field:value
	if tok.Type == "field" {
		field := tok.Value
		if !validFields[field] {
			return nil, fmt.Errorf("invalid field: %s", field)
		}

		p.pos++ // consume field

		// Expect value or regex
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("field %s requires a value", field)
		}

		valTok := p.tokens[p.pos]
		if valTok.Type == "value" {
			p.pos++
			p.nodes = append(p.nodes, valTok.Value)

			// Special validation for level field
			if field == "level" && !validLevels[valTok.Value] {
				return nil, fmt.Errorf("invalid level: %s", valTok.Value)
			}

			return &QueryNode{
				Type:  "FIELD",
				Field: field,
				Value: valTok.Value,
			}, nil
		} else if valTok.Type == "regex" {
			p.pos++
			p.nodes = append(p.nodes, valTok.Value)

			return &QueryNode{
				Type:  "REGEX",
				Field: field,
				Value: valTok.Value,
			}, nil
		}

		return nil, fmt.Errorf("expected value after field %s", field)
	}

	return nil, fmt.Errorf("unexpected token: %s", tok.Value)
}

// peekOperator returns the current operator or empty string.
func (p *Parser) peekOperator() string {
	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == "operator" {
		return p.tokens[p.pos].Value
	}
	return ""
}

// ParsedQuery implementation for QueryParser.Parse
func (qp *QueryParser) Parse(query string) *ParsedQuery {
	result := &ParsedQuery{
		Tokens:      []QueryToken{},
		SearchTerms: []string{},
		IsValid:     false,
	}

	// Tokenize
	lexer := NewLexer(query)
	tokens, err := lexer.Tokenize()
	if err != nil {
		result.ErrorMsg = err.Error()
		return result
	}

	result.Tokens = tokens

	// Parse
	parser := NewParser(tokens)
	node, err := parser.Parse()
	if err != nil {
		result.ErrorMsg = err.Error()
		return result
	}

	result.RootNode = node
	result.SearchTerms = parser.nodes
	result.HasRegex = containsRegex(node)
	result.IsValid = true

	return result
}

// containsRegex checks if AST contains regex nodes.
func containsRegex(node *QueryNode) bool {
	if node == nil {
		return false
	}

	if node.Type == "REGEX" {
		return true
	}

	return containsRegex(node.Left) || containsRegex(node.Right)
}

// Helper functions
func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || (ch >= '0' && ch <= '9') || ch == '_'
}

// QueryParser parses search queries into AST.
type QueryParser struct {
}

// NewQueryParser creates a new query parser instance
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}
