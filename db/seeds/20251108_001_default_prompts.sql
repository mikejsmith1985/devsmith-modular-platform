-- Seed Data: 20251108_001_default_prompts
-- Description: Insert default AI prompts for all review modes and user levels
-- Author: DevSmith Platform
-- Date: 2025-11-08

-- Note: Using "quick" output_mode as default for all combinations
-- Users can customize these or we can add more variations later

-- ====================
-- PREVIEW MODE PROMPTS
-- ====================

-- Preview Mode - Beginner
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_beginner_quick',
    NULL,
    'preview',
    'beginner',
    'quick',
    E'You are a code review assistant helping a beginner developer understand code structure.\n\nAnalyze the following code and provide a simple overview:\n\n```\n{{code}}\n```\n\nProvide:\n1. **File Structure**: List main files and folders in a simple tree format\n2. **Technology Stack**: What programming languages and frameworks are used? (explain in simple terms)\n3. **Main Purpose**: What does this code do? (1-2 sentences)\n4. **Entry Points**: Where does the code start running? (e.g., main.go, index.js)\n\nKeep explanations simple and beginner-friendly. Avoid jargon.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Preview Mode - Intermediate
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_intermediate_quick',
    NULL,
    'preview',
    'intermediate',
    'quick',
    E'You are a code review assistant for intermediate developers.\n\nAnalyze this codebase and provide a structural overview:\n\n```\n{{code}}\n```\n\nProvide:\n1. **File Structure**: Hierarchical tree with brief descriptions\n2. **Bounded Contexts**: Identify domain boundaries (e.g., "Auth domain", "Data layer")\n3. **Technology Stack**: Languages, frameworks, key libraries\n4. **Architectural Pattern**: Is this layered, microservices, MVC, clean architecture?\n5. **Entry Points**: Main functions, startup files, API roots\n6. **External Dependencies**: Databases, APIs, third-party services\n\nBe concise but technical.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Preview Mode - Expert
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_expert_quick',
    NULL,
    'preview',
    'expert',
    'quick',
    E'Analyze this codebase architecture with expert-level depth:\n\n```\n{{code}}\n```\n\nProvide:\n1. **Architecture Pattern**: Identify pattern (hexagonal, CQRS, event-driven, etc.) with evidence\n2. **Bounded Contexts**: Domain boundaries with coupling analysis\n3. **Technology Stack**: Full stack with version implications\n4. **Abstraction Layers**: How concerns are separated (ports/adapters, interfaces, etc.)\n5. **Entry Points & Lifecycle**: Startup sequence, dependency injection, initialization\n6. **Integration Points**: External systems, message queues, webhooks\n7. **Architectural Strengths/Weaknesses**: Brief analysis\n\nAssume deep technical knowledge.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- =================
-- SKIM MODE PROMPTS
-- =================

-- Skim Mode - Beginner
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_beginner_quick',
    NULL,
    'skim',
    'beginner',
    'quick',
    E'You are helping a beginner understand what this code does (not how it works).\n\nCode:\n```\n{{code}}\n```\n\nProvide:\n1. **Functions List**: List each function with a simple description of what it does\n2. **Main Data**: What information does this code work with? (users, products, etc.)\n3. **Key Workflows**: Describe 2-3 main tasks this code performs\n\nUse simple language. Focus on WHAT, not HOW.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Skim Mode - Intermediate
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_intermediate_quick',
    NULL,
    'skim',
    'intermediate',
    'quick',
    E'Analyze abstractions and contracts in this code:\n\n```\n{{code}}\n```\n\nProvide:\n1. **Function Signatures**: List with brief descriptions (inputs/outputs)\n2. **Interface Definitions**: Abstract contracts and their purposes\n3. **Data Models**: Key structs/classes and their relationships\n4. **Major Workflows**: Sequence diagrams or step-by-step descriptions\n5. **API Endpoints**: If applicable, list routes and their purposes\n\nFocus on contracts and abstractions, not implementations.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Skim Mode - Expert
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_expert_quick',
    NULL,
    'skim',
    'expert',
    'quick',
    E'Provide expert-level abstraction analysis:\n\n```\n{{code}}\n```\n\nAnalyze:\n1. **Abstractions**: Interfaces, traits, protocols - evaluate design quality\n2. **Contracts**: Function signatures with behavioral contracts (preconditions, postconditions)\n3. **Type System Usage**: Generics, type constraints, variance (if applicable)\n4. **Data Flow**: High-level data transformations and pipelines\n5. **Design Patterns**: Identified patterns and their appropriateness\n6. **Dependency Direction**: Analyze coupling and adherence to dependency inversion\n\nCritique abstraction quality and suggest improvements.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- =================
