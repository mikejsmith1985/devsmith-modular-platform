# Review 1.1 — Redesigned Review UX & Implementation Notes

Date: 2025-11-01

Summary
-------
This document captures the requested UX redesign for the Review app ("Review 1.1"), the concrete fixes discovered while debugging the current implementation, and a prioritized implementation + testing plan. The goal is a simpler, clearer review experience where the user can: pick a mode, load a repo or paste code, view the code tree, and get AI analysis in a two-pane workspace without confusing demo content or hidden inputs.

High-level goals
----------------
- Keep the UI focused and action-oriented: the main entrypoint should lead into the two-pane Review workspace (detailed below).
- Preserve a small, aesthetic demo/example card (separate from the mode cards) that contains a real sample code file (recommended: `cmd/portal/main.go`) so users can quickly run an example analysis on a known piece of code.
- Provide a two-pane workspace (left = code selection or repo browser; right = model selector, context view, analysis output).
- Optionally, support a mode dropdown on the main page so users can remain on one screen and move from skim → critical without page switches.
- Fix backend and tests so the UI reflects true AI analysis and tests verify the functional output (not just HTML presence).

User flow (final UX)
--------------------
Primary flow (two-pane modal/page):

1. User arrives at /review. They see a clean row/grid of mode cards only (no example/demo content under them).
2. User selects a mode (either by clicking the mode card or via a mode dropdown). That opens the Review workspace in a new window, tab, or in-page two-pane panel (depending on configuration).
3. Workspace layout (two panes):
   - Left pane (Code Source):
     - Tabs: "Paste" | "Upload" | "GitHub Repo"
     - Paste: a large code editor textarea with syntax highlighting and a file name field.
     - Upload: a file selector with drag-and-drop.
     - GitHub Repo: repo URL input + branch selector + repository file tree browser. When a file is selected from the tree it becomes the focused target for analysis.
   - Right pane (Analysis & Model):
     - Model selector dropdown (list of available models) showing currently selected model and a short description.
     - Context viewer: shows a scroll-synced snippet/context of the selected file or pasted code; lines can be highlighted by the analysis.
     - Analysis area: results are displayed here (structured header and expandable sections per mode: Preview summary, Skim signatures, Scan matches, Detailed line-by-line, Critical issues).
     - Actions: Re-run with different model, Download result, Open in editor (optional).

Alternative (single-page mode-dropdown):
- Keep the main review page as the workspace (no new window). Provide a prominent mode dropdown at the top-left of the workspace so the user may switch modes (skim → critical) without changing pages. This keeps the workflow tight and reduces context switching.

Why this is better
-------------------
- Eliminates the noisy demo examples that led to confusion about live analysis vs placeholders.
- Makes it clear when analysis is being run and what inputs are used (model, file/repo/branch).
- Allows fluid transitions across modes for the same file/repo without losing state.

Fixes discovered and included in this doc
---------------------------------------
The following issues were discovered while validating the current implementation. These are included in the Review 1.1 plan and must be addressed:

1. Missing model in mode POST requests
   - Cause: `ModeSelector` buttons' `hx-include` omitted `[name='model']`. Result: user-chosen model wasn't sent with mode requests.
   - Fix: Include `[name='model']` in `hx-include` for all mode buttons, and ensure the session form contains `name="model"` field.

2. Services returning mock/fallback output
   - Cause: Some services are constructed without an Ollama-enabled client (use of `NewPreviewService(...)` vs `NewPreviewServiceWithOllama(...)`) or Ollama calls failed and services fell back to mock data.
   - Fix: Ensure all services use Ollama adapter (`NewXxxServiceWithOllama(...)` or pass adapter) and improve AI error handling (return explicit UI error + retry option). Add metrics/logging around AI call durations and failures.

3. UI contains demo/mock results in `home.templ`
   - Cause: Template includes static mock examples; they looked like real output.
   - Fix: Relocate demo content to a dedicated, clearly-labeled "Sample Project" card (or an Examples page). The main page should show only mode cards and a visible launcher for the two-pane workspace. The sample project card should reference a real file (recommended: `cmd/portal/main.go`) and offer a "Run sample" action that preloads that file into the workspace.

