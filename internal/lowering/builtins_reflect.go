package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerReflectIntrinsics(m map[string]BuiltinIntrinsic) {
	// --- Metadata Reflection APIs ---
	m["Reflect.getMetadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.getMetadata",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectGetMetadata,
	}

	m["Reflect.getOwnMetadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.getOwnMetadata",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectGetOwnMetadata,
	}

	m["Reflect.hasMetadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.hasMetadata",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectHasMetadata,
	}

	m["Reflect.hasOwnMetadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.hasOwnMetadata",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectHasOwnMetadata,
	}

	m["Reflect.defineMetadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.defineMetadata",
		MinArgs:  3,
		MaxArgs:  4,
		Lower:    lowerReflectDefineMetadata,
	}

	m["Reflect.metadata"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.metadata",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: result,
				Value:  "__metadata_decorator",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeString, nil
		},
	}

	// --- Standard ES6 Reflect Object Operations ---
	m["Reflect.get"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.get",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectGet,
	}

	m["Reflect.set"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.set",
		MinArgs:  3,
		MaxArgs:  4,
		Lower:    lowerReflectSet,
	}

	m["Reflect.has"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.has",
		MinArgs:  2,
		MaxArgs:  2,
		Lower:    lowerReflectHas,
	}

	m["Reflect.deleteProperty"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.deleteProperty",
		MinArgs:  2,
		MaxArgs:  2,
		Lower:    lowerReflectDeleteProperty,
	}

	m["Reflect.ownKeys"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.ownKeys",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerReflectOwnKeys,
	}

	m["Reflect.defineProperty"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.defineProperty",
		MinArgs:  3,
		MaxArgs:  3,
		Lower:    lowerReflectDefineProperty,
	}

	m["Reflect.getOwnPropertyDescriptor"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.getOwnPropertyDescriptor",
		MinArgs:  2,
		MaxArgs:  2,
		Lower:    lowerReflectGetOwnPropertyDescriptor,
	}

	m["Reflect.getPrototypeOf"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.getPrototypeOf",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerReflectGetPrototypeOf,
	}

	m["Reflect.setPrototypeOf"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.setPrototypeOf",
		MinArgs:  2,
		MaxArgs:  2,
		Lower:    lowerReflectSetPrototypeOf,
	}

	m["Reflect.isExtensible"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.isExtensible",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerReflectIsExtensible,
	}

	m["Reflect.preventExtensions"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.preventExtensions",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerReflectPreventExtensions,
	}

	m["Reflect.apply"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.apply",
		MinArgs:  3,
		MaxArgs:  3,
		Lower:    lowerReflectApply,
	}

	m["Reflect.construct"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Reflect.construct",
		MinArgs:  2,
		MaxArgs:  3,
		Lower:    lowerReflectConstruct,
	}
}

// --- Metadata Lowering Helpers ---

func extractTargetAndProperty(call IntrinsicCall) (string, string, string) {
	key := ""
	if len(call.Expression.Arguments) > 0 && call.Expression.Arguments[0] != nil {
		if call.Expression.Arguments[0].Kind == "string" {
			key = call.Expression.Arguments[0].Text
		}
	}

	target := ""
	if len(call.Expression.Arguments) > 1 && call.Expression.Arguments[1] != nil {
		arg1 := call.Expression.Arguments[1]
		if arg1.Kind == "identifier" {
			target = arg1.Text
		} else if arg1.Kind == "property" && arg1.Text == "prototype" && arg1.Left != nil {
			target = arg1.Left.Text
		} else {
			target = arg1.Text
		}
	}

	property := ""
	if len(call.Expression.Arguments) > 2 && call.Expression.Arguments[2] != nil {
		arg2 := call.Expression.Arguments[2]
		if arg2.Kind == "string" {
			property = arg2.Text
		} else if arg2.Kind == "identifier" {
			property = arg2.Text
		}
	}

	return target, property, key
}

func lowerReflectGetMetadata(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	target, property, key := extractTargetAndProperty(call)
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	val, found := lookupStaticMetadata(target, property, key)
	if !found && property != "" {
		// Fallback check on class-level if member specific wasn't found
		val, found = lookupStaticMetadata(target, "", key)
	}

	return emitStaticMetadataResult(call, result, val, found)
}

func lowerReflectGetOwnMetadata(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	target, property, key := extractTargetAndProperty(call)
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	val, found := lookupStaticMetadata(target, property, key)
	return emitStaticMetadataResult(call, result, val, found)
}

