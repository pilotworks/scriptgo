package lowering

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerVariableStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	varResultName := statement.Name
	if _, isShadowed := env[statement.Name]; isShadowed {
		varResultName = fmt.Sprintf("%s$%d", statement.Name, *counter)
		*counter++
		env["__ident."+statement.Name] = ir.Type(varResultName)
	}
	localType := toIRTypeForPath(path, statement.Type)
	if localType == "" {
		localType = toIRTypeForPath(path, statement.InferredType)
	}
	if localType == "" {
		localType = ir.TypeUnknown
	}
	function.Locals = append(function.Locals, ir.Parameter{Name: varResultName, Type: localType})
	declTypeStr := statement.Type
	if declTypeStr == "" {
		declTypeStr = statement.InferredType
	}
	if declTypeStr != "" {
		env["__decl_str."+varResultName] = ir.Type(declTypeStr)
		env["__decl_str."+statement.Name] = ir.Type(declTypeStr)
	}
	env["__storage_type."+varResultName] = localType
	env["__storage_type."+statement.Name] = localType
	if statement.Expression == nil {
		if statement.Kind == "using" || statement.Kind == "await_using" {
			return fmt.Errorf("resource %q has no initializer", statement.Name)
		}
		typeStr := statement.Type
		if typeStr == "" {
			typeStr = statement.InferredType
		}
		declaredType := variableStorageType(toIRType(typeStr))
		if declaredType == "" {
			declaredType = ir.TypeUnknown
		}
		env[varResultName] = declaredType
		env[statement.Name] = declaredType
		env["__storage_type."+varResultName] = declaredType
		env["__decl_str."+varResultName] = ir.Type(typeStr)
		var zeroVal string
		hasUndef := strings.Contains(typeStr, "undefined") || strings.Contains(typeStr, "void")
		switch declaredType {
		case ir.TypeNumber:
			zeroVal = "0"
		case ir.TypeBool:
			zeroVal = "false"
		case ir.TypeBigInt:
			zeroVal = "0"
		case ir.TypeString:
			if hasUndef {
				zeroVal = "undefined"
			} else {
				zeroVal = ""
			}
		case ir.TypeUnknown:
			zeroVal = "undefined"
		default:
			if hasUndef {
				zeroVal = "undefined"
			} else {
				zeroVal = "null"
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   declaredType,
			Result: varResultName,
			Value:  zeroVal,
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	inProgressVars[statement.Name] = true
	defer delete(inProgressVars, statement.Name)
	declaredType := toIRTypeForPath(path, statement.Type)
	if statement.Type == "" && statement.InferredType != "" {
		declaredType = toIRTypeForPath(path, statement.InferredType)
	}
	declaredType = variableStorageType(declaredType)
	if statement.Expression.Kind == "identifier" {
		if _, isDynamicAlias := dynamicImports[statement.Expression.Text]; isDynamicAlias {
			env[varResultName] = ir.TypeUnknown
			env[statement.Name] = ir.TypeUnknown
			env["__storage_type."+varResultName] = ir.TypeUnknown
			return nil
		}
	}
	if statement.Expression.Kind == "identifier" || statement.Expression.Kind == "this" {
		identText := statement.Expression.Text
		if statement.Expression.Kind == "this" {
			identText = "this"
		}
		if mangled, hasMangled := env["__ident."+identText]; hasMangled {
			identText = string(mangled)
		}
		srcType, ok := env[identText]
		if ok && (declaredType == "" || declaredType == srcType || declaredType == ir.TypeUnknown) {
			if declaredType == ir.TypeUnknown {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Args:   []string{identText},
					Span:   toIRSpan(path, statement.Span),
				})
				env[varResultName] = ir.TypeUnknown
				env[statement.Name] = ir.TypeUnknown
				return nil
			}
			env[varResultName] = srcType
			env[statement.Name] = srcType
			switch srcType {
			case ir.TypeNumber:
				// x - 0 copies every IEEE value unchanged, including -0
				// (x + 0 would turn -0 into +0).
				zeroConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "-", Args: []string{identText, zeroConst}, Span: toIRSpan(path, statement.Span)})
			case ir.TypeString:
				emptyStr := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "+", Args: []string{identText, emptyStr}, Span: toIRSpan(path, statement.Span)})
			case ir.TypeBool:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "||", Args: []string{identText, identText}, Span: toIRSpan(path, statement.Span)})
			default:
				trueConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: srcType, Result: varResultName, Args: []string{trueConst, identText, identText}, Span: toIRSpan(path, statement.Span)})
			}
			return nil
		}
	}
	if declaredType == ir.TypeUnknown {
		if statement.Expression != nil && statement.Expression.Kind == "null" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeUnknown,
				Result: varResultName,
				Value:  "null",
				Span:   toIRSpan(path, statement.Span),
			})
			env[varResultName] = ir.TypeUnknown
			env[statement.Name] = ir.TypeUnknown
			return nil
		}
		if statement.Expression != nil && statement.Expression.Kind == "undefined" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeUnknown,
				Result: varResultName,
				Value:  "undefined",
				Span:   toIRSpan(path, statement.Span),
			})
			env[varResultName] = ir.TypeUnknown
			env[statement.Name] = ir.TypeUnknown
			return nil
		}
		value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if valType == ir.TypeUnknown {
			if value != varResultName {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Value:  "undefined",
					Span:   toIRSpan(path, statement.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Args:   []string{value},
					Span:   toIRSpan(path, statement.Span),
				})
			}
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: varResultName,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
		}
		env[varResultName] = ir.TypeUnknown
		env[statement.Name] = ir.TypeUnknown
		return nil
	}
	if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && declaredType != "" && declaredType != ir.TypeVoid && declaredType != ir.TypeUnknown {
		defaultVal := "0"
		if declaredType == ir.TypeNumber {
			defaultVal = "NaN"
		} else if declaredType == ir.TypeBool {
			defaultVal = "false"
		} else if declaredType == ir.TypeString {
			if statement.Expression.Kind == "undefined" {
				defaultVal = "undefined"
			} else {
				defaultVal = "null"
			}
		} else if strings.HasPrefix(string(declaredType), "object:") || declaredType == ir.TypePointer {
			if statement.Expression.Kind == "undefined" {
				defaultVal = "undefined"
			} else {
				defaultVal = "null"
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   declaredType,
			Result: varResultName,
			Value:  defaultVal,
			Span:   toIRSpan(path, statement.Span),
		})
		env[varResultName] = declaredType
		env[statement.Name] = declaredType
		return nil
	}
	if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) && statement.Type != "" {
		if toIRType(statement.Type) == ir.TypeUnknownArray {
			statement.Expression.InferredType = "unknown[]"
		} else if _, isTup := tupleFields(statement.Type); isTup {
			statement.Expression.InferredType = statement.Type
		} else if strings.HasSuffix(statement.Type, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" {
			statement.Expression.InferredType = statement.Type
		}
	}
	if statement.Expression != nil && statement.Expression.Kind == "object_literal" && statement.Type != "" {
		statement.Expression.InferredType = statement.Type
	}
	if statement.Expression != nil && (statement.Expression.Kind == "function" || statement.Expression.Function != nil || strings.Contains(statement.Type, "=>") || strings.Contains(statement.InferredType, "=>")) {
		env[varResultName] = ir.TypeClosure
		env[statement.Name] = ir.TypeClosure
		fnSig := statement.Type
		if fnSig == "" || fnSig == "closure" {
			fnSig = statement.InferredType
		}
		if strings.Contains(fnSig, "=>") {
			retStr := extractTopLevelReturnType(fnSig)
			env[varResultName+".retType"] = toIRType(retStr)
			env[statement.Name+".retType"] = toIRType(retStr)
		}
	}
	if statement.Type != "" {
		env["__decl_str."+varResultName] = ir.Type(statement.Type)
		env["__decl_str."+statement.Name] = ir.Type(statement.Type)
	}
	// Intrinsics may honor the requested declaration name and emit an
	// assignment while lowering (for example JSON.stringify). Predeclare a
	// known inferred slot so that assignment validation sees the declaration.
	if declaredType == "" {
		// The checker may omit an expression's return type for a builtin
		// call. Use boxed storage during lowering; the actual value type is
		// installed immediately after the expression is lowered.
		declaredType = ir.TypeUnknown
	}
	env[varResultName] = declaredType
	env[statement.Name] = declaredType
	env["__storage_type."+varResultName] = declaredType
	value, valType, err := lowerExpression(path, statement.Expression, varResultName, function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if env[value+".dynamic"] != "" {
		env[varResultName+".dynamic"] = env[value+".dynamic"]
		env[statement.Name+".dynamic"] = env[value+".dynamic"]
	}
	if declaredType != "" && valType == ir.TypeUnknown && declaredType != ir.TypeUnknown && !isOptionalChainExpr(statement.Expression) {
		if value == varResultName && statement.Expression.Kind != "conditional" {
			tempVal := nextTemp(counter)
			for i := len(function.Body) - 1; i >= 0; i-- {
				if function.Body[i].Result == varResultName {
					function.Body[i].Result = tempVal
					break
				}
			}
			value = tempVal
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   declaredType,
			Result: varResultName,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		valType = declaredType
	} else if value != varResultName {
		assignType := declaredType
		if assignType == "" || assignType == ir.TypeUnknown {
			assignType = valType
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   assignType,
			Result: varResultName,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
	}
	typ := valType
	if declaredType != "" && declaredType != ir.TypeUnknown {
		if valType == ir.TypeDynamicFunction && declaredType == ir.TypeClosure {
			// Keep an engine function handle distinct from native closures even
			// when TypeScript declares the variable with a function signature.
			typ = ir.TypeDynamicFunction
		} else if !(strings.HasPrefix(string(valType), "object:") && strings.HasPrefix(string(declaredType), "object:") && !strings.Contains(string(declaredType), "shape_") && !strings.Contains(string(declaredType), "{")) {
			typ = declaredType
		}
	}
	env[varResultName] = typ
	env[statement.Name] = typ
	if typ == ir.TypeClosure {
		fnSig := statement.Type
		if fnSig == "" || fnSig == "closure" {
			fnSig = statement.InferredType
		}
		if strings.Contains(fnSig, "=>") {
			retStr := extractTopLevelReturnType(fnSig)
			env[varResultName+".retType"] = toIRType(retStr)
			env[statement.Name+".retType"] = toIRType(retStr)
		}
	}
	if statement.Kind == "using" {
		recordUsingResource(varResultName, typ, false, statement.Span)
	} else if statement.Kind == "await_using" {
		recordUsingResource(varResultName, typ, true, statement.Span)
	}
	return nil
}

func lowerNamespaceStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	for _, s := range statement.Body {
		if s.Kind == "variable" || s.Kind == "using" || s.Kind == "await_using" {
			varCopy := s
			varCopy.Name = statement.Name + "." + s.Name
			if err := lowerStatement(path, varCopy, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		} else {
			if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	}
	return nil
}

func lowerImportAliasStatement(path string, statement frontend.SyntaxStatement, env map[string]ir.Type) error {
	if statement.Name != "" && statement.Type != "" {
		cleanFile := filepath.Clean(path)
		if _, isFn := functionImportsByFile[cleanFile][statement.Name]; isFn {
			return nil
		}
		if _, isCls := classImportsByFile[cleanFile][statement.Name]; isCls {
			return nil
		}
		env["__ident."+statement.Name] = ir.Type(statement.Type)
		if origType, ok := env[statement.Type]; ok {
			env[statement.Name] = origType
		}
	}
	return nil
}

// variableStorageType maps a declared variable type to the IR type used for
// its storage. A variable typed void/undefined/never still holds a value
// (undefined), so it is stored boxed rather than as an unsized void slot.
func variableStorageType(declared ir.Type) ir.Type {
	if declared == ir.TypeVoid {
		return ir.TypeUnknown
	}
	return declared
}
