package evaluation

import (
	"time"

	"github.com/Lokee86/grimoire/internal/structure"
)

const FormatVersion = 1

type Category string

const (
	CategoryDirectLocation         Category = "direct-location"
	CategoryMechanismExplanation   Category = "mechanism-explanation"
	CategoryArchitectureOwnership  Category = "architecture-ownership"
	CategoryCallChainInvestigation Category = "call-chain-investigation"
	CategoryLongMixedQuery         Category = "long-mixed-query"
)

var Categories = []Category{
	CategoryDirectLocation,
	CategoryMechanismExplanation,
	CategoryArchitectureOwnership,
	CategoryCallChainInvestigation,
	CategoryLongMixedQuery,
}

type Evidence struct {
	Path    string   `json:"path"`
	Symbols []string `json:"symbols,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Minimum string   `json:"minimum,omitempty"`
}

// StructuralExpectation describes one provider fact that must survive into the
// final package. Empty optional fields are wildcards. Chain is matched as an
// ordered subsequence of Arcana call-chain node names or qualified names.
type StructuralExpectation struct {
	Provider     string   `json:"provider"`
	Kind         string   `json:"kind"`
	Symbol       string   `json:"symbol,omitempty"`
	Path         string   `json:"path,omitempty"`
	Relation     string   `json:"relation,omitempty"`
	Direction    string   `json:"direction,omitempty"`
	Certainty    string   `json:"certainty,omitempty"`
	TargetSymbol string   `json:"target_symbol,omitempty"`
	TargetPath   string   `json:"target_path,omitempty"`
	Chain        []string `json:"chain,omitempty"`
	Expression   string   `json:"expression,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

type Case struct {
	ID                   string                  `json:"id"`
	Query                string                  `json:"query"`
	Category             Category                `json:"category"`
	Budget               int                     `json:"budget"`
	Required             []Evidence              `json:"required,omitempty"`
	Supporting           []Evidence              `json:"supporting,omitempty"`
	Forbidden            []Evidence              `json:"forbidden,omitempty"`
	RequiredStructural   []StructuralExpectation `json:"required_structural,omitempty"`
	SupportingStructural []StructuralExpectation `json:"supporting_structural,omitempty"`
	ForbiddenStructural  []StructuralExpectation `json:"forbidden_structural,omitempty"`
	Notes                string                  `json:"notes,omitempty"`
}

type Corpus struct {
	Version    int    `json:"version"`
	Repository string `json:"repository"`
	SourceURL  string `json:"source_url,omitempty"`
	Revision   string `json:"revision,omitempty"`
	Scope      string `json:"scope,omitempty"`
	JudgedAt   string `json:"judged_at,omitempty"`
	Cases      []Case `json:"cases"`
}

type Timings struct {
	TotalMS                float64 `json:"total_ms"`
	SnapshotValidationMS   float64 `json:"snapshot_validation_ms,omitempty"`
	EmbeddingMS            float64 `json:"embedding_ms,omitempty"`
	VectorSearchMS         float64 `json:"vector_search_ms,omitempty"`
	LexicalSearchMS        float64 `json:"lexical_search_ms,omitempty"`
	LexiconSearchMS        float64 `json:"lexicon_search_ms,omitempty"`
	ArcanaSearchMS         float64 `json:"arcana_search_ms,omitempty"`
	StructuralProviderMS   float64 `json:"structural_provider_ms,omitempty"`
	CandidateMergeMS       float64 `json:"candidate_merge_ms,omitempty"`
	ExactRecoveryMS        float64 `json:"exact_recovery_ms,omitempty"`
	CurationMS             float64 `json:"curation_ms,omitempty"`
	PackageCompilationMS   float64 `json:"package_compilation_ms,omitempty"`
	SelectionCompilationMS float64 `json:"selection_compilation_ms,omitempty"`
	DiagnosticProbeMS      float64 `json:"diagnostic_probe_ms,omitempty"`
}

type ScoreDetail struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Candidate struct {
	Path            string
	StartLine       int
	EndLine         int
	Text            string
	Symbols         []string
	RetrievalSource string
	ProviderRank    int
	Score           float64
	ScoreDetails    []ScoreDetail
	Reasons         []string
	TokenCount      int
}

type CandidateStageDiagnostic struct {
	Rank            int           `json:"rank"`
	RetrievalSource string        `json:"retrieval_source,omitempty"`
	ProviderRank    int           `json:"provider_rank,omitempty"`
	Score           float64       `json:"score"`
	ScoreDetails    []ScoreDetail `json:"score_details,omitempty"`
	Reasons         []string      `json:"reasons,omitempty"`
}

type CandidateDiagnostic struct {
	Path       string                    `json:"path"`
	StartLine  int                       `json:"start_line"`
	EndLine    int                       `json:"end_line"`
	TokenCount int                       `json:"token_count"`
	Required   bool                      `json:"required"`
	Supporting bool                      `json:"supporting"`
	Forbidden  bool                      `json:"forbidden"`
	Retrieved  *CandidateStageDiagnostic `json:"retrieved,omitempty"`
	Exact      *CandidateStageDiagnostic `json:"exact,omitempty"`
	Merged     *CandidateStageDiagnostic `json:"merged,omitempty"`
	Curated    *CandidateStageDiagnostic `json:"curated,omitempty"`
	Included   *CandidateStageDiagnostic `json:"included,omitempty"`
}

type Selection struct {
	Path            string   `json:"path"`
	StartLine       int      `json:"start_line"`
	EndLine         int      `json:"end_line"`
	Symbols         []string `json:"symbols,omitempty"`
	RetrievalSource string   `json:"retrieval_source"`
	ProviderRank    int      `json:"provider_rank"`
	TokenCount      int      `json:"token_count"`
	Relevant        bool     `json:"relevant"`
	Forbidden       bool     `json:"forbidden"`
}

type StructuralSelection struct {
	Evidence  structure.Evidence `json:"evidence"`
	Relevant  bool               `json:"relevant"`
	Forbidden bool               `json:"forbidden"`
}

type EvidenceStatus struct {
	Evidence       Evidence `json:"evidence"`
	Indexed        bool     `json:"indexed"`
	BroadProbe     bool     `json:"broad_probe"`
	Retrieved      bool     `json:"retrieved"`
	RetrievedRank  int      `json:"retrieved_rank,omitempty"`
	ExactRecovered bool     `json:"exact_recovered"`
	Merged         bool     `json:"merged"`
	Curated        bool     `json:"curated"`
	Included       bool     `json:"included"`
	FailureStage   string   `json:"failure_stage,omitempty"`
}

type StructuralEvidenceStatus struct {
	Evidence     StructuralExpectation `json:"evidence"`
	Produced     bool                  `json:"produced"`
	Composed     bool                  `json:"composed"`
	Included     bool                  `json:"included"`
	FailureStage string                `json:"failure_stage,omitempty"`
}

type RankingMetrics struct {
	CandidateCount       int     `json:"candidate_count"`
	RequiredRecallAt10   float64 `json:"required_recall_at_10"`
	RequiredRecallAt20   float64 `json:"required_recall_at_20"`
	SupportingRecallAt10 float64 `json:"supporting_recall_at_10"`
	SupportingRecallAt20 float64 `json:"supporting_recall_at_20"`
	FirstRequiredRank    int     `json:"first_required_rank,omitempty"`
	ReciprocalRank       float64 `json:"reciprocal_rank"`
	RelevantRateAt10     float64 `json:"relevant_rate_at_10"`
	RelevantRateAt20     float64 `json:"relevant_rate_at_20"`
}

type CaseRun struct {
	CaseID                            string                     `json:"case_id"`
	Query                             string                     `json:"query"`
	Category                          Category                   `json:"category"`
	Mode                              string                     `json:"mode"`
	Variant                           string                     `json:"variant"`
	Budget                            int                        `json:"budget"`
	Pass                              bool                       `json:"pass"`
	Error                             string                     `json:"error,omitempty"`
	Warnings                          []string                   `json:"warnings,omitempty"`
	Timings                           Timings                    `json:"timings"`
	RetrievalSources                  []string                   `json:"retrieval_sources"`
	StructuralSources                 []string                   `json:"structural_sources,omitempty"`
	StructuralState                   []structure.ProviderState  `json:"structural_state,omitempty"`
	SelectedPaths                     []string                   `json:"selected_paths"`
	Selections                        []Selection                `json:"selections"`
	StructuralSelections              []StructuralSelection      `json:"structural_selections,omitempty"`
	FinalPackageTokens                int                        `json:"final_package_tokens"`
	CandidateCount                    int                        `json:"candidate_count"`
	CuratedCount                      int                        `json:"curated_count"`
	Ranking                           RankingMetrics             `json:"ranking"`
	CandidateDiagnostics              []CandidateDiagnostic      `json:"candidate_diagnostics,omitempty"`
	OmittedForBudget                  int                        `json:"omitted_for_budget"`
	OmittedStructuralForBudget        int                        `json:"omitted_structural_for_budget"`
	Required                          []EvidenceStatus           `json:"required,omitempty"`
	Supporting                        []EvidenceStatus           `json:"supporting,omitempty"`
	RequiredStructural                []StructuralEvidenceStatus `json:"required_structural,omitempty"`
	SupportingStructural              []StructuralEvidenceStatus `json:"supporting_structural,omitempty"`
	ForbiddenRecovered                []Evidence                 `json:"forbidden_recovered,omitempty"`
	ForbiddenStructuralRecovered      []StructuralExpectation    `json:"forbidden_structural_recovered,omitempty"`
	RequiredEvidenceRecall            float64                    `json:"required_evidence_recall"`
	SupportingEvidenceRecall          float64                    `json:"supporting_evidence_recall"`
	RequiredStructuralRecall          float64                    `json:"required_structural_recall"`
	SupportingStructuralRecall        float64                    `json:"supporting_structural_recall"`
	IrrelevantSelectionRate           float64                    `json:"irrelevant_selection_rate"`
	IrrelevantStructuralRate          float64                    `json:"irrelevant_structural_rate"`
	RequiredNeverRetrieved            int                        `json:"required_never_retrieved"`
	RequiredLostDuringMerge           int                        `json:"required_lost_during_merge"`
	RequiredLostDuringCuration        int                        `json:"required_lost_during_curation"`
	RequiredOmittedForBudget          int                        `json:"required_omitted_for_budget"`
	RequiredStructuralNeverProduced   int                        `json:"required_structural_never_produced"`
	RequiredStructuralLostComposition int                        `json:"required_structural_lost_during_composition"`
	RequiredStructuralOmittedBudget   int                        `json:"required_structural_omitted_for_budget"`
	FailureClassifications            []string                   `json:"failure_classifications,omitempty"`
}

type Aggregate struct {
	Group                      string  `json:"group"`
	Cases                      int     `json:"cases"`
	Passes                     int     `json:"passes"`
	PassRate                   float64 `json:"pass_rate"`
	RequiredEvidenceRecall     float64 `json:"required_evidence_recall"`
	SupportingEvidenceRecall   float64 `json:"supporting_evidence_recall"`
	RequiredStructuralRecall   float64 `json:"required_structural_recall"`
	SupportingStructuralRecall float64 `json:"supporting_structural_recall"`
	IrrelevantSelectionRate    float64 `json:"irrelevant_selection_rate"`
	IrrelevantStructuralRate   float64 `json:"irrelevant_structural_rate"`
	RankingCases               int     `json:"ranking_cases"`
	RequiredRecallAt10         float64 `json:"required_recall_at_10"`
	RequiredRecallAt20         float64 `json:"required_recall_at_20"`
	MeanReciprocalRank         float64 `json:"mean_reciprocal_rank"`
	RelevantRateAt10           float64 `json:"relevant_rate_at_10"`
	RelevantRateAt20           float64 `json:"relevant_rate_at_20"`
	MedianLatencyMS            float64 `json:"median_latency_ms"`
	P95LatencyMS               float64 `json:"p95_latency_ms"`
}

type Report struct {
	Version             int         `json:"version"`
	GeneratedAt         time.Time   `json:"generated_at"`
	Repository          string      `json:"repository"`
	SourceURL           string      `json:"source_url,omitempty"`
	Revision            string      `json:"revision,omitempty"`
	Scope               string      `json:"scope,omitempty"`
	JudgedAt            string      `json:"judged_at,omitempty"`
	Root                string      `json:"root"`
	CasesFile           string      `json:"cases_file"`
	State               string      `json:"state"`
	Variant             string      `json:"variant"`
	Modes               []string    `json:"modes"`
	StructuralProviders []string    `json:"structural_providers,omitempty"`
	Runs                []CaseRun   `json:"runs"`
	ByMode              []Aggregate `json:"by_mode"`
	ByCategory          []Aggregate `json:"by_category"`
	ByModeCategory      []Aggregate `json:"by_mode_category"`
}
