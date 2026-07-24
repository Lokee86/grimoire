use std::fs;
use std::path::Path;
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::Instant;

use crate::snapshot::{GraphSnapshot, OverlayChanges, publish_snapshot, write_overlay};
use crate::storage::{PackedGraph, write_packed};
use crate::synthetic::{GraphDataset, generate};

use super::common::{NamedWorkload, graph_name, shared_workloads, validate_config};
use super::mutation_files::{GeneratedFiles, mutation_path};
use super::mutation_plan::{NamedMutation, standard_mutations};
use super::mutation_query::measure_queries;
use super::{
    Backend, BenchmarkConfig, BenchmarkError, BenchmarkMetric, BenchmarkReport, BenchmarkSample,
};

static RUN_SEQUENCE: AtomicU64 = AtomicU64::new(0);

pub fn run_mutation_benchmark(config: &BenchmarkConfig) -> Result<BenchmarkReport, BenchmarkError> {
    validate_config(config)?;
    fs::create_dir_all(&config.work_dir)?;
    let dataset = generate(&config.graph)?;
    let run_id = RUN_SEQUENCE.fetch_add(1, Ordering::Relaxed);
    let base_path = mutation_path(&config.work_dir, run_id, "base.pack");
    let mut files = GeneratedFiles::new(config.keep_files);
    files.push(base_path.clone());
    write_packed(&base_path, &dataset)?;
    let base = PackedGraph::open(&base_path)?;
    let graph = graph_name(config.graph);
    let mutations = standard_mutations(&dataset, config.graph.seed)?;
    let mut report = BenchmarkReport::new();

    for mutation in mutations {
        let workloads = shared_workloads(
            &mutation.visible,
            config.query_count,
            config.graph.seed ^ mutation.seed_salt,
        )?;
        for sample_index in 0..config.sample_count {
            let sample = u32::try_from(sample_index)
                .map_err(|_| BenchmarkError::InvalidConfig("sample_count exceeds u32"))?;
            run_sample(
                config,
                run_id,
                sample,
                &graph,
                &base_path,
                &base,
                &mutation,
                &workloads,
                &mut files,
                &mut report,
            )?;
        }
    }
    Ok(report)
}

#[allow(clippy::too_many_arguments)]
fn run_sample(
    config: &BenchmarkConfig,
    run_id: u64,
    sample: u32,
    graph: &str,
    base_path: &Path,
    base: &PackedGraph,
    mutation: &NamedMutation,
    workloads: &[NamedWorkload],
    files: &mut GeneratedFiles,
    report: &mut BenchmarkReport,
) -> Result<(), BenchmarkError> {
    let prefix = format!("{}-{sample}", mutation.name);
    let overlay_path = mutation_path(&config.work_dir, run_id, &format!("{prefix}.overlay"));
    let manifest_path = mutation_path(&config.work_dir, run_id, &format!("{prefix}.manifest"));
    let rebuilt_path = mutation_path(&config.work_dir, run_id, &format!("{prefix}.pack"));
    files.extend([
        overlay_path.clone(),
        manifest_path.clone(),
        rebuilt_path.clone(),
    ]);
    let changes = OverlayChanges {
        added: mutation.mutation.added.clone(),
        removed: mutation.mutation.removed.clone(),
    };

    let (overlay_summary, rebuilt_summary) = if sample.is_multiple_of(2) {
        let overlay = measure_overlay_write(
            graph,
            mutation.name,
            sample,
            base,
            &overlay_path,
            &changes,
            report,
        )?;
        let rebuilt = measure_rebuild(
            graph,
            mutation.name,
            sample,
            &rebuilt_path,
            &mutation.visible,
            report,
        )?;
        (overlay, rebuilt)
    } else {
        let rebuilt = measure_rebuild(
            graph,
            mutation.name,
            sample,
            &rebuilt_path,
            &mutation.visible,
            report,
        )?;
        let overlay = measure_overlay_write(
            graph,
            mutation.name,
            sample,
            base,
            &overlay_path,
            &changes,
            report,
        )?;
        (overlay, rebuilt)
    };
    if overlay_summary.visible_dataset_checksum != rebuilt_summary.dataset_checksum {
        return Err(BenchmarkError::MutationMismatch {
            sample,
            workload: mutation.name.to_owned(),
            overlay_items: overlay_summary.visible_edge_count,
            rebuilt_items: rebuilt_summary.edge_count,
            overlay_fingerprint: overlay_summary.visible_dataset_checksum,
            rebuilt_fingerprint: rebuilt_summary.dataset_checksum,
        });
    }

    publish_snapshot(
        &manifest_path,
        base_path.file_name().expect("base path has a name"),
        Some(Path::new(
            overlay_path.file_name().expect("overlay path has a name"),
        )),
        0,
    )?;
    let started = Instant::now();
    let snapshot = GraphSnapshot::open(&manifest_path)?;
    report.push(BenchmarkSample::new(
        graph,
        Backend::Overlay,
        BenchmarkMetric::Reopen,
        mutation.name,
        sample,
        started.elapsed(),
        snapshot.edge_count(),
        u64::from(snapshot.node_count()),
        overlay_summary.file_len,
        snapshot.dataset_checksum(),
    ));
    let started = Instant::now();
    let rebuilt = PackedGraph::open(&rebuilt_path)?;
    report.push(BenchmarkSample::new(
        graph,
        Backend::RebuiltPacked,
        BenchmarkMetric::Reopen,
        mutation.name,
        sample,
        started.elapsed(),
        rebuilt.edge_count(),
        u64::from(rebuilt.node_count()),
        rebuilt_summary.file_len,
        rebuilt.dataset_checksum(),
    ));
    measure_queries(
        graph,
        mutation.name,
        sample,
        workloads,
        &snapshot,
        &rebuilt,
        overlay_summary.file_len,
        rebuilt_summary.file_len,
        report,
    )
}

fn measure_overlay_write(
    graph: &str,
    workload: &str,
    sample: u32,
    base: &PackedGraph,
    path: &Path,
    changes: &OverlayChanges,
    report: &mut BenchmarkReport,
) -> Result<crate::snapshot::OverlayWriteSummary, BenchmarkError> {
    let started = Instant::now();
    let summary = write_overlay(path, base, changes)?;
    report.push(BenchmarkSample::new(
        graph,
        Backend::Overlay,
        BenchmarkMetric::Mutation,
        workload,
        sample,
        started.elapsed(),
        summary.added_count + summary.removed_count,
        summary.visible_edge_count,
        summary.file_len,
        summary.visible_dataset_checksum,
    ));
    Ok(summary)
}

fn measure_rebuild(
    graph: &str,
    workload: &str,
    sample: u32,
    path: &Path,
    visible: &GraphDataset,
    report: &mut BenchmarkReport,
) -> Result<crate::storage::WriteSummary, BenchmarkError> {
    let started = Instant::now();
    let summary = write_packed(path, visible)?;
    report.push(BenchmarkSample::new(
        graph,
        Backend::RebuiltPacked,
        BenchmarkMetric::Mutation,
        workload,
        sample,
        started.elapsed(),
        summary.edge_count,
        summary.edge_count,
        summary.file_len,
        summary.dataset_checksum,
    ));
    Ok(summary)
}
