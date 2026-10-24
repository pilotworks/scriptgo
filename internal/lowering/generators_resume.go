package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

// A generator body is rewritten into a step closure that runs from its last
// suspension point to the next one. Each statement-level yield gets an id
// k >= 1; the factory scope holds `state` (0 before the first next(), k while
// suspended at yield k, -1 once finished) and every local whose scope spans a
// yield, so the closure's captures keep them between calls. On entry the
// closure copies state into `resume` and marks the generator finished; while
// resume is non-zero, statements outside the path to yield `resume` are
// skipped, loop conditions are bypassed and branches follow the yield's
// range, until the yield itself clears resume and execution continues
// normally. A yield stores its id in state and returns { done: false }.
//
// Hoisted locals are bound once per generator instead of once per loop
// iteration, so closures created in a loop that also yields share them.
type generatorResume struct {
	path       string
	span       frontend.SourceSpan
	state      string
	resume     string
	yieldType  string
	nextID     int
	temps      int
	hoisted    []frontend.SyntaxStatement
	hoistTypes map[string]string
	parameters map[string]bool
}

// generatorStep rewrites a generator's body into the declarations its factory
// runs before creating the generator (state and hoisted locals) and the step
// closure that advances it. yieldType is the IR type of yielded values.
func generatorStep(path string, statement frontend.SyntaxStatement, yieldType string) ([]frontend.SyntaxStatement, *frontend.SyntaxExpression, error) {
	suffix := fmt.Sprintf("%d", statement.Span.Start)
	g := &generatorResume{
		path:       path,
		span:       statement.Span,
		state:      "__gen_state_" + suffix,
		resume:     "__gen_resume_" + suffix,
		yieldType:  yieldType,
		hoistTypes: map[string]string{},
		parameters: map[string]bool{},
	}
	for _, parameter := range statement.Parameters {
		g.parameters[parameter.Name] = true
	}
	// Destructured parameters are bound when the generator function is
	// called, so their errors throw there rather than at the first next().
	bindings := min(statement.ParameterBindings, len(statement.Body))
	body, err := g.list(statement.Body[bindings:])
	if err != nil {
		return nil, nil, err
	}
	step := []frontend.SyntaxStatement{
		g.declare(g.resume, "number", g.ident(g.state)),
		g.assign(g.state, g.number("-1")),
	}
	step = append(step, body...)
	step = append(step, g.result(true, nil))
	prelude := append(append(append([]frontend.SyntaxStatement{}, statement.Body[:bindings]...), g.declare(g.state, "number", g.number("0"))), g.hoisted...)
	closure := &frontend.SyntaxExpression{Span: g.span, Kind: "arrow_function", Function: &frontend.SyntaxStatement{
		Span: g.span,
		Kind: "function",
		Type: "object:IteratorResult_" + yieldType,
		Body: step,
	}}
	return prelude, closure, nil
}

// list rewrites a statement list that contains a yield. Runs of statements
// without one execute only when not resuming.
func (g *generatorResume) list(statements []frontend.SyntaxStatement) ([]frontend.SyntaxStatement, error) {
	var out, pending []frontend.SyntaxStatement
	flush := func() {
		if len(pending) > 0 {
			out = append(out, g.ifStmt(g.compare(g.resume, "==", "0"), pending, nil))
			pending = nil
		}
	}
	for index, statement := range statements {
		if statement.Kind == "function" || statement.Kind == "class" {
			// Declarations are hoisted by lowering; keep them visible on every step.
			flush()
			out = append(out, statement)
			continue
		}
		if yieldCount(statement) == 0 {
			if err := g.checkNoNestedYield(statement); err != nil {
				return nil, err
			}
			rewritten, err := g.hoist(statement, statements[index+1:])
			if err != nil {
				return nil, err
			}
			pending = append(pending, g.returns(rewritten)...)
			continue
		}
		flush()
		rewritten, err := g.statement(statement)
		if err != nil {
			return nil, err
		}
		out = append(out, rewritten...)
	}
	flush()
	return out, nil
}

