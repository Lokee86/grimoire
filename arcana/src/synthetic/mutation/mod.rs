use std::fmt;

use self::replacement::{contains_pair, replacement_edges};
use self::selection::selected_indices;
use super::sampling::finish_dataset;
use super::{Edge, GraphDataset, NodeId};

mod replacement;
mod selection;
#[cfg(test)]
mod tests;

const REPLACEMENT_SEED_SALT: u64 = 0x6a09_e667_f3bc_c909;

/// A deterministic update pattern applied to an existing synthetic graph.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum MutationScenario {
    SingleNode {
        node: NodeId,
    },
    LocalRange {
        start: NodeId,
        node_count: u32,
        edge_count: u64,
    },
    Scattered {
        edge_count: u64,
    },
    Hub {
        hub_count: u32,
        edge_count: u64,
    },
    Percentage {
        basis_points: u16,
    },
}

/// The exact edge replacement set used by storage update benchmarks.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct GraphMutation {
    pub removed: Vec<Edge>,
    pub added: Vec<Edge>,
}

/// Builds a deterministic mutation without changing the supplied dataset.
pub fn plan_mutation(
    dataset: &GraphDataset,
    scenario: MutationScenario,
    seed: u64,
) -> Result<GraphMutation, MutationError> {
    let selected = selected_indices(dataset, scenario, seed)?;
    let removed: Vec<Edge> = selected.iter().map(|index| dataset.edges[*index]).collect();
    let added = replacement_edges(dataset, &removed, seed ^ REPLACEMENT_SEED_SALT)?;
    Ok(GraphMutation { removed, added })
}

/// Applies a previously planned mutation and returns a canonical graph.
pub fn apply_mutation(
    dataset: &GraphDataset,
    mutation: &GraphMutation,
) -> Result<GraphDataset, MutationError> {
    if mutation.removed.len() != mutation.added.len() {
        return Err(MutationError::CountMismatch {
            removed: mutation.removed.len() as u64,
            added: mutation.added.len() as u64,
        });
    }

    let mut removed = mutation.removed.clone();
    removed.sort_unstable();
    if removed.windows(2).any(|pair| pair[0] == pair[1]) {
        return Err(MutationError::DuplicateRemovedEdge);
    }

    let mut retained = Vec::with_capacity(dataset.edges.len());
    let mut removed_index = 0;
    for edge in &dataset.edges {
        if removed_index < removed.len() && *edge == removed[removed_index] {
            removed_index += 1;
        } else {
            retained.push(*edge);
        }
    }
    if removed_index != removed.len() {
        return Err(MutationError::RemovedEdgeMissing);
    }

    let mut added = mutation.added.clone();
    added.sort_unstable();
    if added
        .windows(2)
        .any(|pair| pair[0].source == pair[1].source && pair[0].target == pair[1].target)
    {
        return Err(MutationError::DuplicateAddedEdge);
    }
    if added.iter().any(|edge| {
        edge.source == edge.target
            || edge.source.0 >= dataset.node_count
            || edge.target.0 >= dataset.node_count
            || contains_pair(&dataset.edges, edge.source.0, edge.target.0)
    }) {
        return Err(MutationError::InvalidAddedEdge);
    }

    retained.extend(added);
    Ok(finish_dataset(
        dataset.node_count,
        dataset.edges.len() as u64,
        retained,
    ))
}

/// A mutation request or mutation value that cannot be applied.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum MutationError {
    InvalidNode {
        node: NodeId,
        node_count: u32,
    },
    InvalidRange {
        start: NodeId,
        count: u32,
        node_count: u32,
    },
    InvalidHubCount {
        hub_count: u32,
        node_count: u32,
    },
    InvalidPercentage {
        basis_points: u16,
    },
    ZeroRequestedEdges,
    NoMatchingEdges,
    RequestedEdgesExceedCandidates {
        requested: u64,
        candidates: u64,
    },
    InsufficientReplacementCapacity {
        requested: u64,
        available: u64,
    },
    CountMismatch {
        removed: u64,
        added: u64,
    },
    DuplicateRemovedEdge,
    DuplicateAddedEdge,
    RemovedEdgeMissing,
    InvalidAddedEdge,
}

impl fmt::Display for MutationError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::InvalidNode { node, node_count } => write!(
                formatter,
                "node {} is outside the graph's {} nodes",
                node.0, node_count
            ),
            Self::InvalidRange {
                start,
                count,
                node_count,
            } => write!(
                formatter,
                "node range {}..{} is invalid for {} nodes",
                start.0,
                start.0.saturating_add(*count),
                node_count
            ),
            Self::InvalidHubCount {
                hub_count,
                node_count,
            } => write!(
                formatter,
                "hub_count ({hub_count}) must be between 1 and node_count ({node_count})"
            ),
            Self::InvalidPercentage { basis_points } => write!(
                formatter,
                "basis_points ({basis_points}) must be between 1 and 10,000"
            ),
            Self::ZeroRequestedEdges => formatter.write_str("edge_count must be greater than zero"),
            Self::NoMatchingEdges => formatter.write_str("the mutation scenario matched no edges"),
            Self::RequestedEdgesExceedCandidates {
                requested,
                candidates,
            } => write!(
                formatter,
                "requested {requested} edges from only {candidates} matching candidates"
            ),
            Self::InsufficientReplacementCapacity {
                requested,
                available,
            } => write!(
                formatter,
                "requested {requested} replacement edges but only {available} unused node pairs exist"
            ),
            Self::CountMismatch { removed, added } => write!(
                formatter,
                "mutation removes {removed} edges but adds {added} edges"
            ),
            Self::DuplicateRemovedEdge => {
                formatter.write_str("mutation contains duplicate removed edges")
            }
            Self::DuplicateAddedEdge => {
                formatter.write_str("mutation contains duplicate added edges")
            }
            Self::RemovedEdgeMissing => {
                formatter.write_str("mutation removes an edge not present in the dataset")
            }
            Self::InvalidAddedEdge => {
                formatter.write_str("mutation adds an invalid or already occupied node pair")
            }
        }
    }
}

impl std::error::Error for MutationError {}
