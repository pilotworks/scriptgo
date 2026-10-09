package lowering

import (
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func TestRecordParameterDefault(t *testing.T) {
	defaultParamsIndex = map[string]map[int]*frontend.SyntaxExpression{}
	initializer := &frontend.SyntaxExpression{Kind: "number", Text: "1"}
	recordParameterDefault("f", 0, frontend.SyntaxParameter{Name: "required"})
	recordParameterDefault("f", 1, frontend.SyntaxParameter{Name: "withDefault", Initializer: initializer})
	recordParameterDefault("f", 2, frontend.SyntaxParameter{Name: "optional", Optional: true})
	defaults := defaultParamsIndex["f"]
	if _, ok := defaults[0]; ok {
		t.Fatal("required parameter must not record a default")
	}
	if defaults[1] != initializer {
		t.Fatalf("initializer default = %+v", defaults[1])
	}
	if defaults[2] == nil || defaults[2].Kind != "undefined" {
		t.Fatalf("optional default = %+v", defaults[2])
	}
	recordParameterDefault("g", 0, frontend.SyntaxParameter{Name: "required"})
	if _, ok := defaultParamsIndex["g"]; ok {
		t.Fatal("a function without defaults must not get an entry")
	}
}
