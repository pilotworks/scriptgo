package typescriptgo

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func flattenArrayBinding(nameNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	var stmts []SyntaxStatement
	pattern := nameNode.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return nil
	}

	*counter++
	arrVar := fmt.Sprintf("__destruct_arr_%d_%d", nameNode.Pos(), *counter)
	arrInferred := ""
	if initExpr != nil && !isNullishTypeText(initExpr.InferredType) {
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
	} else if isNullishTypeText(arrInferred) {
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
	stmts = append(stmts, requireObjectCoercible(sourceSpan(nameNode), arrVar))

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
