package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerInheritedObjectMethod lowers toString, toLocaleString and valueOf
// that a receiver inherits from Object.prototype (or its built-in class):
// valueOf returns the object, toString its "[object Tag]". An object that
// defines the method as its own property is not handled here, so the call
// reaches that function.
func lowerInheritedObjectMethod(path string, expression *frontend.SyntaxExpression, receiver string, receiverType ir.Type, className string, methodName string, result string, function *ir.Function, counter *int, shapes map[string]ir.ObjectShape) (string, ir.Type, bool) {
	if methodName != "toString" && methodName != "toLocaleString" && methodName != "valueOf" {
		return "", "", false
	}
	if shape, ok := lookupObjectShape(className, shapes); ok && fieldIndex(shape, methodName) >= 0 {
		return "", "", false
	}
	if methodName == "valueOf" {
		return receiver, receiverType, true
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: result, Value: "[object " + objectToStringTag(receiverType) + "]", Span: toIRSpan(path, expression.Span)})
	return result, ir.TypeString, true
}

// objectToStringTag is the tag Object.prototype.toString reports for a
// receiver of a built-in type.
func objectToStringTag(typ ir.Type) string {
	switch {
	case typ == ir.TypeMap:
		return "Map"
	case typ == ir.TypeSet:
		return "Set"
	case typ == ir.TypeClosure:
		return "Function"
	case isPromiseReceiverType(typ):
		return "Promise"
	}
	return "Object"
}

// lowerClassToString converts an instance of a class that defines toString
// (directly or inherited) by calling that method, as ToString does.
func lowerClassToString(path string, value string, valueType ir.Type, span frontend.SourceSpan, function *ir.Function, counter *int, signatures map[string]ir.Function) (string, bool) {
	className, isObject := strings.CutPrefix(string(valueType), "object:")
	if !isObject {
		return "", false
	}
	method, mangled, found := findMethodInHierarchy(className, "toString", signatures, classHierarchy)
	if !found || (method.ReturnType != ir.TypeString && method.ReturnType != "") {
		return "", false
	}
	result := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: mangled, Args: []string{value}, Span: toIRSpan(path, span)})
	return result, true
}

// lowerStringConversion lowers String(value): a class instance converts
// through its toString method, a symbol to "Symbol(desc)", everything else
// through the runtime.
func lowerStringConversion(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	var args []string
	if len(call.Expression.Arguments) == 1 {
		value, valueType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		if converted, ok := lowerClassToString(call.Path, value, valueType, call.Expression.Span, call.Function, call.Counter, call.Signatures); ok {
			return converted, ir.TypeString, nil
		}
		if valueType == ir.TypeSymbol {
			// String(symbol) is its descriptive string, "Symbol(desc)".
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__symbol.toString", Args: []string{value}, Span: toIRSpan(call.Path, call.Expression.Span)})
			return result, ir.TypeString, nil
		}
		args = []string{value}
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.new", Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeString, nil
}
