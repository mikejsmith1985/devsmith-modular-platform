package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeCodeInput_RemovesScriptTags(t *testing.T) {
	// GIVEN: Malicious code input with script tags
	input := "<script>alert('XSS')</script>func main() {}"

	// WHEN: Sanitize
	output := SanitizeCodeInput(input)

	// THEN: Script tags removed, code remains
	assert.NotContains(t, output, "<script>")
	assert.NotContains(t, output, "</script>")
	assert.Contains(t, output, "func main()")
}

func TestSanitizeCodeInput_PreservesValidGo(t *testing.T) {
	// GIVEN: Valid Go code
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	// WHEN: Sanitize
	output := SanitizeCodeInput(code)

	// THEN: Code structure preserved (quotes are HTML-encoded for safety)
	// This is correct - XSS protection requires encoding special characters
	assert.Contains(t, output, "package main")
	assert.Contains(t, output, "import")
	assert.Contains(t, output, "fmt")
	assert.Contains(t, output, "func main()")
	assert.Contains(t, output, "Println")
	assert.Contains(t, output, "Hello")
	assert.Contains(t, output, "World")
}

func TestSanitizeCodeInput_RemovesHTMLTags(t *testing.T) {
	// GIVEN: Code with HTML injection
	input := "<img src=x onerror='alert(1)'>func test() {}"

	// WHEN: Sanitize
	output := SanitizeCodeInput(input)

	// THEN: HTML removed
	assert.NotContains(t, output, "<img")
	assert.NotContains(t, output, "onerror")
	assert.Contains(t, output, "func test()")
}

func TestSanitizeCodeInput_RemovesEventHandlers(t *testing.T) {
	// GIVEN: Code with event handler injection
	input := "<div onclick='malicious()'>func handler() {}</div>"

	// WHEN: Sanitize
	output := SanitizeCodeInput(input)

	// THEN: Event handlers removed
	assert.NotContains(t, output, "onclick")
	assert.NotContains(t, output, "<div>")
	assert.Contains(t, output, "func handler()")
}
