package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var generatorASTIndex = map[string]typescriptgo.SyntaxStatement{}

// RegisterGeneratorStatement indexes a generator function statement by name for delegation analysis.
func RegisterGeneratorStatement(name string, stmt typescriptgo.SyntaxStatement) {
	if generatorASTIndex == nil {
		generatorASTIndex = map[string]typescriptgo.SyntaxStatement{}
	}
	generatorASTIndex[name] = stmt
}

type yieldPoint struct {
	stateIdx int
	expr     *typescriptgo.SyntaxExpression
	isStar   bool
}

func rewriteYieldsToPush(stmts []typescriptgo.SyntaxStatement, itemsName string) []typescriptgo.SyntaxStatement {
	var rewritten []typescriptgo.SyntaxStatement
	for _, s := range stmts {
		cloned := s
		if s.Kind == "expression" && s.Expression != nil && (s.Expression.Kind == "yield" || s.Expression.Kind == "yield_star") {
			cloned = typescriptgo.SyntaxStatement{
				Span: s.Span,
				Kind: "expression",
				Expression: &typescriptgo.SyntaxExpression{
					Span: s.Span,
					Kind: "call",
					Left: &typescriptgo.SyntaxExpression{
						Span: s.Span,
						Kind: "property",
						Left: &typescriptgo.SyntaxExpression{
							Span: s.Span,
							Kind: "identifier",
							Text: itemsName,
						},
						Text: "push",
					},
					Arguments: []*typescriptgo.SyntaxExpression{
						s.Expression.Left,
					},
				},
			}
		}
		if len(cloned.Body) > 0 {
			cloned.Body = rewriteYieldsToPush(cloned.Body, itemsName)
		}
		if len(cloned.Then) > 0 {
			cloned.Then = rewriteYieldsToPush(cloned.Then, itemsName)
		}
		if len(cloned.Else) > 0 {
			cloned.Else = rewriteYieldsToPush(cloned.Else, itemsName)
		}
		if len(cloned.Catch) > 0 {
			cloned.Catch = rewriteYieldsToPush(cloned.Catch, itemsName)
		}
		if len(cloned.Finally) > 0 {
			cloned.Finally = rewriteYieldsToPush(cloned.Finally, itemsName)
		}
		rewritten = append(rewritten, cloned)
	}
	return rewritten
}

func collectYieldPoints(body []typescriptgo.SyntaxStatement) []yieldPoint {
	var yields []yieldPoint
	for _, s := range body {
		if s.Kind == "expression" && s.Expression != nil {
			if s.Expression.Kind == "yield" {
				yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: false})
			} else if s.Expression.Kind == "yield_star" {
				if s.Expression.Left != nil && s.Expression.Left.Kind == "array" {
					for _, arg := range s.Expression.Left.Arguments {
						yields = append(yields, yieldPoint{stateIdx: len(yields), expr: arg, isStar: false})
					}
				} else if s.Expression.Left != nil && s.Expression.Left.Kind == "call" && s.Expression.Left.Left != nil && s.Expression.Left.Left.Kind == "identifier" {
					calleeName := s.Expression.Left.Left.Text
					if subGen, ok := generatorASTIndex[calleeName]; ok {
						subYields := collectYieldPoints(subGen.Body)
						for _, sy := range subYields {
							sy.stateIdx = len(yields)
							yields = append(yields, sy)
						}
					} else {
						yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: true})
					}
				} else {
					yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: true})
				}
			}
		}
	}
	return yields
}

