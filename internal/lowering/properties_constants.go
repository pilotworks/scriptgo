package lowering

import (
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerConstantPropertyPath(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, propertyPath []string) (string, ir.Type, bool, error) {
	firstIdent := propertyPath[0]
	if _, inEnv := env[firstIdent]; !inEnv {
		if val, valType, ok := resolveASTConstantPath(propertyPath); ok {
			if *result == "" {
				*result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   valType,
				Result: *result,
				Value:  val,
				Span:   toIRSpan(path, expression.Span),
			})
			return *result, valType, true, nil
		}
		if len(propertyPath) == 2 {
			if sig, ok := resolveFunctionSignature(path, callName(expression), signatures); ok {
				if *result == "" {
					*result = nextTemp(counter)
				}
				// A static method or namespace function used as a value is
				// called through the closure ABI, so it needs the trampoline
				// that unboxes closure arguments into its typed parameters.
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpClosure,
					Type:   ir.TypeClosure,
					Result: *result,
					Callee: ensureFunctionClosureTrampoline(path, sig, shapes, signatures),
					Span:   toIRSpan(path, expression.Span),
				})
				return *result, ir.TypeClosure, true, nil
			}
			cleanPath := filepath.Clean(path)
			_, isNS := functionNamespacesByFile[cleanPath][firstIdent]
			if !isNS {
				_, isNS = classNamespacesByFile[cleanPath][firstIdent]
			}
			if isNS {
				prop := propertyPath[1]
				if topVar, hasVar := topLevelVars[prop]; hasVar {
					varTyp := toIRType(topVar.Type)
					if varTyp == "" {
						varTyp = toIRType(topVar.InferredType)
					}
					if varTyp == "" {
						varTyp = ir.TypeNumber
					}
					isPrimitiveConst := topVar.VarDeclKind == "const" && topVar.Expression != nil && (topVar.Expression.Kind == "number" || topVar.Expression.Kind == "string" || topVar.Expression.Kind == "bool" || topVar.Expression.Kind == "literal" || topVar.Expression.Kind == "null" || topVar.Expression.Kind == "undefined")
					if !isPrimitiveConst || function.Name != "main" {
						return prop, varTyp, true, nil
					}
					{
						v0, v1, v2 := lowerExpression(path, topVar.Expression, *result, function, env, counter, shapes, signatures)
						return v0, v1, true, v2
					}
				}
			}
		}
	}

	return "", "", false, nil
}
