package lowering

import (
	"sync"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var (
	metadataMu             sync.RWMutex
	staticMetadataRegistry = map[string]map[string]map[string]interface{}{}
)

func clearMetadataRegistry() {
	metadataMu.Lock()
	defer metadataMu.Unlock()
	staticMetadataRegistry = make(map[string]map[string]map[string]interface{})
}

func registerStaticMetadata(className, memberName, key string, val interface{}) {
	metadataMu.Lock()
	defer metadataMu.Unlock()
	if _, ok := staticMetadataRegistry[className]; !ok {
		staticMetadataRegistry[className] = make(map[string]map[string]interface{})
	}
	if _, ok := staticMetadataRegistry[className][memberName]; !ok {
		staticMetadataRegistry[className][memberName] = make(map[string]interface{})
	}
	staticMetadataRegistry[className][memberName][key] = val
}

func lookupStaticMetadata(className, memberName, key string) (interface{}, bool) {
	metadataMu.RLock()
	defer metadataMu.RUnlock()
	if classMap, ok := staticMetadataRegistry[className]; ok {
		if memberMap, ok := classMap[memberName]; ok {
			if val, ok := memberMap[key]; ok {
				return val, true
			}
		}
	}
	return nil, false
}

func hasStaticMetadata(className, memberName, key string) bool {
	_, found := lookupStaticMetadata(className, memberName, key)
	return found
}

func collectClassMetadata(class *typescriptgo.SyntaxClass) {
	if class == nil {
		return
	}
	className := class.Name

	// 1. Class-level decorators
	for _, dec := range class.Decorators {
		registerDecoratorMetadata(className, "", dec)
	}

	// 2. Constructor parameter decorators
	if class.Constructor != nil {
		var paramTypes []string
		for _, p := range class.Constructor.Parameters {
			paramTypes = append(paramTypes, p.Type)
			for _, dec := range p.Decorators {
				registerDecoratorMetadata(className, "constructor", dec)
			}
		}
		if len(paramTypes) > 0 {
			registerStaticMetadata(className, "", "design:paramtypes", paramTypes)
		}
	}

	// 3. Methods & Accessors
	for _, m := range class.Methods {
		var paramTypes []string
		for _, p := range m.Parameters {
			if p.Name != "this" {
				paramTypes = append(paramTypes, p.Type)
			}
			for _, dec := range p.Decorators {
				registerDecoratorMetadata(className, m.Name, dec)
			}
		}
		if len(m.Decorators) > 0 {
			registerStaticMetadata(className, m.Name, "design:type", "function")
			if len(paramTypes) > 0 {
				registerStaticMetadata(className, m.Name, "design:paramtypes", paramTypes)
			}
			if m.Type != "" {
				registerStaticMetadata(className, m.Name, "design:returntype", m.Type)
			}
			for _, dec := range m.Decorators {
				registerDecoratorMetadata(className, m.Name, dec)
			}
		}
	}

	// 4. Fields
	for _, f := range class.Fields {
		if len(f.Decorators) > 0 {
			if f.Type != "" {
				registerStaticMetadata(className, f.Name, "design:type", f.Type)
			}
			for _, dec := range f.Decorators {
				registerDecoratorMetadata(className, f.Name, dec)
			}
		}
	}
}

func registerDecoratorMetadata(className, memberName string, dec typescriptgo.SyntaxDecorator) {
	if dec.DesignType != "" {
		registerStaticMetadata(className, memberName, "design:type", dec.DesignType)
	}
	if len(dec.ParamTypes) > 0 {
		registerStaticMetadata(className, memberName, "design:paramtypes", dec.ParamTypes)
	}
	if dec.ReturnType != "" {
		registerStaticMetadata(className, memberName, "design:returntype", dec.ReturnType)
	}

	// Extract metadata from custom decorators, e.g. @Reflect.metadata(key, val) or @Role("admin")
	if dec.Name != "" {
		if (dec.Name == "metadata" || dec.Name == "Reflect.metadata") && len(dec.Arguments) >= 2 {
			keyArg := dec.Arguments[0]
			valArg := dec.Arguments[1]
			if keyArg != nil && keyArg.Kind == "string" {
				if valArg != nil && valArg.Kind == "string" {
					registerStaticMetadata(className, memberName, keyArg.Text, valArg.Text)
				} else if valArg != nil && valArg.Kind == "number" {
					registerStaticMetadata(className, memberName, keyArg.Text, valArg.Text)
				} else if valArg != nil {
					registerStaticMetadata(className, memberName, keyArg.Text, valArg.Text)
				}
			}
		} else if len(dec.Arguments) == 1 && dec.Arguments[0] != nil {
			arg := dec.Arguments[0]
			registerStaticMetadata(className, memberName, dec.Name, arg.Text)
		} else {
			registerStaticMetadata(className, memberName, dec.Name, "true")
		}
	}
}

func lowerClassDecorators(path string, class *typescriptgo.SyntaxClass, main *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if class == nil {
		return nil
	}

	// Collect static metadata for reflection
	collectClassMetadata(class)

	// Execute Method & Field Decorators at class initialization
	for _, m := range class.Methods {
		for _, dec := range m.Decorators {
			if err := lowerDecoratorInvocation(path, class.Name, m.Name, "method", dec, main, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	}

	for _, f := range class.Fields {
		for _, dec := range f.Decorators {
			if err := lowerDecoratorInvocation(path, class.Name, f.Name, "field", dec, main, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	}

	// Class Decorators
	for _, dec := range class.Decorators {
		if err := lowerDecoratorInvocation(path, class.Name, "", "class", dec, main, env, counter, shapes, signatures); err != nil {
			return err
		}
	}

	return nil
}

func lowerDecoratorInvocation(path, className, memberName, kind string, dec typescriptgo.SyntaxDecorator, fn *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if dec.Expression == nil {
		return nil
	}

	// For factory decorators like @logged("tag"), evaluate the factory call directly
	if dec.Expression.Kind == "call" {
		_, _, err := lowerExpression(path, dec.Expression, "", fn, env, counter, shapes, signatures)
		return err
	}

	decFnName := dec.Name
	if decFnName == "" || decFnName == "metadata" || decFnName == "Reflect.metadata" {
		return nil
	}

	// If the decorator function exists in user code, invoke it during class initialization
	if sig, exists := signatures[decFnName]; exists {
		var args []*typescriptgo.SyntaxExpression
		for pIdx, param := range sig.Parameters {
			if pIdx == 0 {
				if param.Type == ir.TypeString {
					args = append(args, &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "string", Text: memberName})
				} else if param.Type == ir.TypeNumber {
					args = append(args, &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "number", Text: "0"})
				} else {
					args = append(args, &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "undefined"})
				}
			} else if pIdx == 1 {
				if param.Type == ir.TypeString {
					args = append(args, &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "string", Text: kind})
				} else {
					args = append(args, &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "undefined"})
				}
			}
		}
		callExpr := &typescriptgo.SyntaxExpression{
			Span:      dec.Span,
			Kind:      "call",
			Left:      &typescriptgo.SyntaxExpression{Span: dec.Span, Kind: "identifier", Text: decFnName},
			Arguments: args,
		}
		_, _, err := lowerExpression(path, callExpr, "", fn, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
	}

	return nil
}
