package main

import "testing"

func TestAnalyzeResolvesNestedProjectResourcesAndScopedAutoloads(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "game_a/project.godot", `[autoload]
Service="*res://scripts/service.gd"
`)
	writeFixture(t, root, "game_a/scripts/base.gd", `func base_call():
    pass
`)
	writeFixture(t, root, "game_a/scripts/service.gd", `func run():
    pass
`)
	writeFixture(t, root, "game_a/scripts/caller.gd", `extends "res://scripts/base.gd"
const ServiceScript := preload("res://scripts/service.gd")
func call_service():
    var instance := ServiceScript.new()
    instance.run()
    Service.run()
`)
	writeFixture(t, root, "game_b/project.godot", `[autoload]
Service="*res://scripts/service.gd"
`)
	writeFixture(t, root, "game_b/scripts/service.gd", `func run():
    pass
`)
	writeFixture(t, root, "game_b/scripts/caller.gd", `func call_service():
    Service.run()
`)

	records := analyzeFixture(t, root)
	callerA := findNode(records, "function", "call_service", "game_a/scripts/caller.gd")
	callerB := findNode(records, "function", "call_service", "game_b/scripts/caller.gd")
	serviceA := findNode(records, "function", "run", "game_a/scripts/service.gd")
	serviceB := findNode(records, "function", "run", "game_b/scripts/service.gd")
	baseA := findNode(records, "module", "base", "game_a/scripts/base.gd")
	callerModuleA := findNode(records, "module", "caller", "game_a/scripts/caller.gd")
	if callerA == nil || callerB == nil || serviceA == nil || serviceB == nil || baseA == nil || callerModuleA == nil {
		t.Fatalf("missing nested project facts callerA=%#v callerB=%#v serviceA=%#v serviceB=%#v base=%#v module=%#v", callerA, callerB, serviceA, serviceB, baseA, callerModuleA)
	}
	callerAID := callerA["id"].(string)
	callerBID := callerB["id"].(string)
	serviceAID := serviceA["id"].(string)
	serviceBID := serviceB["id"].(string)
	if !hasEdgeFromTo(records, "calls", callerAID, serviceAID) || hasEdgeFromTo(records, "calls", callerAID, serviceBID) {
		t.Fatalf("game_a autoload or preload resolved outside its project")
	}
	if !hasEdgeFromTo(records, "calls", callerBID, serviceBID) || hasEdgeFromTo(records, "calls", callerBID, serviceAID) {
		t.Fatalf("game_b autoload resolved outside its project")
	}
	if !hasEdgeFromTo(records, "extends", callerModuleA["id"].(string), baseA["id"].(string)) {
		t.Fatalf("nested res:// extends did not resolve")
	}
	for _, record := range records {
		if record["record"] != "unresolved" || record["relation"] != "calls" {
			continue
		}
		expression, _ := record["expression"].(string)
		if expression == "ServiceScript.new()" || expression == "instance.run()" || expression == "Service.run()" {
			t.Fatalf("nested project call remained unresolved: %#v", record)
		}
	}
}

func TestAnalyzeFilePreloadShadowsSameNamedAutoload(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "project.godot", `[autoload]
Service="*res://autoload_service.gd"
`)
	writeFixture(t, root, "autoload_service.gd", `func run():
    pass
`)
	writeFixture(t, root, "local_service.gd", `func run():
    pass
`)
	writeFixture(t, root, "caller.gd", `const Service := preload("res://local_service.gd")
func call_service():
    var instance := Service.new()
    instance.run()
`)

	records := analyzeFixture(t, root)
	caller := findNode(records, "function", "call_service", "caller.gd")
	localRun := findNode(records, "function", "run", "local_service.gd")
	autoloadRun := findNode(records, "function", "run", "autoload_service.gd")
	if caller == nil || localRun == nil || autoloadRun == nil {
		t.Fatalf("missing shadowing fixture nodes")
	}
	callerID := caller["id"].(string)
	if !hasEdgeFromTo(records, "calls", callerID, localRun["id"].(string)) {
		t.Fatalf("file-local preload did not shadow same-named autoload")
	}
	if hasEdgeFromTo(records, "calls", callerID, autoloadRun["id"].(string)) {
		t.Fatalf("same-named autoload incorrectly won over file-local preload")
	}
	if hasUnresolved(records, "calls", "Service.new", "external-target") {
		t.Fatalf("shadowed preload constructor remained external")
	}
}

func TestAnalyzeResolvesNestedTypeAliasesFromPreloadedScripts(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "base.gd", `func configure():
    pass
`)
	writeFixture(t, root, "support.gd", `const Base := preload("res://base.gd")
class Nested extends Base:
    pass
`)
	writeFixture(t, root, "caller.gd", `const Support := preload("res://support.gd")
const Alias := Support.Nested
func run():
    var value := Alias.new()
    value.configure()
`)

	records := analyzeFixture(t, root)
	run := findNode(records, "function", "run", "caller.gd")
	configure := findNode(records, "function", "configure", "base.gd")
	nested := findNode(records, "type", "Nested", "support.gd")
	if run == nil || configure == nil || nested == nil {
		t.Fatalf("missing nested preload alias fixture nodes")
	}
	if !hasEdgeFromTo(records, "calls", run["id"].(string), nested["id"].(string)) {
		t.Fatalf("nested preload alias constructor did not resolve")
	}
	if !hasEdgeFromTo(records, "calls", run["id"].(string), configure["id"].(string)) {
		t.Fatalf("nested preload alias instance method did not resolve")
	}
}

func TestAnalyzeResolvesMembersAfterChainedCallResults(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "chain.gd", `class Tracker:
    func needs_resync():
        pass
class Router:
    var tracker: Tracker
class Pipeline:
    func get_router() -> Router:
        return Router.new()
func run():
    var pipeline := Pipeline.new()
    pipeline.get_router().tracker.needs_resync()
`)

	records := analyzeFixture(t, root)
	run := findNode(records, "function", "run", "chain.gd")
	getRouter := findNode(records, "function", "get_router", "chain.gd")
	needsResync := findNode(records, "function", "needs_resync", "chain.gd")
	if run == nil || getRouter == nil || needsResync == nil {
		t.Fatalf("missing chained receiver fixture nodes")
	}
	for _, target := range []map[string]any{getRouter, needsResync} {
		if !hasEdgeFromTo(records, "calls", run["id"].(string), target["id"].(string)) {
			t.Errorf("missing chained call edge to %s", target["name"])
		}
	}
}

func TestAnalyzeKeepsImmediateReceiverAcrossExpressionBoundaries(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "service.gd", `class_name Service
static func static_ready() -> bool:
    return true
func ready() -> bool:
    return true
func stop():
    pass
`)
	writeFixture(t, root, "caller.gd", `func run():
    var service := Service.new()
    if !service.ready():
        return
    var ready := 1 + Service.static_ready()
    service.ready(); service.stop()
`)

	records := analyzeFixture(t, root)
	run := findNode(records, "function", "run", "caller.gd")
	ready := findNode(records, "function", "ready", "service.gd")
	staticReady := findNode(records, "function", "static_ready", "service.gd")
	stop := findNode(records, "function", "stop", "service.gd")
	if run == nil || ready == nil || staticReady == nil || stop == nil {
		t.Fatalf("missing receiver boundary fixture nodes")
	}
	for _, target := range []map[string]any{ready, staticReady, stop} {
		if !hasEdgeFromTo(records, "calls", run["id"].(string), target["id"].(string)) {
			t.Errorf("missing call edge to %s", target["name"])
		}
	}
}