-- SCAN MODE PROMPTS
-- =================

-- Scan Mode - Beginner
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_beginner_quick',
    NULL,
    'scan',
    'beginner',
    'quick',
    E'Help me find specific code:\n\nQuery: {{query}}\n\nCode:\n```\n{{code}}\n```\n\nFind and explain:\n1. **Matches**: Show each place this appears (with file:line)\n2. **Context**: Show 3 lines before and after\n3. **Explanation**: Why does this code matter? What does it do?\n\nUse simple explanations.',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- Scan Mode - Intermediate
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_intermediate_quick',
    NULL,
    'scan',
    'intermediate',
    'quick',
    E'Semantic search query: {{query}}\n\nCodebase:\n```\n{{code}}\n```\n\nFind:\n1. **Exact Matches**: Direct occurrences\n2. **Semantic Matches**: Related code that serves similar purpose (even if different names)\n3. **Usage Tracking**: How and where these elements are used\n4. **Related Code**: Other functions/classes that interact with matches\n\nFor each match provide:\n- File path and line number\n- 3 lines of context before/after\n- Brief explanation of relevance',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- Scan Mode - Expert
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_expert_quick',
    NULL,
    'scan',
    'expert',
    'quick',
    E'Expert semantic search: {{query}}\n\n```\n{{code}}\n```\n\nProvide:\n1. **Semantic Analysis**: Find all related code (exact matches, similar patterns, analogous implementations)\n2. **Data Flow Tracing**: Show how data flows through matched elements\n3. **Call Graph**: Callers and callees of matched functions\n4. **Pattern Recognition**: If query describes a pattern, find all instances\n5. **Cross-Cutting Concerns**: How matches relate to broader system concerns\n\nFormat as JSON with:\n- matches: [{file, line, code, relevance_score, explanation}]\n- call_graph: {callers: [], callees: []}\n- data_flow: []',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- ====================
-- DETAILED MODE PROMPTS
-- ====================

-- Detailed Mode - Beginner
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_beginner_quick',
    NULL,
    'detailed',
    'beginner',
    'quick',
    E'Explain this code step-by-step for a beginner:\n\n```\n{{code}}\n```\n\nFor each important line:\n1. **What it does**: Explain in simple terms\n2. **Why it''s needed**: What would break without it?\n3. **Variables**: What values do variables have at this point?\n\nUse analogies and simple language. Avoid technical jargon where possible.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Detailed Mode - Intermediate
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_intermediate_quick',
    NULL,
    'detailed',
    'intermediate',
    'quick',
    E'Provide detailed line-by-line analysis:\n\n```\n{{code}}\n```\n\nFor each significant section:\n1. **What**: Precise description of operations\n2. **Why**: Rationale for this approach\n3. **Variable State**: Values and types at each point\n4. **Control Flow**: Branches, loops, conditions explained\n5. **Edge Cases**: What could go wrong? How is it handled?\n6. **Algorithm**: If applicable, name the algorithm and analyze complexity\n\nProvide code snippets with inline comments.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Detailed Mode - Expert
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_expert_quick',
    NULL,
    'detailed',
    'expert',
    'quick',
    E'Expert-level deep dive analysis:\n\n```\n{{code}}\n```\n\nAnalyze:\n1. **Implementation Details**: Precise algorithmic analysis with complexity (time/space)\n2. **State Transitions**: How state changes through execution\n3. **Invariants**: What conditions are maintained throughout?\n4. **Memory Model**: Allocation patterns, ownership, lifetimes (if applicable)\n5. **Concurrency**: Race conditions, synchronization, thread safety\n6. **Performance**: Hotspots, optimization opportunities\n7. **Correctness**: Prove correctness or identify bugs\n8. **Alternatives**: Better approaches with trade-off analysis\n\nFormat as detailed technical documentation with code annotations.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- =====================
