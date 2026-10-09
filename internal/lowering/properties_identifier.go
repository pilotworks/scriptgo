package lowering

import (
	"strconv"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerIdentifierReceiverProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, className string) (string, ir.Type, bool, error) {
	propKey := expression.Left.Text + "." + expression.Text

	if propKey == "process.argv" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeStringArray,
			Result: *result,
			Callee: "__process.argv",
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeStringArray, true, nil
	}
	if propKey == "process.env" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpCall, Type: ir.TypeObject, Result: *result, Callee: "__process.env_obj", Span: toIRSpan(path, expression.Span),
		})
		return *result, ir.TypeObject, true, nil
	}

	// 2. Check static getters
	if getter, getterName, ok := findGetterInHierarchy(className, expression.Text, signatures, classHierarchy); ok && getter.Parameters == nil {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   getter.ReturnType,
			Result: *result,
			Callee: getterName,
			Span:   toIRSpan(path, expression.Span),
		})
		return *result, getter.ReturnType, true, nil
	}

	// 2.5. Check function reflection (.length and .name)
	if fn, isFunc := resolveFunctionSignature(path, expression.Left.Text, signatures); isFunc {
		if expression.Text == "length" {
			arity := len(fn.Parameters)
			if defaults, hasDefaults := defaultParamsIndex[fn.Name]; hasDefaults {
				for i := 0; i < len(fn.Parameters); i++ {
					if _, isDefault := defaults[i]; isDefault {
						arity = i
						break
					}
				}
			}
			if restParamsIndex[fn.Name] && arity == len(fn.Parameters) && arity > 0 {
				arity = len(fn.Parameters) - 1
			}
			if *result == "" {
				*result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: *result,
				Value:  strconv.Itoa(arity),
				Span:   toIRSpan(path, expression.Span),
			})
			return *result, ir.TypeNumber, true, nil
		}
		if expression.Text == "name" {
			if *result == "" {
				*result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: *result,
				Value:  functionPublicName(fn.Name),
				Span:   toIRSpan(path, expression.Span),
			})
			return *result, ir.TypeString, true, nil
		}
	}

	// 3. Check static fields in class hierarchy
	if meta, ok := classHierarchy[className]; ok {
		if staticField, isStatic := meta.Statics[expression.Text]; isStatic {
			staticVar := staticFieldGlobal(className, expression.Text)
			typ := toIRTypeForPath(path, staticField.Type)
			if typ == "" {
				typ = ir.TypeNumber
			}
			return staticVar, typ, true, nil
		}
	}

	// 4. Check shape const fields (e.g. Enums)
	if shape, ok := shapes[className]; ok {
		for _, field := range shape.Fields {
			if field.Name == expression.Text && field.Value != "" {
				if *result == "" {
					*result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   field.Type,
					Result: *result,
					Value:  field.Value,
					Span:   toIRSpan(path, expression.Span),
				})
				return *result, field.Type, true, nil
			}
		}
	}

	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "argv" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: *result, Callee: "__process.argv", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeStringArray, true, nil
	}
	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "version" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: *result, Callee: "__process.version", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeString, true, nil
	}
	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "pid" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: *result, Callee: "__process.pid", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeNumber, true, nil
	}
	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "ppid" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: *result, Callee: "__process.ppid", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeNumber, true, nil
	}
	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "platform" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: *result, Callee: "__os.platform", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeString, true, nil
	}
	if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "arch" {
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: *result, Callee: "__os.arch", Args: nil, Span: toIRSpan(path, expression.Span)})
		return *result, ir.TypeString, true, nil
	}

	return "", "", false, nil
}
