-- Week 1 Testing: Create test project with API key for batch ingestion testing
-- This script creates a test project and generates an API key for E2E testing

-- Insert test project (user_id = 1 for testing)
-- The API key hash is for plain key: dsk_test_RK3jP9mL2nQ8vF7dW5tX (SAVE THIS!)
-- Generated using: bcrypt.GenerateFromPassword([]byte("dsk_test_RK3jP9mL2nQ8vF7dW5tX"), bcrypt.DefaultCost)

INSERT INTO logs.projects (
    user_id,
    name,
    slug,
    description,
    repository_url,
    api_key_hash,
    is_active,
    created_at,
    updated_at
) VALUES (
    1,
    'Test Application',
    'test-app',
    'External app for testing batch ingestion API',
    'https://github.com/example/test-app',
    '$2a$10$/tHduRQUv1pDNeEVMAL9gOwmgkefKAoz42Vj8QJZ67DHQIRin4Wjq',  -- Correct bcrypt hash (verified with test_bcrypt.go)
    true,
    NOW(),
    NOW()
) ON CONFLICT (user_id, slug) DO UPDATE 
  SET api_key_hash = EXCLUDED.api_key_hash,
      updated_at = NOW();

-- Verify project was created
SELECT 
    id,
    user_id,
    name,
    slug,
    description,
    is_active,
    created_at
FROM logs.projects
WHERE slug = 'test-app';

-- Show API key for testing (COPY THIS VALUE!)
\echo '================================================'
\echo 'Test Project Created Successfully!'
\echo '================================================'
\echo 'Project Slug: test-app'
\echo 'API Key: dsk_test_RK3jP9mL2nQ8vF7dW5tX'
\echo ''
\echo 'Use this API key in your batch ingestion tests:'
\echo 'Authorization: Bearer dsk_test_RK3jP9mL2nQ8vF7dW5tX'
\echo '================================================'
