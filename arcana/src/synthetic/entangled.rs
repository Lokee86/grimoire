use super::sampling::{
    Permutation, allocate_counts, directed_pair_in_range, edge, finish_dataset,
    hub_incident_capacity, hub_incident_pair, modulo_group_pair, modulo_group_ranges,
};
use super::{GraphDataset, GraphSpec};

const HUB_KIND: u16 = 2;
const CROSS_KIND: u16 = 1;
const WITHIN_KIND: u16 = 0;
const HUB_SEED_SALT: u64 = 0x3bd3_9e10_cb0e_f593;
const CROSS_SEED_SALT: u64 = 0xc0ac_29b7_c97c_50dd;
const WITHIN_SEED_SALT: u64 = 0x3f84_d5b5_b547_0917;

pub(super) fn generate(spec: &GraphSpec, cluster_count: u32, hub_count: u32) -> GraphDataset {
    let hub_capacity = hub_incident_capacity(spec.node_count, hub_count);
    let regular_count = spec.node_count - hub_count;
    let ranges = modulo_group_ranges(hub_count, regular_count, cluster_count);
    let within_capacity = ranges.last().map_or(0, |range| range.edge_end);
    let regular_capacity = u64::from(regular_count) * u64::from(regular_count.saturating_sub(1));
    let cross_capacity = regular_capacity - within_capacity;
    let counts = allocate_counts(
        spec.edge_count,
        &[4_000, 4_500, 1_500],
        &[hub_capacity, cross_capacity, within_capacity],
    );

    let mut edges = Vec::with_capacity(spec.edge_count as usize);
    for ordinal in
        Permutation::new(hub_capacity, spec.seed ^ HUB_SEED_SALT).take(counts[0] as usize)
    {
        let (source, target) = hub_incident_pair(ordinal, spec.node_count, hub_count);
        edges.push(edge(source, target, HUB_KIND));
    }

    if counts[1] > 0 {
        let mut generated = 0;
        for ordinal in Permutation::new(regular_capacity, spec.seed ^ CROSS_SEED_SALT) {
            let (source, target) = directed_pair_in_range(ordinal, hub_count, regular_count);
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

    for ordinal in
        Permutation::new(within_capacity, spec.seed ^ WITHIN_SEED_SALT).take(counts[2] as usize)
    {
        let (source, target) = modulo_group_pair(ordinal, &ranges, cluster_count);
        edges.push(edge(source, target, WITHIN_KIND));
    }

    finish_dataset(spec.node_count, spec.edge_count, edges)
}
