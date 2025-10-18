# Learning Development Platform: Roles and Responsibilities

## Project Overview
The Learning Development Platform, hosted at [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform), is a modular, AI-driven platform for learning, debugging, and building code. This document defines the roles of the project team, adhering to the DevSmith Coding Standards (from DevSmith Logs project) and Test-Driven Development (TDD) principles. The goal is to maintain a clean, recoverable repo with high-quality code and robust testing.

## Roles and Responsibilities

### 1. Project Orchestrator and Manager (Mike)
- **Role**: Oversees the project, manages the agent team, and ensures alignment with project goals.
- **Responsibilities**:
  - Define and prioritize features based on `REQUIREMENTS.md`.
  - Create GitHub issues with clear, single-feature tasks and acceptance criteria using issue templates.
  - Review and approve pull requests (PRs) after Claude’s architectural review.
  - Merge approved PRs into the `development` branch and manage releases to `main`.
  - Monitor project progress and ensure adherence to TDD and coding standards.
  - Validate backups (logs, code states, model configurations) to ensure recoverability.
  - Coordinate sprints (e.g., Sprint 1: Portal + Logging) and track milestones.
  - Configure GitHub branch protection rules to enforce tests, approvals, and changelog updates.
- **Tools**:
  - GitHub for issue tracking, PR approvals, and repo management.
  - Project management tools (e.g., GitHub Projects) for sprint planning.

### 2. Primary Architect and PR Reviewer (Claude)
- **Role**: Designs the platform’s architecture and reviews code for quality and adherence to standards.
- **Responsibilities**:
  - Design the modular architecture, ensuring apps (logging, analytics, review, build) are isolated yet interoperable.
  - Define database schemas (Postgres, pluggable for MongoDB/SQLite) and API contracts.
  - Review Copilot-generated PRs for:
    - Adherence to DevSmith Coding Standards (file organization, naming, error handling).
    - Architectural integrity (modularity, scalability, performance).
    - Alignment with TDD principles (test coverage, test quality).
    - Security and debugging best practices (e.g., friendly error messages, logging).
  - Provide detailed feedback on PRs, suggesting improvements or refactoring.
  - Validate AI-driven features (e.g., Ollama integration, review app reading modes).
  - Ensure WebSocket implementation for real-time logging is robust.
  - Recommend optimizations for the one-click installation process.
- **Tools**:
  - GitHub for PR reviews and comments.
  - FastAPI/Postgres for backend architecture.
  - React for frontend architecture.

### 3. Primary Code and Test Generator (Copilot)
- **Role**: Generates code and tests for a single feature at a time, following DevSmith Coding Standards.
- **Responsibilities**:
  - Implement one feature per task, as defined in GitHub issues by the Orchestrator.
  - Follow DevSmith Coding Standards:
    - **File Organization**:
      - Backend: Python/FastAPI in `backend/{main.py, models/, routes/, services/, tests/, requirements.txt}`.
      - Frontend: React in `frontend/src/{components/, context/, hooks/, styles/, utils/, __tests__}`.
    - **Naming Conventions**:
      - Files: `PascalCase.jsx` for components, `camelCase.js` for utilities, `kebab-case.css` for styles, `ComponentName.test.jsx` for tests.
      - Code: `camelCase` for variables/functions, `UPPER_SNAKE_CASE` for constants, `PascalCase` for classes/components.
    - **React Component Structure**:
      ```javascript
      import React, { useState, useEffect } from 'react';
      export default function ComponentName({ prop1, prop2 = 'default', onAction }) {
        const [data, setData] = useState(null);
        const [loading, setLoading] = useState(false);
        const [error, setError] = useState(null);
        useEffect(() => { /* Effect logic */ }, []);
        const handleClick = () => { /* Handler logic */ };
        if (loading) return <div>Loading...</div>;
        if (error) return <div>Error: {error.message}</div>;
        return <div className="component">{/* JSX */}</div>;
      }
      ```
    - **API Call Pattern**:
      ```javascript
      async function fetchData(params) {
        try {
          const response = await fetch(url, { method: 'GET', headers: { 'Content-Type': 'application/json' } });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          const data = await response.json();
          return data.key || fallbackValue;
        } catch (err) {
          console.error('Error:', err);
          return fallbackValue;
        }
      }
      ```
    - **Error Handling**: Provide user-friendly messages, fallback values, loading states, and console logging.
  - Write TDD-compliant tests (unit, component, API, integration) before coding, targeting 70%+ unit test coverage and 90%+ critical path coverage.
  - Run tests locally (`npm test`, `pytest`) before committing.
  - Perform manual testing checklist:
    - Feature works in browser.
    - No console errors/warnings.
    - Regression check for related features.
    - Light/dark mode compatibility (if applicable).
    - Responsive design for mobile/tablet (if applicable).
  - Test through nginx gateway (`http://localhost:3000`), direct access, authentication, WebSocket connections, and HMR.
  - Create feature branches from `development`, commit with Conventional Commit messages (e.g., `feat: add login button`), and update changelogs.
  - Push code and create PRs to `development` for review.
