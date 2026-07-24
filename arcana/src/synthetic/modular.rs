use super::sampling::{
    Permutation, allocate_counts, directed_pair, edge, finish_dataset, modulo_group_pair,
    modulo_group_ranges,
};
use super::{GraphDataset, GraphSpec};

const WITHIN_KIND: u16 = 0;
const CROSS_KIND: u16 = 1;
const WITHIN_SEED_SALT: u64 = 0x9e37_79b9_7f4a_7c15;
const CROSS_SEED_SALT: u64 = 0xd1b5_4a32_d192_ed03;

pub(super) fn generate(
    spec: &GraphSpec,
    cluster_count: u32,
    cross_cluster_ratio: u16,
) -> GraphDataset {
    let ranges = modulo_group_ranges(0, spec.node_count, cluster_count);
    let within_capacity = ranges.last().map_or(0, |range| range.edge_end);
    let total_capacity = u64::from(spec.node_count) * u64::from(spec.node_count - 1);
    let cross_capacity = total_capacity - within_capacity;
    let counts = allocate_counts(
        spec.edge_count,
        &[10_000 - cross_cluster_ratio, cross_cluster_ratio],
        &[within_capacity, cross_capacity],
    );

    let mut edges = Vec::with_capacity(spec.edge_count as usize);
    for ordinal in
        Permutation::new(within_capacity, spec.seed ^ WITHIN_SEED_SALT).take(counts[0] as usize)
    {
        let (source, target) = modulo_group_pair(ordinal, &ranges, cluster_count);
        edges.push(edge(source, target, WITHIN_KIND));
    }

    if counts[1] > 0 {
        let mut generated = 0;
        for ordinal in Permutation::new(total_capacity, spec.seed ^ CROSS_SEED_SALT) {
            let (source, target) = directed_pair(ordinal, spec.node_count);
            if source % cluster_count == target % cluster_count {
                continue;
            }
            edges.push(edge(source, target, CROSS_KIND));
            generated += 1;
            if generated == counts[1] {
                break;
            }
        }
    }

    finish_dataset(spec.node_count, spec.edge_count, edges)
}

#[cfg(test)]
mod tests {
    use super::generate;
    use crate::synthetic::{GraphSpec, Topology};

    #[test]
    fn requested_extreme_ratios_are_used_when_capacity_allows() {
        let base = GraphSpec {
            topology: Topology::Modular {
                cluster_count: 2,
                cross_cluster_ratio: 0,
            },
            node_count: 8,
            edge_count: 12,
            seed: 3,
        };
        let within = generate(&base, 2, 0);
        assert!(within.edges.iter().all(|edge| edge.kind.0 == 0));

        let cross = generate(&base, 2, 10_000);
        assert!(cross.edges.iter().all(|edge| edge.kind.0 == 1));
    }

    #[test]
    fn large_sparse_generation_does_not_enumerate_pair_capacity() {
        let spec = GraphSpec {
            topology: Topology::Modular {
                cluster_count: 1_000,
                cross_cluster_ratio: 2_500,
            },
            node_count: 1_000_000,
            edge_count: 10_000,
            seed: 17,
        };
        assert_eq!(generate(&spec, 1_000, 2_500).edges.len(), 10_000);
    }
}
