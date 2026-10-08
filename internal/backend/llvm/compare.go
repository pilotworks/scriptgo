package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitCompare(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
	if e.isParamUnknown(instruction.Args[0]) {
		leftType = ir.TypeUnknown
		ok = true
	}
	if !ok {
		for _, g := range e.module.Globals {
			if g.Name == instruction.Args[0] {
				leftType = g.Type
				ok = true
				break
			}
		}
	}
	rightType, okR := e.types[instruction.Args[1]]
	if e.isParamUnknown(instruction.Args[1]) {
		rightType = ir.TypeUnknown
		okR = true
	}
	if !okR {
		for _, g := range e.module.Globals {
			if g.Name == instruction.Args[1] {
				rightType = g.Type
				break
			}
		}
	}
	arg0 := e.resolveArg(out, instruction.Args[0])
	arg1 := e.resolveArg(out, instruction.Args[1])
	if leftType == ir.TypeUnknown && rightType != ir.TypeUnknown {
		if rightType == ir.TypeVoid {
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			b1 := fmt.Sprintf("box.b1.%d", e.loadCounter)
			b2 := fmt.Sprintf("box.b2.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 0, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 0, 2\n", b2, b1))
			arg1 = b2
			rightType = ir.TypeUnknown
		} else if rightType == ir.TypeString {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg0 = ptrVar
			leftType = ir.TypeString
		} else if strings.HasPrefix(string(rightType), "object:") || rightType == ir.TypeObject || llvmType(rightType) == "ptr" || rightType == ir.TypePointer {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg0 = ptrVar
			leftType = "ptr"
		}
	}
	if rightType == ir.TypeUnknown && leftType != ir.TypeUnknown {
		if leftType == ir.TypeVoid {
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			b1 := fmt.Sprintf("box.b1.%d", e.loadCounter)
			b2 := fmt.Sprintf("box.b2.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 0, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 0, 2\n", b2, b1))
			arg0 = b2
			leftType = ir.TypeUnknown
		} else if leftType == ir.TypeString {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg1 = ptrVar
			rightType = ir.TypeString
		} else if strings.HasPrefix(string(leftType), "object:") || leftType == ir.TypeObject || llvmType(leftType) == "ptr" || leftType == ir.TypePointer {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg1 = ptrVar
			rightType = "ptr"
		}
	}
	if instruction.Operator == "==" || instruction.Operator == "===" || instruction.Operator == "!=" || instruction.Operator == "!==" {
		if leftType == ir.TypeNumber && (strings.HasPrefix(string(rightType), "object:") || rightType == ir.TypeObject || llvmType(rightType) == "ptr") {
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", ptrVar))
			arg0 = ptrVar
			leftType = "ptr"
		}
		if rightType == ir.TypeNumber && (strings.HasPrefix(string(leftType), "object:") || leftType == ir.TypeObject || llvmType(leftType) == "ptr") {
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", ptrVar))
			arg1 = ptrVar
			rightType = "ptr"
		}
	}
	if !ok || (rightType != leftType && !((strings.HasPrefix(string(leftType), "object:") || leftType == ir.TypeObject || llvmType(leftType) == "ptr") && (strings.HasPrefix(string(rightType), "object:") || rightType == ir.TypeObject || llvmType(rightType) == "ptr"))) {
		return fmt.Errorf("unknown or mismatched compare operands (left=%s, right=%s)", leftType, rightType)
	}
	if leftType == ir.TypeUnknown {
		slot := fmt.Sprintf("%s.unknown_eq.slot", instruction.Result)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		value0, err := e.emitCanonicalValuePointer(out, arg0, ir.TypeUnknown, fmt.Sprintf("%s.eq.0.%d", instruction.Result, e.loadCounter))
		if err != nil {
			return err
		}
		value1, err := e.emitCanonicalValuePointer(out, arg1, ir.TypeUnknown, fmt.Sprintf("%s.eq.1.%d", instruction.Result, e.loadCounter))
		if err != nil {
			return err
		}
		loose := 0
		if instruction.Operator == "==" || instruction.Operator == "!=" {
			loose = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_equals_unknown(ptr %s, ptr %s, i32 %d, ptr %%%s)\n", status, value0, value1, loose, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		eqValue := fmt.Sprintf("%s.unknown_eq", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", eqValue, slot))

		e.types[instruction.Result] = ir.TypeBool
		if instruction.Operator == "==" || instruction.Operator == "===" {
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, eqValue))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", instruction.Result, eqValue))
		}
		return nil
	}
	if leftType == ir.TypeString {
		cmpResult := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_compare(ptr %%%s, ptr %%%s)\n", cmpResult, arg0, arg1))
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
			"<": "slt", "<=": "sle",
			">": "sgt", ">=": "sge",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM string compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i32 %%%s, 0\n", instruction.Result, predicate, cmpResult))
		return nil
	}
	if leftType == ir.TypeBool {
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM bool compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i1 %%%s, %%%s\n", instruction.Result, predicate, arg0, arg1))
		return nil
	}
	if leftType == ir.TypeBigInt {
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
			"<": "slt", "<=": "sle",
			">": "sgt", ">=": "sge",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM bigint compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i64 %%%s, %%%s\n", instruction.Result, predicate, arg0, arg1))
		return nil
	}
	if isFunctionType(leftType) || isFunctionType(rightType) || leftType == ir.TypeClosure || rightType == ir.TypeClosure {
		if instruction.Operator == "==" || instruction.Operator == "!=" {
			p0Null := fmt.Sprintf("%s.p0_null.%d", instruction.Result, e.loadCounter)
			p0Undef := fmt.Sprintf("%s.p0_undef.%d", instruction.Result, e.loadCounter)
			p0Nullish := fmt.Sprintf("%s.p0_nullish.%d", instruction.Result, e.loadCounter)
			p1Null := fmt.Sprintf("%s.p1_null.%d", instruction.Result, e.loadCounter)
			p1Undef := fmt.Sprintf("%s.p1_undef.%d", instruction.Result, e.loadCounter)
			p1Nullish := fmt.Sprintf("%s.p1_nullish.%d", instruction.Result, e.loadCounter)
			bothNullish := fmt.Sprintf("%s.both_nullish.%d", instruction.Result, e.loadCounter)
			rawEq := fmt.Sprintf("%s.raw_eq.%d", instruction.Result, e.loadCounter)
			looseEq := fmt.Sprintf("%s.loose_eq.%d", instruction.Result, e.loadCounter)
			e.loadCounter++

			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", p0Null, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", p0Undef, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", p0Nullish, p0Null, p0Undef))

			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", p1Null, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", p1Undef, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", p1Nullish, p1Null, p1Undef))

			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", bothNullish, p0Nullish, p1Nullish))
			eqVar := fmt.Sprintf("closure.eq.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_equals(ptr %%%s, ptr %%%s)\n", eqVar, arg0, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", rawEq, eqVar))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", looseEq, rawEq, bothNullish))

			if instruction.Operator == "==" {
				out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %%%s\n", instruction.Result, looseEq))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = xor i1 %%%s, true\n", instruction.Result, looseEq))
			}
			e.types[instruction.Result] = ir.TypeBool
			return nil
		}
		eqVar := fmt.Sprintf("closure.eq.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_equals(ptr %%%s, ptr %%%s)\n", eqVar, arg0, arg1))
		if instruction.Operator == "===" {
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, eqVar))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", instruction.Result, eqVar))
		}
		e.types[instruction.Result] = ir.TypeBool
		return nil
	}
	if leftType == ir.TypeObject || leftType == ir.TypeSymbol || strings.HasPrefix(string(leftType), "object:") || leftType == "ptr" || leftType == ir.TypePointer || leftType == ir.TypeArrayBuffer || leftType == ir.TypeBuffer || isTypedArrayType(leftType) || leftType == ir.TypeDataView || leftType == ir.TypeTextEncoder || leftType == ir.TypeTextDecoder || leftType == ir.TypeMap || leftType == ir.TypeSet || strings.HasSuffix(string(leftType), "[]") {
		e.types[instruction.Result] = ir.TypeBool
		if instruction.Operator == "==" || instruction.Operator == "!=" {
			p0Null := fmt.Sprintf("%s.p0_null.%d", instruction.Result, e.loadCounter)
			p0Undef := fmt.Sprintf("%s.p0_undef.%d", instruction.Result, e.loadCounter)
			p0Nullish := fmt.Sprintf("%s.p0_nullish.%d", instruction.Result, e.loadCounter)
			p1Null := fmt.Sprintf("%s.p1_null.%d", instruction.Result, e.loadCounter)
			p1Undef := fmt.Sprintf("%s.p1_undef.%d", instruction.Result, e.loadCounter)
			p1Nullish := fmt.Sprintf("%s.p1_nullish.%d", instruction.Result, e.loadCounter)
			bothNullish := fmt.Sprintf("%s.both_nullish.%d", instruction.Result, e.loadCounter)
			rawEq := fmt.Sprintf("%s.raw_eq.%d", instruction.Result, e.loadCounter)
			looseEq := fmt.Sprintf("%s.loose_eq.%d", instruction.Result, e.loadCounter)
			e.loadCounter++

			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", p0Null, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", p0Undef, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", p0Nullish, p0Null, p0Undef))

			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", p1Null, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", p1Undef, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", p1Nullish, p1Null, p1Undef))

			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", bothNullish, p0Nullish, p1Nullish))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, %%%s\n", rawEq, arg0, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", looseEq, rawEq, bothNullish))

			if instruction.Operator == "==" {
				out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %%%s\n", instruction.Result, looseEq))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = xor i1 %%%s, true\n", instruction.Result, looseEq))
			}
			return nil
		}
		predicate, ok := map[string]string{
			"===": "eq",
			"!==": "ne",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM pointer compare operator %q", instruction.Operator)
		}
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s ptr %%%s, %%%s\n", instruction.Result, predicate, arg0, arg1))
		return nil
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM compare only supports number, string, symbol, closure, object, or bool operands (got %s)", leftType)
	}
	predicate, ok := map[string]string{
		"==": "oeq", "===": "oeq",
		"!=": "une", "!==": "une",
		"<": "olt", "<=": "ole",
		">": "ogt", ">=": "oge",
	}[instruction.Operator]
	if !ok {
		return fmt.Errorf("unsupported LLVM number compare operator %q", instruction.Operator)
	}
	e.types[instruction.Result] = ir.TypeBool
	out.WriteString(fmt.Sprintf("  %%%s = fcmp %s double %%%s, %%%s\n", instruction.Result, predicate, arg0, arg1))
	return nil
}
