package ir

import "testing"

func TestVerifyValidModule(t *testing.T) {
	module := Module{
		Functions: []Function{
			{
				Name:       "main",
				ReturnType: TypeVoid,
				Body: []Instruction{
					{Op: OpConst, Type: TypeNumber, Result: "t0", Value: "42"},
					{Op: OpPrint, Args: []string{"t0"}},
					{Op: OpReturn, Type: TypeVoid},
				},
			},
		},
	}
	if err := module.Verify(); err != nil {
		t.Fatalf("expected valid module, got error: %v", err)
	}
}

func TestVerifyEmptyModule(t *testing.T) {
	module := Module{}
	if err := module.Verify(); err == nil {
		t.Fatal("expected error for empty module, got nil")
	}
}

func TestVerifyInvalidReturnType(t *testing.T) {
	module := Module{
		Functions: []Function{
			{
				Name:       "main",
				ReturnType: TypeNumber,
				Body: []Instruction{
					{Op: OpReturn, Type: TypeVoid},
				},
			},
		},
	}
	if err := module.Verify(); err == nil {
		t.Fatal("expected error for return type mismatch, got nil")
	}
}

func TestVerifyUnknownArg(t *testing.T) {
	module := Module{
		Functions: []Function{
			{
				Name:       "main",
				ReturnType: TypeVoid,
				Body: []Instruction{
					{Op: OpBinary, Type: TypeNumber, Result: "t1", Operator: "+", Args: []string{"t0", "t0"}},
					{Op: OpReturn, Type: TypeVoid},
				},
			},
		},
	}
	if err := module.Verify(); err == nil {
		t.Fatal("expected error for unknown arg, got nil")
	}
}
