package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerBufferMethod(path string, expression *typescriptgo.SyntaxExpression, receiver, methodName string, receiverType ir.Type, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if receiverType != ir.TypeBuffer {
		return "", "", false, nil
	}

	if result == "" {
		result = nextTemp(counter)
	}

	switch methodName {
	case "toString":
		encVal := nextTemp(counter)
		startVal := nextTemp(counter)
		endVal := nextTemp(counter)
		hasStartVal := nextTemp(counter)
		hasEndVal := nextTemp(counter)

		if len(expression.Arguments) > 0 {
			ev, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			encVal = ev
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: encVal, Value: "utf8", Span: toIRSpan(path, expression.Span),
			})
		}

		if len(expression.Arguments) > 1 {
			sv, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			startVal = sv
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: hasStartVal, Value: "true", Span: toIRSpan(path, expression.Span),
			})
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: startVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: hasStartVal, Value: "false", Span: toIRSpan(path, expression.Span),
			})
		}

		if len(expression.Arguments) > 2 {
			ev, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			endVal = ev
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: hasEndVal, Value: "true", Span: toIRSpan(path, expression.Span),
			})
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: endVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: hasEndVal, Value: "false", Span: toIRSpan(path, expression.Span),
			})
		}

		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: result,
			Callee: "__buffer.toString",
			Args:   []string{receiver, encVal, startVal, endVal, hasStartVal, hasEndVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, true, nil

	case "copy":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("Buffer.copy requires target argument")
		}
		targetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		targetStartVal := nextTemp(counter)
		sourceStartVal := nextTemp(counter)
		sourceEndVal := nextTemp(counter)
		hasTSVal := nextTemp(counter)
		hasSSVal := nextTemp(counter)
		hasSEVal := nextTemp(counter)

		if len(expression.Arguments) > 1 {
			v, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			targetStartVal = v
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasTSVal, Value: "true", Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: targetStartVal, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasTSVal, Value: "false", Span: toIRSpan(path, expression.Span)})
		}

		if len(expression.Arguments) > 2 {
			v, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			sourceStartVal = v
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasSSVal, Value: "true", Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: sourceStartVal, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasSSVal, Value: "false", Span: toIRSpan(path, expression.Span)})
		}

		if len(expression.Arguments) > 3 {
			v, _, err := lowerExpression(path, expression.Arguments[3], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			sourceEndVal = v
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasSEVal, Value: "true", Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: sourceEndVal, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasSEVal, Value: "false", Span: toIRSpan(path, expression.Span)})
		}

		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__buffer.copy",
			Args:   []string{receiver, targetVal, targetStartVal, sourceStartVal, sourceEndVal, hasTSVal, hasSSVal, hasSEVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, true, nil

	case "equals":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("Buffer.equals requires other argument")
		}
		otherVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: result,
			Callee: "__buffer.equals",
			Args:   []string{receiver, otherVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, true, nil

	case "compare":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("Buffer.compare requires other argument")
		}
		otherVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__buffer.compare",
			Args:   []string{receiver, otherVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, true, nil

	case "indexOf":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("Buffer.indexOf requires value argument")
		}
		valVal, valType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		isStrVal := nextTemp(counter)
		if valType == ir.TypeString {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: isStrVal, Value: "true", Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: isStrVal, Value: "false", Span: toIRSpan(path, expression.Span)})
		}
		offVal := nextTemp(counter)
		hasOffVal := nextTemp(counter)
		if len(expression.Arguments) > 1 {
			v, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			offVal = v
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasOffVal, Value: "true", Span: toIRSpan(path, expression.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: offVal, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: hasOffVal, Value: "false", Span: toIRSpan(path, expression.Span)})
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__buffer.indexOf",
			Args:   []string{receiver, valVal, isStrVal, offVal, hasOffVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, true, nil

	// Binary read operations (1 arg: offset)
	case "readUInt8", "readInt8",
		"readUInt16LE", "readUInt16BE",
		"readInt16LE", "readInt16BE",
		"readUInt32LE", "readUInt32BE",
		"readInt32LE", "readInt32BE",
		"readFloatLE", "readFloatBE",
		"readDoubleLE", "readDoubleBE":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("%s requires offset argument", methodName)
		}
		offVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__buffer." + methodName,
			Args:   []string{receiver, offVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, true, nil

	// Binary write operations (2 args: value, offset)
	case "writeUInt8", "writeInt8",
		"writeUInt16LE", "writeUInt16BE",
		"writeInt16LE", "writeInt16BE",
		"writeUInt32LE", "writeUInt32BE",
		"writeInt32LE", "writeInt32BE",
		"writeFloatLE", "writeFloatBE",
		"writeDoubleLE", "writeDoubleBE":
		if len(expression.Arguments) < 2 {
			return "", "", true, fmt.Errorf("%s requires value and offset arguments", methodName)
		}
		valVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		offVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__buffer." + methodName,
			Args:   []string{receiver, valVal, offVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, true, nil
	}

	return "", "", false, nil
}
