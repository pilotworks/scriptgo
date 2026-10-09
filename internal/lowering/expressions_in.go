package lowering

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerObjectHasProperty lowers `key in object` for an object value. The
// runtime answers for own properties; methods of the object's class (and its
// bases) live on the prototype, so a key naming one is also present.
func lowerObjectHasProperty(path string, expression *frontend.SyntaxExpression, object string, className string, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	span := toIRSpan(path, expression.Span)
	methods := classMethodNames(classIdentityForPath(path, className))
	if expression.Left != nil && (expression.Left.Kind == "string" || expression.Left.Kind == "literal") {
		name := strings.Trim(expression.Left.Text, "\"'`")
		if slices.Contains(methods, name) {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: result, Value: "true", Span: span})
			return result, ir.TypeBool, nil
		}
		emitHasOwnProperty(function, counter, span, result, object, name, true)
		return result, ir.TypeBool, nil
	}
	key, keyType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if key, err = propertyKeyString(function, counter, span, key, keyType); err != nil {
		return "", "", err
	}
	present := result
	if len(methods) > 0 {
		present = nextTemp(counter)
	}
	emitHasOwnProperty(function, counter, span, present, object, key, false)
	for index, method := range methods {
		name := nextTemp(counter)
		matches := nextTemp(counter)
		combined := nextTemp(counter)
		if index == len(methods)-1 {
			combined = result
		}
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: name, Value: method, Span: span},
			ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: matches, Operator: "==", Args: []string{key, name}, Span: span},
			ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: combined, Operator: "||", Args: []string{present, matches}, Span: span},
		)
		present = combined
	}
	return result, ir.TypeBool, nil
}

// classMethodNames lists the instance methods and accessors of a class and
// its bases, in declaration order from the class up.
func classMethodNames(className string) []string {
	var names []string
	for seen := map[string]bool{}; className != "" && !seen[className]; className = classHierarchy[className].Extends {
		seen[className] = true
		syntax, ok := classSyntax[className]
		if !ok {
			break
		}
		for _, method := range syntax.Methods {
			if !method.IsStatic && method.Name != "constructor" && !slices.Contains(names, method.Name) {
				names = append(names, method.Name)
			}
		}
	}
	return names
}

// emitHasOwnProperty tests whether object has the own property key (a
// string value, or with keyIsLiteral a key name) at run time.
func emitHasOwnProperty(function *ir.Function, counter *int, span ir.SourceSpan, result string, object string, key string, keyIsLiteral bool) {
	if keyIsLiteral {
		name := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: name, Value: key, Span: span})
		key = name
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__object.has_own", Args: []string{object, key}, Span: span})
}

// propertyKeyString converts a key value to the property-key string the
// runtime looks up (ToPropertyKey for primitives).
func propertyKeyString(function *ir.Function, counter *int, span ir.SourceSpan, key string, keyType ir.Type) (string, error) {
	callee := ""
	switch keyType {
	case ir.TypeString:
		return key, nil
	case ir.TypeNumber:
		callee = "__string.fromNumber"
	case ir.TypeBool:
		callee = "__string.fromBool"
	case ir.TypeUnknown:
		callee = "__string.fromUnknown"
	default:
		return "", fmt.Errorf("property key of type %s is not supported by this check in the native subset", keyType)
	}
	converted := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: converted, Callee: callee, Args: []string{key}, Span: span})
	return converted, nil
}
