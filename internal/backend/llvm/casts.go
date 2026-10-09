package llvm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitBoxValue(out *strings.Builder, argVal string, argType ir.Type, result string) error {
	id := e.loadCounter
	e.loadCounter++

	resolved := e.resolveArg(out, argVal)
	if argType == "" || argType == ir.TypeVoid {
		argType = e.types[argVal]
		if argType == "" {
			argType = e.types[resolved]
		}
	}
	argVal = resolved

	tag := 0
	tagOperand := ""
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
	case ir.TypePointer:
		// A raw pointer is nullish only when it is actually null. Non-null
		// pointers are native object references in the canonical value.
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, argVal))
		nonNull := fmt.Sprintf("nonnull.%d", id)
		tagValue := fmt.Sprintf("tag.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", nonNull, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, i32 5, i32 1\n", tagValue, nonNull))
		tagOperand = "%" + tagValue
	case ir.TypeUnknown:
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", result, argVal))
		return nil
	default:
		if strings.HasSuffix(string(argType), "[]") || argType == ir.TypeNumberArray || argType == ir.TypeStringArray {
			tag = 6 // SCRIPTGO_TAG_ARRAY
		} else if isFunctionType(argType) {
			tag = 7 // SCRIPTGO_TAG_FUNCTION
		} else {
			tag = 5 // SCRIPTGO_TAG_OBJECT
		}
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, argVal))
	}
	if tagOperand == "" {
		tagOperand = strconv.Itoa(tag)
	}

	b0 := fmt.Sprintf("box.b0.%d", id)
	b1 := fmt.Sprintf("box.b1.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %s, 0\n", b0, tagOperand))
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", b1, b0))
	if payloadVal == "0" {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 0, 2\n", result, b1))
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", result, b1, payloadVal))
	}
	return nil
}

func (e *functionEmitter) emitBoxUnknown(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeUnknown
	arg := instruction.Args[0]
	argType := e.types[arg]

	argVal := e.resolveArg(out, arg)
	if argType == "" || argType == ir.TypeVoid {
		argType = e.types[arg]
	}

	return e.emitBoxValue(out, argVal, argType, instruction.Result)
}

