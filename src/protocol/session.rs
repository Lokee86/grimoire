use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

use serde_json::Value;

use crate::repository::{
    CatalogueEntry, NodeKey, RepositoryCatalogue, RepositoryFacts, UnresolvedReferenceFact,
};
use crate::storage::PackedGraph;
use crate::synthetic::NodeId;

use super::error::ProtocolError;
use super::request::{RequestCommand, RequestEnvelope};
use super::response::{failure, success};

/// One opened repository snapshot serving repeated protocol queries.
#[derive(Debug)]
pub struct ProtocolSnapshot {
    pub(crate) root: PathBuf,
    pub(crate) graph: PackedGraph,
    pub(crate) catalogue: RepositoryCatalogue,
    pub(crate) unresolved: Vec<UnresolvedReferenceFact>,
    pub(crate) node_ids: BTreeMap<NodeKey, NodeId>,
}

impl ProtocolSnapshot {
    /// Opens and validates the three repository snapshot artifacts.
    pub fn open(root: impl AsRef<Path>) -> Result<Self, ProtocolError> {
        let root = root.as_ref().to_path_buf();
        let graph = PackedGraph::open(root.join("graph.arcana"))?;
        let catalogue = RepositoryCatalogue::read(root.join("catalogue.tsv"))?;
        let unresolved_text = fs::read_to_string(root.join("unresolved.tsv"))?;
        let unresolved = RepositoryFacts::parse(&unresolved_text)?.unresolved;

        if graph.node_count() as usize != catalogue.len() {
            return Err(ProtocolError::InvalidSnapshot(format!(
                "graph has {} nodes but catalogue has {} entries",
                graph.node_count(),
                catalogue.len()
            )));
        }

        let node_ids = catalogue
            .entries()
            .iter()
            .map(|entry| (entry.fact.key, entry.node_id))
            .collect::<BTreeMap<_, _>>();
        for reference in &unresolved {
            if !node_ids.contains_key(&reference.source) {
                return Err(ProtocolError::InvalidSnapshot(format!(
                    "unresolved source {:016x} is absent from the catalogue",
                    reference.source.0
                )));
            }
        }

        Ok(Self {
            root,
            graph,
            catalogue,
            unresolved,
            node_ids,
        })
    }

    /// Handles one JSON request and always returns one JSON response.
    pub fn handle_line(&self, line: &str) -> Value {
        let request = match serde_json::from_str::<RequestEnvelope>(line) {
            Ok(request) => request,
            Err(error) => return failure(Value::Null, "invalid_json", error.to_string()),
        };
        let id = request.id;
        match self.execute(request.command) {
            Ok(result) => success(id, result),
            Err(error) => failure(id, error.code, error.message),
        }
    }

    fn execute(&self, command: RequestCommand) -> Result<Value, RequestFailure> {
        match command {
            RequestCommand::ResolveSymbol {
                name,
                kind,
                path,
                limit,
            } => self.resolve_symbol(&name, kind.as_deref(), path.as_deref(), limit),
            RequestCommand::ResolveFile { path, limit } => self.resolve_file(&path, limit),
            RequestCommand::ListNodes {
                kind,
                path_prefix,
                limit,
            } => self.list_nodes(kind.as_deref(), path_prefix.as_deref(), limit),
            RequestCommand::Neighbors {
                node_id,
                direction,
                relation,
            } => self.neighbors(node_id, direction, relation.as_deref()),
            RequestCommand::Unresolved {
                node_id,
                path,
                reason,
                relation,
                limit,
            } => self.query_unresolved(
                node_id,
                path.as_deref(),
                reason.as_deref(),
                relation.as_deref(),
                limit,
            ),
            RequestCommand::Stats => self.stats(),
            RequestCommand::Diff {
                other_snapshot,
                limit,
            } => self.diff_snapshot(&other_snapshot, limit),
        }
    }

    pub(crate) fn entry(&self, node_id: NodeId) -> Option<&CatalogueEntry> {
        self.catalogue
            .entries()
            .get(node_id.0 as usize)
            .filter(|entry| entry.node_id == node_id)
    }

    pub(crate) fn node_id(&self, key: NodeKey) -> Option<NodeId> {
        self.node_ids.get(&key).copied()
    }
}

#[derive(Debug)]
pub(crate) struct RequestFailure {
    pub code: &'static str,
    pub message: String,
}

impl RequestFailure {
    pub(crate) fn new(code: &'static str, message: impl Into<String>) -> Self {
        Self {
            code,
            message: message.into(),
        }
    }
}
