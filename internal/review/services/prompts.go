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

You MUST respond with ONLY a valid JSON object (no markdown, no explanation text). Use EXACTLY this structure:

{
  "overall_grade": "B",
  "summary": "Found 1 critical, 2 high, 3 medium severity issues affecting security and performance",
  "issues": [
    {
      "severity": "critical",
      "category": "security",
      "file": "main.go",
      "line": 42,
      "code_snippet": "db.Query(query + userInput)",
      "description": "SQL injection vulnerability - user input not parameterized",
      "impact": "Attacker can execute arbitrary SQL commands",
      "fix_suggestion": "Use parameterized query: db.Query(query, userInput)"
    }
  ]
}

REQUIRED FIELDS (all must be present):
- overall_grade: "A", "B", "C", "D", or "F"
- summary: Brief overview of findings
- issues: Array of issue objects

ISSUE OBJECT FIELDS (all required):
- severity: "critical" | "high" | "medium" | "low"
- category: "security" | "performance" | "maintainability" | "reliability" | "testing"
- file: filename or "unknown" if not determinable
- line: line number or 0 if not determinable
- code_snippet: The problematic code
- description: What's wrong
- impact: What harm this causes
- fix_suggestion: How to fix it

GRADING CRITERIA:
A = No critical/high issues, excellent quality
B = Minor issues, good quality
C = Some concerning issues, acceptable
D = Multiple serious issues, needs work
F = Critical issues present, unsafe

IMPORTANT:
- Return ONLY the JSON object (no json code fences, no explanatory text)
- Focus on SECURITY and CORRECTNESS
- If no issues found, return empty issues array
- Be precise and actionable`, code)
}
