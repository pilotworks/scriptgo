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

func TestVerifyAllowsBufferInUint8ArrayArray(t *testing.T) {
	module := Module{
		Functions: []Function{
			{
				Name:       "main",
				ReturnType: TypeVoid,
				Body: []Instruction{
					{Op: OpCall, Type: TypeBuffer, Result: "buffer", Callee: "__buffer.alloc"},
					{Op: OpArray, Type: TypeUint8Array + "[]", Result: "parts", Args: []string{"buffer"}},
					{Op: OpReturn, Type: TypeVoid},
				},
			},
		},
	}
	if err := module.Verify(); err != nil {
		t.Fatalf("expected Buffer to satisfy Uint8Array array element type, got: %v", err)
	}
}

func TestVerifyAsyncFunctionRequiresStateMachine(t *testing.T) {
	f := Function{Name: "asyncFn", Async: true, ReturnType: Type("object:Promise")}
	if err := f.Verify(); err == nil {
		t.Fatal("expected async state-machine metadata error")
	}
}

func TestVerifyRejectsBlockingAwaitInAsyncFunction(t *testing.T) {
	f := Function{
		Name:       "asyncFn",
		Async:      true,
		EntryBlock: "entry",
		AsyncFrame: &AsyncFrame{Name: "asyncFn_frame"},
		Blocks: []BasicBlock{{
			Name:       "entry",
			Terminator: Terminator{Kind: TermReturn},
		}},
		Body: []Instruction{{Op: OpCall, Type: TypeNumber, Result: "value", Callee: "__async.await", Args: []string{"promise"}}},
	}
	if err := f.Verify(); err == nil {
		t.Fatal("expected blocking await rejection")
	}
}

func TestVerifyAsyncAwaitTargets(t *testing.T) {
	f := Function{
		Name:       "asyncFn",
		Async:      true,
		EntryBlock: "entry",
		AsyncFrame: &AsyncFrame{Name: "asyncFn_frame"},
		Blocks: []BasicBlock{
			{Name: "entry", Terminator: Terminator{Kind: TermAwait, AwaitValue: "promise", Fulfilled: "fulfilled", Rejected: "rejected", State: 1}},
			{Name: "fulfilled", Terminator: Terminator{Kind: TermReturn}},
			{Name: "rejected", Terminator: Terminator{Kind: TermThrow}},
		},
	}
	if err := f.Verify(); err != nil {
		t.Fatalf("expected valid async CFG, got %v", err)
	}
}

func TestVerifyRejectsUnreachableAsyncBlock(t *testing.T) {
	f := Function{
		Name:       "asyncFn",
		Async:      true,
		EntryBlock: "entry",
		AsyncFrame: &AsyncFrame{Name: "asyncFn_frame"},
		Blocks: []BasicBlock{
			{Name: "entry", Terminator: Terminator{Kind: TermReturn}},
			{Name: "stale", Terminator: Terminator{Kind: TermReturn}},
		},
	}
	if err := f.Verify(); err == nil {
		t.Fatal("expected unreachable async block rejection")
	}
}

func TestVerifyRejectsDuplicateAsyncAwaitState(t *testing.T) {
	f := Function{
		Name:       "asyncFn",
		Async:      true,
		EntryBlock: "entry",
		AsyncFrame: &AsyncFrame{Name: "asyncFn_frame"},
		Blocks: []BasicBlock{
			{Name: "entry", Terminator: Terminator{Kind: TermAwait, AwaitValue: "p", Fulfilled: "ok", Rejected: "bad", State: 0}},
			{Name: "ok", Terminator: Terminator{Kind: TermAwait, AwaitValue: "q", Fulfilled: "done", Rejected: "bad2", State: 0}},
			{Name: "bad", Terminator: Terminator{Kind: TermThrow}},
			{Name: "bad2", Terminator: Terminator{Kind: TermThrow}},
			{Name: "done", Terminator: Terminator{Kind: TermReturn}},
		},
	}
	if err := f.Verify(); err == nil {
		t.Fatal("expected duplicate async await state rejection")
	}
}
