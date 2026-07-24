use std::fmt;
use std::io;

use crate::snapshot::{OverlayError, SnapshotError};
use crate::storage::{PackedError, QueryError};
use crate::synthetic::{GraphSpecError, MutationError};

use super::workload::QueryWorkloadError;

/// An error while executing or reporting a benchmark.
#[derive(Debug)]
pub enum BenchmarkError {
    Io(io::Error),
    GraphSpec(GraphSpecError),
    Workload(QueryWorkloadError),
    Query(QueryError),
    Packed(PackedError),
    Overlay(OverlayError),
    Snapshot(SnapshotError),
    Mutation(MutationError),
    InvalidConfig(&'static str),
    MutationMismatch {
        sample: u32,
        workload: String,
        overlay_items: u64,
        rebuilt_items: u64,
        overlay_fingerprint: u64,
        rebuilt_fingerprint: u64,
    },
}

impl fmt::Display for BenchmarkError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::GraphSpec(error) => error.fmt(formatter),
            Self::Workload(error) => error.fmt(formatter),
            Self::Query(error) => error.fmt(formatter),
            Self::Packed(error) => error.fmt(formatter),
            Self::Overlay(error) => error.fmt(formatter),
            Self::Snapshot(error) => error.fmt(formatter),
            Self::Mutation(error) => error.fmt(formatter),
            Self::InvalidConfig(message) => formatter.write_str(message),
            Self::MutationMismatch {
                sample,
                workload,
                overlay_items,
                rebuilt_items,
                overlay_fingerprint,
                rebuilt_fingerprint,
            } => write!(
                formatter,
                "mutation mismatch in sample {sample} workload {workload}: \
                 overlay items={overlay_items} fingerprint={overlay_fingerprint:#x}; \
                 rebuilt packed items={rebuilt_items} fingerprint={rebuilt_fingerprint:#x}"
            ),
        }
    }
}

impl std::error::Error for BenchmarkError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            Self::GraphSpec(error) => Some(error),
            Self::Workload(error) => Some(error),
            Self::Query(error) => Some(error),
            Self::Packed(error) => Some(error),
            Self::Overlay(error) => Some(error),
            Self::Snapshot(error) => Some(error),
            Self::Mutation(error) => Some(error),
            Self::InvalidConfig(_) | Self::MutationMismatch { .. } => None,
        }
    }
}

impl From<io::Error> for BenchmarkError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}

impl From<GraphSpecError> for BenchmarkError {
    fn from(error: GraphSpecError) -> Self {
        Self::GraphSpec(error)
    }
}

impl From<QueryWorkloadError> for BenchmarkError {
    fn from(error: QueryWorkloadError) -> Self {
        Self::Workload(error)
    }
}

impl From<QueryError> for BenchmarkError {
    fn from(error: QueryError) -> Self {
        Self::Query(error)
    }
}

impl From<PackedError> for BenchmarkError {
    fn from(error: PackedError) -> Self {
        Self::Packed(error)
    }
}

impl From<OverlayError> for BenchmarkError {
    fn from(error: OverlayError) -> Self {
        Self::Overlay(error)
    }
}

impl From<SnapshotError> for BenchmarkError {
    fn from(error: SnapshotError) -> Self {
        Self::Snapshot(error)
    }
}

impl From<MutationError> for BenchmarkError {
    fn from(error: MutationError) -> Self {
        Self::Mutation(error)
    }
}
