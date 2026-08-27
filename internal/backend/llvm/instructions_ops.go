package llvm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitConst(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	switch instruction.Type {
	case ir.TypeNumber:
		number, err := strconv.ParseFloat(instruction.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid number %q: %w", instruction.Value, err)
		}
		out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, %s\n", instruction.Result, llvmNumber(number)))
	case ir.TypeString:
		if instruction.Value == "null" {
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", instruction.Result))
			return nil
		}
		global := e.stringsByValue[instruction.Value]
		length := len([]byte(instruction.Value)) + 1
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", instruction.Result, length, global))
	case ir.TypeBigInt:
		clean := strings.TrimSuffix(instruction.Value, "n")
		out.WriteString(fmt.Sprintf("  %%%s = add i64 0, %s\n", instruction.Result, clean))
	case ir.TypeSymbol:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		strGlobal := e.stringsByValue[instruction.Value]
		length := len([]byte(instruction.Value)) + 1
		strPtr := fmt.Sprintf("%s.name", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", strPtr, length, strGlobal))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_symbol_for(ptr %%%s, ptr %%%s)\n", status, strPtr, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	case ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %s\n", instruction.Result, instruction.Value))
	case ir.TypeVoid:
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", instruction.Result))
	case ir.TypeUnknown:
		tag := 0
		if instruction.Value == "null" {
			tag = 5
		}
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 %d, 0\n", instruction.Result, tag))
	default:
		if llvmType(instruction.Type) == "ptr" {
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", instruction.Result))
			return nil
		}
		return fmt.Errorf("unsupported constant type %s", instruction.Type)
	}
	return nil
}

func (e *functionEmitter) emitBinary(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
	if !ok {
		for _, g := range e.module.Globals {
			if g.Name == instruction.Args[0] {
				leftType = g.Type
				ok = true
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("unknown binary value %q", instruction.Args[0])
	}
	rightType, ok := e.types[instruction.Args[1]]
	if !ok {
		for _, g := range e.module.Globals {
			if g.Name == instruction.Args[1] {
				rightType = g.Type
				ok = true
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("unknown binary value %q", instruction.Args[1])
	}
	_ = rightType
	arg0 := e.resolveArg(out, instruction.Args[0])
	arg1 := e.resolveArg(out, instruction.Args[1])
	if instruction.Type == ir.TypeNumber && leftType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		numVar := fmt.Sprintf("num.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar))
		arg0 = numVar
		leftType = ir.TypeNumber
	}
	if instruction.Type == ir.TypeNumber && rightType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		numVar := fmt.Sprintf("num.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar))
		arg1 = numVar
		rightType = ir.TypeNumber
	}
	if instruction.Type == ir.TypeString && leftType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		strVar := fmt.Sprintf("str.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", strVar, payloadVar))
		arg0 = strVar
		leftType = ir.TypeString
	}
	if instruction.Type == ir.TypeString && rightType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		strVar := fmt.Sprintf("str.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", strVar, payloadVar))
		arg1 = strVar
	}
	if leftType == ir.TypeString && instruction.Operator == "+" {
		e.types[instruction.Result] = ir.TypeString
		id := e.loadCounter
		e.loadCounter++
		slot := fmt.Sprintf("concat.slot.%d", id)
		status := fmt.Sprintf("concat.status.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arg0, arg1, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.ownedStrings = append(e.ownedStrings, instruction.Result)
		return nil
	}
	if leftType == ir.TypeBool {
		op, ok := map[string]string{"&&": "and", "||": "or"}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM binary bool operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = %s i1 %%%s, %%%s\n", instruction.Result, op, arg0, arg1))
		return nil
	}
	if leftType == ir.TypeBigInt {
		if op, ok := map[string]string{"+": "add", "-": "sub", "*": "mul", "/": "sdiv", "%": "srem", "&": "and", "|": "or", "^": "xor", "<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
			e.types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = %s i64 %%%s, %%%s\n", instruction.Result, op, arg0, arg1))
			return nil
		}
		return fmt.Errorf("unsupported LLVM bigint binary operator %q", instruction.Operator)
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM binary operator %q only supports number, bool, or string concatenation", instruction.Operator)
	}
	if instruction.Operator == "**" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call double @llvm.pow.f64(double %%%s, double %%%s)\n", instruction.Result, arg0, arg1))
		return nil
	}
	if op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, arg0, arg1))
		return nil
	}
	if bitOp, ok := map[string]string{"&": "and", "|": "or", "^": "xor"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, bitOp, lI32, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if shiftOp, ok := map[string]string{"<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, shiftOp, lI32, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if instruction.Operator == ">>>" {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resU32 := instruction.Result + ".res_u32"
		shift := instruction.Result + ".shift"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = lshr i32 %%%s, %%%s\n", resU32, lI32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = uitofp i32 %%%s to double\n", instruction.Result, resU32))
		return nil
	}
	if instruction.Operator == "||" {
		e.types[instruction.Result] = instruction.Type
		cmp := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = fcmp one double %%%s, 0.0\n", cmp, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, double %%%s, double %%%s\n", instruction.Result, cmp, arg0, arg1))
		return nil
	}
	if instruction.Operator == "&&" {
		e.types[instruction.Result] = instruction.Type
		cmp := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = fcmp one double %%%s, 0.0\n", cmp, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, double %%%s, double %%%s\n", instruction.Result, cmp, arg1, arg0))
		return nil
	}
	return fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
}

