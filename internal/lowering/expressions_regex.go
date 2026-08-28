package lowering

import (
	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func init() {
	// Register RegExp shape globally
	if _, ok := builtinShapes["RegExp"]; !ok {
		builtinShapes["RegExp"] = ir.ObjectShape{
			Name: "RegExp",
			Fields: []ir.Field{
				{Name: "source", Type: ir.TypeString},
				{Name: "flags", Type: ir.TypeString},
				{Name: "lastIndex", Type: ir.TypeNumber},
			},
		}
	}
}

var builtinShapes = map[string]ir.ObjectShape{}

func ensureRegExpShape(shapes map[string]ir.ObjectShape) {
	if _, ok := shapes["RegExp"]; !ok {
		shapes["RegExp"] = ir.ObjectShape{
			Name: "RegExp",
			Fields: []ir.Field{
				{Name: "source", Type: ir.TypeString},
				{Name: "flags", Type: ir.TypeString},
				{Name: "lastIndex", Type: ir.TypeNumber},
			},
		}
	}
}

func lowerRegexLiteral(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	ensureRegExpShape(shapes)

	pattern := expression.Text
	flags := expression.Operator

	res := result
	if res == "" {
		res = nextTemp(counter)
	}

	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       ir.Type("object:RegExp"),
		Result:     res,
		Value:      ":RegExp:",
		FieldCount: 3,
		Span:       toIRSpan(path, expression.Span),
	})

	srcConst := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: srcConst,
		Value:  pattern,
		Span:   toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Callee:     "RegExp",
		Field:      "source",
		FieldIndex: 0,
		Args:       []string{res, srcConst},
		Span:       toIRSpan(path, expression.Span),
	})

	flagsConst := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: flagsConst,
		Value:  flags,
		Span:   toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Callee:     "RegExp",
		Field:      "flags",
		FieldIndex: 1,
		Args:       []string{res, flagsConst},
		Span:       toIRSpan(path, expression.Span),
	})

	lastIdxConst := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeNumber,
		Result: lastIdxConst,
		Value:  "0",
		Span:   toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpFieldSet,
		Type:       ir.TypeVoid,
		Callee:     "RegExp",
		Field:      "lastIndex",
		FieldIndex: 2,
		Args:       []string{res, lastIdxConst},
		Span:       toIRSpan(path, expression.Span),
	})

	return res, ir.Type("object:RegExp"), nil
}
