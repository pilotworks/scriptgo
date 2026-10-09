package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerObjectDefineProperty lowers Object.defineProperty(obj, key, { value })
// for an existing field of obj's shape as a field store and returns obj.
// Native objects have no property attributes, so writable/enumerable/
// configurable are not enforced; accessor descriptors (get/set), descriptors
// that are not object literals, and new properties are rejected.
func lowerObjectDefineProperty(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args := call.Expression.Arguments
	object, objectType, err := bindIntrinsicValue(call, args[0])
	if err != nil {
		return "", "", err
	}
	value, err := dataDescriptorValue(args[2])
	if err != nil {
		return "", "", err
	}
	if err := assignDefinedProperty(call, object, args[1], value); err != nil {
		return "", "", err
	}
	return object, objectType, nil
}

// lowerObjectDefineProperties lowers Object.defineProperties(obj, { k: { value } })
// as one assignment per property, in source order.
func lowerObjectDefineProperties(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args := call.Expression.Arguments
	object, objectType, err := bindIntrinsicValue(call, args[0])
	if err != nil {
		return "", "", err
	}
	if args[1] == nil || args[1].Kind != "object_literal" {
		return "", "", fmt.Errorf("Object.defineProperties requires an object literal of descriptors")
	}
	for _, property := range args[1].Arguments {
		if property.Kind != "property_assignment" {
			return "", "", fmt.Errorf("Object.defineProperties requires plain descriptor properties")
		}
		value, err := dataDescriptorValue(property.Left)
		if err != nil {
			return "", "", err
		}
		key := &frontend.SyntaxExpression{Span: property.Span, Kind: "string", Text: property.Text, InferredType: "string"}
		if err := assignDefinedProperty(call, object, key, value); err != nil {
			return "", "", err
		}
	}
	return object, objectType, nil
}

// lowerUnmodelledObjectReflection rejects reflection native objects cannot
// answer faithfully: a mutable prototype (setPrototypeOf) and the full
// descriptor map (getOwnPropertyDescriptors, whose declared result type has
// no fixed layout).
func lowerUnmodelledObjectReflection(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	return "", "", fmt.Errorf("%s is not supported in the native subset", intrinsic.Name)
}

// dataDescriptorValue returns the value of a data descriptor literal
// ({ value, writable, enumerable, configurable }).
func dataDescriptorValue(descriptor *frontend.SyntaxExpression) (*frontend.SyntaxExpression, error) {
	if descriptor == nil || descriptor.Kind != "object_literal" {
		return nil, fmt.Errorf("property descriptors must be object literals")
	}
	value := &frontend.SyntaxExpression{Span: descriptor.Span, Kind: "undefined", Text: "undefined", InferredType: "undefined"}
	for _, field := range descriptor.Arguments {
		if field.Kind != "property_assignment" {
			return nil, fmt.Errorf("property descriptors must be plain object literals")
		}
		switch field.Text {
		case "value":
			value = field.Left
		case "writable", "enumerable", "configurable":
		case "get", "set":
			return nil, fmt.Errorf("accessor property descriptors (get/set) are not supported")
		default:
			return nil, fmt.Errorf("unknown property descriptor field %q", field.Text)
		}
	}
	return value, nil
}

// assignDefinedProperty stores value in the field a literal key names on
// the object's fixed shape. Defining a property the shape does not have
// would change the shape, which native objects cannot do.
func assignDefinedProperty(call IntrinsicCall, object string, key, value *frontend.SyntaxExpression) error {
	shapeName := strings.TrimPrefix(string(call.Env[object]), "object:")
	shape, ok := lookupObjectShape(shapeName, call.Shapes)
	name, literal := literalPropertyKey(key)
	if !ok || !literal {
		return fmt.Errorf("defining a property is supported only for a literal key of an existing field")
	}
	for index, field := range shape.Fields {
		if field.Name != name {
			continue
		}
		stored, storedType, err := call.LowerExpression(call.Path, value, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return err
		}
		span := toIRSpan(call.Path, call.Expression.Span)
		if field.Type == ir.TypeUnknown && storedType != ir.TypeUnknown {
			boxed := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{stored}, Span: span})
			stored, storedType = boxed, ir.TypeUnknown
		}
		if storedType != field.Type {
			return fmt.Errorf("defining property %q: value of type %s does not fit field type %s", name, storedType, field.Type)
		}
		call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: shapeName, Field: name, FieldIndex: index, Args: []string{object, stored}, Span: span})
		return nil
	}
	return fmt.Errorf("defining new property %q would change a fixed object shape", name)
}

// literalPropertyKey is the property key of a string or numeric literal.
func literalPropertyKey(key *frontend.SyntaxExpression) (string, bool) {
	if key == nil {
		return "", false
	}
	if key.Kind == "string" {
		return key.Text, true
	}
	return numericLiteralKey(key)
}

// bindIntrinsicValue lowers an argument once and records its type so later
// synthetic expressions can refer to the temporary by name.
func bindIntrinsicValue(call IntrinsicCall, argument *frontend.SyntaxExpression) (string, ir.Type, error) {
	value, typ, err := call.LowerExpression(call.Path, argument, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if _, known := call.Env[value]; err == nil && !known {
		call.Env[value] = typ
	}
	return value, typ, err
}

// lowerGetOwnPropertyDescriptor returns the data descriptor of a field a
// literal key names on a known shape (native fields are writable,
// enumerable, and configurable data properties), or undefined when the
// shape has no such field.
func lowerGetOwnPropertyDescriptor(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args := call.Expression.Arguments
	object, objectType, err := bindIntrinsicValue(call, args[0])
	if err != nil {
		return "", "", err
	}
	shape, ok := lookupObjectShape(strings.TrimPrefix(string(objectType), "object:"), call.Shapes)
	name, literal := literalPropertyKey(args[1])
	if !ok || !literal {
		return "", "", fmt.Errorf("Object.getOwnPropertyDescriptor is supported only for a literal key on a known object shape")
	}
	for _, field := range shape.Fields {
		if field.Name == name {
			return call.LowerExpression(call.Path, dataDescriptorLiteral(call.Expression.Span, object, objectType, name), call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		}
	}
	undefined := &frontend.SyntaxExpression{Span: call.Expression.Span, Kind: "undefined", Text: "undefined", InferredType: "undefined"}
	return call.LowerExpression(call.Path, undefined, call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
}

// dataDescriptorLiteral is { value: object[name], writable: true,
// enumerable: true, configurable: true }.
func dataDescriptorLiteral(span frontend.SourceSpan, object string, objectType ir.Type, name string) *frontend.SyntaxExpression {
	read := &frontend.SyntaxExpression{Span: span, Kind: "property", Text: name, Left: &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: object, InferredType: string(objectType)}}
	flag := func(key string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "property_assignment", Text: key, Left: &frontend.SyntaxExpression{Span: span, Kind: "bool", Text: "true", InferredType: "bool"}}
	}
	return &frontend.SyntaxExpression{Span: span, Kind: "object_literal", Arguments: []*frontend.SyntaxExpression{
		{Span: span, Kind: "property_assignment", Text: "value", Left: read},
		flag("writable"), flag("enumerable"), flag("configurable"),
	}}
}
