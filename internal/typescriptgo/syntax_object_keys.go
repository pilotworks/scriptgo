package typescriptgo

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

// objectLiteralKey returns the property key of an object literal member. A
// computed name ([expr]) uses the checker's literal type for the key
// expression, so ["a"], [0x10], and [k] with a const k all produce the
// runtime key (ToPropertyKey of the literal). Well-known symbol keys
// ([Symbol.iterator]) keep their member spelling; other symbol keys are
// handled by symbolKeyExpression first. Any other computed key is only known
// at run time and reports false.
func objectLiteralKey(nameNode *ast.Node, chk *checker.Checker) (string, bool) {
	if nameNode == nil || nameNode.Kind != ast.KindComputedPropertyName {
		return syntaxMemberName(nameNode), true
	}
	expr := nameNode.Expression()
	if expr == nil {
		return "", false
	}
	if chk != nil {
		if t := chk.GetTypeAtLocation(expr); t != nil {
			if t.IsStringLiteral() || t.IsNumberLiteral() {
				return fmt.Sprint(t.AsLiteralType().Value()), true
			}
			if t.Flags()&checker.TypeFlagsESSymbolLike != 0 {
				// Symbol keys are modelled by their binding's spelling, the
				// same name the type's [S] member gets.
				return syntaxMemberName(nameNode), true
			}
		}
	}
	if expr.Kind == ast.KindPropertyAccessExpression {
		return syntaxMemberName(nameNode), true
	}
	return "", false
}

// unsupportedComputedKey marks an object literal member whose key is computed
// at run time; object shapes are fixed at compile time.
func unsupportedComputedKey(member *ast.Node) *SyntaxExpression {
	return &SyntaxExpression{Span: sourceSpan(member), Kind: "unsupported", Text: "computed property name that is not a compile-time constant"}
}

// symbolKeyExpression returns the key expression of a computed member name
// ([k]) whose key is a symbol other than a well-known Symbol.* one. Such a
// key is identified by the symbol value at run time, so it is not a field
// name; well-known symbols keep their Symbol.* spelling (objectLiteralKey).
func symbolKeyExpression(nameNode *ast.Node, chk *checker.Checker) *ast.Node {
	if nameNode == nil || nameNode.Kind != ast.KindComputedPropertyName || chk == nil {
		return nil
	}
	expr := nameNode.Expression()
	if expr == nil {
		return nil
	}
	if expr.Kind == ast.KindPropertyAccessExpression {
		if target := expr.Expression(); target != nil && target.Kind == ast.KindIdentifier && target.Text() == "Symbol" {
			return nil
		}
	}
	if t := chk.GetTypeAtLocation(expr); t != nil && t.Flags()&checker.TypeFlagsESSymbolLike != 0 {
		return expr
	}
	return nil
}

// isComputedReferenceName reports whether a member name is a computed name
// referring to a binding ([k], [Ns.k]) rather than a literal (["a"], [0]);
// in a type such a key is a symbol.
func isComputedReferenceName(nameNode *ast.Node) bool {
	if nameNode == nil || nameNode.Kind != ast.KindComputedPropertyName {
		return false
	}
	expr := nameNode.Expression()
	return expr != nil && (expr.Kind == ast.KindIdentifier || expr.Kind == ast.KindPropertyAccessExpression)
}
