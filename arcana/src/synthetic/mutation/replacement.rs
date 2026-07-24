use super::MutationError;
use crate::synthetic::sampling::{Permutation, directed_pair};
use crate::synthetic::{Edge, GraphDataset, NodeId};

pub(super) fn replacement_edges(
    dataset: &GraphDataset,
    removed: &[Edge],
    seed: u64,
) -> Result<Vec<Edge>, MutationError> {
    let total_capacity =
        u64::from(dataset.node_count) * u64::from(dataset.node_count.saturating_sub(1));
    let occupied_pairs = unique_pair_count(&dataset.edges);
    let available = total_capacity - occupied_pairs;
    if removed.len() as u64 > available {
        return Err(MutationError::InsufficientReplacementCapacity {
            requested: removed.len() as u64,
            available,
        });
    }

    let mut added = Vec::with_capacity(removed.len());
    for ordinal in Permutation::new(total_capacity, seed) {
        let (source, target) = directed_pair(ordinal, dataset.node_count);
        if contains_pair(&dataset.edges, source, target) {
            continue;
        }

        added.push(Edge {
            source: NodeId(source),
            target: NodeId(target),
            kind: removed[added.len()].kind,
        });
        if added.len() == removed.len() {
            break;
        }
    }

    debug_assert_eq!(added.len(), removed.len());
    added.sort_unstable();
    Ok(added)
}

fn unique_pair_count(edges: &[Edge]) -> u64 {
    let mut count = 0;
    let mut previous = None;
    for edge in edges {
        let pair = (edge.source, edge.target);
        if previous != Some(pair) {
            count += 1;
            previous = Some(pair);
        }
    }
    count
}

pub(super) fn contains_pair(edges: &[Edge], source: u32, target: u32) -> bool {
    let index = edges.partition_point(|edge| (edge.source.0, edge.target.0) < (source, target));
    edges
        .get(index)
        .is_some_and(|edge| edge.source.0 == source && edge.target.0 == target)
}
