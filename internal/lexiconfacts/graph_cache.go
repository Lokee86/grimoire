package lexiconfacts

import "sort"

func (corpus *Corpus) prepareGraph() {
	if corpus == nil {
		return
	}
	corpus.graphOnce.Do(func() {
		corpus.adjacency = relationshipAdjacency(corpus.facts.edges)
		corpus.degrees = graphDegrees(corpus.facts.edges)
		for id := range corpus.adjacency {
			values := corpus.adjacency[id]
			sort.SliceStable(values, func(i, j int) bool {
				left := interstackRelationPriority(values[i].edge.Relation)
				right := interstackRelationPriority(values[j].edge.Relation)
				if left != right {
					return left < right
				}
				if values[i].relatedID != values[j].relatedID {
					return values[i].relatedID < values[j].relatedID
				}
				if values[i].direction != values[j].direction {
					return values[i].direction < values[j].direction
				}
				return values[i].edge.Relation < values[j].edge.Relation
			})
			corpus.adjacency[id] = values
		}
	})
}

func (corpus *Corpus) graphAdjacency() map[string][]adjacentRelationship {
	corpus.prepareGraph()
	if corpus == nil {
		return nil
	}
	return corpus.adjacency
}

func (corpus *Corpus) graphDegrees() map[string]int {
	corpus.prepareGraph()
	if corpus == nil {
		return nil
	}
	return corpus.degrees
}

func (corpus *Corpus) queryNeighbors(id, direction string, relations map[string]bool) []adjacentRelationship {
	corpus.prepareGraph()
	values := corpus.adjacency[id]
	if len(values) == 0 {
		return nil
	}
	result := make([]adjacentRelationship, 0, min(len(values), 16))
	for _, value := range values {
		if direction != "" && direction != "both" && value.direction != direction {
			continue
		}
		if len(relations) > 0 && !relations[value.edge.Relation] {
			continue
		}
		result = append(result, value)
	}
	return result
}
