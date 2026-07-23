package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestAnalyzeExtractsGDScriptSlice(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "project.godot", `[application]
config/name="DependencyFixture"
[editor_plugins]
enabled=PackedStringArray("res://addons/example/plugin.cfg")
[autoload]
State="*res://scripts/base.gd"
[rendering]
resource="res://scenes/main.tscn"
`)
	writeFixture(t, root, "addons/example/plugin.cfg", "[plugin]\nname=\"Example\"\n")
	writeFixture(t, root, "scenes/main.tscn", "[gd_scene]\n")
	writeFixture(t, root, "scripts/base.gd", `class_name Base
extends Node
signal changed(value)
const LIMIT = 3
var title: String = "# is not a comment"
func greet(name):
    return name
`)
	writeFixture(t, root, "scripts/player.gd", `# func ignored()
class_name Player extends Base
@onready var scene = preload("res://scripts/base.gd")
signal spawned
func greet(text):
    return text
func run(
    value: int,
):
    var message = "load(\\\"res://fake.gd\\\") # string"
    greet(message)
    load(get_path())
`)
	writeFixture(t, root, ".worktrees/ignored.gd", "class_name Ignored\n")
	writeFixture(t, root, "vendor/ignored.gd", "class_name IgnoredVendor\n")
	for _, directory := range []string{".ddocs", ".lexicon", ".arcana", ".grimoire", ".pitlord", ".cantrip", ".homunculus", ".incubus", ".ritual", ".warlock"} {
		writeFixture(t, root, filepath.Join(directory, "ignored.gd"), "class_name IgnoredState\n")
	}

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	if records[0]["language"] != language || records[0]["schema_version"] != float64(1) {
		t.Fatalf("unexpected header: %#v", records[0])
	}
	playerType := findNode(records, "type", "Player", "scripts/player.gd")
	if playerType["id"] != nodeID("type", "scripts/player.gd::type::Player") {
		t.Fatalf("unexpected stable type ID: %v", playerType["id"])
	}

	var kinds []string
	var names []string
	var relations []string
	var unresolvedReasons []string
	for _, record := range records[1:] {
		switch record["record"] {
		case "node":
			kinds = append(kinds, record["kind"].(string))
			names = append(names, record["name"].(string))
		case "edge":
			relations = append(relations, record["relation"].(string))
		case "unresolved":
			unresolvedReasons = append(unresolvedReasons, record["reason"].(string))
		}
	}
	for _, expected := range []string{"repository", "directory", "file", "module", "type", "function", "signal", "constant", "variable", "import"} {
		if !contains(kinds, expected) {
			t.Errorf("missing node kind %q in %v", expected, kinds)
		}
	}
	for _, expected := range []string{"Player", "Base", "greet", "run", "changed", "LIMIT", "title", "spawned"} {
		if !contains(names, expected) {
			t.Errorf("missing node name %q in %v", expected, names)
		}
	}
	for _, expected := range []string{"contains", "defines", "imports", "references", "extends", "calls"} {
		if !contains(relations, expected) {
			t.Errorf("missing edge relation %q in %v", expected, relations)
		}
	}
	if !contains(unresolvedReasons, "dynamic-target") {
		t.Errorf("dynamic load/call was not reported unresolved: %v", unresolvedReasons)
	}
	if !contains(relations, "depends-on") || !anyDependencyCategory(records, "plugin") || !anyDependencyCategory(records, "autoload") {
		t.Errorf("missing project dependency facts: relations=%v", relations)
	}
	if contains(names, "Ignored") || contains(names, "IgnoredVendor") {
		t.Fatalf("excluded source was scanned: %v", names)
	}
	if contains(names, "IgnoredState") {
		t.Fatalf("Warlock state source was scanned: %v", names)
	}
}

func anyDependencyCategory(records []map[string]any, category string) bool {
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "depends-on" {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["category"] == category {
			return true
		}
	}
	return false
}

