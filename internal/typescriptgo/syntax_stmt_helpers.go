package typescriptgo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func extractParameterBinding(p *ast.Node, pIdx int, chk *checker.Checker) (string, []SyntaxStatement) {
	nameNode := p.Name()
	if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
		return nameNode.Text(), nil
	}
	if nameNode != nil && (nameNode.Kind == ast.KindArrayBindingPattern || nameNode.Kind == ast.KindObjectBindingPattern) {
		pName := fmt.Sprintf("__param_%d", pIdx)
		c := 0
		stmts := flattenDestructuring(nameNode, &SyntaxExpression{
			Span: sourceSpan(nameNode),
			Kind: "identifier",
			Text: pName,
		}, chk, &c)
		return pName, stmts
	}
	if nameNode != nil {
		return fmt.Sprintf("__param_%d", pIdx), nil
	}
	return "", nil
}

func syntaxVariableDeclarations(decls []*ast.Node, span SourceSpan, chk *checker.Checker, varKind string) (SyntaxStatement, bool) {
	var stmts []SyntaxStatement
	destructCounter := 0
	for _, declaration := range decls {
		nameNode := declaration.Name()
		if declaration.Initializer() != nil && declaration.Initializer().Kind == ast.KindClassExpression {
			clsStmt, ok := syntaxClassDeclaration(declaration.Initializer(), sourceSpan(declaration), chk)
			if ok {
				clsStmt.Name = nameNode.Text()
				if clsStmt.Class != nil {
					clsStmt.Class.Name = nameNode.Text()
				}
				stmts = append(stmts, clsStmt)
				continue
			}
		}
		initExpr := syntaxExpression(declaration.Initializer(), chk)
		if nameNode.AsNode().Kind == ast.KindObjectBindingPattern || nameNode.AsNode().Kind == ast.KindArrayBindingPattern {
			nested := flattenDestructuring(nameNode.AsNode(), initExpr, chk, &destructCounter)
			stmts = append(stmts, nested...)
		} else {
			inferredVarType := resolveInferredType(chk, nameNode)
			if inferredVarType == "" {
				inferredVarType = resolveInferredType(chk, declaration)
			}
			varType := syntaxType(declaration.Type())
			if varType == "" && inferredVarType != "" {
				varType = inferredVarType
			}
			if initExpr != nil && varType != "" && !strings.HasPrefix(varType, "{") && (initExpr.InferredType == "" || strings.HasPrefix(initExpr.InferredType, "{")) {
				initExpr.InferredType = varType
			}
			kind := "variable"
			switch varKind {
			case "await_using":
				kind = "await_using"
			case "using":
				kind = "using"
			}
			result := SyntaxStatement{
				Span:         sourceSpan(declaration),
				Kind:         kind,
				VarDeclKind:  varKind,
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
	return evalConstNumberWithEnv(expr, nil)
}

func evalConstNumberWithEnv(expr *SyntaxExpression, env map[string]float64) (float64, bool) {
	if expr == nil {
		return 0, false
	}
	switch expr.Kind {
	case "number":
		v, err := strconv.ParseFloat(expr.Text, 64)
		return v, err == nil
	case "identifier":
		if env != nil {
			if v, ok := env[expr.Text]; ok {
				return v, true
			}
		}
	case "unary":
		v, ok := evalConstNumberWithEnv(expr.Left, env)
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
		lv, lok := evalConstNumberWithEnv(expr.Left, env)
		rv, rok := evalConstNumberWithEnv(expr.Right, env)
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
