package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
)

func EnforceCompatibility(report CompatibilityReport, capabilities CompatibilityCapabilities) error {
	for _, decision := range report.Decisions {
		switch decision.Tier {
		case TierStatic:
			continue
		case TierDynamic:
			// JavaScript module bodies execute inside the Dynamic island. Their
			// property and call semantics are intentionally not lowered natively.
			if capabilities.DynamicRuntime && (decision.DynamicCapability == "local JavaScript module call" || decision.DynamicCapability == "any type" || isJavaScriptFile(decision.FileName)) {
				continue
			}
			message := "This source site requires Dynamic execution, but the Dynamic runtime is not available in this build."
			return fmt.Errorf("%s", typescriptgo.Format(decision.FileName, decision.Start, decision.Length, "error", string(CodeDynamicRuntimeUnavailable), message, decision.Source))
		case TierUnsupported:
			return fmt.Errorf("%s", typescriptgo.Format(decision.FileName, decision.Start, decision.Length, "error", string(decision.Code), decision.Message, decision.Source))
		}
	}
	return nil
}