func TestParserHandlesIndentationStringsCommentsAndMultilineDeclarations(t *testing.T) {
	pf, err := parseFile("scene.gd", []byte(`class_name Scene
var text = "func fake() # not a comment"
# signal fake()
func run(
    value: int,
    label = "signal fake()",
):
    var nested = preload(
        "res://other.gd"
    )
`))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, decl := range pf.declarations {
		got = append(got, decl.kind+":"+decl.name)
	}
	want := []string{"type:Scene", "variable:text", "function:run", "variable:nested"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations = %v, want %v", got, want)
	}
	if len(pf.imports) != 0 || len(pf.calls) != 0 {
		t.Fatalf("parseFile should defer references to the fact pass: imports=%v calls=%v", pf.imports, pf.calls)
	}
}

func TestFactSetDeduplicatesEdgesAndUnresolved(t *testing.T) {
	facts := &factSet{}
	edgeRecord := edge("source", "target", "calls", nil)
	unresolvedRecord := unresolved("source", "calls", "dynamic()", "dynamic-target", nil)

	facts.addEdge(edgeRecord)
	facts.addEdge(edgeRecord)
	facts.addUnresolved(unresolvedRecord)
	facts.addUnresolved(unresolvedRecord)

	if len(facts.edges) != 1 || len(facts.edgeOrderKeys) != 1 {
		t.Fatalf("edge deduplication failed: edges=%d keys=%d", len(facts.edges), len(facts.edgeOrderKeys))
	}
	if len(facts.unresolved) != 1 || len(facts.unresolvedOrderKeys) != 1 {
		t.Fatalf("unresolved deduplication failed: unresolved=%d keys=%d", len(facts.unresolved), len(facts.unresolvedOrderKeys))
	}
}

func TestAnalyzeIsDeterministicAcrossRepeatRuns(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "b.gd", "func z():\n    pass\n")
	writeFixture(t, root, "a.gd", "func a():\n    z()\n")
	first, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeat analysis changed the JSONL output")
	}
	for _, line := range strings.Split(strings.TrimSpace(string(first)), "\n") {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record["record"] == "node" && !strings.HasPrefix(record["id"].(string), "sha256:") {
			t.Fatalf("unstable node ID: %v", record["id"])
		}
	}
}

