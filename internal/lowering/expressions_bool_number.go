package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// numericBoolOperator reports operators that apply ToNumber to a boolean
// operand when the other operand is a boolean or a number: arithmetic,
// bitwise, shift, and relational operators. `+` with a string operand is
// concatenation and is handled elsewhere; equality keeps its own rules.
func numericBoolOperator(operator string) bool {
	switch operator {
	case "+", "-", "*", "/", "%", "**", "&", "|", "^", "<<", ">>", ">>>", "<", "<=", ">", ">=":
		return true
	}
	return false
}

// coerceBoolOperandsToNumber rewrites boolean operands of a numeric operator
// to 1 or 0 (ToNumber) so `true + true` and `2 * false` lower as number
// arithmetic. Operands of any other type are left unchanged.
func coerceBoolOperandsToNumber(path string, expression *frontend.SyntaxExpression, function *ir.Function, counter *int, left *string, leftType *ir.Type, right *string, rightType *ir.Type) {
	if !numericBoolOperator(expression.Operator) {
		return
	}
	if *leftType != ir.TypeBool && *rightType != ir.TypeBool {
		return
	}
	numericOrBool := func(typ ir.Type) bool { return typ == ir.TypeBool || typ == ir.TypeNumber }
	if !numericOrBool(*leftType) || !numericOrBool(*rightType) {
		return
	}
	if *leftType == ir.TypeBool {
		*left, *leftType = boolToNumber(path, expression.Span, function, counter, *left), ir.TypeNumber
	}
	if *rightType == ir.TypeBool {
		*right, *rightType = boolToNumber(path, expression.Span, function, counter, *right), ir.TypeNumber
	}
}

// boolToNumber lowers ToNumber of a boolean value: 1 for true, 0 for false.
func boolToNumber(path string, span frontend.SourceSpan, function *ir.Function, counter *int, value string) string {
	irSpan := toIRSpan(path, span)
	one, zero, number := nextTemp(counter), nextTemp(counter), nextTemp(counter)
	function.Body = append(function.Body,
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: one, Value: "1", Span: irSpan},
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zero, Value: "0", Span: irSpan},
		ir.Instruction{Op: ir.OpSelect, Type: ir.TypeNumber, Result: number, Args: []string{value, one, zero}, Span: irSpan},
	)
	return number
}
