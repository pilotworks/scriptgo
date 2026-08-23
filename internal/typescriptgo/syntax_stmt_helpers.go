package typescriptgo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxVariableDeclarations(decls []*ast.Node, span SourceSpan, chk *checker.Checker, isUsing, isAwaitUsing bool) (SyntaxStatement, bool) {
	var stmts []SyntaxStatement
	for _, declaration := range decls {
		nameNode := declaration.Name()
		initExpr := syntaxExpression(declaration.Initializer(), chk)
		if nameNode.AsNode().Kind == ast.KindObjectBindingPattern {
			pattern := nameNode.AsNode().AsBindingPattern()
			objVar := fmt.Sprintf("__destruct_obj_%d", nameNode.AsNode().Pos())
			if initExpr != nil && initExpr.Kind == "identifier" {
				objVar = initExpr.Text
			} else {
				stmts = append(stmts, SyntaxStatement{
					Span:       sourceSpan(nameNode.AsNode()),
					Kind:       "variable",
					Name:       objVar,
					Expression: initExpr,
				})
			}
			for _, elem := range pattern.Elements.Nodes {
				binding := elem.AsBindingElement()
				propName := binding.Name().Text()
				varName := binding.Name().Text()
				if binding.PropertyName != nil {
					propName = binding.PropertyName.Text()
				}
				propExpr := &SyntaxExpression{
					Span:         sourceSpan(elem),
					Kind:         "property",
					Text:         propName,
					InferredType: resolveInferredType(chk, elem),
					Left: &SyntaxExpression{
						Span: sourceSpan(elem),
						Kind: "identifier",
						Text: objVar,
					},
				}
				stmts = append(stmts, SyntaxStatement{
					Span:         sourceSpan(elem),
					Kind:         "variable",
					Name:         varName,
					InferredType: resolveInferredType(chk, elem),
					Expression:   propExpr,
				})
			}
		} else if nameNode.AsNode().Kind == ast.KindArrayBindingPattern {
			pattern := nameNode.AsNode().AsBindingPattern()
			arrVar := fmt.Sprintf("__destruct_arr_%d", nameNode.AsNode().Pos())
			if initExpr != nil && initExpr.Kind == "identifier" {
				arrVar = initExpr.Text
			} else {
				stmts = append(stmts, SyntaxStatement{
					Span:       sourceSpan(nameNode.AsNode()),
					Kind:       "variable",
					Name:       arrVar,
					Expression: initExpr,
				})
			}
			for idx, elem := range pattern.Elements.Nodes {
				if elem.Kind == ast.KindOmittedExpression {
					continue
				}
				binding := elem.AsBindingElement()
				varName := binding.Name().Text()
				if binding.DotDotDotToken != nil {
					sliceExpr := &SyntaxExpression{
						Span:         sourceSpan(elem),
						Kind:         "call",
						InferredType: resolveInferredType(chk, elem),
						Left: &SyntaxExpression{
							Span: sourceSpan(elem),
							Kind: "property",
							Text: "slice",
							Left: &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						},
						Arguments: []*SyntaxExpression{
							{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
							{Span: sourceSpan(elem), Kind: "property", Text: "length", Left: &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar}, InferredType: "number"},
						},
					}
					stmts = append(stmts, SyntaxStatement{
						Span:         sourceSpan(elem),
						Kind:         "variable",
						Name:         varName,
						InferredType: resolveInferredType(chk, elem),
						Expression:   sliceExpr,
					})
				} else {
					idxExpr := &SyntaxExpression{
						Span:         sourceSpan(elem),
						Kind:         "index",
						InferredType: resolveInferredType(chk, elem),
						Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						Right:        &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
					}
					stmts = append(stmts, SyntaxStatement{
						Span:         sourceSpan(elem),
						Kind:         "variable",
						Name:         varName,
						InferredType: resolveInferredType(chk, elem),
						Expression:   idxExpr,
					})
				}
			}
		} else {
			inferredVarType := resolveInferredType(chk, nameNode)
			if inferredVarType == "" {
				inferredVarType = resolveInferredType(chk, declaration)
			}
			varType := syntaxType(declaration.Type())
			if varType == "" && inferredVarType != "" {
				varType = inferredVarType
			}
			kind := "variable"
			if isAwaitUsing {
				kind = "await_using"
			} else if isUsing {
				kind = "using"
			}
			result := SyntaxStatement{
				Span:         sourceSpan(declaration),
				Kind:         kind,
				Name:         nameNode.Text(),
				Type:         varType,
				InferredType: inferredVarType,
				Expression:   initExpr,
			}
			stmts = append(stmts, result)
		}
	}
	if len(stmts) == 1 {
		return stmts[0], true
	}
	return SyntaxStatement{
		Span: span,
		Kind: "block",
		Body: stmts,
	}, true
}

func syntaxBlockStatements(node *ast.Node, chk *checker.Checker) []SyntaxStatement {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindBlock {
		result := make([]SyntaxStatement, 0, len(node.Statements()))
		for _, statement := range node.Statements() {
			if converted, ok := syntaxStatement(statement, chk); ok {
				result = append(result, converted)
			}
		}
		return result
	}
	if converted, ok := syntaxStatement(node, chk); ok {
		return []SyntaxStatement{converted}
	}
	return nil
}

func isAssignmentOperator(op string) bool {
	if op == "=" {
		return true
	}
	if strings.HasSuffix(op, "=") && op != "==" && op != "===" && op != "!=" && op != "!==" && op != "<=" && op != ">=" {
		return true
	}
	return false
}

func desugarAssignment(expr *SyntaxExpression) (*SyntaxExpression, bool) {
	if expr == nil || expr.Kind != "binary" {
		return nil, false
	}
	if expr.Operator == "=" {
		return expr.Right, true
	}
	if isAssignmentOperator(expr.Operator) {
		baseOp := strings.TrimSuffix(expr.Operator, "=")
		return &SyntaxExpression{
			Span:     expr.Span,
			Kind:     "binary",
			Operator: baseOp,
			Left:     expr.Left,
			Right:    expr.Right,
		}, true
	}
	return nil, false
}

func evalConstNumber(expr *SyntaxExpression) (float64, bool) {
	if expr == nil {
		return 0, false
	}
	switch expr.Kind {
	case "number":
		v, err := strconv.ParseFloat(expr.Text, 64)
		return v, err == nil
	case "unary":
		v, ok := evalConstNumber(expr.Left)
		if !ok {
			return 0, false
		}
		switch expr.Operator {
		case "-":
			return -v, true
		case "+":
			return v, true
		case "~":
			return float64(^int64(v)), true
		}
	case "binary":
		lv, lok := evalConstNumber(expr.Left)
		rv, rok := evalConstNumber(expr.Right)
		if !lok || !rok {
			return 0, false
		}
		li := int64(lv)
		ri := int64(rv)
		switch expr.Operator {
		case "+":
			return lv + rv, true
		case "-":
			return lv - rv, true
		case "*":
			return lv * rv, true
		case "/":
			if rv != 0 {
				return lv / rv, true
			}
		case "%":
			if ri != 0 {
				return float64(li % ri), true
			}
		case "<<":
			return float64(li << uint64(ri)), true
		case ">>":
			return float64(li >> uint64(ri)), true
		case ">>>":
			return float64(uint64(uint32(li)) >> uint64(ri)), true
		case "|":
			return float64(li | ri), true
		case "&":
			return float64(li & ri), true
		case "^":
			return float64(li ^ ri), true
		}
	}
	return 0, false
}
