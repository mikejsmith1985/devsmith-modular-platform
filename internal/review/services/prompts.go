// Package review_services provides analysis services for different reading modes.
// This file contains prompt templates for guiding AI analysis.
package review_services

import "fmt"

// BuildPreviewPrompt creates a prompt for Preview Mode analysis
// Preview: Quick structural assessment (2-3 minutes)
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
func BuildPreviewPrompt(code, userMode, outputMode string) string {
	// Set default values if not provided
	if userMode == "" {
		userMode = "intermediate"
	}
	if outputMode == "" {
		outputMode = "quick"
	}

	// Adjust tone based on user experience level
	var toneGuidance string
	switch userMode {
	case "beginner":
		toneGuidance = "Use simple, non-technical language with analogies. Explain concepts as if teaching someone new to programming. For example, describe 'bounded contexts' as 'separate areas of responsibility' and 'entry points' as 'starting locations where code execution begins'."
	case "novice":
		toneGuidance = "Use clear language with some technical terms, but explain them briefly. Assume basic programming knowledge but not advanced architecture concepts."
	case "expert":
		toneGuidance = "Use precise technical terminology. Be concise and assume deep understanding of software architecture patterns."
	default: // intermediate
		toneGuidance = "Balance technical accuracy with clarity. Use standard software engineering terms without excessive jargon."
	}

	// Add reasoning trace for full mode
	var reasoningSection string
	if outputMode == "full" {
		reasoningSection = `,
  "reasoning_trace": {
    "analysis_approach": "How you analyzed the code structure",
    "key_observations": ["What patterns or structures you noticed first"],
    "confidence_level": "High/Medium/Low - how certain you are about the analysis"
  }`
	}

	return fmt.Sprintf(`YOU MUST RESPOND WITH ONLY VALID JSON. NO TEXT BEFORE OR AFTER THE JSON. START YOUR RESPONSE WITH { AND END WITH }

TONE GUIDANCE: %s

Analyze this code and return ONLY this JSON structure:

{
  "file_tree": [
    {
      "name": "main.go",
      "type": "file",
      "path": "main.go",
      "description": "Entry point with main function"
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
  "summary": "Brief 1-2 sentence summary"%s
}

CODE TO ANALYZE:
%s

CRITICAL RULES:
- Your ENTIRE response must be valid JSON
- Do NOT write "Based on the code" or any explanatory text
- Do NOT use markdown code blocks
- Do NOT add comments or explanations
- START with { and END with }
- file_tree items MUST have name, type, path, and description fields
- Adjust description complexity based on tone guidance above`, toneGuidance, reasoningSection, code)
}

// BuildSkimPrompt creates a prompt for Skim Mode analysis
// Skim: Abstract overview without implementation details (5-7 minutes)
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
func BuildSkimPrompt(code, userMode, outputMode string) string {
	// Set default values if not provided
	if userMode == "" {
		userMode = "intermediate"
	}
	if outputMode == "" {
		outputMode = "quick"
	}

	// Adjust tone based on user experience level
	var toneGuidance string
	switch userMode {
	case "beginner":
		toneGuidance = "Use simple language with everyday analogies. Explain 'functions' as 'actions the code can perform', 'interfaces' as 'contracts that define what something must do', and 'workflows' as 'step-by-step processes'. Avoid jargon."
	case "novice":
		toneGuidance = "Use clear language with basic technical terms. Explain what functions and interfaces do, not just list them. Assume some programming background."
	case "expert":
		toneGuidance = "Use precise technical terminology. Focus on architectural patterns, design decisions, and abstractions. Be concise."
	default: // intermediate
		toneGuidance = "Use standard software engineering terminology with clear, practical descriptions."
	}

	// Add reasoning trace for full mode
	var reasoningSection string
	if outputMode == "full" {
		reasoningSection = `,
  "reasoning_trace": {
    "abstraction_identification": "How you identified key abstractions",
    "pattern_recognition": "What design patterns you noticed",
    "confidence_level": "High/Medium/Low"
  }`
	}

	return fmt.Sprintf(`YOU MUST RESPOND WITH ONLY VALID JSON. NO TEXT BEFORE OR AFTER THE JSON. START YOUR RESPONSE WITH { AND END WITH }

TONE GUIDANCE: %s

Analyze this code and return ONLY this JSON structure:

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
  "summary": "Brief overview of what this code provides"%s
}

CODE TO ANALYZE:
%s

CRITICAL RULES:
- Your ENTIRE response must be valid JSON
- Do NOT write "The code appears to" or any explanatory text
- Do NOT use markdown code blocks
- Do NOT add comments or explanations
- START with { and END with }
- For functions: include name, full signature, and brief description
- For interfaces: describe purpose and list method signatures
- Adjust description complexity based on tone guidance above`, toneGuidance, reasoningSection, code)
}

