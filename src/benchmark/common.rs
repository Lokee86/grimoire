use std::path::PathBuf;

use crate::synthetic::{GraphDataset, GraphSpec, Topology};

use super::{BenchmarkError, QueryDirection, QueryPattern, QueryWorkload, generate_workload};

const HOT_NODE_COUNT: usize = 16;

/// Inputs for one overlay-versus-rebuild benchmark run.
#[derive(Clone, Debug)]
pub struct BenchmarkConfig {
    pub graph: GraphSpec,
    pub query_count: usize,
    pub sample_count: usize,
    pub work_dir: PathBuf,
    pub keep_files: bool,
}

impl BenchmarkConfig {
    pub fn new(
        graph: GraphSpec,
        query_count: usize,
        sample_count: usize,
        work_dir: impl Into<PathBuf>,
        keep_files: bool,
    ) -> Self {
        Self {
            graph,
            query_count,
            sample_count,
            work_dir: work_dir.into(),
            keep_files,
        }
    }
}

pub(super) fn validate_config(config: &BenchmarkConfig) -> Result<(), BenchmarkError> {
    if config.query_count == 0 {
        return Err(BenchmarkError::InvalidConfig(
            "query_count must be greater than zero",
        ));
    }
    if config.sample_count == 0 {
        return Err(BenchmarkError::InvalidConfig(
            "sample_count must be greater than zero",
        ));
    }
    config.graph.validate()?;
    Ok(())
}

pub(super) fn shared_workloads(
    dataset: &GraphDataset,
    query_count: usize,
    seed: u64,
) -> Result<Vec<NamedWorkload>, BenchmarkError> {
    let definitions = [
        (
            "random-forward",
            QueryPattern::Random,
            QueryDirection::Forward,
        ),
        (
            "random-reverse",
            QueryPattern::Random,
            QueryDirection::Reverse,
        ),
        (
            "sequential-forward",
            QueryPattern::Sequential,
            QueryDirection::Forward,
        ),
        (
            "sequential-reverse",
            QueryPattern::Sequential,
            QueryDirection::Reverse,
        ),
        (
            "hot-forward",
            QueryPattern::HotNodes {
                count: HOT_NODE_COUNT,
            },
            QueryDirection::Forward,
        ),
        (
            "hot-reverse",
            QueryPattern::HotNodes {
                count: HOT_NODE_COUNT,
            },
            QueryDirection::Reverse,
        ),
    ];
    definitions
        .into_iter()
        .enumerate()
        .map(|(index, (name, pattern, direction))| {
            Ok(NamedWorkload {
                name,
                workload: generate_workload(
                    dataset,
                    pattern,
                    direction,
                    query_count,
                    seed.wrapping_add(index as u64),
                )?,
            })
        })
        .collect()
}

pub(super) fn graph_name(spec: GraphSpec) -> String {
    let topology = match spec.topology {
        Topology::Modular { .. } => "modular",
        Topology::Entangled { .. } => "entangled",
        Topology::HubHeavy { .. } => "hub-heavy",
        Topology::Layered { .. } => "layered",
        Topology::DenseSubsystem { .. } => "dense-subsystem",
    };
    format!(
        "{topology}-n{}-e{}-seed{}",
        spec.node_count, spec.edge_count, spec.seed
    )
}

pub(super) struct NamedWorkload {
    pub(super) name: &'static str,
    pub(super) workload: QueryWorkload,
}
