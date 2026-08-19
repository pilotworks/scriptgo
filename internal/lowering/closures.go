package lowering

import (
	"fmt"
	"sort"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var extraFunctions []ir.Function
var closureCounter int

func findFreeVariables(fn *typescriptgo.SyntaxStatement, outerEnv map[string]ir.Type) []string {
	params := map[string]bool{}
	for _, p := range fn.Parameters {
		params[p.Name] = true
	}
	locals := map[string]bool{}
	var used []string

	var collectExpr func(e *typescriptgo.SyntaxExpression)
	var collectStmt func(s typescriptgo.SyntaxStatement)

	collectExpr = func(e *typescriptgo.SyntaxExpression) {
		if e == nil {
			return
		}
		if e.Kind == "identifier" {
			name := e.Text
			if !params[name] && !locals[name] {
				if _, ok := outerEnv[name]; ok {
					used = append(used, name)
				}
			}
		}
		if e.Left != nil {
			// If it's a property access obj.prop, only check Left if not method call on global
			collectExpr(e.Left)
		}
		if e.Right != nil {
			collectExpr(e.Right)
		}
		for _, arg := range e.Arguments {
			collectExpr(arg)
		}
	}

	collectStmt = func(s typescriptgo.SyntaxStatement) {
		if s.Kind == "variable" {
			locals[s.Name] = true
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "assign" {
			if !params[s.Name] && !locals[s.Name] {
				if _, ok := outerEnv[s.Name]; ok {
					used = append(used, s.Name)
				}
			}
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "return" || s.Kind == "throw" || s.Kind == "expression" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "if" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
			for _, st := range s.Then {
				collectStmt(st)
			}
			for _, st := range s.Else {
				collectStmt(st)
			}
		} else if s.Kind == "while" || s.Kind == "dowhile" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
			for _, st := range s.Body {
				collectStmt(st)
			}
		} else if s.Kind == "block" {
			for _, st := range s.Body {
				collectStmt(st)
			}
		}
	}

	for _, stmt := range fn.Body {
		collectStmt(stmt)
	}

	// Deduplicate and sort
	seen := map[string]bool{}
	var result []string
	for _, u := range used {
		if !seen[u] {
			seen[u] = true
			result = append(result, u)
		}
	}
	sort.Strings(result)
	return result
}

func lowerClosureExpression(
	path string,
	fnStmt *typescriptgo.SyntaxStatement,
	result string,
	callerFn *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	closureCounter++
	closureName := fmt.Sprintf("__closure_%d", closureCounter)

	capturedVars := findFreeVariables(fnStmt, env)

	// Create closure function
	targetFn := ir.Function{
		Name:       closureName,
		Span:       toIRSpan(path, fnStmt.Span),
		ReturnType: toIRType(fnStmt.Type),
	}
	if targetFn.ReturnType == "" {
		targetFn.ReturnType = ir.TypeVoid
	}

	// Always add __env_ctx as first parameter
	targetFn.Parameters = append(targetFn.Parameters, ir.Parameter{
		Name: "__env_ctx",
		Type: ir.Type("ptr"),
	})

	closureEnv := map[string]ir.Type{
		"__env_ctx": ir.Type("ptr"),
	}

	for _, p := range fnStmt.Parameters {
		typ := toIRType(p.Type)
		if typ == "" || typ == ir.TypeVoid {
			typ = ir.TypeNumber // default to number for untyped parameters
		}
		targetFn.Parameters = append(targetFn.Parameters, ir.Parameter{
			Name: p.Name,
			Type: typ,
		})
		closureEnv[p.Name] = typ
	}

	if targetFn.ReturnType == ir.TypeVoid || targetFn.ReturnType == "" {
		if len(fnStmt.Body) == 1 && fnStmt.Body[0].Kind == "return" && fnStmt.Body[0].Expression != nil {
			expr := fnStmt.Body[0].Expression
			switch expr.Kind {
			case "string", "template":
				targetFn.ReturnType = ir.TypeString
			case "bool", "compare":
				targetFn.ReturnType = ir.TypeBool
			case "binary":
				if isComparison(expr.Operator) || expr.Operator == "&&" || expr.Operator == "||" {
					targetFn.ReturnType = ir.TypeBool
				} else if expr.Operator == "+" && ((expr.Left != nil && expr.Left.Kind == "string") || (expr.Right != nil && expr.Right.Kind == "string")) {
					targetFn.ReturnType = ir.TypeString
				} else {
					targetFn.ReturnType = ir.TypeNumber
				}
			default:
				targetFn.ReturnType = ir.TypeNumber
			}
		}
	}

	// In closure body, captured variables are available from closureEnv
	for _, capVar := range capturedVars {
		closureEnv[capVar] = env[capVar]
		targetFn.Captured = append(targetFn.Captured, ir.Parameter{
			Name: capVar,
			Type: env[capVar],
		})
	}

	// Lower body of closure
	closureBodyCounter := 0
	returned := false
	for _, bodyStatement := range fnStmt.Body {
		if err := lowerStatement(path, bodyStatement, &targetFn, closureEnv, &closureBodyCounter, shapes, signatures); err != nil {
			return "", "", sourceError(path, bodyStatement.Span, err)
		}
		if statementAlwaysReturns(bodyStatement) {
			returned = true
		}
	}
	if !returned {
		if targetFn.ReturnType != ir.TypeVoid {
			// If not returned and return type is not void, return default or error
			targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpReturn, Type: targetFn.ReturnType, Span: targetFn.Span})
		} else {
			targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: targetFn.Span})
		}
	}

	extraFunctions = append(extraFunctions, targetFn)
	signatures[closureName] = targetFn

	if result == "" {
		result = nextTemp(counter)
	}
	signatures[result] = targetFn
	env[result+".retType"] = targetFn.ReturnType

	callerFn.Body = append(callerFn.Body, ir.Instruction{
		Op:     ir.OpClosure,
		Type:   ir.TypeClosure,
		Result: result,
		Callee: closureName,
		Args:   capturedVars,
		Span:   toIRSpan(path, fnStmt.Span),
	})

	return result, ir.TypeClosure, nil
}
