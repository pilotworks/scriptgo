package lowering

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// ValidateSubset rejects checked syntax that the current native IR cannot
// represent. Keeping this gate in lowering prevents backend-specific policy.
func ValidateSubset(program frontend.Program) error {
	lowerMu.Lock()
	defer lowerMu.Unlock()
	return validateSubsetLocked(program)
}

func validateSubsetLocked(program frontend.Program) error {
	var err error
	program, err = SpecializeGenerics(program)
	if err != nil {
		return err
	}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if err := validateStatement(file.FileName, statement, file.Builtin); err != nil {
				return err
			}
		}
	}
	return nil
}

func isObjectBuiltinCall(expr *typescriptgo.SyntaxExpression) bool {
	if expr == nil || expr.Kind != "call" {
		return false
	}
	name := ""
	if expr.Left != nil {
		if expr.Left.Kind == "identifier" && expr.Text != "" {
			name = expr.Left.Text + "." + expr.Text
		} else if (expr.Left.Kind == "property" || expr.Left.Kind == "optional_property") && expr.Left.Left != nil && expr.Left.Left.Kind == "identifier" {
			name = expr.Left.Left.Text + "." + expr.Left.Text
		}
	}
	if strings.HasPrefix(name, "Reflect.") {
		return true
	}
	return name == "Object.values" || name == "Object.entries" || name == "Object.assign" || name == "Object.keys"
}

