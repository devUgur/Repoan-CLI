package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
}

type ScanOptions struct {
	Root             string
	IgnorePatterns   []string
	MaxDepth         int
	RespectGitignore bool
}

func Scan(opts ScanOptions) (*FileNode, error) {
	rootAbs, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}

	var gitIgnoreMatcher *gitignore.GitIgnore
	if opts.RespectGitignore {
		giPath := filepath.Join(rootAbs, ".gitignore")
		if _, err := os.Stat(giPath); err == nil {
			gitIgnoreMatcher, _ = gitignore.CompileIgnoreFile(giPath)
		}
	}

	rootNode := &FileNode{
		Name:  filepath.Base(rootAbs),
		Path:  rootAbs,
		IsDir: true,
	}

	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == rootAbs {
			return nil
		}

		relPath, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}

		// Normalize to forward slashes for gitignore matching
		matchPath := filepath.ToSlash(relPath)
		if d.IsDir() && !strings.HasSuffix(matchPath, "/") {
			matchPath += "/"
		}

		// Check depth
		if opts.MaxDepth > 0 && strings.Count(relPath, string(os.PathSeparator)) >= opts.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check .gitignore
		if gitIgnoreMatcher != nil && gitIgnoreMatcher.MatchesPath(matchPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check ignore patterns
		for _, pattern := range opts.IgnorePatterns {
			if strings.Contains(relPath, pattern) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Add to tree
		addNode(rootNode, relPath, d.IsDir())

		return nil
	})

	return rootNode, err
}

func addNode(root *FileNode, relPath string, isDir bool) {
	parts := strings.Split(relPath, string(filepath.Separator))
	current := root

	for i, part := range parts {
		found := false
		for _, child := range current.Children {
			if child.Name == part {
				current = child
				found = true
				break
			}
		}

		if !found {
			newNode := &FileNode{
				Name:  part,
				Path:  filepath.Join(current.Path, part),
				IsDir: isDir && (i == len(parts)-1),
			}
			current.Children = append(current.Children, newNode)
			current = newNode
		}
	}
}
