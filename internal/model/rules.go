package model

import "time"

// ClusterMetric mirrors the cluster metrics used for rule/boundary evaluation.
type ClusterMetric struct {
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

// RulesReport is a serializable report of inferred boundary rules and cluster metrics.
type RulesReport struct {
    GeneratedAt    time.Time       `json:"generated_at"`
    Root           string          `json:"root"`
    ClusterMetrics []ClusterMetric `json:"cluster_metrics"`
    StrongComponents [][]string    `json:"strong_components,omitempty"`
}
