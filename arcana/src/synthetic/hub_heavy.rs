use super::sampling::{
    Permutation, allocate_counts, directed_pair_in_range, edge, finish_dataset,
    hub_incident_capacity, hub_incident_pair,
};
use super::{GraphDataset, GraphSpec};

const HUB_KIND: u16 = 2;
const REGULAR_KIND: u16 = 0;
const HUB_SEED_SALT: u64 = 0x243f_6a88_85a3_08d3;
const REGULAR_SEED_SALT: u64 = 0x1319_8a2e_0370_7344;

pub(super) fn generate(spec: &GraphSpec, hub_count: u32) -> GraphDataset {
    let hub_capacity = hub_incident_capacity(spec.node_count, hub_count);
    let regular_nodes = spec.node_count - hub_count;
    let regular_capacity = u64::from(regular_nodes) * u64::from(regular_nodes.saturating_sub(1));
    let counts = allocate_counts(
        spec.edge_count,
        &[7_000, 3_000],
        &[hub_capacity, regular_capacity],
    );

    let mut edges = Vec::with_capacity(spec.edge_count as usize);
    for ordinal in
        Permutation::new(hub_capacity, spec.seed ^ HUB_SEED_SALT).take(counts[0] as usize)
    {
        let (source, target) = hub_incident_pair(ordinal, spec.node_count, hub_count);
        edges.push(edge(source, target, HUB_KIND));
    }
    for ordinal in
        Permutation::new(regular_capacity, spec.seed ^ REGULAR_SEED_SALT).take(counts[1] as usize)
    {
        let (source, target) = directed_pair_in_range(ordinal, hub_count, regular_nodes);
        edges.push(edge(source, target, REGULAR_KIND));
    }

    finish_dataset(spec.node_count, spec.edge_count, edges)
}