func (e *functionEmitter) emitCompare(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 0, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 0, 2\n", b2, b1))
			arg1 = b2
			rightType = ir.TypeUnknown
		} else if rightType == ir.TypeString {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg0 = ptrVar
			leftType = ir.TypeString
		} else if strings.HasPrefix(string(rightType), "object:") || rightType == ir.TypeObject || llvmType(rightType) == "ptr" || rightType == ir.TypePointer {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg0))
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 0, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 0, 2\n", b2, b1))
			arg0 = b2
			leftType = ir.TypeUnknown
		} else if leftType == ir.TypeString {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			arg1 = ptrVar
			rightType = ir.TypeString
		} else if strings.HasPrefix(string(leftType), "object:") || leftType == ir.TypeObject || llvmType(leftType) == "ptr" || leftType == ir.TypePointer {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg1))
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
		tag0 := fmt.Sprintf("%s.tag0.%d", instruction.Result, e.loadCounter)
		tag1 := fmt.Sprintf("%s.tag1.%d", instruction.Result, e.loadCounter)
		val0 := fmt.Sprintf("%s.val0.%d", instruction.Result, e.loadCounter)
		val1 := fmt.Sprintf("%s.val1.%d", instruction.Result, e.loadCounter)
		tagEq := fmt.Sprintf("%s.tag_eq.%d", instruction.Result, e.loadCounter)
		valEq := fmt.Sprintf("%s.val_eq.%d", instruction.Result, e.loadCounter)
		bothEq := fmt.Sprintf("%s.both_eq.%d", instruction.Result, e.loadCounter)
		e.loadCounter++

		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag0, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag1, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %%%s\n", tagEq, tag0, tag1))

		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", val0, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", val1, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, %%%s\n", valEq, val0, val1))

		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", bothEq, tagEq, valEq))
		e.types[instruction.Result] = ir.TypeBool
		if instruction.Operator == "==" || instruction.Operator == "===" {
			out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %%%s\n", instruction.Result, bothEq))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = xor i1 %%%s, true\n", instruction.Result, bothEq))
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
	if leftType == ir.TypeClosure {
		eqVar := fmt.Sprintf("closure.eq.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_equals(ptr %%%s, ptr %%%s)\n", eqVar, arg0, arg1))
		if instruction.Operator == "==" || instruction.Operator == "===" {
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, eqVar))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", instruction.Result, eqVar))
		}
		e.types[instruction.Result] = ir.TypeBool
		return nil
	}
	if leftType == ir.TypeObject || leftType == ir.TypeSymbol || strings.HasPrefix(string(leftType), "object:") || leftType == "ptr" || leftType == ir.TypeArrayBuffer || leftType == ir.TypeBuffer || isTypedArrayType(leftType) || leftType == ir.TypeDataView || leftType == ir.TypeTextEncoder || leftType == ir.TypeTextDecoder || leftType == ir.TypeMap || leftType == ir.TypeSet || strings.HasSuffix(string(leftType), "[]") {
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM pointer compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
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

func (e *functionEmitter) coerceSelectOperand(out *strings.Builder, arg string, argType ir.Type, targetLT string) (string, error) {
	currLT := llvmType(argType)
	if currLT == targetLT {
		return arg, nil
	}
	if targetLT == "{ i32, i32, i64 }" {
		boxedVar := fmt.Sprintf("box.sel.%d", e.loadCounter)
		if err := e.emitBoxValue(out, arg, argType, boxedVar); err != nil {
			return "", err
		}
		return boxedVar, nil
	}
	if targetLT == "ptr" {
		if argType == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
			return ptrVar, nil
		}
		if argType == ir.TypeNumber || currLT == "double" {
			intVar := fmt.Sprintf("int.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = fptoui double %%%s to i64\n", intVar, arg))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, intVar))
			return ptrVar, nil
		}
		if argType == ir.TypeBool || currLT == "i1" {
			intVar := fmt.Sprintf("int.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i64\n", intVar, arg))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, intVar))
			return ptrVar, nil
		}
	}
	if targetLT == "double" {
		if argType == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			dblVar := fmt.Sprintf("dbl.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", dblVar, payloadVar))
			return dblVar, nil
		}
	}
	return arg, nil
}

