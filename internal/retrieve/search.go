package retrieve

import (
	"sort"
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexical"
)

type ScoreDetail struct {
	Name  string
	Value float64
}

type Candidate struct {
	Chunk             index.Chunk
	Score             float64
	Source            string
	Rank              int
	Reasons           []string
	ScoreDetails      []ScoreDetail
	GraphScoreDetails []ScoreDetail
	Context           *evidence.Descriptor
}

type lexicalQuerySpec struct {
	phrase      string
	termIndexes []int
}

type declarationVocabularyEntry struct {
	documentFrequency int
}

type declarationAlias struct {
	token      string
	similarity float64
}

func Search(snapshot index.Snapshot, query string, limit int) []Candidate {
	return SearchWithConfig(snapshot, query, limit, DefaultConfig())
}

func SearchWithConfig(snapshot index.Snapshot, query string, limit int, config Config) []Candidate {
	results := SearchManyWithConfig(snapshot, []string{query}, limit, config)
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

// SearchMany scores a bounded set of queries against one shared BM25 corpus.
// Repository content and field tokens are scanned once per request rather than
// once per retrieval intent.
func SearchMany(snapshot index.Snapshot, queries []string, limit int) [][]Candidate {
	return SearchManyWithConfig(snapshot, queries, limit, DefaultConfig())
}

func SearchManyWithConfig(snapshot index.Snapshot, queries []string, limit int, config Config) [][]Candidate {
	return searchManyWithConfig(snapshot, queries, limit, config, nil)
}

func searchManyWithConfig(
	snapshot index.Snapshot,
	queries []string,
	limit int,
	config Config,
	allowedPaths map[string]struct{},
) [][]Candidate {
	config = normalizedConfig(config)
	specs, terms := compileLexicalQueries(queries)
	results := make([][]Candidate, len(queries))
	if len(terms) == 0 {
		return results
	}

	chunks := snapshot.AllChunks()
	lexicalIndex := snapshot.LexicalIndex()
	corpus := newBM25Corpus(lexicalIndex, terms)
	vocabulary := make(map[string]declarationVocabularyEntry)
	for token, frequency := range lexicalIndex.DeclarationVocabulary() {
		vocabulary[token] = declarationVocabularyEntry{documentFrequency: frequency}
	}

	for queryIndex, spec := range specs {
		if len(spec.termIndexes) == 0 {
			continue
		}
		aliases := queryDeclarationAliases(corpus, spec, vocabulary, config.DeclarationAliasBonus)
		candidateTerms := make([]string, 0, len(spec.termIndexes)+len(aliases))
		for _, termIndex := range spec.termIndexes {
			candidateTerms = append(candidateTerms, corpus.terms[termIndex].text)
		}
		for _, alias := range aliases {
			candidateTerms = append(candidateTerms, alias.token)
		}
		candidates := make([]Candidate, 0)
		for _, chunkIndex := range lexicalIndex.CandidateDocuments(candidateTerms) {
			if chunkIndex < 0 || chunkIndex >= len(chunks) {
				continue
			}
			chunk := chunks[chunkIndex]
			if allowedPaths != nil {
				if _, allowed := allowedPaths[chunk.Path]; !allowed {
					continue
				}
			}
			candidate := scoreChunk(
				chunk, chunkIndex, corpus, lexicalIndex.Document(chunkIndex), spec, aliases, config,
			)
			if candidate.Score > 0 {
				candidates = append(candidates, candidate)
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].Score != candidates[j].Score {
				return candidates[i].Score > candidates[j].Score
			}
			if candidates[i].Chunk.Path != candidates[j].Chunk.Path {
				return candidates[i].Chunk.Path < candidates[j].Chunk.Path
			}
			return candidates[i].Chunk.StartLine < candidates[j].Chunk.StartLine
		})
		if limit > 0 && len(candidates) > limit {
			candidates = candidates[:limit]
		}
		for index := range candidates {
			candidates[index].Rank = index + 1
		}
		results[queryIndex] = candidates
	}
	return results
}

func compileLexicalQueries(queries []string) ([]lexicalQuerySpec, []string) {
	specs := make([]lexicalQuerySpec, len(queries))
	termPositions := make(map[string]int)
	var terms []string
	for queryIndex, query := range queries {
		query = strings.TrimSpace(query)
		specs[queryIndex].phrase = strings.ToLower(query)
		for _, term := range queryTerms(query) {
			position, exists := termPositions[term]
			if !exists {
				position = len(terms)
				termPositions[term] = position
				terms = append(terms, term)
			}
			specs[queryIndex].termIndexes = append(specs[queryIndex].termIndexes, position)
		}
	}
	return specs, terms
}

