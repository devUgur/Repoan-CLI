package structure

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	jsImportRe     = regexp.MustCompile(`(?m)(?:import\s+(?:[^"']+\s+from\s+)?|require\()\s*["']([^"']+)["']`)
	pyImportRe     = regexp.MustCompile(`(?m)^\s*import\s+([a-zA-Z0-9_\.]+)\s*$`)
	pyFromImportRe = regexp.MustCompile(`(?m)^\s*from\s+([a-zA-Z0-9_\.]+)\s+import\s+`)
)

type ImportResolver interface {
	Name() string
	Supports(ext string) bool
	Extract(absPath string) ([]parsedImport, error)
	Resolve(imp parsedImport, fromFile string, ctx resolveContext) (ResolveResult, bool)
}

type goImportResolver struct{}
type jsImportResolver struct{}
type pyImportResolver struct{}

type parsedImport struct {
	Path       string
	Source     string
	Confidence float64
}

type ResolveResult struct {
	TargetModule   string
	TargetFile     string
	Resolved       bool
	Confidence     float64
	Resolver       string
	ResolutionKind string
	Reason         string
	External       bool
}

type resolveContext struct {
	goModulePath string
	fileModule   map[string]string
	fileSet      map[string]struct{}
	moduleSet    map[string]struct{}
	ts           *tsResolver
}

type tsResolver struct {
	baseURL string
	paths   map[string][]string
}

type tsConfigFile struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

func extractImports(absPath string) ([]parsedImport, error) {
	ext := strings.ToLower(filepath.Ext(absPath))
	for _, resolver := range defaultImportResolvers() {
		if !resolver.Supports(ext) {
			continue
		}
		return resolver.Extract(absPath)
	}
	return nil, nil
}

func defaultImportResolvers() []ImportResolver {
	return []ImportResolver{
		goImportResolver{},
		jsImportResolver{},
		pyImportResolver{},
	}
}

func (goImportResolver) Name() string { return "go-ast" }

func (goImportResolver) Supports(ext string) bool { return ext == ".go" }

func (goImportResolver) Extract(absPath string) ([]parsedImport, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, absPath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	imports := make([]parsedImport, 0, len(f.Imports))
	for _, imp := range f.Imports {
		if imp == nil || imp.Path == nil {
			continue
		}
		imports = append(imports, parsedImport{Path: strings.Trim(imp.Path.Value, `"`), Source: "go-ast", Confidence: 0.98})
	}
	return imports, nil
}

func (goImportResolver) Resolve(imp parsedImport, fromFile string, ctx resolveContext) (ResolveResult, bool) {
	if imp.Source != "go-ast" {
		return ResolveResult{}, false
	}
	importPath := strings.TrimSpace(imp.Path)
	if importPath == "" {
		return ResolveResult{}, true
	}

	if ctx.goModulePath != "" {
		if importPath == ctx.goModulePath {
			return ResolveResult{TargetModule: "root", Resolved: true, Confidence: 1.0, Resolver: "go", ResolutionKind: "module_root", Reason: "go module root"}, true
		}
		prefix := ctx.goModulePath + "/"
		if strings.HasPrefix(importPath, prefix) {
			rel := strings.TrimPrefix(importPath, prefix)
			return ResolveResult{TargetModule: moduleFromPath(rel), Resolved: true, Confidence: 1.0, Resolver: "go", ResolutionKind: "module_internal", Reason: "go internal module import"}, true
		}
	}

	if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
		return ResolveResult{TargetModule: "external", Resolved: true, Confidence: 1.0, Resolver: "go", ResolutionKind: "stdlib", Reason: "go stdlib", External: true}, true
	}

	return ResolveResult{TargetModule: "external", Resolved: true, Confidence: boost(imp.Confidence, 0.2), Resolver: "go", ResolutionKind: "third_party", Reason: "go third-party module", External: true}, true
}

func (jsImportResolver) Name() string { return "js-regex" }

func (jsImportResolver) Supports(ext string) bool {
	switch ext {
	case ".js", ".mjs", ".cjs", ".ts", ".tsx", ".jsx":
		return true
	default:
		return false
	}
}

func (jsImportResolver) Extract(absPath string) ([]parsedImport, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	matches := jsImportRe.FindAllStringSubmatch(string(data), -1)
	imports := make([]parsedImport, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 && m[1] != "" {
			imports = append(imports, parsedImport{Path: m[1], Source: "js-regex", Confidence: 0.60})
		}
	}
	return imports, nil
}

