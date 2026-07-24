use super::{MutationError, MutationScenario};
use crate::synthetic::sampling::Permutation;
use crate::synthetic::{GraphDataset, NodeId};

pub(super) fn selected_indices(
    dataset: &GraphDataset,
    scenario: MutationScenario,
    seed: u64,
) -> Result<Vec<usize>, MutationError> {
    match scenario {
        MutationScenario::SingleNode { node } => {
            validate_node(node, dataset.node_count)?;
            let candidates = incident_indices(dataset, node.0, node.0 + 1);
            select_candidates(candidates, None, seed)
        }
        MutationScenario::LocalRange {
            start,
            node_count,
            edge_count,
        } => {
            validate_range(start, node_count, dataset.node_count)?;
            validate_requested_count(edge_count)?;
            let candidates = incident_indices(dataset, start.0, start.0 + node_count);
            select_candidates(candidates, Some(edge_count), seed)
        }
        MutationScenario::Scattered { edge_count } => {
            validate_requested_count(edge_count)?;
            select_candidates((0..dataset.edges.len()).collect(), Some(edge_count), seed)
        }
        MutationScenario::Hub {
            hub_count,
            edge_count,
        } => {
            if hub_count == 0 || hub_count > dataset.node_count {
                return Err(MutationError::InvalidHubCount {
                    hub_count,
                    node_count: dataset.node_count,
                });
            }
            validate_requested_count(edge_count)?;
            let candidates = incident_indices(dataset, 0, hub_count);
            select_candidates(candidates, Some(edge_count), seed)
        }
        MutationScenario::Percentage { basis_points } => {
            if basis_points == 0 || basis_points > 10_000 {
                return Err(MutationError::InvalidPercentage { basis_points });
            }
            if dataset.edges.is_empty() {
                return Err(MutationError::NoMatchingEdges);
            }
            let requested =
                (dataset.edges.len() as u128 * u128::from(basis_points)).div_ceil(10_000);
            select_candidates(
                (0..dataset.edges.len()).collect(),
                Some(requested as u64),
                seed,
            )
        }
    }
}

fn validate_node(node: NodeId, node_count: u32) -> Result<(), MutationError> {
    if node.0 >= node_count {
        return Err(MutationError::InvalidNode { node, node_count });
    }
    Ok(())
}

fn validate_range(start: NodeId, count: u32, node_count: u32) -> Result<(), MutationError> {
    if count == 0 || start.0 >= node_count || count > node_count - start.0 {
        return Err(MutationError::InvalidRange {
            start,
            count,
            node_count,
        });
    }
    Ok(())
}

fn validate_requested_count(edge_count: u64) -> Result<(), MutationError> {
    if edge_count == 0 {
        return Err(MutationError::ZeroRequestedEdges);
    }
    Ok(())
}

fn incident_indices(dataset: &GraphDataset, start: u32, end: u32) -> Vec<usize> {
    dataset
        .edges
        .iter()
        .enumerate()
        .filter_map(|(index, edge)| {
            let source_matches = (start..end).contains(&edge.source.0);
            let target_matches = (start..end).contains(&edge.target.0);
            (source_matches || target_matches).then_some(index)
        })
        .collect()
}

fn select_candidates(
    candidates: Vec<usize>,
    requested: Option<u64>,
    seed: u64,
) -> Result<Vec<usize>, MutationError> {
    if candidates.is_empty() {
        return Err(MutationError::NoMatchingEdges);
    }

    let requested = requested.unwrap_or(candidates.len() as u64);
    if requested > candidates.len() as u64 {
        return Err(MutationError::RequestedEdgesExceedCandidates {
            requested,
            candidates: candidates.len() as u64,
        });
    }

    let mut selected: Vec<usize> = Permutation::new(candidates.len() as u64, seed)
        .take(requested as usize)
        .map(|ordinal| candidates[ordinal as usize])
        .collect();
    selected.sort_unstable();
    Ok(selected)
}
