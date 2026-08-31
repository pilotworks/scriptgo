package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func weakValueTag(valueType ir.Type) int {
	switch valueType {
	case ir.TypeVoid:
		return 0 // SCRIPTGO_TAG_UNDEFINED
	case ir.TypePointer:
		return 1 // SCRIPTGO_TAG_NULL
	case ir.TypeBool:
		return 2 // SCRIPTGO_TAG_BOOLEAN
	case ir.TypeNumber:
		return 3 // SCRIPTGO_TAG_NUMBER
	case ir.TypeString:
		return 4 // SCRIPTGO_TAG_STRING
	case ir.TypeBigInt:
		return 8 // SCRIPTGO_TAG_BIGINT
	case ir.TypeSymbol:
		return 9 // SCRIPTGO_TAG_SYMBOL
	default:
		if strings.HasSuffix(string(valueType), "[]") {
			return 6 // SCRIPTGO_TAG_ARRAY
		}
		if valueType == ir.TypeClosure {
			return 7 // SCRIPTGO_TAG_FUNCTION
		}
		return 5 // SCRIPTGO_TAG_OBJECT
	}
}

func emitGcIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"

	switch instruction.Callee {
	case "__gc.collect":
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_gc_collect(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i64Val := instruction.Result + ".i64"
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", i64Val, slot)
		fmt.Fprintf(out, "  %%%s = sitofp i64 %%%s to double\n", instruction.Result, i64Val)
	default:
		return fmt.Errorf("unknown gc intrinsic %q", instruction.Callee)
	}
	return nil
}

func (e *functionEmitter) emitWeakIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	slot := fmt.Sprintf("weak.slot.%d", e.loadCounter)
	e.loadCounter++

	switch instruction.Callee {
	case "__weakref.new":
		targetArg := "null"
		if len(instruction.Args) > 0 && instruction.Args[0] != "" {
			targetArg = "%" + instruction.Args[0]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakref_new(ptr %s, ptr %%%s)\n", status, targetArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__weakref.deref":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakref_deref(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__weakmap.new":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakmap_new(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__weakmap.set":
		valArg := instruction.Args[2]
		valType := e.types[valArg]
		valArg = e.resolveArg(out, valArg)
		var ptrVal string
		tagVal := fmt.Sprintf("%d", weakValueTag(valType))
		if valType == ir.TypeNumber {
			i64Val := fmt.Sprintf("%s.i64.%d", valArg, e.loadCounter)
			ptrVal = fmt.Sprintf("%s.ptr.%d", valArg, e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = bitcast double %%%s to i64\n", i64Val, valArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, i64Val)
		} else if valType == ir.TypeBool {
			i64Val := fmt.Sprintf("%s.i64.%d", valArg, e.loadCounter)
			ptrVal = fmt.Sprintf("%s.ptr.%d", valArg, e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i64\n", i64Val, valArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, i64Val)
		} else if valType == ir.TypeBigInt {
			ptrVal = fmt.Sprintf("%s.ptr.%d", valArg, e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, valArg)
		} else if valType == ir.TypeUnknown {
			rawVal := fmt.Sprintf("%s.raw.%d", valArg, e.loadCounter)
			ptrVal = fmt.Sprintf("%s.ptr.%d", valArg, e.loadCounter)
			tagVal = fmt.Sprintf("%s.tag.%d", valArg, e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", rawVal, valArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, rawVal)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVal, valArg)
		} else if valType == ir.TypeVoid {
			ptrVal = "null"
		} else {
			ptrVal = valArg
		}
		ptrOperand := "null"
		if ptrVal != "null" {
			ptrOperand = "%" + ptrVal
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakmap_set(ptr %%%s, ptr %%%s, ptr %s, i32 %s)\n", status, instruction.Args[0], instruction.Args[1], ptrOperand, tagVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

	case "__weakmap.get":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		tagSlot := slot + ".tag"
		fmt.Fprintf(out, "  %%%s = alloca i32\n", tagSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakmap_get(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot, tagSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		resPtr := fmt.Sprintf("%s.rawptr", instruction.Result)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", resPtr, slot)
		resTag := fmt.Sprintf("%s.tag", instruction.Result)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", resTag, tagSlot)
		if instruction.Type == ir.TypeNumber {
			i64Val := fmt.Sprintf("%s.i64", instruction.Result)
			fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", i64Val, resPtr)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", instruction.Result, i64Val)
		} else if instruction.Type == ir.TypeBool {
			i64Val := fmt.Sprintf("%s.i64", instruction.Result)
			fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", i64Val, resPtr)
			fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", instruction.Result, i64Val)
		} else if instruction.Type == ir.TypeBigInt {
			fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", instruction.Result, resPtr)
		} else if instruction.Type == ir.TypeUnknown {
			payload := fmt.Sprintf("%s.payload", instruction.Result)
			box0 := fmt.Sprintf("%s.box0", instruction.Result)
			box1 := fmt.Sprintf("%s.box1", instruction.Result)
			fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", payload, resPtr)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", box0, resTag)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", box1, box0)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, box1, payload)
		} else {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, resPtr)
		}

	case "__weakmap.has":
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakmap_has(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	case "__weakmap.delete":
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakmap_delete(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	case "__weakset.new":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakset_new(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__weakset.add":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakset_add(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

	case "__weakset.has":
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakset_has(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	case "__weakset.delete":
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_weakset_delete(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	case "__finalization_registry.new":
		cbArg := "null"
		if len(instruction.Args) > 0 && instruction.Args[0] != "" {
			cbArg = "%" + e.resolveArg(out, instruction.Args[0])
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_finalization_registry_new(ptr %s, ptr %%%s)\n", status, cbArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__finalization_registry.register":
		handle := e.resolveArg(out, instruction.Args[0])
		target := e.resolveArg(out, instruction.Args[1])
		held := e.resolveArg(out, instruction.Args[2])
		token := "null"
		if len(instruction.Args) > 3 {
			token = "%" + e.resolveArg(out, instruction.Args[3])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_finalization_registry_register(ptr %%%s, ptr %%%s, ptr %%%s, ptr %s)\n", status, handle, target, held, token)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

	case "__finalization_registry.unregister":
		handle := e.resolveArg(out, instruction.Args[0])
		token := e.resolveArg(out, instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_finalization_registry_unregister(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, handle, token, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	default:
		return fmt.Errorf("unknown weak intrinsic %q", instruction.Callee)
	}
	return nil
}
