package llvm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestEmitArithmeticAndPrint(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "left", Value: "20"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "right", Value: "22"},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "sum", Operator: "+", Args: []string{"left", "right"}},
			{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"sum"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"define i32 @main(", "fadd double", "@scriptgo_console_log_number", "ret i32 0"} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestEmitNumberArrayAndIndex(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "first", Value: "10"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "second", Value: "20"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "index", Value: "1"},
			{Op: ir.OpArray, Type: ir.TypeNumberArray, Result: "values", Args: []string{"first", "second"}},
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "value", Args: []string{"values", "index"}},
			{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"value"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"declare i32 @scriptgo_array_new(i64, i64, ptr)",
		"call i32 @scriptgo_array_set",
		"call i32 @scriptgo_array_get",
		"declare i32 @scriptgo_array_release(ptr)",
		"load double",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestEmitTypedIndexFromUnknownArrayUsesTaggedRead(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeBool, Result: "flag", Value: "true"},
			{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: "boxed", Args: []string{"flag"}},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "index", Value: "0"},
			{Op: ir.OpArray, Type: ir.TypeUnknownArray, Result: "values", Args: []string{"boxed"}},
			{Op: ir.OpIndex, Type: ir.TypeBool, Result: "value", Args: []string{"values", "index"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"call i32 @scriptgo_array_get_unknown",
		"extractvalue { i32, i32, i64, i64 } %value.unknown, 0",
		"extractvalue { i32, i32, i64, i64 } %value.unknown, 2",
		"trunc i64",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "call i32 @scriptgo_array_get(ptr %values") {
		t.Errorf("typed index from unknown[] used an untagged array read:\n%s", output)
	}
}

