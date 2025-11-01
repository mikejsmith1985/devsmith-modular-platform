// Package review_services provides analysis services for different reading modes.
// This file contains prompt templates for guiding AI analysis.
package review_services

import "fmt"

// BuildPreviewPrompt creates a prompt for Preview Mode analysis
// Preview: Quick structural assessment (2-3 minutes)
func BuildPreviewPrompt(code string) string {
	return fmt.Sprintf(`Analyze this code in PREVIEW mode - provide a quick structural overview.

CODE:
%s

Return a JSON object with ONLY these fields (no extra fields):
{
  "file_tree": ["main.go", "handler.go", "models/user.go"],
  "bounded_contexts": ["authentication", "user management"],
  "tech_stack": ["Go", "PostgreSQL", "Gin"],
  "architecture_pattern": "layered",
  "entry_points": ["main()", "NewServer()"],
  "external_dependencies": ["PostgreSQL", "Redis"],
  "summary": "Brief 1-2 sentence summary of what this code does"
}

IMPORTANT:
- Be concise and high-level
- Don't dive into implementation details
- Focus on structure and context
- Return ONLY valid JSON, no markdown or explanation`, code)
}

// BuildSkimPrompt creates a prompt for Skim Mode analysis
// Skim: Abstract overview without implementation details (5-7 minutes)
func BuildSkimPrompt(code string) string {
	return fmt.Sprintf(`Analyze this code in SKIM mode - extract key abstractions without diving into implementations.

CODE:
%s

Return a JSON object with ONLY these fields:
{
  "functions": ["GetUser(id)", "CreateUser(data)", "DeleteUser(id)"],
  "interfaces": ["UserRepository", "AuthService"],
  "data_models": ["User{ID, Name, Email}", "Request{}"],
  "workflows": ["User creation flow: ValidateInput -> StoreDB -> ReturnUser"],
  "summary": "What does this code provide at a high level?"
}

IMPORTANT:
- List function signatures, not full implementations
- Identify key interfaces and abstractions
- Show data structures (not full definitions)
- Don't explain line-by-line logic
- Return ONLY valid JSON`, code)
}

// BuildScanPrompt creates a prompt for Scan Mode analysis
// Scan: Targeted pattern search (3-5 minutes)
func BuildScanPrompt(code, query string) string {
	return fmt.Sprintf(`Analyze this code for Scan mode - find specific patterns/information matching the query.

QUERY: "%s"

CODE:
%s

Return a JSON object with ONLY these fields:
{
  "query": "%s",
  "matches": [
    {
      "file": "handler.go",
      "line": 42,
      "code_snippet": "func HandleLogin(c *gin.Context)",
      "relevance_score": 0.95,
      "reason": "Directly handles authentication"
    }
  ],
  "total_matches": 3,
  "summary": "Found 3 matches for query '%s' in the codebase"
}

IMPORTANT:
- Find code related to the query
- Provide line numbers where possible
- Relevance score 0.0-1.0 (1.0 = perfect match)
- Return ONLY valid JSON`, query, code, query, query)
}

// BuildDetailedPrompt creates a prompt for Detailed Mode analysis
// Detailed: Line-by-line understanding (10-15 minutes)
func BuildDetailedPrompt(code, filename string) string {
	return fmt.Sprintf(`Analyze this code in DETAILED mode - provide line-by-line algorithm explanation.

FILE: %s
CODE:
%s

Return a JSON object with ONLY these fields:
{
  "file": "%s",
  "line_by_line": [
    {
      "line_number": 1,
      "code": "package main",
      "explanation": "Declares this as the main executable package"
    },
    {
      "line_number": 2,
      "code": "import \"fmt\"",
      "explanation": "Imports fmt package for formatted output"
    }
  ],
  "summary": "This function implements X algorithm by doing Y then Z"
}

IMPORTANT:
- Explain EVERY significant line
- Explain the purpose and effect of each line
- Show how lines work together (data flow)
- Explain key logic and conditionals
- Return ONLY valid JSON`, filename, code, filename)
}

// BuildCriticalPrompt creates a prompt for Critical Mode analysis
// Critical: Quality evaluation focusing on issues and improvements
func BuildCriticalPrompt(code string) string {
	return fmt.Sprintf(`Analyze this code in CRITICAL mode - identify quality issues and provide improvement suggestions.

CODE:
%s

Return a JSON object with ONLY these fields:
{
  "issues": [
    {
      "severity": "critical",
      "category": "security",
      "line": 42,
      "code_snippet": "db.Query(query + userInput)",
      "description": "SQL injection vulnerability - user input not parameterized",
      "impact": "Attacker can execute arbitrary SQL",
      "suggestion": "Use parameterized query: db.Query(query, userInput)"
    }
  ],
  "quality_score": 65,
  "summary": "Found 1 critical, 2 high, 3 medium severity issues"
}

SEVERITY LEVELS: critical | high | medium | low
CATEGORIES: security | performance | maintainability | reliability | testing

IMPORTANT:
- Identify actual problems, not style preferences
- Prioritize SECURITY and CORRECTNESS issues
- Provide actionable fix suggestions
- Score 0-100 (100 = perfect, 0 = unusable)
- Return ONLY valid JSON`, code)
}
