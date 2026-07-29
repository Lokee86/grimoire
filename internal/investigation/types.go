package investigation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const ledgerVersion = 1

var (
	ErrCorrupt          = errors.New("investigation ledger is corrupt")
	ErrSnapshotMismatch = errors.New("investigation snapshot mismatch")
	ErrSessionClosed    = errors.New("investigation session is closed")
	ErrSessionExists    = errors.New("investigation session already exists")
	ErrSessionNotFound  = errors.New("investigation session not found")
	ErrSessionLocked    = errors.New("investigation session is locked")
)

type Snapshot struct {
	Repository string            `json:"repository"`
	Providers  map[string]string `json:"providers,omitempty"`
}

func (s Snapshot) Validate() error {
	if strings.TrimSpace(s.Repository) == "" {
		return errors.New("repository snapshot is required")
	}
	for provider, snapshot := range s.Providers {
		if strings.TrimSpace(provider) == "" || strings.TrimSpace(snapshot) == "" {
			return errors.New("provider names and snapshots are required")
		}
	}
	return nil
}

func (s Snapshot) normalized() Snapshot {
	copy := Snapshot{Repository: strings.TrimSpace(s.Repository), Providers: make(map[string]string, len(s.Providers))}
	for provider, snapshot := range s.Providers {
		copy.Providers[strings.TrimSpace(provider)] = strings.TrimSpace(snapshot)
	}
	if len(copy.Providers) == 0 {
		copy.Providers = nil
	}
	return copy
}

func (s Snapshot) Equal(other Snapshot) bool {
	left, right := s.normalized(), other.normalized()
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return string(leftBytes) == string(rightBytes)
}

func (s Snapshot) Digest() string {
	data, _ := json.Marshal(s.normalized())
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

type Node struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind,omitempty"`
	Label    string            `json:"label,omitempty"`
	Path     string            `json:"path,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type SourceRange struct {
	Path        string            `json:"path"`
	StartLine   int               `json:"start_line"`
	StartColumn int               `json:"start_column,omitempty"`
	EndLine     int               `json:"end_line"`
	EndColumn   int               `json:"end_column,omitempty"`
	Text        string            `json:"text,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type GraphPath struct {
	ID       string            `json:"id,omitempty"`
	Nodes    []string          `json:"nodes"`
	Edges    []string          `json:"edges,omitempty"`
	Label    string            `json:"label,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Document struct {
	ID       string            `json:"id"`
	URI      string            `json:"uri,omitempty"`
	Title    string            `json:"title,omitempty"`
	Content  string            `json:"content,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type EvidenceRef struct {
	Kind  string `json:"kind"`
	Index int    `json:"index"`
}

type RetrievalSeed struct {
	Evidence EvidenceRef `json:"evidence"`
	Lane     string      `json:"lane,omitempty"`
	Provider string      `json:"provider,omitempty"`
	Rank     int         `json:"rank,omitempty"`
	Score    float64     `json:"score,omitempty"`
	Reasons  []string    `json:"reasons,omitempty"`
}

type RetrievalHit struct {
	Evidence        EvidenceRef    `json:"evidence"`
	RelatedEvidence []EvidenceRef  `json:"related_evidence,omitempty"`
	Lane            string         `json:"lane"`
	Provider        string         `json:"provider,omitempty"`
	Rank            int            `json:"rank,omitempty"`
	Score           float64        `json:"score,omitempty"`
	Reasons         []string       `json:"reasons,omitempty"`
	DuplicateOf     string         `json:"duplicate_of,omitempty"`
	Direction       string         `json:"direction,omitempty"`
	Relation        string         `json:"relation,omitempty"`
	Certainty       string         `json:"certainty,omitempty"`
	Depth           int            `json:"depth,omitempty"`
	Support         []string       `json:"support,omitempty"`
	Seed            *RetrievalSeed `json:"seed,omitempty"`
}

type UnresolvedQuestion struct {
	Question string `json:"question"`
	Context  string `json:"context,omitempty"`
}

