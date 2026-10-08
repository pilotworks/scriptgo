package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerReturnStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if statement.Expression == nil {
		span := toIRSpan(path, statement.Span)
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
			// A bare return in an async function fulfills its promise with
			// undefined; it does not return a void LLVM value directly.
			promiseVal := appendResolvedPromiseUndefined(function, counter, span)
			emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
			bodyLenBeforeFinally := len(function.Body)
			if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) == bodyLenBeforeFinally || function.Body[len(function.Body)-1].Op != ir.OpReturn {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{promiseVal}, Span: span})
			}
			return nil
		}
		emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
		bodyLenBeforeFinally := len(function.Body)
		if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) == bodyLenBeforeFinally || function.Body[len(function.Body)-1].Op != ir.OpReturn {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: span})
		}
		return nil
	}
	if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeUnknown {
		res := nextTemp(counter)
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
			span := toIRSpan(path, statement.Span)
			prom := ""
			if statement.Expression.Kind == "null" {
				prom = appendResolvedPromiseNull(function, counter, span)
			} else {
				prom = appendResolvedPromiseUndefined(function, counter, span)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{prom}, Span: span})
			return nil
		}
		defaultVal := "0"
		if function.ReturnType == ir.TypeBool {
			defaultVal = "false"
		} else if function.ReturnType == ir.TypeNumber {
			defaultVal = "0"
		} else if statement.Expression.Kind == "undefined" {
			defaultVal = "undefined"
		} else if statement.Expression.Kind == "null" || strings.HasPrefix(string(function.ReturnType), "object:") || isPointerLikeType(function.ReturnType) || function.ReturnType == ir.TypeString {
			defaultVal = "null"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   function.ReturnType,
			Result: res,
			Value:  defaultVal,
			Span:   toIRSpan(path, statement.Span),
		})
		bodyLenBeforeFinally := len(function.Body)
		if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{res}, Span: toIRSpan(path, statement.Span)})
		return nil
	}
	returnExpression := statement.Expression
	expectedReturnType := function.ReturnType
	if resolvedType := env[asyncResolvedReturnTypeEnvKey]; resolvedType != "" {
		expectedReturnType = resolvedType
	}
	if returnExpression.Kind == "object_literal" && strings.HasPrefix(string(expectedReturnType), "object:") && !strings.Contains(string(expectedReturnType), "|") {
		cleanRet := strings.TrimPrefix(string(expectedReturnType), "object:")
		if aliased, ok := typeAliasesIndex[cleanRet]; !ok || !strings.Contains(aliased, "|") {
			// Contextual typing is local to this lowering pass. The frontend AST
			// can be reused by specializations and repeated Lower calls.
			returnExpression = cloneAndSubstituteExpr(returnExpression, nil)
			returnExpression.InferredType = string(expectedReturnType)
		}
	} else if returnExpression.Kind == "array" && strings.HasPrefix(string(expectedReturnType), "object:") {
		// A tuple return type may contain nullable or otherwise heterogeneous
		// elements. Contextualize the array so both branches use one slot layout.
		shapeName := strings.TrimPrefix(string(expectedReturnType), "object:")
		if isTupleShapeName(shapeName) {
			returnExpression = cloneAndSubstituteExpr(returnExpression, nil)
			returnExpression.InferredType = string(expectedReturnType)
		}
	}
	value, typ, err := lowerExpression(path, returnExpression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
		prom := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Promise"),
			Result: prom,
			Callee: "__async.promise_resolve",
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		value = prom
		typ = function.ReturnType
	}
	if function.ReturnType == ir.TypeUnknown && typ != ir.TypeUnknown {
		boxed := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: boxed,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		value = boxed
		typ = ir.TypeUnknown
	} else if function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeUnknown && typ == ir.TypeUnknown {
		castVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   function.ReturnType,
			Result: castVal,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		value = castVal
		typ = function.ReturnType
	}
	if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
		prom := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   function.ReturnType,
			Result: prom,
			Callee: "__async.promise_resolve",
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		value = prom
		typ = function.ReturnType
	} else if strings.HasPrefix(string(function.ReturnType), "object:") && !strings.Contains(string(function.ReturnType), "{") && strings.HasPrefix(string(typ), "object:") {
		typ = function.ReturnType
	} else if function.ReturnType == "ptr" && (typ == ir.TypeString || isPointerLikeType(typ)) {
		typ = "ptr"
	} else if function.ReturnType == ir.TypeString && typ == "ptr" {
		typ = ir.TypeString
	} else if function.ReturnType != "" && function.ReturnType != ir.TypeVoid && isPointerLikeType(function.ReturnType) && (typ == "ptr" || isPointerLikeType(typ)) {
		typ = function.ReturnType
	} else if function.ReturnType == "" || function.ReturnType == ir.TypeVoid {
		function.ReturnType = typ
		if signatures != nil {
			if sig, ok := signatures[function.Name]; ok {
				sig.ReturnType = typ
				signatures[function.Name] = sig
			}
		}
	}
	if len(activeReturnFinallyStack) > 0 && value != "" && typ != ir.TypeVoid {
		savedVal := nextTemp(counter)
		// Return values that must survive a finally clause need a real
		// storage slot; OpAssign is a variable write, not an SSA definition.
		function.Locals = append(function.Locals, ir.Parameter{Name: savedVal, Type: typ})
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   typ,
			Result: savedVal,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		value = savedVal
	}
	emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
	bodyLenBeforeFinally := len(function.Body)
	if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
		return err
	}
	if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
		return nil
	}
	if typ == ir.TypeVoid || value == "" {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
	} else {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	}
	return nil
}

func lowerThrowStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	bodyLenBeforeFinally := len(function.Body)
	if err := lowerActiveThrowFinally(path, function, env, counter, shapes, signatures); err != nil {
		return err
	}
	if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
		return nil
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpThrow,
		Type: ir.TypeVoid,
		Args: []string{val},
		Span: toIRSpan(path, statement.Span),
	})
	return nil
}
