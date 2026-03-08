package structure

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/repoan/repoan/internal/model"
)

type tokenAccumulator struct {
	name    string
	kind    string
	count   int
	depths  map[int]int
	parents map[string]struct{}
	coOccur map[string]int
}

type fileRecord struct {
	relPath string
	module  string
}

type edgeAggregate struct {
	weight          int
	confidenceSum   float64
	resolvedHits    int
	unresolvedHits  int
	sources         map[string]int
	resolvers       map[string]int
	resolutionKinds map[string]int
	reasons         map[string]int
	externalHits    int
}

type rawGraphData struct {
	dependencies []RawDependency
	summary      RawDependencySummary
}

func Analyze(repoRoot string, snap *model.Snapshot, opts Options) (*Report, error) {
	if opts.TopTokens <= 0 {
		opts.TopTokens = 25
	}
	if opts.InstabilityHigh <= 0 {
		opts.InstabilityHigh = 0.85
	}

	files := flattenFiles(repoRoot, snap.Files, opts.IncludeTests)
	tokens, totalParents := buildTokenStats(files)
	rawOpts := normalizeRawDependencyOptions(opts)
	if err := validateRawDependencyOptions(rawOpts); err != nil {
		return nil, err
	}
	moduleNodes, edges, unresolvedByModule, resolvedImports, unresolvedImports, rawData, err := buildModuleGraph(repoRoot, files, rawOpts)
	if err != nil {
		return nil, err
	}

	graphForSCC := map[string]map[string]int{}
	for _, n := range moduleNodes {
		graphForSCC[n] = map[string]int{}
	}
	for from, targets := range edges {
		for to, agg := range targets {
			if agg.weight > 0 {
				graphForSCC[from][to] = agg.weight
			}
		}
	}

	sccs := stronglyConnectedComponents(moduleNodes, graphForSCC)
	inCycle := map[string]bool{}
	cycles := 0
	for _, c := range sccs {
		if len(c) > 1 {
			cycles++
			for _, m := range c {
				inCycle[m] = true
			}
		}
	}

	moduleMetrics := make([]ModuleMetrics, 0, len(moduleNodes))
	adjIn := map[string]map[string]struct{}{}
	adjOut := map[string]map[string]struct{}{}
	incomingWeight := map[string]int{}
	outgoingWeight := map[string]int{}
	for _, n := range moduleNodes {
		adjIn[n] = map[string]struct{}{}
		adjOut[n] = map[string]struct{}{}
		incomingWeight[n] = 0
		outgoingWeight[n] = 0
	}

	moduleEdges := []ModuleEdge{}
	for from, targets := range edges {
		for to, agg := range targets {
			if agg.weight <= 0 {
				continue
			}
			if from == to {
				continue
			}
			adjOut[from][to] = struct{}{}
			adjIn[to][from] = struct{}{}
			outgoingWeight[from] += agg.weight
			incomingWeight[to] += agg.weight

			dominantSource := "mixed"
			maxSrc := 0
			for src, n := range agg.sources {
				if n > maxSrc {
					maxSrc = n
					dominantSource = src
				}
			}

			avgConfidence := 0.0
			if agg.weight > 0 {
				avgConfidence = agg.confidenceSum / float64(agg.weight)
			}
			dominantResolver := dominantKey(agg.resolvers)
			dominantResolutionKind := dominantKey(agg.resolutionKinds)
			dominantReason := dominantKey(agg.reasons)

			moduleEdges = append(moduleEdges, ModuleEdge{
				From:           from,
				To:             to,
				Kind:           "import",
				Weight:         agg.weight,
				Source:         dominantSource,
				Resolver:       dominantResolver,
				ResolutionKind: dominantResolutionKind,
				Reason:         dominantReason,
				External:       agg.externalHits == agg.weight,
				Resolved:       agg.unresolvedHits == 0,
				Confidence:     round(avgConfidence),
				UnresolvedHits: agg.unresolvedHits,
			})
		}
	}

	sort.Slice(moduleEdges, func(i, j int) bool {
		if moduleEdges[i].From != moduleEdges[j].From {
			return moduleEdges[i].From < moduleEdges[j].From
		}
		if moduleEdges[i].To != moduleEdges[j].To {
			return moduleEdges[i].To < moduleEdges[j].To
		}
		return moduleEdges[i].Weight > moduleEdges[j].Weight
	})

	for _, module := range moduleNodes {
		fanIn := len(adjIn[module])
		fanOut := len(adjOut[module])
		den := fanIn + fanOut
		instability := 0.0
		if den > 0 {
			instability = float64(fanOut) / float64(den)
		}
		moduleMetrics = append(moduleMetrics, ModuleMetrics{
			Module:            module,
			FanIn:             fanIn,
			FanOut:            fanOut,
			Instability:       round(instability),
			InCycle:           inCycle[module],
			OutgoingWeight:    outgoingWeight[module],
			IncomingWeight:    incomingWeight[module],
			UnresolvedImports: unresolvedByModule[module],
		})
	}

	sort.Slice(moduleMetrics, func(i, j int) bool {
		if moduleMetrics[i].Instability != moduleMetrics[j].Instability {
			return moduleMetrics[i].Instability > moduleMetrics[j].Instability
		}
		return moduleMetrics[i].Module < moduleMetrics[j].Module
	})

	layerCandidates := inferLayerCandidates(moduleMetrics, opts.InstabilityHigh)
	clusterMetrics := computeClusterMetrics(moduleNodes, moduleEdges, inCycle)

	summary := Summary{
		Files:             len(files),
		Modules:           len(moduleNodes),
		DependencyEdges:   len(moduleEdges),
		ResolvedImports:   resolvedImports,
		UnresolvedImports: unresolvedImports,
		StrongComponents:  len(sccs),
		Cycles:            cycles,
	}

	tokenScores := rankTokens(tokens, totalParents, opts.TopTokens)

	for _, c := range sccs {
		sort.Strings(c)
	}
	sort.Slice(sccs, func(i, j int) bool {
		if len(sccs[i]) != len(sccs[j]) {
			return len(sccs[i]) > len(sccs[j])
		}
		if len(sccs[i]) == 0 || len(sccs[j]) == 0 {
			return len(sccs[i]) > len(sccs[j])
		}
		return sccs[i][0] < sccs[j][0]
	})

	report := &Report{
		GeneratedAt:      time.Now().UTC(),
		Root:             repoRoot,
		Summary:          summary,
		Tokens:           tokenScores,
		ModuleEdges:      moduleEdges,
		ModuleMetrics:    moduleMetrics,
		ClusterMetrics:   clusterMetrics,
		StrongComponents: sccs,
		LayerCandidates:  layerCandidates,
	}
	if rawOpts.Enabled {
		report.RawDependencies = rawData.dependencies
		report.RawDepSummary = &rawData.summary
	}

	return report, nil
}

