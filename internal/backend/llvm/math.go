package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitMathIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.Type != ir.TypeNumber {
		return fmt.Errorf("math intrinsic %q has invalid signature", instruction.Callee)
	}
	switch instruction.Callee {
	case "__Math.abs", "__Math.ceil", "__Math.floor", "__Math.trunc", "__Math.sqrt", "__Math.round", "__Math.sin", "__Math.cos":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("math intrinsic %q requires one number", instruction.Callee)
		}
		name := map[string]string{
			"__Math.abs":   "llvm.fabs.f64",
			"__Math.ceil":  "llvm.ceil.f64",
			"__Math.floor": "llvm.floor.f64",
			"__Math.trunc": "llvm.trunc.f64",
			"__Math.sqrt":  "llvm.sqrt.f64",
			"__Math.round": "llvm.round.f64",
			"__Math.sin":   "llvm.sin.f64",
			"__Math.cos":   "llvm.cos.f64",
		}[instruction.Callee]
		fmt.Fprintf(out, "  %%%s = call double @%s(double %%%s)\n", instruction.Result, name, instruction.Args[0])
		return nil
	case "__Math.min", "__Math.max", "__Math.pow":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("math intrinsic %q requires two numbers", instruction.Callee)
		}
		name := map[string]string{
			"__Math.min": "llvm.minnum.f64",
			"__Math.max": "llvm.maxnum.f64",
			"__Math.pow": "llvm.pow.f64",
		}[instruction.Callee]
		fmt.Fprintf(out, "  %%%s = call double @%s(double %%%s, double %%%s)\n", instruction.Result, name, instruction.Args[0], instruction.Args[1])
		return nil
	default:
		return fmt.Errorf("unknown math intrinsic %q", instruction.Callee)
	}
}
