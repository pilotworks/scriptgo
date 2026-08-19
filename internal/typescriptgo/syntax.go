package typescriptgo

import (
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
)

func syntaxFile(file *ast.SourceFile) SyntaxFile {
	result := SyntaxFile{FileName: file.FileName()}
	if file.Statements == nil {
		return result
	}
	for _, statement := range file.Statements.Nodes {
		if converted, ok := syntaxStatement(statement); ok {
			result.Statements = append(result.Statements, converted)
		}
	}
	return result
}

func sourceSpan(node *ast.Node) SourceSpan {
	if node == nil {
		return SourceSpan{}
	}
	return SourceSpan{Start: node.Pos(), Length: node.End() - node.Pos()}
}

func parameterSpan(node *ast.Node) SourceSpan {
	return sourceSpan(node)
}

func syntaxType(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindNumberKeyword:
		return "number"
	case ast.KindStringKeyword:
		return "string"
	case ast.KindBooleanKeyword:
		return "bool"
	case ast.KindVoidKeyword:
		return "void"
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		if typeRef != nil && typeRef.TypeName != nil {
			name := typeRef.TypeName.Text()
			if (name == "Array" || name == "ReadonlyArray") && typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) == 1 {
				return syntaxType(typeRef.TypeArguments.Nodes[0]) + "[]"
			}
			if typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0 {
				var args []string
				for _, arg := range typeRef.TypeArguments.Nodes {
					args = append(args, syntaxType(arg))
				}
				return name + "<" + strings.Join(args, ", ") + ">"
			}
			return name
		}
		return node.Kind.String()
	case ast.KindArrayType:
		array := node.AsArrayTypeNode()
		return syntaxType(array.ElementType) + "[]"
	default:
		return node.Kind.String()
	}
}

func syntaxTypeParameters(typeParams []*ast.Node) []string {
	if len(typeParams) == 0 {
		return nil
	}
	var result []string
	for _, p := range typeParams {
		if p.Name() != nil {
			result = append(result, p.Name().Text())
		}
	}
	return result
}

func syntaxTypeArguments(typeArgs *ast.NodeList) []string {
	if typeArgs == nil {
		return nil
	}
	var result []string
	for _, a := range typeArgs.Nodes {
		result = append(result, syntaxType(a))
	}
	return result
}

func statementCount(file *ast.SourceFile) int {
	if file.Statements == nil {
		return 0
	}
	return len(file.Statements.Nodes)
}

func isTypeScriptSource(fileName string) bool {
	if strings.HasSuffix(fileName, ".d.ts") {
		return false
	}
	for _, extension := range []string{".ts", ".tsx", ".mts", ".cts"} {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}
