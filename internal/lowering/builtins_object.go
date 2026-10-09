package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectIntrinsics(m map[string]BuiltinIntrinsic) {
	m["Object.is"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.is",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			v1, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			v2, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__object.is",
				Args:   []string{v1, v2},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}

	m["Object.hasOwn"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.hasOwn",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, objType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}

			// A class layout answers statically for a required field (always
			// present) or a key it lacks; objects addressed by name and
			// optional fields are checked at run time.
			if after, ok := strings.CutPrefix(string(objType), "object:"); ok && !dynamicFieldAccess(after) {
				className := after
				shape, exists := call.Shapes[className]
				propName := ""
				if exists && call.Expression.Arguments[1] != nil && call.Expression.Arguments[1].Kind == "string" {
					propName = call.Expression.Arguments[1].Text
				}
				if index := fieldIndex(shape, propName); propName != "" && (index < 0 || !shape.Fields[index].Optional) {
					hasProp := index >= 0
					valStr := "false"
					if hasProp {
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
			}

			propVal, propType, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			span := toIRSpan(call.Path, call.Expression.Span)
			if propVal, err = propertyKeyString(call.Function, call.Counter, span, propVal, propType); err != nil {
				return "", "", err
			}
			emitHasOwnProperty(call.Function, call.Counter, span, result, objVal, propVal, false)
			return result, ir.TypeBool, nil
		},
	}

	registerObjectEnumerationIntrinsics(m)

	registerObjectAssignIntrinsic(m)

	registerObjectFromEntriesIntrinsic(m)

	m["Object.groupBy"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.groupBy",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			itemsVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			cbVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Record"),
				Result: result,
				Callee: "__object.groupBy",
				Args:   []string{itemsVal, cbVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("object:Record"), nil
		},
	}

	registerSimpleObj := func(names []string, callee string, retType ir.Type, minArgs, maxArgs int) {
		for _, name := range names {
			n := name
			m[n] = BuiltinIntrinsic{
				Category: CategoryECMAScript,
				Name:     n,
				MinArgs:  minArgs,
				MaxArgs:  maxArgs,
				Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
					var args []string
					rType := retType
					for i, arg := range call.Expression.Arguments {
						val, typ, err := call.LowerExpression(call.Path, arg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
						if err != nil {
							return "", "", err
						}
						if i == 0 && retType == "" {
							rType = typ
						}
						args = append(args, val)
					}
					result := call.Result
					if result == "" {
						result = nextTemp(call.Counter)
					}
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   rType,
						Result: result,
						Callee: callee,
						Args:   args,
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					return result, rType, nil
				},
			}
		}
	}

	registerCustomIntrinsic(m, []string{"Object.create"}, CategoryECMAScript, "", ir.TypeObject, 1, 2, lowerObjectCreate)
	registerCustomIntrinsic(m, []string{"Object.defineProperty"}, CategoryECMAScript, "", ir.TypeObject, 3, 3, lowerObjectDefineProperty)
	registerCustomIntrinsic(m, []string{"Object.defineProperties"}, CategoryECMAScript, "", ir.TypeObject, 2, 2, lowerObjectDefineProperties)
	registerCustomIntrinsic(m, []string{"Object.getOwnPropertyDescriptor"}, CategoryECMAScript, "", ir.TypeObject, 2, 2, lowerGetOwnPropertyDescriptor)
	registerCustomIntrinsic(m, []string{"Object.getOwnPropertyDescriptors", "Object.setPrototypeOf"}, CategoryECMAScript, "", ir.TypeObject, 1, 2, lowerUnmodelledObjectReflection)
	registerIntegrityLevel(m, "Object.freeze", "__object.freeze")
	registerIntegrityLevel(m, "Object.seal", "__object.seal")
	registerIntegrityLevel(m, "Object.preventExtensions", "__object.preventExtensions")
	registerSimpleObj([]string{"Object.isFrozen"}, "__object.isFrozen", ir.TypeBool, 1, 1)
	registerSimpleObj([]string{"Object.isSealed"}, "__object.isSealed", ir.TypeBool, 1, 1)
	registerSimpleObj([]string{"Object.isExtensible"}, "__object.isExtensible", ir.TypeBool, 1, 1)
	m["Object.getOwnPropertyNames"] = m["Object.keys"]
	m["Object.getOwnPropertySymbols"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.getOwnPropertySymbols",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			_, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:         ir.OpArray,
				Type:       ir.Type("symbol[]"),
				Result:     result,
				FieldCount: 0,
				Span:       toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("symbol[]"), nil
		},
	}
	registerSimpleObj([]string{"Object.getPrototypeOf"}, "__object.getPrototypeOf", ir.TypeObject, 1, 1)
}

// registerIntegrityLevel lowers Object.freeze/seal/preventExtensions, which
// return their argument: an object or array keeps its type and has its
// integrity level raised at run time.
func registerIntegrityLevel(m map[string]BuiltinIntrinsic, name string, callee string) {
	m[name] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     name,
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			value, typ, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			if isTypedArrayType(typ) {
				return "", "", fmt.Errorf("%s of a typed array is not supported in the native subset (JavaScript throws for one with elements)", name)
			}
			if typ != ir.TypeObject && !strings.HasPrefix(string(typ), "object:") && !strings.HasSuffix(string(typ), "[]") {
				// Primitives, and built-ins whose state is not modelled as
				// own properties (Map, Set, functions), are returned as is.
				return value, typ, nil
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: typ, Result: result, Callee: callee, Args: []string{value}, Span: toIRSpan(call.Path, call.Expression.Span)})
			return result, typ, nil
		},
	}
}