func scoreChunk(
	chunk index.Chunk,
	documentIndex int,
	corpus bm25Corpus,
	document lexical.Document,
	spec lexicalQuerySpec,
	aliases map[int]declarationAlias,
	config Config,
) Candidate {
	candidate := Candidate{Chunk: chunk, Source: "lexical"}
	if len(spec.termIndexes) > 1 && len(spec.phrase) > 2 && strings.Contains(strings.ToLower(chunk.Text), spec.phrase) {
		candidate.addScore("exact query phrase in content", 12)
	}

	for _, termIndex := range spec.termIndexes {
		term := corpus.terms[termIndex]
		if lexical.Contains(document.BaseTokens, term.text) {
			candidate.addScore("filename matches "+term.text, 8)
		} else if lexical.Contains(document.PathTokens, term.text) {
			candidate.addScore("path matches "+term.text, 4)
		}
		if lexical.Contains(document.LeadingTokens, term.text) {
			candidate.addScore("leading line matches "+term.text, 4)
		}
		if alias, exists := aliases[termIndex]; exists {
			if lexical.Contains(document.DeclarationTokens, alias.token) {
				candidate.addScore(
					"declaration alias "+term.text+" -> "+alias.token,
					config.DeclarationAliasBonus*alias.similarity,
				)
			}
		}
		if value := corpus.score(documentIndex, termIndex); value > 0 {
			candidate.addScore("BM25 content matches "+term.text, value)
		}
	}
	return candidate
}

func queryDeclarationAliases(
	corpus bm25Corpus,
	spec lexicalQuerySpec,
	vocabulary map[string]declarationVocabularyEntry,
	bonus float64,
) map[int]declarationAlias {
	if bonus <= 0 || len(vocabulary) == 0 {
		return nil
	}
	aliases := make(map[int]declarationAlias)
	for _, termIndex := range spec.termIndexes {
		term := corpus.terms[termIndex].text
		if _, exact := vocabulary[term]; exact {
			continue
		}
		alias, ok := nearestDeclarationAlias(term, vocabulary)
		if ok {
			aliases[termIndex] = alias
		}
	}
	return aliases
}

func nearestDeclarationAlias(
	term string,
	vocabulary map[string]declarationVocabularyEntry,
) (declarationAlias, bool) {
	if !eligibleAliasToken(term) {
		return declarationAlias{}, false
	}
	best := declarationAlias{}
	bestFrequency := 0
	for token, entry := range vocabulary {
		if !eligibleAliasToken(token) || token == term {
			continue
		}
		prefix := commonPrefixLength(term, token)
		if prefix < 4 || absoluteDifference(len(term), len(token)) > 5 {
			continue
		}
		similarity := lexicalSimilarity(term, token)
		if similarity < 0.55 && !(prefix >= 5 && similarity >= 0.5) {
			continue
		}
		if similarity > best.similarity ||
			(similarity == best.similarity && (best.token == "" || entry.documentFrequency < bestFrequency)) ||
			(similarity == best.similarity && entry.documentFrequency == bestFrequency && token < best.token) {
			best = declarationAlias{token: token, similarity: similarity}
			bestFrequency = entry.documentFrequency
		}
	}
	return best, best.token != ""
}

func eligibleAliasToken(value string) bool {
	if len(value) < 5 || len(value) > 32 {
		return false
	}
	for _, current := range value {
		if !unicode.IsLetter(current) {
			return false
		}
	}
	return true
}

func lexicalSimilarity(left, right string) float64 {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	maximum := max(len(leftRunes), len(rightRunes))
	if maximum == 0 {
		return 1
	}
	return 1 - float64(levenshteinDistance(leftRunes, rightRunes))/float64(maximum)
}

func levenshteinDistance(left, right []rune) int {
	previous := make([]int, len(right)+1)
	current := make([]int, len(right)+1)
	for index := range previous {
		previous[index] = index
	}
	for leftIndex, leftRune := range left {
		current[0] = leftIndex + 1
		for rightIndex, rightRune := range right {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[rightIndex+1] = min(
				previous[rightIndex+1]+1,
				current[rightIndex]+1,
				previous[rightIndex]+cost,
			)
		}
		previous, current = current, previous
	}
	return previous[len(right)]
}

func commonPrefixLength(left, right string) int {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	limit := min(len(leftRunes), len(rightRunes))
	for index := 0; index < limit; index++ {
		if leftRunes[index] != rightRunes[index] {
			return index
		}
	}
	return limit
}

func absoluteDifference(left, right int) int {
	if left < right {
		return right - left
	}
	return left - right
}

func (candidate *Candidate) addScore(name string, value float64) {
	candidate.Score += value
	candidate.Reasons = append(candidate.Reasons, name)
	candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{Name: name, Value: value})
}
