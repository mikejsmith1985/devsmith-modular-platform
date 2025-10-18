# Learning Development Platform: TDD Document

## Overview
This Test-Driven Development (TDD) document outlines test cases for the Learning Development Platform, hosted at [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform). The tests ensure that each component meets the requirements specified in `REQUIREMENTS.md`, following a modular, user-centric, and debug-friendly architecture. Tests are organized by component and prioritized to align with the build order (portal → logging → analytics → review → build).

## Test Framework
- **Tools**: Use Jest for JavaScript unit tests, Pytest for Python components (e.g., Ollama integration), and Cypress for end-to-end UI tests.
- **Approach**: Write tests before implementation. Each test case must pass before proceeding to the next development step.
- **Coverage**: Aim for 90%+ code coverage, with focus on critical paths (auth, modularity, AI features, debugging).

## 1. Authentication
### Test Case 1.1: GitHub OAuth Login
- **Description**: Verify that users can log in using GitHub OAuth and access their repositories.
- **Preconditions**: GitHub OAuth app configured with valid client ID and secret.
- **Steps**:
  1. Navigate to the login page.
  2. Click the "Login with GitHub" button.
  3. Enter valid GitHub credentials in the OAuth prompt.
  4. Redirect back to the platform.
- **Expected Outcome**: User is authenticated, and their GitHub profile and repositories are accessible in the portal.
- **Failure Case**: Invalid credentials or OAuth misconfiguration return a clear error message (e.g., "Failed to authenticate with GitHub: Invalid client ID").

### Test Case 1.2: Unauthorized Access
- **Description**: Ensure unauthenticated users cannot access protected endpoints.
- **Preconditions**: User is not logged in.
- **Steps**:
  1. Attempt to access the portal dashboard via URL.
  2. Attempt to fetch a protected API endpoint (e.g., `/api/user/repos`).
- **Expected Outcome**: Redirect to login page with a friendly error: "Please log in with GitHub to continue."
- **Failure Case**: Access granted without authentication.

## 2. Modularity
### Test Case 2.1: App Toggle Functionality
- **Description**: Verify that users can enable/disable apps independently.
- **Preconditions**: User is logged in, portal is loaded.
- **Steps**:
  1. Navigate to the app browser in the portal.
  2. Toggle the logging app to "enabled."
  3. Toggle the analytics app to "disabled."
  4. Access the logging app UI.
  5. Attempt to access the analytics app UI.
- **Expected Outcome**: Logging app is accessible; analytics app returns a clear message: "Analytics app is disabled. Enable it in the portal."
- **Failure Case**: Disabled apps are accessible or enabled apps are inaccessible.

### Test Case 2.2: Isolated App Operation
- **Description**: Ensure apps operate independently without cross-dependencies.
- **Preconditions**: User is logged in, only the review app is enabled.
- **Steps**:
  1. Open the review app.
  2. Import a GitHub repo and perform a code review.
  3. Check for logging or analytics app dependencies.
- **Expected Outcome**: Review app functions fully without requiring logging or analytics apps to be enabled.
- **Failure Case**: Review app fails due to missing dependencies from other apps.

## 3. Database
### Test Case 3.1: Postgres Schema Isolation
- **Description**: Verify that each app’s database schema is isolated.
- **Preconditions**: Postgres database initialized, logging and review apps enabled.
- **Steps**:
  1. Insert sample log data into the logging app’s schema.
  2. Query the review app’s schema for the same data.
- **Expected Outcome**: Review app schema contains no log data, confirming isolation.
- **Failure Case**: Log data appears in the review app’s schema.

### Test Case 3.2: Federated Queries
- **Description**: Ensure cross-app queries work when multiple apps are enabled.
- **Preconditions**: Logging and analytics apps enabled, Postgres configured.
- **Steps**:
  1. Generate logs in the logging app.
  2. Run an analytics query to count log entries by severity.
- **Expected Outcome**: Analytics app correctly aggregates log data across schemas.
- **Failure Case**: Query fails or returns incorrect data due to schema mismatch.