func computeClusterMetrics(moduleNodes []string, moduleEdges []ModuleEdge, inCycle map[string]bool) []ClusterMetrics {
	clusterModules := map[string]map[string]struct{}{}
	for _, module := range moduleNodes {
		cluster := clusterFromModule(module)
		if _, ok := clusterModules[cluster]; !ok {
			clusterModules[cluster] = map[string]struct{}{}
		}
		clusterModules[cluster][module] = struct{}{}
	}

	internalWeight := map[string]int{}
	externalInWeight := map[string]int{}
	externalOutWeight := map[string]int{}
	internalLinks := map[string]map[string]struct{}{}
	cycleModulesByCluster := map[string]int{}

	for cluster := range clusterModules {
		internalLinks[cluster] = map[string]struct{}{}
	}

	for module := range inCycle {
		cluster := clusterFromModule(module)
		cycleModulesByCluster[cluster]++
	}

	for _, edge := range moduleEdges {
		fromCluster := clusterFromModule(edge.From)
		toCluster := clusterFromModule(edge.To)
		if fromCluster == toCluster {
			internalWeight[fromCluster] += edge.Weight
			internalLinks[fromCluster][edge.From+"->"+edge.To] = struct{}{}
			continue
		}
		externalOutWeight[fromCluster] += edge.Weight
		externalInWeight[toCluster] += edge.Weight
	}

	out := make([]ClusterMetrics, 0, len(clusterModules))
	for cluster, modules := range clusterModules {
		moduleCount := len(modules)
		possible := 0
		if moduleCount > 1 {
			possible = moduleCount * (moduleCount - 1)
		}

		internal := internalWeight[cluster]
		externalIn := externalInWeight[cluster]
		externalOut := externalOutWeight[cluster]
		externalTotal := externalIn + externalOut
		total := internal + externalTotal

		cohesion := 0.0
		if possible > 0 {
			cohesion = float64(len(internalLinks[cluster])) / float64(possible)
		}

		coupling := 0.0
		if total > 0 {
			coupling = float64(externalTotal) / float64(total)
		}

		boundary := float64(internal) / math.Max(1, float64(externalTotal))

		directionality := 0.0
		if externalTotal > 0 {
			directionality = float64(externalOut-externalIn) / float64(externalTotal)
		}

		instability := 0.0
		if externalTotal > 0 {
			instability = float64(externalOut) / float64(externalTotal)
		}

		out = append(out, ClusterMetrics{
			Cluster:             cluster,
			Modules:             moduleCount,
			InternalEdges:       internal,
			ExternalInEdges:     externalIn,
			ExternalOutEdges:    externalOut,
			CohesionScore:       round(cohesion),
			CouplingScore:       round(coupling),
			BoundaryStrength:    round(boundary),
			DirectionalityScore: round(directionality),
			Instability:         round(instability),
			InCycleModules:      cycleModulesByCluster[cluster],
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].BoundaryStrength != out[j].BoundaryStrength {
			return out[i].BoundaryStrength > out[j].BoundaryStrength
		}
		return out[i].Cluster < out[j].Cluster
	})

	return out
}

