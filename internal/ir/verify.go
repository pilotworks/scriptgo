package ir

import (
	"fmt"
	"maps"
	"strings"
)

// Verify checks the invariants required by every native backend.
func (m Module) Verify() error {
	if len(m.Functions) == 0 {
		return fmt.Errorf("module has no functions")
	}
	globals := make(map[string]Type, len(m.Globals))
	for _, g := range m.Globals {
		globals[g.Name] = g.Type
	}
	for _, function := range m.Functions {
		if err := function.verifyInternal(globals); err != nil {
			return fmt.Errorf("function %q: %w", function.Name, err)
		}
	}
	for _, shape := range m.Shapes {
		if shape.Name == "" {
			return fmt.Errorf("invalid object shape %q", shape.Name)
		}
		for _, field := range shape.Fields {
			if field.Name == "" || field.Type == "" {
				return fmt.Errorf("invalid field in object shape %q", shape.Name)
			}
		}
	}
	return nil
}

func (f Function) Verify() error {
	return f.verifyInternal(nil)
}

func (f Function) verifyInternal(globals map[string]Type) error {
	if f.Name == "" {
		return fmt.Errorf("function has no name")
	}
	known := make(map[string]Type, len(globals)+len(f.Parameters)+len(f.Captured))
	if globals != nil {
		maps.Copy(known, globals)
	}
	for _, parameter := range f.Parameters {
		if parameter.Name == "" || parameter.Type == "" {
			return fmt.Errorf("invalid parameter")
		}
		known[parameter.Name] = parameter.Type
	}
	for _, captured := range f.Captured {
		if captured.Name == "" || captured.Type == "" {
			return fmt.Errorf("invalid captured variable")
		}
		known[captured.Name] = captured.Type
	}
	terminated := false
	for _, instruction := range f.Body {
		if terminated {
			return fmt.Errorf("instruction follows return")
		}
		for _, arg := range instruction.Args {
			if _, ok := known[arg]; !ok && instruction.Op != OpCall && instruction.Op != OpExternCall && instruction.Op != OpPrint && instruction.Op != OpWhile && instruction.Op != OpDoWhile {
				return fmt.Errorf("unknown value %q", arg)
			}
		}
		switch instruction.Op {
		case OpCall, OpClosureCall, OpExternCall:
			if instruction.Type == "" {
				return fmt.Errorf("%s instruction must define type", instruction.Op)
			}
			if instruction.Type != TypeVoid && instruction.Result == "" {
				return fmt.Errorf("%s instruction with return type %s must define result", instruction.Op, instruction.Type)
			}
			if instruction.Result != "" {
				known[instruction.Result] = instruction.Type
			}
		case OpClosure:
			if instruction.Result == "" || instruction.Type != TypeClosure {
				return fmt.Errorf("closure instruction must define result and closure type")
			}
			known[instruction.Result] = TypeClosure
		case OpBoxUnknown:
			if instruction.Result == "" || instruction.Type != TypeUnknown || len(instruction.Args) != 1 {
				return fmt.Errorf("box.unknown instruction must define result, unknown type, and one arg")
			}
			known[instruction.Result] = TypeUnknown
		case OpCheckedCast:
			if instruction.Result == "" || instruction.Type == "" || len(instruction.Args) != 1 {
				return fmt.Errorf("checked_cast instruction must define result, target type, and one arg")
			}
			srcType, ok := known[instruction.Args[0]]
			if !ok || (srcType != TypeUnknown && srcType != instruction.Type && !strings.Contains(string(srcType), "|") && !strings.HasPrefix(string(srcType), "object:") && !strings.HasPrefix(string(instruction.Type), "object:") && !isTypedArrayType(srcType) && srcType != TypeArrayBuffer) {
				return fmt.Errorf("checked_cast source must be unknown, union, or object type, got %v", srcType)
			}
			known[instruction.Result] = instruction.Type
		case OpTypeOf:
			if instruction.Result == "" || instruction.Type != TypeString || len(instruction.Args) != 1 {
				return fmt.Errorf("typeof instruction must define result, string type, and one arg")
			}
			known[instruction.Result] = TypeString
		case OpConst, OpBinary, OpCompare, OpSelect, OpParam, OpArray, OpIndex, OpObjectNew, OpFieldGet, OpInstanceOf:
			if instruction.Result == "" || instruction.Type == "" {
				return fmt.Errorf("%s instruction must define result and type", instruction.Op)
			}
			if instruction.Op == OpInstanceOf {
				if len(instruction.Args) != 1 || instruction.Type != TypeBool || instruction.Value == "" {
					return fmt.Errorf("instanceof requires one object operand, target class value, and bool result")
				}
			}
			if instruction.Op == OpArray {
				if !strings.HasSuffix(string(instruction.Type), "[]") && instruction.Type != TypeNumberArray && instruction.Type != TypeStringArray && instruction.Type != TypeBigIntArray && instruction.Type != TypeSymbolArray {
					return fmt.Errorf("array instruction has unsupported type %q", instruction.Type)
				}
				for _, argument := range instruction.Args {
					elemType := elementType(instruction.Type)
					argType := known[argument]
					isObjArg := argType == TypeObject || strings.HasPrefix(string(argType), "object:")
					isObjElem := elemType == TypeObject || strings.HasPrefix(string(elemType), "object:")
					if argType != elemType && !(isObjArg && isObjElem) {
						return fmt.Errorf("array element %q has type %s, want %s", argument, argType, elemType)
					}
				}
			}
			if instruction.Op == OpBinary {
				if len(instruction.Args) != 2 {
					return fmt.Errorf("binary instruction requires two operands")
				}
				leftType := known[instruction.Args[0]]
				rightType := known[instruction.Args[1]]
				if leftType != rightType {
					return fmt.Errorf("binary operands must have the same type")
				}
				if leftType == TypeBool {
					if instruction.Operator != "&&" && instruction.Operator != "||" {
						return fmt.Errorf("binary bool instruction has unsupported operator %q", instruction.Operator)
					}
					if instruction.Type != TypeBool {
						return fmt.Errorf("binary bool instruction must have bool result type")
					}
				}
			}
			if instruction.Op == OpCompare {
				if len(instruction.Args) != 2 || instruction.Type != TypeBool {
					return fmt.Errorf("compare instruction requires two operands and bool result")
				}
				if instruction.Operator != "==" && instruction.Operator != "===" && instruction.Operator != "!=" && instruction.Operator != "!==" && instruction.Operator != "<" && instruction.Operator != "<=" && instruction.Operator != ">" && instruction.Operator != ">=" {
					return fmt.Errorf("compare instruction has unsupported operator %q", instruction.Operator)
				}
				leftType := known[instruction.Args[0]]
				rightType := known[instruction.Args[1]]
				if leftType != rightType && !(isPointerType(leftType) && isPointerType(rightType)) {
					return fmt.Errorf("compare operands must have the same type (left=%s, right=%s)", leftType, rightType)
				}
			}
			if instruction.Op == OpSelect {
				if len(instruction.Args) != 3 || known[instruction.Args[0]] != TypeBool {
					return fmt.Errorf("select requires bool condition")
				}
				val1Type := known[instruction.Args[1]]
				val2Type := known[instruction.Args[2]]
				if !isAssignableTo(val1Type, instruction.Type) || !isAssignableTo(val2Type, instruction.Type) {
					return fmt.Errorf("select requires matching values (known[0]=%s, known[1]=%s, known[2]=%s, type=%s)", known[instruction.Args[0]], val1Type, val2Type, instruction.Type)
				}
			}
			if instruction.Op == OpIndex {
				if len(instruction.Args) != 2 {
					return fmt.Errorf("index instruction requires array and index operands")
				}
				arrType, ok := known[instruction.Args[0]]
				if !ok || (!strings.HasSuffix(string(arrType), "[]") && arrType != TypeString && !isTypedArrayType(arrType)) {
					return fmt.Errorf("index instruction requires an array or string operand, got %v", arrType)
				}
				if known[instruction.Args[1]] != TypeNumber {
					return fmt.Errorf("index instruction requires number index operand")
				}
				if arrType == TypeString {
					if instruction.Type != TypeString {
						return fmt.Errorf("index instruction on string must produce string")
					}
				} else if isBigIntTypedArrayType(arrType) {
					if instruction.Type != TypeBigInt {
						return fmt.Errorf("index instruction on bigint typed array must produce bigint")
					}
				} else if isTypedArrayType(arrType) {
					if instruction.Type != TypeNumber {
						return fmt.Errorf("index instruction on typed array must produce number")
					}
				} else if instruction.Type != elementType(arrType) && !isAssignableTo(instruction.Type, elementType(arrType)) {
					return fmt.Errorf("index instruction has incompatible result type %s (expected %s for array %s, result=%s, args=%v)", instruction.Type, elementType(arrType), arrType, instruction.Result, instruction.Args)
				}
			}
			if instruction.Op == OpObjectNew && !strings.HasPrefix(string(instruction.Type), string(TypeObject)+":") {
				return fmt.Errorf("object.new must produce a shaped object")
			}
			if instruction.Op == OpObjectNew && instruction.FieldCount < 0 {
				return fmt.Errorf("object.new has invalid field count")
			}
			if instruction.Op == OpFieldGet && len(instruction.Args) != 1 {
				return fmt.Errorf("field.get requires one object operand")
			}
			if instruction.Op == OpFieldGet && instruction.FieldIndex < 0 {
				return fmt.Errorf("field.get has invalid field index")
			}
			known[instruction.Result] = instruction.Type
		case OpPrint:
			if len(instruction.Args) != 1 {
				return fmt.Errorf("print instruction requires one argument")
			}
		case OpFieldSet:
			if instruction.Type != TypeVoid || len(instruction.Args) != 2 || instruction.Field == "" || instruction.FieldIndex < 0 {
				return fmt.Errorf("field.set requires object, value, and field")
			}
		case OpIndexSet:
			if instruction.Type != TypeVoid || len(instruction.Args) != 3 {
				return fmt.Errorf("index.set requires array, index, and value operands")
			}
			arrType, ok := known[instruction.Args[0]]
			if !ok || (!strings.HasSuffix(string(arrType), "[]") && !isTypedArrayType(arrType)) {
				return fmt.Errorf("index.set requires array operand")
			}
			if known[instruction.Args[1]] != TypeNumber {
				return fmt.Errorf("index.set requires number index")
			}
			if isBigIntTypedArrayType(arrType) {
				if known[instruction.Args[2]] != TypeBigInt {
					return fmt.Errorf("index.set value type mismatch")
				}
			} else if isTypedArrayType(arrType) {
				if known[instruction.Args[2]] != TypeNumber {
					return fmt.Errorf("index.set value type mismatch")
				}
			} else if known[instruction.Args[2]] != elementType(arrType) {
				return fmt.Errorf("index.set value type mismatch")
			}
		case OpAssign:
			if len(instruction.Args) != 1 || instruction.Result == "" {
				return fmt.Errorf("assign instruction requires variable result and one argument")
			}
			varType, ok := known[instruction.Result]
			if !ok {
				return fmt.Errorf("assign to unknown variable %q", instruction.Result)
			}
			valType, ok := known[instruction.Args[0]]
			isObjPair := (varType == TypeObject || strings.HasPrefix(string(varType), "object:")) && (valType == TypeObject || strings.HasPrefix(string(valType), "object:"))
			if !ok || (varType != valType && !isObjPair) {
				return fmt.Errorf("assign type mismatch: %s := %s", varType, valType)
			}
		case OpIf:
			if instruction.Type != TypeVoid || instruction.Result != "" || len(instruction.Args) != 1 || known[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("if requires one bool condition")
			}
			thenKnown := cloneTypes(known)
			elseKnown := cloneTypes(known)
			if err := verifyBlock(f, instruction.Then, thenKnown); err != nil {
				return fmt.Errorf("if then block: %w", err)
			}
			if err := verifyBlock(f, instruction.Else, elseKnown); err != nil {
				return fmt.Errorf("if else block: %w", err)
			}
		case OpWhile, OpDoWhile:
			if instruction.Type != TypeVoid || instruction.Result != "" {
				return fmt.Errorf("%s instruction must have void type and empty result", instruction.Op)
			}
			condKnown := cloneTypes(known)
			if err := verifyBlock(f, instruction.Cond, condKnown); err != nil {
				return fmt.Errorf("%s cond block: %w", instruction.Op, err)
			}
			if len(instruction.Args) != 1 || condKnown[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("%s requires one bool condition", instruction.Op)
			}
			bodyKnown := cloneTypes(condKnown)
			if err := verifyBlock(f, instruction.Body, bodyKnown); err != nil {
				return fmt.Errorf("%s body block: %w", instruction.Op, err)
			}
			if len(instruction.Step) > 0 {
				stepKnown := cloneTypes(bodyKnown)
				if err := verifyBlock(f, instruction.Step, stepKnown); err != nil {
					return fmt.Errorf("%s step block: %w", instruction.Op, err)
				}
			}
		case OpBreak, OpContinue, OpDebugger:
			if instruction.Type != TypeVoid || instruction.Result != "" || len(instruction.Args) != 0 {
				return fmt.Errorf("%s instruction must have void type and no args", instruction.Op)
			}
		case OpThrow:
			if len(instruction.Args) != 1 {
				return fmt.Errorf("throw instruction requires one argument")
			}
			terminated = true
		case OpTry:
			if instruction.Type != TypeVoid || instruction.Result != "" {
				return fmt.Errorf("try instruction must have void type and empty result")
			}
			tryKnown := cloneTypes(known)
			if err := verifyBlock(f, instruction.Body, tryKnown); err != nil {
				return fmt.Errorf("try body block: %w", err)
			}
			if len(instruction.Catch) > 0 {
				catchKnown := cloneTypes(known)
				if instruction.CatchVar != "" {
					catchKnown[instruction.CatchVar] = TypeString
				}
				if err := verifyBlock(f, instruction.Catch, catchKnown); err != nil {
					return fmt.Errorf("try catch block: %w", err)
				}
			}
			if len(instruction.Finally) > 0 {
				finallyKnown := cloneTypes(known)
				if err := verifyBlock(f, instruction.Finally, finallyKnown); err != nil {
					return fmt.Errorf("try finally block: %w", err)
				}
			}
		case OpReturn:
			if instruction.Type != f.ReturnType {
				return fmt.Errorf("return type %q does not match function return type %q", instruction.Type, f.ReturnType)
			}
			terminated = true
		default:
			return fmt.Errorf("unknown instruction %q", instruction.Op)
		}
	}
	return nil
}

