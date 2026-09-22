package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// ValidateSubset rejects checked syntax that the current native IR cannot
// represent. Keeping this gate in lowering prevents backend-specific policy.
func ValidateSubset(program frontend.Program) error {
	return ValidateSubsetWithOptions(program, Options{WarnRuntimeCasts: WarnRuntimeCasts})
}

// ValidateSubsetWithOptions validates compatibility without producing IR.
func ValidateSubsetWithOptions(program frontend.Program, options Options) error {
	lowerMu.Lock()
	defer lowerMu.Unlock()
	previousWarn := WarnRuntimeCasts
	WarnRuntimeCasts = options.WarnRuntimeCasts || previousWarn
	defer func() { WarnRuntimeCasts = previousWarn }()
	ClearDiagnostics()
	if options.Dynamic {
		program = normalizeDynamicAnyProgram(program)
	}
	return validateSubsetLocked(program)
}

func validateSubsetLocked(program frontend.Program) error {
	report, err := analyzeCompatibilityLocked(program, CompatibilityPolicy{Mode: ModeStatic})
	if err != nil {
		return err
	}
	return EnforceCompatibility(report, CompatibilityCapabilities{})
}
