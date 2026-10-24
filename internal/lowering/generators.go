package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerGeneratorFunction lowers a function* or async function* declaration
// into a factory function that creates its generator and the generator's
// next() method. The body runs lazily, one step per next() (generatorStep).
func lowerGeneratorFunction(
	path string,
	statement frontend.SyntaxStatement,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (ir.Function, []ir.Function, []ir.ObjectShape, error) {
	name := functionIdentityForPath(path, statement.Name)
	genClassName := "Generator_" + name
	yieldType := generatorYieldType(statement)
	span := toIRSpan(path, statement.Span)

	nextFn, createdShapes := defineGeneratorClass(span, genClassName, yieldType, shapes)
	signatures[nextFn.Name] = nextFn

	factoryFn := ir.Function{Name: name, Span: span, ReturnType: ir.Type("object:" + genClassName)}
	env := map[string]ir.Type{}
	for _, p := range statement.Parameters {
		pType := toIRType(p.Type)
		if pType == "" && p.InferredType != "" {
			pType = toIRType(p.InferredType)
		}
		if pType == "" {
			pType = ir.TypeNumber
		}
		pType = variableStorageType(pType)
		factoryFn.Parameters = append(factoryFn.Parameters, ir.Parameter{Name: p.Name, Type: pType})
		env[p.Name] = pType
	}
	counter := 0
	if err := lowerGeneratorBody(path, statement, genClassName, yieldType, &factoryFn, env, &counter, shapes, signatures); err != nil {
		return ir.Function{}, nil, nil, err
	}
	signatures[name] = factoryFn
	return factoryFn, []ir.Function{nextFn}, createdShapes, nil
}

// generatorYieldType is the IR type of the values a generator yields: the
// first type argument of its declared or inferred Generator<T> type.
func generatorYieldType(statement frontend.SyntaxStatement) ir.Type {
	for _, typeName := range []string{statement.Type, statement.InferredType} {
		typeName = strings.TrimSpace(typeName)
		if idx := strings.Index(typeName, "<"); idx >= 0 && strings.HasSuffix(typeName, ">") {
			if parts := splitTypeArguments(typeName[idx+1 : len(typeName)-1]); len(parts) > 0 {
				if yieldType := toIRType(parts[0]); yieldType != "" {
					return yieldType
				}
			}
		}
	}
	return ir.TypeNumber
}

// defineGeneratorClass defines a generator class, whose one field __next is
// the step closure, its IteratorResult shape and its next() method. It
// returns next() and the shapes it created.
func defineGeneratorClass(span ir.SourceSpan, genClassName string, yieldType ir.Type, shapes map[string]ir.ObjectShape) (ir.Function, []ir.ObjectShape) {
	var created []ir.ObjectShape
	resultShapeName := "IteratorResult_" + string(yieldType)
	if _, exists := shapes[resultShapeName]; !exists {
		resultShape := ir.ObjectShape{Name: resultShapeName, Span: span, Fields: []ir.Field{
			{Name: "done", Type: ir.TypeBool, Span: span},
			{Name: "value", Type: yieldType, Span: span},
		}}
		shapes[resultShapeName] = resultShape
		created = append(created, resultShape)
	}
	// __next is called with the generator as its argument by
	// generatorNextCall; the step closure ignores it.
	genShape := ir.ObjectShape{Name: genClassName, Span: span, Fields: []ir.Field{{Name: "__next", Type: ir.TypeClosure, Span: span}}}
	shapes[genClassName] = genShape
	created = append(created, genShape)

	if classHierarchy == nil {
		classHierarchy = map[string]ClassMeta{}
	}
	classHierarchy[genClassName] = ClassMeta{Name: genClassName, Extends: "Generator", Statics: map[string]frontend.SyntaxField{}}
	if classSyntax == nil {
		classSyntax = map[string]frontend.SyntaxClass{}
	}
	classSyntax[genClassName] = frontend.SyntaxClass{Name: genClassName, Methods: []frontend.SyntaxMethod{{Name: "next", Type: resultShapeName}}}

	resultType := ir.Type("object:" + resultShapeName)
	next := ir.Function{
		Name:       genClassName + "_next",
		Span:       span,
		ReturnType: resultType,
		Parameters: []ir.Parameter{{Name: "this", Type: ir.Type("object:" + genClassName)}},
		Body: []ir.Instruction{
			{Op: ir.OpFieldGet, Type: ir.TypeClosure, Result: "step", Callee: genClassName, Field: "__next", FieldIndex: 0, Args: []string{"this"}, Span: span},
			{Op: ir.OpClosureCall, Type: resultType, Result: "result", Callee: "step", Span: span},
			{Op: ir.OpReturn, Type: resultType, Args: []string{"result"}, Span: span},
		},
	}
	return next, created
}

// lowerGeneratorBody appends to target, the generator's factory, the
// generator's state, its step closure and the generator object it returns.
func lowerGeneratorBody(path string, statement frontend.SyntaxStatement, genClassName string, yieldType ir.Type, target *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	prelude, step, err := generatorStep(path, statement, string(yieldType))
	if err != nil {
		return err
	}
	for _, s := range prelude {
		if err := lowerStatement(path, s, target, env, counter, shapes, signatures); err != nil {
			return sourceError(path, s.Span, err)
		}
	}
	closure, _, err := lowerExpression(path, step, "", target, env, counter, shapes, signatures)
	if err != nil {
		return sourceError(path, statement.Span, err)
	}
	span := toIRSpan(path, statement.Span)
	genType := ir.Type("object:" + genClassName)
	genObj := nextTemp(counter)
	target.Body = append(target.Body,
		ir.Instruction{Op: ir.OpObjectNew, Type: genType, Result: genObj, Callee: genClassName, FieldCount: 1, Args: []string{closure}, Span: span},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: genClassName, Field: "__next", FieldIndex: 0, Args: []string{genObj, closure}, Span: span},
		ir.Instruction{Op: ir.OpReturn, Type: genType, Args: []string{genObj}, Span: span},
	)
	return nil
}

