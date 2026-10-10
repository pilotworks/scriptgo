package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerMixedTypeBinary(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, env map[string]ir.Type, counter *int, signatures map[string]ir.Function, left *string, leftType *ir.Type, right *string, rightType *ir.Type) (string, ir.Type, bool, error) {
	if isEquality(expression.Operator) && *leftType == ir.TypeNumber && expression.Right != nil && (expression.Right.Kind == "undefined" || expression.Right.Kind == "null") {
		value, typ := lowerNumberNullishCompare(path, expression, *left, expression.Right, result, function, counter)
		return value, typ, true, nil
	}
	if isEquality(expression.Operator) && *rightType == ir.TypeNumber && expression.Left != nil && (expression.Left.Kind == "undefined" || expression.Left.Kind == "null") {
		value, typ := lowerNumberNullishCompare(path, expression, *right, expression.Left, result, function, counter)
		return value, typ, true, nil
	}
	if isComparison(expression.Operator) && (*leftType == ir.TypeBool || *rightType == ir.TypeBool) && (*leftType == ir.TypeVoid || *rightType == ir.TypeVoid || (expression.Left != nil && (expression.Left.Kind == "undefined" || expression.Left.Kind == "null")) || (expression.Right != nil && (expression.Right.Kind == "undefined" || expression.Right.Kind == "null"))) {
		if *result == "" {
			*result = nextTemp(counter)
		}
		isNot := (expression.Operator == "!==" || expression.Operator == "!=")
		constVal := "false"
		if isNot {
			constVal = "true"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: *result,
			Value:  constVal,
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeBool, true, nil
	}
	if isComparison(expression.Operator) && ((*leftType == ir.TypeVoid && (*rightType == ir.TypePointer || *rightType == "ptr")) || (*rightType == ir.TypeVoid && (*leftType == ir.TypePointer || *leftType == "ptr"))) {
		if *result == "" {
			*result = nextTemp(counter)
		}
		val := "false"
		if expression.Operator == "!==" {
			val = "true"
		} else if expression.Operator == "!=" {
			val = "false"
		} else if expression.Operator == "==" {
			val = "true"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: *result,
			Value:  val,
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeBool, true, nil
	}
	if isComparison(expression.Operator) && (expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined")) && isPointerLikeType(*leftType) {
		*right = nextTemp(counter)
		*rightType = *leftType
		nullVal := "null"
		if expression.Right.Kind == "undefined" {
			nullVal = "undefined"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   *leftType,
			Result: *right,
			Value:  nullVal,
			Span:   toIRSpan(path, expression.Right.Span),
		})
	} else if isComparison(expression.Operator) && (expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined")) && isPointerLikeType(*rightType) {
		*left = nextTemp(counter)
		*leftType = *rightType
		nullVal := "null"
		if expression.Left.Kind == "undefined" {
			nullVal = "undefined"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   *rightType,
			Result: *left,
			Value:  nullVal,
			Span:   toIRSpan(path, expression.Left.Span),
		})
	} else if isComparison(expression.Operator) && (*leftType == ir.TypeUnknown || *rightType == ir.TypeUnknown) {
		if *leftType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*left}, Span: toIRSpan(path, expression.Span)})
			*left = boxed
			*leftType = ir.TypeUnknown
		}
		if *rightType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*right}, Span: toIRSpan(path, expression.Span)})
			*right = boxed
			*rightType = ir.TypeUnknown
		}
	} else if expression.Operator == "+" && (*leftType == ir.TypeString || *rightType == ir.TypeString) {
		converted, err := lowerToString(path, expression.Left.Span, *left, *leftType, function, counter, signatures)
		if err != nil {
			return "", "", true, err
		}
		*left, *leftType = converted, ir.TypeString
		converted, err = lowerToString(path, expression.Right.Span, *right, *rightType, function, counter, signatures)
		if err != nil {
			return "", "", true, err
		}
		*right, *rightType = converted, ir.TypeString
	} else if (expression.Operator == "||" || expression.Operator == "&&") &&
		(*leftType == ir.TypeUnknown || *rightType == ir.TypeUnknown ||
			(*rightType == ir.TypeVoid && !isPointerLikeType(*leftType))) {
		if *leftType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*left}, Span: toIRSpan(path, expression.Span)})
			*left = boxed
			*leftType = ir.TypeUnknown
		}
		if *rightType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*right}, Span: toIRSpan(path, expression.Span)})
			*right = boxed
			*rightType = ir.TypeUnknown
		}
	} else if (expression.Operator == "||" || expression.Operator == "&&") && isPointerLikeType(*leftType) {
		if *rightType == "never[]" || *rightType == "unknown[]" {
			*rightType = *leftType
		} else if isPointerLikeType(*rightType) {
			if *leftType != *rightType && !isSubtype(string(*rightType), string(*leftType)) && !isSubtype(string(*leftType), string(*rightType)) {
				boxedLeft := nextTemp(counter)
				env[boxedLeft] = ir.TypeUnknown
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedLeft, Args: []string{*left}, Span: toIRSpan(path, expression.Span)})
				*left = boxedLeft
				*leftType = ir.TypeUnknown

				boxedRight := nextTemp(counter)
				env[boxedRight] = ir.TypeUnknown
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedRight, Args: []string{*right}, Span: toIRSpan(path, expression.Span)})
				*right = boxedRight
				*rightType = ir.TypeUnknown
			} else if isSubtype(string(*rightType), string(*leftType)) {
				*rightType = *leftType
			}
		} else if *rightType == ir.TypeBool {
			boolTemp := nextTemp(counter)
			nullConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: *leftType, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolTemp, Operator: "!=", Args: []string{*left, nullConst}, Span: toIRSpan(path, expression.Span)})
			*left = boolTemp
			*leftType = ir.TypeBool
		} else if *rightType == ir.TypeVoid {
			// Keep the logical result in the left operand's pointer-like
			// representation; the undefined sentinel is a valid falsy value.
			*rightType = *leftType
		} else {
			return "", "", true, fmt.Errorf("operator %q does not support %s and %s", expression.Operator, *leftType, *rightType)
		}
	} else if (expression.Operator == "||" || expression.Operator == "&&") && isPointerLikeType(*rightType) && *leftType == ir.TypeBool {
		boolTemp := nextTemp(counter)
		nullConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: *rightType, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolTemp, Operator: "!=", Args: []string{*right, nullConst}, Span: toIRSpan(path, expression.Span)})
		*right = boolTemp
		*rightType = ir.TypeBool
	} else {
		return "", "", true, fmt.Errorf("operator %q does not support %s and %s", expression.Operator, *leftType, *rightType)
	}

	return "", "", false, nil
}
