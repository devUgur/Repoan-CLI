package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/repoan/repoan/internal/model"
	gitignore "github.com/sabhiram/go-gitignore"
)

type ScanOptions struct {
	Root             string
	IgnorePatterns   []string
	MaxDepth         int
	RespectGitignore bool
}

func Scan(opts ScanOptions) (*model.FileItem, error) {
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

	// Also use gitignore-style matching for manual ignore patterns
	var customIgnoreMatcher *gitignore.GitIgnore
	if len(opts.IgnorePatterns) > 0 {
		customIgnoreMatcher = gitignore.CompileIgnoreLines(opts.IgnorePatterns...)
	}

	rootNode := &model.FileItem{
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

		// Check custom ignore patterns
		if customIgnoreMatcher != nil && customIgnoreMatcher.MatchesPath(matchPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Add to tree
		addNode(rootNode, relPath, d.IsDir(), info.Size())

		return nil
	})

	return rootNode, err
}

func addNode(root *model.FileItem, relPath string, isDir bool, size int64) {
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
			ext := ""
			if !isDir || i < len(parts)-1 {
				ext = filepath.Ext(part)
			}

			newNode := &model.FileItem{
				Name:      part,
				Path:      filepath.Join(current.Path, part),
				IsDir:     isDir && (i == len(parts)-1),
				Size:      size,
				Extension: ext,
			}
			current.Children = append(current.Children, newNode)
			current = newNode
		}
	}
}
