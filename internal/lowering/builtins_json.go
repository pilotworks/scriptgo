package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func isTupleShape(shape ir.ObjectShape) bool {
	if len(shape.Fields) == 0 {
		return false
	}
	for i, f := range shape.Fields {
		if f.Name != strconv.Itoa(i) {
			return false
		}
	}
	return true
}

func findShape(call IntrinsicCall, shapeName string) (ir.ObjectShape, bool) {
	shapeName = strings.TrimPrefix(shapeName, "object:")
	if shape, ok := call.Shapes[shapeName]; ok && len(shape.Fields) > 0 {
		return shape, true
	}
	if shape, ok := anonymousShapes[shapeName]; ok && len(shape.Fields) > 0 {
		return shape, true
	}
	if shape, ok := registeredShapes[shapeName]; ok && len(shape.Fields) > 0 {
		return shape, true
	}
	if parsedShape, ok := parseAnonymousObjectShape(shapeName); ok && len(parsedShape.Fields) > 0 {
		return parsedShape, true
	}
	if fields, ok := tupleFields(shapeName); ok && len(fields) > 0 {
		return ir.ObjectShape{Name: shapeName, Fields: fields}, true
	}
	return ir.ObjectShape{}, false
}

func lowerJSONStringify(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) != 1 {
		return "", "", fmt.Errorf("JSON.stringify expects exactly 1 argument")
	}
	argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	res, err := lowerJSONStringifyValue(call, argVal, argType)
	if err != nil {
		return "", "", err
	}
	if call.Result != "" && res != call.Result {
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeString,
			Result: call.Result,
			Args:   []string{res},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return call.Result, ir.TypeString, nil
	}
	return res, ir.TypeString, nil
}

func lowerJSONStringifyValue(call IntrinsicCall, val string, valType ir.Type) (string, error) {
	switch {
	case valType == ir.TypeString:
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_string",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil

	case valType == ir.TypeNumber:
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_number",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil

	case valType == ir.TypeBool:
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_bool",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil

	case valType == ir.TypeNumberArray:
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_number_array",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil

	case valType == ir.TypeStringArray:
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_string_array",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil

	case strings.HasSuffix(string(valType), "[]"):
		elemType := ir.Type(strings.TrimSuffix(string(valType), "[]"))
		return lowerJSONStringifyArray(call, val, elemType)

	case strings.HasPrefix(string(valType), "object:") || strings.HasPrefix(string(valType), "__shape_") || strings.HasPrefix(string(valType), "["):
		shapeName := strings.TrimPrefix(string(valType), "object:")
		if shape, ok := findShape(call, shapeName); ok {
			res, _, err := lowerJSONStringifyObject(call, val, shape)
			return res, err
		}
		emptyObj := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: emptyObj,
			Value:  "{}",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return emptyObj, nil

	default:
		shapeName := strings.TrimPrefix(string(valType), "object:")
		if shape, ok := findShape(call, shapeName); ok {
			res, _, err := lowerJSONStringifyObject(call, val, shape)
			return res, err
		}
		res := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: res,
			Callee: "__json.stringify_unknown",
			Args:   []string{val},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return res, nil
	}
}

