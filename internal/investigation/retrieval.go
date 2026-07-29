package investigation

import "fmt"

func resolveRetrievalHits(current manifest, response Response) ([]RetrievalHitRecord, error) {
	if len(response.RetrievalHits) == 0 {
		return nil, nil
	}
	resolved := make([]RetrievalHitRecord, 0, len(response.RetrievalHits))
	for _, hit := range response.RetrievalHits {
		handle, err := resolveEvidenceRef(current, response, hit.Evidence)
		if err != nil {
			return nil, fmt.Errorf("resolve retrieval hit: %w", err)
		}
		record := RetrievalHitRecord{
			EvidenceHandle: handle,
			EvidenceKind:   hit.Evidence.Kind,
			Lane:           hit.Lane,
			Provider:       hit.Provider,
			Rank:           hit.Rank,
			Score:          hit.Score,
			Reasons:        append([]string(nil), hit.Reasons...),
			DuplicateOf:    hit.DuplicateOf,
			Direction:      hit.Direction,
			Relation:       hit.Relation,
			Certainty:      hit.Certainty,
			Depth:          hit.Depth,
			Support:        append([]string(nil), hit.Support...),
		}
		for _, related := range hit.RelatedEvidence {
			relatedHandle, err := resolveEvidenceRef(current, response, related)
			if err != nil {
				return nil, fmt.Errorf("resolve related retrieval evidence: %w", err)
			}
			record.RelatedEvidence = append(record.RelatedEvidence, EvidenceHandleRef{Handle: relatedHandle, Kind: related.Kind})
		}
		if hit.Seed != nil {
			seedHandle, err := resolveEvidenceRef(current, response, hit.Seed.Evidence)
			if err != nil {
				return nil, fmt.Errorf("resolve retrieval seed: %w", err)
			}
			record.Seed = &RetrievalSeedRecord{
				EvidenceHandle: seedHandle,
				EvidenceKind:   hit.Seed.Evidence.Kind,
				Lane:           hit.Seed.Lane,
				Provider:       hit.Seed.Provider,
				Rank:           hit.Seed.Rank,
				Score:          hit.Seed.Score,
				Reasons:        append([]string(nil), hit.Seed.Reasons...),
			}
		}
		resolved = append(resolved, record)
	}
	return resolved, nil
}

func resolveEvidenceRef(current manifest, response Response, ref EvidenceRef) (string, error) {
	kind, value, err := evidenceRefValue(response, ref)
	if err != nil {
		return "", err
	}
	key, _, err := evidenceKey(kind, current.Snapshot.Digest(), value)
	if err != nil {
		return "", err
	}
	meta, exists := current.Evidence[key]
	if exists {
		return indexedHandle(meta, kind)
	}
	handle, err := handleFor(kind, current.Snapshot.Digest(), key)
	if err != nil {
		return "", fmt.Errorf("build handle for evidence reference %s[%d]: %w", ref.Kind, ref.Index, err)
	}
	return handle, nil
}

func evidenceRefValue(response Response, ref EvidenceRef) (string, any, error) {
	if err := validateEvidenceRef(response, ref); err != nil {
		return "", nil, err
	}
	switch ref.Kind {
	case "node":
		return "node", response.Nodes[ref.Index], nil
	case "source":
		return "source", response.SourceRanges[ref.Index], nil
	case "path":
		return "path", response.GraphPaths[ref.Index], nil
	case "document":
		return "document", response.Documents[ref.Index], nil
	default:
		return "", nil, fmt.Errorf("unsupported evidence reference kind %q", ref.Kind)
	}
}
