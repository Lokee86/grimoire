use super::{Edge, EdgeKind, GraphDataset, NodeId};

pub(super) const BASIS_POINTS: u64 = 10_000;

pub(super) fn allocate_counts(total: u64, weights: &[u16], capacities: &[u64]) -> Vec<u64> {
    debug_assert_eq!(weights.len(), capacities.len());
    debug_assert_eq!(
        weights.iter().map(|weight| u64::from(*weight)).sum::<u64>(),
        BASIS_POINTS
    );
    debug_assert!(capacities.iter().sum::<u64>() >= total);

    let mut counts = vec![0; weights.len()];
    let mut requested_total = 0;
    for index in 0..weights.len() {
        let requested = if index + 1 == weights.len() {
            total - requested_total
        } else {
            ((u128::from(total) * u128::from(weights[index])) / u128::from(BASIS_POINTS)) as u64
        };
        requested_total += requested;
        counts[index] = requested.min(capacities[index]);
    }

    let mut remaining = total - counts.iter().sum::<u64>();
    for index in 0..counts.len() {
        let extra = (capacities[index] - counts[index]).min(remaining);
        counts[index] += extra;
        remaining -= extra;
        if remaining == 0 {
            break;
        }
    }

    debug_assert_eq!(remaining, 0);
    counts
}

pub(super) fn finish_dataset(
    node_count: u32,
    expected_edges: u64,
    mut edges: Vec<Edge>,
) -> GraphDataset {
    edges.sort_unstable();
    debug_assert_eq!(edges.len(), expected_edges as usize);
    debug_assert!(edges.windows(2).all(|pair| pair[0] != pair[1]));
    debug_assert!(edges.iter().all(|edge| edge.source != edge.target));
    GraphDataset { node_count, edges }
}

pub(super) fn edge(source: u32, target: u32, kind: u16) -> Edge {
    Edge {
        source: NodeId(source),
        target: NodeId(target),
        kind: EdgeKind(kind),
    }
}

pub(super) fn directed_pair(ordinal: u64, node_count: u32) -> (u32, u32) {
    directed_pair_in_range(ordinal, 0, node_count)
}

pub(super) fn directed_pair_in_range(ordinal: u64, start: u32, node_count: u32) -> (u32, u32) {
    let width = u64::from(node_count - 1);
    let local_source = (ordinal / width) as u32;
    let target_offset = (ordinal % width) as u32;
    let local_target = if target_offset >= local_source {
        target_offset + 1
    } else {
        target_offset
    };
    (start + local_source, start + local_target)
}

pub(super) fn hub_incident_pair(ordinal: u64, node_count: u32, hub_count: u32) -> (u32, u32) {
    let outgoing_capacity = u64::from(hub_count) * u64::from(node_count - 1);
    if ordinal < outgoing_capacity {
        let source = (ordinal / u64::from(node_count - 1)) as u32;
        let target_offset = (ordinal % u64::from(node_count - 1)) as u32;
        let target = if target_offset >= source {
            target_offset + 1
        } else {
            target_offset
        };
        (source, target)
    } else {
        let local = ordinal - outgoing_capacity;
        let source = hub_count + (local / u64::from(hub_count)) as u32;
        let target = (local % u64::from(hub_count)) as u32;
        (source, target)
    }
}

pub(super) fn hub_incident_capacity(node_count: u32, hub_count: u32) -> u64 {
    u64::from(hub_count) * u64::from(node_count - 1)
        + u64::from(node_count - hub_count) * u64::from(hub_count)
}

pub(super) fn modulo_group_ranges(
    start: u32,
    node_count: u32,
    group_count: u32,
) -> Vec<GroupRange> {
    let end = start + node_count;
    let start_remainder = start % group_count;
    let mut edge_start = 0;

    (0..group_count)
        .map(|group| {
            let delta = (group + group_count - start_remainder) % group_count;
            let first = start + delta;
            let size = if first >= end {
                0
            } else {
                (end - 1 - first) / group_count + 1
            };
            let capacity = u64::from(size) * u64::from(size.saturating_sub(1));
            let range = GroupRange {
                first,
                size,
                edge_start,
                edge_end: edge_start + capacity,
            };
            edge_start = range.edge_end;
            range
        })
        .collect()
}

pub(super) fn modulo_group_pair(
    ordinal: u64,
    ranges: &[GroupRange],
    group_count: u32,
) -> (u32, u32) {
    let range_index = ranges.partition_point(|range| range.edge_end <= ordinal);
    let range = &ranges[range_index];
    let local_edge = ordinal - range.edge_start;
    let width = u64::from(range.size - 1);
    let local_source = local_edge / width;
    let target_offset = local_edge % width;
    let local_target = if target_offset >= local_source {
        target_offset + 1
    } else {
        target_offset
    };
    (
        range.first + local_source as u32 * group_count,
        range.first + local_target as u32 * group_count,
    )
}

#[derive(Clone, Copy, Debug)]
pub(super) struct GroupRange {
    pub edge_start: u64,
    pub edge_end: u64,
    first: u32,
    size: u32,
}

pub(super) struct Permutation {
    size: u64,
    index: u64,
    start: u64,
    step: u64,
}

impl Permutation {
    pub(super) fn new(size: u64, seed: u64) -> Self {
        if size <= 1 {
            return Self {
                size,
                index: 0,
                start: 0,
                step: 0,
            };
        }

        let mut rng = DeterministicRng::new(seed);
        let start = rng.next_u64() % size;
        let mut step = rng.next_u64() % size;
        if step == 0 {
            step = 1;
        }
        while greatest_common_divisor(step, size) != 1 {
            step += 1;
            if step == size {
                step = 1;
            }
        }

        Self {
            size,
            index: 0,
            start,
            step,
        }
    }
}

impl Iterator for Permutation {
    type Item = u64;

    fn next(&mut self) -> Option<Self::Item> {
        if self.index >= self.size {
            return None;
        }

        let value = ((u128::from(self.step) * u128::from(self.index) + u128::from(self.start))
            % u128::from(self.size)) as u64;
        self.index += 1;
        Some(value)
    }
}

fn greatest_common_divisor(mut left: u64, mut right: u64) -> u64 {
    while right != 0 {
        let remainder = left % right;
        left = right;
        right = remainder;
    }
    left
}

struct DeterministicRng {
    state: u64,
}

impl DeterministicRng {
    fn new(seed: u64) -> Self {
        Self {
            state: if seed == 0 {
                0xa076_1d64_78bd_642f
            } else {
                seed
            },
        }
    }

    fn next_u64(&mut self) -> u64 {
        let mut value = self.state;
        value ^= value << 13;
        value ^= value >> 7;
        value ^= value << 17;
        self.state = value;
        value
    }
}
