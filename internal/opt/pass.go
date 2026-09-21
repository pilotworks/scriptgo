// Package opt provides middle-end optimization passes over ScriptGo typed IR.
package opt

import "github.com/pilotworks/scriptgo/internal/ir"

// Pass represents an optimization transformation on an IR module.
// Run returns true if the pass modified the module.
type Pass interface {
	Name() string
	Run(m *ir.Module) (bool, error)
}
