package lowering

import (
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func resolveThisPropertyClass(expression *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, shapes map[string]ir.ObjectShape, className string) string {
	if expression.Left != nil && expression.Left.Text != "" {
		if varStmt, inTop := topLevelVars[expression.Left.Text]; inTop && varStmt.Type != "" {
			className = varStmt.Type
		} else if t, inEnv := env[expression.Left.Text]; inEnv && string(t) != "" && string(t) != "this" && string(t) != "object" {
			className = strings.TrimPrefix(string(t), "object:")
		}
	}
	if className == "this" || className == "" {
		if t, inEnv := env["this"]; inEnv && string(t) != "this" && string(t) != "object:this" {
			className = strings.TrimPrefix(string(t), "object:")
		} else if function != nil && strings.Contains(function.Name, "_") && !strings.HasPrefix(function.Name, "__closure_") {
			className = strings.Split(function.Name, "_")[0]
		}
	}
	if className == "this" || className == "" {
		for _, sName := range slices.Sorted(maps.Keys(shapes)) {
			s := shapes[sName]
			if fieldIndex(s, expression.Text) >= 0 {
				className = sName
				break
			}
		}
	}

	return className
}

func resolveRegisteredPropertyShape(expression *frontend.SyntaxExpression, shapes map[string]ir.ObjectShape, className string, shape ir.ObjectShape, ok bool) (ir.ObjectShape, bool) {
	if s, exists := registeredShapes[className]; exists {
		shape = s
		shapes[className] = s
		ok = true
	} else if s, exists := anonymousShapes[className]; exists {
		shape = s
		shapes[className] = s
		ok = true
	} else if aliased, hasAlias := typeAliasesIndex[className]; hasAlias {
		if fields, ok2 := anonymousObjectFields(aliased, nil); ok2 {
			shape = ir.ObjectShape{Name: className, Fields: fields}
			shapes[className] = shape
			ok = true
		}
	} else if fields, ok2 := anonymousObjectFields(className, nil); ok2 {
		shape = ir.ObjectShape{Name: className, Fields: fields}
		shapes[className] = shape
		ok = true
	} else if strings.Contains(className, "__") || strings.Contains(className, "_") || strings.Contains(className, "<") {
		baseName := strings.Split(className, "<")[0]
		if strings.Contains(baseName, "__") {
			baseName = strings.Split(baseName, "__")[0]
		} else if strings.Contains(baseName, "_") {
			baseName = strings.Split(baseName, "_")[0]
		}
		if s, exists := shapes[baseName]; exists {
			shape = s
			ok = true
		} else if s, exists := registeredShapes[baseName]; exists {
			shape = s
			ok = true
		} else if baseName == "Partial" || baseName == "Required" || baseName == "Readonly" {
			inner := strings.TrimPrefix(className, baseName+"__")
			if s, exists := shapes[inner]; exists {
				shape = s
				ok = true
			} else if s, exists := registeredShapes[inner]; exists {
				shape = s
				ok = true
			}
		}
	}
	if !ok {
		for _, name := range slices.Sorted(maps.Keys(shapes)) {
			s := shapes[name]
			if (strings.HasPrefix(name, className+"__") || strings.HasPrefix(name, className+"_")) && fieldIndex(s, expression.Text) >= 0 {
				shape = s
				ok = true
				break
			}
		}
	}

	return shape, ok
}

func resolveUnionAliasPropertyShape(expression *frontend.SyntaxExpression, shapes map[string]ir.ObjectShape, className string, shape ir.ObjectShape, ok bool) (string, ir.ObjectShape, bool) {
	unionStr := className
	if typeAliasesIndex != nil && typeAliasesIndex[className] != "" {
		unionStr = typeAliasesIndex[className]
	}
	if strings.Contains(unionStr, "|") {
		for _, m := range splitTopLevelUnion(unionStr) {
			cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m), "object:"))
			var s ir.ObjectShape
			var okS bool
			if s, okS = shapes[cleanM]; !okS {
				s, okS = registeredShapes[cleanM]
			}
			if !okS {
				if fields, okF := anonymousObjectFields(cleanM, nil); okF {
					name := anonymousShapeName(fields)
					s = ir.ObjectShape{Name: name, Fields: fields}
					shapes[name] = s
					okS = true
				}
			}
			if okS && fieldIndex(s, expression.Text) >= 0 {
				shape = s
				className = s.Name
				ok = true
				break
			}
		}
	}

	return className, shape, ok
}

func resolveFieldOwnerShape(expression *frontend.SyntaxExpression, shapes map[string]ir.ObjectShape, className string, shape ir.ObjectShape) (string, ir.ObjectShape) {
	var matchedShape *ir.ObjectShape
	unionStr := className
	if typeAliasesIndex != nil && typeAliasesIndex[className] != "" {
		unionStr = typeAliasesIndex[className]
	}
	if strings.Contains(unionStr, "|") {
		for _, m := range splitTopLevelUnion(unionStr) {
			cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m), "object:"))
			var s ir.ObjectShape
			var okS bool
			if s, okS = shapes[cleanM]; !okS {
				s, okS = registeredShapes[cleanM]
			}
			if !okS {
				if fields, okF := anonymousObjectFields(cleanM, nil); okF {
					name := anonymousShapeName(fields)
					s = ir.ObjectShape{Name: name, Fields: fields}
					shapes[name] = s
					okS = true
				}
			}
			if okS && fieldIndex(s, expression.Text) >= 0 {
				matchedShape = &s
				break
			}
		}
	}
	if matchedShape != nil {
		shape = *matchedShape
		className = shape.Name
	}

	return className, shape
}