func clusterFromModule(module string) string {
	parts := strings.Split(model.NormalizePath(module), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "root"
	}
	return parts[0]
}

func dominantKey(m map[string]int) string {
	maxKey := ""
	maxVal := 0
	for k, v := range m {
		if v > maxVal || (v == maxVal && (maxKey == "" || k < maxKey)) {
			maxVal = v
			maxKey = k
		}
	}
	return maxKey
}

func flattenFiles(repoRoot string, roots []*model.FileItem, includeTests bool) []fileRecord {
	out := []fileRecord{}
	var walk func(n *model.FileItem)
	walk = func(n *model.FileItem) {
		if n == nil {
			return
		}
		if !n.IsDir {
			rel := model.NormalizePath(n.Path)
			if filepath.IsAbs(rel) {
				if rp, err := filepath.Rel(repoRoot, rel); err == nil {
					rel = model.NormalizePath(rp)
				}
			}
			if rel == "." || rel == "" {
				return
			}
			if !includeTests && isTestFile(rel) {
				return
			}
			out = append(out, fileRecord{relPath: rel, module: moduleFromPath(rel)})
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}

	for _, r := range roots {
		walk(r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].relPath < out[j].relPath })
	return out
}

func buildTokenStats(files []fileRecord) (map[string]*tokenAccumulator, int) {
	tokens := map[string]*tokenAccumulator{}
	parents := map[string]struct{}{}

	ensure := func(name, kind string) *tokenAccumulator {
		k := kind + ":" + name
		if t, ok := tokens[k]; ok {
			return t
		}
		t := &tokenAccumulator{
			name:    name,
			kind:    kind,
			depths:  map[int]int{},
			parents: map[string]struct{}{},
			coOccur: map[string]int{},
		}
		tokens[k] = t
		return t
	}

	siblingsByParent := map[string][]string{}
	for _, f := range files {
		parts := strings.Split(model.NormalizePath(f.relPath), "/")
		for i, p := range parts {
			kind := "dir"
			if i == len(parts)-1 {
				kind = "file"
			}
			parent := "__root__"
			if i > 0 {
				parent = strings.Join(parts[:i], "/")
			}
			parents[parent] = struct{}{}
			t := ensure(p, kind)
			t.count++
			t.depths[i+1]++
			t.parents[parent] = struct{}{}
		}

		for i := 0; i < len(parts)-1; i++ {
			parent := strings.Join(parts[:i], "/")
			if parent == "" {
				parent = "__root__"
			}
			siblingsByParent[parent] = append(siblingsByParent[parent], parts[i])
		}
	}

	for _, sibs := range siblingsByParent {
		seen := map[string]struct{}{}
		for _, s := range sibs {
			seen[s] = struct{}{}
		}
		uniq := make([]string, 0, len(seen))
		for s := range seen {
			uniq = append(uniq, s)
		}
		for i := 0; i < len(uniq); i++ {
			for j := i + 1; j < len(uniq); j++ {
				a := ensure(uniq[i], "dir")
				b := ensure(uniq[j], "dir")
				a.coOccur[uniq[j]]++
				b.coOccur[uniq[i]]++
			}
		}
	}

	return tokens, len(parents)
}

