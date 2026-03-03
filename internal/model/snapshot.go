package model

import (
	"time"
)

// Snapshot represents the state of a repository at a point in time.
type Snapshot struct {
	RepoRoot  string      `json:"repo_root"`
	Timestamp time.Time   `json:"timestamp"`
	Files     []*FileItem `json:"files"`
	Version   string      `json:"repoan_version"`
}

// FileItem represents a single file or directory in the repository.
type FileItem struct {
	Path      string      `json:"path"`      // Relative path from repo root
	Name      string      `json:"name"`
	IsDir     bool        `json:"is_dir"`
	Size      int64       `json:"size"`      // Size in bytes
	Extension string      `json:"extension"` // File extension (if any)
	Children  []*FileItem `json:"children,omitempty"`
}