func (e *functionEmitter) emitSelect(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	lt := llvmType(instruction.Type)
	if lt == "" {
		lt = "ptr"
	}
	cond := e.resolveArg(out, instruction.Args[0])
	arg1 := e.resolveArg(out, instruction.Args[1])
	t1 := e.types[arg1]
	if t1 == "" {
		t1 = e.types[instruction.Args[1]]
	}
	var err error
	arg1, err = e.coerceSelectOperand(out, arg1, t1, lt)
	if err != nil {
		return err
	}
	arg2 := e.resolveArg(out, instruction.Args[2])
	t2 := e.types[arg2]
	if t2 == "" {
		t2 = e.types[instruction.Args[2]]
	}
	arg2, err = e.coerceSelectOperand(out, arg2, t2, lt)
	if err != nil {
		return err
	}
	out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, %s %%%s, %s %%%s\n", instruction.Result, cond, lt, arg1, lt, arg2))
	return nil
}

func (e *functionEmitter) emitBoxValue(out *strings.Builder, argVal string, argType ir.Type, result string) error {
	id := e.loadCounter
	e.loadCounter++

	var tag int
	var payloadVal string

	switch argType {
	case ir.TypeNumber:
		tag = 3 // SCRIPTGO_TAG_NUMBER
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = bitcast double %%%s to i64\n", payloadVal, argVal))
	case ir.TypeBool:
		tag = 2 // SCRIPTGO_TAG_BOOLEAN
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i64\n", payloadVal, argVal))
	case ir.TypeString:
		tag = 4 // SCRIPTGO_TAG_STRING
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, argVal))
	case ir.TypeBigInt:
		tag = 8 // SCRIPTGO_TAG_BIGINT
		payloadVal = argVal
	case ir.TypeSymbol:
		tag = 9 // SCRIPTGO_TAG_SYMBOL
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, argVal))
	case ir.TypeVoid:
		tag = 0 // SCRIPTGO_TAG_UNDEFINED
		payloadVal = "0"
	case ir.TypeUnknown:
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", result, argVal))
		return nil
	default:
		if strings.HasSuffix(string(argType), "[]") || argType == ir.TypeNumberArray || argType == ir.TypeStringArray {
			tag = 6 // SCRIPTGO_TAG_ARRAY
		} else if argType == ir.TypeClosure {
			tag = 7 // SCRIPTGO_TAG_FUNCTION
		} else {
			tag = 5 // SCRIPTGO_TAG_OBJECT
		}
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, argVal))
	}

	b0 := fmt.Sprintf("box.b0.%d", id)
	b1 := fmt.Sprintf("box.b1.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 %d, 0\n", b0, tag))
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
	if payloadVal == "0" {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 0, 2\n", result, b1))
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", result, b1, payloadVal))
	}
	return nil
}

