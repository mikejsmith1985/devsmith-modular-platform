package version

import (
	"testing"
)

func TestCacheBuster(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := CommitHash
	origBuild := BuildTime
	defer func() {
		Version = origVersion
		CommitHash = origCommit
		BuildTime = origBuild
	}()

	tests := []struct {
		name        string
		version     string
		commitHash  string
		buildTime   string
		wantPattern string // regex pattern to match
	}{
		{
			name:        "production version",
			version:     "v0.1.0",
			commitHash:  "abc1234567890",
			buildTime:   "2025-11-12T10:00:00Z",
			wantPattern: "v0.1.0-abc1234",
		},
		{
			name:        "development version",
			version:     "dev",
			commitHash:  "xyz9876543210",
			buildTime:   "2025-11-12T10:00:00Z",
			wantPattern: "dev-xyz9876-20251112",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			CommitHash = tt.commitHash
			BuildTime = tt.buildTime

			got := CacheBuster()
			if got != tt.wantPattern {
				t.Errorf("CacheBuster() = %v, want %v", got, tt.wantPattern)
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	Version = "v1.0.0"
	CommitHash = "abc1234"

	got := QueryParam()
	want := "v=v1.0.0-abc1234"

	if got != want {
		t.Errorf("QueryParam() = %v, want %v", got, want)
	}
}

func TestShortVersion(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		commitHash string
		want       string
	}{
		{
			name:       "production",
			version:    "v1.0.0",
			commitHash: "abc1234",
			want:       "v1.0.0",
		},
		{
			name:       "development",
			version:    "dev",
			commitHash: "xyz9876",
			want:       "dev-xyz9876",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			CommitHash = tt.commitHash

			got := ShortVersion()
			if got != tt.want {
				t.Errorf("ShortVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
