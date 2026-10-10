package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerJSONStringifyObject serializes an object of a known shape: a Date as
// its ISO string, a tuple as an array, a class instance field by field in
// declaration order, and a name-addressed object in its own key order.
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

	// Interface and object-literal shapes hold only the keys an object has,
	// in insertion order, and are addressed by name: each field's JSON text
	// is produced from its static type, and the runtime writes the present
	// keys in the object's own order (scriptgo_json_assemble_object).
	byName := dynamicFieldAccess(shape.Name) && !isTuple
	var names, texts []string
	for i, f := range shape.Fields {
		fVal := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:           ir.OpFieldGet,
			Type:         f.Type,
			Result:       fVal,
			Callee:       shape.Name,
			Field:        f.Name,
			FieldIndex:   i,
			DynamicField: byName,
			Args:         []string{argVal},
			Span:         toIRSpan(call.Path, call.Expression.Span),
		})

		subFunc := &ir.Function{}
		subCall := call
		subCall.Function = subFunc
		fStr, err := lowerJSONStringifyValue(subCall, fVal, f.Type)
		if err != nil {
			return "", "", err
		}
		call.Function.Body = append(call.Function.Body, subFunc.Body...)
		if byName {
			name := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: name, Value: f.Name, StringLiteral: true, Span: toIRSpan(call.Path, call.Expression.Span)})
			names = append(names, name)
			texts = append(texts, fStr)
			continue
		}

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
		if byName {
			present := nextTemp(call.Counter)
			emitted := nextTemp(call.Counter)
			emitHasOwnProperty(call.Function, call.Counter, toIRSpan(call.Path, call.Expression.Span), present, argVal, f.Name, true)
			call.Function.Body = append(call.Function.Body,
				ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: emitted, Operator: "&&", Args: []string{isNotUndef, present}, Span: toIRSpan(call.Path, call.Expression.Span)},
			)
			isNotUndef = emitted
		}

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

	if byName {
		span := toIRSpan(call.Path, call.Expression.Span)
		namesArray := nextTemp(call.Counter)
		textsArray := nextTemp(call.Counter)
		finalRes := call.Result
		if finalRes == "" {
			finalRes = nextTemp(call.Counter)
		}
		call.Function.Body = append(call.Function.Body,
			ir.Instruction{Op: ir.OpArray, Type: ir.TypeStringArray, Result: namesArray, Args: names, Span: span},
			ir.Instruction{Op: ir.OpArray, Type: ir.TypeStringArray, Result: textsArray, Args: texts, Span: span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: finalRes, Callee: "__json.assemble_object", Args: []string{argVal, namesArray, textsArray}, Span: span},
		)
		return finalRes, ir.TypeString, nil
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