// statement rewrites one statement that contains a yield.
func (g *generatorResume) statement(statement frontend.SyntaxStatement) ([]frontend.SyntaxStatement, error) {
	lo := g.nextID + 1
	hi := g.nextID + yieldCount(statement)
	entered := g.or(g.compare(g.resume, "==", "0"), g.inRange(lo, hi))
	switch statement.Kind {
	case "expression":
		switch statement.Expression.Kind {
		case "yield":
			if statement.Expression.Left != nil {
				if err := g.checkNoYieldExpression(statement.Expression.Left); err != nil {
					return nil, err
				}
			}
			g.nextID++
			id := fmt.Sprintf("%d", g.nextID)
			suspend := []frontend.SyntaxStatement{g.assign(g.state, g.number(id)), g.result(false, statement.Expression.Left)}
			resumed := []frontend.SyntaxStatement{g.assign(g.resume, g.number("0"))}
			return []frontend.SyntaxStatement{g.ifStmt(g.compare(g.resume, "==", id), resumed, []frontend.SyntaxStatement{
				g.ifStmt(g.compare(g.resume, "==", "0"), suspend, nil),
			})}, nil
		case "yield_star":
			// yield* source behaves as for (const v of source) yield v.
			g.temps++
			item := fmt.Sprintf("__gen_delegate_%s_%d", g.suffix(), g.temps)
			loop := frontend.SyntaxStatement{
				Span: statement.Span, Kind: "forof", Name: item, Type: g.yieldType, InferredType: g.yieldType,
				Expression: statement.Expression.Left,
				Body: []frontend.SyntaxStatement{{Span: statement.Span, Kind: "expression", Expression: &frontend.SyntaxExpression{
					Span: statement.Span, Kind: "yield", Left: g.ident(item),
				}}},
			}
			return g.statement(loop)
		}
	case "block":
		body, err := g.list(statement.Body)
		if err != nil {
			return nil, err
		}
		return []frontend.SyntaxStatement{g.ifStmt(entered, body, nil)}, nil
	case "if":
		thenHi := lo + yieldCountList(statement.Then) - 1
		then, err := g.list(statement.Then)
		if err != nil {
			return nil, err
		}
		otherwise, err := g.list(statement.Else)
		if err != nil {
			return nil, err
		}
		takeThen := g.and(g.compare(g.resume, "==", "0"), statement.Expression)
		if thenHi >= lo {
			takeThen = g.or(takeThen, g.inRange(lo, thenHi))
		}
		branch := g.ifStmt(takeThen, then, otherwise)
		return []frontend.SyntaxStatement{g.ifStmt(entered, []frontend.SyntaxStatement{branch}, nil)}, nil
	case "while":
		body, err := g.list(statement.Body)
		if err != nil {
			return nil, err
		}
		loop := statement
		loop.Expression = g.or(g.compare(g.resume, "!=", "0"), g.condition(statement.Expression))
		loop.Body = body
		return []frontend.SyntaxStatement{g.ifStmt(entered, []frontend.SyntaxStatement{loop}, nil)}, nil
	case "dowhile":
		body, err := g.list(statement.Body)
		if err != nil {
			return nil, err
		}
		loop := statement
		loop.Body = body
		return []frontend.SyntaxStatement{g.ifStmt(entered, []frontend.SyntaxStatement{loop}, nil)}, nil
	case "forof":
		return g.statement(g.forOf(statement))
	case "try":
		if statement.HasFinally || yieldCountList(statement.Catch) > 0 {
			return nil, g.unsupported(statement.Span, "a yield inside a catch block or a try statement with finally")
		}
		body, err := g.list(statement.Body)
		if err != nil {
			return nil, err
		}
		try := statement
		try.Body = body
		try.Catch = g.returnsList(statement.Catch)
		return []frontend.SyntaxStatement{g.ifStmt(entered, []frontend.SyntaxStatement{try}, nil)}, nil
	}
	return nil, g.unsupported(statement.Span, "a yield inside a "+statement.Kind+" statement")
}

