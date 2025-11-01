import { test, expect } from '@playwright/test';

const testCode = `package main
import "fmt"

func main() {
  fmt.Println("Hello")
}`;

const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];

test('All 5 reading modes return analysis', async ({ page }) => {
  for (const mode of modes) {
    const response = await page.request.post(`http://localhost:3000/api/review/modes/${mode}`, {
      data: { code: testCode }
    });
    
    expect(response.status(), `${mode} should return 200`).toBe(200);
    const text = await response.text();
    expect(text.length, `${mode} should return content`).toBeGreaterThan(0);
    console.log(`✅ ${mode}: ${text.length} bytes`);
  }
});
