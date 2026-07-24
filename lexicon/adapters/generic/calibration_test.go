package main

import (
	"reflect"
	"testing"
)

func TestGenericAdapterCalibrationCorpus(t *testing.T) {
	tests := []struct {
		language string
		file     string
		content  string
		nodes    [][2]string
	}{
		{"generic-c", "main.c", `/*
struct Phantom {};
int ghost() {}
*/
#include <stdio.h>
struct Player {};
int launch() { return 0; }
const char *example = "class Specter";
`, [][2]string{{"import", "stdio.h"}, {"type", "Player"}, {"function", "launch"}}},
		{"generic-cpp", "fleet.cpp", `#include <vector>
class Fleet {};
int Fleet::launch(int x) { return x; }
`, [][2]string{{"import", "vector"}, {"type", "Fleet"}, {"function", "launch"}}},
		{"generic-java", "Clock.java", `import java.time.Instant;
interface Clock {
  default long now() { return 0; }
}
`, [][2]string{{"import", "java.time.Instant"}, {"interface", "Clock"}, {"function", "now"}}},
		{"generic-kt", "Worker.kt", `import kotlin.time.Duration
data class Worker(val id: Int) {
  fun run() {}
}
`, [][2]string{{"import", "kotlin.time.Duration"}, {"type", "Worker"}, {"function", "run"}}},
		{"generic-swift", "Runner.swift", `import Foundation
public struct Runner {
  public func execute() {}
}
`, [][2]string{{"import", "Foundation"}, {"type", "Runner"}, {"function", "execute"}}},
		{"generic-php", "Controller.php", `<?php
use App\Core;
namespace App;
final class Controller {
  public function handle() {}
}
`, [][2]string{{"import", `App\Core`}, {"namespace", "App"}, {"type", "Controller"}, {"function", "handle"}}},
		{"generic-cs", "Service.cs", `using System.Threading.Tasks;
public sealed record Service {
  public async Task RunAsync() => await Task.Yield();
}
`, [][2]string{{"import", "System.Threading.Tasks"}, {"type", "Service"}, {"function", "RunAsync"}}},
		{"generic-lua", "service.lua", `local dep = require("dep")
function Service.run() end
`, [][2]string{{"import", "dep"}, {"function", "run"}}},
		{"generic-sh", "deploy.sh", `function deploy {
  echo ready
}
`, [][2]string{{"function", "deploy"}}},
		{"generic-sql", "schema.sql", `CREATE VIEW active_users AS SELECT 1;
CREATE TRIGGER refresh_cache AFTER INSERT ON users BEGIN SELECT 1; END;
`, [][2]string{{"type", "active_users"}, {"function", "refresh_cache"}}},
		{"generic-ps1", "build.ps1", `Import-Module "Build.Tools"
function Start-Build { }
`, [][2]string{{"import", "Build.Tools"}, {"function", "Start-Build"}}},
		{"generic-m", "Clock.m", `#import "Clock.h"
@protocol Clock
- (long)now;
@end
`, [][2]string{{"import", "Clock.h"}, {"interface", "Clock"}, {"function", "now"}}},
		{"generic-proto", "api.proto", `import weak "types.proto";
enum State { UNKNOWN = 0; }
service Jobs {
  rpc Start(State) returns (State);
}
`, [][2]string{{"import", "types.proto"}, {"type", "State"}, {"interface", "Jobs"}, {"function", "Start"}}},
		{"generic-sol", "Math.sol", `import "./Base.sol";
library Math {
  function add() internal {}
}
`, [][2]string{{"import", "./Base.sol"}, {"type", "Math"}, {"function", "add"}}},
		{"generic-pas", "worker.pas", `type TWorker = class
end;
procedure Run();
begin
end;
`, [][2]string{{"type", "TWorker"}, {"function", "Run"}}},
		{"generic-vb", "Service.vb", `Public Class Service
  Public Sub Run()
  End Sub
End Class
`, [][2]string{{"type", "Service"}, {"function", "Run"}}},
		{"generic-pl", "service.pl", "sub run { return 1; }\n", [][2]string{{"function", "run"}}},
		{"generic-r", "service.r", "run <- function(x) { x }\n", [][2]string{{"function", "run"}}},
	}

	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			repository := t.TempDir()
			writeFixture(t, repository, test.file, test.content)
			output, err := analyzeRepository(repository, test.language, nil, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			assertExactSemanticNodes(t, decodeRecords(t, output), test.nodes)
		})
	}
}

func TestGenericAdapterRecognizesAllmanFunctions(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "main.c", "int launch()\n{\n  return 0;\n}\n")
	output, err := analyzeRepository(repository, "generic-c", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertNode(t, decodeRecords(t, output), "function", "launch")
}

func assertExactSemanticNodes(t *testing.T, records []map[string]any, expected [][2]string) {
	t.Helper()
	actual := make(map[[2]string]struct{})
	for _, record := range records {
		kind, kindOK := record["kind"].(string)
		name, nameOK := record["name"].(string)
		if record["record"] == "node" && kindOK && nameOK && kind != "file" && kind != "module" {
			actual[[2]string{kind, name}] = struct{}{}
		}
	}
	wanted := make(map[[2]string]struct{}, len(expected))
	for _, node := range expected {
		wanted[node] = struct{}{}
	}
	if !reflect.DeepEqual(actual, wanted) {
		t.Fatalf("semantic nodes = %#v, want %#v", actual, wanted)
	}
}
