package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerAssignmentExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if expression.Left != nil && expression.Left.Kind == "identifier" {
		varName := expression.Left.Text
		varType, ok := env[varName]
		if !ok {
			if topVar, isTop := topLevelVars[varName]; isTop {
				varType = toIRType(topVar.Type)
				if varType == "" {
					varType = toIRType(topVar.InferredType)
				}
				if varType == "" {
					varType = ir.TypeNumber
				}
				ok = true
			}
		}
		if !ok {
			return "", "", true, fmt.Errorf("assignment to unknown variable %q", varName)
		}
		rhsExpr := expression.Right
		if expression.Operator != "=" {
			baseOp := strings.TrimSuffix(expression.Operator, "=")
			rhsExpr = &frontend.SyntaxExpression{
				Span:         expression.Span,
				Kind:         "binary",
				Operator:     baseOp,
				Left:         expression.Left,
				Right:        expression.Right,
				InferredType: expression.InferredType,
			}
		}
		val, valType, err := lowerExpression(path, rhsExpr, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		if varType == ir.TypeUnknown && valType != ir.TypeUnknown {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: varName,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			if result != "" {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   varType,
					Result: result,
					Args:   []string{varName},
					Span:   toIRSpan(path, expression.Span),
				})
			}
			return result, varType, true, nil
		}
		if valType != varType && varType != ir.TypeUnknown {
			if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(varType), "object:") || varType == ir.TypeObject) {
				// Polymorphic object assignment
			} else if (varType == ir.TypeUint8Array || varType == ir.TypeBuffer) && (valType == ir.TypeUint8Array || valType == ir.TypeBuffer) {
				// Buffer extends Uint8Array and Uint8Array is binary-compatible with Buffer
			} else if isPointerLikeType(varType) && (valType == ir.TypeVoid || valType == ir.TypePointer) {
				// Assigning undefined or null to a pointer-like variable
			} else {
				return "", "", true, fmt.Errorf("assignment type mismatch for %q: %s := %s", varName, varType, valType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   varType,
			Result: varName,
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		if result != "" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: result,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, varType, true, nil
		}
		return val, varType, true, nil
	}

	return "", "", false, nil
}
