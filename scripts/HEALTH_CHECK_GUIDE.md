# Health-check CLI: Usage & Guidance

Quick verdict

- Use `./scripts/health-check-cli.sh` (watch mode) for iterative development and debugging. It's faster and provides real-time diagnostics.
- Keep `./scripts/docker-validate.sh` as the comprehensive pre-PR gate.

Why the health-check CLI is better for development

- Faster feedback loop: lightweight checks (container status, HTTP health endpoints, DB readiness) return quickly.
- Continuous monitoring (--watch): watch mode streams changes so you see when services become healthy.
- Targeted diagnostics: run only the checks you need (JSON, per-service, quick mode).
- Developer UX: eliminates repeated polling (no need to run `docker-compose ps` constantly).
- Automation friendly: JSON output for parsing by scripts or CI when needed.

Where `docker-validate.sh` still shines

- Use it in pre-PR/CI for a full validation suite (endpoint checks, timeouts, aggregated validation report).
- Keep it as the final gate before creating or merging PRs.

Concrete recommendations

1. Dev workflow (interactive)

```bash
# Start the stack (if not already running)
docker-compose up -d

# Monitor services in real-time while you develop
./scripts/health-check-cli.sh --watch

# Quick one-off checks
./scripts/health-check-cli.sh        # human-readable
./scripts/health-check-cli.sh --json # machine-readable
```

2. Pre-PR / full validation

```bash
# Run the thorough validation before creating a PR
./scripts/docker-validate.sh
```

3. Add a small helper wrapper (optional)

Place this wrapper in `scripts/dev-health.sh` (executable) for convenience:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Start the stack, then watch health checks
docker-compose up -d
./scripts/health-check-cli.sh --watch
```

Make executable:

```bash
chmod +x scripts/dev-health.sh
```

4. VS Code / IDE convenience

- Add a task that runs `./scripts/health-check-cli.sh --watch` in an integrated terminal while you develop.

5. Optional pre-commit/soft check (non-blocking)

- You can add a non-blocking quick health check to developer hooks to detect obvious problems early.

Integration notes & small edits I can make

- I can add `scripts/dev-health.sh` and a short README snippet to the project (safe, reversible).
- I can update `scripts/dev.sh` to call the CLI in interactive developer mode if you want that.

Suggested commit message when adding the wrapper & README note:

```
chore(dev): add health-check CLI dev wrapper and documentation

- Add scripts/dev-health.sh to start stack and run health-check CLI in --watch mode
- Add scripts/HEALTH_CHECK_GUIDE.md with guidance and recommended workflows
```

Engineering checklist / rationale

- Inputs/outputs: health-check CLI checks container status, HTTP endpoints, DB readiness; outputs human-readable and JSON modes.
- Error modes: CLI helps spot startup problems quickly (nginx, Postgres); `docker-validate.sh` catches harder-to-reproduce endpoint timeouts.
- Edge cases: For flaky services (maildev, nginx) `--watch` helps spot intermittent failures during fixes.

Next steps (pick one)

- (A) I will add `scripts/dev-health.sh` and `scripts/HEALTH_CHECK_GUIDE.md` and commit them now.
- (B) I will only create the `scripts/HEALTH_CHECK_GUIDE.md` and leave wrapper for later.
- (C) You keep the repo as-is and I'll continue using CLI interactively (no repo changes).

If you want me to commit the wrapper and README, say "Do it" and I'll create the wrapper file, make it executable, and commit to the current branch.
