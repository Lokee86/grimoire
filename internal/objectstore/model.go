package objectstore

import "encoding/json"

const (
	ObjectVersion   = 1
	SnapshotVersion = 1
)

type FactObject struct {
	Version          int               `json:"version"`
	Language         string            `json:"language"`
	Owner            string            `json:"owner,omitempty"`
	SourceContentID  string            `json:"source_content_id,omitempty"`
	AdapterVersion   string            `json:"adapter_version"`
	SchemaVersion    int               `json:"schema_version"`
	AnalysisConfigID string            `json:"analysis_config_id"`
	Records          []json.RawMessage `json:"records"`
}

type FileEntry struct {
	Path      string `json:"path"`
	Language  string `json:"language"`
	ContentID string `json:"content_id"`
	ObjectID  string `json:"object_id"`
}

type LanguageEntry struct {
	Language         string      `json:"language"`
	AdapterVersion   string      `json:"adapter_version"`
	SchemaVersion    int         `json:"schema_version"`
	Repository       string      `json:"repository"`
	AnalysisConfigID string      `json:"analysis_config_id"`
	SharedObjectID   string      `json:"shared_object_id,omitempty"`
	Files            []FileEntry `json:"files"`
}

type Manifest struct {
	Version     int             `json:"version"`
	StateCommit string          `json:"state_commit"`
	Languages   []LanguageEntry `json:"languages"`
}

type Header struct {
	Record         string `json:"record"`
	SchemaVersion  int    `json:"schema_version"`
	AdapterVersion string `json:"adapter_version"`
	Language       string `json:"language"`
	Repository     string `json:"repository"`
	Mode           string `json:"mode"`
}
