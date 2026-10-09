package typescriptgo

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

// objectLiteralKey returns the property key of an object literal member. A
// computed name ([expr]) uses the checker's literal type for the key
// expression, so ["a"], [0x10], and [k] with a const k all produce the
// runtime key (ToPropertyKey of the literal). Symbol keys ([S],
// [Symbol.iterator]) keep their member spelling. Any other computed key is
// only known at run time and reports false.
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
