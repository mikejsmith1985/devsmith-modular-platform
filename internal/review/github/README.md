# GitHub Integration for Review Service

## Overview

The GitHub client provides GitHub API integration for the Review Service, enabling code fetching from GitHub repositories directly into review sessions.

**Issue #27 Implementation**: GitHub Integration & Code Fetching

## Components

### ClientInterface

Defines standard operations for GitHub integration:
- `FetchCode()`: Retrieve code from repository
- `GetRepoMetadata()`: Get repository information
- `ValidateURL()`: Parse and validate GitHub URLs
- `GetRateLimit()`: Check API rate limits

### DefaultClient

Implements ClientInterface with:
- **URL Parsing**: Supports HTTPS and SSH formats
- **Validation**: Ensures valid GitHub URLs
- **Metadata**: Repository information retrieval
- **Rate Limiting**: API quota checking

### Supported URL Formats

1. **HTTPS**: `https://github.com/owner/repo[.git]`
2. **SSH**: `git@github.com:owner/repo[.git]`

## Error Types

- `URLParseError`: Invalid URL format
- `AuthError`: Authentication failed (401)
- `NotFoundError`: Repository not found (404)
- `RateLimitError`: API rate limit exceeded

## Usage

```go
client := github.NewDefaultClient()

// Validate GitHub URL
owner, repo, err := client.ValidateURL("https://github.com/golang/go")

// Fetch code
code, err := client.FetchCode(ctx, owner, repo, "main", token)

// Get repository metadata
meta, err := client.GetRepoMetadata(ctx, owner, repo, token)

// Check rate limits
remaining, resetTime, err := client.GetRateLimit(ctx, token)
```

## Implementation Status

### Current (MVP)
- ✅ URL validation and parsing
- ✅ Stub implementations for API calls
- ✅ Error handling
- ✅ Rate limit interface

### Future Enhancements
- Real GitHub API integration (using github.com/google/go-github)
- OAuth token validation
- Caching via Issue #26 cache layer
- Branch/commit/PR selection
- Private repository support
- Actual code fetching from GitHub API

## Testing

```bash
go test -v ./internal/review/github/...
```

All 10 tests passing - validates URL parsing and error handling.

## Integration with Review Service

The GitHub client integrates with review sessions:
1. User provides GitHub URL when creating session
2. URL is validated by GitHub client
3. Code is fetched using user's GitHub token (from Portal auth)
4. Session is created with fetched code
5. Code is ready for analysis through review modes
