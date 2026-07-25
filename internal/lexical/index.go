package lexical

import (
	"fmt"
	"sort"
)

type Posting struct {
	Document  int
	Frequency int
}

type Index struct {
	documents             []Document
	postings              map[string][]Posting
	candidatePostings     map[string][]int
	declarationVocabulary map[string]int
	averageLength         float64
}

func Build(inputs []Input) *Index {
	return Rebuild(inputs, nil)
}

func Rebuild(inputs []Input, previous *Index) *Index {
	previousDocuments := make(map[string]Document)
	if previous != nil {
		for _, document := range previous.documents {
			previousDocuments[document.Key] = document
		}
	}
	documents := make([]Document, 0, len(inputs))
	for _, input := range inputs {
		if document, exists := previousDocuments[input.Key]; exists {
			documents = append(documents, document)
			continue
		}
		documents = append(documents, Analyze(input))
	}
	result, err := New(documents)
	if err != nil {
		panic(err)
	}
	return result
}

func New(documents []Document) (*Index, error) {
	result := &Index{
		documents:             append([]Document(nil), documents...),
		postings:              make(map[string][]Posting),
		candidatePostings:     make(map[string][]int),
		declarationVocabulary: make(map[string]int),
	}
	totalLength := 0
	for documentIndex, document := range result.documents {
		if document.Key == "" || document.Length < 0 {
			return nil, fmt.Errorf("invalid lexical document %d", documentIndex)
		}
		candidateTerms := make(map[string]struct{})
		previousTerm := ""
		for termIndex, term := range document.Terms {
			if term.Term == "" || term.Frequency <= 0 || termIndex > 0 && term.Term <= previousTerm {
				return nil, fmt.Errorf("invalid lexical term in document %q", document.Key)
			}
			previousTerm = term.Term
			result.postings[term.Term] = append(result.postings[term.Term], Posting{
				Document: documentIndex, Frequency: term.Frequency,
			})
			candidateTerms[term.Term] = struct{}{}
		}
		for _, field := range [][]string{
			document.BaseTokens,
			document.PathTokens,
			document.LeadingTokens,
			document.DeclarationTokens,
		} {
			if !sort.StringsAreSorted(field) {
				return nil, fmt.Errorf("unsorted lexical field in document %q", document.Key)
			}
			for _, term := range field {
				if term == "" {
					return nil, fmt.Errorf("empty lexical field term in document %q", document.Key)
				}
				candidateTerms[term] = struct{}{}
			}
		}
		for _, term := range document.DeclarationTokens {
			result.declarationVocabulary[term]++
		}
		for term := range candidateTerms {
			result.candidatePostings[term] = append(result.candidatePostings[term], documentIndex)
		}
		totalLength += document.Length
	}
	if len(result.documents) > 0 {
		result.averageLength = float64(totalLength) / float64(len(result.documents))
	}
	if result.averageLength == 0 {
		result.averageLength = 1
	}
	return result, nil
}

func (index *Index) Documents() []Document {
	return index.documents
}

func (index *Index) Document(document int) Document {
	if index == nil || document < 0 || document >= len(index.documents) {
		return Document{}
	}
	return index.documents[document]
}

func (index *Index) DocumentCount() int {
	if index == nil {
		return 0
	}
	return len(index.documents)
}

func (index *Index) AverageLength() float64 {
	if index == nil || index.averageLength == 0 {
		return 1
	}
	return index.averageLength
}

func (index *Index) Posting(term string) []Posting {
	if index == nil {
		return nil
	}
	return index.postings[term]
}

func (index *Index) DocumentFrequency(term string) int {
	return len(index.Posting(term))
}

func (index *Index) DeclarationVocabulary() map[string]int {
	if index == nil {
		return nil
	}
	return index.declarationVocabulary
}

func Contains(tokens []string, term string) bool {
	position := sort.SearchStrings(tokens, term)
	return position < len(tokens) && tokens[position] == term
}
