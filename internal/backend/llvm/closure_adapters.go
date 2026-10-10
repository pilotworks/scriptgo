package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func collectClosureCallees(instructions []ir.Instruction, callees map[string]bool) {
	for _, instruction := range instructions {
		if instruction.Op == ir.OpClosure {
			callees[instruction.Callee] = true
		}
		collectClosureCallees(instruction.Then, callees)
		collectClosureCallees(instruction.Else, callees)
		collectClosureCallees(instruction.Cond, callees)
		collectClosureCallees(instruction.Body, callees)
		collectClosureCallees(instruction.Step, callees)
		collectClosureCallees(instruction.Catch, callees)
		collectClosureCallees(instruction.Finally, callees)
	}
}

func emitClosureInvokeAdapter(function ir.Function) string {
	const valueType = "{ i32, i32, i64, i64 }"
	name := mangleFunctionName(function.Name)
	var out strings.Builder
	// The adapter takes the arguments as an array of argc boxed values; a
	// parameter past argc reads scriptgo_closure_absent_arg (undefined).
	fmt.Fprintf(&out, "define internal void %s(ptr %%env, i32 %%argc, ptr %%argv, ptr %%out) nounwind {\n", functionSymbol(name+"$invoke"))
	valueCount := 0
	for _, param := range function.Parameters {
		if param.Name != "__env_ctx" {
			valueCount++
		}
	}
	for i := 0; i < valueCount; i++ {
		fmt.Fprintf(&out, "  %%in%d = icmp sgt i32 %%argc, %d\n", i, i)
		fmt.Fprintf(&out, "  %%slot%d = getelementptr inbounds %s, ptr %%argv, i64 %d\n", i, valueType, i)
		fmt.Fprintf(&out, "  %%arg%d = select i1 %%in%d, ptr %%slot%d, ptr @scriptgo_closure_absent_arg\n", i, i, i)
		fmt.Fprintf(&out, "  %%tp%d = getelementptr inbounds %s, ptr %%arg%d, i32 0, i32 0\n", i, valueType, i)
		fmt.Fprintf(&out, "  %%t%d = load i32, ptr %%tp%d\n", i, i)
		fmt.Fprintf(&out, "  %%fp%d = getelementptr inbounds %s, ptr %%arg%d, i32 0, i32 1\n", i, valueType, i)
		fmt.Fprintf(&out, "  %%f%d = load i32, ptr %%fp%d\n", i, i)
		fmt.Fprintf(&out, "  %%pp%d = getelementptr inbounds %s, ptr %%arg%d, i32 0, i32 2\n", i, valueType, i)
		fmt.Fprintf(&out, "  %%p%d = load i64, ptr %%pp%d\n", i, i)
	}

	var callArgs []string
	valIdx := 0
	for _, param := range function.Parameters {
		if param.Name == "__env_ctx" {
			callArgs = append(callArgs, "ptr %env")
			continue
		}
		if isRawCallbackParameter(param) {
			callArgs = append(callArgs, fmt.Sprintf("i32 %%t%d, i32 %%f%d, i64 %%p%d", valIdx, valIdx, valIdx))
			valIdx++
			continue
		}
		pType := llvmType(param.Type)
		switch param.Type {
		case ir.TypeNumber:
			casted := fmt.Sprintf("arg.num.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = bitcast i64 %%p%d to double\n", casted, valIdx)
			callArgs = append(callArgs, fmt.Sprintf("double %%%s", casted))
		case ir.TypeBool:
			casted := fmt.Sprintf("arg.bool.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = trunc i64 %%p%d to i1\n", casted, valIdx)
			if pType == "zeroext i1" {
				callArgs = append(callArgs, fmt.Sprintf("zeroext i1 %%%s", casted))
			} else {
				callArgs = append(callArgs, fmt.Sprintf("i1 %%%s", casted))
			}
		case ir.TypeUnknown:
			b0 := fmt.Sprintf("arg.box0.%d", valIdx)
			b1 := fmt.Sprintf("arg.box1.%d", valIdx)
			b2 := fmt.Sprintf("arg.box2.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s zeroinitializer, i32 %%t%d, 0\n", b0, valueType, valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s %%%s, i32 %%f%d, 1\n", b1, valueType, b0, valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s %%%s, i64 %%p%d, 2\n", b2, valueType, b1, valIdx)
			callArgs = append(callArgs, fmt.Sprintf("%s %%%s", valueType, b2))
		default:
			if pType == "ptr" {
				casted := fmt.Sprintf("arg.ptr.%d", valIdx)
				fmt.Fprintf(&out, "  %%%s = inttoptr i64 %%p%d to ptr\n", casted, valIdx)
				callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", casted))
			} else {
				callArgs = append(callArgs, fmt.Sprintf("%s zeroinitializer", pType))
			}
		}
		valIdx++
	}
	args := strings.Join(callArgs, ", ")
	if functionReadsMoreArguments(function.Body) {
		// Arguments past the fourth reach the function through the
		// extra-argument slot.
		fmt.Fprintf(&out, "  %%more.count = sub i32 %%argc, %d\n", closureRawParameterCount)
		fmt.Fprintf(&out, "  %%more.args = getelementptr %s, ptr %%argv, i64 %d\n", valueType, closureRawParameterCount)
		fmt.Fprintf(&out, "  call void @scriptgo_closure_more_set(ptr %s, i32 %%more.count, ptr %%more.args)\n", functionSymbol(name))
	}

	if function.ReturnType == ir.TypeUnknown {
		fmt.Fprintf(&out, "  %%result = call %s %s(%s)\n", valueType, functionSymbol(name), args)
		fmt.Fprintf(&out, "  store %s %%result, ptr %%out\n", valueType)
		out.WriteString("  ret void\n}\n\n")
		return out.String()
	}
	if function.ReturnType == ir.TypeVoid {
		fmt.Fprintf(&out, "  call void %s(%s)\n", functionSymbol(name), args)
		fmt.Fprintf(&out, "  store %s zeroinitializer, ptr %%out\n", valueType)
		out.WriteString("  ret void\n}\n\n")
		return out.String()
	}

	returnType := llvmType(function.ReturnType)
	fmt.Fprintf(&out, "  %%typed = call %s %s(%s)\n", returnType, functionSymbol(name), args)
	tag := closureReturnTag(function.ReturnType)
	payload := "%typed"
	switch function.ReturnType {
	case ir.TypeNumber:
		out.WriteString("  %payload = bitcast double %typed to i64\n")
		payload = "%payload"
	case ir.TypeBool:
		out.WriteString("  %payload = zext i1 %typed to i64\n")
		payload = "%payload"
	case ir.TypeBigInt:
	default:
		out.WriteString("  %payload = ptrtoint ptr %typed to i64\n")
		payload = "%payload"
	}
	fmt.Fprintf(&out, "  %%boxed.0 = insertvalue %s zeroinitializer, i32 %d, 0\n", valueType, tag)
	fmt.Fprintf(&out, "  %%boxed.1 = insertvalue %s %%boxed.0, i64 %s, 2\n", valueType, payload)
	fmt.Fprintf(&out, "  store %s %%boxed.1, ptr %%out\n", valueType)
	out.WriteString("  ret void\n}\n\n")
	return out.String()
}

