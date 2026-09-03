package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitNetIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__net.socketCreate":
		// Args: [family: number, sock_type: number] -> fd: number
		famArg := "4.0"
		typeArg := "1.0"
		if len(instruction.Args) > 0 {
			famArg = "%" + resolvedArgs[0]
		}
		if len(instruction.Args) > 1 {
			typeArg = "%" + resolvedArgs[1]
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_socket_create(double %s, double %s, ptr %%%s)\n",
			status, famArg, typeArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__net.socketConnect":
		// Args: [fd: number, host: string, port: number] -> void
		if len(instruction.Args) < 3 {
			return fmt.Errorf("net.socketConnect requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_socket_connect(double %%%s, ptr %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__net.socketWrite":
		// Args: [fd: number, data: string, len: number] -> written: number
		if len(instruction.Args) < 3 {
			return fmt.Errorf("net.socketWrite requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_socket_write(double %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__net.socketRead":
		// Args: [fd: number, max_len: number] -> data: string
		if len(instruction.Args) < 2 {
			return fmt.Errorf("net.socketRead requires 2 arguments")
		}
		dataSlot := instruction.Result + ".data.slot"
		readSlot := instruction.Result + ".read.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dataSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", readSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_socket_read(double %%%s, double %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], dataSlot, readSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, dataSlot)
		return nil

	case "__net.socketClose":
		// Args: [fd: number] -> void
		if len(instruction.Args) < 1 {
			return fmt.Errorf("net.socketClose requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_socket_close(double %%%s)\n",
			status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__net.serverListen":
		// Args: [host: string, port: number, backlog: number] -> server_fd: number
		if len(instruction.Args) < 3 {
			return fmt.Errorf("net.serverListen requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_server_listen(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__net.serverAccept":
		// Args: [server_fd: number] -> object { fd: number, ip: string, port: number }
		if len(instruction.Args) < 1 {
			return fmt.Errorf("net.serverAccept requires 1 argument")
		}
		clientFdSlot := instruction.Result + ".cfd.slot"
		clientIpSlot := instruction.Result + ".cip.slot"
		clientPortSlot := instruction.Result + ".cport.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", clientFdSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", clientIpSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", clientPortSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_net_server_accept(double %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], clientFdSlot, clientIpSlot, clientPortSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		cfdVal := instruction.Result + ".cfd"
		cipVal := instruction.Result + ".cip"
		cportVal := instruction.Result + ".cport"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", cfdVal, clientFdSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", cipVal, clientIpSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", cportVal, clientPortSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		st0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", st0, instruction.Result, cfdVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)

		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 1, ptr %%%s)\n", st1, instruction.Result, cipVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)

		st2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", st2, instruction.Result, cportVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st2)
		return nil

	default:
		return fmt.Errorf("unknown net intrinsic %q", instruction.Callee)
	}
}