func elementType(arrayType Type) Type {
	str := string(arrayType)
	if before, ok := strings.CutSuffix(str, "[]"); ok {
		elem := before
		if elem == "boolean" {
			return TypeBool
		}
		return Type(elem)
	}
	if arrayType == TypeStringArray {
		return TypeString
	}
	if arrayType == TypeBoolArray {
		return TypeBool
	}
	if arrayType == TypeUnknownArray || arrayType == TypeUnknown {
		return TypeUnknown
	}
	return TypeNumber
}

func cloneTypes(types map[string]Type) map[string]Type {
	clone := make(map[string]Type, len(types))
	maps.Copy(clone, types)
	return clone
}

func verifyBlock(f Function, body []Instruction, known map[string]Type) error {
	for _, instruction := range body {
		if instruction.Op == OpIf {
			if instruction.Type != TypeVoid || instruction.Result != "" || len(instruction.Args) != 1 || known[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("nested if requires one bool condition")
			}
			if err := verifyBlock(f, instruction.Then, cloneTypes(known)); err != nil {
				return err
			}
			if err := verifyBlock(f, instruction.Else, cloneTypes(known)); err != nil {
				return err
			}
			continue
		}
		if instruction.Op == OpWhile || instruction.Op == OpDoWhile {
			condKnown := cloneTypes(known)
			if err := verifyBlock(f, instruction.Cond, condKnown); err != nil {
				return err
			}
			if len(instruction.Args) != 1 || condKnown[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("nested %s requires one bool condition", instruction.Op)
			}
			if err := verifyBlock(f, instruction.Body, cloneTypes(condKnown)); err != nil {
				return err
			}
			if len(instruction.Step) > 0 {
				if err := verifyBlock(f, instruction.Step, cloneTypes(condKnown)); err != nil {
					return err
				}
			}
			continue
		}
		if instruction.Op == OpThrow {
			if len(instruction.Args) != 1 {
				return fmt.Errorf("nested throw requires one argument")
			}
			continue
		}
		if instruction.Op == OpTry {
			if err := verifyBlock(f, instruction.Body, cloneTypes(known)); err != nil {
				return err
			}
			if len(instruction.Catch) > 0 {
				catchKnown := cloneTypes(known)
				if instruction.CatchVar != "" {
					catchKnown[instruction.CatchVar] = TypeString
				}
				if err := verifyBlock(f, instruction.Catch, catchKnown); err != nil {
					return err
				}
			}
			if len(instruction.Finally) > 0 {
				if err := verifyBlock(f, instruction.Finally, cloneTypes(known)); err != nil {
					return err
				}
			}
			continue
		}
		if instruction.Op == OpBreak || instruction.Op == OpContinue {
			continue
		}
		if instruction.Op == OpReturn {
			if instruction.Type != f.ReturnType {
				return fmt.Errorf("return type %q does not match function return type %q", instruction.Type, f.ReturnType)
			}
			continue
		}
		if instruction.Result != "" && instruction.Type != "" {
			known[instruction.Result] = instruction.Type
		}
	}
	return nil
}

