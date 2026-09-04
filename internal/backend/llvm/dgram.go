package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitDgramIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__dgram.socketCreate":
		// Args: [family: number] -> fd: number
		famArg := "4.0"
		if len(instruction.Args) > 0 {
			famArg = "%" + resolvedArgs[0]
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_socket_create(double %s, ptr %%%s)\n",
			status, famArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dgram.bind":
		// Args: [fd: number, address: string, port: number] -> void
		if len(instruction.Args) < 3 {
			return fmt.Errorf("dgram.bind requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_bind(double %%%s, ptr %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.send":
		// Args: [fd: number, data: string, len: number, port: number, address: string] -> sent: number
		if len(instruction.Args) < 5 {
			return fmt.Errorf("dgram.send requires 5 arguments")
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_send(double %%%s, ptr %%%s, double %%%s, double %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], resolvedArgs[4], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dgram.recv":
		// Args: [fd: number, max_len: number] -> object { data: string, bytes: number, address: string, port: number, family: number }
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.recv requires 2 arguments")
		}
		dataSlot := instruction.Result + ".data.slot"
		readSlot := instruction.Result + ".read.slot"
		ipSlot := instruction.Result + ".ip.slot"
		portSlot := instruction.Result + ".port.slot"
		famSlot := instruction.Result + ".fam.slot"

		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dataSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", readSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", ipSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", portSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", famSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_recv(double %%%s, double %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], dataSlot, readSlot, ipSlot, portSlot, famSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		dataVal := instruction.Result + ".data"
		readVal := instruction.Result + ".read"
		ipVal := instruction.Result + ".ip"
		portVal := instruction.Result + ".port"
		famVal := instruction.Result + ".fam"

		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", dataVal, dataSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", readVal, readSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", ipVal, ipSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", portVal, portSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", famVal, famSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 5, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		st0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, dataVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)

		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", st1, instruction.Result, readVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)

		st2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 2, ptr %%%s)\n", st2, instruction.Result, ipVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st2)

		st3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 3, double %%%s)\n", st3, instruction.Result, portVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st3)

		st4 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 4, double %%%s)\n", st4, instruction.Result, famVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st4)
		return nil

	case "__dgram.setBroadcast":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setBroadcast requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_broadcast(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setMulticastTTL":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setMulticastTTL requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_multicast_ttl(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setMulticastLoopback":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setMulticastLoopback requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_multicast_loopback(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setRecvBufferSize":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setRecvBufferSize requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_recv_buffer_size(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setSendBufferSize":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setSendBufferSize requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_send_buffer_size(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setTTL":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setTTL requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_ttl(double %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.setMulticastInterface":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dgram.setMulticastInterface requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_set_multicast_interface(double %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.addMembership":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("dgram.addMembership requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_add_membership(double %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.dropMembership":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("dgram.dropMembership requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_drop_membership(double %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.addSourceSpecificMembership":
		if len(instruction.Args) < 4 {
			return fmt.Errorf("dgram.addSourceSpecificMembership requires 4 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_add_source_specific_membership(double %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.dropSourceSpecificMembership":
		if len(instruction.Args) < 4 {
			return fmt.Errorf("dgram.dropSourceSpecificMembership requires 4 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_drop_source_specific_membership(double %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.connect":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("dgram.connect requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_connect(double %%%s, ptr %%%s, double %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.disconnect":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("dgram.disconnect requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_disconnect(double %%%s)\n",
			status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__dgram.close":
		// Args: [fd: number] -> void
		if len(instruction.Args) < 1 {
			return fmt.Errorf("dgram.close requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dgram_close(double %%%s)\n",
			status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	default:
		return fmt.Errorf("unknown dgram intrinsic %q", instruction.Callee)
	}
}
