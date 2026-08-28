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
		if initExpr != nil && initExpr.InferredType != "" && initExpr.InferredType != "void" && initExpr.InferredType != "undefined" {
			inferred = initExpr.InferredType
		} else {
			inferred = resolveInferredType(chk, nameNode)
		}
		if inferred == "" || inferred == "void" || inferred == "undefined" {
			var fields []string
			for _, elem := range pattern.Elements.Nodes {
				if elem != nil && elem.Kind != ast.KindOmittedExpression {
					if b := elem.AsBindingElement(); b != nil {
						pName := ""
						if b.PropertyName != nil {
							pName = b.PropertyName.Text()
						} else if b.Name() != nil {
							pName = b.Name().Text()
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
			propName = binding.PropertyName.Text()
		} else if binding.Name() != nil {
			propName = binding.Name().Text()
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
									pName = sb.PropertyName.Text()
								} else if sb.Name() != nil {
									pName = sb.Name().Text()
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

func flattenArrayBinding(nameNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	var stmts []SyntaxStatement
	pattern := nameNode.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return nil
	}

	*counter++
	arrVar := fmt.Sprintf("__destruct_arr_%d_%d", nameNode.Pos(), *counter)
	arrInferred := ""
	if initExpr != nil && initExpr.InferredType != "" && initExpr.InferredType != "void" && initExpr.InferredType != "undefined" {
		arrInferred = initExpr.InferredType
	} else {
		arrInferred = resolveInferredType(chk, nameNode)
	}
	var elemTypes []string
	for _, elem := range pattern.Elements.Nodes {
		if elem != nil && elem.Kind != ast.KindOmittedExpression {
			if b := elem.AsBindingElement(); b != nil {
				t := resolveInferredType(chk, b.Name())
				if t == "" || t == "void" || t == "undefined" {
					if b.Name() != nil && b.Name().Kind == ast.KindArrayBindingPattern {
						t = "unknown[]"
					} else if b.Name() != nil && b.Name().Kind == ast.KindObjectBindingPattern {
						t = "object"
					} else if b.Initializer != nil {
						t = resolveInferredType(chk, b.Initializer)
					}
				}
				if t != "" && t != "void" && t != "undefined" {
					elemTypes = append(elemTypes, t)
				}
			}
		}
	}
	isHeterogeneousPattern := false
	if len(elemTypes) > 1 {
		for i := 1; i < len(elemTypes); i++ {
			if elemTypes[i] != elemTypes[0] {
				isHeterogeneousPattern = true
				break
			}
		}
	}
	initTupleLen := countTupleElements(arrInferred)
	isTupleType := initTupleLen >= len(pattern.Elements.Nodes)
	if (isHeterogeneousPattern && !isTupleType) || (initTupleLen >= 0 && initTupleLen < len(pattern.Elements.Nodes)) {
		arrInferred = "unknown[]"
		if initExpr != nil && (countTupleElements(initExpr.InferredType) < len(pattern.Elements.Nodes)) {
			initExpr.InferredType = "unknown[]"
		}
	} else if arrInferred == "" || arrInferred == "void" || arrInferred == "undefined" {
		if len(elemTypes) > 0 {
			arrInferred = elemTypes[0] + "[]"
		} else {
			arrInferred = "unknown[]"
		}
	}
	if initExpr != nil && initExpr.Kind == "identifier" {
		arrVar = initExpr.Text
	} else {
		stmts = append(stmts, SyntaxStatement{
			Span:         sourceSpan(nameNode),
			Kind:         "variable",
			Name:         arrVar,
			Type:         arrInferred,
			InferredType: arrInferred,
			Expression:   initExpr,
		})
	}

	for idx, elem := range pattern.Elements.Nodes {
		if elem == nil || elem.Kind == ast.KindOmittedExpression {
			continue
		}
		binding := elem.AsBindingElement()
		if binding == nil {
			continue
		}

		targetNode := binding.Name()
		if binding.Initializer != nil && targetNode != nil && targetNode.Kind != ast.KindObjectBindingPattern && targetNode.Kind != ast.KindArrayBindingPattern {
			varName := targetNode.Text()
			defaultExpr := syntaxExpression(binding.Initializer, chk)
			inferredType := resolveInferredType(chk, targetNode)
			if inferredType == "" || inferredType == "void" || inferredType == "undefined" {
				if defaultExpr != nil && defaultExpr.InferredType != "" && defaultExpr.InferredType != "void" && defaultExpr.InferredType != "undefined" {
					inferredType = defaultExpr.InferredType
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
			rawIdxExpr := &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "index",
				InferredType: inferredType,
				Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
				Right:        &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
			}
			var condExpr *SyntaxExpression = &SyntaxExpression{
				Span:     sourceSpan(elem),
				Kind:     "binary",
				Operator: "!==",
				Left:     rawIdxExpr,
				Right:    &SyntaxExpression{Span: sourceSpan(elem), Kind: "undefined", Text: "undefined"},
			}
			condExpr = &SyntaxExpression{
				Span:     sourceSpan(elem),
				Kind:     "binary",
				Operator: "&&",
				Left: &SyntaxExpression{
					Span:     sourceSpan(elem),
					Kind:     "binary",
					Operator: "<",
					Left:     &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
					Right: &SyntaxExpression{
						Span:         sourceSpan(elem),
						Kind:         "property",
						Text:         "length",
						Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						InferredType: "number",
					},
				},
				Right: condExpr,
			}
			stmts = append(stmts, SyntaxStatement{
				Span:       sourceSpan(elem),
				Kind:       "if",
				Expression: condExpr,
				Then: []SyntaxStatement{
					{
						Span:       sourceSpan(elem),
						Kind:       "assign",
						Name:       varName,
						Expression: rawIdxExpr,
					},
				},
			})
			continue
		}

		var itemExpr *SyntaxExpression
		if binding.DotDotDotToken != nil {
			inferred := resolveInferredType(chk, targetNode)
			if inferred == "" || inferred == "void" || inferred == "undefined" {
				inferred = resolveInferredType(chk, elem)
			}
			if (inferred == "" || inferred == "void" || inferred == "undefined") && arrInferred != "" {
				inferred = arrInferred
			}
			itemExpr = &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "call",
				InferredType: inferred,
				Left: &SyntaxExpression{
					Span:         sourceSpan(elem),
					Kind:         "property",
					Text:         "slice",
					InferredType: inferred,
					Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar, InferredType: arrInferred},
				},
				Arguments: []*SyntaxExpression{
					{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
				},
			}
		} else {
			itemExpr = &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "index",
				InferredType: resolveInferredType(chk, elem),
				Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
				Right:        &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
			}
		}

		if binding.Initializer != nil && targetNode != nil && (targetNode.Kind == ast.KindObjectBindingPattern || targetNode.Kind == ast.KindArrayBindingPattern) {
			defaultExpr := syntaxExpression(binding.Initializer, chk)
			*counter++
			tmpVar := fmt.Sprintf("__destruct_opt_%d_%d", elem.Pos(), *counter)
			inferredType := resolveInferredType(chk, targetNode)
			if inferredType == "" || inferredType == "void" || inferredType == "undefined" || inferredType == "never[]" || inferredType == "void[]" || inferredType == "[]" {
				if defaultExpr != nil && defaultExpr.InferredType != "" && defaultExpr.InferredType != "void" && defaultExpr.InferredType != "never[]" && defaultExpr.InferredType != "void[]" && defaultExpr.InferredType != "[]" {
					inferredType = defaultExpr.InferredType
				} else if targetNode.Kind == ast.KindArrayBindingPattern {
					var subTypes []string
					if pat := targetNode.AsBindingPattern(); pat != nil && pat.Elements != nil {
						for _, sub := range pat.Elements.Nodes {
							if sub != nil && sub.Kind != ast.KindOmittedExpression {
								if sb := sub.AsBindingElement(); sb != nil {
									st := resolveInferredType(chk, sb.Name())
									if st == "" || st == "void" || st == "undefined" {
										if sb.Initializer != nil {
											st = resolveInferredType(chk, sb.Initializer)
										}
									}
									if st != "" && st != "void" && st != "undefined" {
										subTypes = append(subTypes, st)
									}
								}
							}
						}
					}
					if len(subTypes) > 0 {
						inferredType = subTypes[0] + "[]"
					} else {
						inferredType = "unknown[]"
					}
				} else {
					inferredType = "object"
				}
			}
			if defaultExpr != nil && (defaultExpr.InferredType == "" || defaultExpr.InferredType == "never[]" || defaultExpr.InferredType == "void[]" || defaultExpr.InferredType == "[]") {
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
			itemExpr.InferredType = inferredType
			condExpr := &SyntaxExpression{
				Span:     sourceSpan(elem),
				Kind:     "binary",
				Operator: "&&",
				Left: &SyntaxExpression{
					Span:     sourceSpan(elem),
					Kind:     "binary",
					Operator: "<",
					Left:     &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
					Right: &SyntaxExpression{
						Span:         sourceSpan(elem),
						Kind:         "property",
						Text:         "length",
						Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						InferredType: "number",
					},
				},
				Right: &SyntaxExpression{
					Span:     sourceSpan(elem),
					Kind:     "binary",
					Operator: "!==",
					Left:     itemExpr,
					Right:    &SyntaxExpression{Span: sourceSpan(elem), Kind: "undefined", Text: "undefined"},
				},
			}
			stmts = append(stmts, SyntaxStatement{
				Span:       sourceSpan(elem),
				Kind:       "if",
				Expression: condExpr,
				Then: []SyntaxStatement{
					{
						Span:       sourceSpan(elem),
						Kind:       "assign",
						Name:       tmpVar,
						Expression: itemExpr,
					},
				},
			})
			itemExpr = &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "identifier",
				Text:         tmpVar,
				InferredType: inferredType,
			}
		} else if binding.Initializer != nil {
			defaultExpr := syntaxExpression(binding.Initializer, chk)
			itemExpr = &SyntaxExpression{
				Span:         sourceSpan(elem),
				Kind:         "binary",
				Operator:     "??",
				Left:         itemExpr,
				Right:        defaultExpr,
				InferredType: itemExpr.InferredType,
			}
		}

		if targetNode != nil && (targetNode.Kind == ast.KindObjectBindingPattern || targetNode.Kind == ast.KindArrayBindingPattern) {
			nestedStmts := flattenDestructuring(targetNode, itemExpr, chk, counter)
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
				Expression:   itemExpr,
			})
		}
	}
	return stmts
}

func flattenDestructuringAssignment(leftNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	if leftNode == nil {
		return nil
	}
	var stmts []SyntaxStatement
	*counter++
	tmpVar := fmt.Sprintf("__destruct_assign_%d_%d", leftNode.Pos(), *counter)
	inferred := ""
	if initExpr != nil && initExpr.InferredType != "" && initExpr.InferredType != "void" && initExpr.InferredType != "undefined" {
		inferred = initExpr.InferredType
	} else {
		inferred = resolveInferredType(chk, leftNode)
	}
	stmts = append(stmts, SyntaxStatement{
		Span:         sourceSpan(leftNode),
		Kind:         "variable",
		Name:         tmpVar,
		Type:         inferred,
		InferredType: inferred,
		Expression:   initExpr,
	})

	if leftNode.Kind == ast.KindArrayLiteralExpression {
		arrLit := leftNode.AsArrayLiteralExpression()
		if arrLit != nil && arrLit.Elements != nil {
			for idx, elem := range arrLit.Elements.Nodes {
				if elem == nil || elem.Kind == ast.KindOmittedExpression {
					continue
				}
				if elem.Kind == ast.KindIdentifier {
					stmts = append(stmts, SyntaxStatement{
						Span: sourceSpan(elem),
						Kind: "assign",
						Name: elem.Text(),
						Expression: &SyntaxExpression{
							Span:  sourceSpan(elem),
							Kind:  "index",
							Left:  &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
						},
					})
				} else if elem.Kind == ast.KindArrayLiteralExpression || elem.Kind == ast.KindObjectLiteralExpression {
					itemExpr := &SyntaxExpression{
						Span:  sourceSpan(elem),
						Kind:  "index",
						Left:  &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
						Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
					}
					nested := flattenDestructuringAssignment(elem, itemExpr, chk, counter)
					stmts = append(stmts, nested...)
				}
			}
		}
	} else if leftNode.Kind == ast.KindObjectLiteralExpression {
		objLit := leftNode.AsObjectLiteralExpression()
		if objLit != nil && objLit.Properties != nil {
			for _, prop := range objLit.Properties.Nodes {
				if prop.Kind == ast.KindShorthandPropertyAssignment {
					propName := prop.Name().Text()
					stmts = append(stmts, SyntaxStatement{
						Span: sourceSpan(prop),
						Kind: "assign",
						Name: propName,
						Expression: &SyntaxExpression{
							Span: sourceSpan(prop),
							Kind: "property",
							Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Text: propName,
						},
					})
				} else if prop.Kind == ast.KindPropertyAssignment {
					pAssign := prop.AsPropertyAssignment()
					propName := pAssign.Name().Text()
					if pAssign.Initializer != nil {
						if pAssign.Initializer.Kind == ast.KindIdentifier {
							stmts = append(stmts, SyntaxStatement{
								Span: sourceSpan(prop),
								Kind: "assign",
								Name: pAssign.Initializer.Text(),
								Expression: &SyntaxExpression{
									Span: sourceSpan(prop),
									Kind: "property",
									Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
									Text: propName,
								},
							})
						} else if pAssign.Initializer.Kind == ast.KindBinaryExpression {
							bin := pAssign.Initializer.AsBinaryExpression()
							if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
								targetName := bin.Left.Text()
								defaultVal := syntaxExpression(bin.Right, chk)
								propExpr := &SyntaxExpression{
									Span: sourceSpan(prop),
									Kind: "property",
									Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
									Text: propName,
								}
								stmts = append(stmts, SyntaxStatement{
									Span: sourceSpan(prop),
									Kind: "assign",
									Name: targetName,
									Expression: &SyntaxExpression{
										Span:     sourceSpan(prop),
										Kind:     "binary",
										Operator: "??",
										Left:     propExpr,
										Right:    defaultVal,
									},
								})
							}
						}
					}
				} else if prop.Kind == ast.KindBinaryExpression {
					bin := prop.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
						propName := bin.Left.Text()
						defaultVal := syntaxExpression(bin.Right, chk)
						propExpr := &SyntaxExpression{
							Span: sourceSpan(prop),
							Kind: "property",
							Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Text: propName,
						}
						stmts = append(stmts, SyntaxStatement{
							Span: sourceSpan(prop),
							Kind: "assign",
							Name: propName,
							Expression: &SyntaxExpression{
								Span:     sourceSpan(prop),
								Kind:     "binary",
								Operator: "??",
								Left:     propExpr,
								Right:    defaultVal,
							},
						})
					}
				}
			}
		}
	}
	return stmts
}

func countTupleElements(tupleType string) int {
	if !strings.HasPrefix(tupleType, "[") || !strings.HasSuffix(tupleType, "]") {
		return -1
	}
	inner := strings.TrimSpace(tupleType[1 : len(tupleType)-1])
	if inner == "" {
		return 0
	}
	count := 1
	depth := 0
	for _, ch := range inner {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}
