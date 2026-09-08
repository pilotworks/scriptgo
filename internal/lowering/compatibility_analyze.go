package lowering

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

type compatibilitySiteKey struct {
	path   string
	start  int
	length int
	kind   string
}

type compatibilityCollector struct {
	mode            CompatibilityMode
	root            string
	decisions       map[compatibilitySiteKey]CompatibilityDecision
	sources         map[string]string
	dynamicBindings map[string]bool
}

func AnalyzeCompatibility(program frontend.Program, policy CompatibilityPolicy) (CompatibilityReport, error) {
	lowerMu.Lock()
	defer lowerMu.Unlock()
	var err error
	program, err = SpecializeGenerics(cloneCompatibilityProgram(program))
	if err != nil {
		return CompatibilityReport{}, err
	}
	mode := policy.Mode
	if mode == "" {
		mode = ModeStatic
	}
	if mode != ModeStatic && mode != ModeDynamicEnabled {
		return CompatibilityReport{}, fmt.Errorf("unknown compatibility mode %q", mode)
	}
	root := filepath.Dir(program.EntryPath)
	collector := compatibilityCollector{mode: mode, root: root, decisions: make(map[compatibilitySiteKey]CompatibilityDecision), sources: make(map[string]string), dynamicBindings: make(map[string]bool)}
	for _, file := range program.Files {
		if file.Builtin {
			continue
		}
		path := collector.normalizePath(file.FileName)
		collector.sources[path] = file.Source
		for _, reference := range file.Imports {
			if reference.TypeOnly || !isJavaScriptFile(reference.ResolvedFileName) {
				continue
			}
			for _, binding := range reference.Bindings {
				if !binding.TypeOnly {
					collector.dynamicBindings[binding.LocalName] = true
				}
			}
		}
		for i := range file.Syntax.Statements {
			collector.statement(path, &file.Syntax.Statements[i])
		}
	}
	report := CompatibilityReport{Format: CompatibilityReportFormat, Mode: mode}
	for _, decision := range collector.decisions {
		report.Decisions = append(report.Decisions, decision)
	}
	sort.Slice(report.Decisions, func(i, j int) bool {
		a, b := report.Decisions[i], report.Decisions[j]
		if a.FileName != b.FileName {
			return a.FileName < b.FileName
		}
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		if a.Length != b.Length {
			return a.Length < b.Length
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Code < b.Code
	})
	for _, decision := range report.Decisions {
		switch decision.Tier {
		case TierStatic:
			report.Summary.Static++
		case TierDynamic:
			report.Summary.Dynamic++
		case TierUnsupported:
			report.Summary.Unsupported++
		}
	}
	return report, nil
}

func (c *compatibilityCollector) normalizePath(path string) string {
	rel, err := filepath.Rel(c.root, path)
	if err != nil || filepath.IsAbs(rel) && rel == path {
		return filepath.ToSlash(filepath.Base(path))
	}
	return filepath.ToSlash(filepath.Clean(rel))
}

func (c *compatibilityCollector) add(path string, span typescriptgo.SourceSpan, kind string, code SubsetCode, feature, hint string, dynamicEligible bool) {
	decision := CompatibilityDecision{Tier: TierStatic, FileName: path, Span: span, Start: span.Start, Length: span.Length, Kind: kind, DynamicCapability: feature, Source: c.sources[path]}
	if code != "" {
		decision.Tier = TierUnsupported
		decision.Code = code
		decision.Message = formatSubsetMessage(code, feature)
		decision.Hint = hint
		if dynamicEligible && c.mode == ModeDynamicEnabled {
			decision.Tier = TierDynamic
		}
	}
	key := compatibilitySiteKey{path: path, start: span.Start, length: span.Length, kind: kind}
	if old, exists := c.decisions[key]; !exists || compatibilityRank(decision.Tier) > compatibilityRank(old.Tier) {
		c.decisions[key] = decision
	}
}

func compatibilityRank(tier CompatibilityTier) int {
	switch tier {
	case TierUnsupported:
		return 2
	case TierDynamic:
		return 1
	default:
		return 0
	}
}

func (c *compatibilityCollector) typed(path string, span typescriptgo.SourceSpan, kind, typ string) bool {
	if !isOrContainsAny(typ) {
		return false
	}
	c.add(path, span, kind, CodeAnyBoundary, "any type", "use a concrete type, narrow the value, or enable --dynamic", true)
	return true
}

func (c *compatibilityCollector) statement(path string, statement *typescriptgo.SyntaxStatement) {
	if statement == nil {
		return
	}
	classified := false
	if statement.Kind != "type_alias" && (statement.Kind != "variable" || (!isHeterogeneousUnion(statement.Type) && !isObjectBuiltinCall(statement.Expression))) && statement.Kind != "generator_function" && statement.Kind != "async_generator_function" && !statement.IsGenerator {
		classified = c.typed(path, statement.Span, statement.Kind, statement.Type)
	}
	if !classified {
		switch statement.Kind {
		case "unsupported":
			c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, statement.Type, "", false)
			classified = true
		case "class":
			if statement.Class == nil {
				c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, "class declaration is missing its class body", "", false)
				classified = true
			}
		case "variable":
			if statement.Expression != nil && statement.Expression.Kind == "object_literal" && len(statement.Expression.Arguments) == 0 && (statement.Type == "" || statement.Type == "{}") && (statement.InferredType == "" || statement.InferredType == "{}") && !strings.HasPrefix(statement.Name, "__destruct_") {
				c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, "unannotated empty object literal is not supported in native subset", "", false)
				classified = true
			}
		case "using", "await_using":
			if statement.Expression == nil {
				c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, "resource declaration without an initializer", "", false)
				classified = true
			}
		case "assign":
			if statement.Expression == nil {
				c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, "assignment without an expression", "", false)
				classified = true
			}
		case "throw":
			if statement.Expression == nil {
				c.add(path, statement.Span, statement.Kind, CodeLanguageLowering, "throw without an expression", "", false)
				classified = true
			}
		default:
			if !knownStatementKind(statement.Kind) {
				c.add(path, statement.Span, statement.Kind, CodeInternalFallback, statement.Kind, "", false)
				classified = true
			}
		}
	}
	if !classified {
		c.add(path, statement.Span, statement.Kind, "", "", "", false)
	}
	for i := range statement.Parameters {
		c.parameter(path, &statement.Parameters[i])
	}
	if statement.CatchVarType != "" {
		c.typed(path, statement.CatchVarSpan, "catch_parameter", statement.CatchVarType)
	}
	c.expression(path, statement.Expression)
	c.expression(path, statement.Left)
	c.expression(path, statement.Right)
	for i := range statement.Body {
		c.statement(path, &statement.Body[i])
	}
	for i := range statement.Step {
		c.statement(path, &statement.Step[i])
	}
	for i := range statement.Then {
		c.statement(path, &statement.Then[i])
	}
	for i := range statement.Else {
		c.statement(path, &statement.Else[i])
	}
	for i := range statement.Catch {
		c.statement(path, &statement.Catch[i])
	}
	for i := range statement.Finally {
		c.statement(path, &statement.Finally[i])
	}
	for i := range statement.Cases {
		c.expression(path, statement.Cases[i].Expression)
		for j := range statement.Cases[i].Statements {
			c.statement(path, &statement.Cases[i].Statements[j])
		}
	}
	if statement.Class != nil {
		c.class(path, statement.Class)
	}
}

