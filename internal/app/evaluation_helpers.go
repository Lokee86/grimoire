package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func durationMS(value time.Duration) float64 {
	return float64(value) / float64(time.Millisecond)
}

func validateExpectedSymbol(root, caseID, relativePath string, symbols ...string) error {
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("incorrect evaluation expectation for %s: read %s: %w", caseID, relativePath, err)
	}
	for _, symbol := range symbols {
		if !strings.Contains(string(data), symbol) {
			return fmt.Errorf("incorrect evaluation expectation for %s: symbol %s is absent from %s", caseID, symbol, relativePath)
		}
	}
	return nil
}
