package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repoan-test-*")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: No .git directory
	root, err := FindRoot(tempDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if root != "" {
		t.Errorf("expected empty root, got %s", root)
	}

	// Test 2: .git in current directory
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("could not create .git dir: %v", err)
	}
	root, err = FindRoot(tempDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if root != tempDir {
		t.Errorf("expected root %s, got %s", tempDir, root)
	}

	// Test 3: .git in parent directory
	subDir := filepath.Join(tempDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("could not create subdir: %v", err)
	}
	root, err = FindRoot(subDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if root != tempDir {
		t.Errorf("expected root %s, got %s", tempDir, root)
	}
}
