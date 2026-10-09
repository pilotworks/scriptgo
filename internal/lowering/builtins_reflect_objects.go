package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// The Reflect object operations are lowered through the matching Object or
// Function operations, so they share one implementation: getPrototypeOf,
// setPrototypeOf, isExtensible and getOwnPropertyDescriptor are the Object
// functions; preventExtensions and defineProperty report success as a
// boolean; apply is Function.prototype.apply; construct is new.

func reflectSyntax(call IntrinsicCall) syntaxBuilder {
	return syntaxBuilder{span: call.Expression.Span}
}

// syntaxBuilder builds synthetic expressions at one source span.
type syntaxBuilder struct {
	span frontend.SourceSpan
}

func (b syntaxBuilder) identifier(name string, inferred string) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: b.span, Kind: "identifier", Text: name, InferredType: inferred}
}

func (b syntaxBuilder) boolean(value bool) *frontend.SyntaxExpression {
	text := "false"
	if value {
		text = "true"
	}
	return &frontend.SyntaxExpression{Span: b.span, Kind: "bool", Text: text, InferredType: "boolean"}
}

func (b syntaxBuilder) objectCall(method string, inferred string, args ...*frontend.SyntaxExpression) *frontend.SyntaxExpression {
	return b.methodCall(b.identifier("Object", "ObjectConstructor"), method, inferred, args...)
}

func (b syntaxBuilder) methodCall(receiver *frontend.SyntaxExpression, method string, inferred string, args ...*frontend.SyntaxExpression) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: b.span, Kind: "call", InferredType: inferred, Left: &frontend.SyntaxExpression{Span: b.span, Kind: "property", Text: method, Left: receiver}, Arguments: args}
}

// then evaluates first for its effects and yields second (first, second).
func (b syntaxBuilder) then(first, second *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: b.span, Kind: "binary", Operator: ",", Left: first, Right: second, InferredType: second.InferredType}
}

func lowerReflectSyntax(call IntrinsicCall, expression *frontend.SyntaxExpression) (string, ir.Type, error) {
	return call.LowerExpression(call.Path, expression, call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
}

// reflectAsObject lowers Reflect.<name>(...) as Object.<name>(...).
func reflectAsObject(method string) func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error) {
	return func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
		return lowerReflectSyntax(call, reflectSyntax(call).objectCall(method, call.Expression.InferredType, call.Expression.Arguments...))
	}
}

var (
	lowerReflectGetPrototypeOf           = reflectAsObject("getPrototypeOf")
	lowerReflectSetPrototypeOf           = reflectAsObject("setPrototypeOf")
	lowerReflectIsExtensible             = reflectAsObject("isExtensible")
	lowerReflectGetOwnPropertyDescriptor = reflectAsObject("getOwnPropertyDescriptor")
)

// boundTarget evaluates the target once and refers to it by its temporary.
func boundTarget(call IntrinsicCall) (*frontend.SyntaxExpression, error) {
	target := call.Expression.Arguments[0]
	value, _, err := bindIntrinsicValue(call, target)
	if err != nil {
		return nil, err
	}
	return reflectSyntax(call).identifier(value, target.InferredType), nil
}

// Reflect.preventExtensions(o): Object.preventExtensions(o), then true.
func lowerReflectPreventExtensions(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	b := reflectSyntax(call)
	return lowerReflectSyntax(call, b.then(b.objectCall("preventExtensions", call.Expression.Arguments[0].InferredType, call.Expression.Arguments...), b.boolean(true)))
}

// Reflect.defineProperty(o, k, d): false for a frozen target, else
// Object.defineProperty(o, k, d) and true.
func lowerReflectDefineProperty(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) != 3 {
		return "", "", fmt.Errorf("Reflect.defineProperty requires a target, key, and descriptor")
	}
	target, err := boundTarget(call)
	if err != nil {
		return "", "", err
	}
	b := reflectSyntax(call)
	define := b.objectCall("defineProperty", target.InferredType, target, call.Expression.Arguments[1], call.Expression.Arguments[2])
	expression := &frontend.SyntaxExpression{Span: b.span, Kind: "conditional", InferredType: "boolean",
		Left:      b.objectCall("isFrozen", "boolean", target),
		WhenTrue:  b.boolean(false),
		WhenFalse: b.then(define, b.boolean(true)),
	}
	return lowerReflectSyntax(call, expression)
}

// Reflect.apply(f, thisArg, args) is f.apply(thisArg, args).
func lowerReflectApply(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args := call.Expression.Arguments
	if len(args) != 3 {
		return "", "", fmt.Errorf("Reflect.apply requires a function, a this value, and an arguments array")
	}
	return lowerReflectSyntax(call, reflectSyntax(call).methodCall(args[0], "apply", call.Expression.InferredType, args[1], args[2]))
}

// Reflect.construct(C, [a, b]) is new C(a, b).
func lowerReflectConstruct(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args := call.Expression.Arguments
	if len(args) != 2 {
		return "", "", fmt.Errorf("Reflect.construct with a newTarget is not supported in the native subset")
	}
	if args[1].Kind != "array" {
		return "", "", fmt.Errorf("Reflect.construct requires an array literal of arguments in the native subset")
	}
	construct := &frontend.SyntaxExpression{Span: call.Expression.Span, Kind: "new", InferredType: call.Expression.InferredType, Left: args[0], Arguments: args[1].Arguments}
	return lowerReflectSyntax(call, construct)
}

// Reflect.set(o, k, v) and Reflect.deleteProperty(o, k) run in the runtime,
// which reports false where the target's integrity level forbids the change
// (the operators throw in strict code instead).
func lowerReflectSet(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) != 3 {
		return "", "", fmt.Errorf("Reflect.set with a receiver is not supported in the native subset")
	}
	return lowerReflectMutation(call, "__object.reflect_set")
}

func lowerReflectDeleteProperty(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	return lowerReflectMutation(call, "__object.reflect_delete")
}

func lowerReflectMutation(call IntrinsicCall, callee string) (string, ir.Type, error) {
	args := call.Expression.Arguments
	span := toIRSpan(call.Path, call.Expression.Span)
	target, _, err := call.LowerExpression(call.Path, args[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	key, keyType, err := call.LowerExpression(call.Path, args[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	if keyType != ir.TypeString {
		return "", "", fmt.Errorf("%s requires a string key in the native subset, got %s", call.Expression.Left.Text, args[1].InferredType)
	}
	operands := []string{target, key}
	if len(args) == 3 {
		value, valueType, err := call.LowerExpression(call.Path, args[2], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		if valueType != ir.TypeUnknown {
			boxed := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{value}, Span: span})
			value = boxed
		}
		operands = append(operands, value)
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: callee, Args: operands, Span: span})
	return result, ir.TypeBool, nil
}
