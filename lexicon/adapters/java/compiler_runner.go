package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type compilerFact struct {
	Record         string `json:"record"`
	SourceKind     string `json:"source_kind"`
	SourceIdentity string `json:"source_identity"`
	TargetKind     string `json:"target_kind"`
	TargetIdentity string `json:"target_identity"`
	Relation       string `json:"relation"`
	Path           string `json:"path"`
	StartLine      int    `json:"start_line"`
	StartColumn    int    `json:"start_column"`
	EndLine        int    `json:"end_line"`
	EndColumn      int    `json:"end_column"`
	Engine         string `json:"engine"`
	Reason         string `json:"reason"`
}

type compilerRuntime struct {
	classpath string
	java      string
}

func (state *analysisState) resolveCompilerSemantics(snapshot repositorySnapshot) error {
	if len(snapshot.sources) == 0 {
		return nil
	}
	runtime, err := prepareCompilerRuntime()
	if err != nil {
		return err
	}
	sourceList, err := writeCompilerSourceList(snapshot.sources)
	if err != nil {
		return err
	}
	defer os.Remove(sourceList)

	arguments := []string{
		"-Xmx" + compilerHeap(),
		"--add-modules", "jdk.compiler",
		"-cp", runtime.classpath,
		"lexicon.java.CompilerMain",
		"--repo", snapshot.root,
		"--sources-file", sourceList,
	}
	command := exec.Command(runtime.java, arguments...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open javac semantic output: %w", err)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return fmt.Errorf("start javac semantic analysis: %w", err)
	}

	evidence := buildCompilerEvidenceIndex(state.facts)
	mismatches := 0
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var fact compilerFact
		if err := json.Unmarshal(scanner.Bytes(), &fact); err != nil {
			_ = command.Process.Kill()
			_ = command.Wait()
			return fmt.Errorf("decode javac semantic fact: %w", err)
		}
		if !state.applyCompilerFact(fact, &evidence) {
			mismatches++
		}
	}
	scanErr := scanner.Err()
	waitErr := command.Wait()
	if scanErr != nil {
		return fmt.Errorf("read javac semantic output: %w", scanErr)
	}
	if waitErr != nil {
		return fmt.Errorf("javac semantic analysis: %w: %s", waitErr, boundedText(stderr.String(), 2000))
	}
	if mismatches != 0 {
		state.facts.addUnresolved(
			state.repositoryID,
			"defines",
			fmt.Sprintf("%d compiler relationships did not match modeled declarations", mismatches),
			"compiler-identity-mismatch",
			"",
			nil,
			map[string]any{"engine": "javac", "mismatch_count": mismatches},
		)
	}
	return nil
}
