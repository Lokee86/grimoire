package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestEngineNamespaceForwardsNativeArguments(t *testing.T) {
	var command string
	var arguments []string
	dependencies := engineDependencies{
		resolve: func(root, requested, name string) string {
			if root != "sample" || requested != "" || name != "lexicon" {
				t.Fatalf("unexpected resolution request: root=%q requested=%q name=%q", root, requested, name)
			}
			return "C:/bundle/lexicon.exe"
		},
		run: func(actualCommand string, actualArguments []string, stdin io.Reader, stdout, stderr io.Writer) error {
			command = actualCommand
			arguments = append([]string(nil), actualArguments...)
			_, _ = io.WriteString(stdout, "native output")
			return nil
		},
	}
	var stdout bytes.Buffer
	if err := runEngineNamespaceWith(
		lexiconEngine,
		[]string{"doctor", "--repo", "sample"},
		&stdout,
		&bytes.Buffer{},
		dependencies,
	); err != nil {
		t.Fatal(err)
	}
	if command != "C:/bundle/lexicon.exe" {
		t.Fatalf("command = %q", command)
	}
	if !reflect.DeepEqual(arguments, []string{"doctor", "--repo", "sample"}) {
		t.Fatalf("arguments = %#v", arguments)
	}
	if stdout.String() != "native output" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestArcanaVersionAliasUsesNativeFlag(t *testing.T) {
	var arguments []string
	dependencies := engineDependencies{
		resolve: func(_, _, _ string) string { return "arcana" },
		run: func(_ string, actual []string, _ io.Reader, _, _ io.Writer) error {
			arguments = append([]string(nil), actual...)
			return nil
		},
	}
	if err := runEngineNamespaceWith(
		arcanaEngine,
		[]string{"--version"},
		&bytes.Buffer{},
		&bytes.Buffer{},
		dependencies,
	); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(arguments, []string{"--version"}) {
		t.Fatalf("arguments = %#v", arguments)
	}
}

func TestEngineCheckReportsResolvedCommandAndVersion(t *testing.T) {
	dependencies := engineDependencies{
		resolve: func(_, _, _ string) string { return "C:/bundle/arcana.exe" },
		run: func(_ string, arguments []string, _ io.Reader, stdout, _ io.Writer) error {
			if !reflect.DeepEqual(arguments, []string{"--version"}) {
				t.Fatalf("arguments = %#v", arguments)
			}
			_, _ = io.WriteString(stdout, "Arcana 0.4.0\n")
			return nil
		},
	}
	var output bytes.Buffer
	if err := runEngineNamespaceWith(
		arcanaEngine,
		[]string{"check"},
		&output,
		&bytes.Buffer{},
		dependencies,
	); err != nil {
		t.Fatal(err)
	}
	var response struct {
		Engine  string `json:"engine"`
		Command string `json:"command"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Engine != "arcana" || response.Command != "C:/bundle/arcana.exe" || response.Version != "Arcana 0.4.0" {
		t.Fatalf("response = %+v", response)
	}
}

func TestEngineHelpDoesNotRequireProviderBinary(t *testing.T) {
	dependencies := engineDependencies{
		resolve: func(_, _, _ string) string {
			t.Fatal("help attempted provider resolution")
			return ""
		},
		run: func(string, []string, io.Reader, io.Writer, io.Writer) error {
			t.Fatal("help attempted provider execution")
			return nil
		},
	}
	var output bytes.Buffer
	if err := runEngineNamespaceWith(
		lexiconEngine,
		[]string{"--help"},
		&output,
		&bytes.Buffer{},
		dependencies,
	); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte("grimoire lexicon")) || !bytes.Contains(output.Bytes(), []byte("doctor")) {
		t.Fatalf("unexpected help:\n%s", output.String())
	}
}

func TestRunDispatchesEngineHelpWithoutProviderBinary(t *testing.T) {
	for _, command := range []string{"lexicon", "arcana"} {
		var output bytes.Buffer
		if err := Run([]string{command, "help"}, &output, &bytes.Buffer{}); err != nil {
			t.Fatalf("Run(%q help): %v", command, err)
		}
		if !bytes.Contains(output.Bytes(), []byte("grimoire "+command)) {
			t.Fatalf("unexpected %s help:\n%s", command, output.String())
		}
	}
}

func TestEngineProcessExitCodeIsPreserved(t *testing.T) {
	err := &ExitError{Code: 7}
	var actual *ExitError
	if !errors.As(err, &actual) || actual.Code != 7 {
		t.Fatalf("unexpected exit error: %v", err)
	}
}
