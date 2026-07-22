package embedding

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadVerified(
	ctx context.Context,
	client *http.Client,
	url string,
	expectedSHA string,
	target string,
	label string,
	progress io.Writer,
) error {
	if matchesSHA(target, expectedSHA) {
		_, _ = fmt.Fprintf(progress, "%s already installed: %s\n", label, target)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	partial := target + ".partial"
	if err := os.Remove(partial); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	_, _ = fmt.Fprintf(progress, "downloading %s\n", label)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download %s: %w", label, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("download %s: server returned %s", label, response.Status)
	}

	file, err := os.OpenFile(partial, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(file, hash), response.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(partial)
		return fmt.Errorf("download %s: %w", label, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(partial)
		return closeErr
	}
	actualSHA := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualSHA, expectedSHA) {
		_ = os.Remove(partial)
		return fmt.Errorf("verify %s: sha256 %s does not match %s", label, actualSHA, expectedSHA)
	}
	if err := os.Rename(partial, target); err != nil {
		_ = os.Remove(partial)
		return err
	}
	_, _ = fmt.Fprintf(progress, "installed %s: %s\n", label, target)
	return nil
}

func matchesSHA(path, expected string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}
	return strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expected)
}

func extractZip(archivePath, targetDir string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	cleanRoot := filepath.Clean(targetDir) + string(os.PathSeparator)
	for _, entry := range archive.File {
		target := filepath.Join(targetDir, filepath.FromSlash(entry.Name))
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanRoot) {
			return fmt.Errorf("unsafe archive path %q", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := extractFile(entry, target); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(entry *zip.File, target string) error {
	source, err := entry.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entry.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func findNamedFile(root, name string) (string, error) {
	var result string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(entry.Name(), name) {
			result = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", os.ErrNotExist
	}
	return result, nil
}
