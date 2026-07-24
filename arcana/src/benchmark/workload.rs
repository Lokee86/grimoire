use std::fmt;

use crate::synthetic::{GraphDataset, NodeId};

/// Adjacency direction used by a benchmark query workload.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum QueryDirection {
    Forward,
    Reverse,
}

/// Node-selection pattern used by a benchmark query workload.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum QueryPattern {
    Random,
    Sequential,
    HotNodes { count: usize },
}

/// A deterministic sequence of node queries with shared benchmark metadata.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct QueryWorkload {
    pub direction: QueryDirection,
    pub pattern: QueryPattern,
    pub nodes: Vec<NodeId>,
}

impl QueryWorkload {
    /// Builds the same node sequence for either query direction.
    pub fn generate(
        dataset: &GraphDataset,
        pattern: QueryPattern,
        direction: QueryDirection,
        query_count: usize,
        seed: u64,
    ) -> Result<Self, QueryWorkloadError> {
        if query_count > 0 && dataset.node_count == 0 {
            return Err(QueryWorkloadError::ZeroNodes);
        }

        let nodes = match pattern {
            QueryPattern::Random => random_nodes(dataset.node_count, query_count, seed),
            QueryPattern::Sequential => sequential_nodes(dataset.node_count, query_count),
            QueryPattern::HotNodes { count } => {
                if count == 0 && query_count > 0 {
                    return Err(QueryWorkloadError::ZeroHotNodes);
                }
                let hot_nodes = highest_degree_nodes(dataset, count);
                sampled_nodes(&hot_nodes, query_count, seed)
            }
        };

        Ok(Self {
            direction,
            pattern,
            nodes,
        })
    }

    pub fn len(&self) -> usize {
        self.nodes.len()
    }

    pub const fn is_empty(&self) -> bool {
        self.nodes.is_empty()
    }

    pub fn node_ids(&self) -> &[NodeId] {
        &self.nodes
    }
}

/// Generates a shared deterministic query workload for a graph dataset.
pub fn generate_workload(
    dataset: &GraphDataset,
    pattern: QueryPattern,
    direction: QueryDirection,
    query_count: usize,
    seed: u64,
) -> Result<QueryWorkload, QueryWorkloadError> {
    QueryWorkload::generate(dataset, pattern, direction, query_count, seed)
}

fn random_nodes(node_count: u32, query_count: usize, seed: u64) -> Vec<NodeId> {
    let mut state = seed;
    (0..query_count)
        .map(|_| NodeId(next_index(&mut state, node_count)))
        .collect()
}

fn sequential_nodes(node_count: u32, query_count: usize) -> Vec<NodeId> {
    (0..query_count)
        .map(|index| NodeId((index as u64 % u64::from(node_count)) as u32))
        .collect()
}

fn sampled_nodes(candidates: &[NodeId], query_count: usize, seed: u64) -> Vec<NodeId> {
    let mut state = seed;
    (0..query_count)
        .map(|_| candidates[next_index(&mut state, candidates.len() as u32) as usize])
        .collect()
}

fn highest_degree_nodes(dataset: &GraphDataset, requested_count: usize) -> Vec<NodeId> {
    let mut degrees = vec![0usize; dataset.node_count as usize];
    for edge in &dataset.edges {
        if let Some(degree) = degrees.get_mut(edge.source.0 as usize) {
            *degree += 1;
        }
        if let Some(degree) = degrees.get_mut(edge.target.0 as usize) {
            *degree += 1;
        }
    }

    let mut nodes: Vec<_> = (0..dataset.node_count)
        .map(|node| (NodeId(node), degrees[node as usize]))
        .collect();
    nodes.sort_unstable_by(|(left_node, left_degree), (right_node, right_degree)| {
        right_degree
            .cmp(left_degree)
            .then_with(|| left_node.cmp(right_node))
    });
    nodes
        .into_iter()
        .take(requested_count.min(dataset.node_count as usize))
        .map(|(node, _)| node)
        .collect()
}