func buildModuleGraph(repoRoot string, files []fileRecord, rawOpts RawDependencyOptions) ([]string, map[string]map[string]*edgeAggregate, map[string]int, int, int, rawGraphData, error) {
	moduleSet := map[string]struct{}{}
	fileSet := map[string]struct{}{}
	for _, f := range files {
		moduleSet[f.module] = struct{}{}
		fileSet[f.relPath] = struct{}{}
	}

	nodes := make([]string, 0, len(moduleSet))
	for m := range moduleSet {
		nodes = append(nodes, m)
	}
	sort.Strings(nodes)

	edges := map[string]map[string]*edgeAggregate{}
	for _, m := range nodes {
		edges[m] = map[string]*edgeAggregate{}
	}

	goModulePath := parseGoModulePath(filepath.Join(repoRoot, "go.mod"))
	tsResolver := loadTSResolver(repoRoot)
	fileModule := map[string]string{}
	for _, f := range files {
		fileModule[f.relPath] = f.module
	}
	ctx := resolveContext{
		goModulePath: goModulePath,
		fileModule:   fileModule,
		fileSet:      fileSet,
		moduleSet:    moduleSet,
		ts:           tsResolver,
	}

	unresolvedByModule := map[string]int{}
	resolvedImports := 0
	unresolvedImports := 0
	rawDependencies := []RawDependency{}
	rawSummary := RawDependencySummary{
		ByResolver:       map[string]int{},
		ByResolutionKind: map[string]int{},
	}

	for _, f := range files {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(f.relPath))
		imports, err := extractImports(absPath)
		if err != nil {
			continue
		}
		for _, imp := range imports {
			res := resolveImportModule(imp, f.relPath, ctx)
			if rawOpts.Enabled {
				rawSummary.Total++
				if res.Resolved {
					rawSummary.Resolved++
				} else {
					rawSummary.Unresolved++
				}
				if res.External {
					rawSummary.External++
				} else {
					rawSummary.Internal++
				}
				if res.Resolver != "" {
					rawSummary.ByResolver[res.Resolver]++
				}
				if res.ResolutionKind != "" {
					rawSummary.ByResolutionKind[res.ResolutionKind]++
				}

				dep := RawDependency{
					FromFile:       f.relPath,
					FromModule:     f.module,
					ToFile:         res.TargetFile,
					ToModule:       res.TargetModule,
					SourceImport:   imp.Path,
					Resolver:       res.Resolver,
					ResolutionKind: res.ResolutionKind,
					Reason:         res.Reason,
					Resolved:       res.Resolved,
					External:       res.External,
					Confidence:     round(res.Confidence),
				}
				if includeRawDependency(dep, rawOpts) {
					rawDependencies = append(rawDependencies, dep)
				}
			}
			targetModule := res.TargetModule
			if targetModule == "" || targetModule == f.module || targetModule == "external" {
				if !res.Resolved {
					unresolvedImports++
					unresolvedByModule[f.module]++
				}
				continue
			}
			agg, ok := edges[f.module][targetModule]
			if !ok {
				agg = &edgeAggregate{sources: map[string]int{}, resolvers: map[string]int{}, resolutionKinds: map[string]int{}, reasons: map[string]int{}}
				edges[f.module][targetModule] = agg
			}
			agg.weight++
			agg.confidenceSum += res.Confidence
			agg.sources[imp.Source]++
			if res.Resolver != "" {
				agg.resolvers[res.Resolver]++
			}
			if res.ResolutionKind != "" {
				agg.resolutionKinds[res.ResolutionKind]++
			}
			if res.Reason != "" {
				agg.reasons[res.Reason]++
			}
			if res.External {
				agg.externalHits++
			}
			if res.Resolved {
				agg.resolvedHits++
				resolvedImports++
			} else {
				agg.unresolvedHits++
				unresolvedImports++
				unresolvedByModule[f.module]++
			}
		}
	}

	if rawOpts.Enabled {
		sort.Slice(rawDependencies, func(i, j int) bool {
			if rawDependencies[i].Resolved != rawDependencies[j].Resolved {
				return !rawDependencies[i].Resolved
			}
			if rawDependencies[i].Confidence != rawDependencies[j].Confidence {
				return rawDependencies[i].Confidence < rawDependencies[j].Confidence
			}
			if rawDependencies[i].Resolver != rawDependencies[j].Resolver {
				return rawDependencies[i].Resolver < rawDependencies[j].Resolver
			}
			if rawDependencies[i].FromFile != rawDependencies[j].FromFile {
				return rawDependencies[i].FromFile < rawDependencies[j].FromFile
			}
			if rawDependencies[i].SourceImport != rawDependencies[j].SourceImport {
				return rawDependencies[i].SourceImport < rawDependencies[j].SourceImport
			}
			return rawDependencies[i].ToModule < rawDependencies[j].ToModule
		})
		if rawOpts.Limit > 0 && len(rawDependencies) > rawOpts.Limit {
			rawDependencies = rawDependencies[:rawOpts.Limit]
		}
		rawSummary.Emitted = len(rawDependencies)
	}

	return nodes, edges, unresolvedByModule, resolvedImports, unresolvedImports, rawGraphData{
		dependencies: rawDependencies,
		summary:      rawSummary,
	}, nil
}

