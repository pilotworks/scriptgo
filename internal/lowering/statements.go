package lowering

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerFunction(path string, statement frontend.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	if statement.IsAsync || statement.Kind == "async_function" {
		return lowerAsyncFunction(path, statement, shapes, signatures)
	}
	return lowerSyncFunction(path, statement, shapes, signatures)
}

func lowerSyncFunction(path string, statement frontend.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	savedUsingScopes := usingScopeStack
	usingScopeStack = nil
	defer func() {
		usingScopeStack = savedUsingScopes
	}()

	retType := statement.Type
	if retType == "" && statement.InferredType != "" {
		retType = statement.InferredType
	}
	function := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), ReturnType: toIRTypeForPath(path, retType)}
	if function.ReturnType == "" {
		function.ReturnType = ir.TypeVoid
	}
	env := map[string]ir.Type{}
	if resolvedType, ok := asyncResolvedReturnType(retType); ok {
		env[asyncResolvedReturnTypeEnvKey] = resolvedType
	}
	for _, parameter := range statement.Parameters {
		pType := parameter.Type
		if pType == "" && parameter.InferredType != "" {
			pType = parameter.InferredType
		}
		typ := toIRTypeForPath(path, pType)
		if parameter.Rest {
			if typ == "" || typ == ir.TypeUnknown {
				if pType == "number[]" {
					typ = ir.TypeNumberArray
				} else {
					typ = ir.TypeStringArray
				}
			}
		}
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		typ = variableStorageType(typ)
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		if typ == ir.TypeObject && pType != "" && pType != "object" {
			env[parameter.Name] = ir.Type("object:" + pType)
		} else {
			env[parameter.Name] = typ
		}
		// Keep the source-level declaration alongside the storage type. Mixed
		// unions may use unknown storage until a control-flow guard narrows them.
		env["__decl_str."+parameter.Name] = ir.Type(pType)
		env["__param."+parameter.Name] = typ
		env["__storage_type."+parameter.Name] = typ
		fnSig := parameter.Type
		if fnSig == "" || fnSig == "closure" {
			fnSig = parameter.InferredType
		}
		if strings.Contains(fnSig, "=>") {
			retStr := extractTopLevelReturnType(fnSig)
			env[parameter.Name+".retType"] = toIRType(retStr)
		}
	}
	counter := 0
	returned := false
	for _, bodyStatement := range statement.Body {
		if err := lowerStatement(path, bodyStatement, &function, env, &counter, shapes, signatures); err != nil {
			return ir.Function{}, sourceError(path, bodyStatement.Span, err)
		}
		if statementAlwaysReturns(bodyStatement) {
			returned = true
		}
	}
	if !returned {
		if function.ReturnType == ir.TypeVoid {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: function.Span})
		} else if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
			prom := appendResolvedPromiseUndefined(&function, &counter, function.Span)
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{prom},
				Span: function.Span,
			})
		} else {
			defVal := ""
			if function.ReturnType == ir.TypeNumber {
				defVal = "0"
			} else if function.ReturnType == ir.TypeBool {
				defVal = "false"
			} else if strings.HasPrefix(string(function.ReturnType), "object:") || function.ReturnType == "ptr" {
				defVal = "0"
			}
			defTemp := nextTemp(&counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   function.ReturnType,
				Result: defTemp,
				Value:  defVal,
				Span:   function.Span,
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{defTemp},
				Span: function.Span,
			})
		}
	}
	return function, nil
}

func lowerStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	switch statement.Kind {
	case "empty":
		return nil
	case "variable", "using", "await_using":
		return lowerVariableStatement(path, statement, function, env, counter, shapes, signatures)

	case "expression":
		if statement.Expression == nil {
			return fmt.Errorf("empty expression")
		}
		_, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		return err
	case "return":
		return lowerReturnStatement(path, statement, function, env, counter, shapes, signatures)

	case "block":
		blockEnv := maps.Clone(env)
		pushUsingScope()
		for _, s := range statement.Body {
			if err := lowerStatement(path, s, function, blockEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}
		popAndEmitUsingScope(path, function, counter, shapes, signatures)
	case "namespace":
		return lowerNamespaceStatement(path, statement, function, env, counter, shapes, signatures)

	case "assign":
		return lowerAssignStatement(path, statement, function, env, counter, shapes, signatures)

	case "while":
		return lowerWhile(path, statement, function, env, counter, shapes, signatures)
	case "if":
		return lowerIf(path, statement, function, env, counter, shapes, signatures)
	case "break":
		if err := lowerActiveBreakFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBreak, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "continue":
		if err := lowerActiveBreakFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpContinue, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "dowhile":
		return lowerDoWhile(path, statement, function, env, counter, shapes, signatures)
	case "forof", "forawaitof":
		return lowerForOf(path, statement, function, env, counter, shapes, signatures)
	case "forin":
		return lowerForIn(path, statement, function, env, counter, shapes, signatures)
	case "label":
		return lowerLabel(path, statement, function, env, counter, shapes, signatures)
	case "switch":
		return lowerSwitch(path, statement, function, env, counter, shapes, signatures)
	case "function":
		_, typ, err := lowerClosureExpression(path, &statement, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		env[statement.Name] = typ
		return nil
	case "index_set":
		return lowerIndexSetStatement(path, statement, function, env, counter, shapes, signatures)

	case "field_set":
		return lowerFieldSetStatement(path, statement, function, env, counter, shapes, signatures)

	case "class":
		return nil
	case "throw":
		return lowerThrowStatement(path, statement, function, env, counter, shapes, signatures)

	case "try":
		return lowerTry(path, statement, function, env, counter, shapes, signatures)
	case "debugger":
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpDebugger,
			Type: ir.TypeVoid,
			Span: toIRSpan(path, statement.Span),
		})
	case "import_alias":
		return lowerImportAliasStatement(path, statement, env)

	case "export_alias", "module", "enum":
		return nil
	default:
		return fmt.Errorf("unsupported statement %q", statement.Kind)
	}
	return nil
}

func lowerActiveReturnFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	savedStack := activeReturnFinallyStack
	activeReturnFinallyStack = nil
	defer func() {
		activeReturnFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= 0; i-- {
		activeReturnFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerActiveThrowFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	savedStack := activeThrowFinallyStack
	activeThrowFinallyStack = nil
	defer func() {
		activeThrowFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= 0; i-- {
		activeThrowFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerActiveBreakFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	minDepth := 0
	if len(loopFinallyScopeStack) > 0 {
		minDepth = loopFinallyScopeStack[len(loopFinallyScopeStack)-1]
	}
	savedStack := activeReturnFinallyStack
	activeReturnFinallyStack = nil
	defer func() {
		activeReturnFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= minDepth; i-- {
		activeReturnFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerBranch(path string, statements []frontend.SyntaxStatement, returnType ir.Type, parentEnv map[string]ir.Type, parent *ir.Function, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]ir.Instruction, error) {
	branch := ir.Function{Name: "branch", ReturnType: returnType}
	env := make(map[string]ir.Type, len(parentEnv))
	maps.Copy(env, parentEnv)
	for _, statement := range statements {
		if err := lowerStatement(path, statement, &branch, env, counter, shapes, signatures); err != nil {
			return nil, err
		}
	}
	for _, local := range branch.Locals {
		found := false
		for _, existing := range parent.Locals {
			if existing.Name == local.Name {
				found = true
				break
			}
		}
		if !found {
			parent.Locals = append(parent.Locals, local)
		}
	}
	return branch.Body, nil
}

func isOptionalChainExpr(expr *frontend.SyntaxExpression) bool {
	if expr == nil {
		return false
	}
	if expr.Kind == "optional_call" || expr.Kind == "optional_property" || expr.Kind == "optional_index" {
		return true
	}
	return isOptionalChainExpr(expr.Left) || isOptionalChainExpr(expr.Right)
}
