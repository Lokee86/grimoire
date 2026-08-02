package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
)

//go:embed compiler/src/*.java
var compilerSources embed.FS

var compilerBuild struct {
	sync.Once
	runtime compilerRuntime
	err     error
}

func prepareCompilerRuntime() (compilerRuntime, error) {
	compilerBuild.Do(func() {
		if packaged, ok := packagedCompilerRuntime(); ok {
			compilerBuild.runtime = packaged
			return
		}
		compilerBuild.runtime, compilerBuild.err = compileEmbeddedCompiler()
	})
	return compilerBuild.runtime, compilerBuild.err
}

func packagedCompilerRuntime() (compilerRuntime, bool) {
	executable, err := os.Executable()
	if err != nil {
		return compilerRuntime{}, false
	}
	root := filepath.Dir(executable)
	java := filepath.Join(root, "runtime", "bin", executableName("java"))
	jar := filepath.Join(root, "compiler", "lexicon-java-compiler.jar")
	if regularFile(java) && regularFile(jar) {
		return compilerRuntime{java: java, classpath: jar}, true
	}
	return compilerRuntime{}, false
}

func compileEmbeddedCompiler() (compilerRuntime, error) {
	java, err := javaExecutable("LEXICON_JAVA", "java")
	if err != nil {
		return compilerRuntime{}, err
	}
	javac, err := javaExecutable("LEXICON_JAVAC", "javac")
	if err != nil {
		return compilerRuntime{}, err
	}
	root, err := os.MkdirTemp("", "lexicon-java-compiler-")
	if err != nil {
		return compilerRuntime{}, fmt.Errorf("create javac compiler directory: %w", err)
	}
	sourceRoot := filepath.Join(root, "src")
	classes := filepath.Join(root, "classes")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		return compilerRuntime{}, fmt.Errorf("create javac source directory: %w", err)
	}
	if err := os.MkdirAll(classes, 0o755); err != nil {
		return compilerRuntime{}, fmt.Errorf("create javac class directory: %w", err)
	}

	entries, err := fs.Glob(compilerSources, "compiler/src/*.java")
	if err != nil {
		return compilerRuntime{}, fmt.Errorf("enumerate embedded compiler sources: %w", err)
	}
	sort.Strings(entries)
	arguments := []string{"-encoding", "UTF-8", "-d", classes}
	for _, entry := range entries {
		content, readErr := compilerSources.ReadFile(entry)
		if readErr != nil {
			return compilerRuntime{}, fmt.Errorf("read embedded compiler source: %w", readErr)
		}
		path := filepath.Join(sourceRoot, filepath.Base(entry))
		if writeErr := os.WriteFile(path, content, 0o644); writeErr != nil {
			return compilerRuntime{}, fmt.Errorf("write embedded compiler source: %w", writeErr)
		}
		arguments = append(arguments, path)
	}
	command := exec.Command(javac, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		return compilerRuntime{}, fmt.Errorf("compile javac semantic engine: %w: %s", err, boundedText(string(output), 2000))
	}
	return compilerRuntime{java: java, classpath: classes}, nil
}