func normalizeRawDependencyOptions(opts Options) RawDependencyOptions {
	raw := opts.RawDependencies
	if !raw.Enabled && (raw.Limit > 0 || raw.OnlyUnresolved || raw.OnlyExternal || raw.MinConfidence > 0) {
		raw.Enabled = true
	}
	return raw
}

func validateRawDependencyOptions(opts RawDependencyOptions) error {
	if opts.MinConfidence < 0 || opts.MinConfidence > 1 {
		return fmt.Errorf("raw dependency min confidence must be between 0 and 1, got %.3f", opts.MinConfidence)
	}
	return nil
}

func includeRawDependency(dep RawDependency, opts RawDependencyOptions) bool {
	if opts.OnlyUnresolved && dep.Resolved {
		return false
	}
	if opts.OnlyExternal && !dep.External {
		return false
	}
	if dep.Confidence < opts.MinConfidence {
		return false
	}
	return true
}

func moduleFromPath(relPath string) string {
	relPath = model.NormalizePath(relPath)
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "root"
	}
	if len(parts) == 1 {
		return "root"
	}
	if len(parts) == 2 {
		if strings.Contains(parts[1], ".") {
			return parts[0]
		}
		return parts[0] + "/" + parts[1]
	}
	if len(parts) >= 3 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