type Branch struct {
	ID          string   `json:"id"`
	Description string   `json:"description,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	References  []string `json:"references,omitempty"`
}

type Response struct {
	Snapshot            Snapshot             `json:"snapshot"`
	Nodes               []Node               `json:"nodes,omitempty"`
	SourceRanges        []SourceRange        `json:"source_ranges,omitempty"`
	GraphPaths          []GraphPath          `json:"graph_paths,omitempty"`
	Documents           []Document           `json:"documents,omitempty"`
	RetrievalHits       []RetrievalHit       `json:"retrieval_hits,omitempty"`
	UnresolvedQuestions []UnresolvedQuestion `json:"unresolved_questions,omitempty"`
	RejectedBranches    []Branch             `json:"rejected_branches,omitempty"`
	AcceptedBranches    []Branch             `json:"accepted_branches,omitempty"`
}

type NodeHandle struct{ token string }
type SourceRangeHandle struct{ token string }
type GraphPathHandle struct{ token string }
type DocumentHandle struct{ token string }

func (h NodeHandle) String() string                      { return h.token }
func (h SourceRangeHandle) String() string               { return h.token }
func (h GraphPathHandle) String() string                 { return h.token }
func (h DocumentHandle) String() string                  { return h.token }
func (h NodeHandle) IsZero() bool                        { return h.token == "" }
func (h SourceRangeHandle) IsZero() bool                 { return h.token == "" }
func (h GraphPathHandle) IsZero() bool                   { return h.token == "" }
func (h DocumentHandle) IsZero() bool                    { return h.token == "" }
func (h NodeHandle) MarshalJSON() ([]byte, error)        { return json.Marshal(h.token) }
func (h SourceRangeHandle) MarshalJSON() ([]byte, error) { return json.Marshal(h.token) }
func (h GraphPathHandle) MarshalJSON() ([]byte, error)   { return json.Marshal(h.token) }
func (h DocumentHandle) MarshalJSON() ([]byte, error)    { return json.Marshal(h.token) }

func (h *NodeHandle) UnmarshalJSON(data []byte) error {
	return decodeJSONHandle(data, "node", &h.token)
}
func (h *SourceRangeHandle) UnmarshalJSON(data []byte) error {
	return decodeJSONHandle(data, "source", &h.token)
}
func (h *GraphPathHandle) UnmarshalJSON(data []byte) error {
	return decodeJSONHandle(data, "path", &h.token)
}
func (h *DocumentHandle) UnmarshalJSON(data []byte) error {
	return decodeJSONHandle(data, "document", &h.token)
}

func decodeJSONHandle(data []byte, kind string, target *string) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if _, err := decodeHandle(value, kind); err != nil {
		return err
	}
	*target = value
	return nil
}

type NodeRecord struct {
	Handle   NodeHandle `json:"handle"`
	Evidence Node       `json:"evidence"`
}
type SourceRangeRecord struct {
	Handle   SourceRangeHandle `json:"handle"`
	Evidence SourceRange       `json:"evidence"`
}
type GraphPathRecord struct {
	Handle   GraphPathHandle `json:"handle"`
	Evidence GraphPath       `json:"evidence"`
}
type DocumentRecord struct {
	Handle   DocumentHandle `json:"handle"`
	Evidence Document       `json:"evidence"`
}

type RetrievalSeedRecord struct {
	EvidenceHandle string   `json:"evidence_handle"`
	EvidenceKind   string   `json:"evidence_kind"`
	Lane           string   `json:"lane,omitempty"`
	Provider       string   `json:"provider,omitempty"`
	Rank           int      `json:"rank,omitempty"`
	Score          float64  `json:"score,omitempty"`
	Reasons        []string `json:"reasons,omitempty"`
}

type EvidenceHandleRef struct {
	Handle string `json:"handle"`
	Kind   string `json:"kind"`
}

type RetrievalHitRecord struct {
	EvidenceHandle  string               `json:"evidence_handle"`
	EvidenceKind    string               `json:"evidence_kind"`
	RelatedEvidence []EvidenceHandleRef  `json:"related_evidence,omitempty"`
	Lane            string               `json:"lane"`
	Provider        string               `json:"provider,omitempty"`
	Rank            int                  `json:"rank,omitempty"`
	Score           float64              `json:"score,omitempty"`
	Reasons         []string             `json:"reasons,omitempty"`
	DuplicateOf     string               `json:"duplicate_of,omitempty"`
	Direction       string               `json:"direction,omitempty"`
	Relation        string               `json:"relation,omitempty"`
	Certainty       string               `json:"certainty,omitempty"`
	Depth           int                  `json:"depth,omitempty"`
	Support         []string             `json:"support,omitempty"`
	Seed            *RetrievalSeedRecord `json:"seed,omitempty"`
}

type Delta struct {
	ResponseID string `json:"response_id"`

	NewNodes          []NodeRecord         `json:"new_nodes,omitempty"`
	PriorNodeHandles  []NodeHandle         `json:"prior_node_handles,omitempty"`
	NewSourceRanges   []SourceRangeRecord  `json:"new_source_ranges,omitempty"`
	PriorSourceRanges []SourceRangeHandle  `json:"prior_source_range_handles,omitempty"`
	NewGraphPaths     []GraphPathRecord    `json:"new_graph_paths,omitempty"`
	PriorGraphPaths   []GraphPathHandle    `json:"prior_graph_path_handles,omitempty"`
	NewDocuments      []DocumentRecord     `json:"new_documents,omitempty"`
	PriorDocuments    []DocumentHandle     `json:"prior_document_handles,omitempty"`
	RetrievalHits     []RetrievalHitRecord `json:"retrieval_hits,omitempty"`

	NewQuestions        []UnresolvedQuestion `json:"new_unresolved_questions,omitempty"`
	PriorQuestionIDs    []string             `json:"prior_unresolved_question_ids,omitempty"`
	NewRejectedBranches []Branch             `json:"new_rejected_branches,omitempty"`
	PriorRejectedIDs    []string             `json:"prior_rejected_branch_ids,omitempty"`
	NewAcceptedBranches []Branch             `json:"new_accepted_branches,omitempty"`
	PriorAcceptedIDs    []string             `json:"prior_accepted_branch_ids,omitempty"`
}

type Status struct {
	Version            int      `json:"version"`
	SessionID          string   `json:"session_id"`
	Snapshot           Snapshot `json:"snapshot"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	ClosedAt           string   `json:"closed_at,omitempty"`
	Responses          int      `json:"responses"`
	UniqueNodes        int      `json:"unique_nodes"`
	UniqueSourceRanges int      `json:"unique_source_ranges"`
	UniqueGraphPaths   int      `json:"unique_graph_paths"`
	UniqueDocuments    int      `json:"unique_documents"`
	UniqueQuestions    int      `json:"unique_questions"`
	UniqueRejected     int      `json:"unique_rejected_branches"`
	UniqueAccepted     int      `json:"unique_accepted_branches"`
}

