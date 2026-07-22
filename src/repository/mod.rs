//! Language-neutral repository facts and their deterministic text format.

mod catalogue;
mod compiler;
#[cfg(test)]
mod compiler_catalogue_tests;
mod fact_file;
#[cfg(test)]
mod fact_file_tests;
mod model;

pub use catalogue::{
    CatalogueEntry, CatalogueError, RepositoryCatalogue, read_catalogue, write_catalogue,
};
pub use compiler::{
    CompiledRepository, RepositoryCompileError, compile_facts, compile_repository_facts,
    edge_kind_to_relation, relation_to_edge_kind,
};
pub use fact_file::{FactFileError, encode_facts, parse_facts};
pub use model::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryPathError,
    SourceSpan, normalize_repository_path,
};

/// The complete set of facts extracted from one repository.
#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct RepositoryFacts {
    pub nodes: Vec<NodeFact>,
    pub edges: Vec<EdgeFact>,
}

impl RepositoryFacts {
    pub fn new(nodes: Vec<NodeFact>, edges: Vec<EdgeFact>) -> Self {
        Self { nodes, edges }
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
        facts
    }
}