func (jsImportResolver) Resolve(imp parsedImport, fromFile string, ctx resolveContext) (ResolveResult, bool) {
	if imp.Source != "js-regex" {
		return ResolveResult{}, false
	}
	importPath := strings.TrimSpace(imp.Path)
	if importPath == "" {
		return ResolveResult{}, true
	}

	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		baseDir := filepath.Dir(fromFile)
		resolved := filepath.Clean(filepath.Join(baseDir, importPath))
		resolved = modelNormalize(resolved, "")
		if m, target, ok := resolveByPathCandidates(resolved, ctx.fileModule, ctx.fileSet); ok {
			return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.25), Resolver: "js", ResolutionKind: "relative_file", Reason: "relative JS/TS import"}, true
		}
		return ResolveResult{TargetModule: "external", Resolved: false, Confidence: imp.Confidence, Resolver: "js", ResolutionKind: "relative_unresolved", Reason: "unresolved relative JS/TS import", External: true}, true
	}

	if ctx.ts != nil {
		if mapped := ctx.ts.resolve(importPath); mapped != "" {
			if m, target, ok := resolveByPathCandidates(mapped, ctx.fileModule, ctx.fileSet); ok {
				return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.2), Resolver: "js", ResolutionKind: "tsconfig_path", Reason: "tsconfig path alias"}, true
			}
		}
	}

	if m, target, ok := resolveByPathCandidates(modelNormalize(importPath, ""), ctx.fileModule, ctx.fileSet); ok {
		return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.1), Resolver: "js", ResolutionKind: "workspace_absolute", Reason: "workspace absolute JS/TS import"}, true
	}

	if !strings.HasPrefix(importPath, "/") {
		return ResolveResult{TargetModule: "external", Resolved: true, Confidence: boost(imp.Confidence, 0.1), Resolver: "js", ResolutionKind: "package_external", Reason: "npm package import", External: true}, true
	}

	return ResolveResult{TargetModule: "external", Resolved: false, Confidence: imp.Confidence, Resolver: "js", ResolutionKind: "unresolved", Reason: "unresolved JS/TS import", External: true}, true
}

func (pyImportResolver) Name() string { return "py-regex" }

func (pyImportResolver) Supports(ext string) bool { return ext == ".py" }

func (pyImportResolver) Extract(absPath string) ([]parsedImport, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	content := string(data)
	imports := []parsedImport{}
	for _, m := range pyImportRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			imports = append(imports, parsedImport{Path: m[1], Source: "py-regex", Confidence: 0.65})
		}
	}
	for _, m := range pyFromImportRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			imports = append(imports, parsedImport{Path: m[1], Source: "py-regex", Confidence: 0.65})
		}
	}
	return imports, nil
}

func (pyImportResolver) Resolve(imp parsedImport, fromFile string, ctx resolveContext) (ResolveResult, bool) {
	if imp.Source != "py-regex" {
		return ResolveResult{}, false
	}
	importPath := strings.TrimSpace(imp.Path)
	if importPath == "" {
		return ResolveResult{}, true
	}

	if strings.HasPrefix(importPath, ".") {
		leadingDots := 0
		for leadingDots < len(importPath) && importPath[leadingDots] == '.' {
			leadingDots++
		}
		rest := strings.TrimPrefix(importPath, strings.Repeat(".", leadingDots))
		baseDir := filepath.Dir(fromFile)
		for i := 1; i < leadingDots && baseDir != "." && baseDir != ""; i++ {
			baseDir = filepath.Dir(baseDir)
		}
		candidate := modelNormalize(filepath.Join(baseDir, strings.ReplaceAll(rest, ".", "/")), "")
		if m, target, ok := resolveByPathCandidates(candidate, ctx.fileModule, ctx.fileSet); ok {
			return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.2), Resolver: "python", ResolutionKind: "relative_module", Reason: "relative python import"}, true
		}
		return ResolveResult{TargetModule: "external", Resolved: false, Confidence: imp.Confidence, Resolver: "python", ResolutionKind: "relative_unresolved", Reason: "unresolved relative python import", External: true}, true
	}

	pyCandidate := strings.ReplaceAll(importPath, ".", "/")
	if m, target, ok := resolveByPathCandidates(pyCandidate, ctx.fileModule, ctx.fileSet); ok {
		return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.15), Resolver: "python", ResolutionKind: "absolute_module", Reason: "python module import"}, true
	}

	parts := strings.Split(pyCandidate, "/")
	if candidate := safePart(parts, 0); candidate != "" {
		if _, ok := ctx.moduleSet[candidate]; ok {
			return ResolveResult{TargetModule: candidate, Resolved: true, Confidence: imp.Confidence, Resolver: "python", ResolutionKind: "top_level_package", Reason: "python top-level package maps to module"}, true
		}
	}

	return ResolveResult{TargetModule: "external", Resolved: true, Confidence: boost(imp.Confidence, 0.1), Resolver: "python", ResolutionKind: "package_external", Reason: "external python package", External: true}, true
}

