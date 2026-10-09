package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerIdentifierExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Text == "undefined" {
		if _, inEnv := env["undefined"]; !inEnv {
			typ := ir.TypeVoid
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "undefined", Span: toIRSpan(path, expression.Span)})
			return result, typ, nil
		}
	}
	identName := expression.Text
	if mangled, hasMangled := env["__ident."+expression.Text]; hasMangled {
		identName = string(mangled)
	}
	rawEnvType, ok := env[identName]
	if !ok {
		rawEnvType, ok = env[expression.Text]
	}
	if ok {
		typ := rawEnvType
		// TypeScript-Go may preserve a const string literal as the
		// variable's environment type. Normalize it before property and
		// operator lowering so literal values use their runtime primitive.
		if normalized := toIRType(string(typ)); normalized != "" && normalized != ir.TypeUnknown {
			typ = normalized
		}
		storageType := env["__storage_type."+identName]
		if storageType == "" {
			storageType = env["__storage_type."+expression.Text]
		}
		if storageType == "" {
			if topVar, ok := topLevelVars[expression.Text]; ok {
				vType := topVar.Type
				if vType == "" && topVar.InferredType != "" {
					vType = topVar.InferredType
				}
				storageType = toIRType(vType)
			}
		}
		origParamType, isParam := env["__param."+identName]
		if !isParam {
			origParamType, isParam = env["__param."+expression.Text]
		}
		isOriginallyUnknown := (rawEnvType == "" || (isParam && (origParamType == ir.TypeUnknown || origParamType == "" || strings.Contains(string(origParamType), "|"))))
		if isOriginallyUnknown {
			switch expression.InferredType {
			case "number":
				typ = ir.TypeNumber
			case "bigint":
				typ = ir.TypeBigInt
			case "symbol":
				typ = ir.TypeSymbol
			case "string":
				typ = ir.TypeString
			case "bool", "boolean":
				typ = ir.TypeBool
			case "number[]":
				typ = ir.TypeNumberArray
			case "string[]":
				typ = ir.TypeStringArray
			case "unknown[]":
				typ = ir.TypeUnknownArray
			case "Uint8Array":
				typ = ir.TypeUint8Array
			case "Int8Array":
				typ = ir.TypeInt8Array
			case "Uint16Array":
				typ = ir.TypeUint16Array
			case "Int16Array":
				typ = ir.TypeInt16Array
			case "Uint32Array":
				typ = ir.TypeUint32Array
			case "Int32Array":
				typ = ir.TypeInt32Array
			case "Float32Array":
				typ = ir.TypeFloat32Array
			case "Float64Array":
				typ = ir.TypeFloat64Array
			case "BigInt64Array":
				typ = ir.TypeBigInt64Array
			case "BigUint64Array":
				typ = ir.TypeBigUint64Array
			case "ArrayBuffer":
				typ = ir.TypeArrayBuffer
			case "DataView":
				typ = ir.TypeDataView
			default:
				if strings.HasPrefix(expression.InferredType, "object:") {
					typ = ir.Type(expression.InferredType)
				} else if _, isShape := shapes[expression.InferredType]; isShape {
					typ = ir.Type("object:" + expression.InferredType)
				}
			}
		}
		isNarrowedParameter := isParam && origParamType != "" && typ != "" && typ != ir.TypeUnknown && origParamType != typ
		if (storageType == ir.TypeUnknown && rawEnvType != ir.TypeUnknown && rawEnvType != "") || (isOriginallyUnknown && typ != ir.TypeUnknown && typ != "") || isNarrowedParameter {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   typ,
				Result: result,
				Args:   []string{identName},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, typ, nil
		}
		if result != "" && result != identName {
			if typ == ir.TypeNumber {
				zeroConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "-", Args: []string{identName, zeroConst}, Span: toIRSpan(path, expression.Span)})
				return result, typ, nil
			}
			if typ == ir.TypeString {
				emptyStr := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, expression.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "+", Args: []string{identName, emptyStr}, Span: toIRSpan(path, expression.Span)})
				return result, typ, nil
			}
			if typ == ir.TypeBool {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "||", Args: []string{identName, identName}, Span: toIRSpan(path, expression.Span)})
				return result, typ, nil
			}
			if strings.HasPrefix(string(typ), "object:") || strings.HasSuffix(string(typ), "[]") {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   typ,
					Result: result,
					Args:   []string{identName},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, typ, nil
			}
		}
		return identName, typ, nil
	}
	if sig, ok := resolveFunctionSignature(path, expression.Text, signatures); ok {
		if result == "" {
			result = nextTemp(counter)
		}
		calleeName := ensureFunctionClosureTrampoline(path, sig, signatures)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpClosure,
			Type:   ir.TypeClosure,
			Result: result,
			Callee: calleeName,
			Args:   nil,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeClosure, nil
	}
	if topVar, ok := topLevelVars[expression.Text]; ok && topVar.Expression != nil && !inProgressVars[expression.Text] {
		isPrimitiveConst := topVar.VarDeclKind == "const" && (topVar.Expression.Kind == "number" || topVar.Expression.Kind == "string" || topVar.Expression.Kind == "bool" || topVar.Expression.Kind == "literal" || topVar.Expression.Kind == "null" || topVar.Expression.Kind == "undefined")
		if !isPrimitiveConst || function.Name != "main" {
			varTyp := toIRType(topVar.Type)
			if varTyp == "" {
				varTyp = toIRType(topVar.InferredType)
			}
			if varTyp == "" {
				varTyp = ir.TypeNumber
			}
			return expression.Text, varTyp, nil
		}
		inProgressVars[expression.Text] = true
		res, typ, err := lowerExpression(path, topVar.Expression, result, function, env, counter, shapes, signatures)
		delete(inProgressVars, expression.Text)
		return res, typ, err
	}
	if inProgressVars[expression.Text] {
		varTyp := ir.TypeNumber
		if topVar, ok := topLevelVars[expression.Text]; ok {
			if topVar.Type != "" {
				varTyp = toIRTypeForPath(path, topVar.Type)
			} else if topVar.InferredType != "" {
				varTyp = toIRTypeForPath(path, topVar.InferredType)
			}
		}
		if varTyp == "" {
			varTyp = ir.TypeNumber
		}
		return expression.Text, varTyp, nil
	}
	global, ok := builtinGlobal(expression.Text)
	if !ok {
		return "", "", fmt.Errorf("unknown identifier %q", expression.Text)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: result, Value: global.Value, Span: toIRSpan(path, expression.Span)})
	return result, global.Type, nil
}
