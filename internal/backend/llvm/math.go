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
	case "__Math.abs", "__Math.ceil", "__Math.floor", "__Math.trunc", "__Math.sqrt", "__Math.round", "__Math.sin", "__Math.cos", "__Math.log", "__Math.log2", "__Math.log10", "__Math.exp", "__Math.tan", "__Math.atan", "__Math.asin", "__Math.acos", "__Math.cbrt", "__Math.sinh", "__Math.cosh", "__Math.tanh", "__Math.asinh", "__Math.acosh", "__Math.atanh", "__Math.expm1", "__Math.log1p":
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
			"__Math.log":   "llvm.log.f64",
			"__Math.log2":  "llvm.log2.f64",
			"__Math.log10": "llvm.log10.f64",
			"__Math.exp":   "llvm.exp.f64",
			"__Math.tan":   "tan",
			"__Math.atan":  "atan",
			"__Math.asin":  "asin",
			"__Math.acos":  "acos",
			"__Math.cbrt":  "cbrt",
			"__Math.sinh":  "sinh",
			"__Math.cosh":  "cosh",
			"__Math.tanh":  "tanh",
			"__Math.asinh": "asinh",
			"__Math.acosh": "acosh",
			"__Math.atanh": "atanh",
			"__Math.expm1": "expm1",
			"__Math.log1p": "log1p",
		}[instruction.Callee]
		fmt.Fprintf(out, "  %%%s = call double @%s(double %%%s)\n", instruction.Result, name, instruction.Args[0])
		return nil
	case "__Math.fround":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("math intrinsic %q requires one number", instruction.Callee)
		}
		f32 := instruction.Result + "_f32"
		fmt.Fprintf(out, "  %%%s = fptrunc double %%%s to float\n", f32, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = fpext float %%%s to double\n", instruction.Result, f32)
		return nil
	case "__Math.f16round":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("math intrinsic %q requires one number", instruction.Callee)
		}
		f16 := instruction.Result + "_f16"
		fmt.Fprintf(out, "  %%%s = fptrunc double %%%s to half\n", f16, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = fpext half %%%s to double\n", instruction.Result, f16)
		return nil
	case "__Math.clz32":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("math intrinsic %q requires one number", instruction.Callee)
		}
		u32 := instruction.Result + "_u32"
		clz := instruction.Result + "_clz"
		fmt.Fprintf(out, "  %%%s = fptoui double %%%s to i32\n", u32, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @llvm.ctlz.i32(i32 %%%s, i1 false)\n", clz, u32)
		fmt.Fprintf(out, "  %%%s = uitofp i32 %%%s to double\n", instruction.Result, clz)
		return nil
	case "__Math.imul":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("math intrinsic %q requires two numbers", instruction.Callee)
		}
		a32 := instruction.Result + "_a32"
		b32 := instruction.Result + "_b32"
		mul := instruction.Result + "_mul"
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i32\n", a32, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i32\n", b32, instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = mul i32 %%%s, %%%s\n", mul, a32, b32)
		fmt.Fprintf(out, "  %%%s = sitofp i32 %%%s to double\n", instruction.Result, mul)
		return nil
	case "__Math.sign":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("math intrinsic %q requires one number", instruction.Callee)
		}
		gt := instruction.Result + "_gt"
		lt := instruction.Result + "_lt"
		neg := instruction.Result + "_neg"
		fmt.Fprintf(out, "  %%%s = fcmp ogt double %%%s, 0.0\n", gt, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = fcmp olt double %%%s, 0.0\n", lt, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = select i1 %%%s, double -1.0, double %%%s\n", neg, lt, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = select i1 %%%s, double 1.0, double %%%s\n", instruction.Result, gt, neg)
		return nil
	case "__Math.random":
		fmt.Fprintf(out, "  %%%s = call double @drand48()\n", instruction.Result)
		return nil
	case "__Math.min", "__Math.max", "__Math.pow", "__Math.atan2", "__Math.hypot":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("math intrinsic %q requires two numbers", instruction.Callee)
		}
		name := map[string]string{
			"__Math.min":   "llvm.minnum.f64",
			"__Math.max":   "llvm.maxnum.f64",
			"__Math.pow":   "llvm.pow.f64",
			"__Math.atan2": "atan2",
			"__Math.hypot": "hypot",
		}[instruction.Callee]
		fmt.Fprintf(out, "  %%%s = call double @%s(double %%%s, double %%%s)\n", instruction.Result, name, instruction.Args[0], instruction.Args[1])
		return nil
	default:
		return fmt.Errorf("unknown math intrinsic %q", instruction.Callee)
	}
}
