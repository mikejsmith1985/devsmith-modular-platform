# DevSmith LLM Coding Guide

## Overview

This document outlines the guidelines for coding with the DevSmith LLM project. It covers topics such as coding standards, testing, and best practices.

## Code Standards

* Use Go 1.18+
* Follow the Go standard library's naming conventions
* Use consistent variable naming conventions throughout the codebase

...

## TODO List

### Syntax Error
- File: claude/models/claude_model.go (Lines 12-15)
  - Missing closing parenthesis for struct declaration.
    Suggested fix: Add a closing parenthesis to complete the struct definition.

### Coding Standard Deviation
- File: services/reviewAI.go (Line 30)
  - Variable naming convention inconsistency (mix of camelCase and underscore notation).
    Suggested fix: Use consistent variable naming conventions throughout the file.
- File: services/reviewAI_test.go (Line 50)
  - Missing tests for error handling cases.
    Suggested fix: Add additional tests to cover error handling scenarios.

### Potential Bug
- File: services/reviewAI.go (Lines 101-200)
  - Incorrect usage of the `llama3.2:3b` API.
    Suggested fix: Review the API documentation to ensure correct usage.
- File: internal/review/services/predict.go (Lines 50-100)
  - Missing validation for input data.
    Suggested fix: Add validation to ensure the input data is valid and consistent.
- File: internal/review/services/detailed_mode_test.go (Lines 20-50)
  - Incorrect assertions.
    Suggested fix: Review and correct the assertions to ensure they accurately reflect the expected behavior.

### .husky/
- Update `golangci-lint` to the latest version.
- Remove unnecessary test commands from `pre-commit.sh`.
- Remove unnecessary test commands from `.husky/skipping-tests`.

...

Please let me know if you have any questions or need further assistance with implementing these changes!

(Note: I'll keep this file up-to-date as we work through the repository assessment.)

I've created a new file at the root of your repository with the TODO list and its corresponding suggested fixes. This will ensure that the document is backed up on GitHub until the full repository assessment is complete.