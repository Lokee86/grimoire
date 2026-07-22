package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
	"github.com/Lokee86/lexicon/internal/scan"
	"github.com/fsnotify/fsnotify"
)

type Options struct {
	Debounce  time.Duration
	Reconcile time.Duration
	Output    io.Writer
}

func Run(ctx context.Context, scanner *scan.Scanner, options Options) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	if err := addTree(watcher, scanner.Repository); err != nil {
		return err
	}
	if report, err := scanner.Scan(ctx); err != nil {
		return err
	} else {
		writeReport(options.Output, "startup", report)
	}
	if options.Debounce <= 0 {
		options.Debounce = 150 * time.Millisecond
	}
	if options.Reconcile <= 0 {
		options.Reconcile = 30 * time.Second
	}
	pending := make(map[string]struct{})
	var timer *time.Timer
	var timerChannel <-chan time.Time
	reconcile := time.NewTicker(options.Reconcile)
	defer reconcile.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if ignored(scanner.Repository, event.Name) {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addTree(watcher, event.Name)
				}
			}
			if relevantEvent(event) {
				pending[event.Name] = struct{}{}
				if timer == nil {
					timer = time.NewTimer(options.Debounce)
				} else if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(options.Debounce)
				timerChannel = timer.C
			}
		case <-timerChannel:
			paths := sortedKeys(pending)
			clear(pending)
			timerChannel = nil
			report, scanErr := scanner.ScanPaths(ctx, paths)
			if scanErr != nil {
				fmt.Fprintf(options.Output, "lexicon daemon scan failed: %v\n", scanErr)
				continue
			}
			writeReport(options.Output, "watch", report)
		case watchErr, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(options.Output, "lexicon watcher error: %v; reconciling\n", watchErr)
			report, scanErr := scanner.Scan(ctx)
			if scanErr != nil {
				fmt.Fprintf(options.Output, "lexicon reconciliation failed: %v\n", scanErr)
			} else {
				writeReport(options.Output, "reconcile", report)
			}
		case <-reconcile.C:
			report, scanErr := scanner.Scan(ctx)
			if scanErr != nil {
				fmt.Fprintf(options.Output, "lexicon reconciliation failed: %v\n", scanErr)
			} else {
				writeReport(options.Output, "reconcile", report)
			}
		}
	}
}

func addTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if path != root && lexfiles.IgnoredDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

func ignored(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return true
	}
	for _, part := range strings.Split(filepath.Clean(relative), string(filepath.Separator)) {
		if lexfiles.IgnoredDirectory(part) {
			return true
		}
	}
	return false
}

func relevantEvent(event fsnotify.Event) bool {
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		return true
	}
	if lexfiles.Relevant(event.Name) {
		return true
	}
	info, err := os.Stat(event.Name)
	return err == nil && info.IsDir()
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func writeReport(output io.Writer, source string, report scan.Report) {
	if output == nil || len(report.Changed) == 0 {
		return
	}
	fmt.Fprintf(output, "%s scan: %d files, %s\n", source, len(report.Changed), strings.Join(report.Languages, ", "))
}
