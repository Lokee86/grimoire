use super::sampling::{Permutation, allocate_counts, directed_pair_in_range, edge, finish_dataset};
use super::{GraphDataset, GraphSpec};

const DENSE_KIND: u16 = 5;
const OTHER_KIND: u16 = 0;
const DENSE_SEED_SALT: u64 = 0xa409_3822_299f_31d0;
const OTHER_SEED_SALT: u64 = 0x082e_fa98_ec4e_6c89;

pub(super) fn generate(spec: &GraphSpec, dense_node_count: u32) -> GraphDataset {
    let sparse_node_count = spec.node_count - dense_node_count;
    let dense_capacity =
        u64::from(dense_node_count) * u64::from(dense_node_count.saturating_sub(1));
    let other_capacity = u64::from(dense_node_count) * u64::from(sparse_node_count) * 2
        + u64::from(sparse_node_count) * u64::from(sparse_node_count.saturating_sub(1));
    let counts = allocate_counts(
        spec.edge_count,
        &[7_000, 3_000],
        &[dense_capacity, other_capacity],
    );

    let mut edges = Vec::with_capacity(spec.edge_count as usize);
    for ordinal in
        Permutation::new(dense_capacity, spec.seed ^ DENSE_SEED_SALT).take(counts[0] as usize)
    {
        let (source, target) = directed_pair_in_range(ordinal, 0, dense_node_count);
        edges.push(edge(source, target, DENSE_KIND));
    }
    for ordinal in
        Permutation::new(other_capacity, spec.seed ^ OTHER_SEED_SALT).take(counts[1] as usize)
    {
        let (source, target) = other_pair(ordinal, dense_node_count, sparse_node_count);
        edges.push(edge(source, target, OTHER_KIND));
    }

    finish_dataset(spec.node_count, spec.edge_count, edges)
}

fn other_pair(ordinal: u64, dense_count: u32, sparse_count: u32) -> (u32, u32) {
    let dense_to_sparse = u64::from(dense_count) * u64::from(sparse_count);
    if ordinal < dense_to_sparse {
        let source = (ordinal / u64::from(sparse_count)) as u32;
        let target = dense_count + (ordinal % u64::from(sparse_count)) as u32;
        return (source, target);
    }

    let ordinal = ordinal - dense_to_sparse;
    if ordinal < dense_to_sparse {
        let source = dense_count + (ordinal / u64::from(dense_count)) as u32;
        let target = (ordinal % u64::from(dense_count)) as u32;
        return (source, target);
    }

    directed_pair_in_range(ordinal - dense_to_sparse, dense_count, sparse_count)
}
