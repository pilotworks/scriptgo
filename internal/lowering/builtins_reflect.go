package lowering

import (
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
