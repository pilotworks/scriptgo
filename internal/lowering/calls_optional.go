package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerOptionalCallExpression lowers optional call expressions: fn?.(args) or obj?.method?.(args)
func lowerOptionalCallExpression(
	path string,
	expression *typescriptgo.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if expression.Left == nil {
		return "", "", fmt.Errorf("optional call missing target")
	}

	var calleeVal string
	var calleeType ir.Type
	var cond string
	var standardCall *typescriptgo.SyntaxExpression

	fnRetType := ""
	if expression.InferredType != "" {
		parts := strings.Split(expression.InferredType, "|")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "undefined" && trimmed != "null" && trimmed != "void" && trimmed != "" {
				fnRetType = trimmed
				break
			}
		}
	}

	isClassMethod := false
	if (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil {
		receiverExpr := expression.Left.Left
		receiverVal, receiverType, err := lowerExpression(path, receiverExpr, "", function, env, counter, shapes, signatures)
		if err == nil {
			className := strings.TrimPrefix(string(receiverType), "object:")
			if className != "" {
				if _, _, ok := findMethodInHierarchy(className, expression.Left.Text, signatures, classHierarchy); ok {
					isClassMethod = true
					cond, err = coerceToBool(path, receiverVal, receiverType, function, counter, expression.Span)
					if err != nil {
						return "", "", err
					}
					calleeVal = receiverVal
					standardCall = &typescriptgo.SyntaxExpression{
						Span: expression.Span,
						Kind: "call",
						Left: &typescriptgo.SyntaxExpression{
							Span: expression.Left.Span,
							Kind: "property",
							Left: &typescriptgo.SyntaxExpression{
								Span:         receiverExpr.Span,
								Kind:         "identifier",
								Text:         receiverVal,
								InferredType: string(receiverType),
							},
							Text:  expression.Left.Text,
							Right: expression.Left.Right,
						},
						Arguments:     expression.Arguments,
						TypeArguments: expression.TypeArguments,
						InferredType:  fnRetType,
					}
				}
			}
		}
	}

	if !isClassMethod {
		var err error
		calleeVal, calleeType, err = lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if (calleeType == "" || calleeType == ir.TypeUnknown) && expression.Left.InferredType != "" {
			if t := toIRType(expression.Left.InferredType); t != "" {
				calleeType = t
			}
		}
		env[calleeVal] = calleeType
		cond, err = coerceToBool(path, calleeVal, calleeType, function, counter, expression.Span)
		if err != nil {
			return "", "", err
		}
		standardCall = &typescriptgo.SyntaxExpression{
			Span:          expression.Span,
			Kind:          "call",
			Left:          &typescriptgo.SyntaxExpression{Span: expression.Span, Kind: "identifier", Text: calleeVal, InferredType: string(calleeType)},
			Arguments:     expression.Arguments,
			TypeArguments: expression.TypeArguments,
			InferredType:  fnRetType,
		}
	}

	if result == "" {
		result = nextTemp(counter)
	}

	retType := ir.TypeUnknown
	if expression.InferredType != "" {
		if t := toIRType(expression.InferredType); t != "" {
			retType = t
		}
	} else if rt, ok := env[calleeVal+".retType"]; ok && rt != "" {
		retType = rt
	} else if target, ok := signatures[calleeVal]; ok && target.ReturnType != "" {
		retType = target.ReturnType
	}
	if retType == ir.TypeNumber || retType == ir.TypeBool {
		retType = ir.TypeUnknown
	}

	initVal := "undefined"
	if retType != ir.TypeString && retType != ir.TypeUnknown {
		initVal = "null"
	}

	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   retType,
		Result: result,
		Value:  initVal,
		Span:   toIRSpan(path, expression.Span),
	})

	thenFn := &ir.Function{}

	callRes, callType, err := lowerCallExpression(path, standardCall, "", thenFn, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if retType == ir.TypeUnknown && callType != ir.TypeUnknown {
		thenFn.Body = append(thenFn.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: result,
			Args:   []string{callRes},
			Span:   toIRSpan(path, expression.Span),
		})
	} else {
		thenFn.Body = append(thenFn.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   callType,
			Result: result,
			Args:   []string{callRes},
			Span:   toIRSpan(path, expression.Span),
		})
	}

	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{cond},
		Then: thenFn.Body,
		Span: toIRSpan(path, expression.Span),
	})

	return result, retType, nil
}
