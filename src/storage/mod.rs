//! Packed immutable graph storage and correctness reference structures.

mod dataset;
mod error;
mod format;
mod oracle;
mod reader;
mod sqlite;
mod writer;

#[cfg(test)]
mod corruption_tests;
#[cfg(test)]
mod tests;

pub use error::{DatasetError, Direction, PackedError, QueryError};
pub use oracle::InMemoryGraph;
pub use reader::PackedGraph;
pub use sqlite::{SqliteError, SqliteGraph, SqliteWriteSummary, write_sqlite};
pub use writer::{WriteSummary, write_packed};

pub(crate) use dataset::canonical_edges;
pub(crate) use format::{StableHasher, checksum};

use crate::synthetic::{EdgeKind, NodeId};

/// One adjacent node and the relationship kind connecting it.
#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct Neighbor {
    pub node: NodeId,
    pub kind: EdgeKind,
}
