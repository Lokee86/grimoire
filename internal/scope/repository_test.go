package scope

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCopiesOnlySelectedSourcesAndLanguageConfiguration(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "a.py", "value = 1\n")
	writeFixture(t, root, "b.py", "value = 2\n")
	writeFixture(t, root, "pyproject.toml", "[project]\nname = 'example'\n")
	repository, err := Build(root, t.TempDir(), "python", []string{"a.py"})
	if err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(repository, "a.py"))
	assertExists(t, filepath.Join(repository, "pyproject.toml"))
	assertMissing(t, filepath.Join(repository, "b.py"))
}

func TestBuildScopesJavaScriptAndCopiesJavaScriptConfiguration(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "src/a.js", "export const value = 1;\n")
	writeFixture(t, root, "src/b.js", "export const value = 2;\n")
	writeFixture(t, root, "jsconfig.json", "{\"compilerOptions\":{}}\n")
	writeFixture(t, root, "package.json", "{\"type\":\"module\"}\n")
	repository, err := Build(root, t.TempDir(), "typescript", []string{"src/a.js"})
	if err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(repository, "src", "a.js"))
	assertExists(t, filepath.Join(repository, "jsconfig.json"))
	assertExists(t, filepath.Join(repository, "package.json"))
	assertMissing(t, filepath.Join(repository, "src", "b.js"))
}

func TestBuildExpandsGoPackageCompanions(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module example.com/test\n")
	writeFixture(t, root, "pkg/a.go", "package pkg\n")
	writeFixture(t, root, "pkg/b.go", "package pkg\n")
	writeFixture(t, root, "other/c.go", "package other\n")
	repository, err := Build(root, t.TempDir(), "go", []string{"pkg/a.go"})
	if err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(repository, "pkg", "a.go"))
	assertExists(t, filepath.Join(repository, "pkg", "b.go"))
	assertExists(t, filepath.Join(repository, "go.mod"))
	assertMissing(t, filepath.Join(repository, "other", "c.go"))
}

func writeFixture(t *testing.T, root, relative, data string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be absent: %v", path, err)
	}
}
