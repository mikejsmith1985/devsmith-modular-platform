package review_db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// InMemoryGitHubRepository implements GitHubRepository interface for testing
type InMemoryGitHubRepository struct {
	mu              sync.RWMutex
	sessions        map[int64]*review_models.GitHubSession
	openFiles       map[int64]*review_models.OpenFile
	analyses        map[int64]*review_models.MultiFileAnalysis
	nextSessionID   int64
	nextFileID      int64
	nextAnalysisID  int64
	sessionFiles    map[int64][]int64 // session_id -> []file_ids
	sessionAnalyses map[int64][]int64 // session_id -> []analysis_ids
}

// NewInMemoryGitHubRepository creates a new in-memory repository for testing
func NewInMemoryGitHubRepository() *InMemoryGitHubRepository {
	return &InMemoryGitHubRepository{
		sessions:        make(map[int64]*review_models.GitHubSession),
		openFiles:       make(map[int64]*review_models.OpenFile),
		analyses:        make(map[int64]*review_models.MultiFileAnalysis),
		sessionFiles:    make(map[int64][]int64),
		sessionAnalyses: make(map[int64][]int64),
		nextSessionID:   1,
		nextFileID:      1,
		nextAnalysisID:  1,
	}
}

// CreateSession creates a new GitHub session
func (r *InMemoryGitHubRepository) CreateSession(ctx context.Context, session *review_models.GitHubSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.ID = r.nextSessionID
	r.nextSessionID++
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	r.sessions[session.ID] = session
	return nil
}

// GetSession retrieves a session by ID
func (r *InMemoryGitHubRepository) GetSession(ctx context.Context, sessionID int64) (*review_models.GitHubSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %d", sessionID)
	}
	return session, nil
}

// UpdateSession updates an existing session
func (r *InMemoryGitHubRepository) UpdateSession(ctx context.Context, session *review_models.GitHubSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %d", session.ID)
	}

	session.UpdatedAt = time.Now()
	r.sessions[session.ID] = session
	return nil
}

// DeleteSession deletes a session
func (r *InMemoryGitHubRepository) DeleteSession(ctx context.Context, sessionID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, sessionID)
	delete(r.sessionFiles, sessionID)
	delete(r.sessionAnalyses, sessionID)
	return nil
}