func (e *functionEmitter) emitCheckedCast(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	arg := instruction.Args[0]

	if instruction.Type == ir.TypeUnknown {
		if e.types[arg] == ir.TypeUnknown {
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", instruction.Result, arg))
			return nil
		}
		return e.emitBoxValue(out, arg, e.types[arg], instruction.Result)
	}

	argType, hasType := e.types[arg]
	if hasType && argType != ir.TypeUnknown {
		if instruction.Result != arg {
			if instruction.Type == ir.TypeVoid {
				return nil
			}
			srcType := llvmType(argType)
			dstType := llvmType(instruction.Type)
			if srcType == dstType {
				out.WriteString(fmt.Sprintf("  %%%s = bitcast %s %%%s to %s\n", instruction.Result, srcType, arg, dstType))
			} else if srcType == "void" {
				if dstType == "ptr" {
					out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 0 to ptr\n", instruction.Result))
				}
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
		} else if isFunctionType(instruction.Type) {
			expectedTag = 7
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
		out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
		argVal = loaded
	}
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tagVar, argVal))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", rawPayload, argVal))
	if isFunctionType(instruction.Type) {
		cmpFn := fmt.Sprintf("cast.cmp_fn.%d", id)
		cmpObj := fmt.Sprintf("cast.cmp_obj.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 7\n", cmpFn, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 5\n", cmpObj, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", cmpVar, cmpFn, cmpObj))
	} else if expectedTag == 5 {
		cmpObj := fmt.Sprintf("cast.cmp_obj.%d", id)
		cmpFn := fmt.Sprintf("cast.cmp_fn.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 5\n", cmpObj, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 7\n", cmpFn, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", cmpVar, cmpObj, cmpFn))
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %d\n", cmpVar, tagVar, expectedTag))
	}
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
		isNotNumber := fmt.Sprintf("cast.not_num.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 3\n", isNotNumber, tagVar))
		// undefined and null keep their number-storage markers; any
		// other non-number becomes NaN.
		isUndefined := fmt.Sprintf("cast.is_undef.%d", id)
		isNull := fmt.Sprintf("cast.is_null.%d", id)
		marker := fmt.Sprintf("cast.marker.%d", id)
		markerOrNaN := fmt.Sprintf("cast.marker_or_nan.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", isUndefined, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 1\n", isNull, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, i64 %s, i64 %s\n", marker, isNull, numberMarkerBits["null"], numberMarkerBits["NaN"]))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, i64 %s, i64 %%%s\n", markerOrNaN, isUndefined, numberMarkerBits["undefined"], marker))
		out.WriteString(fmt.Sprintf("  %%%s.bits = select i1 %%%s, i64 %%%s, i64 %%%s\n", instruction.Result, isNotNumber, markerOrNaN, rawPayload))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s.bits to double\n", instruction.Result, instruction.Result))
	case ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", instruction.Result, rawPayload))
	case ir.TypeBigInt:
		out.WriteString(fmt.Sprintf("  %%%s = add i64 0, %%%s\n", instruction.Result, rawPayload))
	case ir.TypeString:
		stringPayload := fmt.Sprintf("cast.string_ptr.%d", id)
		isUndefined := fmt.Sprintf("cast.string_undefined.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", stringPayload, rawPayload))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", isUndefined, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr @scriptgo_undefined_sentinel, ptr %%%s\n", instruction.Result, isUndefined, stringPayload))
	default:
		pointerPayload := fmt.Sprintf("cast.pointer.%d", id)
		isUndefined := fmt.Sprintf("cast.pointer_undefined.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", pointerPayload, rawPayload))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 0\n", isUndefined, tagVar))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr @scriptgo_undefined_sentinel, ptr %%%s\n", instruction.Result, isUndefined, pointerPayload))
	}
	if slot, ok := e.varSlots[instruction.Result]; ok {
		// A self-cast consumes a boxed slot and leaves the known value in SSA.
		// Do not overwrite that boxed slot with a differently sized value.
		if instruction.Result != instruction.Args[0] {
			out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(instruction.Type), instruction.Result, slot))
		}
	}
	if instruction.Result == instruction.Args[0] {
		// Subsequent uses should read the cast result, not reload the old box.
		delete(e.varSlots, instruction.Result)
	}
	return nil
}

