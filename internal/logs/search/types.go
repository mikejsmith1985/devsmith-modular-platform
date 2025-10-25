package search

// ParsedQuery represents a fully parsed search query.
// nolint:govet // fieldalignment: fields ordered logically for domain model clarity
type ParsedQuery struct {
	Tokens      []QueryToken
	SearchTerms []string
	RootNode    *QueryNode
	ErrorMsg    string
	IsValid     bool
	HasRegex    bool
}

// QueryNode represents a node in the query tree.
type QueryNode struct {
	Left      *QueryNode
	Right     *QueryNode
	Type      string // "AND", "OR", "NOT", "FIELD", "REGEX"
	Field     string // field name (e.g., "message", "service", "level")
	Value     string // value or pattern
	IsNegated bool
}

// QueryToken represents a parsed token from a search query.
type QueryToken struct {
	Type  string // "field", "operator", "value", "paren"
	Value string
}
