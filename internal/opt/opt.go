package opt

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// Options controls optimization behavior.
type Options struct {
	Level string // "0", "1", "2", "3", "s", "z", "fast"
}

// Optimize runs optimization passes on the module according to options.
// It verifies the module invariants before returning.
func Optimize(m ir.Module, opts Options) (ir.Module, error) {
	if opts.Level == "0" {
		return m, nil
	}
	pipeline := buildPipeline(opts.Level)
	const maxIters = 5
	for iter := 0; iter < maxIters; iter++ {
		changed := false
		for _, pass := range pipeline {
			passChanged, err := pass.Run(&m)
			if err != nil {
				return m, fmt.Errorf("pass %s: %w", pass.Name(), err)
			}
			if err := m.Verify(); err != nil {
				return m, fmt.Errorf("pass %s produced invalid IR: %w", pass.Name(), err)
			}
			if passChanged {
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	if err := m.Verify(); err != nil {
		return m, fmt.Errorf("optimized module verification failed: %w", err)
	}
	return m, nil
}

func buildPipeline(level string) []Pass {
	switch level {
	case "0":
		return nil
	case "1":
		return []Pass{
			NewConstFoldPass(),
			NewBCEPass(),
			NewDCEPass(),
		}
	default: // "2", "3", "s", "z", "fast"
		return []Pass{
			NewConstFoldPass(),
			NewCSEPass(),
			NewLICMPass(),
			NewBCEPass(),
			NewTemporaryObjectRegionPass(),
			NewDCEPass(),
		}
	}
}
