package review_db

import (
	"context"

	"github.com/google/uuid"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// GitHubRepositoryInterface defines the contract for GitHub session repository operations.
// This interface enables dependency injection and testing with in-memory implementations.
type GitHubRepositoryInterface interface {
	// Session operations
	CreateGitHubSession(ctx context.Context, session *review_models.GitHubSession) error
	GetGitHubSession(ctx context.Context, id int64) (*review_models.GitHubSession, error)
	GetGitHubSessionBySessionID(ctx context.Context, sessionID int64) (*review_models.GitHubSession, error)
	UpdateFileTree(ctx context.Context, id int64, tree []byte, totalFiles, totalDirs int) error

	// File operations
	CreateOpenFile(ctx context.Context, file *review_models.OpenFile) error
	GetOpenFiles(ctx context.Context, githubSessionID int64) ([]*review_models.OpenFile, error)
	GetOpenFileByTabID(ctx context.Context, tabID uuid.UUID) (*review_models.OpenFile, error)
	SetActiveTab(ctx context.Context, githubSessionID int64, tabID uuid.UUID) error
	CloseFile(ctx context.Context, tabID uuid.UUID) error
	IncrementAnalysisCount(ctx context.Context, tabID uuid.UUID) error

	// Multi-file analysis operations
	CreateMultiFileAnalysis(ctx context.Context, analysis *review_models.MultiFileAnalysis) error
	GetMultiFileAnalyses(ctx context.Context, githubSessionID int64) ([]*review_models.MultiFileAnalysis, error)
	GetLatestMultiFileAnalysis(ctx context.Context, githubSessionID int64) (*review_models.MultiFileAnalysis, error)
}