func (c *compatibilityCollector) parameter(path string, parameter *typescriptgo.SyntaxParameter) {
	if parameter == nil {
		return
	}
	if !c.typed(path, parameter.Span, "parameter", parameter.Type) {
		c.add(path, parameter.Span, "parameter", "", "", "", false)
	}
	c.expression(path, parameter.Initializer)
}

func (c *compatibilityCollector) class(path string, class *typescriptgo.SyntaxClass) {
	if class.Constructor != nil {
		for i := range class.Constructor.Parameters {
			c.parameter(path, &class.Constructor.Parameters[i])
		}
		for i := range class.Constructor.Body {
			c.statement(path, &class.Constructor.Body[i])
		}
	}
	for i := range class.Methods {
		method := &class.Methods[i]
		if !c.typed(path, method.Span, "method", method.Type) {
			c.add(path, method.Span, "method", "", "", "", false)
		}
		for j := range method.Parameters {
			c.parameter(path, &method.Parameters[j])
		}
		for j := range method.Body {
			c.statement(path, &method.Body[j])
		}
	}
	for i := range class.Fields {
		field := &class.Fields[i]
		if !c.typed(path, field.Span, "field", field.Type) {
			c.add(path, field.Span, "field", "", "", "", false)
		}
		c.expression(path, field.Initializer)
	}
	for i := range class.StaticBlocks {
		for j := range class.StaticBlocks[i] {
			c.statement(path, &class.StaticBlocks[i][j])
		}
	}
}