func mangleFunctionName(name string) string {
	switch name {
	case "close", "open", "read", "write", "exit", "abort", "link", "unlink", "remove", "rename", "stat", "pipe", "fork", "kill", "signal", "listen", "connect", "bind", "accept", "send", "recv", "select", "poll", "system", "pause", "sleep", "alarm":
		return name + "$user"
	default:
		return name
	}
}

// functionSymbol returns the LLVM global identifier for a lowered function.
// Names that are not valid bare LLVM identifiers (for example private
// members such as "C_#method_impl") are emitted as quoted names, which LLVM
// accepts for any byte sequence.
func functionSymbol(name string) string {
	bare := name != ""
	for i := 0; i < len(name) && bare; i++ {
		c := name[i]
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '$' || c == '.' || c == '_' || c == '-'
		bare = isLetter || (i > 0 && c >= '0' && c <= '9')
	}
	if bare {
		return "@" + name
	}
	var b strings.Builder
	b.WriteString(`@"`)
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c == '"' || c == '\\' || c < 0x20 || c >= 0x7f {
			fmt.Fprintf(&b, `\%02X`, c)
			continue
		}
		b.WriteByte(c)
	}
	b.WriteByte('"')
	return b.String()
}

// closureRawParameterCount is the number of arguments the closure ABI
// passes in registers; a closure call passes the rest through
// scriptgo_closure_more_set (see internal/runtime/native/closures).
const closureRawParameterCount = 4

// emitClosureMoreIntrinsic lowers the closure argument intrinsics: the reads
// a closure with more than four parameters makes on entry
// (__closure.more_arg [index] is argument 4 + index of the current call,
// undefined when absent; __closure.more_done releases the arguments), and
// __closure.apply [closure, array, restIndex], a call with an array's
// elements, those from restIndex on packed into one rest array.
func (e *functionEmitter) emitClosureMoreIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	self := functionSymbol(mangleFunctionName(e.function.Name))
	switch instruction.Callee {
	case "__closure.more_arg":
		if len(instruction.Args) != 1 || instruction.Result == "" {
			return fmt.Errorf("closure.more_arg has invalid signature")
		}
		index := e.resolveArg(out, instruction.Args[0])
		slot := instruction.Result + ".more.slot"
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", slot)
		fmt.Fprintf(out, "  %%%s.more.index = fptosi double %%%s to i32\n", instruction.Result, index)
		fmt.Fprintf(out, "  call void @scriptgo_closure_more_arg(ptr %s, i32 %%%s.more.index, ptr %%%s)\n", self, instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__closure.more_done":
		fmt.Fprintf(out, "  call void @scriptgo_closure_more_done(ptr %s)\n", self)
		return nil
	case "__closure.apply":
		if len(instruction.Args) != 3 || instruction.Result == "" {
			return fmt.Errorf("closure.apply requires a closure, an argument array and a rest index")
		}
		closure := e.ensurePointerArg(out, instruction.Args[0])
		array := e.ensurePointerArg(out, instruction.Args[1])
		restIndex := e.resolveArg(out, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s.apply.rest = fptosi double %%%s to i32\n", instruction.Result, restIndex)
		slot := instruction.Result + ".apply.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_closure_apply(ptr %%%s, ptr %%%s, i32 %%%s.apply.rest, ptr %%%s)\n", status, closure, array, instruction.Result, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, slot)
		e.types[instruction.Result] = ir.TypeUnknown
		return nil
	}
	return fmt.Errorf("unknown closure intrinsic %q", instruction.Callee)
}

// functionReadsMoreArguments reports whether a closure reads arguments past
// the fourth (it has more than four parameters).
func functionReadsMoreArguments(body []ir.Instruction) bool {
	for _, instruction := range body {
		if instruction.Op == ir.OpCall && instruction.Callee == "__closure.more_arg" {
			return true
		}
	}
	return false
}