func validateStatement(fileName string, statement typescriptgo.SyntaxStatement, isBuiltin bool) error {
	if statement.Kind != "type_alias" && (statement.Kind != "variable" || (!isHeterogeneousUnion(statement.Type) && !isObjectBuiltinCall(statement.Expression))) && statement.Kind != "generator_function" && statement.Kind != "async_generator_function" && !statement.IsGenerator {
		if err := validateStaticType(fileName, statement.Span, statement.Type, isBuiltin); err != nil {
			return err
		}
	}
	switch statement.Kind {
	case "break", "continue", "debugger":
		return nil
	case "dowhile":
		if err := validateExpression(fileName, statement.Expression, isBuiltin); err != nil {
			return err
		}
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, stepStatement := range statement.Step {
			if err := validateStatement(fileName, stepStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "forof", "forin", "forawaitof", "label":
		if statement.Expression != nil {
			if err := validateExpression(fileName, statement.Expression, isBuiltin); err != nil {
				return err
			}
		}
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, stepStatement := range statement.Step {
			if err := validateStatement(fileName, stepStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "switch":
		if err := validateExpression(fileName, statement.Expression, isBuiltin); err != nil {
			return err
		}
		for _, c := range statement.Cases {
			if c.Expression != nil {
				if err := validateExpression(fileName, c.Expression, isBuiltin); err != nil {
					return err
				}
			}
			for _, stmt := range c.Statements {
				if err := validateStatement(fileName, stmt, isBuiltin); err != nil {
					return err
				}
			}
		}
		return nil
	case "index_set":
		if err := validateExpression(fileName, statement.Left, isBuiltin); err != nil {
			return err
		}
		if err := validateExpression(fileName, statement.Right, isBuiltin); err != nil {
			return err
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "field_set":
		if err := validateExpression(fileName, statement.Left, isBuiltin); err != nil {
			return err
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "module", "enum", "interface", "type_alias":
		return nil
	case "class":
		if statement.Class != nil && len(statement.Class.TypeParameters) > 0 {
			return subsetError(fileName, statement.Span, CodeGenericSpecialize, fmt.Sprintf("unspecialized generic class %q", statement.Class.Name))
		}
		if statement.Class == nil || (statement.Class.Extends == "" && len(statement.Class.Fields) == 0 && statement.Class.Constructor == nil && len(statement.Class.Methods) == 0 && len(statement.Class.StaticBlocks) == 0 && len(statement.Class.Decorators) == 0) {
			return subsetError(fileName, statement.Span, CodeLanguageLowering, "empty class shape")
		}
		if statement.Class.Constructor != nil {
			for _, p := range statement.Class.Constructor.Parameters {
				if err := validateStaticType(fileName, p.Span, p.Type, isBuiltin); err != nil {
					return err
				}
			}
			for _, stmt := range statement.Class.Constructor.Body {
				if err := validateStatement(fileName, stmt, isBuiltin); err != nil {
					return err
				}
			}
		}
		for _, method := range statement.Class.Methods {
			if len(method.Body) == 0 {
				continue
			}
			for _, p := range method.Parameters {
				if err := validateStaticType(fileName, p.Span, p.Type, isBuiltin); err != nil {
					return err
				}
			}
			for _, stmt := range method.Body {
				if err := validateStatement(fileName, stmt, isBuiltin); err != nil {
					return err
				}
			}
		}
		for _, field := range statement.Class.Fields {
			if !isBuiltin && isUnknownType(field.Type) {
				return subsetError(fileName, field.Span, CodeUnknownBoundary, "class field of unknown type")
			}
			if err := validateStaticType(fileName, field.Span, field.Type, isBuiltin); err != nil {
				return err
			}
			if field.Initializer != nil {
				if err := validateExpression(fileName, field.Initializer, isBuiltin); err != nil {
					return err
				}
			}
		}
		for _, block := range statement.Class.StaticBlocks {
			for _, stmt := range block {
				if err := validateStatement(fileName, stmt, isBuiltin); err != nil {
					return err
				}
			}
		}
		return nil
	case "variable", "using", "await_using":
		if statement.Expression == nil {
			return subsetError(fileName, statement.Span, CodeLanguageLowering, "variable declaration without an initializer")
		}
		if statement.Expression.Kind == "object_literal" && len(statement.Expression.Arguments) == 0 && (statement.Type == "" || statement.Type == "{}") && (statement.InferredType == "" || statement.InferredType == "{}") && !strings.HasPrefix(statement.Name, "__destruct_") {
			return subsetError(fileName, statement.Span, CodeLanguageLowering, "unannotated empty object literal is not supported in native subset")
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "namespace":
		for _, s := range statement.Body {
			if s.Kind == "declare_function" || s.Kind == "interface" || s.Kind == "type_alias" || s.Kind == "module" || s.Kind == "enum" {
				continue
			}
			if err := validateStatement(fileName, s, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "expression", "return":
		if statement.Expression == nil {
			return nil
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "declare_function":
		for _, parameter := range statement.Parameters {
			if err := validateStaticType(fileName, parameter.Span, parameter.Type, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "function", "generator_function", "async_function", "async_generator_function":
		if len(statement.TypeParameters) > 0 {
			return subsetError(fileName, statement.Span, CodeGenericSpecialize, fmt.Sprintf("unspecialized generic function %q", statement.Name))
		}
		for _, parameter := range statement.Parameters {
			if !isBuiltin && isHeterogeneousUnion(parameter.Type) {
				return subsetError(fileName, parameter.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union type in parameter %q", parameter.Name))
			}
			if err := validateStaticType(fileName, parameter.Span, parameter.Type, isBuiltin); err != nil {
				return err
			}
		}
		if statement.Type != "" {
			if err := validateStaticType(fileName, statement.Span, statement.Type, isBuiltin); err != nil {
				return err
			}
		}
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "block":
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "assign":
		if statement.Expression == nil {
			return subsetError(fileName, statement.Span, CodeLanguageLowering, "assignment without an expression")
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "while":
		if err := validateExpression(fileName, statement.Expression, isBuiltin); err != nil {
			return err
		}
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, stepStatement := range statement.Step {
			if err := validateStatement(fileName, stepStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "if":
		if statement.Expression != nil && statement.Expression.Kind != "identifier" {
			if err := validateExpression(fileName, statement.Expression, isBuiltin); err != nil {
				return err
			}
		}
		for _, branchStatement := range statement.Then {
			if err := validateStatement(fileName, branchStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, branchStatement := range statement.Else {
			if err := validateStatement(fileName, branchStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "throw":
		if statement.Expression == nil {
			return subsetError(fileName, statement.Span, CodeLanguageLowering, "throw without an expression")
		}
		return validateExpression(fileName, statement.Expression, isBuiltin)
	case "try":
		if statement.CatchVarType != "" {
			if err := validateStaticType(fileName, statement.CatchVarSpan, statement.CatchVarType, isBuiltin); err != nil {
				return err
			}
		}
		for _, tryStatement := range statement.Body {
			if err := validateStatement(fileName, tryStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, catchStatement := range statement.Catch {
			if err := validateStatement(fileName, catchStatement, isBuiltin); err != nil {
				return err
			}
		}
		for _, finallyStatement := range statement.Finally {
			if err := validateStatement(fileName, finallyStatement, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "unsupported":
		return subsetError(fileName, statement.Span, CodeLanguageLowering, statement.Type)
	default:
		return subsetError(fileName, statement.Span, CodeInternalFallback, statement.Kind)
	}
}

func isHeterogeneousUnion(typ string) bool {
	normalized := strings.ToLower(strings.TrimSpace(typ))
	if strings.HasSuffix(normalized, "[]") {
		elem := strings.TrimSuffix(normalized, "[]")
		elem = strings.TrimPrefix(elem, "(")
		elem = strings.TrimSuffix(elem, ")")
		return isHeterogeneousUnion(elem)
	}
	if strings.Contains(normalized, "generator") || strings.Contains(normalized, "iterator") || strings.Contains(normalized, "string | symbol") || strings.Contains(normalized, "symbol | string") {
		return false
	}
	if !strings.Contains(normalized, "|") {
		return false
	}
	parts := strings.Split(normalized, "|")
	var nonNullish []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		trimmed = strings.TrimPrefix(trimmed, "(")
		trimmed = strings.TrimSuffix(trimmed, ")")
		if trimmed != "null" && trimmed != "undefined" && trimmed != "void" && trimmed != "" {
			nonNullish = append(nonNullish, trimmed)
		}
	}
	if len(nonNullish) <= 1 {
		return false
	}
	first := toPrimitiveCategory(nonNullish[0])
	for _, other := range nonNullish[1:] {
		if toPrimitiveCategory(other) != first {
			return true
		}
	}
	return false
}

func toPrimitiveCategory(t string) string {
	t = strings.TrimSpace(t)
	if (strings.HasPrefix(t, "\"") && strings.HasSuffix(t, "\"")) || (strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'")) || t == "string" {
		return "string"
	}
	if _, err := strconv.ParseFloat(t, 64); err == nil || t == "number" {
		return "number"
	}
	if t == "boolean" || t == "bool" || t == "true" || t == "false" {
		return "bool"
	}
	if t == "bigint" {
		return "bigint"
	}
	if t == "symbol" {
		return "symbol"
	}
	if strings.HasSuffix(t, "[]") {
		return "array"
	}
	return "object"
}

func isUnknownType(typ string) bool {
	normalized := strings.ToLower(strings.TrimSpace(typ))
	return normalized == "unknown" || normalized == "unknownkeyword" || normalized == "kindunknownkeyword"
}

func isOrContainsAny(typ string) bool {
	typ = strings.ToLower(strings.TrimSpace(typ))
	if typ == "" {
		return false
	}
	if typ == "any" || typ == "anykeyword" || typ == "kindanykeyword" {
		return true
	}
	n := len(typ)
	for i := 0; i < n; {
		b := typ[i]
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '$' {
			start := i
			for i < n {
				c := typ[i]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$' {
					i++
				} else {
					break
				}
			}
			word := typ[start:i]
			if word == "any" || word == "anykeyword" || word == "kindanykeyword" {
				return true
			}
		} else {
			i++
		}
	}
	return false
}

func validateStaticType(fileName string, span typescriptgo.SourceSpan, typ string, isBuiltin bool) error {
	normalized := strings.ToLower(strings.TrimSpace(typ))
	if !isBuiltin && strings.HasSuffix(normalized, "[]") && isUnknownType(strings.TrimSuffix(normalized, "[]")) {
		return subsetError(fileName, span, CodeUnknownBoundary, "unknown array type")
	}
	if isOrContainsAny(typ) {
		return subsetError(fileName, span, CodeAnyBoundary, "any type")
	}
	return nil
}

type Warning struct {
	FileName string
	Span     typescriptgo.SourceSpan
	Code     SubsetCode
	Message  string
}

func (w Warning) Format() string {
	return typescriptgo.Format(w.FileName, w.Span.Start, w.Span.Length, "warning", string(w.Code), w.Message, "")
}

var (
	warnings         []Warning
	WarnRuntimeCasts bool
)

func ClearDiagnostics() {
	warnings = nil
}

func GetWarnings() []Warning {
	lowerMu.Lock()
	defer lowerMu.Unlock()
	return append([]Warning(nil), warnings...)
}

func recordWarning(fileName string, span typescriptgo.SourceSpan, code SubsetCode, message string) {
	warnings = append(warnings, Warning{
		FileName: fileName,
		Span:     span,
		Code:     code,
		Message:  message,
	})
}

func validateExpression(fileName string, expression *typescriptgo.SyntaxExpression, isBuiltin bool) error {
	switch expression.Kind {
	case "as":
		if err := validateStaticType(fileName, expression.Span, expression.Text, isBuiltin); err != nil {
			return err
		}
		if expression.Left != nil {
			if expression.Left.Kind == "as" && isUnknownType(expression.Left.Text) && !isUnknownType(expression.Text) {
				innerVal := "value"
				if expression.Left.Left != nil && expression.Left.Left.Text != "" {
					innerVal = expression.Left.Left.Text
				}
				recordWarning(fileName, expression.Span, CodeUnsafeDoubleCast, fmt.Sprintf("unsafe escape-hatch double assertion (%s as unknown as %s)", innerVal, expression.Text))
			} else if WarnRuntimeCasts && (isUnknownType(expression.Left.InferredType) || (expression.Left.Kind == "as" && isUnknownType(expression.Left.Text))) && !isUnknownType(expression.Text) {
				recordWarning(fileName, expression.Span, CodeWarnCheckedCast, fmt.Sprintf("runtime checked cast to %s", expression.Text))
			}
			if expression.Left.Kind != "identifier" {
				return validateExpression(fileName, expression.Left, isBuiltin)
			}
			return nil
		}
		return nil
	case "identifier":
		return nil
	case "number", "bigint", "regex", "string", "bool", "null", "undefined":
		return nil
	case "arrow_function":
		if expression.Function != nil {
			for _, p := range expression.Function.Parameters {
				if !isBuiltin && isHeterogeneousUnion(p.Type) {
					return subsetError(fileName, p.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union type in parameter %q", p.Name))
				}
				if err := validateStaticType(fileName, p.Span, p.Type, isBuiltin); err != nil {
					return err
				}
			}
			if expression.Function.Type != "" {
				if err := validateStaticType(fileName, expression.Span, expression.Function.Type, isBuiltin); err != nil {
					return err
				}
			}
			for _, stmt := range expression.Function.Body {
				if err := validateStatement(fileName, stmt, isBuiltin); err != nil {
					return err
				}
			}
		}
		return nil
	case "array":
		for _, element := range expression.Arguments {
			if err := validateExpression(fileName, element, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "object_literal":
		for _, prop := range expression.Arguments {
			if prop.Left != nil {
				if err := validateExpression(fileName, prop.Left, isBuiltin); err != nil {
					return err
				}
			}
		}
		return nil
	case "spread":
		return validateExpression(fileName, expression.Left, isBuiltin)
	case "optional_index", "index":
		if expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Left.InferredType))
		}
		if err := validateExpression(fileName, expression.Left, isBuiltin); err != nil {
			return err
		}
		return validateExpression(fileName, expression.Right, isBuiltin)
	case "optional_property", "property":
		if expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Left.InferredType))
		}
		if expression.Left != nil && (expression.Left.Kind == "identifier" || expression.Left.Kind == "string" || expression.Left.Kind == "call" || expression.Left.Kind == "optional_call" || expression.Left.Kind == "property" || expression.Left.Kind == "optional_property" || expression.Left.Kind == "index" || expression.Left.Kind == "optional_index" || expression.Left.Kind == "object_literal" || expression.Left.Kind == "as") {
			return validateExpression(fileName, expression.Left, isBuiltin)
		}
		return subsetError(fileName, expression.Span, CodeStructuralFlow, "nested property access")
	case "new":
		for _, argument := range expression.Arguments {
			if err := validateExpression(fileName, argument, isBuiltin); err != nil {
				return err
			}
		}
		if callName(expression.Left) == "" {
			return subsetError(fileName, expression.Span, CodeLanguageLowering, "dynamic constructor target")
		}
		return nil
	case "typeof":
		if expression.Left != nil && expression.Left.Kind != "identifier" {
			return validateExpression(fileName, expression.Left, isBuiltin)
		}
		return nil
	case "await", "yield", "yield_star":
		if expression.Left != nil {
			return validateExpression(fileName, expression.Left, isBuiltin)
		}
		return nil
	case "unary":
		if expression.Operator != "!" && expression.Operator != "-" && expression.Operator != "+" && expression.Operator != "~" && expression.Operator != "++" && expression.Operator != "--" && expression.Operator != "void" {
			return subsetError(fileName, expression.Span, CodeLanguageLowering, "unary operator "+expression.Operator)
		}
		if expression.Operator != "!" && expression.Operator != "void" && expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Left.InferredType))
		}
		if expression.Operator == "!" && expression.Left != nil && expression.Left.Kind == "identifier" {
			return nil
		}
		return validateExpression(fileName, expression.Left, isBuiltin)
	case "postfix_unary":
		if expression.Operator != "++" && expression.Operator != "--" {
			return subsetError(fileName, expression.Span, CodeLanguageLowering, "postfix operator "+expression.Operator)
		}
		if expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Left.InferredType))
		}
		return validateExpression(fileName, expression.Left, isBuiltin)
	case "binary":
		if expression.Operator == "instanceof" || expression.Operator == "in" || expression.Operator == "==" || expression.Operator == "!=" || expression.Operator == "===" || expression.Operator == "!==" || expression.Operator == "&&" || expression.Operator == "||" || expression.Operator == "??" {
			if expression.Left != nil && expression.Left.Kind != "identifier" {
				if err := validateExpression(fileName, expression.Left, isBuiltin); err != nil {
					return err
				}
			}
			if expression.Right != nil && expression.Right.Kind != "identifier" {
				if err := validateExpression(fileName, expression.Right, isBuiltin); err != nil {
					return err
				}
			}
			return nil
		}
		if expression.Left != nil && isHeterogeneousUnion(expression.Left.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Left.InferredType))
		}
		if expression.Right != nil && isHeterogeneousUnion(expression.Right.InferredType) {
			return subsetError(fileName, expression.Span, CodeUnionNarrowing, fmt.Sprintf("unresolved union operation on type %q", expression.Right.InferredType))
		}
		if err := validateExpression(fileName, expression.Left, isBuiltin); err != nil {
			return err
		}
		return validateExpression(fileName, expression.Right, isBuiltin)
	case "template", "tagged_template":
		if expression.Left != nil {
			if err := validateExpression(fileName, expression.Left, isBuiltin); err != nil {
				return err
			}
		}
		for _, argument := range expression.Arguments {
			if err := validateExpression(fileName, argument, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "conditional":
		if expression.Left != nil && expression.Left.Kind != "identifier" {
			if err := validateExpression(fileName, expression.Left, isBuiltin); err != nil {
				return err
			}
		}
		if expression.WhenTrue != nil {
			if err := validateExpression(fileName, expression.WhenTrue, isBuiltin); err != nil {
				return err
			}
		}
		if expression.WhenFalse != nil {
			if err := validateExpression(fileName, expression.WhenFalse, isBuiltin); err != nil {
				return err
			}
		}
		return nil
	case "call", "optional_call":
		for _, argument := range expression.Arguments {
			if err := validateExpression(fileName, argument, isBuiltin); err != nil {
				return err
			}
		}
		target := expression.Left
		for target != nil && (target.Kind == "as" || target.Kind == "paren" || target.Kind == "cast" || target.Kind == "type_assertion" || target.Kind == "non_null") {
			target = target.Left
		}
		if target != nil && target.Kind == "arrow_function" {
			return validateExpression(fileName, target, isBuiltin)
		}
		if callName(expression.Left) == "" && stringMethod(expression.Left) == "" && arrayMethod(expression.Left) == "" && (target == nil || (target.Kind != "property" && target.Kind != "optional_property" && target.Kind != "index" && target.Kind != "optional_index" && target.Kind != "identifier")) {
			return subsetError(fileName, expression.Span, CodeFunctionValue, "dynamic call target")
		}
		return nil
	case "unsupported":
		return subsetError(fileName, expression.Span, CodeLanguageLowering, expression.Text)
	default:
		return subsetError(fileName, expression.Span, CodeInternalFallback, expression.Kind)
	}
}

func formatSubsetMessage(code SubsetCode, feature string) string {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return "unsupported feature in native subset"
	}
	if strings.Contains(feature, "in native subset") || strings.Contains(feature, "native subset") {
		return feature
	}
	if strings.HasSuffix(feature, ".") {
		return strings.TrimSuffix(feature, ".") + " in native subset."
	}
	if strings.ToLower(feature) == "any type" {
		return "The any type is not supported in native subset."
	}
	r, size := utf8.DecodeRuneInString(feature)
	capitalized := string(unicode.ToUpper(r)) + feature[size:]
	if strings.Contains(strings.ToLower(feature), "not supported") {
		return capitalized + " in native subset."
	}
	return capitalized + " is not supported in native subset."
}

func subsetError(fileName string, span typescriptgo.SourceSpan, code SubsetCode, feature string) error {
	msg := formatSubsetMessage(code, feature)
	return fmt.Errorf("%s", typescriptgo.Format(fileName, span.Start, span.Length, "error", string(code), msg, ""))
}
