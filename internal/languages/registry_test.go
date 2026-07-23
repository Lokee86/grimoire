package languages

import (
	"reflect"
	"testing"
)

func TestSvelteUsesTypeScriptAdapter(t *testing.T) {
	if got, want := ForPath("src/App.svelte"), []string{"typescript"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ForPath(App.svelte) = %v, want %v", got, want)
	}
	if !OwnsSource("typescript", "src/App.svelte") {
		t.Fatal("typescript adapter does not own .svelte source")
	}
}
