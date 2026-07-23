use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

use serde_json::Value;

use crate::repository::{
    CatalogueEntry, NodeKey, REPOSITORY_MANIFEST_FILE, RepositoryCatalogue, RepositorySnapshot,
    UnresolvedReferenceFact,
};
use crate::snapshot::GraphSnapshot;
use crate::synthetic::NodeId;

use super::error::ProtocolError;
use super::request::{RequestCommand, RequestEnvelope};
use super::response::{failure, success};

/// One opened repository snapshot serving repeated protocol queries.
#[derive(Debug)]
pub struct ProtocolSnapshot {
    pub(crate) root: PathBuf,
    pub(crate) graph: GraphSnapshot,
    pub(crate) catalogue: RepositoryCatalogue,
    pub(crate) unresolved: Vec<UnresolvedReferenceFact>,
    pub(crate) unresolved_by_source: BTreeMap<NodeKey, Vec<usize>>,
}

impl ProtocolSnapshot {
    /// Opens and validates a manifest-bound repository snapshot.
    pub fn open(root: impl AsRef<Path>) -> Result<Self, ProtocolError> {
        let root = root.as_ref().to_path_buf();
        let repository = RepositorySnapshot::open(root.join(REPOSITORY_MANIFEST_FILE))
            .map_err(|error| ProtocolError::InvalidSnapshot(error.to_string()))?;
        let (graph, catalogue, unresolved_facts) = repository.into_protocol_parts();
        let unresolved = unresolved_facts.unresolved;

        for reference in &unresolved {
            if catalogue.node_id_by_key(reference.source).is_none() {
                return Err(ProtocolError::InvalidSnapshot(format!(
                    "unresolved source {:016x} is absent from the catalogue",
                    reference.source.0
                )));
            }
        }
        let mut unresolved_by_source = BTreeMap::new();
        for (index, reference) in unresolved.iter().enumerate() {
            unresolved_by_source
                .entry(reference.source)
                .or_insert_with(Vec::new)
                .push(index);
        }

        Ok(Self {
            root,
            graph,
            catalogue,
            unresolved,
            unresolved_by_source,
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
            RequestCommand::Paths {
                from_node_id,
                to_node_id,
                relations,
                max_depth,
                limit,
            } => self.paths(
                from_node_id,
                to_node_id,
                relations.as_deref(),
                max_depth,
                limit,
            ),
            RequestCommand::Reachability {
                entry_node_ids,
                include_possible,
                max_depth,
                limit,
            } => self.reachability(
                &entry_node_ids,
                include_possible.unwrap_or(true),
                max_depth,
                limit,
            ),
            RequestCommand::Impact {
                node_id,
                relations,
                max_depth,
                limit,
            } => self.impact(node_id, relations.as_deref(), max_depth, limit),
            RequestCommand::ShortestCallChain {
                from_node_id,
                to_node_id,
                include_possible,
                max_depth,
            } => self.shortest_call_chain(
                from_node_id,
                to_node_id,
                include_possible.unwrap_or(true),
                max_depth,
            ),
            RequestCommand::DeadSymbols {
                entry_node_ids,
                include_possible,
                kinds,
                max_depth,
                limit,
            } => self.dead_symbols(
                &entry_node_ids,
                include_possible.unwrap_or(true),
                kinds.as_deref(),
                max_depth,
                limit,
            ),
            RequestCommand::OperationalRole {
                node_id,
                entry_node_ids,
                include_possible,
                max_depth,
            } => self.operational_role(
                node_id,
                entry_node_ids.as_deref(),
                include_possible.unwrap_or(true),
                max_depth,
            ),
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
        self.catalogue.node_id_by_key(key)
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
