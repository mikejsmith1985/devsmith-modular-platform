# Learning Development Platform: Requirements

## Repository
- Hosted at: [github.com/mikejsmith1985/devsmith-modular-platform](https://github.com/mikejsmith1985/devsmith-modular-platform)
- Built on the DevSmith reusable template with standardized branch structure and project standards.

## Overview
- **Modularity**: Fully modular architecture, allowing users to opt-in to specific apps (logging, analytics, review, build) without dependency on others. Each app operates in isolation but interoperates seamlessly when selected.
- **Ease of Use**: One-click installation (Docker-based, Homebrew, or Winget) to handle all dependencies, including Postgres setup, Ollama model downloads, and GitHub OAuth configuration. Includes health check to verify services before launch.
- **Debugging**: All endpoints must have high-level debugging with friendly, actionable error messages, including stack traces, context dumps, retry options, and AI-driven "fix this" prompts via Ollama. No cryptic errors (e.g., 504s).
- **Build Order**: Develop portal first (home screen with app browser), followed by logging app, which will monitor development of subsequent apps (analytics, review, build).
- **Hardware Requirements**: Minimum 8GB RAM, modern CPU. Optional GPU with 16GB VRAM for local LLM performance. Macs with M1 or later (32GB RAM) supported without GPU.
- **Backups**: Automatic snapshots of logs, code states, and model configurations to prevent data loss.

## Authentication
- GitHub OAuth for login, providing secure access to user repositories and profile data via the DevSmith platform.

## Modularity
- Each app is toggleable, with independent database schemas but capable of federated queries for cross-app insights.
- Apps function standalone or in concert, with user control over active components via the portal.

## Database
- Default: Postgres for robust, relational data storage, aligned with DevSmith standards.
- Pluggable architecture to support future integration of MongoDB or SQLite for lightweight use cases.
- Schemas isolated per app but designed for cross-app querying when needed.

## Logging App
- Real-time log tracking via WebSockets (or optimized alternative) for live debugging.
- Optional AI-driven context analysis (via Ollama) for logs, including tags, severity levels, and historical data pulled from the database.
- Configurable to capture metadata for debugging and analytics, feeding into the analytics app.

## Analytics App
- Analyzes logs for frequency, trends, severity heatmaps, and anomaly detection.
- Provides tag-based filtering and exportable reports (e.g., CSV).
- Surfaces actionable statistics, such as most common log types or error patterns.
- Optional AI-driven insights for deeper analysis (e.g., pattern recognition, predictive alerts).

## Review App
- Supports code import via GitHub (browse user or public repos from `mikejsmith1985/devsmith-modular-platform` or others) or direct paste into UI.
- AI-driven code analysis with five reading modes, each with distinct purposes and approaches:
  - **Previewing**: Quick once-over to get the gist of a codebase with minimal investment. Use case: Exploring a GitHub repo for the first time to assess relevance or interest. Provides surface-level understanding only.
  - **Skimming**: Understands functionality at a high level without deep detail. Use case: Clicking through files to grasp what the code does generally, without intent to modify. One level deeper than previewing.
  - **Scanning**: Targeted search for specific information. Use case: Debugging to locate a variable or error (e.g., null pointer exception). Task-oriented, focused on finding particular code elements.
  - **Detailed Reading**: Deep dive into specific code to understand algorithms and inputs thoroughly. Use case: Explaining exactly what code does in granular detail. Infrequent but demanding; less common in daily engineering tasks.
  - **Critical Reading**: Evaluative, mentally demanding mode to identify weaknesses and improvements. Use case: Reviewing another engineer’s or LLM-generated code, or pre-production reviews to mitigate risk. Mindset: “How can I break this?” Focuses on quality assurance, bug detection, security issues, and refactoring suggestions.
- Features syntax highlighting, bug detection, security issue flagging, and refactoring suggestions tailored to each reading mode.
- Real-time collaboration: multiple users can view, highlight, annotate, and comment on code together.
- Integrates with logging app to track review session telemetry.

## Build App
- **Phase 1**: Terminal interface supporting Cloud CLI and Copilot CLI, enabling full development, debugging, and learning within the platform.
- Logs from terminal feed directly into the logging app for real-time monitoring.
- **Phase 2**: Autonomous coding via OpenHands, powered by local Ollama models for secure, offline builds.
- Integrates with review and analytics apps for seamless code-to-log-to-stats workflow.

## LLM Integration
- **Local Option**: Ollama for offline, secure LLM operation, supporting all AI-driven features (logging context, review analysis, build automation).
- **Online Option**: Toggle to external APIs (defaults: OpenAI, Anthropic, Together) with user-provided API keys for flexibility.
- Configurable via UI to switch between local and online models effortlessly.

## Collaboration
- Real-time collaboration in the review app for shared code reviews, annotations, and discussions.
- Optional shared terminal sessions in the build app for co-piloting or mentoring.
- Session history preserved and accessible via the portal.

## Portal
- Clean, intuitive dashboard with app browser for selecting and launching apps.
- Displays session history, recent logs, and quick access to active apps.
- Integrates GitHub OAuth for seamless authentication and repository access.
