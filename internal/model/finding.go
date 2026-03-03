package model

type Severity string

const (
	SevInfo     Severity = "info"
	SevLow      Severity = "low"
	SevMedium   Severity = "medium"
	SevHigh     Severity = "high"
	SevCritical Severity = "critical"
)

type Category string

const (
	CatSecurity  Category = "security"
	CatHygiene   Category = "hygiene"
	CatStructure Category = "structure"
	CatPolicy    Category = "policy"
)

// Finding represents a single issue found by an analyzer.
type Finding struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Message     string   `json:"message"`
	Severity    Severity `json:"severity"`
	Category    Category `json:"category"`

	// Location
	Path   string `json:"path"` // repo-relative, forward slashes
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`

	// Evidence (keep it small; never dump secrets)
	Evidence    string `json:"evidence,omitempty"` // truncated/sanitized
	Remediation string `json:"remediation,omitempty"`

	// Baseline
	Fingerprint string `json:"fingerprint"`

	// Optional metadata
	Tags  []string          `json:"tags,omitempty"`
	Props map[string]string `json:"props,omitempty"`
}