// ListSessions lists all sessions for a user (not supported in in-memory repo - used for demo)
func (r *InMemoryGitHubRepository) ListSessions(ctx context.Context, userID int64) ([]*review_models.GitHubSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*review_models.GitHubSession
	for _, session := range r.sessions {
		// In-memory repo doesn't track user_id on session (not in model)
		// Just return all sessions for testing
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// OpenFile creates a new open file entry
func (r *InMemoryGitHubRepository) OpenFile(ctx context.Context, file *review_models.OpenFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file.ID = r.nextFileID
	r.nextFileID++
	file.OpenedAt = time.Now()

	r.openFiles[file.ID] = file
	r.sessionFiles[file.GitHubSessionID] = append(r.sessionFiles[file.GitHubSessionID], file.ID)
	return nil
}

// GetOpenFile retrieves an open file by ID
func (r *InMemoryGitHubRepository) GetOpenFile(ctx context.Context, fileID int64) (*review_models.OpenFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, exists := r.openFiles[fileID]
	if !exists {
		return nil, fmt.Errorf("file not found: %d", fileID)
	}
	return file, nil
}

// ListOpenFiles lists all open files for a session
func (r *InMemoryGitHubRepository) ListOpenFiles(ctx context.Context, sessionID int64) ([]*review_models.OpenFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fileIDs := r.sessionFiles[sessionID]
	files := make([]*review_models.OpenFile, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		if file, exists := r.openFiles[fileID]; exists {
			files = append(files, file)
		}
	}
	return files, nil
}

// closeFileByID marks a file as closed (internal method using int64 ID)
func (r *InMemoryGitHubRepository) closeFileByID(ctx context.Context, fileID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, exists := r.openFiles[fileID]
	if !exists {
		return fmt.Errorf("file not found: %d", fileID)
	}

	// Remove from session files
	sessionFiles := r.sessionFiles[file.GitHubSessionID]
	for i, id := range sessionFiles {
		if id == fileID {
			r.sessionFiles[file.GitHubSessionID] = append(sessionFiles[:i], sessionFiles[i+1:]...)
			break
		}
	}
	delete(r.openFiles, fileID)
	return nil
}

// SetActiveFile sets the active file for a session
func (r *InMemoryGitHubRepository) SetActiveFile(ctx context.Context, sessionID, fileID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Deactivate all files for this session
	for _, id := range r.sessionFiles[sessionID] {
		if file, exists := r.openFiles[id]; exists {
			file.IsActive = false
		}
	}

	// Activate the target file
	file, exists := r.openFiles[fileID]
	if !exists {
		return fmt.Errorf("file not found: %d", fileID)
	}
	file.IsActive = true
	return nil
}

// CreateAnalysis creates a new multi-file analysis
func (r *InMemoryGitHubRepository) CreateAnalysis(ctx context.Context, analysis *review_models.MultiFileAnalysis) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	analysis.ID = r.nextAnalysisID
	r.nextAnalysisID++
	analysis.CreatedAt = time.Now()

	r.analyses[analysis.ID] = analysis
	r.sessionAnalyses[analysis.GitHubSessionID] = append(r.sessionAnalyses[analysis.GitHubSessionID], analysis.ID)
	return nil
}

// GetAnalysis retrieves an analysis by ID
func (r *InMemoryGitHubRepository) GetAnalysis(ctx context.Context, analysisID int64) (*review_models.MultiFileAnalysis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	analysis, exists := r.analyses[analysisID]
	if !exists {
		return nil, fmt.Errorf("analysis not found: %d", analysisID)
	}
	return analysis, nil
}

// UpdateAnalysis updates an existing analysis
func (r *InMemoryGitHubRepository) UpdateAnalysis(ctx context.Context, analysis *review_models.MultiFileAnalysis) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.analyses[analysis.ID]; !exists {
		return fmt.Errorf("analysis not found: %d", analysis.ID)
	}

	r.analyses[analysis.ID] = analysis
	return nil
}

// ListAnalyses lists all analyses for a session
func (r *InMemoryGitHubRepository) ListAnalyses(ctx context.Context, sessionID int64) ([]*review_models.MultiFileAnalysis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	analysisIDs := r.sessionAnalyses[sessionID]
	analyses := make([]*review_models.MultiFileAnalysis, 0, len(analysisIDs))
	for _, id := range analysisIDs {
		if analysis, exists := r.analyses[id]; exists {
			analyses = append(analyses, analysis)
		}
	}
	return analyses, nil
}

// UpdateTreeCache updates the cached tree for a session
func (r *InMemoryGitHubRepository) UpdateTreeCache(ctx context.Context, sessionID int64, treeSHA string, treeData []byte, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %d", sessionID)
	}

	session.CommitSHA = treeSHA
	session.FileTree = treeData
	session.TreeLastSynced = time.Now()
	session.UpdatedAt = time.Now()
	return nil
}

// GetTreeCache retrieves the cached tree for a session
func (r *InMemoryGitHubRepository) GetTreeCache(ctx context.Context, sessionID int64) ([]byte, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, false, fmt.Errorf("session not found: %d", sessionID)
	}

	if session.FileTree == nil {
		return nil, false, nil
	}

	// Check if cache is expired (24 hour TTL)
	if time.Since(session.TreeLastSynced) > 24*time.Hour {
		return nil, false, nil // Cache expired
	}

	return session.FileTree, true, nil
}

// InvalidateTreeCache invalidates the cached tree for a session
func (r *InMemoryGitHubRepository) InvalidateTreeCache(ctx context.Context, sessionID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %d", sessionID)
	}

	session.FileTree = nil
	session.TreeLastSynced = time.Time{}
	session.UpdatedAt = time.Now()
	return nil
}

// ========== Interface Adapter Methods ==========
// These methods adapt the in-memory repository to match the GitHubRepositoryInterface

func (r *InMemoryGitHubRepository) CreateGitHubSession(ctx context.Context, session *review_models.GitHubSession) error {
	return r.CreateSession(ctx, session)
}