func lowerJSONStringifyArray(call IntrinsicCall, arrVal string, elemType ir.Type) (string, error) {
	res := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: res,
		Value:  "[",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	lenRes := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeNumber,
		Result: lenRes,
		Callee: "__array.length",
		Args:   []string{arrVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	idxVar := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeNumber,
		Result: idxVar,
		Value:  "0",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	condRes := nextTemp(call.Counter)
	condBlock := []ir.Instruction{
		{
			Op:       ir.OpCompare,
			Type:     ir.TypeBool,
			Result:   condRes,
			Operator: "<",
			Args:     []string{idxVar, lenRes},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		},
	}

	bodyBlock := []ir.Instruction{}

	zeroConst := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeNumber,
		Result: zeroConst,
		Value:  "0",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	gtZero := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:       ir.OpCompare,
		Type:     ir.TypeBool,
		Result:   gtZero,
		Operator: ">",
		Args:     []string{idxVar, zeroConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})

	commaConst := nextTemp(call.Counter)
	afterComma := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{gtZero},
		Then: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeString, Result: commaConst, Value: ",", Span: toIRSpan(call.Path, call.Expression.Span)},
			{Op: ir.OpBinary, Type: ir.TypeString, Result: afterComma, Operator: "+", Args: []string{res, commaConst}, Span: toIRSpan(call.Path, call.Expression.Span)},
			{Op: ir.OpAssign, Type: ir.TypeString, Result: res, Args: []string{afterComma}, Span: toIRSpan(call.Path, call.Expression.Span)},
		},
		Span: toIRSpan(call.Path, call.Expression.Span),
	})

	elemVal := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:     ir.OpIndex,
		Type:   elemType,
		Result: elemVal,
		Args:   []string{arrVal, idxVar},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	subFunc := &ir.Function{}
	subCall := call
	subCall.Function = subFunc
	elemStr, err := lowerJSONStringifyValue(subCall, elemVal, elemType)
	if err != nil {
		return "", err
	}
	bodyBlock = append(bodyBlock, subFunc.Body...)

	afterElem := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:       ir.OpBinary,
		Type:     ir.TypeString,
		Result:   afterElem,
		Operator: "+",
		Args:     []string{res, elemStr},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:     ir.OpAssign,
		Type:   ir.TypeString,
		Result: res,
		Args:   []string{afterElem},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	oneConst := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeNumber,
		Result: oneConst,
		Value:  "1",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	nextIdx := nextTemp(call.Counter)
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:       ir.OpBinary,
		Type:     ir.TypeNumber,
		Result:   nextIdx,
		Operator: "+",
		Args:     []string{idxVar, oneConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})
	bodyBlock = append(bodyBlock, ir.Instruction{
		Op:     ir.OpAssign,
		Type:   ir.TypeNumber,
		Result: idxVar,
		Args:   []string{nextIdx},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:   ir.OpWhile,
		Type: ir.TypeVoid,
		Args: []string{condRes},
		Cond: condBlock,
		Body: bodyBlock,
		Span: toIRSpan(call.Path, call.Expression.Span),
	})

	closeConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: closeConst,
		Value:  "]",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	finalRes := call.Result
	if finalRes == "" {
		finalRes = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:       ir.OpBinary,
		Type:     ir.TypeString,
		Result:   finalRes,
		Operator: "+",
		Args:     []string{res, closeConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})
	return finalRes, nil
}

