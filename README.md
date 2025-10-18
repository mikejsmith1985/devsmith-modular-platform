# [Project Name]

[Brief project description]

## Features

- Feature 1
- Feature 2
- Feature 3

## Tech Stack

**Frontend:**
- React 18
- Vite
- Bootstrap 5

**Backend:**
- FastAPI
- PostgreSQL

**Infrastructure:**
- Docker
- Docker Compose

## Quick Start

```bash
# Clone repository
git clone https://github.com/mikejsmith1985/PROJECT-NAME.git
cd PROJECT-NAME

# Start with Docker
docker-compose up -d

# Or develop locally
cd frontend && npm install && npm run dev
cd backend && pip install -r requirements.txt && uvicorn main:app --reload
```

## Development

See [STANDARDS.md](STANDARDS.md) for coding standards.

### Branch Strategy

- `main` - Production releases
- `development` - Integration branch
- `feature/*` - Feature branches

### Making Changes

1. Create feature branch: `git checkout -b feature/my-feature development`
2. Make changes
3. Update CLAUDE_CHANGELOG.md or AI_CHANGELOG.md
4. Test changes
5. Commit and push
6. Create PR to `development`

## Testing

```bash
# Frontend tests
cd frontend && npm test

# Backend tests
cd backend && pytest
```

## Documentation

- [STANDARDS.md](STANDARDS.md) - Coding standards
- [STYLING_GUIDE.md](STYLING_GUIDE.md) - Design system
- [CLAUDE_CHANGELOG.md](CLAUDE_CHANGELOG.md) - Claude AI changes
- [AI_CHANGELOG.md](AI_CHANGELOG.md) - Copilot changes

## License

MIT
