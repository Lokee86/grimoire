package main

import "testing"

func TestBuildTaggedFilesRetainCallableContracts(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/tagged\n\ngo 1.22\n",
		"enabled_default.go": `//go:build !special

package tagged

func Enabled() bool { return true }
`,
		"enabled_special.go": `//go:build special

package tagged

func Enabled() bool { return false }
`,
		"special_test.go": `//go:build special

package tagged

import (
	"os"
	"os/exec"
	"testing"
)

func TestTagged(t *testing.T) {
	if Enabled() { t.Fatal("unexpected") }
	exe, _ := os.Executable()
	cmd := exec.Command(exe)
	_ = cmd.Start()
	_ = append([]int{}, 1)
}
`,
	})

	facts, summary, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.UnresolvedCalls != 0 {
		t.Fatalf("unresolved calls = %d: %#v", summary.UnresolvedCalls, facts.Unresolved)
	}
	test := hashIdentity("test:example.com/tagged:TestTagged")
	enabled := hashIdentity("function:example.com/tagged:Enabled")
	if !hasEdge(facts, test, enabled, RelCalls) {
		t.Fatal("build-tagged call did not collapse to canonical function identity")
	}
	for _, name := range []string{"Executable", "Command", "Start", "append"} {
		if !hasNode(facts, KindFunction, name) && !hasNode(facts, KindMethod, name) {
			t.Fatalf("missing callable contract node %q", name)
		}
	}
}

func TestFunctionParameterFlowProducesPossibleCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/callbacks\n\ngo 1.22\n",
		"main.go": `package callbacks

func first() {}
func second() {}
func apply(callback func()) { callback() }
func caller(flag bool) {
	callback := first
	if flag { callback = second }
	apply(callback)
}
`,
	})

	facts, summary, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.UnresolvedCalls != 0 {
		t.Fatalf("unresolved calls = %d: %#v", summary.UnresolvedCalls, facts.Unresolved)
	}
	apply := hashIdentity("function:example.com/callbacks:apply")
	first := hashIdentity("function:example.com/callbacks:first")
	second := hashIdentity("function:example.com/callbacks:second")
	if !hasEdge(facts, apply, first, RelPossibleCalls) || !hasEdge(facts, apply, second, RelPossibleCalls) {
		t.Fatal("callback parameter did not retain both conservative targets")
	}
}