func TestAnalyzeResolvesUniqueClassCallsAndKeepsDynamicCallsUnresolved(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "actor.gd", `class_name Actor
static func spawn():
    pass
`)
	writeFixture(t, root, "caller.gd", `class_name Caller
func run(instance):
    Actor.spawn()
    Actor.new()
    instance.spawn()
`)
	writeFixture(t, root, "ambiguous_one.gd", `class_name Ambiguous
func ping():
    pass
`)
	writeFixture(t, root, "ambiguous_two.gd", `class_name Ambiguous
func ping():
    pass
`)
	writeFixture(t, root, "ambiguous_caller.gd", `func run():
    Ambiguous.ping()
    Ambiguous.new()
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	actor := findNode(records, "type", "Actor", "actor.gd")
	spawn := findNode(records, "function", "spawn", "actor.gd")
	if actor == nil || spawn == nil {
		t.Fatalf("missing Actor declarations: actor=%#v spawn=%#v", actor, spawn)
	}

	resolved := map[string]bool{}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "calls" && record["target"] == spawn["id"] {
			resolved["Actor.spawn()"] = true
		}
		if record["record"] == "edge" && record["relation"] == "calls" && record["target"] == actor["id"] {
			resolved["Actor.new()"] = true
		}
	}
	if !resolved["Actor.spawn()"] || !resolved["Actor.new()"] {
		t.Fatalf("class calls were not resolved: %v", resolved)
	}

	var unresolvedCalls []string
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == "calls" {
			unresolvedCalls = append(unresolvedCalls, record["expression"].(string))
		}
	}
	for _, expected := range []string{"instance.spawn()", "Ambiguous.ping()", "Ambiguous.new()"} {
		if !contains(unresolvedCalls, expected) {
			t.Errorf("missing unresolved dotted call %q in %v", expected, unresolvedCalls)
		}
	}
	for _, unexpected := range []string{"Actor.spawn()", "Actor.new()"} {
		if contains(unresolvedCalls, unexpected) {
			t.Errorf("unique class call remained unresolved: %q", unexpected)
		}
	}
}

func TestAnalyzeResolvesOnlyExplicitlyTypedReceivers(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "actor.gd", `class_name Actor
func spawn():
    pass
`)
	writeFixture(t, root, "caller.gd", `class_name Caller
var member: Actor
func run(argument: Actor):
    var local: Actor
    argument.spawn()
    local.spawn()
    member.spawn()
`)
	writeFixture(t, root, "ambiguous_one.gd", `class_name Duplicate
func spawn():
    pass
`)
	writeFixture(t, root, "ambiguous_two.gd", `class_name Duplicate
func spawn():
    pass
`)
	writeFixture(t, root, "ambiguous_caller.gd", `func run(argument: Duplicate):
    argument.spawn()
    Node.spawn()
    missing.spawn()
    untyped.spawn()
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	spawn := findNode(records, "function", "spawn", "actor.gd")
	if spawn == nil {
		t.Fatal("missing Actor.spawn declaration")
	}

	resolved := 0
	var unresolvedCalls []string
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "calls" && record["target"] == spawn["id"] {
			resolved++
		}
		if record["record"] == "unresolved" && record["relation"] == "calls" {
			unresolvedCalls = append(unresolvedCalls, record["expression"].(string))
		}
	}
	if resolved != 3 {
		t.Fatalf("typed parameter/local/member calls resolved=%d, want 3", resolved)
	}
	for _, expected := range []string{"argument.spawn()", "Node.spawn()", "missing.spawn()", "untyped.spawn()"} {
		if !contains(unresolvedCalls, expected) {
			t.Errorf("receiver call was not left unresolved: %q in %v", expected, unresolvedCalls)
		}
	}
}