func validateNode(node Node) error {
	if strings.TrimSpace(node.ID) == "" {
		return errors.New("node id is required")
	}
	return nil
}
func validateSourceRange(value SourceRange) error {
	if strings.TrimSpace(value.Path) == "" || value.StartLine <= 0 || value.EndLine < value.StartLine {
		return errors.New("source range requires a path and valid line range")
	}
	return nil
}
func validateGraphPath(value GraphPath) error {
	if len(value.Nodes) == 0 {
		return errors.New("graph path requires at least one node")
	}
	for _, node := range value.Nodes {
		if strings.TrimSpace(node) == "" {
			return errors.New("graph path nodes must not be empty")
		}
	}
	return nil
}
func validateDocument(value Document) error {
	if strings.TrimSpace(value.ID) == "" {
		return errors.New("document id is required")
	}
	return nil
}
func validateEvidenceRef(response Response, value EvidenceRef) error {
	if value.Index < 0 {
		return errors.New("evidence reference index must not be negative")
	}
	switch value.Kind {
	case "node":
		if value.Index >= len(response.Nodes) {
			return errors.New("node evidence reference is out of range")
		}
	case "source":
		if value.Index >= len(response.SourceRanges) {
			return errors.New("source evidence reference is out of range")
		}
	case "path":
		if value.Index >= len(response.GraphPaths) {
			return errors.New("path evidence reference is out of range")
		}
	case "document":
		if value.Index >= len(response.Documents) {
			return errors.New("document evidence reference is out of range")
		}
	default:
		return fmt.Errorf("unsupported evidence reference kind %q", value.Kind)
	}
	return nil
}

func validateRetrievalHit(response Response, value RetrievalHit) error {
	if strings.TrimSpace(value.Lane) == "" {
		return errors.New("retrieval hit lane is required")
	}
	if err := validateEvidenceRef(response, value.Evidence); err != nil {
		return err
	}
	for _, related := range value.RelatedEvidence {
		if err := validateEvidenceRef(response, related); err != nil {
			return fmt.Errorf("related evidence: %w", err)
		}
	}
	if value.Seed != nil {
		if err := validateEvidenceRef(response, value.Seed.Evidence); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
	}
	return nil
}

func validateQuestion(value UnresolvedQuestion) error {
	if strings.TrimSpace(value.Question) == "" {
		return errors.New("unresolved question is required")
	}
	return nil
}
func validateBranch(value Branch) error {
	if strings.TrimSpace(value.ID) == "" {
		return errors.New("branch id is required")
	}
	return nil
}

func responseValidate(response Response) error {
	if err := response.Snapshot.Validate(); err != nil {
		return err
	}
	for _, value := range response.Nodes {
		if err := validateNode(value); err != nil {
			return fmt.Errorf("node: %w", err)
		}
	}
	for _, value := range response.SourceRanges {
		if err := validateSourceRange(value); err != nil {
			return fmt.Errorf("source range: %w", err)
		}
	}
	for _, value := range response.GraphPaths {
		if err := validateGraphPath(value); err != nil {
			return fmt.Errorf("graph path: %w", err)
		}
	}
	for _, value := range response.Documents {
		if err := validateDocument(value); err != nil {
			return fmt.Errorf("document: %w", err)
		}
	}
	for _, value := range response.RetrievalHits {
		if err := validateRetrievalHit(response, value); err != nil {
			return fmt.Errorf("retrieval hit: %w", err)
		}
	}
	for _, value := range response.UnresolvedQuestions {
		if err := validateQuestion(value); err != nil {
			return fmt.Errorf("question: %w", err)
		}
	}
	for _, value := range response.RejectedBranches {
		if err := validateBranch(value); err != nil {
			return fmt.Errorf("rejected branch: %w", err)
		}
	}
	for _, value := range response.AcceptedBranches {
		if err := validateBranch(value); err != nil {
			return fmt.Errorf("accepted branch: %w", err)
		}
	}
	return nil
}

func sortedStrings(values []string) []string {
	copy := append([]string(nil), values...)
	sort.Strings(copy)
	return copy
}
