package embedding

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadVerifiedPublishesMatchingContent(t *testing.T) {
	content := []byte("verified model data")
	sum := sha256.Sum256(content)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(content)
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "model.gguf")
	if err := downloadVerified(
		context.Background(), server.Client(), server.URL,
		hex.EncodeToString(sum[:]), target, "test model", nilWriter{},
	); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("unexpected content %q", got)
	}
	if _, err := os.Stat(target + ".partial"); !os.IsNotExist(err) {
		t.Fatalf("partial file remains: %v", err)
	}
}

func TestDownloadVerifiedRejectsHashMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("wrong data"))
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "model.gguf")
	err := downloadVerified(
		context.Background(), server.Client(), server.URL,
		strings.Repeat("0", 64), target, "test model", nilWriter{},
	)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("mismatched file was published: %v", statErr)
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "unsafe.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(archiveFile)
	entry, err := archive.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("escape"))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "runtime")
	err = extractZip(archivePath, target)
	if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type nilWriter struct{}

func (nilWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
