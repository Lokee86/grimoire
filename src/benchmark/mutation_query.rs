use std::time::Instant;

use crate::snapshot::GraphSnapshot;
use crate::storage::{Neighbor, PackedGraph};
use crate::synthetic::NodeId;

use super::common::NamedWorkload;
use super::{
    Backend, BenchmarkError, BenchmarkMetric, BenchmarkReport, BenchmarkSample, QueryDirection,
    QueryWorkload,
};

#[allow(clippy::too_many_arguments)]
pub(super) fn measure_queries(
    graph_name: &str,
    scenario: &str,
    sample: u32,
    workloads: &[NamedWorkload],
    snapshot: &GraphSnapshot,
    rebuilt: &PackedGraph,
    overlay_size: u64,
    rebuilt_size: u64,
    report: &mut BenchmarkReport,
) -> Result<(), BenchmarkError> {
    for (index, named) in workloads.iter().enumerate() {
        let workload_name = format!("{scenario}:{}", named.name);
        std::hint::black_box(execute_snapshot(snapshot, &named.workload)?);
        std::hint::black_box(execute_packed(rebuilt, &named.workload)?);
        let overlay_first = (sample as usize + index).is_multiple_of(2);
        let (overlay, packed) = if overlay_first {
            let overlay = timed_snapshot(
                graph_name,
                &workload_name,
                sample,
                snapshot,
                &named.workload,
                overlay_size,
                report,
            )?;
            let packed = timed_packed(
                graph_name,
                &workload_name,
                sample,
                rebuilt,
                &named.workload,
                rebuilt_size,
                report,
            )?;
            (overlay, packed)
        } else {
            let packed = timed_packed(
                graph_name,
                &workload_name,
                sample,
                rebuilt,
                &named.workload,
                rebuilt_size,
                report,
            )?;
            let overlay = timed_snapshot(
                graph_name,
                &workload_name,
                sample,
                snapshot,
                &named.workload,
                overlay_size,
                report,
            )?;
            (overlay, packed)
        };
        if overlay != packed {
            return Err(BenchmarkError::MutationMismatch {
                sample,
                workload: workload_name,
                overlay_items: overlay.0,
                rebuilt_items: packed.0,
                overlay_fingerprint: overlay.1,
                rebuilt_fingerprint: packed.1,
            });
        }
    }
    Ok(())
}

fn timed_snapshot(
    graph_name: &str,
    workload_name: &str,
    sample: u32,
    graph: &GraphSnapshot,
    workload: &QueryWorkload,
    file_size: u64,
    report: &mut BenchmarkReport,
) -> Result<(u64, u64), BenchmarkError> {
    let started = Instant::now();
    let result = execute_snapshot(graph, workload)?;
    report.push(BenchmarkSample::new(
        graph_name,
        Backend::Overlay,
        BenchmarkMetric::Query,
        workload_name,
        sample,
        started.elapsed(),
        workload.len() as u64,
        result.0,
        file_size,
        result.1,
    ));
    Ok(result)
}

fn timed_packed(
    graph_name: &str,
    workload_name: &str,
    sample: u32,
    graph: &PackedGraph,
    workload: &QueryWorkload,
    file_size: u64,
    report: &mut BenchmarkReport,
) -> Result<(u64, u64), BenchmarkError> {
    let started = Instant::now();
    let result = execute_packed(graph, workload)?;
    report.push(BenchmarkSample::new(
        graph_name,
        Backend::RebuiltPacked,
        BenchmarkMetric::Query,
        workload_name,
        sample,
        started.elapsed(),
        workload.len() as u64,
        result.0,
        file_size,
        result.1,
    ));
    Ok(result)
}

fn execute_snapshot(
    graph: &GraphSnapshot,
    workload: &QueryWorkload,
) -> Result<(u64, u64), BenchmarkError> {
    execute_workload(workload, |node, direction| match direction {
        QueryDirection::Forward => Ok(graph.forward_neighbors(node)?),
        QueryDirection::Reverse => Ok(graph.reverse_neighbors(node)?),
    })
}

fn execute_packed(
    graph: &PackedGraph,
    workload: &QueryWorkload,
) -> Result<(u64, u64), BenchmarkError> {
    execute_workload(workload, |node, direction| match direction {
        QueryDirection::Forward => Ok(graph.forward_neighbors(node)?),
        QueryDirection::Reverse => Ok(graph.reverse_neighbors(node)?),
    })
}

fn execute_workload<F>(
    workload: &QueryWorkload,
    mut neighbors: F,
) -> Result<(u64, u64), BenchmarkError>
where
    F: FnMut(NodeId, QueryDirection) -> Result<Vec<Neighbor>, BenchmarkError>,
{
    let mut items = 0_u64;
    let mut fingerprint = 0xcbf2_9ce4_8422_2325_u64;
    for &node in workload.node_ids() {
        let adjacent = neighbors(node, workload.direction)?;
        fingerprint = mix(fingerprint, u64::from(node.0));
        fingerprint = mix(fingerprint, adjacent.len() as u64);
        for neighbor in adjacent {
            items += 1;
            fingerprint = mix(
                mix(fingerprint, u64::from(neighbor.node.0)),
                u64::from(neighbor.kind.0),
            );
        }
    }
    Ok((items, std::hint::black_box(fingerprint)))
}

fn mix(value: u64, input: u64) -> u64 {
    (value ^ input.wrapping_add(0x9e37_79b9_7f4a_7c15))
        .wrapping_mul(0x1000_0000_01b3)
        .rotate_left(13)
}
