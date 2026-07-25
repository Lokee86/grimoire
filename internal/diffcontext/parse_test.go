package diffcontext

import (
	"reflect"
	"testing"
)

func TestParseModifiedAddedDeletedAndAnchoredHunks(t *testing.T) {
	diff := `diff --git a/internal\old.go b/internal\new.go
similarity index 70%
rename from internal\old.go
rename to internal\new.go
--- a/internal\old.go
+++ b/internal\new.go
@@ -4,2 +4,3 @@ func Resolve() {
 old context
+added
+more
@@ -20,2 +21,0 @@ func Resolve() {
-old
-removed
diff --git a/new.txt b/new.txt
new file mode 100644
--- /dev/null
+++ b/new.txt
@@ -0,0 +1,2 @@
+one
+two
diff --git a/deleted.txt b/deleted.txt
deleted file mode 100644
--- a/deleted.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-old
-lines
`

	got := Parse(diff)
	want := []Change{
		{Path: "deleted.txt", StartLine: 1, EndLine: 1, Deleted: true},
		{Path: "internal/new.go", OldPath: "internal/old.go", StartLine: 4, EndLine: 6, Summary: "func Resolve() {"},
		{Path: "internal/new.go", OldPath: "internal/old.go", StartLine: 21, EndLine: 21, Summary: "func Resolve() {"},
		{Path: "new.txt", StartLine: 1, EndLine: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}

func TestParseFileLevelChangesWithoutHunks(t *testing.T) {
	diff := "diff --git a/assets/icon.bin b/assets/icon.bin\nBinary files a/assets/icon.bin and b/assets/icon.bin differ\n" +
		"diff --git a/old.bin b/old.bin\ndeleted file mode 100644\nBinary files a/old.bin and /dev/null differ\n"
	got := Parse(diff)
	want := []Change{
		{Path: "assets/icon.bin", StartLine: 1, EndLine: 1, Summary: "changed file"},
		{Path: "old.bin", StartLine: 1, EndLine: 1, Deleted: true, Summary: "deleted file"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}

func TestParsePureRenameUsesCurrentPathAndAnchor(t *testing.T) {
	got := Parse("diff --git a/pkg/old.go b/pkg/new.go\nsimilarity index 100%\nrename from pkg/old.go\nrename to pkg/new.go\n")
	want := []Change{{Path: "pkg/new.go", OldPath: "pkg/old.go", StartLine: 1, EndLine: 1, Summary: "renamed file"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}
