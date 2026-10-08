package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	switch expression.Kind {
	case "number":
		typ := ir.TypeNumber
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bigint":
		typ := ir.TypeBigInt
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "regex":
		return lowerRegexLiteral(path, expression, result, function, env, counter, shapes, signatures)
	case "string":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, StringLiteral: true, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bool":
		typ := ir.TypeBool
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "null":
		typ := ir.Type("ptr")
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "null", Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "undefined":
		typ := ir.TypeVoid
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "undefined", Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "array":
		return lowerArrayLiteral(path, expression, result, function, env, counter, shapes, signatures)

	case "index", "optional_index":
		return lowerIndexExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "this":
		typ, ok := env["this"]
		if !ok {
			typ = ir.TypeObject
		}
		return "this", typ, nil
	case "identifier":
		return lowerIdentifierExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "unary":
		return lowerUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "postfix_unary":
		return lowerPostfixUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "as":
		return lowerAsExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "non_null":
		return lowerNonNullExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "typeof":
		return lowerTypeofExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "binary":
		return lowerBinaryExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "template":
		return lowerTemplateLiteral(path, expression, result, function, env, counter, shapes, signatures)

	case "conditional":
		return lowerConditionalExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "property", "optional_property":
		return lowerPropertyExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "object_literal":
		return lowerObjectLiteralExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "new":
		return lowerNewExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "call":
		return lowerCallExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "optional_call":
		return lowerOptionalCallExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "tagged_template":
		return lowerTaggedTemplate(path, expression, result, function, env, counter, shapes, signatures)
	case "arrow_function":
		if expression.Function != nil {
			return lowerClosureExpression(path, expression.Function, result, function, env, counter, shapes, signatures)
		}
		return "", "", fmt.Errorf("arrow function expression missing body")
	case "await":
		return lowerAwaitExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "spread":
		if expression.Left != nil {
			return lowerExpression(path, expression.Left, result, function, env, counter, shapes, signatures)
		}
		if expression.Right != nil {
			return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
		}
		return "", "", fmt.Errorf("invalid spread expression")
	case "yield", "yield_star":
		if expression.Left != nil {
			return lowerExpression(path, expression.Left, result, function, env, counter, shapes, signatures)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: result, Value: "0", Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, nil
	default:
		return "", "", fmt.Errorf("unsupported expression %q", expression.Kind)
	}
}

func arrayLiteralElementType(arrayType ir.Type) ir.Type {
	if element, ok := strings.CutSuffix(string(arrayType), "[]"); ok {
		return toIRType(element)
	}
	switch arrayType {
	case ir.TypeNumberArray:
		return ir.TypeNumber
	case ir.TypeStringArray:
		return ir.TypeString
	case ir.TypeBoolArray:
		return ir.TypeBool
	case ir.TypeBigIntArray:
		return ir.TypeBigInt
	case ir.TypeSymbolArray:
		return ir.TypeSymbol
	default:
		return ir.TypeUnknown
	}
}

func nextTemp(counter *int) string {
	for {
		value := "t" + strconv.Itoa(*counter)
		*counter++
		if _, exists := topLevelVars[value]; exists {
			continue
		}
		return value
	}
}

func isComparison(operator string) bool {
	return operator == "==" || operator == "===" || operator == "!=" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}

func arrayElementType(arrType ir.Type) ir.Type {
	switch arrType {
	case ir.TypeNumberArray:
		return ir.TypeNumber
	case ir.TypeStringArray:
		return ir.TypeString
	case ir.TypeBoolArray:
		return ir.TypeBool
	case ir.TypeBigIntArray:
		return ir.TypeBigInt
	case ir.TypeSymbolArray:
		return ir.TypeSymbol
	default:
		s := string(arrType)
		if strings.HasSuffix(s, "[]") {
			return toIRType(strings.TrimSuffix(s, "[]"))
		}
		return ir.TypeNumber
	}
}
