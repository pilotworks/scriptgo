package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryFillConstructorDefaults(path string, expression *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, ctor ir.Function, ctorName string, args *[]string) (string, ir.Type, bool, error) {
	defaults := defaultParamsIndex[ctorName]
	if defaults == nil {
		defaults = defaultParamsIndex[strings.Split(ctorName, "__")[0]]
	}
	for i := len(*args); i < len(ctor.Parameters); i++ {
		var val string
		var valType ir.Type
		paramType := ctor.Parameters[i].Type
		if defaults != nil && defaults[i] != nil {
			defExpr := defaults[i]
			if paramType == ir.TypeNumber && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
				numConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: numConst,
					Value:  "0",
					Span:   toIRSpan(path, defExpr.Span),
				})
				val = numConst
				valType = ir.TypeNumber
			} else if paramType == ir.TypeBool && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
				boolConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeBool,
					Result: boolConst,
					Value:  "false",
					Span:   toIRSpan(path, defExpr.Span),
				})
				val = boolConst
				valType = ir.TypeBool
			} else if paramType == ir.TypeBigInt && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
				biConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeBigInt,
					Result: biConst,
					Value:  "0",
					Span:   toIRSpan(path, defExpr.Span),
				})
				val = biConst
				valType = ir.TypeBigInt
			} else if (strings.HasPrefix(string(paramType), "object:") || isPointerLikeType(paramType)) && (defExpr.Kind == "null" || defExpr.Kind == "undefined") {
				nullConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   paramType,
					Result: nullConst,
					Value:  map[bool]string{true: "undefined", false: "null"}[defExpr.Kind == "undefined"],
					Span:   toIRSpan(path, defExpr.Span),
				})
				val = nullConst
				valType = paramType
			} else {
				v, vt, err := lowerExpression(path, defExpr, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				val = v
				valType = vt
			}
		}
		if val == "" {
			val = nextTemp(counter)
			valType = paramType
			defStr := "0"
			if paramType == ir.TypeBool {
				defStr = "false"
			} else if paramType == ir.TypeString {
				defStr = ""
			} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
				defStr = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   paramType,
				Result: val,
				Value:  defStr,
				Span:   toIRSpan(path, expression.Span),
			})
		}
		if paramType == ir.TypeUnknown && valType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			val = boxed
		}
		*args = append(*args, val)
	}

	return "", "", false, nil
}
