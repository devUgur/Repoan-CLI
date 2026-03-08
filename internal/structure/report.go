package structure

import "time"

type TokenScore struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	Count      int      `json:"count"`
	Spread     float64  `json:"spread"`
	DepthKinds int      `json:"depth_kinds"`
	Weight     float64  `json:"weight"`
	TopCoOccur []string `json:"top_co_occur,omitempty"`
}

type ModuleEdge struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	Kind           string  `json:"kind"`
	Weight         int     `json:"weight"`
	Source         string  `json:"source"`
	Resolver       string  `json:"resolver,omitempty"`
	ResolutionKind string  `json:"resolution_kind,omitempty"`
	Reason         string  `json:"reason,omitempty"`
	External       bool    `json:"external"`
	Resolved       bool    `json:"resolved"`
	Confidence     float64 `json:"confidence"`
	UnresolvedHits int     `json:"unresolved_hits,omitempty"`
}

type RawDependency struct {
	FromFile       string  `json:"from_file"`
	FromModule     string  `json:"from_module"`
	ToFile         string  `json:"to_file,omitempty"`
	ToModule       string  `json:"to_module,omitempty"`
	SourceImport   string  `json:"source_import"`
	Resolver       string  `json:"resolver,omitempty"`
	ResolutionKind string  `json:"resolution_kind,omitempty"`
	Reason         string  `json:"reason,omitempty"`
	Resolved       bool    `json:"resolved"`
	External       bool    `json:"external"`
	Confidence     float64 `json:"confidence"`
}

type RawDependencySummary struct {
	Total            int            `json:"total"`
	Emitted          int            `json:"emitted"`
	Resolved         int            `json:"resolved"`
	Unresolved       int            `json:"unresolved"`
	External         int            `json:"external"`
	Internal         int            `json:"internal"`
	ByResolver       map[string]int `json:"by_resolver,omitempty"`
	ByResolutionKind map[string]int `json:"by_resolution_kind,omitempty"`
}

type RawDependencyOptions struct {
	Enabled        bool
	Limit          int
	OnlyUnresolved bool
	OnlyExternal   bool
	MinConfidence  float64
}

type ModuleMetrics struct {
	Module            string  `json:"module"`
	FanIn             int     `json:"fan_in"`
	FanOut            int     `json:"fan_out"`
	Instability       float64 `json:"instability"`
	InCycle           bool    `json:"in_cycle"`
	OutgoingWeight    int     `json:"outgoing_weight"`
	IncomingWeight    int     `json:"incoming_weight"`
	UnresolvedImports int     `json:"unresolved_imports"`
}

type LayerCandidate struct {
	Module     string   `json:"module"`
	Role       string   `json:"role"`
	Confidence float64  `json:"confidence"`
	Reasons    []string `json:"reasons"`
}

type ClusterMetrics struct {
	Cluster             string  `json:"cluster"`
	Modules             int     `json:"modules"`
	InternalEdges       int     `json:"internal_edges"`
	ExternalInEdges     int     `json:"external_in_edges"`
	ExternalOutEdges    int     `json:"external_out_edges"`
	CohesionScore       float64 `json:"cohesion_score"`
	CouplingScore       float64 `json:"coupling_score"`
	BoundaryStrength    float64 `json:"boundary_strength"`
	DirectionalityScore float64 `json:"directionality_score"`
	Instability         float64 `json:"instability"`
	InCycleModules      int     `json:"in_cycle_modules"`
}

type Summary struct {
	Files             int `json:"files"`
	Modules           int `json:"modules"`
	DependencyEdges   int `json:"dependency_edges"`
	ResolvedImports   int `json:"resolved_imports"`
	UnresolvedImports int `json:"unresolved_imports"`
	StrongComponents  int `json:"strong_components"`
	Cycles            int `json:"cycles"`
}

type Report struct {
	GeneratedAt      time.Time             `json:"generated_at"`
	Root             string                `json:"root"`
	Summary          Summary               `json:"summary"`
	Tokens           []TokenScore          `json:"tokens"`
	RawDependencies  []RawDependency       `json:"raw_dependencies,omitempty"`
	RawDepSummary    *RawDependencySummary `json:"raw_dependency_summary,omitempty"`
	ModuleEdges      []ModuleEdge          `json:"module_edges"`
	ModuleMetrics    []ModuleMetrics       `json:"module_metrics"`
	ClusterMetrics   []ClusterMetrics      `json:"cluster_metrics"`
	StrongComponents [][]string            `json:"strong_components"`
	LayerCandidates  []LayerCandidate      `json:"layer_candidates"`
}

type Options struct {
	TopTokens           int
	IncludeTests        bool
	InstabilityHigh     float64
	ClusterCouplingHigh float64
	ClusterBoundaryLow  float64
	RawDependencies     RawDependencyOptions
}
