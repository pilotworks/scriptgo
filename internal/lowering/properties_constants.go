package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerConstantPropertyPath(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, propertyPath []string) (string, ir.Type, bool, error) {
	firstIdent := propertyPath[0]
	if _, inEnv := env[firstIdent]; !inEnv {
		// A deeper path through a namespace (fs.constants.F_OK) reads the
		// exported binding and then its properties at run time.
		if len(propertyPath) > 2 {
			if binding, ok := moduleExportBinding(path, firstIdent, propertyPath[1]); ok {
				return lowerNamespaceMemberPath(path, expression, *result, function, env, counter, shapes, signatures, len(propertyPath)-2, binding)
			}
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
			if binding, ok := moduleExportBinding(path, firstIdent, propertyPath[1]); ok {
				decl := binding.Decl
				isPrimitiveConst := decl.VarDeclKind == "const" && decl.Expression != nil && (decl.Expression.Kind == "number" || decl.Expression.Kind == "string" || decl.Expression.Kind == "bool" || decl.Expression.Kind == "literal" || decl.Expression.Kind == "null" || decl.Expression.Kind == "undefined")
				if !isPrimitiveConst || function.Name != "main" {
					return binding.Storage, bindingType(binding), true, nil
				}
				value, typ, err := lowerExpression(path, decl.Expression, *result, function, env, counter, shapes, signatures)
				return value, typ, true, err
			}
		}
	}

	return "", "", false, nil
}

// lowerNamespaceMemberPath lowers a property chain whose innermost access is
// ns.binding, depth property accesses below expression, by replacing that
// access with the exported binding itself.
func lowerNamespaceMemberPath(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, depth int, binding topLevelBinding) (string, ir.Type, bool, error) {
	rewritten := *expression
	node := &rewritten
	// Descend like extractPropertyPath: `as` wrappers add no path segment.
	for depth > 0 || node.Kind == "as" {
		if node.Kind != "as" {
			depth--
		}
		inner := *node.Left
		node.Left = &inner
		node = node.Left
	}
	*node = frontend.SyntaxExpression{Span: node.Span, Kind: "identifier", Text: binding.Storage, InferredType: node.InferredType}
	value, typ, err := lowerExpression(path, &rewritten, result, function, env, counter, shapes, signatures)
	return value, typ, true, err
}
