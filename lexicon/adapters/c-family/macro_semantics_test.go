package main

import (
	"reflect"
	"testing"
)

func TestGrowArrayMacroRecordsParametersAndNestedBodyCall(t *testing.T) {
	macro := loadMacro(t, "GROW_ARRAY", `#define GROW_ARRAY(type, array, length, capacity) do { if ((length) == (capacity)) { (capacity) *= 2; (array) = realloc((array), sizeof(*(array)) * (capacity)); } } while (0)`)
	if !reflect.DeepEqual(macro.MacroParameters, []string{"type", "array", "length", "capacity"}) {
		t.Fatalf("parameters = %#v", macro.MacroParameters)
	}
	if len(macro.MacroCalls) != 1 {
		t.Fatalf("calls = %#v", macro.MacroCalls)
	}
	call := macro.MacroCalls[0]
	if call.Callee != "realloc" || !reflect.DeepEqual(call.Arguments, []string{"(array)", "sizeof(*(array)) * (capacity)"}) {
		t.Fatalf("realloc call = %#v", call)
	}
	if call.Unsupported {
		t.Fatalf("supported call marked unsupported: %#v", call)
	}
}

func TestMacroSemanticModelKeepsNestedAndMultipleCallsInOrder(t *testing.T) {
	macro := loadMacro(t, "MULTI", `#define MULTI(value) first(inner(value)); second(value)`)
	if got := []string{macro.MacroCalls[0].Callee, macro.MacroCalls[1].Callee, macro.MacroCalls[2].Callee}; !reflect.DeepEqual(got, []string{"first", "inner", "second"}) {
		t.Fatalf("call order = %#v", got)
	}
	if !reflect.DeepEqual(macro.MacroCalls[0].Arguments, []string{"inner(value)"}) || !reflect.DeepEqual(macro.MacroCalls[1].Arguments, []string{"value"}) || !reflect.DeepEqual(macro.MacroCalls[2].Arguments, []string{"value"}) {
		t.Fatalf("call arguments = %#v", macro.MacroCalls)
	}
}

func TestMacroSemanticModelPreservesParameterSubstitution(t *testing.T) {
	macro := loadMacro(t, "APPLY", `#define APPLY(array, index) push(array[index], index)`)
	if len(macro.MacroCalls) != 1 || macro.MacroCalls[0].Callee != "push" {
		t.Fatalf("calls = %#v", macro.MacroCalls)
	}
	if macro.MacroTarget != "push" {
		t.Fatalf("legacy macro target = %q", macro.MacroTarget)
	}
	if !reflect.DeepEqual(macro.MacroCalls[0].Arguments, []string{"array[index]", "index"}) {
		t.Fatalf("raw arguments = %#v", macro.MacroCalls[0].Arguments)
	}
}

func TestMacroSemanticModelMarksUnsupportedSubstitutions(t *testing.T) {
	pasted := loadMacro(t, "PASTE", `#define PASTE(left, right) invoke(left ## right)`)
	if len(pasted.MacroCalls) != 1 || !pasted.MacroCalls[0].TokenPasting || !pasted.MacroCalls[0].Unsupported {
		t.Fatalf("token-pasting call = %#v", pasted.MacroCalls)
	}
	stringified := loadMacro(t, "STRINGIFY", `#define STRINGIFY(value) log(#value)`)
	if len(stringified.MacroCalls) != 1 || !stringified.MacroCalls[0].Stringification || !stringified.MacroCalls[0].Unsupported {
		t.Fatalf("stringification call = %#v", stringified.MacroCalls)
	}
	variadic := loadMacro(t, "LOG", `#define LOG(format, ...) log(format, __VA_ARGS__)`)
	if !reflect.DeepEqual(variadic.MacroParameters, []string{"format", "__VA_ARGS__"}) || len(variadic.MacroCalls) != 1 || !variadic.MacroCalls[0].VariadicSubstitution || !variadic.MacroCalls[0].Unsupported {
		t.Fatalf("variadic model = %#v", variadic)
	}
}

func loadMacro(t *testing.T, name, definition string) *declaration {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, map[string]string{"macro.h": definition + "\n"})
	model, err := buildRepositoryModel(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range model.Declarations {
		if declaration.Name == name && declaration.MacroFunction {
			return declaration
		}
	}
	t.Fatalf("macro %q not found", name)
	return nil
}