func TestAnalyzePrefersSameFileClassDeclarations(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "external.gd", `class_name Widget
static func ping():
    pass
`)
	writeFixture(t, root, "caller.gd", `class_name Widget
static func ping():
    pass
var member: Widget
func run():
    Widget.ping()
    Widget.new()
    member.ping()
`)
	writeFixture(t, root, "ambiguous.gd", `class_name Duplicate
static func ping():
    pass
`)
	writeFixture(t, root, "ambiguous_caller.gd", `class_name Duplicate
static func ping():
    pass
class_name Duplicate
static func ping():
    pass
var member: Duplicate
func run():
    Duplicate.ping()
    Duplicate.new()
    member.ping()
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	localWidget := findNode(records, "type", "Widget", "caller.gd")
	localPing := findNode(records, "function", "ping", "caller.gd")
	externalPing := findNode(records, "function", "ping", "external.gd")
	if localWidget == nil || localPing == nil || externalPing == nil {
		t.Fatalf("missing Widget declarations: local=%#v local method=%#v external method=%#v", localWidget, localPing, externalPing)
	}

	var localCalls, externalCalls int
	var unresolvedCalls []string
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "calls" {
			switch record["target"] {
			case localPing["id"], localWidget["id"]:
				localCalls++
			case externalPing["id"]:
				externalCalls++
			}
		}
		if record["record"] == "unresolved" && record["relation"] == "calls" {
			unresolvedCalls = append(unresolvedCalls, record["expression"].(string))
		}
	}
	if localCalls != 3 || externalCalls != 0 {
		t.Fatalf("same-file calls resolved local=%d external=%d unresolved=%v", localCalls, externalCalls, unresolvedCalls)
	}
	for _, expected := range []string{"Duplicate.ping()", "Duplicate.new()", "member.ping()"} {
		if !contains(unresolvedCalls, expected) {
			t.Errorf("ambiguous same-file call was resolved: %q in %v", expected, unresolvedCalls)
		}
	}
}

func TestAnalyzeResolvesStaticPreloadAliasesConservatively(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "scripts/tool.gd", `static func build():
    pass
func instance_method():
    pass
`)
	writeFixture(t, root, "scripts/typed_tool.gd", `class_name TypedTool
`)
	writeFixture(t, root, "scripts/ambiguous.gd", `class_name Duplicate
class_name Duplicate
`)
	writeFixture(t, root, "caller.gd", `const Tool = preload("res://scripts/tool.gd")
const TypedAlias = preload("res://scripts/typed_tool.gd")
const Missing = preload("res://scripts/missing.gd")
const Dynamic = preload(script_path)
const Loaded = load("res://scripts/tool.gd")
const ResourceAlias = preload("res://data/resource.tres")
const Ambiguous = preload("res://scripts/ambiguous.gd")
func run():
    Tool.new()
    Tool.build()
    Tool.instance_method()
    Tool.missing()
    TypedAlias.new()
    Missing.new()
    Dynamic.new()
    Loaded.new()
    ResourceAlias.new()
    Ambiguous.new()
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	toolModule := findNode(records, "module", "tool", "scripts/tool.gd")
	build := findNode(records, "function", "build", "scripts/tool.gd")
	typedTool := findNode(records, "type", "TypedTool", "scripts/typed_tool.gd")
	if toolModule == nil || build == nil || typedTool == nil {
		t.Fatalf("missing preload targets: module=%#v build=%#v type=%#v", toolModule, build, typedTool)
	}

	resolvedTargets := map[string]int{}
	var unresolvedCalls []string
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "calls" {
			resolvedTargets[record["target"].(string)]++
		}
		if record["record"] == "unresolved" && record["relation"] == "calls" {
			unresolvedCalls = append(unresolvedCalls, record["expression"].(string))
		}
	}
	if resolvedTargets[toolModule["id"].(string)] != 2 || resolvedTargets[build["id"].(string)] != 1 || resolvedTargets[typedTool["id"].(string)] != 1 {
		t.Fatalf("preload aliases resolved to unexpected targets: %v", resolvedTargets)
	}
	for _, expected := range []string{"Tool.instance_method()", "Tool.missing()", "Missing.new()", "Dynamic.new()", "ResourceAlias.new()", "Ambiguous.new()"} {
		if !contains(unresolvedCalls, expected) {
			t.Errorf("preload call was not left unresolved: %q in %v", expected, unresolvedCalls)
		}
	}
}

func TestAnalyzeResolvesInheritanceSelfAndSuperCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "base.gd", `class_name Base
func _init():
    pass
func ping():
    pass
`)
	writeFixture(t, root, "child.gd", `class_name Child extends Base
func run():
    ping()
    self.ping()
    super.ping()
    super()
`)
	records := analyzeFixture(t, root)
	ping := findNode(records, "function", "ping", "base.gd")
	init := findNode(records, "function", "_init", "base.gd")
	if countEdges(records, "calls", ping["id"].(string)) != 3 {
		t.Fatalf("inherited ping calls were not resolved")
	}
	if countEdges(records, "calls", init["id"].(string)) != 1 {
		t.Fatalf("super constructor was not resolved")
	}
}

