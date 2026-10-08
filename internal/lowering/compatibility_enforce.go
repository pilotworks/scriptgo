package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func EnforceCompatibility(report CompatibilityReport, capabilities CompatibilityCapabilities) error {
	for _, decision := range report.Decisions {
		switch decision.Tier {
		case TierStatic:
			continue
		case TierDynamic:
			if capabilities.DynamicRuntime && (decision.DynamicCapability == "local JavaScript module call" || decision.DynamicCapability == "any type" || decision.DynamicCapability == "Proxy constructor" || decision.DynamicCapability == "delete operator" || isJavaScriptFile(decision.FileName)) {
				continue
			}
			message := "This source site requires Dynamic execution, but the Dynamic runtime is not available in this build."
			return fmt.Errorf("%s", frontend.FormatSpan(decision.FileName, decision.Start, decision.Length, "error", string(CodeDynamicRuntimeUnavailable), message, decision.Source))
		case TierUnsupported:
			return fmt.Errorf("%s", frontend.FormatSpan(decision.FileName, decision.Start, decision.Length, "error", string(decision.Code), decision.Message, decision.Source))
		}
	}
	return nil
}
