package llvm

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) isInteger(name string) bool {
	if e.integerVars == nil {
		return false
	}
	clean := strings.TrimPrefix(name, "%")
	return e.integerVars[clean] || e.integerVars[name]
}

func (e *functionEmitter) integerUpperBound(name string) (float64, bool) {
	if e.integerUpperBounds == nil {
		return 0, false
	}
	clean := strings.TrimPrefix(name, "%")
	upper, ok := e.integerUpperBounds[clean]
	if !ok || upper < 0 || upper > float64(1<<53-1) {
		return 0, false
	}
	return upper, true
}

func (e *functionEmitter) setIntegerUpperBound(name string, upper float64) {
	if e.integerUpperBounds == nil || upper < 0 || upper > float64(1<<53-1) {
		return
	}
	e.integerUpperBounds[name] = upper
}

func (e *functionEmitter) emitConst(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	switch instruction.Type {
	case ir.TypeNumber:
		if bits, ok := numberMarkerBits[instruction.Value]; ok {
			// Bit-exact: undefined, null and NaN are distinct NaN payloads.
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %s to double\n", instruction.Result, bits))
			return nil
		}
		number, err := strconv.ParseFloat(instruction.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid number %q: %w", instruction.Value, err)
		}
		if e.integerVars != nil && number == math.Floor(number) && !math.IsNaN(number) && !math.IsInf(number, 0) {
			e.integerVars[instruction.Result] = true
		}
		if number >= 0 && number == math.Floor(number) && number <= float64(1<<53-1) {
			e.setIntegerUpperBound(instruction.Result, number)
		}
		out.WriteString(fmt.Sprintf("  %%%s = fadd double -0.0, %s\n", instruction.Result, llvmNumber(number)))
	case ir.TypeString:
		if !instruction.StringLiteral && instruction.Value == "null" {
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", instruction.Result))
			return nil
		}
		if !instruction.StringLiteral && instruction.Value == "undefined" {
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr i8, ptr @scriptgo_undefined_sentinel, i64 0\n", instruction.Result))
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
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr i8, ptr @scriptgo_undefined_sentinel, i64 0\n", instruction.Result))
	case ir.TypeUnknown:
		tag := 0
		if instruction.Value == "null" {
			tag = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", instruction.Result, tag))
	default:
		if llvmType(instruction.Type) == "ptr" {
			if instruction.Value == "undefined" {
				out.WriteString(fmt.Sprintf("  %%%s = getelementptr i8, ptr @scriptgo_undefined_sentinel, i64 0\n", instruction.Result))
				return nil
			}
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
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar))
		arg0 = numVar
		leftType = ir.TypeNumber
	}
	if instruction.Type == ir.TypeNumber && rightType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		numVar := fmt.Sprintf("num.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar))
		arg1 = numVar
		rightType = ir.TypeNumber
	}
	if instruction.Type == ir.TypeString && leftType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		strVar := fmt.Sprintf("str.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", strVar, payloadVar))
		arg0 = strVar
		leftType = ir.TypeString
	}
	if instruction.Type == ir.TypeString && rightType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		strVar := fmt.Sprintf("str.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg1))
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
		if instruction.Operator == "**" {
			e.types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = call i64 @scriptgo_bigint_pow(i64 %%%s, i64 %%%s)\n", instruction.Result, arg0, arg1))
			return nil
		}
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
		out.WriteString(fmt.Sprintf("  %%%s = call double @scriptgo_math_pow(double %%%s, double %%%s)\n", instruction.Result, arg0, arg1))
		return nil
	}
	if op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		if (instruction.Operator == "+" || instruction.Operator == "-" || instruction.Operator == "*" || instruction.Operator == "%") &&
			e.isInteger(instruction.Args[0]) && e.isInteger(instruction.Args[1]) {
			e.integerVars[instruction.Result] = true
		}
		out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, arg0, arg1))
		if instruction.Operator == "+" || instruction.Operator == "*" {
			if leftUpper, ok := e.integerUpperBound(instruction.Args[0]); ok {
				if rightUpper, ok := e.integerUpperBound(instruction.Args[1]); ok {
					upper := leftUpper + rightUpper
					if instruction.Operator == "*" {
						upper = leftUpper * rightUpper
					}
					e.setIntegerUpperBound(instruction.Result, upper)
				}
			}
		}
		return nil
	}
	if bitOp, ok := map[string]string{"&": "and", "|": "or", "^": "xor"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		if e.integerVars != nil {
			e.integerVars[instruction.Result] = true
		}
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		leftUpper, leftIsSafeUnsigned := e.integerUpperBound(instruction.Args[0])
		rightUpper, rightIsSafeUnsigned := e.integerUpperBound(instruction.Args[1])
		if leftIsSafeUnsigned {
			lI64 := instruction.Result + ".l_i64"
			out.WriteString(fmt.Sprintf("  %%%s = fptoui double %%%s to i64\n", lI64, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i32\n", lI32, lI64))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", lI32, arg0))
		}
		if rightIsSafeUnsigned {
			rI64 := instruction.Result + ".r_i64"
			out.WriteString(fmt.Sprintf("  %%%s = fptoui double %%%s to i64\n", rI64, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i32\n", rI32, rI64))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", rI32, arg1))
		}
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, bitOp, lI32, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		if instruction.Operator == "&" && rightIsSafeUnsigned && rightUpper <= float64(1<<31-1) {
			e.setIntegerUpperBound(instruction.Result, rightUpper)
		} else if instruction.Operator == "&" && leftIsSafeUnsigned && leftUpper <= float64(1<<31-1) {
			e.setIntegerUpperBound(instruction.Result, leftUpper)
		}
		return nil
	}
	if shiftOp, ok := map[string]string{"<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		if e.integerVars != nil {
			e.integerVars[instruction.Result] = true
		}
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		shift := instruction.Result + ".shift"
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", lI32, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", rI32, arg1))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, shiftOp, lI32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if instruction.Operator == ">>>" {
		e.types[instruction.Result] = instruction.Type
		if e.integerVars != nil {
			e.integerVars[instruction.Result] = true
		}
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resU32 := instruction.Result + ".res_u32"
		shift := instruction.Result + ".shift"
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", lI32, arg0))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @__scriptgo_to_int32(double %%%s)\n", rI32, arg1))
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

func (e *functionEmitter) coerceSelectOperand(out *strings.Builder, arg string, argType ir.Type, targetLT string) (string, error) {
	currLT := llvmType(argType)
	if currLT == targetLT {
		return arg, nil
	}
	if targetLT == "{ i32, i32, i64, i64 }" {
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
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg))
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
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, arg))
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
