package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitDnsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__dns.lookup":
		// Args: [hostname: string, family: number]
		// Result: object { address: string, family: number }
		if len(instruction.Args) < 1 {
			return fmt.Errorf("dns.lookup requires at least 1 argument")
		}
		famArg := "0.0"
		if len(instruction.Args) > 1 {
			famArg = "%" + resolvedArgs[1]
		}

		addrSlot := instruction.Result + ".addr.slot"
		famSlot := instruction.Result + ".fam.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", addrSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", famSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_lookup(ptr %%%s, double %s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], famArg, addrSlot, famSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		addrVal := instruction.Result + ".addr"
		famVal := instruction.Result + ".fam"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", addrVal, addrSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", famVal, famSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)
		if err := e.emitDnsObjectDescriptor(out, instruction.Result, instruction.Callee); err != nil {
			return err
		}

		st0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, addrVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)

		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", st1, instruction.Result, famVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)
		return nil

	case "__dns.lookupService":
		// Args: [address: string, port: number]
		// Result: object { hostname: string, service: string }
		if len(instruction.Args) < 2 {
			return fmt.Errorf("dns.lookupService requires 2 arguments")
		}
		hostSlot := instruction.Result + ".host.slot"
		servSlot := instruction.Result + ".serv.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", hostSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", servSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_lookup_service(ptr %%%s, double %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], hostSlot, servSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		hostVal := instruction.Result + ".host"
		servVal := instruction.Result + ".serv"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", hostVal, hostSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", servVal, servSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)
		if err := e.emitDnsObjectDescriptor(out, instruction.Result, instruction.Callee); err != nil {
			return err
		}

		st0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, hostVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)

		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 1, ptr %%%s)\n", st1, instruction.Result, servVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)
		return nil

	case "__dns.reverse":
		// Args: [ip: string] -> string[]
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_reverse(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dns.resolveStrings":
		// Args: [hostname: string, rrtype: string] -> string[]
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_strings(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown dns intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitDnsObjectDescriptor(out *strings.Builder, result string, callee string) error {
	descriptor := intrinsicObjectDescriptor(callee)
	if descriptor == "" {
		return nil
	}
	global, ok := e.stringsByValue[descriptor]
	if !ok {
		return fmt.Errorf("missing DNS object descriptor %q", descriptor)
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_type_set(ptr %%%s, ptr %s)\n", status, result, global)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	return nil
}
