// Package llvm emits LLVM IR from verified scriptgo IR.
package llvm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// Options controls deterministic LLVM artifact metadata.
type Options struct {
	CompilerVersion string
	RuntimeABI      string
	Target          string
	SourceHash      string
	Debug           bool
}

// Emit converts a verified module into LLVM IR using opaque pointers.
func Emit(module ir.Module) (string, error) {
	return EmitWithOptions(module, Options{
		CompilerVersion: "dev",
		RuntimeABI:      "scriptgo.runtime.v1",
		Target:          "native",
	})
}

// EmitWithOptions converts verified IR into LLVM IR with stable build metadata.
func EmitWithOptions(module ir.Module, options Options) (string, error) {
	if options.CompilerVersion == "" {
		options.CompilerVersion = "dev"
	}
	if options.RuntimeABI == "" {
		options.RuntimeABI = "scriptgo.runtime.v1"
	}
	if options.Target == "" {
		options.Target = "native"
	}
	var debug *debugInfo
	if options.Debug {
		debug = newDebugInfo(module)
	}
	if err := module.Verify(); err != nil {
		return "", err
	}
	functions := make(map[string]ir.Function, len(module.Functions))
	stringsByValue := map[string]string{}
	for _, function := range module.Functions {
		functions[function.Name] = function
		for _, instruction := range function.Body {
			if instruction.Op == ir.OpConst && instruction.Type == ir.TypeString {
				if _, ok := stringsByValue[instruction.Value]; !ok {
					stringsByValue[instruction.Value] = fmt.Sprintf("@.str.%d", len(stringsByValue))
				}
			}
		}
	}
	if _, ok := functions["main"]; !ok {
		return "", fmt.Errorf("module has no main function")
	}

	var out strings.Builder
	out.WriteString("; ModuleID = 'scriptgo'\n")
	fmt.Fprintf(&out, "; scriptgo.compiler = %q\n", options.CompilerVersion)
	fmt.Fprintf(&out, "; scriptgo.runtime-abi = %q\n", options.RuntimeABI)
	fmt.Fprintf(&out, "; scriptgo.target = %q\n", options.Target)
	if options.SourceHash != "" {
		fmt.Fprintf(&out, "; scriptgo.source-sha256 = %q\n", options.SourceHash)
	}
	out.WriteString("declare void @scriptgo_runtime_abort_if_failed(i32)\n\n")
	out.WriteString("declare i32 @scriptgo_print_number(double)\n")
	out.WriteString("declare i32 @scriptgo_print_string(ptr)\n")
	out.WriteString("declare i32 @scriptgo_print_bool(i32)\n\n")
	out.WriteString("declare i32 @scriptgo_array_number_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_number_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_number_set(ptr, double, double)\n")
	out.WriteString("declare i32 @scriptgo_array_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_number_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_array_string_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_string_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_string_set(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_string_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_object_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_number_set(ptr, i64, double)\n")
	out.WriteString("declare i32 @scriptgo_object_number_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_set(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_string_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_last_index(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_release(ptr)\n\n")
	for value, name := range stringsByValue {
		encoded := escapeString(value)
		out.WriteString(fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", name, len([]byte(value))+1, encoded))
	}
	out.WriteString("\n")

	for _, function := range module.Functions {
		text, err := emitFunction(function, functions, stringsByValue, debug)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
	}
	if debug != nil {
		out.WriteString(debug.metadata(module, options.CompilerVersion))
	}
	return out.String(), nil
}

func emitFunction(function ir.Function, functions map[string]ir.Function, stringsByValue map[string]string, debug *debugInfo) (string, error) {
	returnType := llvmType(function.ReturnType)
	name := function.Name
	if name == "main" {
		returnType = "i32"
	}
	var out strings.Builder
	out.WriteString(fmt.Sprintf("define %s @%s(", returnType, name))
	for index, parameter := range function.Parameters {
		if index > 0 {
			out.WriteString(", ")
		}
		out.WriteString(fmt.Sprintf("%s %%%s", llvmType(parameter.Type), parameter.Name))
	}
	out.WriteString(")")
	if debug != nil {
		fmt.Fprintf(&out, " !dbg !%d", debug.functions[function.Name])
	}
	out.WriteString(" {\n")

	types := map[string]ir.Type{}
	type arrayReference struct {
		name string
		typ  ir.Type
	}
	arrayTypes := []arrayReference{}
	objects := []string{}
	ownedStrings := []string{}
	for _, parameter := range function.Parameters {
		types[parameter.Name] = parameter.Type
	}
	terminated := false
	labelCounter := 0
	runtimeStatus := 0
	for _, instruction := range function.Body {
		if terminated {
			return "", fmt.Errorf("function %q contains instruction after return", function.Name)
		}
		switch instruction.Op {
		case ir.OpConst:
			types[instruction.Result] = instruction.Type
			switch instruction.Type {
			case ir.TypeNumber:
				number, err := strconv.ParseFloat(instruction.Value, 64)
				if err != nil {
					return "", fmt.Errorf("invalid number %q: %w", instruction.Value, err)
				}
				out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, %s\n", instruction.Result, llvmNumber(number)))
			case ir.TypeString:
				global := stringsByValue[instruction.Value]
				length := len([]byte(instruction.Value)) + 1
				out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", instruction.Result, length, global))
			case ir.TypeBool:
				out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %s\n", instruction.Result, instruction.Value))
			default:
				return "", fmt.Errorf("unsupported constant type %s", instruction.Type)
			}
		case ir.OpBinary:
			leftType, ok := types[instruction.Args[0]]
			if !ok {
				return "", fmt.Errorf("unknown binary value %q", instruction.Args[0])
			}
			if _, ok := types[instruction.Args[1]]; !ok {
				return "", fmt.Errorf("unknown binary value %q", instruction.Args[1])
			}
			if leftType == ir.TypeString && instruction.Operator == "+" {
				types[instruction.Result] = ir.TypeString
				slot := instruction.Result + ".slot"
				status := instruction.Result + ".status"
				out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
				out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
				ownedStrings = append(ownedStrings, instruction.Result)
				break
			}
			if leftType != ir.TypeNumber {
				return "", fmt.Errorf("LLVM binary operator %q only supports number or string concatenation", instruction.Operator)
			}
			op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]
			if !ok {
				return "", fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
			}
			types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		case ir.OpCompare:
			leftType, ok := types[instruction.Args[0]]
			if !ok || types[instruction.Args[1]] != leftType {
				return "", fmt.Errorf("unknown or mismatched compare operands")
			}
			if leftType != ir.TypeNumber {
				return "", fmt.Errorf("LLVM compare only supports number operands")
			}
			predicate, ok := map[string]string{"==": "oeq", "!==": "une", "<": "olt", "<=": "ole", ">": "ogt", ">=": "oge"}[instruction.Operator]
			if !ok {
				return "", fmt.Errorf("unsupported LLVM compare operator %q", instruction.Operator)
			}
			types[instruction.Result] = ir.TypeBool
			out.WriteString(fmt.Sprintf("  %%%s = fcmp %s double %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
		case ir.OpSelect:
			if types[instruction.Args[0]] != ir.TypeBool || types[instruction.Args[1]] != types[instruction.Args[2]] {
				return "", fmt.Errorf("select operands have incompatible types")
			}
			types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, %s %%%s, %s %%%s\n", instruction.Result, instruction.Args[0], llvmType(instruction.Type), instruction.Args[1], llvmType(instruction.Type), instruction.Args[2]))
		case ir.OpIf:
			if len(instruction.Args) != 1 || types[instruction.Args[0]] != ir.TypeBool || len(instruction.Then) != 1 || instruction.Then[0].Op != ir.OpReturn || len(instruction.Else) != 0 {
				return "", fmt.Errorf("LLVM if currently requires a returning then branch and empty else branch")
			}
			thenLabel := fmt.Sprintf("if.then.%d", labelCounter)
			continueLabel := fmt.Sprintf("if.continue.%d", labelCounter)
			labelCounter++
			out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", instruction.Args[0], thenLabel, continueLabel))
			out.WriteString(fmt.Sprintf("%s:\n", thenLabel))
			branchReturn := instruction.Then[0]
			if len(branchReturn.Args) == 0 {
				out.WriteString("  ret void\n")
			} else {
				out.WriteString(fmt.Sprintf("  ret %s %%%s\n", llvmType(branchReturn.Type), branchReturn.Args[0]))
			}
			out.WriteString(fmt.Sprintf("%s:\n", continueLabel))
		case ir.OpPrint:
			valueType, ok := types[instruction.Args[0]]
			if !ok {
				return "", fmt.Errorf("unknown print value %q", instruction.Args[0])
			}
			switch valueType {
			case ir.TypeNumber:
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_print_number(double %%%s)\n", status, instruction.Args[0]))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			case ir.TypeString:
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_print_string(ptr %%%s)\n", status, instruction.Args[0]))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			case ir.TypeBool:
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				boolValue := fmt.Sprintf("print.bool.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolValue, instruction.Args[0]))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_print_bool(i32 %%%s)\n", status, boolValue))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			default:
				return "", fmt.Errorf("unsupported print type %s", valueType)
			}
		case ir.OpArray:
			if instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray {
				return "", fmt.Errorf("unsupported LLVM array type %s", instruction.Type)
			}
			types[instruction.Result] = instruction.Type
			arrayTypes = append(arrayTypes, arrayReference{name: instruction.Result, typ: instruction.Type})
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			constructor := "scriptgo_array_number_new"
			setter := "scriptgo_array_number_set"
			if instruction.Type == ir.TypeStringArray {
				constructor = "scriptgo_array_string_new"
				setter = "scriptgo_array_string_set"
			}
			status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
			runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i64 %d, ptr %%%s)\n", status, constructor, len(instruction.Args), slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			for index, argument := range instruction.Args {
				if instruction.Type == ir.TypeStringArray {
					status = fmt.Sprintf("runtime.status.%d", runtimeStatus)
					runtimeStatus++
					out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, double %s, ptr %%%s)\n", status, setter, instruction.Result, llvmNumber(float64(index)), argument))
				} else {
					status = fmt.Sprintf("runtime.status.%d", runtimeStatus)
					runtimeStatus++
					out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, double %s, double %%%s)\n", status, setter, instruction.Result, llvmNumber(float64(index)), argument))
				}
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			}
		case ir.OpIndex:
			if len(instruction.Args) != 2 {
				return "", fmt.Errorf("index instruction requires array and index operands")
			}
			arrayType, ok := types[instruction.Args[0]]
			if !ok {
				return "", fmt.Errorf("unknown index array %q", instruction.Args[0])
			}
			types[instruction.Result] = instruction.Type
			slot := instruction.Result + ".slot"
			status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
			runtimeStatus++
			if arrayType == ir.TypeStringArray {
				out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_string_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
				out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_number_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
				out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
			}
		case ir.OpObjectNew:
			if instruction.FieldCount < 0 {
				return "", fmt.Errorf("object shape %q has invalid field count", instruction.Callee)
			}
			types[instruction.Result] = instruction.Type
			objects = append(objects, instruction.Result)
			slot := instruction.Result + ".slot"
			status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
			runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 %d, ptr %%%s)\n", status, instruction.FieldCount, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		case ir.OpFieldSet:
			if instruction.FieldIndex < 0 {
				return "", fmt.Errorf("object field %q has invalid index", instruction.Field)
			}
			valueType, ok := types[instruction.Args[1]]
			if !ok {
				return "", fmt.Errorf("unknown object field value %q", instruction.Args[1])
			}
			switch valueType {
			case ir.TypeNumber:
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			case ir.TypeString:
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			default:
				return "", fmt.Errorf("unsupported object field type %s", valueType)
			}
		case ir.OpFieldGet:
			if instruction.FieldIndex < 0 {
				return "", fmt.Errorf("object field %q has invalid index", instruction.Field)
			}
			types[instruction.Result] = instruction.Type
			slot := instruction.Result + ".slot"
			switch instruction.Type {
			case ir.TypeNumber:
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
				out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
			case ir.TypeString:
				out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
				status := fmt.Sprintf("runtime.status.%d", runtimeStatus)
				runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_string_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
				out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			default:
				return "", fmt.Errorf("unsupported object field type %s", instruction.Type)
			}
		case ir.OpCall:
			if strings.HasPrefix(instruction.Callee, "__array.") {
				arrayType, ok := types[instruction.Args[0]]
				if !ok {
					return "", fmt.Errorf("unknown array intrinsic argument %q", instruction.Args[0])
				}
				if err := emitArrayIntrinsic(&out, instruction, arrayType); err != nil {
					return "", err
				}
				types[instruction.Result] = instruction.Type
				break
			}
			if strings.HasPrefix(instruction.Callee, "__string.") {
				if err := emitStringIntrinsic(&out, instruction); err != nil {
					return "", err
				}
				types[instruction.Result] = instruction.Type
				if instruction.Type == ir.TypeString && (instruction.Callee == "__string.slice" || instruction.Callee == "__string.concat") {
					ownedStrings = append(ownedStrings, instruction.Result)
				}
				break
			}
			callee, ok := functions[instruction.Callee]
			if !ok {
				return "", fmt.Errorf("unknown function %q", instruction.Callee)
			}
			if len(callee.Parameters) != len(instruction.Args) {
				return "", fmt.Errorf("call to %q has wrong arity", instruction.Callee)
			}
			returnType := llvmType(callee.ReturnType)
			if returnType == "void" {
				out.WriteString(fmt.Sprintf("  call void @%s(", instruction.Callee))
			} else {
				types[instruction.Result] = callee.ReturnType
				out.WriteString(fmt.Sprintf("  %%%s = call %s @%s(", instruction.Result, returnType, instruction.Callee))
			}
			for index, argument := range instruction.Args {
				if index > 0 {
					out.WriteString(", ")
				}
				out.WriteString(fmt.Sprintf("%s %%%s", llvmType(callee.Parameters[index].Type), argument))
			}
			out.WriteString(")\n")
		case ir.OpReturn:
			for _, arrayReference := range arrayTypes {
				array := arrayReference.name
				arrayType := arrayReference.typ
				release := "scriptgo_array_number_release"
				if arrayType == ir.TypeStringArray {
					release = "scriptgo_array_string_release"
				}
				out.WriteString(fmt.Sprintf("  call i32 @%s(ptr %%%s)\n", release, array))
			}
			for _, object := range objects {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
			}
			returnValue := ""
			if len(instruction.Args) != 0 {
				returnValue = instruction.Args[0]
			}
			if function.Name == "main" {
				for _, value := range ownedStrings {
					if value != returnValue {
						out.WriteString(fmt.Sprintf("  call i32 @scriptgo_string_release(ptr %%%s)\n", value))
					}
				}
			}
			if function.Name == "main" {
				out.WriteString("  ret i32 0\n")
			} else if len(instruction.Args) == 0 {
				out.WriteString("  ret void\n")
			} else {
				out.WriteString(fmt.Sprintf("  ret %s %%%s\n", llvmType(instruction.Type), instruction.Args[0]))
			}
			terminated = true
		default:
			return "", fmt.Errorf("unsupported LLVM instruction %q", instruction.Op)
		}
	}
	if !terminated {
		for _, arrayReference := range arrayTypes {
			array := arrayReference.name
			arrayType := arrayReference.typ
			release := "scriptgo_array_number_release"
			if arrayType == ir.TypeStringArray {
				release = "scriptgo_array_string_release"
			}
			out.WriteString(fmt.Sprintf("  call i32 @%s(ptr %%%s)\n", release, array))
		}
		for _, object := range objects {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
		}
		if function.Name == "main" {
			for _, value := range ownedStrings {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_string_release(ptr %%%s)\n", value))
			}
		}
		if function.Name == "main" {
			out.WriteString("  ret i32 0\n")
		} else if function.ReturnType == ir.TypeVoid {
			out.WriteString("  ret void\n")
		} else {
			return "", fmt.Errorf("function %q has no return", function.Name)
		}
	}
	out.WriteString("}\n\n")
	return out.String(), nil
}

func llvmType(typ ir.Type) string {
	switch typ {
	case ir.TypeNumber:
		return "double"
	case ir.TypeString:
		return "ptr"
	case ir.TypeNumberArray, ir.TypeStringArray:
		return "ptr"
	case ir.TypeBool:
		return "i1"
	case ir.TypeVoid:
		return "void"
	case ir.TypeObject:
		return "ptr"
	default:
		if strings.HasPrefix(string(typ), string(ir.TypeObject)+":") {
			return "ptr"
		}
		return string(typ)
	}
}

func llvmNumber(value float64) string {
	return strconv.FormatFloat(value, 'e', 17, 64)
}

func escapeString(value string) string {
	var out strings.Builder
	for _, b := range []byte(value) {
		if b >= 32 && b <= 126 && b != '"' && b != '\\' {
			out.WriteByte(b)
			continue
		}
		out.WriteString(fmt.Sprintf("\\%02X", b))
	}
	return out.String()
}