func isTypedArrayType(t Type) bool {
	switch t {
	case TypeBuffer, TypeUint8Array, TypeInt8Array, TypeUint8ClampedArray,
		TypeInt16Array, TypeUint16Array, TypeInt32Array, TypeUint32Array,
		TypeFloat32Array, TypeFloat64Array, TypeBigInt64Array, TypeBigUint64Array:
		return true
	default:
		return false
	}
}

func isPointerType(t Type) bool {
	return t == TypePointer || t == "ptr" || t == TypeObject || strings.HasPrefix(string(t), "object:") || strings.HasSuffix(string(t), "[]") || t == TypeClosure || t == TypeArrayBuffer || t == TypeBuffer || t == TypeDataView || t == TypeTextEncoder || t == TypeTextDecoder || t == TypeMap || t == TypeSet || t == TypeSymbol || isTypedArrayType(t)
}

func isBigIntTypedArrayType(t Type) bool {
	return t == TypeBigInt64Array || t == TypeBigUint64Array
}

func isAssignableTo(actual, expected Type) bool {
	if actual == expected {
		return true
	}
	if actual == TypeUnknown || expected == TypeUnknown {
		return true
	}
	if (actual == "never[]" || actual == "object:never[]" || actual == "any[]" || actual == "object:any[]") && strings.HasSuffix(string(expected), "[]") {
		return true
	}
	if strings.HasPrefix(string(actual), "object:") && strings.HasPrefix(string(expected), "object:") {
		return true
	}
	if actual == TypePointer && (strings.HasPrefix(string(expected), "object:") || expected == TypeClosure) {
		return true
	}
	if expected == TypePointer && (strings.HasPrefix(string(actual), "object:") || actual == TypeClosure) {
		return true
	}
	return false
}
