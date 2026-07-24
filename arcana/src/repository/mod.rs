//! Language-neutral repository facts and their deterministic text format.

mod catalogue;
mod compiler;
#[cfg(test)]
mod compiler_catalogue_tests;
mod fact_file;
mod fact_file_error;
#[cfg(test)]
mod fact_file_tests;
mod incremental;
#[cfg(test)]
mod incremental_tests;
#[cfg(test)]
mod lexicon_fact_file_tests;
mod model;
mod ownership;
#[cfg(test)]
mod ownership_tests;
mod path;
mod repository_snapshot;
mod repository_snapshot_error;
mod repository_snapshot_format;
#[cfg(test)]
mod repository_snapshot_tests;
mod repository_snapshot_validation;
mod unresolved;

pub use catalogue::{
    CatalogueEntry, CatalogueError, RepositoryCatalogue, read_catalogue, write_catalogue,
};
pub use compiler::{
    CompiledRepository, RepositoryCompileError, compile_facts, compile_repository_facts,
    edge_kind_to_relation, relation_to_edge_kind,
};
pub use fact_file::{FACT_SCHEMA_VERSION, encode_facts, parse_facts};
pub use fact_file_error::FactFileError;
pub use incremental::{IncrementalError, IncrementalUpdate, plan_file_update};
pub use model::{ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, SourceSpan};
pub use ownership::{FactOwnershipError, FactPartitions, partition_facts, replace_changed_files};
pub use path::{RepositoryPathError, normalize_repository_path};
pub use repository_snapshot::{
    PublishRepositorySnapshot, REPOSITORY_MANIFEST_FILE, RepositorySnapshot,
    derive_repository_snapshot_id, publish_repository_snapshot,
};
pub use repository_snapshot_error::RepositorySnapshotError;
pub use repository_snapshot_format::{REPOSITORY_MANIFEST_VERSION, RepositorySnapshotManifest};
pub use unresolved::{UnresolvedReason, UnresolvedReferenceFact};

/// The complete set of facts extracted from one repository.
#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct RepositoryFacts {
    pub nodes: Vec<NodeFact>,
    pub edges: Vec<EdgeFact>,
    pub unresolved: Vec<UnresolvedReferenceFact>,
}

impl RepositoryFacts {
    pub fn new(nodes: Vec<NodeFact>, edges: Vec<EdgeFact>) -> Self {
        Self {
            nodes,
            edges,
            unresolved: Vec::new(),
        }
    }

    pub fn with_unresolved(
        nodes: Vec<NodeFact>,
        edges: Vec<EdgeFact>,
        unresolved: Vec<UnresolvedReferenceFact>,
    ) -> Self {
        Self {
            nodes,
            edges,
            unresolved,
        }
    }

    /// Returns the canonical tab-separated representation of these facts.
    pub fn encode(&self) -> String {
        encode_facts(self)
    }

    /// Parses a canonical tab-separated fact file.
    pub fn parse(input: &str) -> Result<Self, FactFileError> {
        parse_facts(input)
    }

    /// Produces a copy with nodes and edges in canonical order.
    pub fn canonicalized(&self) -> Self {
        let mut facts = self.clone();
        facts.nodes.sort_unstable();
        facts.edges.sort_unstable();
        facts.unresolved.sort_unstable();
        facts
    }
}
