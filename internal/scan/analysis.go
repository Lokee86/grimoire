package scan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/library"
	analysisscope "github.com/Lokee86/lexicon/internal/scope"
)

func (s *Scanner) analyzeFull(ctx context.Context, languages []string) error {
	plans := make([]analysisPlan, 0, len(languages))
	for _, language := range languages {
		plans = append(plans, analysisPlan{Language: language, Full: true})
	}
	return s.analyzePlans(ctx, plans)
}

func (s *Scanner) analyzePlans(ctx context.Context, plans []analysisPlan) error {
	libraryRoot := filepath.Join(s.StateRoot, "library")
	temporary := filepath.Join(config.StateRoot(s.Repository), "tmp")
	if err := os.MkdirAll(temporary, 0o755); err != nil {
		return err
	}
	for _, plan := range plans {
		if err := s.analyzePlan(ctx, plan, libraryRoot, temporary); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scanner) analyzePlan(ctx context.Context, plan analysisPlan, libraryRoot, temporary string) error {
	output := filepath.Join(libraryRoot, plan.Language+".jsonl")
	sourceRoot := filepath.Join(s.StateRoot, "source")
	present, err := hasLanguage(sourceRoot, plan.Language)
	if err != nil {
		return err
	}
	if !present {
		if s.Output != nil {
			fmt.Fprintf(s.Output, "removing %s library\n", plan.Language)
		}
		if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	adapterOutput := filepath.Join(temporary, plan.Language+".jsonl")
	_ = os.Remove(adapterOutput)
	if s.Output != nil {
		if plan.Full {
			fmt.Fprintf(s.Output, "analyzing %s\n", plan.Language)
		} else {
			fmt.Fprintf(s.Output, "analyzing %s files: %d\n", plan.Language, len(plan.ChangedFiles))
		}
	}
	request, err := s.analysisRequest(plan, sourceRoot, temporary, adapterOutput)
	if err != nil {
		return err
	}
	if err := s.Analyzer.Run(ctx, request); err != nil {
		if plan.Full {
			return err
		}
		return s.retryFull(ctx, request, sourceRoot, output, err)
	}
	if plan.Full {
		return replace(adapterOutput, output)
	}
	fullRequired, err := s.Store.RequiresFullAnalysis(plan.Language, plan.ChangedFiles, adapterOutput)
	if err != nil {
		return err
	}
	if fullRequired {
		return s.retryFull(ctx, request, sourceRoot, output, nil)
	}
	if err := library.SetSharedComplete(adapterOutput, false); err != nil {
		return err
	}
	merged := filepath.Join(temporary, plan.Language+".merged.jsonl")
	_ = os.Remove(merged)
	if err := library.Merge(output, adapterOutput, merged); err != nil {
		return err
	}
	return replace(merged, output)
}

func (s *Scanner) analysisRequest(plan analysisPlan, sourceRoot, temporary, output string) (adapters.Request, error) {
	repository := sourceRoot
	if !plan.Full {
		var err error
		repository, err = analysisscope.Build(sourceRoot, filepath.Join(temporary, "scopes"), plan.Language, plan.ContextFiles)
		if err != nil {
			return adapters.Request{}, err
		}
	}
	return adapters.Request{
		Language: plan.Language, Repository: repository, Output: output,
		ChangedFiles: plan.ChangedFiles, RemovedFiles: plan.RemovedFiles,
	}, nil
}

func (s *Scanner) retryFull(ctx context.Context, request adapters.Request, sourceRoot, output string, scopedErr error) error {
	if s.Output != nil {
		fmt.Fprintf(s.Output, "expanding %s to full analysis\n", request.Language)
	}
	_ = os.Remove(request.Output)
	request.Repository = sourceRoot
	request.ChangedFiles = nil
	request.RemovedFiles = nil
	if err := s.Analyzer.Run(ctx, request); err != nil {
		if scopedErr != nil {
			return fmt.Errorf("scoped %s analysis failed: %v; full retry failed: %w", request.Language, scopedErr, err)
		}
		return err
	}
	return replace(request.Output, output)
}

func planLanguages(plans []analysisPlan) []string {
	languages := make([]string, 0, len(plans))
	for _, plan := range plans {
		languages = append(languages, plan.Language)
	}
	return languages
}

func replace(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(source, destination)
}
