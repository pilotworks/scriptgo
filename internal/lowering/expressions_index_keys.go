package lowering

import (
	"strconv"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerNumericKeyRead lowers obj[n] on a non-tuple object shape. A numeric
// literal key names a field ("16" for obj[16], "16" for obj[0x10]); any other
// key is converted with ToString and looked up at run time.
func lowerNumericKeyRead(path string, expression *frontend.SyntaxExpression, object, shapeName, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	span := toIRSpan(path, expression.Span)
	if key, ok := numericLiteralKey(expression.Right); ok {
		if shape, found := lookupObjectShape(shapeName, shapes); found {
			for index, field := range shape.Fields {
				if field.Name == key {
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: result, Callee: shapeName, Field: field.Name, FieldIndex: index, Args: []string{object}, Span: span})
					return result, field.Type, nil
				}
			}
		}
	}
	keyValue, keyType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	keyValue, _ = coercePrimitiveToString(path, expression.Right.Span, keyValue, keyType, function, counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeUnknown, Result: result, Callee: "__object.get_prop", Args: []string{object, keyValue}, Span: span})
	return result, ir.TypeUnknown, nil
}

// numericLiteralKey is ToString of a numeric literal property key.
func numericLiteralKey(expression *frontend.SyntaxExpression) (string, bool) {
	if expression == nil || expression.Kind != "number" {
		return "", false
	}
	if n, err := strconv.ParseInt(expression.Text, 0, 64); err == nil {
		return strconv.FormatInt(n, 10), true
	}
	if f, err := strconv.ParseFloat(expression.Text, 64); err == nil {
		return formatJSNumber(f), true
	}
	return "", false
}

func lookupObjectShape(name string, shapes map[string]ir.ObjectShape) (ir.ObjectShape, bool) {
	if shape, ok := shapes[name]; ok {
		return shape, true
	}
	if shape, ok := registeredShapes[name]; ok {
		return shape, true
	}
	shape, ok := anonymousShapes[name]
	return shape, ok
}

// isKeyedObjectShape reports a known object shape whose fields are named
// properties, as opposed to a tuple or positional entry shape (fields named
// "0", "1", ... or left unnamed), which numeric indexes address by position.
func isKeyedObjectShape(name string, shapes map[string]ir.ObjectShape) bool {
	shape, ok := lookupObjectShape(name, shapes)
	if !ok || len(shape.Fields) == 0 || isTupleShape(shape) {
		return false
	}
	for _, field := range shape.Fields {
		if field.Name == "" {
			return false
		}
	}
	return true
}