func TestAnalyzeEmitsOverridesAndPolymorphicDispatch(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "base.gd", `class_name Base
func ping():
    pass
`)
	writeFixture(t, root, "child.gd", `class_name Child extends Base
func ping():
    pass
`)
	writeFixture(t, root, "caller.gd", `func dispatch(value: Base):
    value.ping()
func exact():
    Child.new().ping()
func dynamic(value, method_name):
    value.call(method_name)
`)
	records := analyzeFixture(t, root)
	basePing := findNode(records, "function", "ping", "base.gd")
	childPing := findNode(records, "function", "ping", "child.gd")
	dispatch := findNode(records, "function", "dispatch", "caller.gd")
	exact := findNode(records, "function", "exact", "caller.gd")
	if basePing == nil || childPing == nil || dispatch == nil || exact == nil {
		t.Fatal("missing polymorphic dispatch declarations")
	}
	if !hasEdgeFromTo(records, "overrides", childPing["id"].(string), basePing["id"].(string)) {
		t.Fatal("child override relationship was not emitted")
	}
	if !hasEdgeFromTo(records, "possible-calls", dispatch["id"].(string), basePing["id"].(string)) ||
		!hasEdgeFromTo(records, "possible-calls", dispatch["id"].(string), childPing["id"].(string)) {
		t.Fatal("typed base receiver did not retain both runtime dispatch targets")
	}
	if !hasEdgeFromTo(records, "calls", exact["id"].(string), childPing["id"].(string)) {
		t.Fatal("concrete Child receiver did not narrow to its override")
	}
	if !hasUnresolved(records, "calls", "value.call", "dynamic-target") {
		t.Fatal("computed GDScript call remained resolved")
	}
}

func TestAnalyzePropagatesConstructorArgumentsAssignmentsAndReturns(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "dependency.gd", `class_name Dependency
func work():
    pass
`)
	writeFixture(t, root, "factory.gd", `class_name Factory
static func make() -> Dependency:
    return Dependency.new()
`)
	writeFixture(t, root, "consumer.gd", `class_name Consumer
var dependency
func configure(value):
    dependency = value
func run():
    dependency.work()
`)
	writeFixture(t, root, "caller.gd", `func execute():
    var dependency := Factory.make()
    var consumer := Consumer.new()
    consumer.configure(dependency)
    consumer.run()
    Factory.make().work()
`)
	records := analyzeFixture(t, root)
	work := findNode(records, "function", "work", "dependency.gd")
	configure := findNode(records, "function", "configure", "consumer.gd")
	run := findNode(records, "function", "run", "consumer.gd")
	for _, target := range []map[string]any{work, configure, run} {
		if target == nil || countEdges(records, "calls", target["id"].(string)) == 0 {
			t.Fatalf("missing propagated call target %#v", target)
		}
	}
	if countEdges(records, "calls", work["id"].(string)) != 2 {
		t.Fatalf("dependency.work calls were not propagated through member and return flow")
	}
}

func TestAnalyzeKeepsInnerClassMethodOwnership(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "nested.gd", `class Inner:
    func ping():
        pass
    func run():
        ping()
func outer():
    Inner.new().run()
`)
	records := analyzeFixture(t, root)
	ping := findNode(records, "function", "ping", "nested.gd")
	run := findNode(records, "function", "run", "nested.gd")
	if ping == nil || run == nil {
		t.Fatal("inner class methods were not emitted")
	}
	if countEdges(records, "calls", ping["id"].(string)) != 1 || countEdges(records, "calls", run["id"].(string)) != 1 {
		t.Fatal("inner class call ownership or resolution is incorrect")
	}
}

func TestAnalyzeResolvesLiteralLoadAliases(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "loaded.gd", `func ping():
    pass
`)
	writeFixture(t, root, "caller.gd", `static var Loaded = load("res://loaded.gd")
var direct = load("res://loaded.gd").new()
func run():
    var value = Loaded.new()
    value.ping()
    direct.ping()
`)
	records := analyzeFixture(t, root)
	ping := findNode(records, "function", "ping", "loaded.gd")
	if ping == nil || countEdges(records, "calls", ping["id"].(string)) != 2 {
		t.Fatal("literal load alias and direct loaded instance did not resolve through construction")
	}
}

