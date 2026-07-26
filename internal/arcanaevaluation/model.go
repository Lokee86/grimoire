package arcanaevaluation

import (
	"time"

	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/structure"
)

const FormatVersion = 1

const (
	ModeLexiconSeeds       = "lexicon-seeds"
	ModeLexiconVectorSeeds = "lexicon-plus-vector"
	MaxSeedLimit           = 6
)

type SeedExpectation struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Kind   string `json:"kind,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Case struct {
	ID                   string                             `json:"id"`
	Query                string                             `json:"query"`
	Category             string                             `json:"category"`
	RequiredSeeds        []SeedExpectation                  `json:"required_seeds"`
	SupportingSeeds      []SeedExpectation                  `json:"supporting_seeds,omitempty"`
	RequiredStructural   []evaluation.StructuralExpectation `json:"required_structural"`
	SupportingStructural []evaluation.StructuralExpectation `json:"supporting_structural,omitempty"`
	Notes                string                             `json:"notes,omitempty"`
}

type Corpus struct {
	Version    int    `json:"version"`
	Repository string `json:"repository"`
	SourceURL  string `json:"source_url,omitempty"`
	Revision   string `json:"revision,omitempty"`
	Scope      string `json:"scope,omitempty"`
	JudgedAt   string `json:"judged_at,omitempty"`
	TopK       int    `json:"top_k"`
	RecallAtK  []int  `json:"recall_at_k"`
	Cases      []Case `json:"cases"`
}

type RankedSeed struct {
	Node   structure.Node `json:"node"`
	Source string         `json:"source"`
}

type SeedResult struct {
	Rank       int            `json:"rank"`
	Source     string         `json:"source"`
	Node       structure.Node `json:"node"`
	Required   bool           `json:"required"`
	Supporting bool           `json:"supporting"`
	Relevant   bool           `json:"relevant"`
}

type RecallMetric struct {
	K     int     `json:"k"`
	Value float64 `json:"value"`
}

type StructuralJudgment struct {
	Expectation evaluation.StructuralExpectation `json:"expectation"`
	Matched     bool                             `json:"matched"`
}

type Timings struct {
	LexiconSeedMS  float64 `json:"lexicon_seed_ms"`
	SemanticSeedMS float64 `json:"semantic_seed_ms,omitempty"`
	GraphSearchMS  float64 `json:"graph_search_ms"`
	TotalMS        float64 `json:"total_ms"`
}

type Measurement struct {
	Mode          string
	Seeds         []RankedSeed
	Structural    []structure.Evidence
	VectorUsed    bool
	ProviderCalls int
	Timings       Timings
	Error         string
}

type CaseResult struct {
	CaseID                    string               `json:"case_id"`
	Query                     string               `json:"query"`
	Category                  string               `json:"category"`
	Mode                      string               `json:"mode"`
	Pass                      bool                 `json:"pass"`
	Error                     string               `json:"error,omitempty"`
	VectorUsed                bool                 `json:"vector_used"`
	ProviderCalls             int                  `json:"provider_calls"`
	Timings                   Timings              `json:"timings"`
	SeedPayloadBytes          int                  `json:"seed_payload_bytes"`
	StructuralPayloadBytes    int                  `json:"structural_payload_bytes"`
	PayloadBytes              int                  `json:"payload_bytes"`
	RequiredSeedRecall        float64              `json:"required_seed_recall"`
	RecallAtK                 []RecallMetric       `json:"recall_at_k"`
	MRR                       float64              `json:"mrr"`
	FirstRequiredSeedRank     int                  `json:"first_required_seed_rank,omitempty"`
	RequiredSeedsMatched      []SeedExpectation    `json:"required_seeds_matched,omitempty"`
	RequiredSeedsMissing      []SeedExpectation    `json:"required_seeds_missing,omitempty"`
	Seeds                     []SeedResult         `json:"seeds"`
	RequiredStructuralRecall  float64              `json:"required_structural_recall"`
	RequiredStructuralMatched int                  `json:"required_structural_matched"`
	RequiredStructuralMissing int                  `json:"required_structural_missing"`
	StructuralJudgments       []StructuralJudgment `json:"structural_judgments"`
	StructuralEvidence        []structure.Evidence `json:"structural_evidence"`
}

type Aggregate struct {
	Mode                     string         `json:"mode"`
	Cases                    int            `json:"cases"`
	Passes                   int            `json:"passes"`
	PassRate                 float64        `json:"pass_rate"`
	RequiredSeedRecall       float64        `json:"required_seed_recall"`
	RecallAtK                []RecallMetric `json:"recall_at_k"`
	MRR                      float64        `json:"mrr"`
	RequiredStructuralRecall float64        `json:"required_structural_recall"`
	MedianLatencyMS          float64        `json:"median_latency_ms"`
	P95LatencyMS             float64        `json:"p95_latency_ms"`
	MedianPayloadBytes       float64        `json:"median_payload_bytes"`
	P95PayloadBytes          float64        `json:"p95_payload_bytes"`
	MeanProviderCalls        float64        `json:"mean_provider_calls"`
	VectorUsedCases          int            `json:"vector_used_cases"`
	ErrorCases               int            `json:"error_cases"`
}

type Comparison struct {
	BaselineMode                  string         `json:"baseline_mode"`
	VectorMode                    string         `json:"vector_mode"`
	PassRateDelta                 float64        `json:"pass_rate_delta"`
	RequiredSeedRecallDelta       float64        `json:"required_seed_recall_delta"`
	RecallAtKDelta                []RecallMetric `json:"recall_at_k_delta"`
	MRRDelta                      float64        `json:"mrr_delta"`
	RequiredStructuralRecallDelta float64        `json:"required_structural_recall_delta"`
	MedianLatencyMSDelta          float64        `json:"median_latency_ms_delta"`
	MedianPayloadBytesDelta       float64        `json:"median_payload_bytes_delta"`
	MeanProviderCallsDelta        float64        `json:"mean_provider_calls_delta"`
}

type Report struct {
	Version           int          `json:"version"`
	GeneratedAt       time.Time    `json:"generated_at"`
	Repository        string       `json:"repository"`
	SourceURL         string       `json:"source_url,omitempty"`
	Revision          string       `json:"revision,omitempty"`
	Scope             string       `json:"scope,omitempty"`
	JudgedAt          string       `json:"judged_at,omitempty"`
	Root              string       `json:"root"`
	State             string       `json:"state"`
	LexiconSnapshot   string       `json:"lexicon_snapshot,omitempty"`
	ArcanaSnapshot    string       `json:"arcana_snapshot"`
	EmbeddingIdentity string       `json:"embedding_identity"`
	CasesFile         string       `json:"cases_file"`
	Variant           string       `json:"variant"`
	TopK              int          `json:"top_k"`
	RecallAtK         []int        `json:"recall_at_k"`
	Modes             []string     `json:"modes"`
	Cases             []CaseResult `json:"cases"`
	Aggregates        []Aggregate  `json:"aggregates"`
	Comparison        Comparison   `json:"comparison"`
}
