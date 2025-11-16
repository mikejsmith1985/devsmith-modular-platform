#!/usr/bin/env node

/**
 * Test Issue #2 Fix: Details button HTTP 500 → Now Fixed
 * 
 * WHAT WAS BROKEN:
 * - UI used 'full_learn' value for outputMode
 * - Database only accepts 'quick', 'detailed', 'comprehensive'
 * - 'full_learn' didn't match any prompts → 500 error
 * 
 * WHAT WAS FIXED:
 * - Changed UI to use 'detailed' instead of 'full_learn'
 * - Seeded 15 'detailed' prompts matching database schema
 * - Updated ReviewPage.jsx radio buttons
 * - Cleaned PromptEditorModal.jsx (no hacks)
 * 
 * THIS TEST:
 * - Verifies both 'quick' and 'detailed' modes work
 * - Uses authenticated session (user_id=1)
 * - Tests the fixed endpoint
 */

const http = require('http');

// Test configuration
const HOST = 'localhost';
const PORT = 8081;
const USER_ID = 1; // Simulated authenticated user

// ANSI colors
const GREEN = '\x1b[32m';
const RED = '\x1b[31m';
const YELLOW = '\x1b[33m';
const BLUE = '\x1b[34m';
const RESET = '\x1b[0m';

function makeRequest(path, method = 'GET', headers = {}) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: HOST,
      port: PORT,
      path: path,
      method: method,
      headers: {
        'Content-Type': 'application/json',
        ...headers
      }
    };

    const req = http.request(options, (res) => {
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          body: data
        });
      });
    });

    req.on('error', reject);
    req.end();
  });
}

async function testPromptFetch(outputMode, label) {
  console.log(`\n${BLUE}Testing ${label} mode (output_mode='${outputMode}'):${RESET}`);
  console.log('━'.repeat(60));
  
  // Endpoint: GET /api/review/prompts?mode=preview&user_level=intermediate&output_mode=...
  const path = `/api/review/prompts?mode=preview&user_level=intermediate&output_mode=${outputMode}`;
  
  try {
    const response = await makeRequest(path, 'GET', {
      // Simulate authentication middleware setting user_id
      'X-User-ID': USER_ID.toString()
    });

    console.log(`Status: ${response.statusCode}`);
    console.log(`Path: ${path}`);

    if (response.statusCode === 200) {
      const data = JSON.parse(response.body);
      console.log(`${GREEN}✓ SUCCESS${RESET} - Prompt retrieved`);
      console.log(`Prompt ID: ${data.id}`);
      console.log(`Output Mode: ${data.output_mode}`);
      console.log(`Mode: ${data.mode}`);
      console.log(`User Level: ${data.user_level}`);
      console.log(`Is Default: ${data.is_default}`);
      console.log(`Prompt Text Preview: ${data.prompt_text.substring(0, 100)}...`);
      return true;
    } else {
      console.log(`${RED}✗ FAILED${RESET}`);
      console.log(`Response: ${response.body}`);
      return false;
    }
  } catch (error) {
    console.log(`${RED}✗ ERROR${RESET}`);
    console.log(`Error: ${error.message}`);
    return false;
  }
}

async function main() {
  console.log(`${YELLOW}╔═══════════════════════════════════════════════════════════╗${RESET}`);
  console.log(`${YELLOW}║  Issue #2 Fix Verification Test                          ║${RESET}`);
  console.log(`${YELLOW}║  Testing: Details button 500 error fixed                 ║${RESET}`);
  console.log(`${YELLOW}╚═══════════════════════════════════════════════════════════╝${RESET}`);

  console.log(`\n${BLUE}Background:${RESET}`);
  console.log('  - UI previously used "full_learn" value');
  console.log('  - Database only accepts: quick, detailed, comprehensive');
  console.log('  - Mismatch caused HTTP 500 "Failed to retrieve prompt"');
  console.log(`\n${BLUE}Fix Applied:${RESET}`);
  console.log('  - UI now uses "detailed" instead of "full_learn"');
  console.log('  - Seeded 15 "detailed" prompts in database');
  console.log('  - Both modes should now work');

  // Test both modes
  const quickOk = await testPromptFetch('quick', 'Quick Learn');
  const detailedOk = await testPromptFetch('detailed', 'Full Learn (detailed)');

  // Summary
  console.log(`\n${YELLOW}═══════════════════════════════════════════════════════════${RESET}`);
  console.log(`${YELLOW}Test Summary:${RESET}`);
  console.log(`  Quick Learn:    ${quickOk ? GREEN + '✓ PASS' : RED + '✗ FAIL'}${RESET}`);
  console.log(`  Full Learn:     ${detailedOk ? GREEN + '✓ PASS' : RED + '✗ FAIL'}${RESET}`);

  if (quickOk && detailedOk) {
    console.log(`\n${GREEN}╔═══════════════════════════════════════════════════════════╗${RESET}`);
    console.log(`${GREEN}║  ✓ Issue #2 FIXED - Both modes working correctly         ║${RESET}`);
    console.log(`${GREEN}╚═══════════════════════════════════════════════════════════╝${RESET}`);
    process.exit(0);
  } else {
    console.log(`\n${RED}╔═══════════════════════════════════════════════════════════╗${RESET}`);
    console.log(`${RED}║  ✗ Issue #2 NOT FIXED - Some modes still failing         ║${RESET}`);
    console.log(`${RED}╚═══════════════════════════════════════════════════════════╝${RESET}`);
    process.exit(1);
  }
}

main();