// forOf rewrites a for-of loop around a yield into a while loop whose
// iteration state lives in hoisted locals: an index into an array (other
// finite iterables are spread into one) or a generator advanced by next().
func (g *generatorResume) forOf(statement frontend.SyntaxStatement) frontend.SyntaxStatement {
	g.temps++
	source := fmt.Sprintf("__gen_source_%s_%d", g.suffix(), g.temps)
	itemType := statement.Type
	if itemType == "" {
		itemType = statement.InferredType
	}
	sourceType := ""
	if statement.Expression != nil {
		sourceType = strings.TrimSpace(statement.Expression.InferredType)
	}
	span := statement.Span
	item := g.declare(statement.Name, itemType, nil)
	if isGeneratorTypeName(sourceType) {
		result := source + "_result"
		advance := frontend.SyntaxStatement{Span: span, Kind: "block", Body: []frontend.SyntaxStatement{
			g.declare(result, "", &frontend.SyntaxExpression{Span: span, Kind: "call", Left: g.property(g.ident(source), "next", "")}),
			g.ifStmt(g.property(g.ident(result), "done", "boolean"), []frontend.SyntaxStatement{{Span: span, Kind: "break"}}, nil),
			g.assign(statement.Name, g.property(g.ident(result), "value", itemType)),
		}}
		body := append([]frontend.SyntaxStatement{item, advance}, statement.Body...)
		return frontend.SyntaxStatement{Span: span, Kind: "block", Body: []frontend.SyntaxStatement{
			g.declare(source, sourceType, statement.Expression),
			{Span: span, Kind: "while", Label: statement.Label, Expression: &frontend.SyntaxExpression{Span: span, Kind: "bool", Text: "true", InferredType: "boolean"}, Body: body},
		}}
	}
	sourceValue := statement.Expression
	if !strings.HasSuffix(sourceType, "[]") && !strings.HasPrefix(sourceType, "Array<") {
		sourceType = itemType + "[]"
		sourceValue = &frontend.SyntaxExpression{Span: span, Kind: "array", InferredType: sourceType, Arguments: []*frontend.SyntaxExpression{
			{Span: span, Kind: "spread", Left: statement.Expression},
		}}
	}
	index := source + "_index"
	item.Expression = &frontend.SyntaxExpression{Span: span, Kind: "index", Left: g.ident(source), Right: g.ident(index), InferredType: itemType}
	step := g.assign(index, &frontend.SyntaxExpression{Span: span, Kind: "binary", Operator: "+", Left: g.ident(index), Right: g.number("1"), InferredType: "number"})
	body := append([]frontend.SyntaxStatement{item, step}, statement.Body...)
	return frontend.SyntaxStatement{Span: span, Kind: "block", Body: []frontend.SyntaxStatement{
		g.declare(source, sourceType, sourceValue),
		g.declare(index, "number", g.number("0")),
		{Span: span, Kind: "while", Label: statement.Label, Expression: &frontend.SyntaxExpression{
			Span: span, Kind: "binary", Operator: "<", InferredType: "boolean",
			Left: g.ident(index), Right: g.property(g.ident(source), "length", "number"),
		}, Body: body},
	}}
}

// hoist moves a declaration whose scope reaches a later yield into the
// factory scope and turns it into an assignment.
func (g *generatorResume) hoist(statement frontend.SyntaxStatement, rest []frontend.SyntaxStatement) (frontend.SyntaxStatement, error) {
	if statement.Kind != "variable" && statement.Kind != "using" && statement.Kind != "await_using" {
		return statement, nil
	}
	if yieldCountList(rest) == 0 {
		return statement, nil
	}
	if statement.Kind != "variable" {
		return statement, g.unsupported(statement.Span, "a yield in the scope of a using declaration")
	}
	typ := statement.Type
	if typ == "" {
		typ = statement.InferredType
	}
	if g.parameters[statement.Name] {
		return statement, g.unsupported(statement.Span, fmt.Sprintf("a yield in the scope of %q, which shadows a parameter", statement.Name))
	}
	if previous, ok := g.hoistTypes[statement.Name]; ok {
		if previous != typ {
			return statement, g.unsupported(statement.Span, fmt.Sprintf("a yield in the scope of %q, which is redeclared with another type", statement.Name))
		}
	} else {
		g.hoistTypes[statement.Name] = typ
		g.hoisted = append(g.hoisted, g.declare(statement.Name, typ, nil))
	}
	if statement.Expression == nil {
		return frontend.SyntaxStatement{Span: statement.Span, Kind: "block"}, nil
	}
	return g.assign(statement.Name, statement.Expression), nil
}