func lowerJSONStringifyObject(call IntrinsicCall, argVal string, shape ir.ObjectShape) (string, ir.Type, error) {
	if shape.Name == "Date" {
		timeVal := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       ir.TypeNumber,
			Result:     timeVal,
			Callee:     "Date",
			Field:      "time",
			FieldIndex: 0,
			Args:       []string{argVal},
			Span:       toIRSpan(call.Path, call.Expression.Span),
		})
		isoStr := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: isoStr,
			Callee: "__date.toISOString",
			Args:   []string{timeVal},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		jsonStr := call.Result
		if jsonStr == "" {
			jsonStr = nextTemp(call.Counter)
		}
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: jsonStr,
			Callee: "__json.stringify_string",
			Args:   []string{isoStr},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return jsonStr, ir.TypeString, nil
	}

	isTuple := isTupleShape(shape)
	openBracket := "{"
	closeBracket := "}"
	if isTuple {
		openBracket = "["
		closeBracket = "]"
	}

	res := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: res,
		Value:  openBracket,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	hasFieldsConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: hasFieldsConst,
		Value:  "false",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	for i, f := range shape.Fields {
		fVal := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       f.Type,
			Result:     fVal,
			Callee:     shape.Name,
			Field:      f.Name,
			FieldIndex: i,
			Args:       []string{argVal},
			Span:       toIRSpan(call.Path, call.Expression.Span),
		})

		subFunc := &ir.Function{}
		subCall := call
		subCall.Function = subFunc
		fStr, err := lowerJSONStringifyValue(subCall, fVal, f.Type)
		if err != nil {
			return "", "", err
		}
		call.Function.Body = append(call.Function.Body, subFunc.Body...)

		undefStr := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: undefStr,
			Value:  "undefined",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		isNotUndef := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpCompare,
			Type:     ir.TypeBool,
			Result:   isNotUndef,
			Operator: "!=",
			Args:     []string{fStr, undefStr},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})

		var thenInstructions []ir.Instruction

		commaConst := nextTemp(call.Counter)
		afterComma := nextTemp(call.Counter)
		thenInstructions = append(thenInstructions, ir.Instruction{
			Op:   ir.OpIf,
			Type: ir.TypeVoid,
			Args: []string{hasFieldsConst},
			Then: []ir.Instruction{
				{Op: ir.OpConst, Type: ir.TypeString, Result: commaConst, Value: ",", Span: toIRSpan(call.Path, call.Expression.Span)},
				{Op: ir.OpBinary, Type: ir.TypeString, Result: afterComma, Operator: "+", Args: []string{res, commaConst}, Span: toIRSpan(call.Path, call.Expression.Span)},
				{Op: ir.OpAssign, Type: ir.TypeString, Result: res, Args: []string{afterComma}, Span: toIRSpan(call.Path, call.Expression.Span)},
			},
			Span: toIRSpan(call.Path, call.Expression.Span),
		})

		if !isTuple {
			prefix := fmt.Sprintf("\"%s\":", f.Name)
			prefConst := nextTemp(call.Counter)
			afterPref := nextTemp(call.Counter)
			thenInstructions = append(thenInstructions,
				ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: prefConst, Value: prefix, Span: toIRSpan(call.Path, call.Expression.Span)},
				ir.Instruction{Op: ir.OpBinary, Type: ir.TypeString, Result: afterPref, Operator: "+", Args: []string{res, prefConst}, Span: toIRSpan(call.Path, call.Expression.Span)},
				ir.Instruction{Op: ir.OpAssign, Type: ir.TypeString, Result: res, Args: []string{afterPref}, Span: toIRSpan(call.Path, call.Expression.Span)},
			)
		}

		afterVal := nextTemp(call.Counter)
		trueConst := nextTemp(call.Counter)
		thenInstructions = append(thenInstructions,
			ir.Instruction{Op: ir.OpBinary, Type: ir.TypeString, Result: afterVal, Operator: "+", Args: []string{res, fStr}, Span: toIRSpan(call.Path, call.Expression.Span)},
			ir.Instruction{Op: ir.OpAssign, Type: ir.TypeString, Result: res, Args: []string{afterVal}, Span: toIRSpan(call.Path, call.Expression.Span)},
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(call.Path, call.Expression.Span)},
			ir.Instruction{Op: ir.OpAssign, Type: ir.TypeBool, Result: hasFieldsConst, Args: []string{trueConst}, Span: toIRSpan(call.Path, call.Expression.Span)},
		)

		if !isTuple {
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{isNotUndef},
				Then: thenInstructions,
				Span: toIRSpan(call.Path, call.Expression.Span),
			})
		} else {
			call.Function.Body = append(call.Function.Body, thenInstructions...)
		}
	}

	closeConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: closeConst,
		Value:  closeBracket,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	finalRes := call.Result
	if finalRes == "" {
		finalRes = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:       ir.OpBinary,
		Type:     ir.TypeString,
		Result:   finalRes,
		Operator: "+",
		Args:     []string{res, closeConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})
	return finalRes, ir.TypeString, nil
}

func parseAnonymousObjectShape(shapeName string) (ir.ObjectShape, bool) {
	clean := strings.TrimSpace(shapeName)
	clean = strings.TrimPrefix(clean, "object:")
	if !strings.HasPrefix(clean, "{") || !strings.HasSuffix(clean, "}") {
		return ir.ObjectShape{}, false
	}
	inner := strings.Trim(clean, "{}")
	var fields []ir.Field
	rawParts := strings.FieldsFunc(inner, func(r rune) bool {
		return r == ';' || r == ','
	})
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		fName := strings.TrimSpace(kv[0])
		fType := strings.TrimSpace(kv[1])
		fields = append(fields, ir.Field{
			Name: fName,
			Type: toIRType(fType),
		})
	}
	if len(fields) == 0 {
		return ir.ObjectShape{}, false
	}
	return ir.ObjectShape{
		Name:   shapeName,
		Fields: fields,
	}, true
}