func (e *functionEmitter) emitBoxUnknown(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeUnknown
	arg := instruction.Args[0]
	argType := e.types[arg]

	argVal := arg
	if slot, ok := e.varSlots[arg]; ok {
		if argType != ir.TypeVoid && llvmType(argType) != "void" {
			loaded := fmt.Sprintf("%s.box_load.%d", arg, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot))
			argVal = loaded
		}
	}

	return e.emitBoxValue(out, argVal, argType, instruction.Result)
}

func (e *functionEmitter) emitCheckedCast(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	arg := instruction.Args[0]

	if instruction.Type == ir.TypeUnknown || instruction.Type == "any" {
		if e.types[arg] == ir.TypeUnknown || e.types[arg] == "any" {
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", instruction.Result, arg))
			return nil
		}
		return e.emitBoxValue(out, arg, e.types[arg], instruction.Result)
	}

	argType, hasType := e.types[arg]
	if hasType && argType != ir.TypeUnknown && argType != "any" {
		if instruction.Result != arg {
			srcType := llvmType(argType)
			dstType := llvmType(instruction.Type)
			if srcType == dstType {
				out.WriteString(fmt.Sprintf("  %%%s = bitcast %s %%%s to %s\n", instruction.Result, srcType, arg, dstType))
			} else if dstType == "ptr" && srcType != "ptr" {
				out.WriteString(fmt.Sprintf("  %%%s = inttoptr %s %%%s to ptr\n", instruction.Result, srcType, arg))
			} else if srcType == "ptr" && dstType != "ptr" {
				out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to %s\n", instruction.Result, arg, dstType))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = bitcast %s %%%s to %s\n", instruction.Result, srcType, arg, dstType))
			}
			if slot, ok := e.varSlots[instruction.Result]; ok {
				out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(instruction.Type), instruction.Result, slot))
			}
		}
		return nil
	}

	id := e.labelCounter
	e.labelCounter++

	var expectedTag int
	switch instruction.Type {
	case ir.TypeNumber:
		expectedTag = 3 // SCRIPTGO_TAG_NUMBER
	case ir.TypeBool:
		expectedTag = 2 // SCRIPTGO_TAG_BOOLEAN
	case ir.TypeString:
		expectedTag = 4 // SCRIPTGO_TAG_STRING
	case ir.TypeVoid:
		expectedTag = 0 // SCRIPTGO_TAG_UNDEFINED
	case ir.TypeClosure:
		expectedTag = 7 // SCRIPTGO_TAG_FUNCTION
	case ir.TypeBigInt:
		expectedTag = 8 // SCRIPTGO_TAG_BIGINT
	case ir.TypeSymbol:
		expectedTag = 9 // SCRIPTGO_TAG_SYMBOL
	case ir.TypeNumberArray, ir.TypeStringArray:
		expectedTag = 6 // SCRIPTGO_TAG_ARRAY
	default:
		if strings.HasSuffix(string(instruction.Type), "[]") {
			expectedTag = 6
		} else {
			expectedTag = 5
		}
	}

	tagVar := fmt.Sprintf("cast.tag.%d", id)
	rawPayload := fmt.Sprintf("cast.raw.%d", id)
	cmpVar := fmt.Sprintf("cast.cmp.%d", id)
	isNullVar := fmt.Sprintf("cast.isnull.%d", id)
	cmpFinal := fmt.Sprintf("cast.cmp_final.%d", id)
	castOk := fmt.Sprintf("cast_ok.%d", id)
	castFail := fmt.Sprintf("cast_fail.%d", id)

	argVal := arg
	if slot, ok := e.varSlots[arg]; ok {
		loaded := fmt.Sprintf("%s.cast_load.%d", arg, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
		argVal = loaded
	}
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, argVal))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", rawPayload, argVal))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %d\n", cmpVar, tagVar, expectedTag))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 0\n", isNullVar, rawPayload))
	out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", cmpFinal, cmpVar, isNullVar))
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", cmpFinal, castOk, castFail))

	out.WriteString(fmt.Sprintf("\n%s:\n", castFail))
	fnGlobal, ok := e.stringsByValue[e.function.Name]
	fnArg := "null"
	if ok {
		fnArg = fmt.Sprintf("getelementptr inbounds ([%d x i8], ptr %s, i64 0, i64 0)", len(e.function.Name)+1, fnGlobal)
	}
	out.WriteString(fmt.Sprintf("  call void @__scriptgo_fail_checked_cast(i32 %%%s, i32 %d, ptr %s)\n", tagVar, expectedTag, fnArg))
	out.WriteString("  unreachable\n")

	out.WriteString(fmt.Sprintf("\n%s:\n", castOk))

	switch instruction.Type {
	case ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", instruction.Result, rawPayload))
	case ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", instruction.Result, rawPayload))
	case ir.TypeString:
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, rawPayload))
	default:
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, rawPayload))
	}
	if slot, ok := e.varSlots[instruction.Result]; ok {
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(instruction.Type), instruction.Result, slot))
	}
	return nil
}