- **Tools**:
  - GitHub Copilot for code generation.
  - Jest, Pytest, Cypress for test generation.
  - VS Code for coding and debugging.

## Workflow
1. **Issue Creation**: Orchestrator creates a GitHub issue for a single feature with clear acceptance criteria using issue templates.
2. **Feature Development**:
   - Copilot creates a feature branch from `development`.
   - Copilot writes tests first (per `TDD.md`), then implements the feature, adhering to DevSmith Coding Standards.
   - Copilot runs tests (`npm test`, `pytest`) and manual checklist.
   - Copilot commits with a Conventional Commit message (e.g., `feat(logging): add WebSocket log streaming`) and updates the changelog.
3. **PR Creation**:
   - Copilot pushes the branch and creates a PR to `development`.
   - GitHub Actions run automated tests, linting, and coverage checks.
4. **Review Process**:
   - Claude reviews the PR for architecture, standards, and TDD compliance.
   - Orchestrator reviews Claude’s feedback and approves or requests changes.
5. **Merge and Release**:
   - Orchestrator merges the PR to `development` after approval.
   - For releases, Orchestrator merges `development` to `main`.
6. **Backup Validation**:
   - Orchestrator periodically verifies backups (logs, code states, models) via automated tests and manual checks.

## Testing Requirements
- **Automated Testing** (Copilot):
  - Unit tests for utilities (70% coverage).
  - Component tests for React components.
  - API endpoint tests for backend routes.
  - Integration tests for critical paths (e.g., login → portal → app launch).
  - Commands:
    ```bash
    # Frontend tests
    cd apps/platform-frontend && npm test
    cd apps/review-frontend && npm test
    cd apps/logs-frontend && npm test
    # Backend tests
    cd apps/platform-backend && pytest
    cd apps/review-backend && pytest
    cd apps/logs-backend && pytest
    # Integration tests
    ./tests/integration-tests.sh
    ```
- **Manual Testing Checklist** (Copilot):
  - Feature works in browser.
  - No console errors/warnings.
  - Regression check for related features.
  - Light/dark mode compatibility.
  - Responsive design for mobile/tablet.
  - Gateway/proxy testing via `http://localhost:3000`.
- **CI Pipeline** (Orchestrator):
  - GitHub Actions enforce tests, linting, and coverage.
  - Branch protection requires passing tests and one approval.

## Workflow Improvements
- **Automated PR Checks**: GitHub Actions run tests, linting, and coverage checks on PRs to catch issues early.
- **Branch Protection**: Require passing tests, one approval (Orchestrator), and updated changelogs for `development` and `main`.
- **Conventional Commits**: Use `feat:`, `fix:`, `docs:`, etc., for clear commit history and automated changelog generation.
- **Pre-Commit Hooks**: Use Husky (frontend) and pre-commit (backend) to run linting and tests locally before commits.
- **Issue Templates**: Standardize feature/bug reports with acceptance criteria to ensure Copilot focuses on single features.
- **Backup Tests**: Add automated tests for backup system to verify recoverability.
- **Sprints**: Organize development into sprints (e.g., Sprint 1: Portal + Logging) with milestones for tracking.

## Notes
- Copilot must focus on one feature per issue to avoid scope creep and maintain repo clarity.
- Claude’s architectural reviews ensure modularity and scalability, reducing technical debt.
- Orchestrator’s oversight and approval process ensures alignment with project goals and recoverability.
- All team members adhere to TDD principles per `TDD.md` and DevSmith Coding Standards.
