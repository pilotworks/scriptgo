package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) tlsStatus(out *strings.Builder) string {
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	return status
}

func (e *functionEmitter) emitTLSString(out *strings.Builder, call string, result string) {
	slot := result + ".slot"
	status := e.tlsStatus(out)
	fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
	separator := ", "
	if strings.HasSuffix(call, "(") {
		separator = ""
	}
	fmt.Fprintf(out, "  %%%s = call i32 %s%s ptr %%%s)\n", status, call, separator, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", result, slot)
}

func (e *functionEmitter) emitTLSHandle(out *strings.Builder, call string, result string) {
	slot := result + ".slot"
	status := e.tlsStatus(out)
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	separator := ", "
	if strings.HasSuffix(call, "(") {
		separator = ""
	}
	fmt.Fprintf(out, "  %%%s = call i32 %s%s ptr %%%s)\n", status, call, separator, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", result, slot)
}

func (e *functionEmitter) emitTLSBool(out *strings.Builder, call string, result string) {
	slot := result + ".slot"
	status := e.tlsStatus(out)
	fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
	separator := ", "
	if strings.HasSuffix(call, "(") {
		separator = ""
	}
	fmt.Fprintf(out, "  %%%s = call i32 %s%s ptr %%%s)\n", status, call, separator, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s.i32 = load i32, ptr %%%s\n", result, slot)
	fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s.i32, 0\n", result, result)
}

func (e *functionEmitter) emitTLSIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	args := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		args[i] = e.resolveArg(out, arg)
	}
	ptr := func(index int) string { return "ptr %" + args[index] }
	double := func(index int) string { return "double %" + args[index] }
	bool32 := func(index int) string {
		name := fmt.Sprintf("tls.bool.%d", e.loadCounter)
		value := fmt.Sprintf("tls.bool.f64.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", name, args[index])
		fmt.Fprintf(out, "  %%%s = uitofp i32 %%%s to double\n", value, name)
		return "double %" + value
	}

	switch instruction.Callee {
	case "__tls.contextCreate":
		if len(args) != 7 {
			return fmt.Errorf("tls.contextCreate requires 7 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_context_create(%s, %s, %s, %s, %s, %s, %s", ptr(0), ptr(1), ptr(2), ptr(3), ptr(4), ptr(5), bool32(6)), instruction.Result)
	case "__tls.socketCreate":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketCreate requires 2 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_socket_create(%s, %s", double(0), bool32(1)), instruction.Result)
	case "__tls.socketConnect":
		if len(args) != 6 {
			return fmt.Errorf("tls.socketConnect requires 6 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_socket_connect(%s, %s, %s, %s, %s, %s", double(0), ptr(1), double(2), ptr(3), bool32(4), ptr(5)), instruction.Result)
	case "__tls.socketAdopt":
		if len(args) != 7 {
			return fmt.Errorf("tls.socketAdopt requires 7 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_socket_adopt(%s, %s, %s, %s, %s, %s, %s", double(0), double(1), ptr(2), bool32(3), bool32(4), bool32(5), ptr(6)), instruction.Result)
	case "__tls.socketWrite":
		if len(args) != 3 {
			return fmt.Errorf("tls.socketWrite requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_write(%s, %s, %s, ptr %%%s)\n", status, double(0), ptr(1), double(2), slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__tls.socketWriteBytes":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketWriteBytes requires 2 arguments")
		}
		slot := instruction.Result + ".slot"
		status := e.tlsStatus(out)
		view := e.ensurePointerArg(out, instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_write_bytes(%s, ptr %%%s, ptr %%%s)\n", status, double(0), view, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__tls.socketRead":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketRead requires 2 arguments")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_socket_read(%s, %s", double(0), double(1)), instruction.Result)
	case "__tls.pairWrite":
		if len(args) != 4 {
			return fmt.Errorf("tls.pairWrite requires 4 arguments")
		}
		slot := instruction.Result + ".slot"
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_pair_write(%s, %s, %s, %s, ptr %%%s)\n", status, double(0), double(1), ptr(2), double(3), slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__tls.pairWriteBytes":
		if len(args) != 3 {
			return fmt.Errorf("tls.pairWriteBytes requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := e.tlsStatus(out)
		view := e.ensurePointerArg(out, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_pair_write_bytes(%s, %s, ptr %%%s, ptr %%%s)\n", status, double(0), double(1), view, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__tls.pairRead":
		if len(args) != 3 {
			return fmt.Errorf("tls.pairRead requires 3 arguments")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_socket_pair_read(%s, %s, %s", double(0), double(1), double(2)), instruction.Result)
	case "__tls.socketClose":
		if len(args) != 1 {
			return fmt.Errorf("tls.socketClose requires 1 argument")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_close(%s)\n", status, double(0))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.socketInfo":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketInfo requires 2 arguments")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_socket_info(%s, %s", double(0), ptr(1)), instruction.Result)
	case "__tls.socketNumber":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketNumber requires 2 arguments")
		}
		slot := instruction.Result + ".slot"
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_number(%s, %s, ptr %%%s)\n", status, double(0), ptr(1), slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__tls.socketBool":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketBool requires 2 arguments")
		}
		e.emitTLSBool(out, fmt.Sprintf("@scriptgo_tls_socket_bool(%s, %s", double(0), ptr(1)), instruction.Result)
	case "__tls.exportKeyingMaterial":
		if len(args) != 4 {
			return fmt.Errorf("tls.exportKeyingMaterial requires 4 arguments")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_socket_export_keying_material(%s, %s, %s, %s", double(0), double(1), ptr(2), ptr(3)), instruction.Result)
	case "__tls.socketSetOption":
		if len(args) != 3 {
			return fmt.Errorf("tls.socketSetOption requires 3 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_set_option(%s, %s, %s)\n", status, double(0), ptr(1), double(2))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.socketSetServername":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketSetServername requires 2 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_set_servername(%s, %s)\n", status, double(0), ptr(1))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.socketSetSession":
		if len(args) != 2 {
			return fmt.Errorf("tls.socketSetSession requires 2 arguments")
		}
		status := e.tlsStatus(out)
		view := e.ensurePointerArg(out, instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_set_session(%s, ptr %%%s)\n", status, double(0), view)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.socketSetKeyCert":
		if len(args) != 3 {
			return fmt.Errorf("tls.socketSetKeyCert requires 3 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_socket_set_key_cert(%s, %s, %s)\n", status, double(0), ptr(1), ptr(2))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.socketRenegotiate":
		if len(args) != 1 {
			return fmt.Errorf("tls.socketRenegotiate requires 1 argument")
		}
		e.emitTLSBool(out, fmt.Sprintf("@scriptgo_tls_socket_renegotiate(%s", double(0)), instruction.Result)
	case "__tls.pairCreate":
		if len(args) != 4 {
			return fmt.Errorf("tls.pairCreate requires 4 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_pair_create(%s, %s, %s, %s", double(0), bool32(1), bool32(2), bool32(3)), instruction.Result)
	case "__tls.serverListen":
		if len(args) != 6 {
			return fmt.Errorf("tls.serverListen requires 6 arguments")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_server_listen(%s, %s, %s, %s, %s, %s", double(0), bool32(1), bool32(2), ptr(3), double(4), double(5)), instruction.Result)
	case "__tls.serverAccept":
		if len(args) != 1 {
			return fmt.Errorf("tls.serverAccept requires 1 argument")
		}
		e.emitTLSHandle(out, fmt.Sprintf("@scriptgo_tls_server_accept(%s", double(0)), instruction.Result)
	case "__tls.serverClose":
		if len(args) != 1 {
			return fmt.Errorf("tls.serverClose requires 1 argument")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_server_close(%s)\n", status, double(0))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.serverInfo":
		if len(args) != 2 {
			return fmt.Errorf("tls.serverInfo requires 2 arguments")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_server_info(%s, %s", double(0), ptr(1)), instruction.Result)
	case "__tls.serverSetContext":
		if len(args) != 4 {
			return fmt.Errorf("tls.serverSetContext requires 4 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_server_set_context(%s, %s, %s, %s)\n", status, double(0), double(1), bool32(2), bool32(3))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.serverAddContext":
		if len(args) != 3 {
			return fmt.Errorf("tls.serverAddContext requires 3 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_server_add_context(%s, %s, %s)\n", status, double(0), ptr(1), double(2))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.serverSetTicketKeys":
		if len(args) != 2 {
			return fmt.Errorf("tls.serverSetTicketKeys requires 2 arguments")
		}
		status := e.tlsStatus(out)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tls_server_set_ticket_keys(%s, %s)\n", status, double(0), ptr(1))
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	case "__tls.x509ParsePem":
		if len(args) != 1 {
			return fmt.Errorf("tls.x509ParsePem requires 1 argument")
		}
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_x509_parse_pem(%s", ptr(0)), instruction.Result)
	case "__tls.x509ParseBytes":
		if len(args) != 1 {
			return fmt.Errorf("tls.x509ParseBytes requires 1 argument")
		}
		arg := e.ensurePointerArg(out, instruction.Args[0])
		e.emitTLSString(out, fmt.Sprintf("@scriptgo_tls_x509_parse_bytes(ptr %%%s", arg), instruction.Result)
	case "__tls.ciphers":
		if len(args) != 0 {
			return fmt.Errorf("tls.ciphers takes no arguments")
		}
		e.emitTLSString(out, "@scriptgo_tls_ciphers(", instruction.Result)
	case "__tls.rootCertificates":
		if len(args) != 0 {
			return fmt.Errorf("tls.rootCertificates takes no arguments")
		}
		e.emitTLSString(out, "@scriptgo_tls_root_certificates(", instruction.Result)
	case "__tls.systemCertificates":
		if len(args) != 0 {
			return fmt.Errorf("tls.systemCertificates takes no arguments")
		}
		e.emitTLSString(out, "@scriptgo_tls_system_certificates(", instruction.Result)
	case "__tls.extraCertificates":
		if len(args) != 0 {
			return fmt.Errorf("tls.extraCertificates takes no arguments")
		}
		e.emitTLSString(out, "@scriptgo_tls_extra_certificates(", instruction.Result)
	default:
		return fmt.Errorf("unknown TLS intrinsic %q", instruction.Callee)
	}
	return nil
}