func TestAnalyzeResolvesPreloadInstancesAndNestedTypes(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "tools.gd", `class Inner:
    func run():
        pass
func instance_method():
    pass
`)
	writeFixture(t, root, "caller.gd", `const Tools = preload("res://tools.gd")
func execute():
    var tool = Tools.new()
    tool.instance_method()
    Tools.Inner.new().run()
    preload("res://tools.gd").new().instance_method()
`)
	records := analyzeFixture(t, root)
	instanceMethod := findNode(records, "function", "instance_method", "tools.gd")
	run := findNode(records, "function", "run", "tools.gd")
	if instanceMethod == nil || countEdges(records, "calls", instanceMethod["id"].(string)) != 2 {
		t.Fatal("preload instance flow was not resolved")
	}
	if run == nil || countEdges(records, "calls", run["id"].(string)) != 1 {
		t.Fatal("nested preload type flow was not resolved")
	}
}

func TestAnalyzeResolvesConfiguredAutoloadSingletons(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "project.godot", `[application]
config/name="Fixture"

[autoload]
Logger="*res://logger.gd"
`)
	writeFixture(t, root, "logger.gd", `class_name LocalLogger
func write_message():
    pass
`)
	writeFixture(t, root, "caller.gd", `func run():
    Logger.write_message()
`)
	records := analyzeFixture(t, root)
	writeMessage := findNode(records, "function", "write_message", "logger.gd")
	if writeMessage == nil || countEdges(records, "calls", writeMessage["id"].(string)) != 1 {
		t.Fatal("autoload singleton call was not resolved")
	}
}

func TestAnalyzeScopesInlineLambdaCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "service.gd", `class_name LambdaService
func ping():
    pass
`)
	writeFixture(t, root, "lambda.gd", `class_name LambdaOwner
var service := LambdaService.new()
func wire(signal_value: Signal):
    signal_value.connect(
        func():
            service.ping()
    )
`)
	records := analyzeFixture(t, root)
	ping := findNode(records, "function", "ping", "service.gd")
	var lambda map[string]any
	for _, record := range records {
		if record["record"] == "node" && record["kind"] == "function" && strings.HasPrefix(record["name"].(string), "<lambda@") {
			lambda = record
			break
		}
	}
	if ping == nil || lambda == nil {
		t.Fatal("lambda declarations were not emitted")
	}
	if !hasEdgeFromTo(records, "calls", lambda["id"].(string), ping["id"].(string)) {
		t.Fatal("lambda body call was not owned by the lambda")
	}
	if countEdges(records, "possible-calls", lambda["id"].(string)) != 1 {
		t.Fatal("lambda callback target was not emitted")
	}
}

func TestAnalyzeClassifiesExplicitBuiltinReceiverCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "builtin_receiver.gd", `class_name BuiltinReceiver
var control: Control
func wire():
    control.connect("ready", Callable(self, "wire"))
    control.has_signal("ready")
`)
	records := analyzeFixture(t, root)
	for _, candidate := range []string{"control.connect", "control.has_signal"} {
		if !hasUnresolved(records, "calls", candidate, "builtin-target") {
			t.Fatalf("expected %s to be classified as builtin", candidate)
		}
		if hasUnresolved(records, "calls", candidate, "missing-target") {
			t.Fatalf("typed builtin receiver %s was marked missing", candidate)
		}
	}
}

func TestAnalyzeClassifiesUnknownInstanceMethodsAsExternal(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "worker.gd", `class_name Worker
func run():
    var child := Worker.new()
    child.queue_free()
`)
	records := analyzeFixture(t, root)
	if !hasUnresolved(records, "calls", "child.queue_free", "external-target") {
		t.Fatal("unknown instance method was not kept as external dispatch")
	}
}

func TestAnalyzePropagatesCallablePropertyAssignments(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "session.gd", `class_name Session
var route: Callable
func invoke():
    route.call()
`)
	writeFixture(t, root, "composer.gd", `class_name Composer
func handle():
    pass
func configure(session: Session):
    session.route = Callable(self, "handle")
    session.invoke()
`)
	records := analyzeFixture(t, root)
	handle := findNode(records, "function", "handle", "composer.gd")
	if handle == nil || countEdges(records, "calls", handle["id"].(string)) != 1 {
		t.Fatal("callable property assignment did not propagate to invocation")
	}
}

