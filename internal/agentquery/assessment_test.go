package agentquery

import (
	"slices"
	"testing"
)

func TestNarrowSearchAssessmentGuidesTargetedInspection(t *testing.T) {
	response := Response{
		ExactMatches: []Result{{
			Provider: "exact",
			Node: Node{
				Kind: "method", Name: "BackgroundCompaction", Path: "db/db_impl.cc",
				Handle: Handle{Value: "owner", Provider: "source", Path: "db/db_impl.cc", StartLine: 10, EndLine: 30},
			},
		}},
		SymbolMatches: []Result{{
			Provider: "lexicon",
			Node: Node{
				Kind: "interface", Name: "DB", Path: "include/leveldb/db.h",
				Handle: Handle{Value: "api", Provider: "lexicon", Path: "include/leveldb/db.h"},
			},
		}},
		SourceMatches: []Result{{
			Provider: "lexical",
			Node: Node{
				Kind: "test", Name: "BackgroundCompactionTest", Path: "db/db_test.cc",
				Handle: Handle{Value: "test", Provider: "source", Path: "db/db_test.cc", StartLine: 40, EndLine: 60},
			},
		}},
	}

	assessEvidence(Request{Mode: "search", Breadth: "narrow"}, &response)
	if response.Assessment == nil {
		t.Fatal("narrow search returned no assessment")
	}
	assessment := response.Assessment
	if assessment.Status != "ready-for-targeted-inspection" || assessment.NextAction != "inspect-selected-handles" {
		t.Fatalf("unexpected narrow assessment: %+v", assessment)
	}
	for _, dimension := range narrowEvidenceDimensions {
		if !slices.Contains(assessment.ObservedDimensions, dimension) {
			t.Fatalf("assessment missed %q: %+v", dimension, assessment)
		}
	}
	if len(assessment.MissingDimensions) != 0 || assessment.DistinctPaths != 3 {
		t.Fatalf("unexpected narrow coverage: %+v", assessment)
	}
}

func TestInspectionAssessmentStopsWhenDimensionsAreGrounded(t *testing.T) {
	response := Response{Inspections: []Inspection{
		{
			Handle:         Handle{Provider: "lexicon"},
			Node:           &Node{Kind: "method", Name: "MaybeScheduleCompaction", Path: "db/db_impl.cc"},
			ContainingSpan: Range{Path: "db/db_impl.cc", StartLine: 1, EndLine: 20},
			Source:         "void DBImpl::MaybeScheduleCompaction() {}",
		},
		{
			Handle:         Handle{Provider: "source"},
			Node:           &Node{Kind: "interface", Name: "DB", Path: "include/leveldb/db.h"},
			ContainingSpan: Range{Path: "include/leveldb/db.h", StartLine: 1, EndLine: 20},
			Source:         "class DB {};",
		},
		{
			Handle:         Handle{Provider: "source"},
			Node:           &Node{Kind: "test", Name: "PauseCompactionTest", Path: "db/db_test.cc"},
			ContainingSpan: Range{Path: "db/db_test.cc", StartLine: 1, EndLine: 20},
			Source:         "TEST(PauseCompactionTest, Resume) {}",
		},
	}}

	assessEvidence(Request{Mode: "inspect"}, &response)
	if response.Assessment == nil || response.Assessment.Status != "ready-to-synthesize" ||
		response.Assessment.NextAction != "synthesize" {
		t.Fatalf("inspection did not produce a stopping signal: %+v", response.Assessment)
	}
}

func TestInspectionAssessmentNamesMissingDimensions(t *testing.T) {
	response := Response{Inspections: []Inspection{{
		Handle:         Handle{Provider: "source"},
		Node:           &Node{Kind: "method", Name: "BackgroundCompaction", Path: "db/db_impl.cc"},
		ContainingSpan: Range{Path: "db/db_impl.cc", StartLine: 1, EndLine: 20},
		Source:         "void DBImpl::BackgroundCompaction() {}",
	}}}

	assessEvidence(Request{Mode: "inspect"}, &response)
	assessment := response.Assessment
	if assessment == nil || assessment.Status != "partial" || assessment.NextAction != "inspect-missing-dimensions" {
		t.Fatalf("unexpected partial assessment: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingDimensions, "public_boundary") || !slices.Contains(assessment.MissingDimensions, "tests") {
		t.Fatalf("missing dimensions were not identified: %+v", assessment)
	}
}
