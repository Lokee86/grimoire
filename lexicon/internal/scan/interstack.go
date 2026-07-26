package scan

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/interstack"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func (s *Scanner) refreshInterstack(manifest objectstore.Manifest) (objectstore.Manifest, error) {
	ordinaryLanguages := 0
	for _, entry := range manifest.Languages {
		if entry.Language != interstack.Language {
			ordinaryLanguages++
		}
	}
	if ordinaryLanguages == 0 {
		return manifest.WithoutLanguage(interstack.Language), nil
	}
	output := filepath.Join(s.StateRoot, "tmp", interstack.Language+".jsonl")
	analysis, summary, err := interstack.Build(
		filepath.Join(s.StateRoot, "source"),
		s.Store,
		manifest,
		output,
	)
	if err != nil {
		return objectstore.Manifest{}, fmt.Errorf("build interstack analysis: %w", err)
	}
	entry, err := s.Store.BuildSharedLanguage(analysis, config.AnalysisID(), interstack.AdapterFingerprint())
	if err != nil {
		return objectstore.Manifest{}, fmt.Errorf("store interstack analysis: %w", err)
	}
	s.writeOutput(
		"interstack: %d HTTP contracts, %d HTTP links, %d message channels, %d message links, %d config keys\n",
		summary.HTTPContracts,
		summary.HTTPLinks,
		summary.MessageChannels,
		summary.MessageLinks,
		summary.ConfigKeys,
	)
	return manifest.WithLanguage(entry), nil
}

func interstackDrifted(manifest objectstore.Manifest) bool {
	ordinaryLanguages := 0
	var derived *objectstore.LanguageEntry
	for index := range manifest.Languages {
		entry := &manifest.Languages[index]
		if entry.Language == interstack.Language {
			derived = entry
			continue
		}
		ordinaryLanguages++
	}
	if ordinaryLanguages == 0 {
		return derived != nil
	}
	if derived == nil {
		return true
	}
	return derived.AdapterVersion != interstack.AdapterVersion ||
		derived.AdapterFingerprint != interstack.AdapterFingerprint() || derived.SchemaVersion != 1
}