func TestAnalyzePropagatesCallableMapsThroughDictionaryGet(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "callback_map.gd", `class_name CallbackMap
var callbacks: Dictionary
func configure(value: Dictionary):
    callbacks = value
func invoke():
    var handler: Callable = callbacks.get("ready", Callable())
    handler.call()
func ready():
    pass
func wire():
    configure({"ready": Callable(self, "ready")})
    invoke()
`)
	records := analyzeFixture(t, root)
	ready := findNode(records, "function", "ready", "callback_map.gd")
	if ready == nil || countEdges(records, "calls", ready["id"].(string)) != 1 {
		t.Fatal("dictionary callback target did not propagate to its invocation site")
	}
}

func TestAnalyzePropagatesCallableArgumentsToInvocationSites(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "routes.gd", `class_name Routes
var route
func configure(callback):
    route = callback
func invoke():
    route.call()
func handle():
    pass
func wire():
    configure(Callable(self, "handle"))
    invoke()
`)
	records := analyzeFixture(t, root)
	handle := findNode(records, "function", "handle", "routes.gd")
	if handle == nil || countEdges(records, "calls", handle["id"].(string)) != 1 {
		t.Fatal("callable target did not propagate to its invocation site")
	}
}

func TestAnalyzeEmitsPossibleCallbackCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "callbacks.gd", `class_name Callbacks
func handle():
    pass
func wire(signal_value: Signal, values: Array):
    var callback = Callable(self, "handle")
    signal_value.connect(handle)
    values.map(handle)
`)
	records := analyzeFixture(t, root)
	handle := findNode(records, "function", "handle", "callbacks.gd")
	if handle == nil || countEdges(records, "possible-calls", handle["id"].(string)) < 3 {
		t.Fatalf("callback references were not emitted")
	}
}

func TestJSONLRecordsUseContractOrder(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "one.gd", "class_name One\n")
	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatal("expected header and facts")
	}
	previous := ""
	for _, line := range lines[1:] {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		kind := record["record"].(string)
		key := kind
		if kind == "node" {
			key = "0/" + record["id"].(string) + "/" + record["kind"].(string) + "/" + record["path"].(string) + "/" + record["qualified_name"].(string)
		} else if kind == "edge" {
			key = "1/" + record["source"].(string) + "/" + record["target"].(string) + "/" + record["relation"].(string)
		} else {
			key = "2/" + record["source"].(string) + "/" + record["relation"].(string) + "/" + record["expression"].(string) + "/" + record["reason"].(string)
		}
		if previous != "" && key < previous {
			t.Fatalf("records are not ordered: %q before %q", previous, key)
		}
		previous = key
	}
}

func analyzeFixture(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	return decodeRecords(t, data)
}

func hasUnresolved(records []map[string]any, relation, candidate, reason string) bool {
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == relation && record["candidate_name"] == candidate && record["reason"] == reason {
			return true
		}
	}
	return false
}

func hasEdgeFromTo(records []map[string]any, relation, source, target string) bool {
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source && record["target"] == target {
			return true
		}
	}
	return false
}

func countEdges(records []map[string]any, relation, target string) int {
	count := 0
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["target"] == target {
			count++
		}
	}
	return count
}

func writeFixture(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return records
}

func findNode(records []map[string]any, kind, name, path string) map[string]any {
	for _, record := range records {
		if record["record"] == "node" && record["kind"] == kind && record["name"] == name && record["path"] == path {
			return record
		}
	}
	return nil
}

func contains(values []string, value string) bool {
	return sort.SearchStrings(appendSorted(values), value) < len(values) && appendSorted(values)[sort.SearchStrings(appendSorted(values), value)] == value
}

func appendSorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
