package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerClassDeclaration(fileName string, statement frontend.SyntaxStatement, module *ir.Module, hierarchy map[string]ClassMeta, shapes map[string]ir.ObjectShape, main *ir.Function, env map[string]ir.Type, signatures map[string]ir.Function, counter *int) error {
	if statement.Class == nil || len(statement.Class.TypeParameters) > 0 {
		return nil
	}
	className := classIdentityForPath(fileName, statement.Class.Name)
	var fieldInits []frontend.SyntaxStatement
	for _, f := range statement.Class.Fields {
		if !f.IsStatic && f.Initializer != nil {
			fieldInits = append(fieldInits, frontend.SyntaxStatement{
				Span: f.Span,
				Kind: "field_set",
				Name: f.Name,
				Left: &frontend.SyntaxExpression{
					Span: f.Span,
					Kind: "identifier",
					Text: "this",
				},
				Expression: f.Initializer,
			})
		}
	}

	// Lower constructor if present
	if statement.Class.Constructor != nil {
		ctorMangled := className + "_constructor"
		var ctorBody []frontend.SyntaxStatement
		if len(statement.Class.Constructor.Body) > 0 && statement.Class.Constructor.Body[0].Expression != nil && statement.Class.Constructor.Body[0].Expression.Kind == "call" && statement.Class.Constructor.Body[0].Expression.Left != nil && statement.Class.Constructor.Body[0].Expression.Left.Text == "super" {
			ctorBody = append(ctorBody, statement.Class.Constructor.Body[0])
			ctorBody = append(ctorBody, fieldInits...)
			ctorBody = append(ctorBody, statement.Class.Constructor.Body[1:]...)
		} else {
			ctorBody = append(ctorBody, fieldInits...)
			ctorBody = append(ctorBody, statement.Class.Constructor.Body...)
		}
		ctorStmt := frontend.SyntaxStatement{
			Span: statement.Class.Constructor.Span,
			Kind: "function",
			Name: ctorMangled,
			Type: "void",
			Parameters: append([]frontend.SyntaxParameter{
				{Name: "this", Type: "object:" + className},
			}, statement.Class.Constructor.Parameters...),
			Body: ctorBody,
		}
		function, err := lowerFunction(fileName, ctorStmt, shapes, signatures)
		if err != nil {
			return fmt.Errorf("lower class constructor %q: %w", ctorMangled, sourceError(fileName, statement.Class.Constructor.Span, err))
		}
		module.Functions = append(module.Functions, function)
		signatures[ctorMangled] = function
	} else if len(fieldInits) > 0 || statement.Class.Extends != "" {
		ctorMangled := className + "_constructor"
		var ctorBody []frontend.SyntaxStatement
		var ctorParams []frontend.SyntaxParameter
		if statement.Class.Extends != "" {
			var superArgs []*frontend.SyntaxExpression
			baseClass := qualifyClassType(fileName, statement.Class.Extends)
			if baseCtor, _, found := findConstructorInHierarchy(baseClass, signatures, hierarchy); found && len(baseCtor.Parameters) > 1 {
				for _, p := range baseCtor.Parameters[1:] {
					ctorParams = append(ctorParams, frontend.SyntaxParameter{
						Name: p.Name,
						Type: string(p.Type),
					})
					superArgs = append(superArgs, &frontend.SyntaxExpression{
						Span: statement.Span,
						Kind: "identifier",
						Text: p.Name,
					})
				}
			}
			ctorBody = append(ctorBody, frontend.SyntaxStatement{
				Span: statement.Span,
				Kind: "expression",
				Expression: &frontend.SyntaxExpression{
					Span:      statement.Span,
					Kind:      "call",
					Left:      &frontend.SyntaxExpression{Span: statement.Span, Kind: "identifier", Text: "super"},
					Arguments: superArgs,
				},
			})
		}
		ctorBody = append(ctorBody, fieldInits...)
		ctorStmt := frontend.SyntaxStatement{
			Span: statement.Span,
			Kind: "function",
			Name: ctorMangled,
			Type: "void",
			Parameters: append([]frontend.SyntaxParameter{
				{Name: "this", Type: "object:" + className},
			}, ctorParams...),
			Body: ctorBody,
		}
		function, err := lowerFunction(fileName, ctorStmt, shapes, signatures)
		if err != nil {
			return fmt.Errorf("lower class default constructor %q: %w", ctorMangled, sourceError(fileName, statement.Span, err))
		}
		module.Functions = append(module.Functions, function)
		signatures[ctorMangled] = function
	}

	// Lower methods, static methods, getters, setters
	allMethods := getInheritedMethods(className, hierarchy)
	for _, method := range allMethods {
		if method.IsAbstract || method.Body == nil {
			continue
		}
		var mangled string
		var params []frontend.SyntaxParameter
		retType := method.Type
		// TypeScript's polymorphic `this` return type is the concrete
		// class type at each method boundary. Keeping it as the literal
		// `this` type makes the next chained call lose its receiver class.
		if retType == "this" || retType == "object:this" {
			retType = "object:" + className
		}
		var cleanParams []frontend.SyntaxParameter
		for _, p := range method.Parameters {
			if p.Name != "this" {
				cleanParams = append(cleanParams, p)
			}
		}
		if method.IsStatic {
			mangled = className + "_static_" + method.Name
			params = cleanParams
		} else if method.Kind == "get" {
			mangled = className + "_get_" + method.Name
			params = []frontend.SyntaxParameter{{Name: "this", Type: "object:" + className}}
		} else if method.Kind == "set" {
			mangled = className + "_set_" + method.Name
			params = append([]frontend.SyntaxParameter{{Name: "this", Type: "object:" + className}}, cleanParams...)
			retType = "void"
		} else {
			mangled = methodImplementationName(className, method.Name)
			params = append([]frontend.SyntaxParameter{{Name: "this", Type: "object:" + className}}, cleanParams...)
		}
		methodStmt := frontend.SyntaxStatement{
			Span:       method.Span,
			Kind:       "function",
			IsAsync:    method.IsAsync,
			Name:       mangled,
			Type:       retType,
			Parameters: params,
			Body:       method.Body,
		}
		function, err := lowerFunction(fileName, methodStmt, shapes, signatures)
		if err != nil {
			return fmt.Errorf("lower class method %q: %w", mangled, sourceError(fileName, method.Span, err))
		}
		module.Functions = append(module.Functions, function)
		signatures[mangled] = function
		if method.IsStatic {
			signatures[className+"."+method.Name] = function
		}
	}

	// Lower static field initializers and static blocks in class definition order
	if len(statement.Class.StaticElements) > 0 {
		for _, elem := range statement.Class.StaticElements {
			switch elem.Kind {
			case frontend.StaticElementField:
				f := elem.Field
				if f != nil && f.IsStatic && f.Initializer != nil {
					staticVar := className + "_" + f.Name
					_, valType, err := lowerExpression(fileName, f.Initializer, staticVar, main, env, counter, shapes, signatures)
					if err == nil {
						env[staticVar] = valType
					}
				}
			case frontend.StaticElementBlock:
				for _, stmt := range elem.Statements {
					if err := lowerStatement(fileName, stmt, main, env, counter, shapes, signatures); err != nil {
						return fmt.Errorf("lower class %s static block: %w", statement.Class.Name, sourceError(fileName, stmt.Span, err))
					}
				}
			}
		}
	} else {
		for _, f := range statement.Class.Fields {
			if f.IsStatic && f.Initializer != nil {
				staticVar := className + "_" + f.Name
				_, valType, err := lowerExpression(fileName, f.Initializer, staticVar, main, env, counter, shapes, signatures)
				if err == nil {
					env[staticVar] = valType
				}
			}
		}
		for _, block := range statement.Class.StaticBlocks {
			for _, stmt := range block {
				if err := lowerStatement(fileName, stmt, main, env, counter, shapes, signatures); err != nil {
					return fmt.Errorf("lower class %s static block: %w", statement.Class.Name, sourceError(fileName, stmt.Span, err))
				}
			}
		}
	}
	if err := lowerClassDecorators(fileName, statement.Class, main, env, counter, shapes, signatures); err != nil {
		return fmt.Errorf("lower class %s decorators: %w", statement.Class.Name, err)
	}
	return nil
}