// returns makes each `return` in a statement finish the generator.
func (g *generatorResume) returns(statement frontend.SyntaxStatement) []frontend.SyntaxStatement {
	switch statement.Kind {
	case "return":
		if statement.Expression == nil {
			return []frontend.SyntaxStatement{g.result(true, nil)}
		}
		// The return value is evaluated; next() reports it only when it has
		// the yielded type.
		if toIRType(statement.Expression.InferredType) == toIRType(g.yieldType) {
			return []frontend.SyntaxStatement{g.result(true, statement.Expression)}
		}
		return []frontend.SyntaxStatement{{Span: statement.Span, Kind: "expression", Expression: statement.Expression}, g.result(true, nil)}
	case "function", "class":
		return []frontend.SyntaxStatement{statement}
	}
	statement.Body = g.returnsList(statement.Body)
	statement.Then = g.returnsList(statement.Then)
	statement.Else = g.returnsList(statement.Else)
	statement.Catch = g.returnsList(statement.Catch)
	statement.Finally = g.returnsList(statement.Finally)
	if len(statement.Cases) > 0 {
		cases := make([]frontend.SyntaxSwitchCase, len(statement.Cases))
		for index, c := range statement.Cases {
			c.Statements = g.returnsList(c.Statements)
			cases[index] = c
		}
		statement.Cases = cases
	}
	return []frontend.SyntaxStatement{statement}
}

func (g *generatorResume) returnsList(statements []frontend.SyntaxStatement) []frontend.SyntaxStatement {
	if statements == nil {
		return nil
	}
	out := make([]frontend.SyntaxStatement, 0, len(statements))
	for _, statement := range statements {
		out = append(out, g.returns(statement)...)
	}
	return out
}

// yieldCount counts the statement-level yields in a statement, outside nested
// functions and classes.
func yieldCount(statement frontend.SyntaxStatement) int {
	switch statement.Kind {
	case "function", "class":
		return 0
	case "expression":
		if statement.Expression != nil && (statement.Expression.Kind == "yield" || statement.Expression.Kind == "yield_star") {
			return 1
		}
	}
	count := yieldCountList(statement.Body) + yieldCountList(statement.Then) + yieldCountList(statement.Else) + yieldCountList(statement.Catch) + yieldCountList(statement.Finally)
	for _, c := range statement.Cases {
		count += yieldCountList(c.Statements)
	}
	return count
}

func yieldCountList(statements []frontend.SyntaxStatement) int {
	count := 0
	for _, statement := range statements {
		count += yieldCount(statement)
	}
	return count
}

