package lexical

import "sort"

func (index *Index) CandidateDocuments(terms []string) []int {
	if index == nil || len(terms) == 0 {
		return nil
	}
	seen := make(map[int]struct{})
	for _, term := range terms {
		for _, document := range index.candidatePostings[term] {
			seen[document] = struct{}{}
		}
	}
	result := make([]int, 0, len(seen))
	for document := range seen {
		result = append(result, document)
	}
	sort.Ints(result)
	return result
}

func (index *Index) CandidateDocumentsAll(terms []string) []int {
	if index == nil || len(terms) == 0 {
		return nil
	}
	orderedTerms := append([]string(nil), terms...)
	sort.Slice(orderedTerms, func(left, right int) bool {
		return len(index.candidatePostings[orderedTerms[left]]) < len(index.candidatePostings[orderedTerms[right]])
	})
	seed := index.candidatePostings[orderedTerms[0]]
	if len(seed) == 0 {
		return []int{}
	}
	result := append([]int(nil), seed...)
	for _, term := range orderedTerms[1:] {
		postings := index.candidatePostings[term]
		if len(postings) == 0 {
			return []int{}
		}
		result = intersectSortedDocuments(result, postings)
		if len(result) == 0 {
			return result
		}
	}
	return result
}

func intersectSortedDocuments(left, right []int) []int {
	result := make([]int, 0, min(len(left), len(right)))
	for leftIndex, rightIndex := 0, 0; leftIndex < len(left) && rightIndex < len(right); {
		switch {
		case left[leftIndex] < right[rightIndex]:
			leftIndex++
		case right[rightIndex] < left[leftIndex]:
			rightIndex++
		default:
			result = append(result, left[leftIndex])
			leftIndex++
			rightIndex++
		}
	}
	return result
}