-- CRITICAL MODE PROMPTS
-- =====================

-- Critical Mode - Beginner
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_beginner_quick',
    NULL,
    'critical',
    'beginner',
    'quick',
    E'Review this code for common issues:\n\n```\n{{code}}\n```\n\nCheck for:\n1. **Basic Errors**: Syntax errors, typos, undefined variables\n2. **Simple Security**: Are passwords or secrets visible in code?\n3. **Code Quality**: Are variable names clear? Are functions too long?\n4. **Common Mistakes**: Typical beginner errors\n\nFor each issue:\n- Explain why it''s a problem (in simple terms)\n- Show how to fix it (with code example)\n- Rate severity: High, Medium, Low',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Critical Mode - Intermediate
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_intermediate_quick',
    NULL,
    'critical',
    'intermediate',
    'quick',
    E'Comprehensive code review:\n\n```\n{{code}}\n```\n\nEvaluate:\n1. **Architecture**: Layer violations, missing abstractions, tight coupling\n2. **Code Quality**: Naming, error handling, documentation, testability\n3. **Security**: SQL injection, XSS, unvalidated input, exposed secrets\n4. **Performance**: N+1 queries, unnecessary allocations, inefficient algorithms\n5. **Best Practices**: Language idioms, framework conventions\n\nFor each issue:\n- Type: architecture | security | performance | quality\n- Severity: critical | important | minor\n- Location: file:line\n- Description: What''s wrong and why\n- Fix: Code example showing correction\n- Priority: Order by impact',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Critical Mode - Expert
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_expert_quick',
    NULL,
    'critical',
    'expert',
    'quick',
    E'Expert architectural and security audit:\n\n```\n{{code}}\n```\n\nDeep analysis:\n1. **Architectural Integrity**: DDD violations, bounded context leakage, dependency inversions\n2. **Design Patterns**: Misapplied patterns, missing patterns, over-engineering\n3. **Security**: OWASP Top 10, cryptographic weaknesses, authentication/authorization flaws\n4. **Performance**: Algorithmic complexity, caching opportunities, database optimization\n5. **Concurrency**: Data races, deadlocks, lock contention, actor model violations\n6. **Testing**: Missing test coverage, test quality, integration test gaps\n7. **Maintainability**: Technical debt, code smells, refactoring opportunities\n8. **Scalability**: Bottlenecks, stateful operations, distributed system concerns\n\nFormat as structured JSON:\n```json\n{\n  "critical_issues": [{type, severity, location, description, fix, rationale}],\n  "architectural_assessment": {...},\n  "security_audit": {...},\n  "performance_analysis": {...},\n  "recommendations": [{priority, effort, impact}]\n}\n```',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Verification: Ensure we have exactly 15 default prompts
DO $$
DECLARE
    prompt_count INT;
BEGIN
    SELECT COUNT(*) INTO prompt_count
    FROM review.prompt_templates
    WHERE is_default = true;
    
    IF prompt_count != 15 THEN
        RAISE EXCEPTION 'Expected 15 default prompts, found %', prompt_count;
    END IF;
    
    RAISE NOTICE 'Successfully seeded 15 default prompts (5 modes × 3 user levels)';
END $$;