func resolveImportModule(imp parsedImport, fromFile string, ctx resolveContext) ResolveResult {
	for _, resolver := range defaultImportResolvers() {
		res, ok := resolver.Resolve(imp, fromFile, ctx)
		if !ok {
			continue
		}
		if res.Resolver == "" {
			res.Resolver = resolver.Name()
		}
		if res.TargetModule == "" {
			return ResolveResult{TargetModule: "external", Resolved: false, Confidence: imp.Confidence, Resolver: resolver.Name(), ResolutionKind: "empty_target", Reason: "resolver returned empty target", External: true}
		}
		res.Confidence = clamp01(res.Confidence)
		return res
	}

	importPath := strings.TrimSpace(imp.Path)
	if m, target, ok := resolveByPathCandidates(modelNormalize(importPath, ""), ctx.fileModule, ctx.fileSet); ok {
		return ResolveResult{TargetModule: m, TargetFile: target, Resolved: true, Confidence: boost(imp.Confidence, 0.1), Resolver: "fallback", ResolutionKind: "workspace_path", Reason: "fallback workspace path candidate"}
	}
	parts := strings.Split(strings.ReplaceAll(importPath, ".", "/"), "/")
	if candidate := safePart(parts, 0); candidate != "" {
		if _, ok := ctx.moduleSet[candidate]; ok {
			return ResolveResult{TargetModule: candidate, Resolved: true, Confidence: imp.Confidence, Resolver: "fallback", ResolutionKind: "module_prefix", Reason: "fallback top-level module match"}
		}
	}
	return ResolveResult{TargetModule: "external", Resolved: false, Confidence: imp.Confidence, Resolver: "fallback", ResolutionKind: "unresolved", Reason: "no resolver matched import", External: true}
}

func resolveByPathCandidates(base string, fileModule map[string]string, fileSet map[string]struct{}) (string, string, bool) {
	candidates := []string{
		base,
		base + ".go",
		base + ".ts",
		base + ".tsx",
		base + ".js",
		base + ".jsx",
		base + ".mjs",
		base + ".cjs",
		base + ".py",
		base + "/index.ts",
		base + "/index.tsx",
		base + "/index.js",
		base + "/index.jsx",
		base + "/index.py",
		base + "/__init__.py",
	}
	for _, c := range candidates {
		c = modelNormalize(c, "")
		if _, ok := fileSet[c]; !ok {
			continue
		}
		if m, ok := fileModule[c]; ok {
			return m, c, true
		}
		return moduleFromPath(c), c, true
	}
	return "", "", false
}

func isTestFile(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(base))
	if ext == ".go" {
		return strings.HasSuffix(base, "_test.go")
	}
	if ext == ".py" {
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	}
	return strings.Contains(strings.ToLower(base), ".test.") || strings.Contains(strings.ToLower(base), ".spec.")
}

func parseGoModulePath(goModPath string) string {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func loadTSResolver(repoRoot string) *tsResolver {
	paths := []string{
		filepath.Join(repoRoot, "tsconfig.json"),
		filepath.Join(repoRoot, "jsconfig.json"),
	}
	for _, cfgPath := range paths {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		var cfg tsConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		base := cfg.CompilerOptions.BaseURL
		if base == "" {
			base = "."
		}
		return &tsResolver{
			baseURL: modelNormalize(filepath.Join(filepath.Dir(cfgPath), base), repoRoot),
			paths:   cfg.CompilerOptions.Paths,
		}
	}
	return nil
}

func (r *tsResolver) resolve(importPath string) string {
	if r == nil {
		return ""
	}
	for pattern, targets := range r.paths {
		if strings.Contains(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if !strings.HasPrefix(importPath, prefix) {
				continue
			}
			suffix := strings.TrimPrefix(importPath, prefix)
			for _, t := range targets {
				replaced := strings.ReplaceAll(t, "*", suffix)
				return modelNormalize(filepath.Join(r.baseURL, replaced), "")
			}
			continue
		}
		if pattern == importPath {
			if len(targets) > 0 {
				return modelNormalize(filepath.Join(r.baseURL, targets[0]), "")
			}
		}
	}

	if !strings.HasPrefix(importPath, ".") {
		return modelNormalize(filepath.Join(r.baseURL, importPath), "")
	}
	return ""
}

func modelNormalize(path string, repoRoot string) string {
	path = filepath.Clean(path)
	if repoRoot != "" {
		if rel, err := filepath.Rel(repoRoot, path); err == nil {
			path = rel
		}
	}
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return path
}

func boost(v, delta float64) float64 {
	return clamp01(v + delta)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func safePart(parts []string, idx int) string {
	if idx >= 0 && idx < len(parts) {
		return parts[idx]
	}
	return ""
}
