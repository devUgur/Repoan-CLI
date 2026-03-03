package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repoan-scan-test-*")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some files and directories
	files := []string{
		"file1.txt",
		"dir1/file2.txt",
		"dir1/subdir1/file3.txt",
		"node_modules/some-package/index.js",
		".gitignore",
	}

	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("could not create dir for %s: %v", f, err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("could not create file %s: %v", f, err)
		}
	}

	// Create .gitignore
	gitignoreContent := "node_modules/\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("could not create .gitignore: %v", err)
	}

	t.Run("Scan with gitignore", func(t *testing.T) {
		opts := ScanOptions{
			Root:             tempDir,
			RespectGitignore: true,
		}
		root, err := Scan(opts)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Check if node_modules is ignored
		for _, child := range root.Children {
			if child.Name == "node_modules" {
				t.Errorf("node_modules should have been ignored")
			}
		}
	})

	t.Run("Scan with max depth", func(t *testing.T) {
		opts := ScanOptions{
			Root:     tempDir,
			MaxDepth: 1,
		}
		root, err := Scan(opts)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Depth 1 should see dir1 but not its contents
		var dir1 *FileNode
		for _, child := range root.Children {
			if child.Name == "dir1" {
				dir1 = child
				break
			}
		}

		if dir1 == nil {
			t.Fatalf("dir1 not found")
		}
		if len(dir1.Children) > 0 {
			t.Errorf("dir1 should have no children with max-depth 1, got %d", len(dir1.Children))
		}
	})
}
