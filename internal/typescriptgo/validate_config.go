package typescriptgo

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// ValidateCompilerOptions verifies that compilerOptions in a tsconfig.json satisfy
// ScriptGo's native compilation and type safety invariants.
func ValidateCompilerOptions(opts *core.CompilerOptions, configPath string) []Diagnostic {
	if opts == nil {
		return nil
	}

	var diags []Diagnostic

	// SG6001: Strict mode invariant
	if opts.Strict != core.TSTrue {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6001,
			Message:  "ScriptGo requires 'compilerOptions.strict: true' in tsconfig.json for native type safety",
		})
	} else if opts.NoImplicitAny == core.TSFalse {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6001,
			Message:  "ScriptGo requires 'compilerOptions.noImplicitAny' to not be disabled ('false')",
		})
	}

	// SG6002: Minimum target invariant (>= ES2020)
	if opts.Target > 0 && opts.Target < core.ScriptTargetES2020 {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6002,
			Message:  fmt.Sprintf("ScriptGo requires 'compilerOptions.target' to be 'ES2020' or higher (found %q) to preserve native class and async semantics", opts.Target.String()),
		})
	}

	// SG6003: Unsupported moduleResolution
	if opts.ModuleResolution == core.ModuleResolutionKindClassic {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6003,
			Message:  "'compilerOptions.moduleResolution: \"classic\"' is unsupported; use 'Bundler', 'NodeNext', or 'Node16'",
		})
	} else if opts.ModuleResolution == core.ModuleResolutionKindNode10 {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6003,
			Message:  "'compilerOptions.moduleResolution: \"node10\"' (node) is unsupported; use 'Bundler', 'NodeNext', or 'Node16'",
		})
	}

	// SG6004: Unsupported module kinds
	if opts.Module == core.ModuleKindAMD || opts.Module == core.ModuleKindUMD || opts.Module == core.ModuleKindSystem {
		diags = append(diags, Diagnostic{
			FileName: configPath,
			Kind:     "config",
			Code:     6004,
			Message:  fmt.Sprintf("'compilerOptions.module: %q' is unsupported for native compilation; use 'ESNext', 'NodeNext', 'Node16', or 'CommonJS'", opts.Module.String()),
		})
	}

	return diags
}