func (e *functionEmitter) emitTypeOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeString
	arg := instruction.Args[0]
	argType, ok := e.types[arg]
	if ok && argType != ir.TypeUnknown && argType != "any" && !strings.Contains(string(argType), "|") {
		if argType == ir.TypeClosure {
			nullPtr := fmt.Sprintf("typeof.null.%d", e.loadCounter)
			isNonNull := fmt.Sprintf("typeof.is_nonnull.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", nullPtr))
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, %%%s\n", isNonNull, arg, nullPtr))
			fnGlobal := e.stringsByValue["function"]
			undefGlobal := e.stringsByValue["undefined"]
			fnPtr := fmt.Sprintf("typeof.fn.%d", e.loadCounter)
			undefPtr := fmt.Sprintf("typeof.undef.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [9 x i8], ptr %s, i64 0, i64 0\n", fnPtr, fnGlobal))
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [10 x i8], ptr %s, i64 0, i64 0\n", undefPtr, undefGlobal))
			out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", instruction.Result, isNonNull, fnPtr, undefPtr))
			return nil
		}
		var typeStr string
		switch {
		case argType == ir.TypeNumber:
			typeStr = "number"
		case argType == ir.TypeString:
			typeStr = "string"
		case argType == ir.TypeBool:
			typeStr = "boolean"
		case argType == ir.TypeBigInt:
			typeStr = "bigint"
		case argType == ir.TypeSymbol:
			typeStr = "symbol"
		case argType == ir.TypeVoid:
			typeStr = "undefined"
		default:
			typeStr = "object"
		}
		if strGlobal, ok := e.stringsByValue[typeStr]; ok {
			length := len([]byte(typeStr)) + 1
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", instruction.Result, length, strGlobal))
			return nil
		}
	}
	id := e.loadCounter
	e.loadCounter++

	argVal := arg
	if slot, ok := e.varSlots[arg]; ok {
		loaded := fmt.Sprintf("%s.typeof_load.%d", arg, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
		argVal = loaded
	}

	tagVar := fmt.Sprintf("typeof.tag.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, argVal))
	out.WriteString(fmt.Sprintf("  %%%s = call ptr @__scriptgo_typeof_unknown(i32 %%%s)\n", instruction.Result, tagVar))
	return nil
}