## 4. Logging App
### Test Case 4.1: Real-Time Log Tracking
- **Description**: Verify that logs are displayed in real-time via WebSockets.
- **Preconditions**: User is logged in, logging app enabled.
- **Steps**:
  1. Open the logging app UI.
  2. Trigger a test event (e.g., API call) to generate a log.
  3. Observe the log display in the UI.
- **Expected Outcome**: Log appears in the UI within 1 second, with timestamp, severity, and message.
- **Failure Case**: Log is delayed, missing, or lacks metadata.

### Test Case 4.2: AI-Driven Log Analysis
- **Description**: Ensure optional AI context analysis enhances logs.
- **Preconditions**: Logging app enabled, Ollama configured locally.
- **Steps**:
  1. Generate a log with an error (e.g., "Null pointer exception").
  2. Enable AI analysis in the logging app.
  3. View the log’s AI-generated context.
- **Expected Outcome**: AI adds context (e.g., “Possible uninitialized variable at line 42”) with tags and severity.
- **Failure Case**: AI analysis fails or provides irrelevant context.

## 5. Analytics App
### Test Case 5.1: Log Frequency Analysis
- **Description**: Verify that the analytics app calculates log frequency correctly.
- **Preconditions**: Logging app enabled, 100 sample logs generated (50 errors, 30 warnings, 20 info).
- **Steps**:
  1. Open the analytics app.
  2. Run a frequency report on log types.
- **Expected Outcome**: Report shows 50 errors, 30 warnings, 20 info, with a downloadable CSV.
- **Failure Case**: Incorrect counts or CSV export fails.

### Test Case 5.2: Anomaly Detection
- **Description**: Ensure the analytics app detects log anomalies.
- **Preconditions**: Logging app enabled, 100 logs with 5 unusual entries (e.g., rare error code).
- **Steps**:
  1. Open the analytics app.
  2. Run anomaly detection.
- **Expected Outcome**: App highlights the 5 unusual logs with details (e.g., “Rare error code detected 5 times”).
- **Failure Case**: Anomalies missed or misidentified.

## 6. Review App
### Test Case 6.1: Code Import via GitHub
- **Description**: Verify that users can import code from GitHub repos.
- **Preconditions**: User logged in with GitHub OAuth, review app enabled.
- **Steps**:
  1. Open the review app.
  2. Select a public repo (e.g., `mikejsmith1985/devsmith-modular-platform`).
  3. Choose a file and import it.
- **Expected Outcome**: File contents load in the UI with syntax highlighting.
- **Failure Case**: Import fails or file contents are corrupted.

### Test Case 6.2: Five Reading Modes
- **Description**: Ensure AI-driven analysis supports all five reading modes: Previewing, Skimming, Scanning, Detailed Reading, Critical Reading.
- **Preconditions**: Review app enabled, sample code file loaded, Ollama configured.
- **Steps**:
  1. Select each reading mode in the UI.
  2. Run AI analysis on the same code file for each mode.
- **Expected Outcomes**:
  - **Previewing**: Returns a brief summary (e.g., “This is a Python script for a REST API”).
  - **Skimming**: Lists high-level functionality (e.g., “Handles user authentication and data queries”).
  - **Scanning**: Finds specific elements (e.g., highlights variable `user_id` in 3 locations).
  - **Detailed Reading**: Explains code logic in depth (e.g., “Function `auth_user` validates JWT tokens”).
  - **Critical Reading**: Identifies issues (e.g., “Missing error handling in `auth_user` could lead to crashes”).
- **Failure Case**: Mode produces incorrect or incomplete output.

### Test Case 6.3: Real-Time Collaboration
- **Description**: Verify that multiple users can collaborate on code reviews.
- **Preconditions**: Two users logged in, review app enabled.
- **Steps**:
  1. User A loads a code file and starts a review session.
  2. User B joins the session and adds an annotation.
  3. User A responds to the annotation.
- **Expected Outcome**: Annotation appears in real-time for both users, with logged telemetry.
- **Failure Case**: Annotations fail to sync or are not visible.