func rankTokens(tokens map[string]*tokenAccumulator, totalParents, limit int) []TokenScore {
	if totalParents <= 0 {
		totalParents = 1
	}
	totalTokens := len(tokens)
	if totalTokens <= 1 {
		totalTokens = 2
	}

	noise := map[string]struct{}{
		"dist": {}, "build": {}, "coverage": {}, "tmp": {}, "temp": {}, "node_modules": {}, "vendor": {},
	}

	out := make([]TokenScore, 0, len(tokens))
	for _, t := range tokens {
		freq := math.Log1p(float64(t.count)) / math.Log1p(float64(totalTokens))
		spread := float64(len(t.parents)) / float64(totalParents)
		depthKinds := len(t.depths)
		depthConsistency := 1.0
		if depthKinds > 1 {
			depthConsistency = 1.0 / float64(depthKinds)
		}
		coMax := 0
		coTop := topCoOccur(t.coOccur, 3)
		for _, v := range t.coOccur {
			if v > coMax {
				coMax = v
			}
		}
		coScore := math.Min(1.0, float64(coMax)/math.Max(1.0, float64(t.count)))
		penalty := 0.0
		if _, ok := noise[strings.ToLower(t.name)]; ok {
			penalty = 0.2
		}
		weight := 0.35*freq + 0.25*spread + 0.2*depthConsistency + 0.2*coScore - penalty
		if weight < 0 {
			weight = 0
		}
		if weight > 1 {
			weight = 1
		}
		out = append(out, TokenScore{
			Name:       t.name,
			Kind:       t.kind,
			Count:      t.count,
			Spread:     round(spread),
			DepthKinds: depthKinds,
			Weight:     round(weight),
			TopCoOccur: coTop,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Weight != out[j].Weight {
			return out[i].Weight > out[j].Weight
		}
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})

	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func topCoOccur(m map[string]int, n int) []string {
	type kv struct {
		key string
		val int
	}
	pairs := make([]kv, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kv{key: k, val: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].val != pairs[j].val {
			return pairs[i].val > pairs[j].val
		}
		return pairs[i].key < pairs[j].key
	})
	if n > len(pairs) {
		n = len(pairs)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, pairs[i].key)
	}
	return out
}

func inferLayerCandidates(metrics []ModuleMetrics, instabilityHigh float64) []LayerCandidate {
	out := []LayerCandidate{}
	for _, m := range metrics {
		role := "probable_shared"
		conf := 0.55
		reasons := []string{"balanced fan-in/fan-out profile"}

		switch {
		case m.FanIn == 0 && m.FanOut >= 2:
			role = "probable_entry"
			conf = 0.80
			reasons = []string{"high fan-out with no incoming dependencies"}
		case m.FanIn >= 2 && m.Instability <= 0.30:
			role = "probable_core"
			conf = 0.82
			reasons = []string{"high fan-in and low instability"}
		case m.FanIn > 0 && m.FanOut > 0:
			role = "probable_orchestration"
			conf = 0.70
			reasons = []string{"acts as dependency bridge with incoming and outgoing edges"}
		case m.FanOut > m.FanIn && m.Instability >= instabilityHigh:
			role = "probable_infra"
			conf = 0.66
			reasons = []string{fmt.Sprintf("instability %.2f above threshold %.2f", m.Instability, instabilityHigh)}
		}

		if m.OutgoingWeight > m.IncomingWeight {
			reasons = append(reasons, "outgoing dependency weight is higher than incoming")
		}
		if m.UnresolvedImports > 0 {
			reasons = append(reasons, fmt.Sprintf("contains %d unresolved imports", m.UnresolvedImports))
		}

		if m.InCycle {
			conf = math.Max(0.50, conf-0.12)
			reasons = append(reasons, "participates in dependency cycle")
		}

		out = append(out, LayerCandidate{
			Module:     m.Module,
			Role:       role,
			Confidence: round(conf),
			Reasons:    reasons,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		return out[i].Module < out[j].Module
	})
	return out
}

func round(v float64) float64 {
	return math.Round(v*1000) / 1000
}
