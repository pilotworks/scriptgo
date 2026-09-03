package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func typedArrayKind(t ir.Type) int {
	switch t {
	case ir.TypeInt8Array:
		return 1
	case ir.TypeUint8Array:
		return 2
	case ir.TypeUint8ClampedArray:
		return 3
	case ir.TypeInt16Array:
		return 4
	case ir.TypeUint16Array:
		return 5
	case ir.TypeInt32Array:
		return 6
	case ir.TypeUint32Array:
		return 7
	case ir.TypeFloat32Array:
		return 8
	case ir.TypeFloat64Array:
		return 9
	case ir.TypeBigInt64Array:
		return 10
	case ir.TypeBigUint64Array:
		return 11
	default:
		return 2
	}
}

func (e *functionEmitter) emitTypedArrayIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__arraybuffer.new":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.new requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[0]+".i64", instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_new(i64 %%%s, ptr %%%s)\n", status, instruction.Args[0]+".i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.byteLength":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.byteLength requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_byte_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.slice":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("arraybuffer.slice requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_slice(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.isView":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.isView requires 1 argument")
		}
		argName := instruction.Args[0]
		arg := ""
		if e.types[argName] == ir.TypeUnknown || e.isParamUnknown(argName) {
			if slot, ok := e.varSlots[argName]; ok {
				loaded := fmt.Sprintf("%s.is_view_load.%d", argName, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load volatile { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
				argName = loaded
			}
			tag := fmt.Sprintf("%s.is_view_tag.%d", argName, e.loadCounter)
			payload := fmt.Sprintf("%s.is_view_payload.%d", argName, e.loadCounter)
			payloadPtr := fmt.Sprintf("%s.is_view_ptr.%d", argName, e.loadCounter)
			objectPtr := fmt.Sprintf("%s.is_view_object.%d", argName, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, argName))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, argName))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", payloadPtr, payload))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, 5\n", objectPtr, tag))
			arg = fmt.Sprintf("%s.is_view_safe.%d", argName, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr %%%s, ptr null\n", arg, objectPtr, payloadPtr))
		} else {
			arg = e.ensurePointerArg(out, argName)
		}
		slot := instruction.Result + ".is_view.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_is_view(ptr %%%s, ptr %%%s)\n", status, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__arraybuffer_view.byteLength", "__arraybuffer_view.byteOffset":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("%s requires 1 argument", strings.TrimPrefix(instruction.Callee, "__"))
		}
		arg := e.ensurePointerArg(out, instruction.Args[0])
		function := "scriptgo_arraybuffer_view_byte_length"
		if instruction.Callee == "__arraybuffer_view.byteOffset" {
			function = "scriptgo_arraybuffer_view_byte_offset"
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, function, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer_view.buffer":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer_view.buffer requires 1 argument")
		}
		arg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_view_buffer(ptr %%%s, ptr %%%s)\n", status, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.new_length":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.new_length requires 1 argument")
		}
		kind := typedArrayKind(instruction.Type)
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[0]+".i64", instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_new(i64 %d, i64 %%%s, ptr null, i64 0, ptr %%%s)\n", status, kind, instruction.Args[0]+".i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.new_buffer":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.new_buffer requires 3 arguments")
		}
		kind := typedArrayKind(instruction.Type)
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[1]+".offset.i64", instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[2]+".len.i64", instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_new(i64 %d, i64 %%%s, ptr %%%s, i64 %%%s, ptr %%%s)\n", status, kind, instruction.Args[2]+".len.i64", instruction.Args[0], instruction.Args[1]+".offset.i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.length":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.length requires 1 argument")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_length(ptr %%%s, ptr %%%s)\n", status, ptrArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.byteLength":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.byteLength requires 1 argument")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_byte_length(ptr %%%s, ptr %%%s)\n", status, ptrArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.byteOffset":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.byteOffset requires 1 argument")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_byte_offset(ptr %%%s, ptr %%%s)\n", status, ptrArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.buffer":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.buffer requires 1 argument")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_buffer(ptr %%%s, ptr %%%s)\n", status, ptrArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.subarray":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.subarray requires 3 arguments")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_subarray(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, ptrArg, instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.slice":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.slice requires 3 arguments")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_slice(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, ptrArg, instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.set":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.set requires 3 arguments")
		}
		ptrArg0 := e.ensurePointerArg(out, instruction.Args[0])
		ptrArg1 := e.ensurePointerArg(out, instruction.Args[1])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_set_array(ptr %%%s, ptr %%%s, double %%%s)\n", status, ptrArg0, ptrArg1, instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__typedarray.set_array":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.set_array requires 3 arguments")
		}
		ptrArg0 := e.ensurePointerArg(out, instruction.Args[0])
		ptrArg1 := e.ensurePointerArg(out, instruction.Args[1])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_set_js_array(ptr %%%s, ptr %%%s, double %%%s)\n", status, ptrArg0, ptrArg1, instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__typedarray.new_array":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("typedarray.new_array requires at least 1 argument")
		}
		kind := typedArrayKind(instruction.Type)
		if instruction.Value != "" {
			kind = typedArrayKind(ir.Type(instruction.Value))
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_from_array(i64 %d, ptr %%%s, ptr %%%s)\n", status, kind, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.new_typed_array":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.new_typed_array requires 1 argument")
		}
		kind := typedArrayKind(instruction.Type)
		if instruction.Value != "" {
			kind = typedArrayKind(ir.Type(instruction.Value))
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_from_typed_array(i64 %d, ptr %%%s, ptr %%%s)\n", status, kind, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.toString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.toString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.fill":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("typedarray.fill requires 4 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		valArg := instruction.Args[1]
		if e.types[valArg] == ir.TypeBigInt {
			fmt.Fprintf(out, "  %%%s = sitofp i64 %%%s to double\n", valArg+".f64", valArg)
			valArg = valArg + ".f64"
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_fill(ptr %%%s, double %%%s, double %%%s, double %%%s)\n", status, instruction.Args[0], valArg, instruction.Args[2], instruction.Args[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	// -------------------------------------------------------------------------
	// DataView intrinsics
	// -------------------------------------------------------------------------
	case "__dataview.new":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("dataview.new requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dataview_new(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.byteLength":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("dataview.byteLength requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dataview_byte_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.byteOffset":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("dataview.byteOffset requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dataview_byte_offset(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.buffer":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("dataview.buffer requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dataview_buffer(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.getInt8", "__dataview.getUint8":
		cFuncName := "scriptgo_dataview_get_int8"
		if instruction.Callee == "__dataview.getUint8" {
			cFuncName = "scriptgo_dataview_get_uint8"
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, ptr %%%s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.setInt8", "__dataview.setUint8":
		cFuncName := "scriptgo_dataview_set_int8"
		if instruction.Callee == "__dataview.setUint8" {
			cFuncName = "scriptgo_dataview_set_uint8"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, double %%%s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dataview.getInt16", "__dataview.getUint16", "__dataview.getInt32", "__dataview.getUint32", "__dataview.getFloat32", "__dataview.getFloat64":
		method := strings.TrimPrefix(instruction.Callee, "__dataview.get")
		cFuncName := fmt.Sprintf("scriptgo_dataview_get_%s", strings.ToLower(method))
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		leArg := "0"
		if len(instruction.Args) > 2 {
			leArg = instruction.Args[2]
			if e.types[leArg] == ir.TypeBool {
				leVar := fmt.Sprintf("%s.le_i32.%d", instruction.Result, e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", leVar, leArg)
				leArg = "%" + leVar
			} else {
				leArg = "%" + leArg
			}
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, i32 %s, ptr %%%s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], leArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.setUint16", "__dataview.setInt16", "__dataview.setUint32", "__dataview.setInt32", "__dataview.setFloat32", "__dataview.setFloat64":
		method := strings.TrimPrefix(instruction.Callee, "__dataview.set")
		cFuncName := fmt.Sprintf("scriptgo_dataview_set_%s", strings.ToLower(method))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		leArg := "0"
		if len(instruction.Args) > 3 {
			leArg = instruction.Args[3]
			if e.types[leArg] == ir.TypeBool {
				leVar := fmt.Sprintf("%s.le_i32.%d", instruction.Result, e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", leVar, leArg)
				leArg = "%" + leVar
			} else {
				leArg = "%" + leArg
			}
		}
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, double %%%s, i32 %s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], instruction.Args[2], leArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dataview.getBigInt64", "__dataview.getBigUint64":
		cFuncName := "scriptgo_dataview_get_bigint64"
		if instruction.Callee == "__dataview.getBigUint64" {
			cFuncName = "scriptgo_dataview_get_biguint64"
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		leArg := "0"
		if len(instruction.Args) > 2 {
			leArg = instruction.Args[2]
			if e.types[leArg] == ir.TypeBool {
				leVar := fmt.Sprintf("%s.le_i32.%d", instruction.Result, e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", leVar, leArg)
				leArg = "%" + leVar
			} else {
				leArg = "%" + leArg
			}
		}
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, i32 %s, ptr %%%s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], leArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dataview.setBigInt64", "__dataview.setBigUint64":
		cFuncName := "scriptgo_dataview_set_bigint64"
		if instruction.Callee == "__dataview.setBigUint64" {
			cFuncName = "scriptgo_dataview_set_biguint64"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		leArg := "0"
		if len(instruction.Args) > 3 {
			leArg = instruction.Args[3]
			if e.types[leArg] == ir.TypeBool {
				leVar := fmt.Sprintf("%s.le_i32.%d", instruction.Result, e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", leVar, leArg)
				leArg = "%" + leVar
			} else {
				leArg = "%" + leArg
			}
		}
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, i64 %%%s, i32 %s)\n", status, cFuncName, instruction.Args[0], instruction.Args[1], instruction.Args[2], leArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dataview.toString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("dataview.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dataview_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown typedarray intrinsic %q", instruction.Callee)
	}
}
