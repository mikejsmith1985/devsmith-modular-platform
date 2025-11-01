import { test, expect } from '@playwright/test';

test('Ollama integration: Critical mode endpoint returns analysis', async ({ page }) => {
  const response = await page.request.post('http://localhost:3000/api/review/modes/critical', {
    data: {
      code: `package main
import "fmt"

func main() {
  x := 1
  fmt.Println(x)
}`
    }
  });
  
  console.log('Critical mode response status:', response.status());
  const text = await response.text();
  console.log('Response length:', text.length);
  console.log('Response contains analysis:', text.includes('analysis') || text.includes('quality') || text.includes('error'));
  console.log('First 200 chars:', text.substring(0, 200));
  
  expect(response.status()).toBe(200);
  expect(text.length).toBeGreaterThan(0);
});
