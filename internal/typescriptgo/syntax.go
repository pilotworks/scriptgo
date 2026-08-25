package typescriptgo

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxFile(file *ast.SourceFile, chk *checker.Checker) SyntaxFile {
	result := SyntaxFile{FileName: file.FileName()}
	if file.Statements == nil {
		return result
	}
	for _, statement := range file.Statements.Nodes {
		if converted, ok := syntaxStatement(statement, chk); ok {
			result.Statements = append(result.Statements, converted)
		}
	}
	return result
}

func resolveInferredType(chk *checker.Checker, node *ast.Node) string {
	if chk == nil || node == nil {
		return ""
	}
	typ := chk.GetTypeAtLocation(node)
	if typ == nil {
		return ""
	}
	typeStr := chk.TypeToString(typ)
	return normalizeInferredType(typeStr)
}

func resolveFunctionReturnType(chk *checker.Checker, node *ast.Node) string {
	if chk == nil || node == nil {
		return ""
	}
	sig := chk.GetSignatureFromDeclaration(node)
	if sig == nil {
		return ""
	}
	retType := chk.GetReturnTypeOfSignature(sig)
	if retType == nil {
		return ""
	}
	return normalizeInferredType(chk.TypeToString(retType))
}

func normalizeInferredType(typeStr string) string {
	typeStr = strings.TrimSpace(typeStr)
	typeStr = strings.ReplaceAll(typeStr, "\n", " ")
	for strings.Contains(typeStr, "  ") {
		typeStr = strings.ReplaceAll(typeStr, "  ", " ")
	}
	typeStr = strings.TrimSpace(typeStr)
	switch typeStr {
	case "boolean", "true", "false":
		return "bool"
	case "number", "string", "void", "any", "unknown", "never", "undefined", "null", "bigint", "symbol",
		"Uint8Array", "Int8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
		"Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array", "DataView", "ArrayBuffer", "Map", "Set":
		return typeStr
	default:
		if strings.HasPrefix(typeStr, "\"") && strings.HasSuffix(typeStr, "\"") {
			return "string"
		}
		if _, err := strconv.ParseFloat(typeStr, 64); err == nil {
			return "number"
		}
		if strings.HasPrefix(typeStr, "Array<") && strings.HasSuffix(typeStr, ">") {
			elem := strings.TrimSuffix(strings.TrimPrefix(typeStr, "Array<"), ">")
			return normalizeInferredType(elem) + "[]"
		}
		if strings.HasPrefix(typeStr, "ReadonlyArray<") && strings.HasSuffix(typeStr, ">") {
			elem := strings.TrimSuffix(strings.TrimPrefix(typeStr, "ReadonlyArray<"), ">")
			return normalizeInferredType(elem) + "[]"
		}
		if strings.Contains(typeStr, "|") {
			parts := strings.Split(typeStr, "|")
			var nonNullish []string
			for _, p := range parts {
				pNorm := normalizeInferredType(p)
				if pNorm != "undefined" && pNorm != "null" && pNorm != "void" {
					nonNullish = append(nonNullish, pNorm)
				}
			}
			if len(nonNullish) == 1 {
				return nonNullish[0]
			}
			var normParts []string
			for _, p := range parts {
				normParts = append(normParts, normalizeInferredType(p))
			}
			return strings.Join(normParts, " | ")
		}
		if strings.HasSuffix(typeStr, "[]") {
			elem := strings.TrimSuffix(typeStr, "[]")
			return normalizeInferredType(elem) + "[]"
		}
		return typeStr
	}
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
	case ast.KindBigIntKeyword:
		return "bigint"
	case ast.KindSymbolKeyword:
		return "symbol"
	case ast.KindStringKeyword:
		return "string"
	case ast.KindBooleanKeyword:
		return "bool"
	case ast.KindVoidKeyword:
		return "void"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindUndefinedKeyword:
		return "undefined"
	case ast.KindAnyKeyword:
		return "any"
	case ast.KindUnknownKeyword:
		return "unknown"
	case ast.KindNeverKeyword:
		return "never"
	case ast.KindThisType, ast.KindThisKeyword:
		return "this"
	case ast.KindUnionType:
		unionNode := node.AsUnionTypeNode()
		if unionNode != nil && unionNode.Types != nil {
			var types []string
			for _, elem := range unionNode.Types.Nodes {
				types = append(types, syntaxType(elem))
			}
			return strings.Join(types, " | ")
		}
		return "union"
	case ast.KindLiteralType:
		lit := node.AsLiteralTypeNode()
		if lit != nil && lit.Literal != nil {
			switch lit.Literal.Kind {
			case ast.KindNullKeyword:
				return "null"
			case ast.KindTrueKeyword:
				return "true"
			case ast.KindFalseKeyword:
				return "false"
			case ast.KindStringLiteral:
				return strconv.Quote(lit.Literal.Text())
			case ast.KindNumericLiteral:
				return lit.Literal.Text()
			default:
				return lit.Literal.Kind.String()
			}
		}
		return "literal"
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		if typeRef != nil && typeRef.TypeName != nil {
			name := entityNameText(typeRef.TypeName)
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
	case ast.KindTypeLiteral:
		typeLit := node.AsTypeLiteralNode()
		if typeLit != nil && typeLit.Members != nil {
			var members []string
			for _, m := range typeLit.Members.Nodes {
				if m.Kind == ast.KindPropertySignature && m.Name() != nil {
					name := m.Name().Text()
					t := syntaxType(m.Type())
					members = append(members, name+": "+t)
				}
			}
			return "{" + strings.Join(members, "; ") + "}"
		}
		return "object"
	case ast.KindIntersectionType:
		interNode := node.AsIntersectionTypeNode()
		if interNode != nil && interNode.Types != nil {
			var types []string
			for _, elem := range interNode.Types.Nodes {
				types = append(types, syntaxType(elem))
			}
			return strings.Join(types, " & ")
		}
		return "object"
	case ast.KindIndexedAccessType:
		idxNode := node.AsIndexedAccessTypeNode()
		if idxNode != nil && idxNode.ObjectType != nil && idxNode.IndexType != nil {
			objT := syntaxType(idxNode.ObjectType)
			idxT := syntaxType(idxNode.IndexType)
			idxT = strings.Trim(idxT, "\"")
			return objT + "[" + idxT + "]"
		}
		return "any"
	case ast.KindArrayType:
		array := node.AsArrayTypeNode()
		return syntaxType(array.ElementType) + "[]"
	case ast.KindTupleType:
		tuple := node.AsTupleTypeNode()
		if tuple != nil && tuple.Elements != nil {
			var elements []string
			for _, elem := range tuple.Elements.Nodes {
				elements = append(elements, syntaxType(elem))
			}
			return "[" + strings.Join(elements, ", ") + "]"
		}
		return "object"
	case ast.KindNamedTupleMember:
		named := node.AsNamedTupleMember()
		if named != nil && named.Type != nil {
			return syntaxType(named.Type)
		}
		return "any"
	case ast.KindRestType:
		restNode := node.AsRestTypeNode()
		if restNode != nil && restNode.Type != nil {
			return "..." + syntaxType(restNode.Type)
		}
		return "..."
	case ast.KindOptionalType:
		optNode := node.AsOptionalTypeNode()
		if optNode != nil && optNode.Type != nil {
			return syntaxType(optNode.Type) + "?"
		}
		return "any?"
	case ast.KindTypeOperator:
		op := node.AsTypeOperatorNode()
		if op != nil && op.Type != nil {
			return syntaxType(op.Type)
		}
		return "any"
	case ast.KindFunctionType:
		fnNode := node.AsFunctionTypeNode()
		if fnNode != nil {
			var params []string
			if fnNode.Parameters != nil {
				for _, p := range fnNode.Parameters.Nodes {
					pName := ""
					if p.Name() != nil {
						pName = p.Name().Text()
					}
					pType := syntaxType(p.Type())
					if pName != "" {
						params = append(params, pName+": "+pType)
					} else {
						params = append(params, pType)
					}
				}
			}
			retType := syntaxType(fnNode.Type)
			return "(" + strings.Join(params, ", ") + ") => " + retType
		}
		return "closure"
	case ast.KindParenthesizedType:
		paren := node.AsParenthesizedTypeNode()
		if paren != nil && paren.Type != nil {
			return "(" + syntaxType(paren.Type) + ")"
		}
		return "any"
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
		if p == nil {
			continue
		}
		if p.Kind == ast.KindTypeParameter {
			tp := p.AsTypeParameterDeclaration()
			if tp != nil && tp.Name() != nil {
				result = append(result, tp.Name().Text())
				continue
			}
		}
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

func entityNameText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindIdentifier {
		return node.Text()
	}
	if node.Kind == ast.KindQualifiedName {
		qn := node.AsQualifiedName()
		if qn != nil {
			left := entityNameText(qn.Left)
			right := entityNameText(qn.Right)
			if left != "" && right != "" {
				return left + "." + right
			}
			if right != "" {
				return right
			}
			return left
		}
	}
	return ""
}