func TestEmitIndexedAssignmentUsesTypedArrayWrite(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeString, Result: "value", Value: "updated"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "index", Value: "0"},
			{Op: ir.OpArray, Type: ir.TypeStringArray, Result: "values", Args: []string{"value"}},
			{Op: ir.OpIndexSet, Type: ir.TypeVoid, Args: []string{"values", "index", "value"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"declare i32 @scriptgo_array_set_typed(ptr, double, ptr, i64, i64)",
		"call i32 @scriptgo_array_set_typed(ptr %values",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestEmitArrayFromAsyncPreservesArrayPromiseTag(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
			{Op: ir.OpArray, Type: ir.TypeNumberArray, Result: "values", Args: []string{"one"}},
			{Op: ir.OpCall, Type: ir.Type("object:Promise<number[]>"), Result: "promise", Callee: "__async.array_from_async", Args: []string{"values"}},
			{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: "outer", Callee: "__async.promise_create"},
			{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{"outer", "values"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "call i32 @scriptgo_promise_resolve_existing_array") {
		t.Fatalf("Array.fromAsync did not preserve the array payload tag:\n%s", output)
	}
	if strings.Contains(output, "call i32 @scriptgo_promise_resolve(ptr %promise") {
		t.Fatalf("Array.fromAsync used generic pointer resolution:\n%s", output)
	}
}

func TestEmitAsyncFramePersistsEntryParameter(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{
		{Name: "main", ReturnType: ir.TypeVoid, Body: []ir.Instruction{{Op: ir.OpReturn, Type: ir.TypeVoid}}},
		{
			Name:       "asyncFn",
			Parameters: []ir.Parameter{{Name: "value", Type: ir.TypeNumber}},
			ReturnType: ir.Type("object:Promise"),
			Async:      true,
			EntryBlock: "entry",
			AsyncFrame: &ir.AsyncFrame{Fields: []ir.AsyncField{
				{Name: "state", Type: ir.TypeNumber},
				{Name: "promise", Type: ir.Type("object:Promise")},
				{Name: "value", Type: ir.TypeNumber},
			}},
			Blocks: []ir.BasicBlock{{Name: "entry", Terminator: ir.Terminator{Kind: ir.TermReturn}}},
			Body: []ir.Instruction{
				{Op: ir.OpCall, Type: ir.TypePointer, Result: "frame", Callee: "__async.frame_new", Value: "3"},
				{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{"frame"}},
			},
		}}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "call i32 @scriptgo_async_frame_set") {
		t.Fatalf("async entry did not persist frame values:\n%s", output)
	}
}

func TestEmitClosureRecordsCanonicalReturnABI(t *testing.T) {
	rawParameters := []ir.Parameter{{Name: "__env_ctx", Type: ir.TypePointer}}
	for index := 0; index < 4; index++ {
		rawParameters = append(rawParameters, ir.Parameter{Name: fmt.Sprintf("arg_%d$raw", index), Type: ir.TypeUnknown})
	}
	module := ir.Module{Functions: []ir.Function{
		{
			Name:       "main",
			ReturnType: ir.TypeVoid,
			Body: []ir.Instruction{
				{Op: ir.OpClosure, Type: ir.TypeClosure, Result: "callback", Callee: "__closure_callback"},
				{Op: ir.OpReturn, Type: ir.TypeVoid},
			},
		},
		{
			Name:       "__closure_callback",
			Parameters: rawParameters,
			ReturnType: ir.TypeUnknown,
			Body:       []ir.Instruction{{Op: ir.OpReturn, Type: ir.TypeUnknown}},
		},
	}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "call i32 @scriptgo_closure_create(ptr @__closure_callback, ptr null, ptr @__closure_callback$invoke, i32 -1") {
		t.Fatalf("closure did not record its canonical aggregate return ABI:\n%s", output)
	}
}

func TestEmitUnknownClosureCallUsesReturnTagDispatcher(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{
		{
			Name:       "main",
			ReturnType: ir.TypeVoid,
			Body: []ir.Instruction{
				{Op: ir.OpClosure, Type: ir.TypeClosure, Result: "callback", Callee: "__closure_callback"},
				{Op: ir.OpClosureCall, Type: ir.TypeUnknown, Result: "result", Callee: "callback"},
				{Op: ir.OpReturn, Type: ir.TypeVoid},
			},
		},
		{
			Name:       "__closure_callback",
			ReturnType: ir.TypeVoid,
			Body:       []ir.Instruction{{Op: ir.OpReturn, Type: ir.TypeVoid}},
		},
	}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "call i32 @scriptgo_closure_create(ptr @__closure_callback, ptr null, ptr @__closure_callback$invoke, i32 0") {
		t.Fatalf("void closure did not record its return ABI:\n%s", output)
	}
	if !strings.Contains(output, "call i32 @scriptgo_closure_invoke_value(") {
		t.Fatalf("unknown closure call bypassed the ABI dispatcher:\n%s", output)
	}
	if strings.Contains(output, "call { i32, i32, i64, i64 } (ptr, i32") {
		t.Fatalf("unknown closure call directly used the aggregate return ABI:\n%s", output)
	}
}

func TestEmitFunctionNounwind(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{
		{
			Name:       "main",
			ReturnType: ir.TypeVoid,
			Body: []ir.Instruction{
				{Op: ir.OpClosure, Type: ir.TypeClosure, Result: "fn", Callee: "helper"},
				{Op: ir.OpReturn, Type: ir.TypeVoid},
			},
		},
		{
			Name:       "helper",
			ReturnType: ir.TypeNumber,
			Body: []ir.Instruction{
				{Op: ir.OpConst, Type: ir.TypeNumber, Result: "res", Value: "42"},
				{Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"res"}},
			},
		},
	}}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "define i32 @main(i32 %argc, ptr %argv) nounwind") {
		t.Errorf("main function definition missing nounwind attribute:\n%s", output)
	}
	if !strings.Contains(output, "define internal double @helper() nounwind") {
		t.Errorf("internal function definition missing nounwind attribute:\n%s", output)
	}
	if !strings.Contains(output, "define internal void @helper$invoke(ptr %env, i32 %t0, i32 %f0, i64 %p0, i32 %t1, i32 %f1, i64 %p1, i32 %t2, i32 %f2, i64 %p2, i32 %t3, i32 %f3, i64 %p3, ptr %out) nounwind") {
		t.Errorf("invoke adapter definition missing nounwind attribute:\n%s", output)
	}
	if !strings.Contains(output, "define internal i32 @__scriptgo_to_int32(double %val) alwaysinline nounwind") {
		t.Errorf("__scriptgo_to_int32 definition missing nounwind attribute:\n%s", output)
	}
}
