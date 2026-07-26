package knowledgeevaluation

import (
	"path/filepath"
	"strings"
	"time"
)

const FormatVersion = 1

type SectionExpectation struct {
	Path        string   `json:"path"`
	Heading     string   `json:"heading,omitempty"`
	HeadingPath []string `json:"heading_path,omitempty"`
	SectionID   string   `json:"section_id,omitempty"`
}

func (expectation SectionExpectation) Key() string {
	if expectation.SectionID != "" {
		return filepath.ToSlash(expectation.Path) + "\x00" + expectation.SectionID
	}
	return filepath.ToSlash(expectation.Path) + "\x00" + expectation.Heading + "\x00" + strings.Join(expectation.HeadingPath, "\x1f")
}

type Case struct {
	ID         string               `json:"id"`
	Query      string               `json:"query"`
	Category   string               `json:"category"`
	Required   []SectionExpectation `json:"required"`
	Supporting []SectionExpectation `json:"supporting,omitempty"`
	Forbidden  []SectionExpectation `json:"forbidden,omitempty"`
	Notes      string               `json:"notes,omitempty"`
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

type RecallMetric struct {
	K     int     `json:"k"`
	Value float64 `json:"value"`
}

type Result struct {
	Handle     string  `json:"handle"`
	Path       string  `json:"path"`
	Heading    string  `json:"heading"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
	Required   bool    `json:"required"`
	Supporting bool    `json:"supporting"`
	Forbidden  bool    `json:"forbidden"`
	Relevant   bool    `json:"relevant"`
}

type CaseResult struct {
	CaseID                  string               `json:"case_id"`
	Query                   string               `json:"query"`
	Category                string               `json:"category"`
	Pass                    bool                 `json:"pass"`
	RequiredSectionRecall   float64              `json:"required_section_recall"`
	RecallAtK               []RecallMetric       `json:"recall_at_k"`
	MRR                     float64              `json:"mrr"`
	FirstRelevantRank       int                  `json:"first_relevant_rank,omitempty"`
	RelevantSelections      int                  `json:"relevant_selections"`
	IrrelevantSelections    int                  `json:"irrelevant_selections"`
	IrrelevantSelectionRate float64              `json:"irrelevant_selection_rate"`
	ForbiddenSelections     int                  `json:"forbidden_selections"`
	ResultCount             int                  `json:"result_count"`
	VectorUsed              bool                 `json:"vector_used"`
	VectorError             string               `json:"vector_error,omitempty"`
	LatencyMS               float64              `json:"latency_ms"`
	Results                 []Result             `json:"results"`
	RequiredMatched         []SectionExpectation `json:"required_matched,omitempty"`
	RequiredMissing         []SectionExpectation `json:"required_missing,omitempty"`
}

type Aggregate struct {
	Cases                   int            `json:"cases"`
	Passes                  int            `json:"passes"`
	PassRate                float64        `json:"pass_rate"`
	RequiredSectionRecall   float64        `json:"required_section_recall"`
	RecallAtK               []RecallMetric `json:"recall_at_k"`
	MRR                     float64        `json:"mrr"`
	IrrelevantSelections    int            `json:"irrelevant_selections"`
	IrrelevantSelectionRate float64        `json:"irrelevant_selection_rate"`
	VectorUsedCases         int            `json:"vector_used_cases"`
	VectorUsageRate         float64        `json:"vector_usage_rate"`
	VectorErrorCases        int            `json:"vector_error_cases"`
	MedianLatencyMS         float64        `json:"median_latency_ms"`
	P95LatencyMS            float64        `json:"p95_latency_ms"`
}

type Report struct {
	Version     int          `json:"version"`
	GeneratedAt time.Time    `json:"generated_at"`
	Repository  string       `json:"repository"`
	SourceURL   string       `json:"source_url,omitempty"`
	Revision    string       `json:"revision,omitempty"`
	Scope       string       `json:"scope,omitempty"`
	JudgedAt    string       `json:"judged_at,omitempty"`
	Root        string       `json:"root"`
	State       string       `json:"state"`
	CasesFile   string       `json:"cases_file"`
	Variant     string       `json:"variant"`
	Vectors     bool         `json:"vectors"`
	TopK        int          `json:"top_k"`
	RecallAtK   []int        `json:"recall_at_k"`
	Cases       []CaseResult `json:"cases"`
	Aggregate   Aggregate    `json:"aggregate"`
}

func effectiveTopK(corpus Corpus) int {
	if corpus.TopK > 0 {
		return corpus.TopK
	}
	return 10
}

func effectiveRecallAtK(corpus Corpus) []int {
	if len(corpus.RecallAtK) > 0 {
		return append([]int(nil), corpus.RecallAtK...)
	}
	return []int{1, 3, 5, 10}
}
