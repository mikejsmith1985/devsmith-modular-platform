-- Seed Data: 20251109_002_detailed_prompts
-- Description: Insert "detailed" output_mode prompts (Full Learn mode with AI reasoning)
-- Author: DevSmith Platform  
-- Date: 2025-11-09

-- Note: "detailed" mode shows AI reasoning process step-by-step
-- This is for users who want to understand HOW the AI analyzes code

-- ====================
-- PREVIEW MODE PROMPTS - DETAILED
-- ====================

-- Preview Mode - Beginner - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_beginner_detailed',
    NULL,
    'preview',
    'beginner',
    'detailed',
    E'You are a code review assistant helping a beginner developer understand code structure.\n\n**Step 1: Initial Analysis**\nLet me analyze the code structure systematically:\n\n```\n{{code}}\n```\n\n**Step 2: File Structure Identification**\n*[AI explains thought process for identifying files/folders]*\n\n**Step 3: Technology Stack Detection**\n*[AI explains how it recognizes languages and frameworks]*\n\n**Step 4: Purpose Understanding**\n*[AI explains reasoning for determining main purpose]*\n\n**Step 5: Entry Points**\n*[AI shows logic for finding where code starts]*\n\n**Final Summary:**\nProvide beginner-friendly overview based on analysis above.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Preview Mode - Intermediate - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_intermediate_detailed',
    NULL,
    'preview',
    'intermediate',
    'detailed',
    E'You are a code review assistant analyzing code structure.\n\n**Analysis Process:**\n\n```\n{{code}}\n```\n\n**Step 1: Structural Analysis**\n*Explain reasoning for file organization patterns*\n\n**Step 2: Architecture Pattern Recognition**\n*Show thought process for identifying MVC, microservices, etc.*\n\n**Step 3: Technology Stack**\n*Detail how frameworks and libraries are detected*\n\n**Step 4: Entry Points and Flow**\n*Trace execution paths and explain logic*\n\n**Conclusion:**\nProvide architectural overview with reasoning.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Preview Mode - Expert - Detailed  
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_preview_expert_detailed',
    NULL,
    'preview',
    'expert',
    'detailed',
    E'Code structure analysis with detailed reasoning:\n\n```\n{{code}}\n```\n\n**Analysis Methodology:**\n1. Pattern recognition: *[explain detection algorithms used]*\n2. Bounded context identification: *[reasoning for domain boundaries]*\n3. Architectural decisions: *[show inference process]*\n4. Technical debt assessment: *[explain evaluation criteria]*\n\n**Detailed findings** with justification for each conclusion.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- ====================
-- SKIM MODE PROMPTS - DETAILED
-- ====================

-- Skim Mode - Beginner - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_beginner_detailed',
    NULL,
    'skim',
    'beginner',
    'detailed',
    E'Let me help you understand what this code does, step by step:\n\n```\n{{code}}\n```\n\n**My Analysis Process:**\n\n**Step 1: Reading the code**\n*[Explain what I see first]*\n\n**Step 2: Finding key functions**\n*[Show how I identify important parts]*\n\n**Step 3: Understanding data flow**\n*[Trace how information moves through code]*\n\n**Step 4: Identifying patterns**\n*[Explain common structures found]*\n\n**Summary:** Simple explanation based on analysis above.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Skim Mode - Intermediate - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_intermediate_detailed',
    NULL,
    'skim',
    'intermediate',
    'detailed',
    E'Code functionality analysis with reasoning:\n\n```\n{{code}}\n```\n\n**Analysis Steps:**\n\n**1. Abstract Syntax Tree** parsing: *[explain code structure]*\n**2. Function signature** analysis: *[show interface detection]*\n**3. Data model** identification: *[explain entity recognition]*\n**4. Workflow mapping**: *[trace execution paths]*\n\n**Detailed Summary** of functionality with supporting evidence.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Skim Mode - Expert - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_skim_expert_detailed',
    NULL,
    'skim',
    'expert',
    'detailed',
    E'Advanced code analysis with methodology:\n\n```\n{{code}}\n```\n\n**Analytical Framework:**\n- Pattern matching: *[specific algorithms applied]*\n- Semantic parsing: *[AST analysis details]*\n- Abstraction layers: *[interface vs implementation reasoning]*\n- Side effect detection: *[purity analysis methods]*\n\n**Comprehensive findings** with technical justification.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- ====================
-- SCAN MODE PROMPTS - DETAILED
-- ====================

