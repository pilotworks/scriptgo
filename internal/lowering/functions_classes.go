package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// indexClassDeclaration records the constructor and method signatures of one
// class declaration, top-level or inside a namespace. Method parameters start
// at slot 1 when the lowered function takes `this` first.
func indexClassDeclaration(fileName string, class frontend.SyntaxClass, hierarchy map[string]ClassMeta, index map[string]ir.Function, functionsByFile map[string][]indexedFunction) {
	className := classIdentityForPath(fileName, class.Name)
	thisParameter := ir.Parameter{Name: "this", Type: ir.Type("object:" + className)}
	if class.Constructor != nil {
		ctorMangled := className + "_constructor"
		ctorFn := ir.Function{Name: ctorMangled, ReturnType: ir.TypeVoid, Parameters: []ir.Parameter{thisParameter}}
		for pIdx, parameter := range class.Constructor.Parameters {
			recordParameterDefault(ctorMangled, pIdx+1, parameter)
			if parameter.Rest {
				restParamsIndex[ctorMangled] = true
			}
			ctorFn.Parameters = append(ctorFn.Parameters, ir.Parameter{Name: parameter.Name, Type: variableStorageType(toIRTypeForPath(fileName, parameter.Type))})
		}
		index[ctorMangled] = ctorFn
		functionsByFile[fileName] = append(functionsByFile[fileName], indexedFunction{Function: ctorFn, PublicName: ctorFn.Name})
	}

	for _, method := range getInheritedMethods(className, hierarchy) {
		var mangled string
		var function ir.Function
		switch {
		case method.IsStatic:
			mangled = className + "_static_" + method.Name
			function = ir.Function{Name: mangled, ReturnType: toIRTypeForPath(fileName, method.Type)}
			if function.ReturnType == "" {
				function.ReturnType = ir.TypeVoid
			}
			for pIdx, parameter := range method.Parameters {
				recordParameterDefault(mangled, pIdx, parameter)
				if parameter.Rest {
					restParamsIndex[mangled] = true
				}
				function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: variableStorageType(toIRTypeForPath(fileName, parameter.Type))})
			}
			index[mangled] = function
			index[className+"."+method.Name] = function
		case method.Kind == "get":
			mangled = className + "_get_" + method.Name
			function = ir.Function{Name: mangled, ReturnType: toIRTypeForPath(fileName, method.Type), Parameters: []ir.Parameter{thisParameter}}
		case method.Kind == "set":
			mangled = className + "_set_" + method.Name
			function = ir.Function{Name: mangled, ReturnType: ir.TypeVoid, Parameters: []ir.Parameter{thisParameter}}
			if len(method.Parameters) > 0 {
				function.Parameters = append(function.Parameters, ir.Parameter{Name: method.Parameters[0].Name, Type: variableStorageType(toIRTypeForPath(fileName, method.Parameters[0].Type))})
			}
		default:
			mangled = methodImplementationName(className, method.Name)
			retType := toIRTypeForPath(fileName, method.Type)
			if method.Type == "this" || retType == "this" || retType == "object:this" {
				retType = ir.Type("object:" + className)
			}
			function = ir.Function{Name: mangled, ReturnType: retType, Parameters: []ir.Parameter{thisParameter}}
			if function.ReturnType == "" {
				function.ReturnType = ir.TypeVoid
			}
			for pIdx, parameter := range method.Parameters {
				recordParameterDefault(mangled, pIdx+1, parameter)
				if parameter.Rest {
					restParamsIndex[mangled] = true
				}
				function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: variableStorageType(toIRTypeForPath(fileName, parameter.Type))})
			}
		}
		if !method.IsAbstract {
			index[mangled] = function
			functionsByFile[fileName] = append(functionsByFile[fileName], indexedFunction{Function: function, PublicName: function.Name})
		}
	}
}

// recordParameterDefault registers the value a call site supplies for an
// omitted argument: the parameter initializer, or undefined for an optional
// parameter. Required parameters record nothing.
func recordParameterDefault(function string, slot int, parameter frontend.SyntaxParameter) {
	initializer := parameter.Initializer
	if initializer == nil {
		if !parameter.Optional {
			return
		}
		initializer = &frontend.SyntaxExpression{Kind: "undefined"}
	}
	if defaultParamsIndex[function] == nil {
		defaultParamsIndex[function] = map[int]*frontend.SyntaxExpression{}
	}
	defaultParamsIndex[function][slot] = initializer
}
