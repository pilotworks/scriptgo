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
	out.WriteString("declare i32 @printf(ptr, ...)\n")
	out.WriteString("declare i32 @puts(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_array_number_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_number_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_number_set(ptr, double, double)\n")
	out.WriteString("declare i32 @scriptgo_array_number_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_object_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_number_set(ptr, i64, double)\n")
	out.WriteString("declare i32 @scriptgo_object_number_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_set(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_release(ptr)\n\n")
	out.WriteString("@.fmt.num = private unnamed_addr constant [4 x i8] c\"%g\\0A\\00\"\n")
	for value, name := range stringsByValue {
		encoded := escapeString(value)
		out.WriteString(fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", name, len([]byte(value))+1, encoded))
	}
	out.WriteString("@.true = private unnamed_addr constant [5 x i8] c\"true\\00\"\n")
	out.WriteString("@.false = private unnamed_addr constant [6 x i8] c\"false\\00\"\n\n")

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
	arrays := []string{}
	objects := []string{}
	for _, parameter := range function.Parameters {
		types[parameter.Name] = parameter.Type
	}
	terminated := false
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
			if leftType != ir.TypeNumber {
				return "", fmt.Errorf("LLVM binary operator %q only supports number", instruction.Operator)
			}
			op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]
			if !ok {
				return "", fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
			}
			types[instruction.Result] = instruction.Type
			out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		case ir.OpPrint:
			valueType, ok := types[instruction.Args[0]]
			if !ok {
				return "", fmt.Errorf("unknown print value %q", instruction.Args[0])
			}
			switch valueType {
			case ir.TypeNumber:
				out.WriteString(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr @.fmt.num, double %%%s)\n", instruction.Args[0]))
			case ir.TypeString:
				out.WriteString(fmt.Sprintf("  call i32 @puts(ptr %%%s)\n", instruction.Args[0]))
			case ir.TypeBool:
				out.WriteString(fmt.Sprintf("  %%print.bool = select i1 %%%s, ptr @.true, ptr @.false\n", instruction.Args[0]))
				out.WriteString("  call i32 @puts(ptr %print.bool)\n")
			default:
				return "", fmt.Errorf("unsupported print type %s", valueType)
			}
		case ir.OpArray:
			if instruction.Type != ir.TypeNumberArray {
				return "", fmt.Errorf("unsupported LLVM array type %s", instruction.Type)
			}
			types[instruction.Result] = instruction.Type
			arrays = append(arrays, instruction.Result)
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_number_new(i64 %d, ptr %%%s)\n", len(instruction.Args), slot))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			for index, argument := range instruction.Args {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_number_set(ptr %%%s, double %s, double %%%s)\n", instruction.Result, llvmNumber(float64(index)), argument))
			}
		case ir.OpIndex:
			if len(instruction.Args) != 2 {
				return "", fmt.Errorf("index instruction requires array and index operands")
			}
			types[instruction.Result] = instruction.Type
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_number_get(ptr %%%s, double %%%s, ptr %%%s)\n", instruction.Args[0], instruction.Args[1], slot))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		case ir.OpObjectNew:
			if instruction.FieldCount < 0 {
				return "", fmt.Errorf("object shape %q has invalid field count", instruction.Callee)
			}
			types[instruction.Result] = instruction.Type
			objects = append(objects, instruction.Result)
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_new(i64 %d, ptr %%%s)\n", instruction.FieldCount, slot))
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
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
			case ir.TypeString:
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_string_set(ptr %%%s, i64 %d, ptr %%%s)\n", instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
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
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_number_get(ptr %%%s, i64 %d, ptr %%%s)\n", instruction.Args[0], instruction.FieldIndex, slot))
				out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
			case ir.TypeString:
				out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_string_get(ptr %%%s, i64 %d, ptr %%%s)\n", instruction.Args[0], instruction.FieldIndex, slot))
				out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			default:
				return "", fmt.Errorf("unsupported object field type %s", instruction.Type)
			}
		case ir.OpCall:
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
			for _, array := range arrays {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_number_release(ptr %%%s)\n", array))
			}
			for _, object := range objects {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
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
		for _, array := range arrays {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_number_release(ptr %%%s)\n", array))
		}
		for _, object := range objects {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
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
