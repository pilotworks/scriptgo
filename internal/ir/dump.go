package ir

import (
	"fmt"
	"strconv"
	"strings"
)

// Dump returns a deterministic human-readable representation of verified IR.
func (m Module) Dump() (string, error) {
	if err := m.Verify(); err != nil {
		return "", err
	}
	var out strings.Builder
	fmt.Fprintf(&out, "module %s statements=%d\n", strconv.Quote(m.SourcePath), m.StatementCount)
	for _, shape := range m.Shapes {
		fmt.Fprintf(&out, "shape %s %s\n", shape.Name, dumpSpan(shape.Span))
		for _, field := range shape.Fields {
			fmt.Fprintf(&out, "  field %s: %s %s\n", field.Name, field.Type, dumpSpan(field.Span))
		}
	}
	for _, function := range m.Functions {
		fmt.Fprintf(&out, "function %s(%s) -> %s %s\n", function.Name, dumpParameters(function.Parameters), function.ReturnType, dumpSpan(function.Span))
		for _, instruction := range function.Body {
			fmt.Fprintf(&out, "  %s %s\n", dumpInstruction(instruction), dumpSpan(instruction.Span))
		}
	}
	return out.String(), nil
}

func dumpParameters(parameters []Parameter) string {
	values := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		values = append(values, fmt.Sprintf("%s: %s", parameter.Name, parameter.Type))
	}
	return strings.Join(values, ", ")
}

func dumpInstructionWithIndent(instruction Instruction, indent string) string {
	result := ""
	if instruction.Result != "" {
		result = "%" + instruction.Result + ": " + string(instruction.Type) + " = "
	}
	parts := []string{instruction.Op}
	if instruction.Operator != "" {
		parts = append(parts, strconv.Quote(instruction.Operator))
	}
	if instruction.Callee != "" {
		parts = append(parts, instruction.Callee)
	}
	if instruction.Field != "" {
		parts = append(parts, fmt.Sprintf(".%s[%d]", instruction.Field, instruction.FieldIndex))
	}
	if instruction.Value != "" {
		parts = append(parts, strconv.Quote(instruction.Value))
	}
	if len(instruction.Args) > 0 {
		args := make([]string, len(instruction.Args))
		for index, argument := range instruction.Args {
			args[index] = "%" + argument
		}
		parts = append(parts, "["+strings.Join(args, ", ")+"]")
	}
	res := result + strings.Join(parts, " ")
	if len(instruction.Cond) > 0 {
		res += "\n" + indent + "  cond:"
		for _, inst := range instruction.Cond {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	if len(instruction.Then) > 0 {
		res += "\n" + indent + "  then:"
		for _, inst := range instruction.Then {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	if len(instruction.Else) > 0 {
		res += "\n" + indent + "  else:"
		for _, inst := range instruction.Else {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	if len(instruction.Body) > 0 {
		res += "\n" + indent + "  body:"
		for _, inst := range instruction.Body {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	if len(instruction.Catch) > 0 {
		res += "\n" + indent + "  catch:"
		for _, inst := range instruction.Catch {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	if len(instruction.Finally) > 0 {
		res += "\n" + indent + "  finally:"
		for _, inst := range instruction.Finally {
			res += "\n" + indent + "    " + dumpInstructionWithIndent(inst, indent+"    ")
		}
	}
	return res
}

func dumpInstruction(instruction Instruction) string {
	return dumpInstructionWithIndent(instruction, "")
}

func dumpSpan(span SourceSpan) string {
	if span.Path == "" {
		return "@<unknown>"
	}
	return fmt.Sprintf("@%s:%d+%d", span.Path, span.Offset, span.Length)
}
