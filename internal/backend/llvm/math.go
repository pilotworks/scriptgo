package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitMathIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
		return fmt.Errorf("math intrinsic %q has invalid signature", instruction.Callee)
	}
	name := map[string]string{
		"__Math.abs":   "llvm.fabs.f64",
		"__Math.ceil":  "llvm.ceil.f64",
		"__Math.floor": "llvm.floor.f64",
		"__Math.trunc": "llvm.trunc.f64",
	}[instruction.Callee]
	if name == "" {
		return fmt.Errorf("unknown math intrinsic %q", instruction.Callee)
	}
	fmt.Fprintf(out, "  %%%s = call double @%s(double %%%s)\n", instruction.Result, name, instruction.Args[0])
	return nil
}