// BuildScanPrompt creates a prompt for Scan Mode analysis
// Scan: Targeted pattern search (3-5 minutes)
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
func BuildScanPrompt(code, query, userMode, outputMode string) string {
	// Set default values if not provided
	if userMode == "" {
		userMode = "intermediate"
	}
	if outputMode == "" {
		outputMode = "quick"
	}

	// Adjust tone based on user experience level
	var toneGuidance string
	switch userMode {
	case "beginner":
		toneGuidance = "Explain why each match is relevant using simple terms. Use analogies like 'this code is like a gatekeeper checking IDs' for authentication. Avoid technical jargon."
	case "novice":
		toneGuidance = "Explain the relevance of each match clearly. Use basic technical terms but keep explanations practical and concrete."
	case "expert":
		toneGuidance = "Be precise and technical. Focus on pattern recognition, architectural implications, and code quality aspects of each match."
	default: // intermediate
		toneGuidance = "Explain match relevance using clear software engineering terminology. Balance detail with conciseness."
	}

	// Add reasoning trace for full mode
	var reasoningSection string
	if outputMode == "full" {
		reasoningSection = `,
  "reasoning_trace": {
    "search_strategy": "How you approached finding matches",
    "filtering_logic": "Why you included/excluded certain results",
    "confidence_level": "High/Medium/Low"
  }`
	}

	return fmt.Sprintf(`YOU MUST RESPOND WITH ONLY VALID JSON. NO TEXT BEFORE OR AFTER THE JSON. START YOUR RESPONSE WITH { AND END WITH }

TONE GUIDANCE: %s

Find patterns matching this query and return ONLY this JSON structure:

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
  "summary": "Found 3 matches for query in the codebase"%s
}

QUERY: "%s"

CODE TO ANALYZE:
%s

CRITICAL RULES:
- Your ENTIRE response must be valid JSON
- Do NOT write "Based on the query" or any explanatory text
- Do NOT use markdown code blocks
- Do NOT add comments or explanations
- START with { and END with }
- Relevance score 0.0-1.0 (1.0 = perfect match)
- Adjust reason explanations based on tone guidance above`, toneGuidance, query, reasoningSection, query, code)
}

// BuildDetailedPrompt creates a prompt for Detailed Mode analysis
// Detailed: Line-by-line understanding (10-15 minutes)
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
func BuildDetailedPrompt(code, filename, userMode, outputMode string) string {
	// Set default values if not provided
	if userMode == "" {
		userMode = "intermediate"
	}
	if outputMode == "" {
		outputMode = "quick"
	}

	// Adjust tone based on user experience level
	var toneGuidance string
	switch userMode {
	case "beginner":
		toneGuidance = "Use detailed analogies for every concept. For example, explain loops as 'repeating the same task until a condition is met, like checking each item in a shopping cart'. Explain variables as 'named boxes that store information'. Avoid assuming any prior knowledge."
	case "novice":
		toneGuidance = "Explain each line clearly with practical examples. Use analogies where helpful but introduce technical terms. Assume basic programming knowledge but explain intermediate concepts."
	case "expert":
		toneGuidance = "Focus on algorithmic complexity, design patterns, optimization opportunities, and subtle implementation details. Be concise and technical."
	default: // intermediate
		toneGuidance = "Provide clear technical explanations. Explain the 'why' not just the 'what'. Balance detail with readability."
	}

	// Add reasoning trace for full mode
	var reasoningSection string
	if outputMode == "full" {
		reasoningSection = `,
  "reasoning_trace": {
    "analysis_method": "How you approached understanding this code",
    "key_insights": ["Important patterns or techniques you identified"],
    "complexity_assessment": "Why you rated complexity the way you did",
    "confidence_level": "High/Medium/Low"
  }`
	}

	return fmt.Sprintf(`YOU MUST RESPOND WITH ONLY VALID JSON. NO TEXT BEFORE OR AFTER THE JSON. START YOUR RESPONSE WITH { AND END WITH }

TONE GUIDANCE: %s

Analyze this code and return ONLY this JSON structure:

{
  "line_explanations": [
    {
      "line_number": 1,
      "code": "func BinarySearch(arr []int, target int) int {",
      "explanation": "Function declaration that takes a sorted array and target value",
      "variables": "arr: input array, target: value to find"
    }
  ],
  "algorithm_summary": "Binary search algorithm implementation",
  "complexity": "Time: O(log n), Space: O(1)",
  "edge_cases": [
    "Empty array: returns -1",
    "Target not found: returns -1"
  ],
  "variable_tracking": [
    {
      "line_number": 2,
      "variables": {
        "left": "0",
        "right": "len(arr)-1"
      }
    }
  ],
  "control_flow": [
    {
      "type": "loop",
      "line_number": 3,
      "description": "While loop continues until left exceeds right"
    }
  ],
  "summary": "Binary search implementation"%s
}

FILE: %s

CODE TO ANALYZE:
%s

CRITICAL RULES:
- Your ENTIRE response must be valid JSON
- Do NOT write "This code" or any explanatory text
- Do NOT use markdown code blocks
- Do NOT add comments or explanations
- START with { and END with }
- line_explanations is the PRIMARY OUTPUT - explain EVERY significant line
- Adjust explanation depth and language based on tone guidance above`, toneGuidance, reasoningSection, filename, code)
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
