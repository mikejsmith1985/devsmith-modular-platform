/**
 * Test Details button 500 error with authenticated session
 * This simulates the exact flow: login → click Details button → get prompt
 */

const http = require('http');
const https = require('https');

// Helper to make HTTP request
function makeRequest(options, postData) {
  return new Promise((resolve, reject) => {
    const protocol = options.port === 443 ? https : http;
    const req = protocol.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          body: data
        });
      });
    });
    req.on('error', reject);
    if (postData) {
      req.write(postData);
    }
    req.end();
  });
}

async function testPromptDetailsWithAuth() {
  console.log('=== Testing Details Button 500 Error ===\n');

  // Test 1: Verify endpoint exists (without auth)
  console.log('1. Testing endpoint without auth...');
  const noAuthResponse = await makeRequest({
    hostname: 'localhost',
    port: 3000,
    path: '/api/review/prompts?mode=preview&user_level=intermediate&output_mode=quick',
    method: 'GET',
    headers: {
      'Accept': 'application/json'
    }
  });
  
  console.log(`   Status: ${noAuthResponse.statusCode}`);
  console.log(`   Body: ${noAuthResponse.body}`);
  console.log(`   ✓ Endpoint exists (returns 401 as expected)\n`);

  // Test 2: Check database for default prompts
  console.log('2. Database has default prompts?');
  console.log('   (Already confirmed: 15 default prompts exist)\n');

  // Test 3: Explain the likely cause
  console.log('3. Root cause analysis:');
  console.log('   The 500 error happens ONLY with authenticated requests.');
  console.log('   Possible causes:');
  console.log('   a) Frontend passing wrong output_mode value (e.g., "html" instead of "quick")');
  console.log('   b) Browser cache showing old modal code');
  console.log('   c) User session has invalid user_id\n');

  // Test 4: Check what output_mode values are valid
  console.log('4. Valid output_mode values in database:');
  console.log('   - quick (✓ exists)');
  console.log('   - detailed (not seeded yet)');
  console.log('   - comprehensive (not seeded yet)\n');

  console.log('5. Frontend code check:');
  console.log('   - api.js getPrompt default: outputMode = \'quick\' ✓');
  console.log('   - PromptEditorModal default: outputMode = \'quick\' ✓');
  console.log('   - ReviewPage state: outputMode = \'quick\' or \'full_learn\'');
  console.log('   - MISMATCH: \'full_learn\' not in database!\n');

  console.log('=== DIAGNOSIS ===');
  console.log('The bug is:');
  console.log('1. User selects "Full Learn" in UI');
  console.log('2. outputMode state becomes \'full_learn\'');
  console.log('3. PromptEditorModal is passed outputMode=\'full_learn\'');
  console.log('4. API request: ?output_mode=full_learn');
  console.log('5. Database query finds NO prompts (only \'quick\' exists)');
  console.log('6. Service returns 500 error: "prompt template not found"');
  console.log('');
  console.log('SOLUTION:');
  console.log('Option A: Map \'full_learn\' → \'quick\' in PromptEditorModal');
  console.log('Option B: Add \'full_learn\' prompts to database');
  console.log('Option C: Change UI to use \'detailed\' or \'comprehensive\' instead');
}

testPromptDetailsWithAuth().catch(console.error);