// checkNoNestedYield rejects a yield used as a value anywhere in a statement
// that has no statement-level yield.
func (g *generatorResume) checkNoNestedYield(statement frontend.SyntaxStatement) error {
	if statement.Kind == "function" || statement.Kind == "class" {
		return nil
	}
	for _, expression := range []*frontend.SyntaxExpression{statement.Expression, statement.Left, statement.Right} {
		if err := g.checkNoYieldExpression(expression); err != nil {
			return err
		}
	}
	for _, list := range [][]frontend.SyntaxStatement{statement.Body, statement.Step, statement.Then, statement.Else, statement.Catch, statement.Finally} {
		for _, nested := range list {
			if err := g.checkNoNestedYield(nested); err != nil {
				return err
			}
		}
	}
	for _, c := range statement.Cases {
		if err := g.checkNoYieldExpression(c.Expression); err != nil {
			return err
		}
		for _, nested := range c.Statements {
			if err := g.checkNoNestedYield(nested); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generatorResume) checkNoYieldExpression(expression *frontend.SyntaxExpression) error {
	if expression == nil {
		return nil
	}
	if expression.Kind == "yield" || expression.Kind == "yield_star" {
		return g.unsupported(expression.Span, "a yield used as a value")
	}
	for _, nested := range append([]*frontend.SyntaxExpression{expression.Left, expression.Right, expression.WhenTrue, expression.WhenFalse}, expression.Arguments...) {
		if err := g.checkNoYieldExpression(nested); err != nil {
			return err
		}
	}
	return nil
}

func (g *generatorResume) unsupported(span frontend.SourceSpan, what string) error {
	return sourceError(g.path, span, fmt.Errorf("generators do not support %s", what))
}

func isGeneratorTypeName(typeName string) bool {
	return strings.HasPrefix(typeName, "Generator<") || strings.HasPrefix(typeName, "IterableIterator<") || strings.HasPrefix(typeName, "Iterator<") || strings.HasPrefix(typeName, "IteratorObject<")
}

func (g *generatorResume) suffix() string { return fmt.Sprintf("%d", g.span.Start) }

func (g *generatorResume) ident(name string) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "identifier", Text: name}
}

func (g *generatorResume) number(value string) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "number", Text: value, InferredType: "number"}
}

func (g *generatorResume) property(object *frontend.SyntaxExpression, name, typ string) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "property", Left: object, Text: name, InferredType: typ}
}

func (g *generatorResume) compare(name, operator, value string) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "binary", Operator: operator, Left: g.ident(name), Right: g.number(value), InferredType: "boolean"}
}

func (g *generatorResume) and(left, right *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "binary", Operator: "&&", Left: left, Right: right, InferredType: "boolean"}
}

func (g *generatorResume) or(left, right *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	return &frontend.SyntaxExpression{Span: g.span, Kind: "binary", Operator: "||", Left: left, Right: right, InferredType: "boolean"}
}

func (g *generatorResume) inRange(lo, hi int) *frontend.SyntaxExpression {
	if lo == hi {
		return g.compare(g.resume, "==", fmt.Sprintf("%d", lo))
	}
	return g.and(g.compare(g.resume, ">=", fmt.Sprintf("%d", lo)), g.compare(g.resume, "<=", fmt.Sprintf("%d", hi)))
}

// condition is a loop condition, true when absent (`for (;;)`).
func (g *generatorResume) condition(expression *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	if expression == nil || expression.Kind == "empty" {
		return &frontend.SyntaxExpression{Span: g.span, Kind: "bool", Text: "true", InferredType: "boolean"}
	}
	return expression
}

func (g *generatorResume) ifStmt(condition *frontend.SyntaxExpression, then, otherwise []frontend.SyntaxStatement) frontend.SyntaxStatement {
	return frontend.SyntaxStatement{Span: g.span, Kind: "if", Expression: condition, Then: then, Else: otherwise}
}

func (g *generatorResume) declare(name, typ string, value *frontend.SyntaxExpression) frontend.SyntaxStatement {
	return frontend.SyntaxStatement{Span: g.span, Kind: "variable", VarDeclKind: "let", Name: name, Type: typ, InferredType: typ, Expression: value}
}

func (g *generatorResume) assign(name string, value *frontend.SyntaxExpression) frontend.SyntaxStatement {
	return frontend.SyntaxStatement{Span: g.span, Kind: "assign", Name: name, Expression: value}
}

// result returns { done, value } from the step closure; a nil value is the
// yielded type's undefined.
func (g *generatorResume) result(done bool, value *frontend.SyntaxExpression) frontend.SyntaxStatement {
	return frontend.SyntaxStatement{Span: g.span, Kind: "return", Expression: &frontend.SyntaxExpression{
		Span: g.span, Kind: "generator_result", Text: fmt.Sprintf("%t", done), InferredType: g.yieldType, Left: value,
	}}
}