## 7. Build App
### Test Case 7.1: Terminal Interface
- **Description**: Ensure the terminal supports Cloud CLI and Copilot CLI commands.
- **Preconditions**: Build app enabled, user logged in.
- **Steps**:
  1. Open the build app terminal.
  2. Run a Cloud CLI command (e.g., `gcloud version`).
  3. Run a Copilot CLI command (e.g., `gh repo list`).
- **Expected Outcome**: Commands execute successfully, with output logged to the logging app.
- **Failure Case**: Commands fail or logs are not captured.

### Test Case 7.2: OpenHands Autonomous Coding
- **Description**: Verify autonomous coding with OpenHands and Ollama.
- **Preconditions**: Build app in Phase 2, Ollama configured, sample project loaded.
- **Steps**:
  1. Initiate an OpenHands task (e.g., “Generate a Python function for sorting”).
  2. Monitor the autonomous coding process.
- **Expected Outcome**: OpenHands generates correct code, logged to the logging app, and verifiable in the review app.
- **Failure Case**: Code generation fails or produces incorrect results.

## 8. LLM Integration
### Test Case 8.1: Local Ollama Operation
- **Description**: Ensure AI features work offline with Ollama.
- **Preconditions**: Ollama configured locally, internet disabled.
- **Steps**:
  1. Enable AI analysis in the logging app.
  2. Run a code review with the review app.
- **Expected Outcome**: AI features (e.g., log context, code analysis) function without internet.
- **Failure Case**: AI features fail or require online connectivity.

### Test Case 8.2: API Toggle
- **Description**: Verify seamless switching between local and online LLMs.
- **Preconditions**: User logged in, OpenAI API key provided.
- **Steps**:
  1. Set LLM to Ollama and run a review app analysis.
  2. Switch to OpenAI API and rerun the analysis.
- **Expected Outcome**: Both analyses complete successfully with consistent outputs.
- **Failure Case**: Switching fails or outputs differ significantly.

## 9. Collaboration
### Test Case 9.1: Shared Terminal Session
- **Description**: Ensure users can co-pilot in the build app’s terminal.
- **Preconditions**: Two users logged in, build app enabled.
- **Steps**:
  1. User A starts a terminal session and invites User B.
  2. User B enters a command (e.g., `ls`).
- **Expected Outcome**: Command output is visible to both users in real-time, logged to the logging app.
- **Failure Case**: Commands or outputs do not sync.

## 10. Portal
### Test Case 10.1: Dashboard Functionality
- **Description**: Verify the portal dashboard displays session history and app access.
- **Preconditions**: User logged in, logging and review apps enabled.
- **Steps**:
  1. Open the portal dashboard.
  2. Check session history for recent activity.
  3. Launch the logging app from the app browser.
- **Expected Outcome**: Dashboard shows recent logs, and logging app launches correctly.
- **Failure Case**: History is missing or app fails to launch.

### Test Case 10.2: One-Click Installation
- **Description**: Ensure the platform installs with a single command.
- **Preconditions**: Clean system with Docker installed.
- **Steps**:
  1. Run the installation script (e.g., `docker-compose up`).
  2. Check for Postgres, Ollama, and OAuth setup.
- **Expected Outcome**: All services start successfully, and the portal is accessible.
- **Failure Case**: Installation fails or requires manual configuration.

## 11. Debugging
### Test Case 11.1: Friendly Error Messages
- **Description**: Verify that errors are clear and actionable.
- **Preconditions**: User logged in, logging app enabled.
- **Steps**:
  1. Trigger an error (e.g., invalid API key in LLM settings).
  2. Observe the error message in the UI.
- **Expected Outcome**: Error includes stack trace, context, and an AI-driven “fix this” prompt (e.g., “Invalid API key. Enter a valid OpenAI key in settings.”).
- **Failure Case**: Error is cryptic or lacks actionable guidance.

### Test Case 11.2: Backup Functionality
- **Description**: Ensure automatic snapshots preserve data.
- **Preconditions**: Logging app enabled, sample logs generated.
- **Steps**:
  1. Simulate a system crash (e.g., kill Docker container).
  2. Restart the platform and check log availability.
- **Expected Outcome**: Logs are restored from the latest snapshot.
- **Failure Case**: Logs are lost or corrupted.