// lowerGeneratorFunction lowers a function* or async function* into:
// 1. An IteratorResult_<T> object shape
// 2. A Generator_<name> object shape
// 3. A Generator_<name>_next(this) method implementing the state-machine
// 4. A factory function <name>(params...) that instantiates and returns the generator
func lowerGeneratorFunction(
	path string,
	statement typescriptgo.SyntaxStatement,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (ir.Function, []ir.Function, []ir.ObjectShape, error) {
	name := statement.Name
	genClassName := "Generator_" + name

	// Determine yielded type from statements or type annotation
	yieldType := ir.TypeNumber
	if strings.Contains(statement.Type, "<") && strings.HasSuffix(statement.Type, ">") {
		idx := strings.Index(statement.Type, "<")
		inner := statement.Type[idx+1 : len(statement.Type)-1]
		parts := splitTypeArguments(inner)
		if len(parts) > 0 {
			yieldType = toIRType(parts[0])
		}
	} else if strings.Contains(statement.InferredType, "<") && strings.HasSuffix(statement.InferredType, ">") {
		idx := strings.Index(statement.InferredType, "<")
		inner := statement.InferredType[idx+1 : len(statement.InferredType)-1]
		parts := splitTypeArguments(inner)
		if len(parts) > 0 {
			yieldType = toIRType(parts[0])
		}
	} else {
		for _, stmt := range statement.Body {
			if stmt.Expression != nil && (stmt.Expression.Kind == "yield" || stmt.Expression.Kind == "yield_star") {
				if stmt.Expression.Left != nil {
					if stmt.Expression.Left.Kind == "string" || stmt.Expression.Left.InferredType == "string" {
						yieldType = ir.TypeString
					} else if stmt.Expression.Left.Kind == "bigint" || stmt.Expression.Left.InferredType == "bigint" {
						yieldType = ir.TypeBigInt
					} else if stmt.Expression.Left.Kind == "bool" || stmt.Expression.Left.InferredType == "bool" {
						yieldType = ir.TypeBool
					}
				}
			}
		}
		if strings.Contains(statement.Type, "string") || strings.Contains(statement.InferredType, "string") {
			yieldType = ir.TypeString
		} else if strings.Contains(statement.Type, "bigint") || strings.Contains(statement.InferredType, "bigint") {
			yieldType = ir.TypeBigInt
		}
	}

	resultShapeName := fmt.Sprintf("IteratorResult_%s", yieldType)

	// 1. Define IteratorResult shape if not present
	var createdShapes []ir.ObjectShape
	if _, exists := shapes[resultShapeName]; !exists {
		resultShape := ir.ObjectShape{
			Name: resultShapeName,
			Span: toIRSpan(path, statement.Span),
			Fields: []ir.Field{
				{Name: "value", Type: yieldType, Span: toIRSpan(path, statement.Span)},
				{Name: "done", Type: ir.TypeBool, Span: toIRSpan(path, statement.Span)},
			},
		}
		shapes[resultShapeName] = resultShape
		createdShapes = append(createdShapes, resultShape)
	}

	// 2. Define Generator struct shape
	genFields := []ir.Field{
		{Name: "__state", Type: ir.TypeNumber, Value: "0", Span: toIRSpan(path, statement.Span)},
		{Name: "__done", Type: ir.TypeBool, Value: "false", Span: toIRSpan(path, statement.Span)},
		{Name: "__value", Type: yieldType, Span: toIRSpan(path, statement.Span)},
	}

	// Add parameter fields
	for _, p := range statement.Parameters {
		pType := toIRType(p.Type)
		if pType == "" && p.InferredType != "" {
			pType = toIRType(p.InferredType)
		}
		if pType == "" {
			pType = ir.TypeNumber
		}
		genFields = append(genFields, ir.Field{
			Name: p.Name,
			Type: pType,
			Span: toIRSpan(path, p.Span),
		})
	}

	genFields = append(genFields, ir.Field{
		Name: "__items",
		Type: ir.Type(string(yieldType) + "[]"),
		Span: toIRSpan(path, statement.Span),
	})

	genShape := ir.ObjectShape{
		Name:   genClassName,
		Span:   toIRSpan(path, statement.Span),
		Fields: genFields,
	}
	shapes[genClassName] = genShape
	createdShapes = append(createdShapes, genShape)

	// Register in classHierarchy and classSyntax
	if classHierarchy == nil {
		classHierarchy = map[string]ClassMeta{}
	}
	classHierarchy[genClassName] = ClassMeta{
		Name:    genClassName,
		Statics: map[string]typescriptgo.SyntaxField{},
	}
	if classSyntax == nil {
		classSyntax = map[string]typescriptgo.SyntaxClass{}
	}
	classSyntax[genClassName] = typescriptgo.SyntaxClass{
		Name: genClassName,
		Methods: []typescriptgo.SyntaxMethod{
			{
				Name: "next",
				Type: resultShapeName,
			},
		},
	}

	// 3. Build next() method: Generator_<name>_next(this: Generator_<name>) -> IteratorResult_<T>
	nextMethodName := genClassName + "_next"
	nextFn := ir.Function{
		Name:       nextMethodName,
		Span:       toIRSpan(path, statement.Span),
		ReturnType: ir.Type("object:" + resultShapeName),
		Parameters: []ir.Parameter{
			{Name: "this", Type: ir.Type("object:" + genClassName)},
		},
	}
	signatures[nextMethodName] = nextFn

	nextEnv := map[string]ir.Type{
		"this": ir.Type("object:" + genClassName),
	}
	counter := 0

	// Compile state machine inside nextFn
	// Read this.__state
	stateVal := nextTemp(&counter)
	nextFn.Body = append(nextFn.Body, ir.Instruction{
		Op:         ir.OpFieldGet,
		Type:       ir.TypeNumber,
		Result:     stateVal,
		Callee:     genClassName,
		Field:      "__state",
		FieldIndex: 0,
		Args:       []string{"this"},
		Span:       toIRSpan(path, statement.Span),
	})

	yields := collectYieldPoints(statement.Body)

	// Dispatch states
	if len(yields) == 0 {
		// Empty generator: immediately returns done: true
		resObj := nextTemp(&counter)
		defVal := nextTemp(&counter)
		nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: defVal, Value: "0", Span: toIRSpan(path, statement.Span)})
		trueVal := nextTemp(&counter)
		nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueVal, Value: "true", Span: toIRSpan(path, statement.Span)})
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpObjectNew,
			Type:       ir.Type("object:" + resultShapeName),
			Result:     resObj,
			Callee:     resultShapeName,
			FieldCount: 2,
			Args:       []string{defVal, trueVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     resultShapeName,
			Field:      "value",
			FieldIndex: 0,
			Args:       []string{resObj, defVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     resultShapeName,
			Field:      "done",
			FieldIndex: 1,
			Args:       []string{resObj, trueVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:" + resultShapeName), Args: []string{resObj}, Span: toIRSpan(path, statement.Span)})
	} else {
		for i, y := range yields {
			// Compare state == i
			cmpVal := nextTemp(&counter)
			iStr := fmt.Sprintf("%d", i)
			iConst := nextTemp(&counter)
			nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: iConst, Value: iStr, Span: toIRSpan(path, statement.Span)})
			nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpVal, Operator: "==", Args: []string{stateVal, iConst}, Span: toIRSpan(path, statement.Span)})

			// Then branch for state i
			var thenBranch []ir.Instruction
			thenCounter := counter

			// Advance state to i + 1
			nextStateStr := fmt.Sprintf("%d", i+1)
			nextStateVal := nextTemp(&thenCounter)
			thenBranch = append(thenBranch, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: nextStateVal, Value: nextStateStr, Span: toIRSpan(path, statement.Span)})
			thenBranch = append(thenBranch, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     genClassName,
				Field:      "__state",
				FieldIndex: 0,
				Args:       []string{"this", nextStateVal},
				Span:       toIRSpan(path, statement.Span),
			})

			// Evaluate yield expression
			var valTemp string
			if y.expr != nil {
				// If identifier, check if parameter on this
				if y.expr.Kind == "identifier" {
					for fIdx, f := range genFields {
						if f.Name == y.expr.Text {
							valTemp = nextTemp(&thenCounter)
							thenBranch = append(thenBranch, ir.Instruction{
								Op:         ir.OpFieldGet,
								Type:       f.Type,
								Result:     valTemp,
								Callee:     genClassName,
								Field:      f.Name,
								FieldIndex: fIdx,
								Args:       []string{"this"},
								Span:       toIRSpan(path, y.expr.Span),
							})
							break
						}
					}
				}
				if valTemp == "" {
					var err error
					thenFn := &ir.Function{Body: thenBranch}
					valTemp, _, err = lowerExpression(path, y.expr, "", thenFn, nextEnv, &thenCounter, shapes, signatures)
					thenBranch = thenFn.Body
					if err != nil {
						valTemp = nextTemp(&thenCounter)
						thenBranch = append(thenBranch, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: valTemp, Value: "0", Span: toIRSpan(path, statement.Span)})
					}
				}
			} else {
				valTemp = nextTemp(&thenCounter)
				thenBranch = append(thenBranch, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: valTemp, Value: "0", Span: toIRSpan(path, statement.Span)})
			}

			// Store to this.__value
			thenBranch = append(thenBranch, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     genClassName,
				Field:      "__value",
				FieldIndex: 2,
				Args:       []string{"this", valTemp},
				Span:       toIRSpan(path, statement.Span),
			})

			// Construct IteratorResult: { value: valTemp, done: false }
			falseVal := nextTemp(&thenCounter)
			thenBranch = append(thenBranch, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: falseVal, Value: "false", Span: toIRSpan(path, statement.Span)})

			resObj := nextTemp(&thenCounter)
			thenBranch = append(thenBranch, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       ir.Type("object:" + resultShapeName),
				Result:     resObj,
				Callee:     resultShapeName,
				FieldCount: 2,
				Args:       []string{valTemp, falseVal},
				Span:       toIRSpan(path, statement.Span),
			})
			thenBranch = append(thenBranch, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     resultShapeName,
				Field:      "value",
				FieldIndex: 0,
				Args:       []string{resObj, valTemp},
				Span:       toIRSpan(path, statement.Span),
			})
			thenBranch = append(thenBranch, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     resultShapeName,
				Field:      "done",
				FieldIndex: 1,
				Args:       []string{resObj, falseVal},
				Span:       toIRSpan(path, statement.Span),
			})
			thenBranch = append(thenBranch, ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:" + resultShapeName), Args: []string{resObj}, Span: toIRSpan(path, statement.Span)})

			counter = thenCounter
			nextFn.Body = append(nextFn.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{cmpVal},
				Then: thenBranch,
				Span: toIRSpan(path, statement.Span),
			})
		}

		// State >= len(yields): finished, returns { value: this.__value, done: true }
		doneVal := nextTemp(&counter)
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       yieldType,
			Result:     doneVal,
			Callee:     genClassName,
			Field:      "__value",
			FieldIndex: 2,
			Args:       []string{"this"},
			Span:       toIRSpan(path, statement.Span),
		})
		trueVal := nextTemp(&counter)
		nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueVal, Value: "true", Span: toIRSpan(path, statement.Span)})
		doneResObj := nextTemp(&counter)
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpObjectNew,
			Type:       ir.Type("object:" + resultShapeName),
			Result:     doneResObj,
			Callee:     resultShapeName,
			FieldCount: 2,
			Args:       []string{doneVal, trueVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     resultShapeName,
			Field:      "value",
			FieldIndex: 0,
			Args:       []string{doneResObj, doneVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     resultShapeName,
			Field:      "done",
			FieldIndex: 1,
			Args:       []string{doneResObj, trueVal},
			Span:       toIRSpan(path, statement.Span),
		})
		nextFn.Body = append(nextFn.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:" + resultShapeName), Args: []string{doneResObj}, Span: toIRSpan(path, statement.Span)})
	}

	// 4. Build Factory Function: <name>(params...) -> Generator_<name>
	factoryFn := ir.Function{
		Name:       name,
		Span:       toIRSpan(path, statement.Span),
		ReturnType: ir.Type("object:" + genClassName),
	}
	factoryCounter := 0
	var factoryArgs []string

	// __state = 0, __done = false, __value = 0
	stateZero := nextTemp(&factoryCounter)
	factoryFn.Body = append(factoryFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: stateZero, Value: "0", Span: toIRSpan(path, statement.Span)})
	factoryArgs = append(factoryArgs, stateZero)

	doneFalse := nextTemp(&factoryCounter)
	factoryFn.Body = append(factoryFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: doneFalse, Value: "false", Span: toIRSpan(path, statement.Span)})
	factoryArgs = append(factoryArgs, doneFalse)

	defVal := nextTemp(&factoryCounter)
	factoryFn.Body = append(factoryFn.Body, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: defVal, Value: "0", Span: toIRSpan(path, statement.Span)})
	factoryArgs = append(factoryArgs, defVal)

	for _, p := range statement.Parameters {
		pType := toIRType(p.Type)
		if pType == "" && p.InferredType != "" {
			pType = toIRType(p.InferredType)
		}
		if pType == "" {
			pType = ir.TypeNumber
		}
		factoryFn.Parameters = append(factoryFn.Parameters, ir.Parameter{Name: p.Name, Type: pType})
		factoryArgs = append(factoryArgs, p.Name)
	}

	factoryEnv := map[string]ir.Type{}
	for _, p := range factoryFn.Parameters {
		factoryEnv[p.Name] = p.Type
	}

	itemsTemp := nextTemp(&factoryCounter)
	factoryFn.Body = append(factoryFn.Body, ir.Instruction{Op: ir.OpArray, Type: ir.Type(string(yieldType) + "[]"), Result: itemsTemp, Span: toIRSpan(path, statement.Span)})
	factoryArgs = append(factoryArgs, itemsTemp)
	factoryEnv["__items"] = ir.Type(string(yieldType) + "[]")

	factoryEnv[itemsTemp] = ir.Type(string(yieldType) + "[]")

	rewrittenBody := rewriteYieldsToPush(statement.Body, itemsTemp)
	for _, s := range rewrittenBody {
		_ = lowerStatement(path, s, &factoryFn, factoryEnv, &factoryCounter, shapes, signatures)
	}

	genObj := nextTemp(&factoryCounter)
	factoryFn.Body = append(factoryFn.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       ir.Type("object:" + genClassName),
		Result:     genObj,
		Callee:     genClassName,
		FieldCount: len(genFields),
		Args:       factoryArgs,
		Span:       toIRSpan(path, statement.Span),
	})

	// Set initial fields on genObj for interpreter
	for fIdx, f := range genFields {
		factoryFn.Body = append(factoryFn.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     genClassName,
			Field:      f.Name,
			FieldIndex: fIdx,
			Args:       []string{genObj, factoryArgs[fIdx]},
			Span:       toIRSpan(path, statement.Span),
		})
	}

	factoryFn.Body = append(factoryFn.Body, ir.Instruction{
		Op:   ir.OpReturn,
		Type: ir.Type("object:" + genClassName),
		Args: []string{genObj},
		Span: toIRSpan(path, statement.Span),
	})

	signatures[name] = factoryFn

	return factoryFn, []ir.Function{nextFn}, createdShapes, nil
}
