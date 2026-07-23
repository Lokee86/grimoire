package embedding

import (
	"reflect"
	"testing"
)

func TestServeArgsForLlamaServer(t *testing.T) {
	got := ServeArgs("llama-server.exe", ServeOptions{
		Host: "127.0.0.1", Port: 9090, ContextSize: 4096, UbatchSize: 1024, Parallel: 8,
	})
	want := []string{
		"-hf", ModelReference,
		"--embedding",
		"--pooling", "last",
		"--host", "127.0.0.1",
		"--port", "9090",
		"--ctx-size", "4096",
		"--ubatch-size", "1024",
		"--parallel", "8",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected args:\n got: %v\nwant: %v", got, want)
	}
}

func TestServeArgsForLlamaMulticallAndLocalModel(t *testing.T) {
	got := ServeArgs("llama", ServeOptions{ModelPath: "model.gguf"})
	if len(got) < 4 || got[0] != "serve" || got[1] != "-m" || got[2] != "model.gguf" {
		t.Fatalf("unexpected args: %v", got)
	}
	for index := range got[:len(got)-1] {
		if got[index] == "--port" && got[index+1] == "9876" {
			return
		}
	}
	t.Fatalf("default port missing from args: %v", got)
}
