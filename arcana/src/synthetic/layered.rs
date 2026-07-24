use super::sampling::{Permutation, allocate_counts, directed_pair, edge, finish_dataset};
use super::{GraphDataset, GraphSpec};

const FORWARD_KIND: u16 = 3;
const OTHER_KIND: u16 = 4;
const FORWARD_SEED_SALT: u64 = 0x4528_21e6_38d0_1377;
const OTHER_SEED_SALT: u64 = 0xbe54_66cf_34e9_0c6c;

pub(super) fn generate(spec: &GraphSpec, layer_count: u32) -> GraphDataset {
    let spans = adjacent_spans(spec.node_count, layer_count);
    let forward_capacity = spans.last().map_or(0, |span| span.edge_end);
    let total_capacity = u64::from(spec.node_count) * u64::from(spec.node_count - 1);
    let other_capacity = total_capacity - forward_capacity;
    let counts = allocate_counts(
        spec.edge_count,
        &[8_000, 2_000],
        &[forward_capacity, other_capacity],
    );

    let mut edges = Vec::with_capacity(spec.edge_count as usize);
    for ordinal in
        Permutation::new(forward_capacity, spec.seed ^ FORWARD_SEED_SALT).take(counts[0] as usize)
    {
        let (source, target) = adjacent_pair(ordinal, &spans, layer_count);
        edges.push(edge(source, target, FORWARD_KIND));
    }

    if counts[1] > 0 {
        let mut generated = 0;
        for ordinal in Permutation::new(total_capacity, spec.seed ^ OTHER_SEED_SALT) {
            let (source, target) = directed_pair(ordinal, spec.node_count);
            if target % layer_count == source % layer_count + 1 {
                continue;
            }
            edges.push(edge(source, target, OTHER_KIND));
            generated += 1;
            if generated == counts[1] {
                break;
            }
        }
    }

    finish_dataset(spec.node_count, spec.edge_count, edges)
}

fn adjacent_spans(node_count: u32, layer_count: u32) -> Vec<LayerSpan> {
    let sizes: Vec<u32> = (0..layer_count)
        .map(|layer| (node_count - 1 - layer) / layer_count + 1)
        .collect();
    let mut edge_start = 0;
    (0..layer_count - 1)
        .map(|layer| {
            let capacity = u64::from(sizes[layer as usize]) * u64::from(sizes[layer as usize + 1]);
            let span = LayerSpan {
                layer,
                source_size: sizes[layer as usize],
                target_size: sizes[layer as usize + 1],
                edge_start,
                edge_end: edge_start + capacity,
            };
            edge_start = span.edge_end;
            span
        })
        .collect()
}

fn adjacent_pair(ordinal: u64, spans: &[LayerSpan], layer_count: u32) -> (u32, u32) {
    let span_index = spans.partition_point(|span| span.edge_end <= ordinal);
    let span = &spans[span_index];
    let local = ordinal - span.edge_start;
    let source_local = local / u64::from(span.target_size);
    let target_local = local % u64::from(span.target_size);
    debug_assert!(source_local < u64::from(span.source_size));
    (
        span.layer + source_local as u32 * layer_count,
        span.layer + 1 + target_local as u32 * layer_count,
    )
}

struct LayerSpan {
    layer: u32,
    source_size: u32,
    target_size: u32,
    edge_start: u64,
    edge_end: u64,
}
