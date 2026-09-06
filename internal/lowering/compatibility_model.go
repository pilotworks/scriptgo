package lowering

import typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"

const CompatibilityReportFormat = 1

type CompatibilityTier string

const (
	TierStatic      CompatibilityTier = "static"
	TierDynamic     CompatibilityTier = "dynamic"
	TierUnsupported CompatibilityTier = "unsupported"
)

type CompatibilityMode string

const (
	ModeStatic         CompatibilityMode = "static"
	ModeDynamicEnabled CompatibilityMode = "dynamic-enabled"
)

type CompatibilityPolicy struct {
	Mode CompatibilityMode
}

type CompatibilityDecision struct {
	Tier     CompatibilityTier       `json:"tier"`
	FileName string                  `json:"path"`
	Span     typescriptgo.SourceSpan `json:"-"`
	Start    int                     `json:"start"`
	Length   int                     `json:"length"`
	Kind     string                  `json:"kind"`
	Code     SubsetCode              `json:"code,omitempty"`
	Message  string                  `json:"message,omitempty"`
	Hint     string                  `json:"hint,omitempty"`
	Source   string                  `json:"-"`
}

type CompatibilitySummary struct {
	Static      int `json:"static"`
	Dynamic     int `json:"dynamic"`
	Unsupported int `json:"unsupported"`
}

type CompatibilityReport struct {
	Format    int                     `json:"format"`
	Mode      CompatibilityMode       `json:"mode"`
	Summary   CompatibilitySummary    `json:"summary"`
	Decisions []CompatibilityDecision `json:"sites"`
}

type CompatibilityCapabilities struct {
	DynamicRuntime bool
}
