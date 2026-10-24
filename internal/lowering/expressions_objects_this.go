package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// objectMethodUsesThis reports an object literal member whose value is a
// method that reads `this`.
func objectMethodUsesThis(member *frontend.SyntaxExpression) bool {
	return member.Left != nil && isObjectLiteralMethod(member.Left) && usesThis(member.Left.Function.Body, "object")
}

// bindObjectMethodThis returns a member's method value with `this` read as
// object, the IR value of the literal being built: object literal lowering
// creates the object first and then sets such methods, which capture it.
// Unlike JavaScript, a method detached from its object or called on another
// receiver still sees the object it was defined on.
func bindObjectMethodThis(member *frontend.SyntaxExpression, object string) *frontend.SyntaxExpression {
	method := *member.Left.Function
	method.Body = replaceThisInStatements(method.Body, "object", object)
	value := *member.Left
	value.Function = &method
	return &value
}

// objectThisLocal copies the object an object literal is building into a
// local its methods capture as `this`. Temporaries cannot be captured by
// name: the method's own temporaries reuse those names.
func objectThisLocal(path string, span frontend.SourceSpan, object string, objectType ir.Type, function *ir.Function, env map[string]ir.Type, counter *int) string {
	name := fmt.Sprintf("__object_this_%d_%d", span.Start, *counter)
	*counter++
	function.Locals = append(function.Locals, ir.Parameter{Name: name, Type: objectType})
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: objectType, Result: name, Args: []string{object}, Span: toIRSpan(path, span)})
	env[name] = objectType
	return name
}