4. Test gaps: E2E asserts on HTML existence only
   - Cause: Playwright tests checked that a container had text content or that a DOM element exists; mock output makes these tests pass even when analysis is not performed.
   - Fix: Update tests to assert structured, mode-specific content (e.g., Preview must include "File Tree" entries, Skim must include function signatures, Critical must include an issue list with severities). Use `page.waitForResponse` and then inspect the response body for JSON/HTML markers.

5. Missing model override propagation
   - Cause: Adapter expected model in context; UI did not provide it; sometimes handlers used default model.
   - Fix: Ensure `bindCodeRequest` extracts `model` and handlers call `context.WithValue(..., modelContextKey, req.Model)`. On the adapter, read typed context key and use it when building the request to Ollama.

6. DB connection pooling
   - Cause: Default DB usage caused connection exhaustion under load.
   - Fix: Limit `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and `SetConnMaxIdleTime` per service (already added for review/portal/logs). Ensure Compose Postgres max_connections is tuned for per-service limits.

7. SSE progress is a simulator
   - Cause: Session progress SSE currently uses a simulated ticker rather than real task progress.
   - Fix: Integrate the real analysis pipeline (job queue or background worker) to stream progress updates; keep simulator for demos.

Acceptance criteria (Review 1.1)
-------------------------------
The change is done when the following are true:

- [ ] The mode cards have no demo content beneath them on `/review`.
-- [ ] The workspace can be opened by selecting a mode (card click or mode dropdown) and by the sample project's "Run sample" action.
- [ ] The workspace has a two-pane layout (left: source/tree/editor; right: model selector/context/analysis) or a single-page workspace with a mode dropdown (decide before implementation).
- [ ] The chosen model is always sent with mode requests and displayed in the result header.
- [ ] Services use Ollama adapter for real analysis. When Ollama fails, an explicit UI error and a retry button are shown; fallback mock content is only used for demo mode or when a special toggle is set.
- [ ] Playwright E2E tests validate functional analysis (structured content) for at least Preview, Skim, and Critical modes.
- [ ] Unit tests cover `bindCodeRequest` binding (form + JSON + multipart) and adapter `Generate` use of context model override.

Implementation plan (phases)
---------------------------
Phase 1 — Spec + tests (small, fast)
- Create `Review 1.1` spec (this doc). ✅
- Add unit tests for `bindCodeRequest` and adapter model override.
- Update Playwright tests to assert structured results and change selectors to use `hx-post` where necessary.

Phase 2 — UX changes and wiring (medium)
- Update `apps/review/templates/home.templ`:
  - Remove `@PreviewModeResult` demo sections from home.
  - Modify `ModeSelector()` buttons to:
    - Use `hx-include="[name='pasted_code'], [name='github_url'], [name='file'], [name='model']"`.
    - Use `role="button"` and ensure the entire card is clickable (wrap the card in `<button>` or add JS to handle card click).
    - Change label to `Start <Mode> Analysis`.
- Add a two-pane workspace template (`apps/review/templates/workspace.templ`) or modify `home.templ` to support mode dropdown + workspace embedding.

Phase 3 — Backend & reliability (medium)
- Ensure all review services are constructed with the Ollama adapter (no nil clients).
- Improve Ollama adapter to read typed context key for `model` and to log the actual model used in responses.
- Add robust error handling and a retry endpoint for analyses.
- Replace SSE simulator with a real job progress stream (using a queue or goroutine + channels). Start with a background worker that runs the service call and writes progress to an in-memory channel for SSE.

Phase 4 — Testing & roll-out (short)
- Update Playwright E2E tests to cover the full workspace flow: open workspace, paste code or select repo/branch/file, select model, run analysis, assert structured output and highlights.
- Add curl-based API integration tests for each mode.
- Run tests and fix any flakiness (increase timeouts for real AI calls; skip long tests in CI by default).

Playwright test guidance (what to assert)
--------------------------------------
- Preview mode: response contains `File Tree` header and at least one file name. The result area includes a `summary` paragraph of > 50 chars.
- Skim mode: response contains a list of function names/signatures and interfaces (assert the presence of `func ` patterns in the HTML or a JSON blob in a data attribute).
- Scan mode: response includes `Matches` or `Search Results` and at least one code snippet with context (3 lines before/after).
- Detailed mode: response includes `Line-by-Line` or an explanation block per line and an algorithm summary (if applicable).
- Critical mode: response includes an issue list with severities (Critical/Important/Minor) and suggested fixes (assert presence of the word `parameterized` for SQL injection suggestions when the input includes a vulnerable query).

Notes on testing performance and flakiness
---------------------------------------
- Real AI calls can be slow; tests should set reasonable timeouts (e.g., 2-5 minutes for critical long analyses) or use an Ollama test/mock endpoint for CI.
- Add a test mode toggle or environment variable to let CI use deterministic mock responses while local dev can validate live AI behavior.

Deliverables for Review 1.1
--------------------------
- This design doc (this file).
- Template changes: `home.templ` modifications, new `workspace.templ`.
- Backend: ensure adapter wiring and model override propagation; add progress job worker scaffolding.
- Tests: new unit tests and updated Playwright tests.

Rollout checklist
-----------------
- [ ] Merge design and unit tests to `development` branch.
- [ ] Deploy review image to development via `docker-compose up -d --build review` and run full e2e smoke suite.
- [ ] Confirm user flows manually: paste, upload, GitHub → branch selection → file selection → model select → analysis results + highlights.
- [ ] Iterate on UX per feedback.

Design language and look & feel
--------------------------------
You asked for the Review app to match the look-and-feel of the `devsmith-logs` application. Important constraints and preferences:

- Keep `templ` and `htmx` (server-side rendering + small JS) — avoid a heavy client-side framework.
- Preferred CSS toolkit: Tailwind (current) — prefer to stay with Tailwind unless a desired UI pattern cannot be implemented reasonably. Bootstrap is an acceptable fallback if a required UI component or aesthetic is dramatically easier to achieve.
- The implementation will begin with a component audit of `devsmith-logs` (header, sidebar, cards, typographic scale, color system) and mapping those components to Tailwind utility classes used in Review templates.

Model-driven design audit (recommended first step)
-------------------------------------------------
Plan: use a higher-tier LLM (as you requested) to analyze `devsmith-logs` templates and CSS and produce a component inventory and Tailwind theme recommendations for Review. The audit output will include:

- Component inventory (header, nav, card, sidebar, form elements) with file references.
- Tailwind class mappings and a suggested color/spacing scale.
- Annotated diffs for templates showing where to swap classes and where to add component wrappers.

Why this helps:
- Speeds up the visual refactor by producing precise, copy-paste-ready class and template updates.
- Reduces guesswork and ensures Review visually aligns with `devsmith-logs` while keeping server-side rendering.

Governance & timeline (estimate)
--------------------------------
- Design audit (LLM-driven): 1 day
- Template refactor & demo card: 1-2 days
- Backend wiring & unit tests: 1 day
- Playwright test updates and runs: 1 day (longer if AI calls slow down tests)
- Final visual polish & accessibility checks: 1 day

Total: 4–6 working days to ship Review 1.1 to development (dependent on model-assisted design output quality and review cycles).

Next step choice
----------------
Please pick one to start:

1. "Go audit" — Run the LLM-driven design audit and produce `design-audit.md` with component inventory and Tailwind theme recommendations. (Recommended)
2. "Implement now" — Immediately implement `hx-include` + sample demo card + workspace templ scaffold and rebuild the review service.

Reply with "Go audit" or "Implement now" and I'll start right away.

Appendix: Example HTML snippets (developer hints)
-----------------------------------------------
Mode button: ensure model is included

```html
<button
  hx-post="/api/review/modes/preview"
  hx-target="#results-container"
  hx-swap="innerHTML"
  hx-include="[name='pasted_code'], [name='github_url'], [name='file'], [name='model']"
  hx-indicator="#progress-indicator-container"
>
  Start Preview Analysis
</button>
```

Two-pane workspace (struct outline)

- Left pane: `#source-pane`
  - Paste editor (`textarea[name='pasted_code']`)
  - File tree (`#repo-tree`)
  - Branch selector
- Right pane: `#config-pane`
  - Model selector (`select[name='model']`)
  - Context viewer (`#context-viewer`)
  - Results container (`#results-container`)

Contact
-------
If this matches your intent, I can:
- implement the template changes (remove demo sections, make cards clickable, add hx-include model) and rebuild the review service now, or
- create the workspace templates and wire the front-end scaffolding, or
- update Playwright tests to the structured assertions listed above.

Which do you want me to do next? (I recommend: 1) remove demo content + make cards clickable + include model in hx-include, 2) update Playwright tests to assert structured analysis, 3) then implement two-pane workspace.)