func emitStaticMetadataResult(call IntrinsicCall, result string, val interface{}, found bool) (string, ir.Type, error) {
	if found {
		switch v := val.(type) {
		case []string:
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:         ir.OpArray,
				Type:       ir.TypeStringArray,
				Result:     result,
				FieldCount: len(v),
				Span:       toIRSpan(call.Path, call.Expression.Span),
			})
			for _, item := range v {
				itemConst := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: itemConst,
					Value:  item,
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				pushRes := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: pushRes,
					Callee: "__array.push",
					Args:   []string{result, itemConst},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
			}
			return result, ir.TypeStringArray, nil
		case string:
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: result,
				Value:  v,
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeString, nil
		}
	}

	// If not found statically, return empty string / undefined
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: result,
		Value:  "",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeString, nil
}

func lowerReflectHasMetadata(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	target, property, key := extractTargetAndProperty(call)
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	found := hasStaticMetadata(target, property, key)
	if !found && property != "" {
		found = hasStaticMetadata(target, "", key)
	}

	valStr := "false"
	if found {
		valStr = "true"
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  valStr,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectHasOwnMetadata(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	target, property, key := extractTargetAndProperty(call)
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	found := hasStaticMetadata(target, property, key)
	valStr := "false"
	if found {
		valStr = "true"
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  valStr,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectDefineMetadata(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	key := ""
	if len(call.Expression.Arguments) > 0 && call.Expression.Arguments[0] != nil {
		key = call.Expression.Arguments[0].Text
	}
	val := ""
	if len(call.Expression.Arguments) > 1 && call.Expression.Arguments[1] != nil {
		val = call.Expression.Arguments[1].Text
	}
	target := ""
	if len(call.Expression.Arguments) > 2 && call.Expression.Arguments[2] != nil {
		arg2 := call.Expression.Arguments[2]
		if arg2.Kind == "identifier" {
			target = arg2.Text
		} else if arg2.Kind == "property" && arg2.Text == "prototype" && arg2.Left != nil {
			target = arg2.Left.Text
		} else {
			target = arg2.Text
		}
	}
	property := ""
	if len(call.Expression.Arguments) > 3 && call.Expression.Arguments[3] != nil {
		arg3 := call.Expression.Arguments[3]
		property = arg3.Text
	}

	if key != "" && target != "" {
		registerStaticMetadata(target, property, key, val)
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeVoid,
		Result: result,
		Value:  "",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeVoid, nil
}

// --- Standard ES6 Reflect Operation Implementations ---

func lowerReflectGet(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.get requires at least 2 arguments")
	}

	targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	propArg := call.Expression.Arguments[1]
	propName := propArg.Text

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	if after, ok := strings.CutPrefix(string(targetType), "object:"); ok {
		className := after
		if shape, exists := call.Shapes[className]; exists {
			for i, f := range shape.Fields {
				if f.Name == propName {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       f.Type,
						Result:     result,
						Args:       []string{targetVal},
						Field:      f.Name,
						FieldIndex: i,
						Span:       toIRSpan(call.Path, call.Expression.Span),
					})
					return result, f.Type, nil
				}
			}
		}
	}

	// Fallback to dynamic property get
	propVal, _, err := call.LowerExpression(call.Path, propArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeUnknown,
		Result: result,
		Callee: "__object.get_prop",
		Args:   []string{targetVal, propVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeUnknown, nil
}

func lowerReflectSet(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 3 {
		return "", "", fmt.Errorf("reflect.set requires at least 3 arguments")
	}

	targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	propArg := call.Expression.Arguments[1]
	propName := propArg.Text

	valArg := call.Expression.Arguments[2]
	valVal, _, err := call.LowerExpression(call.Path, valArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	if after, ok := strings.CutPrefix(string(targetType), "object:"); ok {
		className := after
		if shape, exists := call.Shapes[className]; exists {
			for i, f := range shape.Fields {
				if f.Name == propName {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:         ir.OpFieldSet,
						Type:       ir.TypeVoid,
						Field:      f.Name,
						FieldIndex: i,
						Args:       []string{targetVal, valVal},
						Span:       toIRSpan(call.Path, call.Expression.Span),
					})
					break
				}
			}
		}
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectHas(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.has requires 2 arguments")
	}

	targetVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	propArg := call.Expression.Arguments[1]
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	if propArg.Kind == "string" || propArg.Kind == "literal" {
		fieldName := strings.Trim(propArg.Text, "\"'`")
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpInstanceOf,
			Type:   ir.TypeBool,
			Result: result,
			Value:  fieldName,
			Args:   []string{targetVal},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return result, ir.TypeBool, nil
	}

	propVal, _, err := call.LowerExpression(call.Path, propArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpInstanceOf,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "",
		Args:   []string{targetVal, propVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectDeleteProperty(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectOwnKeys(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 1 {
		return "", "", fmt.Errorf("Reflect.ownKeys requires 1 argument")
	}

	targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Env[result] = ir.TypeStringArray

	if after, ok := strings.CutPrefix(string(targetType), "object:"); ok {
		className := after
		if shape, exists := call.Shapes[className]; exists {
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:         ir.OpArray,
				Type:       ir.TypeStringArray,
				Result:     result,
				FieldCount: len(shape.Fields),
				Span:       toIRSpan(call.Path, call.Expression.Span),
			})
			for _, f := range shape.Fields {
				constName := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: constName,
					Value:  f.Name,
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				pushRes := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: pushRes,
					Callee: "__array.push",
					Args:   []string{result, constName},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
			}
			return result, ir.TypeStringArray, nil
		}
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeStringArray,
		Result: result,
		Callee: "__object.keys",
		Args:   []string{targetVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeStringArray, nil
}

func lowerReflectDefineProperty(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) >= 3 {
		targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err == nil {
			propArg := call.Expression.Arguments[1]
			propName := propArg.Text
			descArg := call.Expression.Arguments[2]
			if descArg != nil && (descArg.Kind == "object" || descArg.Kind == "object_literal") {
				for _, prop := range descArg.Arguments {
					if prop.Text == "value" && prop.Left != nil {
						valVal, _, err := call.LowerExpression(call.Path, prop.Left, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
						if err == nil && strings.HasPrefix(string(targetType), "object:") {
							className := strings.TrimPrefix(string(targetType), "object:")
							if shape, exists := call.Shapes[className]; exists {
								for i, f := range shape.Fields {
									if f.Name == propName {
										call.Function.Body = append(call.Function.Body, ir.Instruction{
											Op:         ir.OpFieldSet,
											Type:       ir.TypeVoid,
											Field:      f.Name,
											FieldIndex: i,
											Args:       []string{targetVal, valVal},
											Span:       toIRSpan(call.Path, call.Expression.Span),
										})
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectGetOwnPropertyDescriptor(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	descShapeName := "PropertyDescriptor"
	if _, exists := call.Shapes[descShapeName]; !exists {
		call.Shapes[descShapeName] = ir.ObjectShape{
			Name: descShapeName,
			Fields: []ir.Field{
				{Name: "writable", Type: ir.TypeBool},
				{Name: "enumerable", Type: ir.TypeBool},
				{Name: "configurable", Type: ir.TypeBool},
			},
		}
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       ir.Type("object:" + descShapeName),
		Result:     result,
		FieldCount: 3,
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})

	writableConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: writableConst,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Field:      "writable",
		FieldIndex: 0,
		Args:       []string{result, writableConst},
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Field:      "enumerable",
		FieldIndex: 1,
		Args:       []string{result, writableConst},
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Field:      "configurable",
		FieldIndex: 2,
		Args:       []string{result, writableConst},
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})

	return result, ir.Type("object:" + descShapeName), nil
}

func lowerReflectGetPrototypeOf(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: result,
		Value:  "Object.prototype",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeString, nil
}

func lowerReflectSetPrototypeOf(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectIsExtensible(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectPreventExtensions(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectApply(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 3 {
		return "", "", fmt.Errorf("reflect.apply requires 3 arguments")
	}

	fnArg := call.Expression.Arguments[0]
	argsArg := call.Expression.Arguments[2]

	var directArgs []string
	if argsArg != nil && (argsArg.Kind == "array" || argsArg.Kind == "array_literal") {
		for _, el := range argsArg.Arguments {
			elVal, _, err := call.LowerExpression(call.Path, el, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			directArgs = append(directArgs, elVal)
		}
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	fnName := fnArg.Text
	if sig, exists := call.Signatures[fnName]; exists {
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   sig.ReturnType,
			Result: result,
			Callee: fnName,
			Args:   directArgs,
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return result, sig.ReturnType, nil
	}

	fnVal, fnType, err := call.LowerExpression(call.Path, fnArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeString,
		Result: result,
		Callee: fnVal,
		Args:   directArgs,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, fnType, nil
}

func lowerReflectConstruct(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.construct requires at least 2 arguments")
	}

	targetArg := call.Expression.Arguments[0]
	className := targetArg.Text

	shape, exists := call.Shapes[className]
	if !exists {
		return "", "", fmt.Errorf("reflect.construct: unknown class or constructor %q", className)
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       ir.Type("object:" + className),
		Result:     result,
		FieldCount: len(shape.Fields),
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})

	for i, field := range shape.Fields {
		defVal := field.Value
		if defVal == "" {
			switch field.Type {
			case ir.TypeNumber:
				defVal = "0"
			case ir.TypeBool:
				defVal = "false"
			case ir.TypeBigInt:
				defVal = "0"
			}
		}
		initTemp := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   field.Type,
			Result: initTemp,
			Value:  defVal,
			Span:   field.Span,
		})
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Field:      field.Name,
			FieldIndex: i,
			Args:       []string{result, initTemp},
			Span:       field.Span,
		})
	}

	return result, ir.Type("object:" + className), nil
}
