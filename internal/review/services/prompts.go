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
  "file_tree": [
    {
      "name": "main.go",
      "type": "file",
      "path": "main.go",
      "description": "Entry point with main function"
    },
    {
      "name": "handler.go",
      "type": "file",
      "path": "handler.go",
      "description": "HTTP request handlers"
    }
  ],
  "bounded_contexts": ["authentication", "user management"],
  "tech_stack": ["Go", "PostgreSQL", "Gin"],
  "architecture_style": "layered",
  "entry_points": ["main()", "NewServer()"],
  "external_dependencies": ["PostgreSQL", "Redis"],
  "stats": {
    "total_files": 5,
    "total_lines": 500,
    "total_functions": 20,
    "total_interfaces": 3,
    "total_tests": 15
  },
  "summary": "Brief 1-2 sentence summary of what this code does"
}

IMPORTANT:
- Be concise and high-level
- Don't dive into implementation details
- Focus on structure and context
- file_tree items MUST have name, type, path, and description fields
- Return ONLY valid JSON, no markdown or explanation`, code)
}

// BuildSkimPrompt creates a prompt for Skim Mode analysis
// Skim: Abstract overview without implementation details (5-7 minutes)
func BuildSkimPrompt(code string) string {
	return fmt.Sprintf(`Analyze this code in SKIM mode - extract key abstractions without diving into implementations.

CODE:
%s

Return a JSON object with ONLY these fields matching this exact structure:
{
  "functions": [
    {
      "name": "GetUser",
      "signature": "func GetUser(id int) (*User, error)",
      "description": "Retrieves user by ID from database"
    }
  ],
  "interfaces": [
    {
      "purpose": "Manages user data persistence",
      "methods": ["GetUser(id int)", "CreateUser(user *User)", "DeleteUser(id int)"]
    }
  ],
  "data_models": [
    {
      "purpose": "Represents a user account",
      "fields": ["ID int", "Name string", "Email string", "CreatedAt time.Time"]
    }
  ],
  "workflows": [
    {
      "name": "User creation flow",
      "steps": ["1. Validate input data", "2. Check if user exists", "3. Store in database", "4. Return created user"]
    }
  ],
  "summary": "Brief overview of what this code provides at a high level"
}

IMPORTANT:
- For functions: include name, full signature, and brief description
- For interfaces: describe purpose and list method signatures
- For data_models: explain purpose and list field names with types
- For workflows: name the workflow and list sequential steps
- Don't explain implementation details or line-by-line logic
- Return ONLY valid JSON matching the structure above`, code)
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
	return fmt.Sprintf(`Analyze this code in DETAILED mode - provide comprehensive line-by-line explanation with algorithm analysis.

FILE: %s
CODE:
%s

Return a JSON object with these fields in this EXACT structure:
{
  "line_explanations": [
    {
      "line_number": 1,
      "code": "func BinarySearch(arr []int, target int) int {",
      "explanation": "Function declaration that takes a sorted array and target value, returns index or -1",
      "variables": "arr: input array, target: value to find"
    },
    {
      "line_number": 2,
      "code": "left, right := 0, len(arr)-1",
      "explanation": "Initialize pointers to search boundaries",
      "variables": "left=0, right=len(arr)-1"
    }
  ],
  "algorithm_summary": "This implements binary search algorithm, which efficiently finds a target value by repeatedly dividing the search space in half",
  "complexity": "Time: O(log n) - halves search space each iteration. Space: O(1) - only uses constant extra space",
  "edge_cases": [
    "Empty array: returns -1",
    "Target not found: returns -1",
    "Array with one element: checks element and returns 0 or -1"
  ],
  "variable_tracking": [
    {
      "line_number": 2,
      "variables": {
        "left": "0",
        "right": "len(arr)-1"
      }
    },
    {
      "line_number": 5,
      "variables": {
        "left": "0",
        "right": "4",
        "mid": "2"
      }
    }
  ],
  "control_flow": [
    {
      "type": "loop",
      "line_number": 3,
      "description": "While loop continues until left exceeds right",
      "children": ["condition_check", "binary_search_logic"]
    },
    {
      "type": "if",
      "line_number": 5,
      "description": "Check if middle element matches target",
      "children": ["return_mid", "check_less_than", "check_greater_than"]
    }
  ],
  "summary": "Binary search implementation with O(log n) complexity"
}

CRITICAL INSTRUCTIONS:
1. line_explanations is the PRIMARY OUTPUT - explain EVERY significant line
2. For each line, include:
   - What the code does
   - Why it's needed
   - Current state of variables at that point
3. algorithm_summary: explain the overall algorithm/pattern used
4. complexity: provide time and space complexity analysis
5. edge_cases: identify boundary conditions and special cases
6. variable_tracking: show variable values at key execution points
7. control_flow: map out if/else branches, loops, function calls
8. summary: brief 1-sentence overview (LEAST important field)

IMPORTANT:
- Focus on building complete mental model of code execution
- Explain logic flow, not just syntax
- Track how data changes through execution
- Return ONLY valid JSON matching structure above`, filename, code)
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
