package retrieve

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
)

type ScoreDetail struct {
	Name  string
	Value float64
}

type Candidate struct {
	Chunk        index.Chunk
	Score        float64
	Source       string
	Rank         int
	Reasons      []string
	ScoreDetails []ScoreDetail
	Context      *evidence.Descriptor
}

type lexicalQuerySpec struct {
	phrase      string
	termIndexes []int
}

type lexicalFields struct {
	text              string
	baseMatches       []bool
	pathMatches       []bool
	leadingMatches    []bool
	declarationTokens map[string]struct{}
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
	config = normalizedConfig(config)
	specs, terms := compileLexicalQueries(queries)
	results := make([][]Candidate, len(queries))
	if len(terms) == 0 {
		return results
	}

	chunks := snapshot.AllChunks()
	texts := make([]string, len(chunks))
	for chunkIndex, chunk := range chunks {
		texts[chunkIndex] = chunk.Text
	}
	corpus := newBM25Corpus(texts, terms)
	fields := make([]lexicalFields, len(chunks))
	vocabulary := make(map[string]declarationVocabularyEntry)
	for chunkIndex, chunk := range chunks {
		text := strings.ToLower(chunk.Text)
		firstLine := text
		if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
			firstLine = firstLine[:newline]
		}
		declarationTokens := lexicalTokenSet(
			filepath.Base(chunk.Path) + "\n" + chunk.Path + "\n" + declarationHeader(chunk.Text),
		)
		for token := range declarationTokens {
			entry := vocabulary[token]
			entry.documentFrequency++
			vocabulary[token] = entry
		}
		fields[chunkIndex] = lexicalFields{
			text:              text,
			baseMatches:       corpus.matches(filepath.Base(chunk.Path)),
			pathMatches:       corpus.matches(chunk.Path),
			leadingMatches:    corpus.matches(firstLine),
			declarationTokens: declarationTokens,
		}
	}

	for queryIndex, spec := range specs {
		if len(spec.termIndexes) == 0 {
			continue
		}
		aliases := queryDeclarationAliases(corpus, spec, vocabulary, config.DeclarationAliasBonus)
		candidates := make([]Candidate, 0)
		for chunkIndex, chunk := range chunks {
			candidate := scoreChunk(
				chunk, chunkIndex, corpus, fields[chunkIndex], spec, aliases, config,
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
	fields lexicalFields,
	spec lexicalQuerySpec,
	aliases map[int]declarationAlias,
	config Config,
) Candidate {
	candidate := Candidate{Chunk: chunk, Source: "lexical"}
	if len(spec.termIndexes) > 1 && len(spec.phrase) > 2 && strings.Contains(fields.text, spec.phrase) {
		candidate.addScore("exact query phrase in content", 12)
	}

	for _, termIndex := range spec.termIndexes {
		term := corpus.terms[termIndex]
		if fields.baseMatches[termIndex] {
			candidate.addScore("filename matches "+term.text, 8)
		} else if fields.pathMatches[termIndex] {
			candidate.addScore("path matches "+term.text, 4)
		}
		if fields.leadingMatches[termIndex] {
			candidate.addScore("leading line matches "+term.text, 4)
		}
		if alias, exists := aliases[termIndex]; exists {
			if _, matched := fields.declarationTokens[alias.token]; matched {
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

func lexicalTokenSet(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, token := range lexicalTokens(value) {
		result[token] = struct{}{}
	}
	return result
}

func declarationHeader(text string) string {
	const maxLines = 6
	const maxBytes = 768
	lines := strings.Split(text, "\n")
	var header strings.Builder
	linesAdded := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || commentOnlyLine(line) {
			continue
		}
		if header.Len() > 0 {
			header.WriteByte('\n')
		}
		header.WriteString(line)
		linesAdded++
		if linesAdded >= maxLines || header.Len() >= maxBytes {
			break
		}
	}
	value := header.String()
	if len(value) > maxBytes {
		value = value[:maxBytes]
	}
	return value
}

func commentOnlyLine(line string) bool {
	for _, prefix := range []string{"//", "#", "--", "/*", "*", "<!--"} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func (candidate *Candidate) addScore(name string, value float64) {
	candidate.Score += value
	candidate.Reasons = append(candidate.Reasons, name)
	candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{Name: name, Value: value})
}