-- Scan Mode - Beginner - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_beginner_detailed',
    NULL,
    'scan',
    'beginner',
    'detailed',
    E'Let me search for "{{query}}" and show you how I find it:\n\n```\n{{code}}\n```\n\n**My Search Process:**\n\n**Step 1: Understanding your question**\n*[Explain what I think you''re looking for]*\n\n**Step 2: Scanning the code**\n*[Show where I''m looking and why]*\n\n**Step 3: Finding matches**\n*[Explain why each result matches]*\n\n**Results** with simple explanations.',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- Scan Mode - Intermediate - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_intermediate_detailed',
    NULL,
    'scan',
    'intermediate',
    'detailed',
    E'Semantic search for: {{query}}\n\n```\n{{code}}\n```\n\n**Search Methodology:**\n\n**1. Query parsing:** *[show keyword extraction]*\n**2. Context mapping:** *[explain semantic matching]*\n**3. Relevance scoring:** *[detail ranking algorithm]*\n**4. Result filtering:** *[show selection criteria]*\n\n**Findings** with match confidence scores.',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- Scan Mode - Expert - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_scan_expert_detailed',
    NULL,
    'scan',
    'expert',
    'detailed',
    E'Advanced pattern matching for: {{query}}\n\n```\n{{code}}\n```\n\n**Search Algorithm:**\n- Tokenization: *[lexical analysis method]*\n- Semantic embedding: *[vector space model used]*\n- Context weighting: *[TF-IDF or attention mechanism]*\n- Graph traversal: *[AST navigation strategy]*\n\n**Results** with algorithmic justification.',
    '["{{code}}", "{{query}}"]'::jsonb,
    true,
    1
);

-- ====================
-- DETAILED MODE PROMPTS - DETAILED
-- ====================

-- Detailed Mode - Beginner - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_beginner_detailed',
    NULL,
    'detailed',
    'beginner',
    'detailed',
    E'Let me explain this code line by line, showing my thinking:\n\n```\n{{code}}\n```\n\n**How I Analyze Code:**\n\n**For each important line:**\n- What does this line do?\n- Why is it here?\n- What happens if we change it?\n- *[Show my reasoning process]*\n\n**Overall understanding** with step-by-step logic.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Detailed Mode - Intermediate - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_intermediate_detailed',
    NULL,
    'detailed',
    'intermediate',
    'detailed',
    E'Line-by-line analysis with reasoning:\n\n```\n{{code}}\n```\n\n**Analysis Method:**\n\nFor each code block:\n- **Execution trace:** *[show state changes]*\n- **Control flow:** *[explain branching logic]*\n- **Side effects:** *[identify mutations]*\n- **Complexity:** *[time/space analysis]*\n\n**Comprehensive explanation** with technical details.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Detailed Mode - Expert - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_detailed_expert_detailed',
    NULL,
    'detailed',
    'expert',
    'detailed',
    E'Deep algorithmic analysis:\n\n```\n{{code}}\n```\n\n**Formal Analysis:**\n- **Operational semantics:** *[state transition system]*\n- **Invariants:** *[loop invariants, preconditions/postconditions]*\n- **Complexity bounds:** *[asymptotic analysis with proof sketch]*\n- **Correctness:** *[reasoning about algorithm correctness]*\n\n**Detailed findings** with formal justification.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- ====================
-- CRITICAL MODE PROMPTS - DETAILED
-- ====================

-- Critical Mode - Beginner - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_beginner_detailed',
    NULL,
    'critical',
    'beginner',
    'detailed',
    E'Let me check this code for problems and show you how:\n\n```\n{{code}}\n```\n\n**My Review Process:**\n\n**Step 1: Safety checks**\n*[Explain what I''m looking for]*\n\n**Step 2: Finding issues**\n*[Show why something might be a problem]*\n\n**Step 3: Suggesting fixes**\n*[Explain how to make it better]*\n\n**Findings** with simple explanations and fixes.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Critical Mode - Intermediate - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_intermediate_detailed',
    NULL,
    'critical',
    'intermediate',
    'detailed',
    E'Quality analysis with methodology:\n\n```\n{{code}}\n```\n\n**Review Framework:**\n\n**1. Architecture:** *[explain layering analysis]*\n**2. Security:** *[detail vulnerability scanning]*\n**3. Performance:** *[show optimization identification]*\n**4. Maintainability:** *[explain code smell detection]*\n\n**Issues found** with reasoning and solutions.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Critical Mode - Expert - Detailed
INSERT INTO review.prompt_templates (id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
VALUES (
    'default_critical_expert_detailed',
    NULL,
    'critical',
    'expert',
    'detailed',
    E'Advanced code review with formal methods:\n\n```\n{{code}}\n```\n\n**Analysis Framework:**\n- **Static analysis:** *[abstract interpretation methods]*\n- **Security:** *[formal verification approach]*\n- **Performance:** *[profiling and complexity proofs]*\n- **Architecture:** *[design pattern compliance checking]*\n\n**Comprehensive findings** with formal justification and solutions.',
    '["{{code}}"]'::jsonb,
    true,
    1
);

-- Verification query
SELECT mode, user_level, output_mode, COUNT(*) as count
FROM review.prompt_templates  
WHERE is_default = true
GROUP BY mode, user_level, output_mode
ORDER BY mode, user_level, output_mode;
