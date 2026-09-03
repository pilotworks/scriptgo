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

	case "__dns.resolveTxt":
		// Args: [hostname: string] -> string[]
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_txt(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__dns.resolveMx":
		// Args: [hostname: string] -> object { exchanges: string[], priorities: number[] }
		exchSlot := instruction.Result + ".exch.slot"
		prioSlot := instruction.Result + ".prio.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", exchSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", prioSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_mx(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], exchSlot, prioSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		exchVal := instruction.Result + ".exch"
		prioVal := instruction.Result + ".prio"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", exchVal, exchSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", prioVal, prioSlot)
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, exchVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)
		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", st1, instruction.Result, prioVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)
		return nil

	case "__dns.resolveSrv":
		// Args: [hostname: string] -> object { names: string[], ports: number[], priorities: number[], weights: number[] }
		namesSlot := instruction.Result + ".names.slot"
		portsSlot := instruction.Result + ".ports.slot"
		prioSlot := instruction.Result + ".prio.slot"
		weightSlot := instruction.Result + ".weight.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", namesSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", portsSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", prioSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", weightSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_srv(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], namesSlot, portsSlot, prioSlot, weightSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		namesVal := instruction.Result + ".names"
		portsVal := instruction.Result + ".ports"
		prioVal := instruction.Result + ".prio"
		weightVal := instruction.Result + ".weight"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", namesVal, namesSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", portsVal, portsSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", prioVal, prioSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", weightVal, weightSlot)
		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 4, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)
		if err := e.emitDnsObjectDescriptor(out, instruction.Result, instruction.Callee); err != nil {
			return err
		}
		for idx, val := range []string{namesVal, portsVal, prioVal, weightVal} {
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 %d, ptr %%%s)\n", st, instruction.Result, idx, val)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st)
		}
		return nil

	case "__dns.resolveSoa":
		// Args: [hostname: string] -> object { nsname: string, hostmaster: string, serial: number, refresh: number, retry: number, expire: number, minttl: number }
		nsSlot := instruction.Result + ".ns.slot"
		hmSlot := instruction.Result + ".hm.slot"
		serSlot := instruction.Result + ".ser.slot"
		refSlot := instruction.Result + ".ref.slot"
		retSlot := instruction.Result + ".ret.slot"
		expSlot := instruction.Result + ".exp.slot"
		minSlot := instruction.Result + ".min.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", nsSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", hmSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", serSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", refSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", retSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", expSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", minSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_soa(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], nsSlot, hmSlot, serSlot, refSlot, retSlot, expSlot, minSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		nsVal := instruction.Result + ".ns"
		hmVal := instruction.Result + ".hm"
		serVal := instruction.Result + ".ser"
		refVal := instruction.Result + ".ref"
		retVal := instruction.Result + ".ret"
		expVal := instruction.Result + ".exp"
		minVal := instruction.Result + ".min"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", nsVal, nsSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", hmVal, hmSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", serVal, serSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", refVal, refSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", retVal, retSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", expVal, expSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", minVal, minSlot)
		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 7, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)
		if err := e.emitDnsObjectDescriptor(out, instruction.Result, instruction.Callee); err != nil {
			return err
		}
		st0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, nsVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)
		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 1, ptr %%%s)\n", st1, instruction.Result, hmVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)
		for idx, val := range []string{serVal, refVal, retVal, expVal, minVal} {
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", st, instruction.Result, idx+2, val)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st)
		}
		return nil

	case "__dns.resolveCaa":
		// Args: [hostname: string] -> object { criticals: number[], issues: string[] }
		critSlot := instruction.Result + ".crit.slot"
		issueSlot := instruction.Result + ".issue.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", critSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", issueSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_caa(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], critSlot, issueSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		critVal := instruction.Result + ".crit"
		issueVal := instruction.Result + ".issue"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", critVal, critSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", issueVal, issueSlot)
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", st0, instruction.Result, critVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st0)
		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", st1, instruction.Result, issueVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1)
		return nil

	case "__dns.resolveNaptr":
		// Args: [hostname: string] -> object { flags: string[], services: string[], regexps: string[], replacements: string[], orders: number[], preferences: number[] }
		flagsSlot := instruction.Result + ".flags.slot"
		servSlot := instruction.Result + ".serv.slot"
		regSlot := instruction.Result + ".reg.slot"
		replSlot := instruction.Result + ".repl.slot"
		ordSlot := instruction.Result + ".ord.slot"
		prefSlot := instruction.Result + ".pref.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", flagsSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", servSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", regSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", replSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", ordSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", prefSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dns_resolve_naptr(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], flagsSlot, servSlot, regSlot, replSlot, ordSlot, prefSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		flagsVal := instruction.Result + ".flags"
		servVal := instruction.Result + ".serv"
		regVal := instruction.Result + ".reg"
		replVal := instruction.Result + ".repl"
		ordVal := instruction.Result + ".ord"
		prefVal := instruction.Result + ".pref"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", flagsVal, flagsSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", servVal, servSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", regVal, regSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", replVal, replSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", ordVal, ordSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", prefVal, prefSlot)
		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 6, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)
		if err := e.emitDnsObjectDescriptor(out, instruction.Result, instruction.Callee); err != nil {
			return err
		}
		for idx, val := range []string{flagsVal, servVal, regVal, replVal, ordVal, prefVal} {
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 %d, ptr %%%s)\n", st, instruction.Result, idx, val)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st)
		}
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
