package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitChildProcessIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__child_process.execSync":
		cmdArg := "%" + resolvedArgs[0]
		cwdArg := "null"
		inputArg := "null"
		if len(instruction.Args) > 1 {
			cwdArg = "%" + resolvedArgs[1]
		}
		if len(instruction.Args) > 2 {
			inputArg = "%" + resolvedArgs[2]
		}
		stdoutSlot := instruction.Result + ".stdout.slot"
		stderrSlot := instruction.Result + ".stderr.slot"
		statusSlot := instruction.Result + ".status.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stdoutSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stderrSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_exec_sync(ptr %s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, cmdArg, cwdArg, inputArg, stdoutSlot, stderrSlot, statusSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		if instruction.Type == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, stdoutSlot)
		} else {
			stdoutVal := instruction.Result + ".stdout"
			stderrVal := instruction.Result + ".stderr"
			statusVal := instruction.Result + ".status"
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stdoutVal, stdoutSlot)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stderrVal, stderrSlot)
			fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)

			objSlot := instruction.Result + ".obj_slot"
			objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", objStatus, objSlot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

			s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, stdoutVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, stderrVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, statusVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		}
		return nil

	case "__child_process.spawnSync":
		cmdArg := "%" + resolvedArgs[0]
		argsArg := "null"
		cwdArg := "null"
		inputArg := "null"
		if len(instruction.Args) > 1 {
			argsArg = "%" + resolvedArgs[1]
		}
		if len(instruction.Args) > 2 {
			cwdArg = "%" + resolvedArgs[2]
		}
		if len(instruction.Args) > 3 {
			inputArg = "%" + resolvedArgs[3]
		}
		stdoutSlot := instruction.Result + ".stdout.slot"
		stderrSlot := instruction.Result + ".stderr.slot"
		statusSlot := instruction.Result + ".status.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stdoutSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stderrSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_spawn_sync(ptr %s, ptr %s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, cmdArg, argsArg, cwdArg, inputArg, stdoutSlot, stderrSlot, statusSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		stdoutVal := instruction.Result + ".stdout"
		stderrVal := instruction.Result + ".stderr"
		statusVal := instruction.Result + ".status"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stdoutVal, stdoutSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stderrVal, stderrSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, stdoutVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, stderrVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, statusVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		return nil

	case "__child_process.spawnAsync":
		cmdArg := "%" + resolvedArgs[0]
		argsArg := "null"
		cwdArg := "null"
		if len(instruction.Args) > 1 {
			argsArg = "%" + resolvedArgs[1]
		}
		if len(instruction.Args) > 2 {
			cwdArg = "%" + resolvedArgs[2]
		}

		pidSlot := instruction.Result + ".pid.slot"
		stdinSlot := instruction.Result + ".stdin.slot"
		stdoutSlot := instruction.Result + ".stdout.slot"
		stderrSlot := instruction.Result + ".stderr.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", pidSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", stdinSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", stdoutSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", stderrSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_spawn_async(ptr %s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, cmdArg, argsArg, cwdArg, pidSlot, stdinSlot, stdoutSlot, stderrSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		pidVal := instruction.Result + ".pid"
		stdinVal := instruction.Result + ".stdin"
		stdoutVal := instruction.Result + ".stdout"
		stderrVal := instruction.Result + ".stderr"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", pidVal, pidSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", stdinVal, stdinSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", stdoutVal, stdoutSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", stderrVal, stderrSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 4, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s4 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", s1, instruction.Result, pidVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", s2, instruction.Result, stdinVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, stdoutVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 3, double %%%s)\n", s4, instruction.Result, stderrVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s4)
		return nil

	case "__child_process.pollStatus":
		pidArg := "%" + resolvedArgs[0]
		statusSlot := instruction.Result + ".status.slot"
		exitedSlot := instruction.Result + ".exited.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", exitedSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_poll_status(double %s, ptr %%%s, ptr %%%s)\n",
			status, pidArg, statusSlot, exitedSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		statusVal := instruction.Result + ".status"
		exitedVal := instruction.Result + ".exited"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", exitedVal, exitedSlot)

		cmpExited := fmt.Sprintf("cmp.exited.%d", e.loadCounter)
		e.loadCounter++
		zextExited := fmt.Sprintf("zext.exited.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = fcmp ogt double %%%s, 0.0\n", cmpExited, exitedVal)
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", zextExited, cmpExited)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", s1, instruction.Result, statusVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 1, i32 %%%s)\n", s2, instruction.Result, zextExited)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		return nil

	case "__child_process.pipeRead":
		fdArg := "%" + resolvedArgs[0]
		maxLenArg := "%" + resolvedArgs[1]

		dataSlot := instruction.Result + ".data.slot"
		bytesReadSlot := instruction.Result + ".bytes_read.slot"
		eofSlot := instruction.Result + ".eof.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dataSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", bytesReadSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", eofSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_pipe_read(double %s, double %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, fdArg, maxLenArg, dataSlot, bytesReadSlot, eofSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		dataVal := instruction.Result + ".data"
		bytesReadVal := instruction.Result + ".bytes_read"
		eofVal := instruction.Result + ".eof"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", dataVal, dataSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", bytesReadVal, bytesReadSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", eofVal, eofSlot)

		cmpEof := fmt.Sprintf("cmp.eof.%d", e.loadCounter)
		e.loadCounter++
		zextEof := fmt.Sprintf("zext.eof.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = fcmp ogt double %%%s, 0.0\n", cmpEof, eofVal)
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", zextEof, cmpEof)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, dataVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", s2, instruction.Result, bytesReadVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 2, i32 %%%s)\n", s3, instruction.Result, zextEof)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		return nil

	case "__child_process.pipeWrite":
		fdArg := "%" + resolvedArgs[0]
		dataArg := "%" + resolvedArgs[1]
		lenArg := "%" + resolvedArgs[2]

		writtenSlot := instruction.Result + ".written.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", writtenSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_pipe_write(double %s, ptr %s, double %s, ptr %%%s)\n",
			status, fdArg, dataArg, lenArg, writtenSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, writtenSlot)
		return nil

	case "__child_process.pipeClose":
		fdArg := "%" + resolvedArgs[0]
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_pipe_close(double %s)\n", status, fdArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__child_process.kill":
		pidArg := "%" + resolvedArgs[0]
		sigArg := "%" + resolvedArgs[1]
		successSlot := instruction.Result + ".success.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", successSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_kill(double %s, double %s, ptr %%%s)\n",
			status, pidArg, sigArg, successSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		resVal := instruction.Result + ".raw"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", resVal, successSlot)
		fmt.Fprintf(out, "  %%%s = fcmp ogt double %%%s, 0.0\n", instruction.Result, resVal)
		return nil

	default:
		return fmt.Errorf("unknown child_process intrinsic %q", instruction.Callee)
	}
}