// lowerGeneratorResult lowers the { done, value } a step closure returns
// (generatorResume.result); a missing value is the yielded type's undefined.
func lowerGeneratorResult(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	span := toIRSpan(path, expression.Span)
	yieldType := ir.Type(expression.InferredType)
	resultShapeName := "IteratorResult_" + string(yieldType)
	resultType := ir.Type("object:" + resultShapeName)
	var value string
	if expression.Left != nil {
		lowered, valueType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		value = coerceGeneratorValue(function, counter, lowered, valueType, yieldType, span)
	} else {
		value = nextTemp(counter)
		undefinedValue := "undefined"
		if yieldType == ir.TypeBool || yieldType == ir.TypeBigInt {
			undefinedValue = "0"
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: value, Value: undefinedValue, Span: span})
	}
	done := nextTemp(counter)
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body,
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: done, Value: expression.Text, Span: span},
		ir.Instruction{Op: ir.OpObjectNew, Type: resultType, Result: result, Callee: resultShapeName, FieldCount: 2, Args: []string{done, value}, Span: span},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: resultShapeName, Field: "done", FieldIndex: 0, Args: []string{result, done}, Span: span},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: resultShapeName, Field: "value", FieldIndex: 1, Args: []string{result, value}, Span: span},
	)
	return result, resultType, nil
}

// coerceGeneratorValue boxes a yielded value into an unknown yield type, or
// checks an unknown value against a concrete one, as a return would.
func coerceGeneratorValue(function *ir.Function, counter *int, value string, valueType, yieldType ir.Type, span ir.SourceSpan) string {
	op := ""
	switch {
	case yieldType == ir.TypeUnknown && valueType != ir.TypeUnknown:
		op = ir.OpBoxUnknown
	case yieldType != ir.TypeUnknown && valueType == ir.TypeUnknown:
		op = ir.OpCheckedCast
	default:
		return value
	}
	converted := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: op, Type: yieldType, Result: converted, Args: []string{value}, Span: span})
	return converted
}