fn next_index(state: &mut u64, upper_bound: u32) -> u32 {
    *state = state.wrapping_add(0x9e37_79b9_7f4a_7c15);
    let mut value = *state;
    value = (value ^ (value >> 30)).wrapping_mul(0xbf58_476d_1ce4_e5b9);
    value = (value ^ (value >> 27)).wrapping_mul(0x94d0_49bb_1331_11eb);
    value ^= value >> 31;
    (value % u64::from(upper_bound)) as u32
}

/// A workload request that cannot produce valid node queries.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum QueryWorkloadError {
    ZeroNodes,
    ZeroHotNodes,
}

impl fmt::Display for QueryWorkloadError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::ZeroNodes => {
                formatter.write_str("a non-empty workload requires at least one node")
            }
            Self::ZeroHotNodes => {
                formatter.write_str("a hot-node workload requires at least one hot node")
            }
        }
    }
}

impl std::error::Error for QueryWorkloadError {}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::synthetic::{Edge, EdgeKind};

    fn dataset() -> GraphDataset {
        GraphDataset {
            node_count: 8,
            edges: vec![
                Edge {
                    source: NodeId(0),
                    target: NodeId(1),
                    kind: EdgeKind(0),
                },
                Edge {
                    source: NodeId(0),
                    target: NodeId(2),
                    kind: EdgeKind(0),
                },
                Edge {
                    source: NodeId(3),
                    target: NodeId(0),
                    kind: EdgeKind(0),
                },
                Edge {
                    source: NodeId(4),
                    target: NodeId(0),
                    kind: EdgeKind(0),
                },
                Edge {
                    source: NodeId(5),
                    target: NodeId(6),
                    kind: EdgeKind(0),
                },
            ],
        }
    }

    #[test]
    fn all_six_combinations_are_deterministic_and_valid() {
        let patterns = [
            QueryPattern::Random,
            QueryPattern::Sequential,
            QueryPattern::HotNodes { count: 2 },
        ];

        for pattern in patterns {
            let forward = generate_workload(&dataset(), pattern, QueryDirection::Forward, 31, 17)
                .expect("workload should be valid");
            let reverse = generate_workload(&dataset(), pattern, QueryDirection::Reverse, 31, 17)
                .expect("workload should be valid");
            let repeat = generate_workload(&dataset(), pattern, QueryDirection::Forward, 31, 17)
                .expect("workload should be valid");

            assert_eq!(forward.len(), 31);
            assert_eq!(forward.nodes, reverse.nodes);
            assert_eq!(forward, repeat);
            assert!(forward.nodes.iter().all(|node| node.0 < 8));
        }
    }

    #[test]
    fn sequential_queries_wrap_at_the_node_count() {
        let workload = generate_workload(
            &dataset(),
            QueryPattern::Sequential,
            QueryDirection::Forward,
            10,
            0,
        )
        .expect("workload should be valid");

        assert_eq!(
            workload.nodes,
            (0..10).map(|node| NodeId(node % 8)).collect::<Vec<_>>()
        );
    }

    #[test]
    fn hot_nodes_are_selected_by_total_degree() {
        let workload = generate_workload(
            &dataset(),
            QueryPattern::HotNodes { count: 1 },
            QueryDirection::Reverse,
            20,
            23,
        )
        .expect("workload should be valid");

        assert!(workload.nodes.iter().all(|&node| node == NodeId(0)));
    }

    #[test]
    fn random_workloads_change_with_seed() {
        let first = generate_workload(
            &dataset(),
            QueryPattern::Random,
            QueryDirection::Forward,
            12,
            1,
        )
        .expect("workload should be valid");
        let second = generate_workload(
            &dataset(),
            QueryPattern::Random,
            QueryDirection::Forward,
            12,
            2,
        )
        .expect("workload should be valid");

        assert_ne!(first.nodes, second.nodes);
    }
}