func (e *functionEmitter) emitTypeOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeString
	arg := instruction.Args[0]
	argType, ok := e.types[arg]
	if (instruction.RuntimeTypeOf || strings.Contains(string(argType), "|")) && ok && llvmType(argType) == "ptr" {
		argVal := e.resolveArg(out, arg)
		id := e.loadCounter
		e.loadCounter++
		isUndefined := fmt.Sprintf("typeof.runtime_undefined.%d", id)
		isNull := fmt.Sprintf("typeof.runtime_null.%d", id)
		nonNull := fmt.Sprintf("typeof.runtime_nonnull.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", isUndefined, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", isNull, argVal))
		objectGlobal := e.stringsByValue["object"]
		objectPtr := fmt.Sprintf("typeof.runtime_object.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [7 x i8], ptr %s, i64 0, i64 0\n", objectPtr, objectGlobal))

		var valuePtr string
		hasFn := strings.Contains(string(argType), "=>") || strings.Contains(string(argType), "function") || strings.Contains(string(argType), "Function")
		if argType == ir.TypeString {
			strGlobal := e.stringsByValue["string"]
			sPtr := fmt.Sprintf("typeof.runtime_str.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [7 x i8], ptr %s, i64 0, i64 0\n", sPtr, strGlobal))
			valuePtr = sPtr
		} else if hasFn {
			tagVal := fmt.Sprintf("typeof.gc_tag.%d", e.loadCounter)
			isFn := fmt.Sprintf("typeof.is_fn.%d", e.loadCounter)
			fnPtr := fmt.Sprintf("typeof.runtime_fn.%d", e.loadCounter)
			selPtr := fmt.Sprintf("typeof.runtime_fn_or_obj.%d", e.loadCounter)
			e.loadCounter++
			fnGlobal := e.stringsByValue["function"]
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_gc_get_tag(ptr %%%s)\n", tagVal, argVal))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 3\n", isFn, tagVal))
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [9 x i8], ptr %s, i64 0, i64 0\n", fnPtr, fnGlobal))
			out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", selPtr, isFn, fnPtr, objectPtr))
			valuePtr = selPtr
		} else {
			nonNullType := "object"
			if isFunctionType(argType) {
				nonNullType = "function"
			}
			valueGlobal := e.stringsByValue[nonNullType]
			vPtr := fmt.Sprintf("typeof.runtime_value.%d", e.loadCounter)
			e.loadCounter++
			valueLength := len([]byte(nonNullType)) + 1
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", vPtr, valueLength, valueGlobal))
			valuePtr = vPtr
		}

		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", nonNull, isNull, objectPtr, valuePtr))
		undefGlobal := e.stringsByValue["undefined"]
		undefPtr := fmt.Sprintf("typeof.runtime_undefined_value.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [10 x i8], ptr %s, i64 0, i64 0\n", undefPtr, undefGlobal))
		out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", instruction.Result, isUndefined, undefPtr, nonNull))
		return nil
	}
	if ok && argType != ir.TypeUnknown && !strings.Contains(string(argType), "|") && !e.isParamUnknown(arg) {
		if isFunctionType(argType) {
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
		if argType == ir.TypeNumber {
			return e.emitNumberTypeOf(out, instruction.Result, e.resolveArg(out, arg))
		}
		var typeStr string
		switch {
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
	valuePtr, err := e.emitCanonicalValuePointer(out, arg, ir.TypeUnknown, "typeof.value")
	if err != nil {
		return err
	}
	out.WriteString(fmt.Sprintf("  %%%s = call ptr @__scriptgo_typeof_unknown(ptr %s)\n", instruction.Result, valuePtr))
	return nil
}

// emitNumberTypeOf is typeof for a value in number storage, where undefined
// and null are markers (numberMarkerBits): "undefined", "object" or "number".
func (e *functionEmitter) emitNumberTypeOf(out *strings.Builder, result, value string) error {
	id := e.loadCounter
	e.loadCounter++
	name := func(text string) (string, error) {
		global, ok := e.stringsByValue[text]
		if !ok {
			return "", fmt.Errorf("typeof string %q is not interned", text)
		}
		ptr := fmt.Sprintf("typeof.num.%s.%d", text, id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", ptr, len(text)+1, global))
		return ptr, nil
	}
	numberPtr, err := name("number")
	if err != nil {
		return err
	}
	undefinedPtr, err := name("undefined")
	if err != nil {
		return err
	}
	objectPtr, err := name("object")
	if err != nil {
		return err
	}
	bits := fmt.Sprintf("typeof.num.bits.%d", id)
	isUndefined := fmt.Sprintf("typeof.num.is_undef.%d", id)
	isNull := fmt.Sprintf("typeof.num.is_null.%d", id)
	nullOrNumber := fmt.Sprintf("typeof.num.null_or_number.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = bitcast double %%%s to i64\n", bits, value))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, %s\n", isUndefined, bits, numberMarkerBits["undefined"]))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, %s\n", isNull, bits, numberMarkerBits["null"]))
	out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", nullOrNumber, isNull, objectPtr, numberPtr))
	out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", result, isUndefined, undefinedPtr, nullOrNumber))
	return nil
}