func (c *compatibilityCollector) expression(path string, expression *typescriptgo.SyntaxExpression) {
	if expression == nil {
		return
	}
	classified := false
	skipLeft := false
	if expression.Kind == "as" {
		classified = c.typed(path, expression.Span, expression.Kind, expression.Text)
	}
	if !classified {
		switch expression.Kind {
		case "optional_index", "index", "optional_property", "property", "unary", "postfix_unary", "binary":
			var unionType string
			if expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
				unionType = expression.Left.InferredType
			}
			if unionType == "" && expression.Right != nil && isHeterogeneousUnion(expression.Right.InferredType) {
				unionType = expression.Right.InferredType
			}
			if unionType != "" && !(expression.Kind == "binary" && expression.Operator == "+" && hasStringConcatUnion(expression)) {
				c.add(path, expression.Span, expression.Kind, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", unionType), "narrow the union before this operation", false)
				classified = true
			}
		}
	}
	if !classified && (expression.Kind == "property" || expression.Kind == "optional_property") && expression.Left != nil {
		allowed := map[string]bool{"identifier": true, "string": true, "call": true, "optional_call": true, "property": true, "optional_property": true, "index": true, "optional_index": true, "object_literal": true, "as": true, "non_null": true}
		if !allowed[expression.Left.Kind] {
			c.add(path, expression.Span, expression.Kind, CodeStructuralFlow, "nested property access", "enable --dynamic for JavaScript property semantics", true)
			classified = true
		}
	}
	if !classified && (expression.Kind == "call" || expression.Kind == "optional_call") {
		if expression.Left != nil && expression.Left.Kind == "identifier" && c.dynamicBindings[expression.Left.Text] {
			c.add(path, expression.Span, expression.Kind, CodeFunctionValue, "local JavaScript module call", "enable --dynamic for JavaScript execution", true)
			classified = true
		}
	}
	if !classified && (expression.Kind == "call" || expression.Kind == "optional_call") {
		name := callName(expression.Left)
		if name == "eval" || name == "Function" {
			c.add(path, expression.Span, expression.Kind, CodeFunctionValue, "dynamic call target", "enable --dynamic for JavaScript call semantics", true)
			classified = true
		} else {
			target := expression.Left
			for target != nil && (target.Kind == "as" || target.Kind == "paren" || target.Kind == "cast" || target.Kind == "type_assertion" || target.Kind == "non_null") {
				target = target.Left
			}
			if callName(expression.Left) == "" && stringMethod(expression.Left) == "" && arrayMethod(expression.Left) == "" && (target == nil || (target.Kind != "property" && target.Kind != "optional_property" && target.Kind != "index" && target.Kind != "optional_index" && target.Kind != "identifier" && target.Kind != "call" && target.Kind != "optional_call" && target.Kind != "arrow_function")) {
				c.add(path, expression.Span, expression.Kind, CodeFunctionValue, "dynamic call target", "enable --dynamic for JavaScript call semantics", true)
				classified = true
				skipLeft = target != nil && target.Kind == "unsupported"
			}
		}
	}
	if !classified {
		switch expression.Kind {
		case "unsupported":
			c.add(path, expression.Span, expression.Kind, CodeLanguageLowering, expression.Text, "", false)
			classified = true
		case "new":
			if callName(expression.Left) == "Function" {
				c.add(path, expression.Span, expression.Kind, CodeFunctionValue, "dynamic constructor target", "enable --dynamic for new Function", true)
				classified = true
			} else if callName(expression.Left) == "" {
				c.add(path, expression.Span, expression.Kind, CodeLanguageLowering, "dynamic constructor target", "", false)
				classified = true
			}
		case "unary":
			if expression.Operator != "!" && expression.Operator != "-" && expression.Operator != "+" && expression.Operator != "~" && expression.Operator != "++" && expression.Operator != "--" && expression.Operator != "void" {
				c.add(path, expression.Span, expression.Kind, CodeLanguageLowering, "unary operator "+expression.Operator, "", false)
				classified = true
			}
		case "postfix_unary":
			if expression.Operator != "++" && expression.Operator != "--" {
				c.add(path, expression.Span, expression.Kind, CodeLanguageLowering, "postfix operator "+expression.Operator, "", false)
				classified = true
			}
		default:
			if !knownExpressionKind(expression.Kind) {
				c.add(path, expression.Span, expression.Kind, CodeInternalFallback, expression.Kind, "", false)
				classified = true
			}
		}
	}
	if !classified {
		c.add(path, expression.Span, expression.Kind, "", "", "", false)
	}
	if !skipLeft {
		c.expression(path, expression.Left)
	}
	c.expression(path, expression.Right)
	c.expression(path, expression.WhenTrue)
	c.expression(path, expression.WhenFalse)
	for _, argument := range expression.Arguments {
		c.expression(path, argument)
	}
	if expression.Function != nil {
		c.statement(path, expression.Function)
	}
}

func knownStatementKind(kind string) bool {
	switch kind {
	case "break", "continue", "debugger", "empty", "dowhile", "forof", "forin", "forawaitof", "label", "switch", "index_set", "field_set", "module", "enum", "interface", "type_alias", "import_alias", "export_alias", "class", "variable", "using", "await_using", "namespace", "expression", "return", "declare_function", "function", "generator_function", "async_function", "async_generator_function", "block", "assign", "while", "if", "throw", "try":
		return true
	}
	return false
}

func knownExpressionKind(kind string) bool {
	switch kind {
	case "as", "non_null", "identifier", "number", "bigint", "regex", "string", "bool", "null", "undefined", "arrow_function", "array", "object_literal", "spread", "optional_index", "index", "optional_property", "property", "new", "typeof", "await", "yield", "yield_star", "unary", "postfix_unary", "binary", "template", "tagged_template", "conditional", "call", "optional_call":
		return true
	}
	return false
}
