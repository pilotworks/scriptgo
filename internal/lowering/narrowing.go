package lowering

import (
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func findMatchingDiscriminatedType(propName string, targetVal string, unionType string, shapes map[string]ir.ObjectShape) string {
	cleanTarget := strings.ToLower(strings.Trim(targetVal, "\"'`"))
	if cleanTarget == "" {
		return ""
	}
	cleanUnion := strings.TrimPrefix(unionType, "object:")
	if aliased, ok := typeAliasesIndex[cleanUnion]; ok {
		cleanUnion = aliased
	}
	isUnion := strings.Contains(cleanUnion, "|")
	if !isUnion && cleanUnion != "" && cleanUnion != "object" && !strings.HasPrefix(cleanUnion, "__shape_") {
		return ""
	}
	unionParts := map[string]bool{}
	if isUnion {
		for _, typePart := range strings.Split(cleanUnion, "|") {
			part := strings.TrimSpace(typePart)
			part = strings.TrimPrefix(part, "object:")
			unionParts[part] = true
			if strings.EqualFold(part, cleanTarget) {
				return part
			}
		}
	}
	for _, shapeName := range slices.Sorted(maps.Keys(shapes)) {
		s := shapes[shapeName]
		if isUnion && !unionParts[shapeName] {
			continue
		}
		for _, f := range s.Fields {
			if propName != "" && !strings.EqualFold(f.Name, propName) {
				continue
			}
			if propName != "" || strings.EqualFold(f.Name, "kind") || strings.EqualFold(f.Name, "type") || strings.EqualFold(f.Name, "tag") || strings.EqualFold(f.Name, "status") {
				cleanVal := strings.ToLower(strings.Trim(f.Value, "\"'`"))
				if cleanVal == cleanTarget {
					return shapeName
				}
			}
		}
	}
	return ""
}

func applyConditionNarrowing(expr *frontend.SyntaxExpression, thenEnv, elseEnv map[string]ir.Type, baseEnv map[string]ir.Type, shapes map[string]ir.ObjectShape) {
	if expr == nil {
		return
	}
	if expr.Kind == "unary" && (expr.Operator == "!" || expr.Operator == "not") {
		applyConditionNarrowing(expr.Left, elseEnv, thenEnv, baseEnv, shapes)
		return
	}
	if expr.Kind == "binary" && (expr.Operator == "&&" || expr.Operator == "and") {
		applyConditionNarrowing(expr.Left, thenEnv, elseEnv, baseEnv, shapes)
		applyConditionNarrowing(expr.Right, thenEnv, elseEnv, baseEnv, shapes)
		return
	}
	if expr.Kind == "binary" && (expr.Operator == "||" || expr.Operator == "or") {
		// The false branch of A || B requires both operands to be false. The
		// true branch cannot safely inherit a narrowing from either operand.
		unusedThen := make(map[string]ir.Type)
		applyConditionNarrowing(expr.Left, unusedThen, elseEnv, baseEnv, shapes)
		applyConditionNarrowing(expr.Right, unusedThen, elseEnv, baseEnv, shapes)
		return
	}
	if expr.Kind == "binary" && (expr.Operator == "===" || expr.Operator == "==" || expr.Operator == "!==" || expr.Operator == "!=") {
		if applyTypeofNarrowing(expr, thenEnv, elseEnv, baseEnv) {
			return
		}
	}
	if expr.Kind == "binary" && expr.Operator == "in" && expr.Left != nil && (expr.Left.Kind == "string" || expr.Left.Kind == "literal") && expr.Right != nil && expr.Right.Kind == "identifier" {
		varName := expr.Right.Text
		propName := strings.Trim(expr.Left.Text, "\"'`")
		if currType, ok := baseEnv[varName]; ok {
			cleanCurr := strings.TrimPrefix(string(currType), "object:")
			if aliased, ok := typeAliasesIndex[cleanCurr]; ok && strings.Contains(aliased, "|") {
				cleanCurr = aliased
			}
			parts := []string{cleanCurr}
			if strings.Contains(cleanCurr, "|") {
				parts = nil
				for _, p := range strings.Split(cleanCurr, "|") {
					parts = append(parts, strings.TrimSpace(strings.TrimPrefix(p, "object:")))
				}
			}
			var thenTypes []string
			var elseTypes []string
			for _, part := range parts {
				typeDef := part
				if aliased, ok := typeAliasesIndex[part]; ok {
					typeDef = aliased
				}
				var fields []ir.Field
				if f, ok := anonymousObjectFields(typeDef, nil); ok {
					fields = f
				} else if s, ok := shapes[part]; ok {
					fields = s.Fields
				}
				hasProp := false
				for _, f := range fields {
					if f.Name == propName {
						hasProp = true
						break
					}
				}
				if hasProp {
					thenTypes = append(thenTypes, part)
				} else {
					elseTypes = append(elseTypes, part)
				}
			}
			if len(thenTypes) == 1 {
				thenEnv[varName] = toIRType(thenTypes[0])
			} else if len(thenTypes) > 1 {
				thenEnv[varName] = toIRType(strings.Join(thenTypes, " | "))
			}
			if len(elseTypes) == 1 {
				elseEnv[varName] = toIRType(elseTypes[0])
			} else if len(elseTypes) > 1 {
				elseEnv[varName] = toIRType(strings.Join(elseTypes, " | "))
			}
		}
		return
	}
	if expr.Kind == "binary" && expr.Operator == "instanceof" && expr.Left != nil && expr.Right != nil && (expr.Right.Kind == "identifier" || expr.Right.Kind == "type") {
		varName := expr.Left.Text
		className := expr.Right.Text
		if expr.Left.Kind == "identifier" {
			thenEnv[varName] = toIRType(className)
			if currType, ok := baseEnv[varName]; ok {
				cleanCurr := strings.TrimPrefix(string(currType), "object:")
				if strings.Contains(cleanCurr, "|") {
					var remaining []string
					for _, part := range strings.Split(cleanCurr, "|") {
						trimmed := strings.TrimSpace(part)
						if trimmed != className && trimmed != "object:"+className {
							remaining = append(remaining, trimmed)
						}
					}
					if len(remaining) > 0 {
						elseEnv[varName] = toIRType(strings.Join(remaining, " | "))
					}
				}
			}
			return
		}
		if propertyPath := extractPropertyPath(expr.Left); len(propertyPath) > 0 {
			// TypeScript narrows dotted names (for example result.value) as
			// well as identifiers. Keep that fact in the branch environment.
			thenEnv[strings.Join(propertyPath, ".")] = toIRType(className)
		}
		return
	}
	if expr.Kind == "call" && expr.Left != nil && (expr.Left.Kind == "property" || expr.Left.Kind == "member") &&
		expr.Left.Left != nil && expr.Left.Left.Kind == "identifier" && expr.Left.Left.Text == "Array" &&
		expr.Left.Text == "isArray" && len(expr.Arguments) == 1 && expr.Arguments[0].Kind == "identifier" {
		varName := expr.Arguments[0].Text
		thenEnv[varName] = ir.TypeUnknownArray
		return
	}
	if expr.Kind == "binary" && (expr.Operator == "!==" || expr.Operator == "!=") {
		left := expr.Left
		right := expr.Right
		var targetIdent *frontend.SyntaxExpression
		var targetProperty *frontend.SyntaxExpression
		var nullishKind string
		if left != nil && left.Kind == "identifier" {
			targetIdent = left
			if right != nil && (right.Kind == "undefined" || right.Text == "undefined") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if right != nil && (right.Kind == "null" || right.Text == "null") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if left != nil && (left.Kind == "property" || left.Kind == "member") {
			targetProperty = left
			if right != nil && (right.Kind == "undefined" || right.Text == "undefined") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if right != nil && (right.Kind == "null" || right.Text == "null") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if right != nil && right.Kind == "identifier" {
			targetIdent = right
			if left != nil && (left.Kind == "undefined" || left.Text == "undefined") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if left != nil && (left.Kind == "null" || left.Text == "null") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if right != nil && (right.Kind == "property" || right.Kind == "member") {
			targetProperty = right
			if left != nil && (left.Kind == "undefined" || left.Text == "undefined") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if left != nil && (left.Kind == "null" || left.Text == "null") {
				if expr.Operator == "!=" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		}
		if targetIdent != nil && nullishKind != "" {
			varName := targetIdent.Text
			declStr := string(baseEnv["__decl_str."+varName])
			if declStr == "" {
				if mangled, ok := baseEnv["__ident."+varName]; ok {
					declStr = string(baseEnv["__decl_str."+string(mangled)])
				}
			}
			if declStr == "" {
				if topVar, ok := topLevelVars[varName]; ok {
					declStr = topVar.Type
					if declStr == "" {
						declStr = topVar.InferredType
					}
				}
			}
			if declStr != "" && strings.Contains(declStr, "|") {
				narrowedType := narrowUnionTypeString(declStr, nullishKind)
				if narrowedType != "" {
					thenEnv[varName] = narrowedType
					if mangled, ok := thenEnv["__ident."+varName]; ok {
						thenEnv[string(mangled)] = narrowedType
					}
				}
			}
		}
		if targetProperty != nil && nullishKind != "" {
			if propertyPath := extractPropertyPath(targetProperty); len(propertyPath) > 1 {
				if narrowed := narrowPropertyPathType(propertyPath, targetProperty, baseEnv, shapes, nullishKind); narrowed != "" {
					thenEnv[strings.Join(propertyPath, ".")] = narrowed
				}
			}
		}
	}
	if expr.Kind == "binary" && (expr.Operator == "===" || expr.Operator == "==") {
		left := expr.Left
		right := expr.Right
		var targetIdent *frontend.SyntaxExpression
		var targetProperty *frontend.SyntaxExpression
		var nullishKind string
		if left != nil && left.Kind == "identifier" {
			targetIdent = left
			if right != nil && (right.Kind == "undefined" || right.Text == "undefined") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if right != nil && (right.Kind == "null" || right.Text == "null") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if left != nil && (left.Kind == "property" || left.Kind == "member") {
			targetProperty = left
			if right != nil && (right.Kind == "undefined" || right.Text == "undefined") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if right != nil && (right.Kind == "null" || right.Text == "null") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if right != nil && right.Kind == "identifier" {
			targetIdent = right
			if left != nil && (left.Kind == "undefined" || left.Text == "undefined") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if left != nil && (left.Kind == "null" || left.Text == "null") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		} else if right != nil && (right.Kind == "property" || right.Kind == "member") {
			targetProperty = right
			if left != nil && (left.Kind == "undefined" || left.Text == "undefined") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "undefined"
				}
			} else if left != nil && (left.Kind == "null" || left.Text == "null") {
				if expr.Operator == "==" {
					nullishKind = "nullish"
				} else {
					nullishKind = "null"
				}
			}
		}
		if targetIdent != nil && nullishKind != "" {
			varName := targetIdent.Text
			declStr := string(baseEnv["__decl_str."+varName])
			if declStr == "" {
				if mangled, ok := baseEnv["__ident."+varName]; ok {
					declStr = string(baseEnv["__decl_str."+string(mangled)])
				}
			}
			if declStr == "" {
				if topVar, ok := topLevelVars[varName]; ok {
					declStr = topVar.Type
					if declStr == "" {
						declStr = topVar.InferredType
					}
				}
			}
			if declStr != "" && strings.Contains(declStr, "|") {
				narrowedType := narrowUnionTypeString(declStr, nullishKind)
				if narrowedType != "" {
					elseEnv[varName] = narrowedType
					if mangled, ok := elseEnv["__ident."+varName]; ok {
						elseEnv[string(mangled)] = narrowedType
					}
				}
			}
		}
		if targetProperty != nil && nullishKind != "" {
			if propertyPath := extractPropertyPath(targetProperty); len(propertyPath) > 1 {
				if narrowed := narrowPropertyPathType(propertyPath, targetProperty, baseEnv, shapes, nullishKind); narrowed != "" {
					elseEnv[strings.Join(propertyPath, ".")] = narrowed
				}
			}
		}
		// typeof narrowing is applied above so that both the true and false
		// branches retain the useful remainder of a source-level union.
		var propAccess *frontend.SyntaxExpression
		var literalVal *frontend.SyntaxExpression
		if left != nil && (left.Kind == "property" || left.Kind == "member") && right != nil && (right.Kind == "string" || right.Kind == "literal") {
			propAccess = left
			literalVal = right
		} else if right != nil && (right.Kind == "property" || right.Kind == "member") && left != nil && (left.Kind == "string" || left.Kind == "literal") {
			propAccess = right
			literalVal = left
		}
		if propAccess != nil && propAccess.Left != nil && propAccess.Left.Kind == "identifier" && literalVal != nil {
			varName := propAccess.Left.Text
			propName := propAccess.Text
			if propAccess.Right != nil && propAccess.Right.Text != "" {
				propName = propAccess.Right.Text
			}
			valStr := literalVal.Text
			if currType, ok := baseEnv[varName]; ok {
				matched := findMatchingDiscriminatedType(propName, valStr, string(currType), shapes)
				if matched != "" {
					thenEnv[varName] = ir.Type("object:" + matched)
				}
			}
		}
	}
	if expr.Kind == "identifier" {
		varName := expr.Text
		declStr := string(baseEnv["__decl_str."+varName])
		if declStr == "" {
			if mangled, ok := baseEnv["__ident."+varName]; ok {
				declStr = string(baseEnv["__decl_str."+string(mangled)])
			}
		}
		if declStr == "" {
			if topVar, ok := topLevelVars[varName]; ok {
				declStr = topVar.Type
				if declStr == "" {
					declStr = topVar.InferredType
				}
			}
		}
		if declStr != "" && strings.Contains(declStr, "|") {
			narrowedType := narrowUnionTypeString(declStr, "nullish")
			if narrowedType != "" {
				thenEnv[varName] = narrowedType
				if mangled, ok := thenEnv["__ident."+varName]; ok {
					thenEnv[string(mangled)] = narrowedType
				}
			}
		} else if expr.InferredType != "" {
			if nonNull := nonNullishIRType(expr.InferredType); nonNull != "" && nonNull != ir.TypeUnknown && nonNull != ir.TypeVoid {
				thenEnv[varName] = nonNull
				if mangled, ok := thenEnv["__ident."+varName]; ok {
					thenEnv[string(mangled)] = nonNull
				}
			}
		}
	}
	if expr.Kind == "property" || expr.Kind == "member" {
		if propertyPath := extractPropertyPath(expr); len(propertyPath) > 1 {
			if narrowed := narrowPropertyPathType(propertyPath, expr, baseEnv, shapes, "nullish"); narrowed != "" {
				thenEnv[strings.Join(propertyPath, ".")] = narrowed
			}
		}
	}
}
