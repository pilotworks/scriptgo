package typescriptgo

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

// flattenDestructuring recursively flattens ObjectBindingPattern and ArrayBindingPattern
// into a sequential list of SSA-ready variable declarations.
func flattenDestructuring(nameNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	if nameNode == nil {
		return nil
	}

	switch nameNode.Kind {
	case ast.KindObjectBindingPattern:
		return flattenObjectBinding(nameNode, initExpr, chk, counter)
	case ast.KindArrayBindingPattern:
		return flattenArrayBinding(nameNode, initExpr, chk, counter)
	default:
		inferredVarType := resolveInferredType(chk, nameNode)
		varType := inferredVarType
		return []SyntaxStatement{{
			Span:         sourceSpan(nameNode),
			Kind:         "variable",
			Name:         nameNode.Text(),
			Type:         varType,
			InferredType: inferredVarType,
			Expression:   initExpr,
		}}
	}
}

func flattenObjectBinding(nameNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	var stmts []SyntaxStatement
	pattern := nameNode.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return nil
	}

	*counter++
	objVar := fmt.Sprintf("__destruct_obj_%d_%d", nameNode.Pos(), *counter)
	if initExpr != nil && initExpr.Kind == "identifier" {
		objVar = initExpr.Text
	} else {
		inferred := ""
		if initExpr != nil && !isNullishTypeText(initExpr.InferredType) {
			inferred = initExpr.InferredType
		} else {
			inferred = resolveInferredType(chk, nameNode)
		}
		if isNullishTypeText(inferred) {
			var fields []string
			for _, elem := range pattern.Elements.Nodes {
				if elem != nil && elem.Kind != ast.KindOmittedExpression {
					if b := elem.AsBindingElement(); b != nil {
						pName := ""
						if b.PropertyName != nil {
							pName = syntaxMemberName(b.PropertyName)
						} else if b.Name() != nil {
							pName = syntaxMemberName(b.Name())
						}
						t := resolveInferredType(chk, b.Name())
						if t == "" || t == "void" {
							if b.Initializer != nil {
								t = resolveInferredType(chk, b.Initializer)
							}
						}
						if pName != "" && t != "" && t != "void" {
							fields = append(fields, fmt.Sprintf("%s?: %s", pName, t))
						}
					}
				}
			}
			if len(fields) > 0 {
				inferred = "{" + strings.Join(fields, "; ") + "}"
			} else {
				inferred = "object"
			}
		}
		stmts = append(stmts, SyntaxStatement{
			Span:         sourceSpan(nameNode),
			Kind:         "variable",
			Name:         objVar,
			Type:         inferred,
			InferredType: inferred,
			Expression:   initExpr,
		})
	}

	stmts = append(stmts, requireObjectCoercible(sourceSpan(nameNode), objVar))

	for _, elem := range pattern.Elements.Nodes {
		if elem == nil || elem.Kind == ast.KindOmittedExpression {
			continue
		}
		binding := elem.AsBindingElement()
		if binding == nil {
			continue
		}

		propName := ""
		if binding.PropertyName != nil {
			propName = syntaxMemberName(binding.PropertyName)
		} else if binding.Name() != nil {
			propName = syntaxMemberName(binding.Name())
		}

		targetNode := binding.Name()
		var propExpr *SyntaxExpression
		if binding.DotDotDotToken != nil {
			inferred := resolveInferredType(chk, targetNode)
			if inferred == "" || inferred == "void" || inferred == "undefined" {
				inferred = resolveInferredType(chk, elem)
			}
			var fieldArgs []*SyntaxExpression
			if chk != nil && targetNode != nil {
				if t := chk.GetTypeAtLocation(targetNode); t != nil {
					for _, sym := range chk.GetPropertiesOfType(t) {
						if sym != nil {
							symName := sym.Name
							symType := normalizeInferredType(chk.TypeToString(chk.GetTypeOfSymbolAtLocation(sym, targetNode)))
							fieldArgs = append(fieldArgs, &SyntaxExpression{
								Span:         sourceSpan(elem),
								Kind:         "property",
								Text:         symName,
								InferredType: symType,
								Left: &SyntaxExpression{
									Span:         sourceSpan(elem),
									Kind:         "property",
									Text:         symName,
									InferredType: symType,
									Left: &SyntaxExpression{
										Span: sourceSpan(elem),
										Kind: "identifier",
										Text: objVar,
									},
								},
							})
						}
					}
				}
			}
			if len(fieldArgs) > 0 {
				propExpr = &SyntaxExpression{
					Span:         sourceSpan(elem),
					Kind:         "object_literal",
					Arguments:    fieldArgs,
					InferredType: inferred,
				}
			} else {
				propExpr = &SyntaxExpression{
					Span:         sourceSpan(elem),
					Kind:         "identifier",
					Text:         objVar,
					InferredType: inferred,
				}
			}
		} else {
			propExpr = &SyntaxExpression{
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
		}

		if binding.Initializer != nil && targetNode != nil && targetNode.Kind != ast.KindObjectBindingPattern && targetNode.Kind != ast.KindArrayBindingPattern {
			varName := targetNode.Text()
			defaultExpr := syntaxExpression(binding.Initializer, chk)
			inferredType := resolveInferredType(chk, targetNode)
			if inferredType == "" || inferredType == "void" || inferredType == "undefined" {
				if defaultExpr != nil && defaultExpr.InferredType != "" && defaultExpr.InferredType != "void" && defaultExpr.InferredType != "undefined" {
					inferredType = defaultExpr.InferredType
				} else if propExpr.InferredType != "" {
					inferredType = propExpr.InferredType
				} else {
					inferredType = resolveInferredType(chk, elem)
				}
			}
			stmts = append(stmts, SyntaxStatement{
				Span:         sourceSpan(elem),
				Kind:         "variable",
				Name:         varName,
				Type:         inferredType,
				InferredType: inferredType,
				Expression:   defaultExpr,
			})
			stmts = append(stmts, SyntaxStatement{
				Span: sourceSpan(elem),
				Kind: "if",
				Expression: &SyntaxExpression{
					Span:     sourceSpan(elem),
					Kind:     "binary",
					Operator: "!==",
					Left:     propExpr,
					Right:    &SyntaxExpression{Span: sourceSpan(elem), Kind: "undefined", Text: "undefined"},
				},
				Then: []SyntaxStatement{
					{
						Span:       sourceSpan(elem),
						Kind:       "assign",
						Name:       varName,
						Expression: propExpr,
					},
				},
			})
			continue
		}

		if binding.Initializer != nil {
			defaultExpr := syntaxExpression(binding.Initializer, chk)
			*counter++
			tmpVar := fmt.Sprintf("__destruct_opt_%d_%d", elem.Pos(), *counter)
			inferredType := resolveInferredType(chk, targetNode)
			if inferredType == "" || inferredType == "void" || inferredType == "undefined" || inferredType == "{}" || inferredType == "object" {
				if propExpr != nil && propExpr.InferredType != "" && propExpr.InferredType != "void" && propExpr.InferredType != "undefined" {
					inferredType = propExpr.InferredType
				} else if defaultExpr != nil && defaultExpr.InferredType != "" && defaultExpr.InferredType != "void" && defaultExpr.InferredType != "{}" {
					inferredType = defaultExpr.InferredType
				}
			}
			if targetNode != nil && targetNode.Kind == ast.KindObjectBindingPattern && (inferredType == "" || inferredType == "void" || inferredType == "undefined" || inferredType == "{}" || inferredType == "object") {
				var fields []string
				if pat := targetNode.AsBindingPattern(); pat != nil && pat.Elements != nil {
					for _, sub := range pat.Elements.Nodes {
						if sub != nil && sub.Kind != ast.KindOmittedExpression {
							if sb := sub.AsBindingElement(); sb != nil {
								pName := ""
								if sb.PropertyName != nil {
									pName = syntaxMemberName(sb.PropertyName)
								} else if sb.Name() != nil {
									pName = syntaxMemberName(sb.Name())
								}
								st := resolveInferredType(chk, sb.Name())
								if st == "" || st == "void" || st == "undefined" {
									if sb.Initializer != nil {
										st = resolveInferredType(chk, sb.Initializer)
									}
								}
								if pName != "" && st != "" && st != "void" && st != "undefined" {
									fields = append(fields, fmt.Sprintf("%s?: %s", pName, st))
								}
							}
						}
					}
				}
				if len(fields) > 0 {
					inferredType = "{" + strings.Join(fields, "; ") + "}"
				} else {
					inferredType = "object"
				}
			}
			if defaultExpr != nil && (defaultExpr.InferredType == "" || defaultExpr.InferredType == "{}" || defaultExpr.InferredType == "void" || defaultExpr.InferredType == "object") {
				defaultExpr.InferredType = inferredType
			}
			stmts = append(stmts, SyntaxStatement{
				Span:         sourceSpan(elem),
				Kind:         "variable",
				Name:         tmpVar,
				Type:         inferredType,
				InferredType: inferredType,
				Expression:   defaultExpr,
			})
			stmts = append(stmts, SyntaxStatement{
				Span: sourceSpan(elem),
				Kind: "if",
				Expression: &SyntaxExpression{
					Span:     sourceSpan(elem),
					Kind:     "binary",
					Operator: "!==",
					Left:     propExpr,
					Right:    &SyntaxExpression{Span: sourceSpan(elem), Kind: "undefined", Text: "undefined"},
				},
				Then: []SyntaxStatement{
					{
						Span:       sourceSpan(elem),
						Kind:       "assign",
						Name:       tmpVar,
						Expression: propExpr,
					},
				},
			})
			propExpr = &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "identifier",
				Text:         tmpVar,
				InferredType: inferredType,
			}
		}

		if targetNode != nil && (targetNode.Kind == ast.KindObjectBindingPattern || targetNode.Kind == ast.KindArrayBindingPattern) {
			nestedStmts := flattenDestructuring(targetNode, propExpr, chk, counter)
			stmts = append(stmts, nestedStmts...)
		} else if targetNode != nil {
			varName := targetNode.Text()
			inferredType := resolveInferredType(chk, targetNode)
			if inferredType == "" {
				inferredType = resolveInferredType(chk, elem)
			}
			stmts = append(stmts, SyntaxStatement{
				Span:         sourceSpan(elem),
				Kind:         "variable",
				Name:         varName,
				InferredType: inferredType,
				Expression:   propExpr,
			})
		}
	}
	return stmts
}

// requireObjectCoercible is the RequireObjectCoercible step every
// destructuring pattern performs on its source before reading from it
// (destructuring null or undefined throws a TypeError). It is emitted as a
// call to a scriptgo intrinsic so lowering can drop it for storage that
// cannot hold null or undefined.
func requireObjectCoercible(span SourceSpan, source string) SyntaxStatement {
	return SyntaxStatement{
		Span: span,
		Kind: "expression",
		Expression: &SyntaxExpression{
			Span:         span,
			Kind:         "call",
			InferredType: "void",
			Left: &SyntaxExpression{
				Span: span,
				Kind: "property",
				Text: "requireObjectCoercible",
				Left: &SyntaxExpression{Span: span, Kind: "identifier", Text: "__scriptgo"},
			},
			Arguments: []*SyntaxExpression{{Span: span, Kind: "identifier", Text: source}},
		},
	}
}

// isNullishTypeText reports a type that names no object shape (absent, void,
// undefined, or null), so a destructuring source of that type takes its
// shape from the pattern; RequireObjectCoercible then throws at runtime.
func isNullishTypeText(typ string) bool {
	return typ == "" || typ == "void" || typ == "undefined" || typ == "null"
}
