use crate::synthetic::{
    GraphDataset, GraphMutation, MutationScenario, NodeId, apply_mutation, plan_mutation,
};

use super::BenchmarkError;

pub(super) struct NamedMutation {
    pub name: &'static str,
    pub seed_salt: u64,
    pub mutation: GraphMutation,
    pub visible: GraphDataset,
}

pub(super) fn standard_mutations(
    dataset: &GraphDataset,
    seed: u64,
) -> Result<Vec<NamedMutation>, BenchmarkError> {
    let target = (dataset.edges.len() as u64 / 1_000).max(1);
    let hot_node = hottest_node(dataset);
    let local_count = (dataset.node_count / 100).max(1);
    let local_start = hot_node.0.min(dataset.node_count - local_count);
    let local_candidates = incident_count(dataset, local_start, local_start + local_count);
    let hub_count = (dataset.node_count / 1_000)
        .max(1)
        .max(hot_node.0.saturating_add(1))
        .min(dataset.node_count);
    let hub_candidates = incident_count(dataset, 0, hub_count);
    let definitions = [
        (
            "single-node",
            0x11,
            MutationScenario::SingleNode { node: hot_node },
        ),
        (
            "local-range",
            0x22,
            MutationScenario::LocalRange {
                start: NodeId(local_start),
                node_count: local_count,
                edge_count: target.min(local_candidates).max(1),
            },
        ),
        (
            "scattered",
            0x33,
            MutationScenario::Scattered {
                edge_count: target.min(dataset.edges.len() as u64),
            },
        ),
        (
            "hub",
            0x44,
            MutationScenario::Hub {
                hub_count,
                edge_count: target.min(hub_candidates).max(1),
            },
        ),
        (
            "percentage-1pct",
            0x55,
            MutationScenario::Percentage { basis_points: 100 },
        ),
    ];

    definitions
        .into_iter()
        .map(|(name, seed_salt, scenario)| {
            let mutation = plan_mutation(dataset, scenario, seed ^ seed_salt)?;
            let visible = apply_mutation(dataset, &mutation)?;
            Ok(NamedMutation {
                name,
                seed_salt,
                mutation,
                visible,
            })
        })
        .collect()
}

fn hottest_node(dataset: &GraphDataset) -> NodeId {
    let mut degree = vec![0_u64; dataset.node_count as usize];
    for edge in &dataset.edges {
        degree[edge.source.0 as usize] += 1;
        degree[edge.target.0 as usize] += 1;
    }
    NodeId(
        degree
            .iter()
            .enumerate()
            .max_by_key(|&(node, count)| (*count, std::cmp::Reverse(node)))
            .map_or(0, |(node, _)| node as u32),
    )
}

fn incident_count(dataset: &GraphDataset, start: u32, end: u32) -> u64 {
    dataset
        .edges
        .iter()
        .filter(|edge| {
            (start..end).contains(&edge.source.0) || (start..end).contains(&edge.target.0)
        })
        .count() as u64
}
