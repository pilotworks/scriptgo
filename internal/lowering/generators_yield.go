package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
)

type yieldPoint struct {
	stateIdx int
	expr     *frontend.SyntaxExpression
	isStar   bool
}

func rewriteYieldsToPush(stmts []frontend.SyntaxStatement, itemsName string) []frontend.SyntaxStatement {
	var rewritten []frontend.SyntaxStatement
	for _, s := range stmts {
		cloned := s
		if s.Kind == "expression" && s.Expression != nil {
			if s.Expression.Kind == "yield" {
				yielded := s.Expression.Left
				if yielded == nil {
					// A bare `yield;` produces undefined.
					yielded = &frontend.SyntaxExpression{Span: s.Span, Kind: "undefined"}
				}
				cloned = frontend.SyntaxStatement{
					Span: s.Span,
					Kind: "expression",
					Expression: &frontend.SyntaxExpression{
						Span: s.Span,
						Kind: "call",
						Left: &frontend.SyntaxExpression{
							Span: s.Span,
							Kind: "property",
							Left: &frontend.SyntaxExpression{
								Span: s.Span,
								Kind: "identifier",
								Text: itemsName,
							},
							Text: "push",
						},
						Arguments: []*frontend.SyntaxExpression{yielded},
					},
				}
			} else if s.Expression.Kind == "yield_star" {
				if s.Expression.Left != nil && s.Expression.Left.Kind == "call" && s.Expression.Left.Left != nil && s.Expression.Left.Left.Kind == "identifier" {
					calleeName := s.Expression.Left.Left.Text
					if subGen, ok := generatorASTIndex[calleeName]; ok {
						subRewritten := rewriteYieldsToPush(subGen.Body, itemsName)
						rewritten = append(rewritten, subRewritten...)
						continue
					}
				}
				if s.Expression.Left != nil && s.Expression.Left.Kind == "array" {
					for _, elem := range s.Expression.Left.Arguments {
						rewritten = append(rewritten, frontend.SyntaxStatement{
							Span: s.Span,
							Kind: "expression",
							Expression: &frontend.SyntaxExpression{
								Span: s.Span,
								Kind: "call",
								Left: &frontend.SyntaxExpression{
									Span: s.Span,
									Kind: "property",
									Left: &frontend.SyntaxExpression{
										Span: s.Span,
										Kind: "identifier",
										Text: itemsName,
									},
									Text: "push",
								},
								Arguments: []*frontend.SyntaxExpression{
									elem,
								},
							},
						})
					}
					continue
				}
				cloned = frontend.SyntaxStatement{
					Span:       s.Span,
					Kind:       "forof",
					Name:       "___yield_star_item",
					Expression: s.Expression.Left,
					Body: []frontend.SyntaxStatement{
						{
							Span: s.Span,
							Kind: "expression",
							Expression: &frontend.SyntaxExpression{
								Span: s.Span,
								Kind: "call",
								Left: &frontend.SyntaxExpression{
									Span: s.Span,
									Kind: "property",
									Left: &frontend.SyntaxExpression{
										Span: s.Span,
										Kind: "identifier",
										Text: itemsName,
									},
									Text: "push",
								},
								Arguments: []*frontend.SyntaxExpression{
									{
										Span: s.Span,
										Kind: "identifier",
										Text: "___yield_star_item",
									},
								},
							},
						},
					},
				}
			}
		}
		if len(cloned.Body) > 0 {
			cloned.Body = rewriteYieldsToPush(cloned.Body, itemsName)
		}
		if len(cloned.Then) > 0 {
			cloned.Then = rewriteYieldsToPush(cloned.Then, itemsName)
		}
		if len(cloned.Else) > 0 {
			cloned.Else = rewriteYieldsToPush(cloned.Else, itemsName)
		}
		if len(cloned.Catch) > 0 {
			cloned.Catch = rewriteYieldsToPush(cloned.Catch, itemsName)
		}
		if len(cloned.Finally) > 0 {
			cloned.Finally = rewriteYieldsToPush(cloned.Finally, itemsName)
		}
		rewritten = append(rewritten, cloned)
	}
	return rewritten
}

func collectYieldPoints(body []frontend.SyntaxStatement) []yieldPoint {
	var yields []yieldPoint
	for _, s := range body {
		if s.Kind == "expression" && s.Expression != nil {
			switch s.Expression.Kind {
			case "yield":
				yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: false})
			case "yield_star":
				if s.Expression.Left != nil && s.Expression.Left.Kind == "array" {
					for _, arg := range s.Expression.Left.Arguments {
						yields = append(yields, yieldPoint{stateIdx: len(yields), expr: arg, isStar: false})
					}
				} else if s.Expression.Left != nil && s.Expression.Left.Kind == "call" && s.Expression.Left.Left != nil && s.Expression.Left.Left.Kind == "identifier" {
					calleeName := s.Expression.Left.Left.Text
					if subGen, ok := generatorASTIndex[calleeName]; ok {
						subYields := collectYieldPoints(subGen.Body)
						for _, sy := range subYields {
							sy.stateIdx = len(yields)
							yields = append(yields, sy)
						}
					} else {
						yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: true})
					}
				} else {
					yields = append(yields, yieldPoint{stateIdx: len(yields), expr: s.Expression.Left, isStar: true})
				}
			}
		}
	}
	return yields
}
