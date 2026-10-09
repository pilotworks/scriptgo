package typescriptgo

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

// syntaxObjectLiteral converts an object literal: members with a key known
// at compile time become property_assignment, members keyed by a symbol
// value become computed_property (the key expression in Right), and spreads
// stay spread.
func syntaxObjectLiteral(node *ast.Node, chk *checker.Checker) *SyntaxExpression {
	objLit := node.AsObjectLiteralExpression()
	result := &SyntaxExpression{Span: sourceSpan(node), Kind: "object_literal", InferredType: resolveInferredType(chk, node)}
	if properties := objLit.Properties; properties != nil {
		for _, propNode := range properties.Nodes {
			switch propNode.Kind {
			case ast.KindPropertyAssignment:
				prop := propNode.AsPropertyAssignment()
				if keyExpr := symbolKeyExpression(prop.Name(), chk); keyExpr != nil {
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "computed_property",
						Right:        syntaxExpression(keyExpr, chk),
						Left:         syntaxExpression(prop.Initializer, chk),
						InferredType: resolveInferredType(chk, propNode),
					})
					continue
				}
				key, ok := objectLiteralKey(prop.Name(), chk)
				if !ok {
					result.Arguments = append(result.Arguments, unsupportedComputedKey(propNode))
					continue
				}
				result.Arguments = append(result.Arguments, &SyntaxExpression{
					Span:         sourceSpan(propNode),
					Kind:         "property_assignment",
					Text:         key,
					Left:         syntaxExpression(prop.Initializer, chk),
					InferredType: resolveInferredType(chk, propNode),
				})
			case ast.KindShorthandPropertyAssignment:
				prop := propNode.AsShorthandPropertyAssignment()
				name := syntaxMemberName(prop.Name())
				result.Arguments = append(result.Arguments, &SyntaxExpression{
					Span:         sourceSpan(propNode),
					Kind:         "property_assignment",
					Text:         name,
					Left:         &SyntaxExpression{Span: sourceSpan(propNode), Kind: "identifier", Text: name, InferredType: resolveInferredType(chk, propNode)},
					InferredType: resolveInferredType(chk, propNode),
				})
			case ast.KindSpreadAssignment:
				spread := propNode.AsSpreadAssignment()
				result.Arguments = append(result.Arguments, &SyntaxExpression{
					Span:         sourceSpan(propNode),
					Kind:         "spread",
					Left:         syntaxExpression(spread.Expression, chk),
					InferredType: resolveInferredType(chk, propNode),
				})
			case ast.KindMethodDeclaration:
				if keyExpr := symbolKeyExpression(propNode.Name(), chk); keyExpr != nil {
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "computed_property",
						Right:        syntaxExpression(keyExpr, chk),
						Left:         syntaxExpression(propNode, chk),
						InferredType: resolveInferredType(chk, propNode),
					})
					continue
				}
				name, ok := objectLiteralKey(propNode.Name(), chk)
				if !ok {
					result.Arguments = append(result.Arguments, unsupportedComputedKey(propNode))
					continue
				}
				result.Arguments = append(result.Arguments, &SyntaxExpression{
					Span:         sourceSpan(propNode),
					Kind:         "property_assignment",
					Text:         name,
					Left:         syntaxExpression(propNode, chk),
					InferredType: resolveInferredType(chk, propNode),
				})
			}
		}
	}
	return result
}
