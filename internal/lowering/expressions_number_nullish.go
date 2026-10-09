package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerNumberNullishCompare lowers `x === undefined`, `x == null` and their
// negations for a value x in number storage. undefined and null are NaNs
// with reserved payloads there (the runtime's number-storage markers), so
// the test compares bits rather than asking whether x is NaN: a NaN number
// is neither.
func lowerNumberNullishCompare(path string, expression *frontend.SyntaxExpression, operand string, nullish *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int) (string, ir.Type) {
	if *result == "" {
		*result = nextTemp(counter)
	}
	span := toIRSpan(path, expression.Span)
	negate := expression.Operator == "!==" || expression.Operator == "!="
	callee := "__number.isNullish"
	if expression.Operator == "===" || expression.Operator == "!==" {
		callee = "__number.isUndefined"
		if nullish.Kind == "null" {
			callee = "__number.isNull"
		}
	}
	test := *result
	if negate {
		test = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: test, Callee: callee, Args: []string{operand}, Span: span})
	if negate {
		falseConst := nextTemp(counter)
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: falseConst, Value: "false", Span: span},
			ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: *result, Operator: "==", Args: []string{test, falseConst}, Span: span},
		)
	}
	return *result, ir.TypeBool
}

func isEquality(operator string) bool {
	return operator == "===" || operator == "!==" || operator == "==" || operator == "!="
}
