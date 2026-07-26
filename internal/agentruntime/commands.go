package agentruntime

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const providerConfigVersion = 1

type providerConfiguration struct {
	Version        int    `json:"version"`
	LexiconCommand string `json:"lexicon_command"`
	ArcanaCommand  string `json:"arcana_command"`
}

func adjacentCommand(name string) string {
	executable, err := os.Executable()
	if err != nil {
		return name
	}
	return adjacentCommandFrom(executable, name)
}

func resolveProviderCommand(root, requested, name string) string {
	if requested = strings.TrimSpace(requested); requested != "" {
		return requested
	}
	if configured := repositoryConfiguredCommand(root, name); configured != "" {
		return configured
	}
	if executable, err := os.Executable(); err == nil {
		if adjacent := adjacentCommandFrom(executable, name); adjacent != name {
			return adjacent
		}
	}
	for _, checkout := range providerCheckoutRoots(root) {
		if command := checkedOutCommand(checkout, name); command != "" {
			return command
		}
	}
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return name
}

func repositoryConfiguredCommand(root, name string) string {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	data, err := os.ReadFile(filepath.Join(root, ".grimoire", "providers.json"))
	if err != nil {
		return ""
	}
	var configuration providerConfiguration
	if json.Unmarshal(data, &configuration) != nil ||
		(configuration.Version != 0 && configuration.Version != providerConfigVersion) {
		return ""
	}
	command := ""
	switch name {
	case "lexicon":
		command = configuration.LexiconCommand
	case "arcana":
		command = configuration.ArcanaCommand
	}
	return resolveConfiguredCommand(root, command)
}

func resolveConfiguredCommand(root, command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	if filepath.IsAbs(command) {
		if regularFile(command) {
			return filepath.Clean(command)
		}
		return ""
	}
	candidate := filepath.Join(root, command)
	if regularFile(candidate) {
		absolute, err := filepath.Abs(candidate)
		if err == nil {
			return absolute
		}
	}
	if path, err := exec.LookPath(command); err == nil {
		return path
	}
	return ""
}

func providerCheckoutRoots(repositoryRoot string) []string {
	seen := make(map[string]bool)
	roots := make([]string, 0, 4)
	add := func(path string) {
		if path = strings.TrimSpace(path); path == "" {
			return
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return
		}
		absolute = filepath.Clean(absolute)
		if !seen[absolute] {
			seen[absolute] = true
			roots = append(roots, absolute)
		}
	}
	add(os.Getenv("GRIMOIRE_HOME"))
	add(checkoutFromLexiconConfig(repositoryRoot))
	if cwd, err := os.Getwd(); err == nil {
		add(findGrimoireCheckout(cwd))
	}
	if executable, err := os.Executable(); err == nil {
		add(findGrimoireCheckout(filepath.Dir(executable)))
	}
	return roots
}

func checkoutFromLexiconConfig(repositoryRoot string) string {
	if strings.TrimSpace(repositoryRoot) == "" {
		repositoryRoot = "."
	}
	data, err := os.ReadFile(filepath.Join(repositoryRoot, ".lexicon", "config.json"))
	if err != nil {
		return ""
	}
	var configuration struct {
		AdapterRoot string `json:"adapter_root"`
	}
	if json.Unmarshal(data, &configuration) != nil || strings.TrimSpace(configuration.AdapterRoot) == "" {
		return ""
	}
	adapterRoot := configuration.AdapterRoot
	if !filepath.IsAbs(adapterRoot) {
		adapterRoot = filepath.Join(repositoryRoot, adapterRoot)
	}
	return findGrimoireCheckout(adapterRoot)
}

func findGrimoireCheckout(start string) string {
	if strings.TrimSpace(start) == "" {
		return ""
	}
	path, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}
	for {
		data, readErr := os.ReadFile(filepath.Join(path, "go.mod"))
		if readErr == nil && strings.Contains(string(data), "module github.com/Lokee86/grimoire") {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

func checkedOutCommand(checkout, name string) string {
	if checkout == "" {
		return ""
	}
	var directories []string
	switch name {
	case "lexicon":
		directories = []string{
			checkout,
			filepath.Join(checkout, "bin"),
			filepath.Join(checkout, "lexicon"),
			filepath.Join(checkout, "lexicon", "bin"),
		}
	case "arcana":
		directories = []string{
			checkout,
			filepath.Join(checkout, "bin"),
			filepath.Join(checkout, "arcana"),
			filepath.Join(checkout, "arcana", "bin"),
			filepath.Join(checkout, "arcana", "target", "release"),
			filepath.Join(checkout, "arcana", "target", "debug"),
		}
	default:
		return ""
	}
	for _, directory := range directories {
		for _, candidate := range commandNames(name) {
			path := filepath.Join(directory, candidate)
			if regularFile(path) {
				return path
			}
		}
	}
	return ""
}

func commandNames(name string) []string {
	if runtime.GOOS == "windows" {
		return []string{name + ".exe", name}
	}
	return []string{name}
}

func adjacentCommandFrom(executable, name string) string {
	for _, candidate := range commandNames(name) {
		path := filepath.Join(filepath.Dir(executable), candidate)
		if regularFile(path) {
			return path
		}
	}
	return name
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
