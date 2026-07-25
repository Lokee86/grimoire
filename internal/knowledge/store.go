package knowledge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var ErrIncompatibleIndex = errors.New("incompatible knowledge index")

func DefaultState(root string) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve knowledge root: %w", err)
	}
	return filepath.Join(absolute, ".grimoire", "knowledge"), nil
}

func Load(state string) (Index, error) {
	data, err := os.ReadFile(filepath.Join(state, "index.json"))
	if err != nil {
		return Index{}, err
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return Index{}, fmt.Errorf("decode knowledge index: %w", err)
	}
	if index.Version != FormatVersion {
		return Index{}, fmt.Errorf("%w: version %d", ErrIncompatibleIndex, index.Version)
	}
	if err := validate(index); err != nil {
		return Index{}, err
	}
	return index, nil
}

func Save(state string, index Index) error {
	if index.Version != FormatVersion {
		return fmt.Errorf("cannot save knowledge index version %d", index.Version)
	}
	if err := validate(index); err != nil {
		return err
	}
	sort.Slice(index.Documents, func(i, j int) bool { return index.Documents[i].Path < index.Documents[j].Path })
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("encode knowledge index: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(state, 0o755); err != nil {
		return fmt.Errorf("create knowledge state: %w", err)
	}
	temporary, err := os.CreateTemp(state, "index-*.tmp")
	if err != nil {
		return fmt.Errorf("create knowledge state temp file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write knowledge state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, filepath.Join(state, "index.json")); err != nil {
		return fmt.Errorf("publish knowledge state: %w", err)
	}
	return nil
}

func Inspect(index Index, path, handle string) (any, error) {
	if path == "" && handle == "" {
		sections := 0
		for _, document := range index.Documents {
			sections += len(document.Sections)
		}
		return struct {
			Version   int    `json:"version"`
			Root      string `json:"root"`
			GitCommit string `json:"git_commit,omitempty"`
			Documents int    `json:"documents"`
			Sections  int    `json:"sections"`
		}{index.Version, index.Root, index.GitCommit, len(index.Documents), sections}, nil
	}
	for _, document := range index.Documents {
		if path != "" && document.Path != path {
			continue
		}
		for _, section := range document.Sections {
			if handle == "" || document.Handle(section) == handle {
				return Result{Handle: document.Handle(section), DocumentID: document.ID, SectionID: section.ID, Path: document.Path, Kind: document.Kind, Heading: section.Heading, HeadingPath: section.HeadingPath, StartByte: section.StartByte, EndByte: section.EndByte, StartLine: section.StartLine, EndLine: section.EndLine, Hash: section.Hash, Text: section.Text, CommitID: document.CommitID, CommitTime: document.CommitTime, CodeLinks: section.CodeLinks}, nil
			}
		}
	}
	return nil, fmt.Errorf("knowledge section not found")
}

func validate(index Index) error {
	if index.Version != FormatVersion {
		return fmt.Errorf("%w: version %d", ErrIncompatibleIndex, index.Version)
	}
	previous := ""
	for _, document := range index.Documents {
		if document.ID == "" || document.Path == "" || document.Hash == "" {
			return fmt.Errorf("invalid knowledge document %q", document.Path)
		}
		if previous != "" && document.Path < previous {
			return fmt.Errorf("knowledge documents are not sorted")
		}
		previous = document.Path
		if len(document.Sections) == 0 {
			return fmt.Errorf("knowledge document %q has no sections", document.Path)
		}
		for _, section := range document.Sections {
			if section.ID == "" || section.StartByte < 0 || section.EndByte < section.StartByte || section.EndByte > int(document.Size) {
				return fmt.Errorf("invalid knowledge section %q", section.ID)
			}
		}
	}
	return nil
}
