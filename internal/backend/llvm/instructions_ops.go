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
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 0, 0\n", instruction.Result))
	default:
		if strings.HasPrefix(string(instruction.Type), "object:") || instruction.Type == "ptr" || instruction.Type == ir.TypeClosure || strings.HasSuffix(string(instruction.Type), "[]") || instruction.Type == ir.TypeUint8Array || instruction.Type == ir.TypeInt32Array || instruction.Type == ir.TypeFloat64Array || instruction.Type == ir.TypeArrayBuffer {
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
		return fmt.Errorf("unknown binary value %q", instruction.Args[0])
	}
	if _, ok := e.types[instruction.Args[1]]; !ok {
		return fmt.Errorf("unknown binary value %q", instruction.Args[1])
	}
	if leftType == ir.TypeString && instruction.Operator == "+" {
		e.types[instruction.Result] = ir.TypeString
		slot := instruction.Result + ".slot"
		status := instruction.Result + ".status"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
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
		out.WriteString(fmt.Sprintf("  %%%s = %s i1 %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if leftType == ir.TypeBigInt {
		if op, ok := map[string]string{"+": "add", "-": "sub", "*": "mul", "/": "sdiv", "%": "srem", "&": "and", "|": "or", "^": "xor", "<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
			e.types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = %s i64 %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
			return nil
		}
		return fmt.Errorf("unsupported LLVM bigint binary operator %q", instruction.Operator)
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM binary operator %q only supports number, bool, or string concatenation", instruction.Operator)
	}
	if instruction.Operator == "**" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call double @llvm.pow.f64(double %%%s, double %%%s)\n", instruction.Result, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if bitOp, ok := map[string]string{"&": "and", "|": "or", "^": "xor"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, bitOp, lI32, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if shiftOp, ok := map[string]string{"<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		shift := instruction.Result + ".shift"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, shiftOp, lI32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if instruction.Operator == ">>>" {
		e.types[instruction.Result] = instruction.Type
		lU32 := instruction.Result + ".l_u32"
		rU32 := instruction.Result + ".r_u32"
		shift := instruction.Result + ".shift"
		resU32 := instruction.Result + ".res_u32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lU32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rU32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rU32))
		out.WriteString(fmt.Sprintf("  %%%s = lshr i32 %%%s, %%%s\n", resU32, lU32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = uitofp i32 %%%s to double\n", instruction.Result, resU32))
		return nil
	}
	if instruction.Operator == "||" {
		e.types[instruction.Result] = instruction.Type
		cmp := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = fcmp one double %%%s, 0.0\n", cmp, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, double %%%s, double %%%s\n", instruction.Result, cmp, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if instruction.Operator == "&&" {
		e.types[instruction.Result] = instruction.Type
		cmp := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = fcmp one double %%%s, 0.0\n", cmp, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, double %%%s, double %%%s\n", instruction.Result, cmp, instruction.Args[1], instruction.Args[0]))
		return nil
	}
	return fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
}

func (e *functionEmitter) emitCompare(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
	if !ok || e.types[instruction.Args[1]] != leftType {
		return fmt.Errorf("unknown or mismatched compare operands")
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

		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag0, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag1, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %%%s\n", tagEq, tag0, tag1))

		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", val0, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", val1, instruction.Args[1]))
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
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_compare(ptr %%%s, ptr %%%s)\n", cmpResult, instruction.Args[0], instruction.Args[1]))
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
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i1 %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
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
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i64 %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if leftType == ir.TypeSymbol || leftType == ir.TypeClosure || strings.HasPrefix(string(leftType), "object:") || leftType == "ptr" || leftType == ir.TypeArrayBuffer || leftType == ir.TypeBuffer || leftType == ir.TypeDataView || leftType == ir.TypeTextEncoder || leftType == ir.TypeTextDecoder || leftType == ir.TypeMap || leftType == ir.TypeSet || strings.HasSuffix(string(leftType), "[]") {
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM pointer/closure compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s ptr %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
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
	out.WriteString(fmt.Sprintf("  %%%s = fcmp %s double %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
	return nil
}

func (e *functionEmitter) emitSelect(out *strings.Builder, instruction ir.Instruction) error {
	if e.types[instruction.Args[0]] != ir.TypeBool || e.types[instruction.Args[1]] != e.types[instruction.Args[2]] {
		return fmt.Errorf("select operands have incompatible types")
	}
	e.types[instruction.Result] = instruction.Type
	out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, %s %%%s, %s %%%s\n", instruction.Result, instruction.Args[0], llvmType(instruction.Type), instruction.Args[1], llvmType(instruction.Type), instruction.Args[2]))
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
		loaded := fmt.Sprintf("%s.box_load.%d", arg, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(argType), slot))
		argVal = loaded
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
	if e.types[arg] != ir.TypeUnknown && e.types[arg] != "any" {
		argVal := arg
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.cast_load.%d", arg, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", loaded, slot))
			argVal = loaded
		}
		out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, argVal))
		return nil
	}
	id := e.labelCounter
	e.labelCounter++

	var expectedTag int
	switch instruction.Type {
	case ir.TypeNumber:
		expectedTag = 3
	case ir.TypeBool:
		expectedTag = 2
	case ir.TypeString:
		expectedTag = 4
	case ir.TypeVoid:
		expectedTag = 0
	case ir.TypeClosure:
		expectedTag = 7
	case ir.TypeNumberArray, ir.TypeStringArray:
		expectedTag = 6
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

	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", rawPayload, arg))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %d\n", cmpVar, tagVar, expectedTag))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 0\n", isNullVar, rawPayload))
	out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", cmpFinal, cmpVar, isNullVar))
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", cmpFinal, castOk, castFail))

	out.WriteString(fmt.Sprintf("\n%s:\n", castFail))
	out.WriteString(fmt.Sprintf("  call void @__scriptgo_fail_checked_cast(i32 %%%s, i32 %d, ptr null)\n", tagVar, expectedTag))
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
	return nil
}

func (e *functionEmitter) emitTypeOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeString
	arg := instruction.Args[0]
	argType, ok := e.types[arg]
	if ok && argType != ir.TypeUnknown {
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

	tagVar := fmt.Sprintf("typeof.tag.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg))
	out.WriteString(fmt.Sprintf("  %%%s = call ptr @__scriptgo_typeof_unknown(i32 %%%s)\n", instruction.Result, tagVar))
	return nil
}