func (r *InMemoryGitHubRepository) GetGitHubSession(ctx context.Context, id int64) (*review_models.GitHubSession, error) {
	return r.GetSession(ctx, id)
}

func (r *InMemoryGitHubRepository) GetGitHubSessionBySessionID(ctx context.Context, sessionID int64) (*review_models.GitHubSession, error) {
	// In the in-memory version, the ID and SessionID are the same
	return r.GetSession(ctx, sessionID)
}

func (r *InMemoryGitHubRepository) UpdateFileTree(ctx context.Context, id int64, tree []byte, totalFiles, totalDirs int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %d", id)
	}

	session.FileTree = tree
	session.TotalFiles = totalFiles
	session.TotalDirectories = totalDirs
	session.TreeLastSynced = time.Now()
	session.UpdatedAt = time.Now()
	return nil
}

func (r *InMemoryGitHubRepository) CreateOpenFile(ctx context.Context, file *review_models.OpenFile) error {
	return r.OpenFile(ctx, file)
}

func (r *InMemoryGitHubRepository) GetOpenFiles(ctx context.Context, githubSessionID int64) ([]*review_models.OpenFile, error) {
	return r.ListOpenFiles(ctx, githubSessionID)
}

func (r *InMemoryGitHubRepository) GetOpenFileByTabID(ctx context.Context, tabID uuid.UUID) (*review_models.OpenFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, file := range r.openFiles {
		if file.TabID == tabID {
			return file, nil
		}
	}

	return nil, fmt.Errorf("file not found with tab_id: %s", tabID)
}

func (r *InMemoryGitHubRepository) SetActiveTab(ctx context.Context, githubSessionID int64, tabID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Set all files in this session to inactive
	for _, file := range r.openFiles {
		if file.GitHubSessionID == githubSessionID {
			file.IsActive = false
		}
	}

	// Find and activate the target file
	for _, file := range r.openFiles {
		if file.TabID == tabID {
			file.IsActive = true
			file.LastAccessed = time.Now()
			return nil
		}
	}

	return fmt.Errorf("file not found with tab_id: %s", tabID)
}

func (r *InMemoryGitHubRepository) CloseFile(ctx context.Context, tabID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, file := range r.openFiles {
		if file.TabID == tabID {
			// Remove from session files
			if fileIDs, ok := r.sessionFiles[file.GitHubSessionID]; ok {
				for i, fid := range fileIDs {
					if fid == id {
						r.sessionFiles[file.GitHubSessionID] = append(fileIDs[:i], fileIDs[i+1:]...)
						break
					}
				}
			}
			delete(r.openFiles, id)
			return nil
		}
	}

	return fmt.Errorf("file not found with tab_id: %s", tabID)
}

func (r *InMemoryGitHubRepository) IncrementAnalysisCount(ctx context.Context, tabID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, file := range r.openFiles {
		if file.TabID == tabID {
			file.AnalysisCount++
			file.LastAccessed = time.Now()
			return nil
		}
	}

	return fmt.Errorf("file not found with tab_id: %s", tabID)
}

func (r *InMemoryGitHubRepository) CreateMultiFileAnalysis(ctx context.Context, analysis *review_models.MultiFileAnalysis) error {
	return r.CreateAnalysis(ctx, analysis)
}

func (r *InMemoryGitHubRepository) GetMultiFileAnalyses(ctx context.Context, githubSessionID int64) ([]*review_models.MultiFileAnalysis, error) {
	return r.ListAnalyses(ctx, githubSessionID)
}

func (r *InMemoryGitHubRepository) GetLatestMultiFileAnalysis(ctx context.Context, githubSessionID int64) (*review_models.MultiFileAnalysis, error) {
	analyses, err := r.ListAnalyses(ctx, githubSessionID)
	if err != nil {
		return nil, err
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no analyses found for session: %d", githubSessionID)
	}

	// Return the most recently created analysis
	latest := analyses[0]
	for _, a := range analyses {
		if a.CreatedAt.After(latest.CreatedAt) {
			latest = a
		}
	}

	return latest, nil
}
