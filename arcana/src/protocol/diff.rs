use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;

use serde_json::{Value, json};

use crate::repository::{CatalogueEntry, NodeKey};
use crate::synthetic::NodeId;

use super::response::node_value;
use super::session::{ProtocolSnapshot, RequestFailure};

const DEFAULT_DIFF_LIMIT: usize = 1_000;
const MAX_DIFF_LIMIT: usize = 10_000;

impl ProtocolSnapshot {
    pub(crate) fn diff_snapshot(
        &self,
        other_path: &Path,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let other = Self::open(other_path)
            .map_err(|error| RequestFailure::new("snapshot_open_failed", error.to_string()))?;
        let current = entries_by_key(self.catalogue.entries());
        let previous = entries_by_key(other.catalogue.entries());

        let added = current
            .iter()
            .filter(|(key, _)| !previous.contains_key(key))
            .map(|(_, entry)| *entry)
            .collect::<Vec<_>>();
        let removed = previous
            .iter()
            .filter(|(key, _)| !current.contains_key(key))
            .map(|(_, entry)| *entry)
            .collect::<Vec<_>>();
        let metadata_changed = current
            .iter()
            .filter_map(|(key, entry)| {
                previous
                    .get(key)
                    .filter(|previous| previous.fact != entry.fact)
                    .map(|_| *entry)
            })
            .collect::<Vec<_>>();

        let mut relationship_changed = Vec::new();
        for (key, entry) in &current {
            let Some(previous_entry) = previous.get(key) else {
                continue;
            };
            if logical_outgoing(self, entry.node_id)?
                != logical_outgoing(&other, previous_entry.node_id)?
            {
                relationship_changed.push(*entry);
            }
        }

        let logical_graph_changed =
            !added.is_empty() || !removed.is_empty() || !relationship_changed.is_empty();
        let snapshot_changed = logical_graph_changed || !metadata_changed.is_empty();
        let packed_checksum_changed =
            self.graph.dataset_checksum() != other.graph.dataset_checksum();
        let limit = limit.unwrap_or(DEFAULT_DIFF_LIMIT).min(MAX_DIFF_LIMIT);
        Ok(json!({
            "current_snapshot": self.root.display().to_string(),
            "other_snapshot": other.root.display().to_string(),
            "current_checksum": format!("{:016x}", self.graph.dataset_checksum()),
            "other_checksum": format!("{:016x}", other.graph.dataset_checksum()),
            "snapshot_changed": snapshot_changed,
            "graph_changed": logical_graph_changed,
            "packed_checksum_changed": packed_checksum_changed,
            "counts": {
                "added": added.len(),
                "removed": removed.len(),
                "metadata_changed": metadata_changed.len(),
                "relationship_changed": relationship_changed.len(),
            },
            "truncated": added.len() > limit
                || removed.len() > limit
                || metadata_changed.len() > limit
                || relationship_changed.len() > limit,
            "nodes": {
                "added": values(&added, limit),
                "removed": values(&removed, limit),
                "metadata_changed": values(&metadata_changed, limit),
                "relationship_changed": values(&relationship_changed, limit),
            },
        }))
    }
}

fn entries_by_key(entries: &[CatalogueEntry]) -> BTreeMap<NodeKey, &CatalogueEntry> {
    entries
        .iter()
        .map(|entry| (entry.fact.key, entry))
        .collect()
}

fn logical_outgoing(
    snapshot: &ProtocolSnapshot,
    node_id: NodeId,
) -> Result<BTreeSet<(NodeKey, u16)>, RequestFailure> {
    snapshot
        .graph
        .forward_neighbors(node_id)
        .map_err(|error| RequestFailure::new("query_failed", error.to_string()))?
        .into_iter()
        .map(|neighbor| {
            snapshot
                .entry(neighbor.node)
                .map(|entry| (entry.fact.key, neighbor.kind.0))
                .ok_or_else(|| {
                    RequestFailure::new(
                        "invalid_snapshot",
                        format!("catalogue is missing graph node {}", neighbor.node.0),
                    )
                })
        })
        .collect()
}

fn values(entries: &[&CatalogueEntry], limit: usize) -> Vec<Value> {
    entries
        .iter()
        .take(limit)
        .map(|entry| node_value(entry))
        .collect()
}
