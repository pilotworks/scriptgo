package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerInstanceofExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	left, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	targetClass := callName(expression.Right)
	if targetClass == "" && expression.Right != nil && (expression.Right.Kind == "identifier" || expression.Right.Kind == "type") {
		targetClass = expression.Right.Text
	}
	if targetClass == "" {
		return "", "", fmt.Errorf("instanceof requires a class identifier on the right")
	}
	targetClass = classIdentityForPath(path, targetClass)
	if idx := strings.LastIndex(targetClass, "."); idx != -1 {
		if _, exists := classHierarchy[targetClass]; !exists {
			shortName := targetClass[idx+1:]
			if _, exists2 := classHierarchy[shortName]; exists2 {
				targetClass = shortName
			}
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpInstanceOf,
		Type:   ir.TypeBool,
		Result: result,
		Args:   []string{left},
		Value:  targetClass,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.TypeBool, nil
}

func tryLowerReferenceBinary(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, env map[string]ir.Type, counter *int, left *string, leftType *ir.Type, right *string, rightType ir.Type) (string, ir.Type, bool, error) {
	if isComparison(expression.Operator) {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: *result, Operator: expression.Operator, Args: []string{*left, *right}, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeBool, true, nil
	}
	if expression.Operator == "||" || expression.Operator == "&&" {
		if *leftType == ir.TypeUnknown && rightType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*right}, Span: toIRSpan(path, expression.Span)})
			*right = boxed
		} else if rightType == ir.TypeUnknown && *leftType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{*left}, Span: toIRSpan(path, expression.Span)})
			*left = boxed
			*leftType = ir.TypeUnknown
		}
		var cond string
		if *leftType == ir.TypeUnknown {
			condTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: condTemp,
				Callee: "__scriptgo.is_truthy",
				Args:   []string{*left},
				Span:   toIRSpan(path, expression.Span),
			})
			cond = condTemp
		} else {
			nullConst := nextTemp(counter)
			nullVal := "null"
			if *leftType == ir.TypeString {
				nullVal = ""
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: *leftType, Result: nullConst, Value: nullVal, Span: toIRSpan(path, expression.Span)})
			cmpNull := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{*left, nullConst}, Span: toIRSpan(path, expression.Span)})
			cond = cmpNull
			if *leftType == ir.TypeString {
				undefConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
				cmpUndef := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{*left, undefConst}, Span: toIRSpan(path, expression.Span)})
				condTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
				cond = condTemp
			} else if isPointerLikeType(*leftType) {
				undefConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: *leftType, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
				cmpUndef := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{*left, undefConst}, Span: toIRSpan(path, expression.Span)})
				condTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
				cond = condTemp
			}
		}
		if *result == "" {
			*result = nextTemp(counter)
		}
		selType := *leftType
		selLeft := *left
		selRight := *right
		if *leftType != rightType {
			if *leftType != ir.TypeUnknown {
				bLeft := nextTemp(counter)
				env[bLeft] = ir.TypeUnknown
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: bLeft, Args: []string{*left}, Span: toIRSpan(path, expression.Span)})
				selLeft = bLeft
			}
			if rightType != ir.TypeUnknown {
				bRight := nextTemp(counter)
				env[bRight] = ir.TypeUnknown
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: bRight, Args: []string{*right}, Span: toIRSpan(path, expression.Span)})
				selRight = bRight
			}
			selType = ir.TypeUnknown
		}
		env[*result] = selType
		if expression.Operator == "||" {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: selType, Result: *result, Args: []string{cond, selLeft, selRight}, Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: selType, Result: *result, Args: []string{cond, selRight, selLeft}, Span: toIRSpan(path, expression.Span)})
		}
		return *result, selType, true, nil
	}
	if *leftType == ir.TypeString && expression.Operator == "+" {
		// continue to binary +
	} else {
		return "", "", true, fmt.Errorf("operator %q does not support %s operands", expression.Operator, *leftType)
	}

	return "", "", false, nil
}
