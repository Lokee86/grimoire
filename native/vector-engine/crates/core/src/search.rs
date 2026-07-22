use std::{cmp::Ordering, collections::BinaryHeap};

use rayon::prelude::*;

use crate::{Error, Result, Snapshot};

#[derive(Clone, Debug, PartialEq)]
pub struct SearchHit {
    pub id: String,
    pub score: f32,
    pub index: usize,
}

#[derive(Clone, Copy, Debug)]
struct Candidate {
    index: usize,
    score: f32,
}

impl Eq for Candidate {}

impl PartialEq for Candidate {
    fn eq(&self, other: &Self) -> bool {
        self.index == other.index && self.score.to_bits() == other.score.to_bits()
    }
}

impl Ord for Candidate {
    fn cmp(&self, other: &Self) -> Ordering {
        other
            .score
            .total_cmp(&self.score)
            .then_with(|| self.index.cmp(&other.index))
    }
}

impl PartialOrd for Candidate {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

pub fn search(snapshot: &Snapshot, query: &[f32], top_k: usize) -> Result<Vec<SearchHit>> {
    if query.len() != snapshot.info().dimensions || query.iter().any(|value| !value.is_finite()) {
        return Err(Error::InvalidInput(
            "query dimensions or values are invalid".into(),
        ));
    }
    if top_k == 0 {
        return Ok(Vec::new());
    }
    let dimensions = query.len();
    let vectors = snapshot.vectors();
    let count = snapshot.info().count;
    let heap = if count < 2048 {
        scan(vectors.chunks_exact(dimensions).enumerate(), query, top_k)
    } else {
        vectors
            .par_chunks_exact(dimensions)
            .enumerate()
            .fold(
                || BinaryHeap::with_capacity(top_k + 1),
                |mut heap, (index, vector)| {
                    push(
                        &mut heap,
                        Candidate {
                            index,
                            score: dot(query, vector),
                        },
                        top_k,
                    );
                    heap
                },
            )
            .reduce(
                || BinaryHeap::with_capacity(top_k + 1),
                |mut left, right| {
                    for candidate in right {
                        push(&mut left, candidate, top_k);
                    }
                    left
                },
            )
    };
    let mut candidates = heap.into_vec();
    candidates.sort_by(|left, right| {
        right
            .score
            .total_cmp(&left.score)
            .then_with(|| left.index.cmp(&right.index))
    });
    Ok(candidates
        .into_iter()
        .map(|candidate| SearchHit {
            id: snapshot.id(candidate.index).to_owned(),
            score: candidate.score,
            index: candidate.index,
        })
        .collect())
}

fn scan<'a>(
    vectors: impl Iterator<Item = (usize, &'a [f32])>,
    query: &[f32],
    top_k: usize,
) -> BinaryHeap<Candidate> {
    let mut heap = BinaryHeap::with_capacity(top_k + 1);
    for (index, vector) in vectors {
        push(
            &mut heap,
            Candidate {
                index,
                score: dot(query, vector),
            },
            top_k,
        );
    }
    heap
}

fn push(heap: &mut BinaryHeap<Candidate>, candidate: Candidate, top_k: usize) {
    if heap.len() < top_k {
        heap.push(candidate);
        return;
    }
    let worst = heap.peek().copied().expect("non-empty top-k heap");
    if candidate.score > worst.score
        || (candidate.score == worst.score && candidate.index < worst.index)
    {
        heap.pop();
        heap.push(candidate);
    }
}

fn dot(left: &[f32], right: &[f32]) -> f32 {
    left.iter().zip(right).map(|(a, b)| a * b).sum()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{ObjectStore, RecordRef, materialize};

    #[test]
    fn returns_exact_top_results_with_stable_ties() {
        let temp = tempfile::tempdir().unwrap();
        let store = ObjectStore::new(temp.path().join("store"));
        store.put("m", "a", &[1.0, 0.0]).unwrap();
        store.put("m", "b", &[0.5, 0.5]).unwrap();
        store.put("m", "c", &[0.0, 1.0]).unwrap();
        let records = [
            RecordRef {
                id: "a".into(),
                source: "a".into(),
            },
            RecordRef {
                id: "b".into(),
                source: "b".into(),
            },
            RecordRef {
                id: "c".into(),
                source: "c".into(),
            },
        ];
        let path = temp.path().join("snapshot.gvs");
        materialize(&store, "m", &records, &path).unwrap();
        let snapshot = Snapshot::open(path).unwrap();
        let hits = search(&snapshot, &[1.0, 0.0], 2).unwrap();
        assert_eq!(
            hits.iter().map(|hit| hit.id.as_str()).collect::<Vec<_>>(),
            vec!["a", "b"]
        );
    }
}
