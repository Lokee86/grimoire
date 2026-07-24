use serde_json::{Value, json};

use crate::repository::{
    NodeKind, RelationKind, UnresolvedReason, edge_kind_to_relation, normalize_repository_path,
};
use crate::synthetic::NodeId;

use super::request::QueryDirection;
use super::response::{node_value, relationship_value, unresolved_value};
use super::session::{ProtocolSnapshot, RequestFailure};

const DEFAULT_LIMIT: usize = 1_000;
const MAX_LIMIT: usize = 10_000;

impl ProtocolSnapshot {
    pub(crate) fn resolve_symbol(
        &self,
        name: &str,
        kind: Option<&str>,
        path: Option<&str>,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let kind = parse_kind(kind)?;
        let path = normalize_optional_path(path)?;
        let mut matches = self.catalogue.node_ids_by_name(name).to_vec();
        if let Some(kind) = kind {
            matches = intersect_sorted_ids(&matches, self.catalogue.node_ids_by_kind(&kind));
        }
        if let Some(path) = path {
            matches = intersect_sorted_ids(
                &matches,
                self.catalogue
                    .node_ids_by_path(&path)
                    .map_err(|error| RequestFailure::new("invalid_path", error.to_string()))?,
            );
        }
        Ok(node_list(self, matches, limit))
    }

    pub(crate) fn resolve_file(
        &self,
        path: &str,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let matches = self
            .catalogue
            .node_ids_by_path(path)
            .map_err(|error| RequestFailure::new("invalid_path", error.to_string()))?
            .to_vec();
        let matches =
            intersect_sorted_ids(&matches, self.catalogue.node_ids_by_kind(&NodeKind::File));
        Ok(node_list(self, matches, limit))
    }

    pub(crate) fn list_nodes(
        &self,
        kind: Option<&str>,
        path_prefix: Option<&str>,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let kind = parse_kind(kind)?;
        let path_prefix = normalize_optional_path(path_prefix)?;
        let matches = match (kind, path_prefix) {
            (None, None) => (0..self.catalogue.len())
                .map(|node_id| NodeId(node_id as u32))
                .collect(),
            (Some(kind), None) => self.catalogue.node_ids_by_kind(&kind).to_vec(),
            (None, Some(prefix)) => self
                .catalogue
                .node_ids_by_path_prefix(&prefix)
                .map_err(|error| RequestFailure::new("invalid_path", error.to_string()))?,
            (Some(kind), Some(prefix)) => {
                let path_ids = self
                    .catalogue
                    .node_ids_by_path_prefix(&prefix)
                    .map_err(|error| RequestFailure::new("invalid_path", error.to_string()))?;
                intersect_sorted_ids(&path_ids, self.catalogue.node_ids_by_kind(&kind))
            }
        };
        Ok(node_list(self, matches, limit))
    }

    pub(crate) fn neighbors(
        &self,
        node_id: u32,
        direction: QueryDirection,
        relation: Option<&str>,
    ) -> Result<Value, RequestFailure> {
        let node_id = NodeId(node_id);
        let source = self.entry(node_id).ok_or_else(|| {
            RequestFailure::new("unknown_node", format!("node {node_id:?} does not exist"))
        })?;
        let wanted = parse_relation(relation)?;
        let neighbors = match direction {
            QueryDirection::Outgoing => self.graph.forward_neighbors_iter(node_id),
            QueryDirection::Incoming => self.graph.reverse_neighbors_iter(node_id),
        }
        .map_err(|error| RequestFailure::new("query_failed", error.to_string()))?;

        let mut relationships = Vec::new();
        for neighbor in neighbors {
            let relation = edge_kind_to_relation(neighbor.kind).ok_or_else(|| {
                RequestFailure::new(
                    "corrupt_graph",
                    format!("unknown edge kind {}", neighbor.kind.0),
                )
            })?;
            if wanted.as_ref().is_some_and(|wanted| wanted != &relation) {
                continue;
            }
            let entry = self.entry(neighbor.node).ok_or_else(|| {
                RequestFailure::new(
                    "invalid_snapshot",
                    format!("catalogue is missing graph node {}", neighbor.node.0),
                )
            })?;
            relationships.push(relationship_value(&relation, entry));
        }
        Ok(json!({
            "node": node_value(source),
            "direction": match direction {
                QueryDirection::Incoming => "incoming",
                QueryDirection::Outgoing => "outgoing",
            },
            "count": relationships.len(),
            "relationships": relationships,
        }))
    }

    pub(crate) fn query_unresolved(
        &self,
        node_id: Option<u32>,
        path: Option<&str>,
        reason: Option<&str>,
        relation: Option<&str>,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let source_key = node_id
            .map(|node_id| {
                self.entry(NodeId(node_id))
                    .map(|entry| entry.fact.key)
                    .ok_or_else(|| {
                        RequestFailure::new(
                            "unknown_node",
                            format!("node {node_id} does not exist"),
                        )
                    })
            })
            .transpose()?;
        let path = normalize_optional_path(path)?;
        let reason = parse_reason(reason)?;
        let relation = parse_relation(relation)?;
        let source_matches = source_key.map(|key| {
            self.unresolved_by_source
                .get(&key)
                .into_iter()
                .flat_map(|indices| indices.iter())
                .map(|&index| &self.unresolved[index])
                .collect::<Vec<_>>()
        });
        let matches = source_matches
            .unwrap_or_else(|| self.unresolved.iter().collect::<Vec<_>>())
            .into_iter()
            .filter(|reference| source_key.is_none_or(|key| reference.source == key))
            .filter(|reference| {
                reason
                    .as_ref()
                    .is_none_or(|reason| &reference.reason == reason)
            })
            .filter(|reference| {
                relation
                    .as_ref()
                    .is_none_or(|relation| &reference.relation == relation)
            })
            .filter(|reference| {
                path.as_ref().is_none_or(|path| {
                    self.node_id(reference.source)
                        .and_then(|id| self.entry(id))
                        .is_some_and(|entry| &entry.fact.path == path)
                })
            })
            .collect::<Vec<_>>();
        let total = matches.len();
        let limit = bounded_limit(limit);
        let items = matches
            .into_iter()
            .take(limit)
            .map(|reference| {
                let source = self
                    .node_id(reference.source)
                    .expect("snapshot validation checked unresolved sources");
                unresolved_value(reference, source)
            })
            .collect::<Vec<_>>();
        Ok(json!({
            "count": total,
            "returned": items.len(),
            "truncated": total > items.len(),
            "unresolved": items,
        }))
    }
}

fn node_list(snapshot: &ProtocolSnapshot, matches: Vec<NodeId>, limit: Option<usize>) -> Value {
    let total = matches.len();
    let limit = bounded_limit(limit);
    let nodes = matches
        .into_iter()
        .take(limit)
        .map(|node_id| {
            node_value(
                snapshot
                    .entry(node_id)
                    .expect("catalogue index contains only valid node IDs"),
            )
        })
        .collect::<Vec<_>>();
    json!({
        "count": total,
        "returned": nodes.len(),
        "truncated": total > nodes.len(),
        "nodes": nodes,
    })
}

fn intersect_sorted_ids(left: &[NodeId], right: &[NodeId]) -> Vec<NodeId> {
    let mut matches = Vec::new();
    let mut left_index = 0;
    let mut right_index = 0;
    while left_index < left.len() && right_index < right.len() {
        match left[left_index].cmp(&right[right_index]) {
            std::cmp::Ordering::Less => left_index += 1,
            std::cmp::Ordering::Greater => right_index += 1,
            std::cmp::Ordering::Equal => {
                matches.push(left[left_index]);
                left_index += 1;
                right_index += 1;
            }
        }
    }
    matches
}

fn bounded_limit(limit: Option<usize>) -> usize {
    limit.unwrap_or(DEFAULT_LIMIT).min(MAX_LIMIT)
}

fn normalize_optional_path(path: Option<&str>) -> Result<Option<String>, RequestFailure> {
    path.map(|path| {
        normalize_repository_path(path)
            .map_err(|error| RequestFailure::new("invalid_path", error.to_string()))
    })
    .transpose()
}

fn parse_kind(kind: Option<&str>) -> Result<Option<NodeKind>, RequestFailure> {
    kind.map(|kind| {
        NodeKind::parse(kind).ok_or_else(|| {
            RequestFailure::new("invalid_node_kind", format!("unknown node kind '{kind}'"))
        })
    })
    .transpose()
}

fn parse_relation(relation: Option<&str>) -> Result<Option<RelationKind>, RequestFailure> {
    relation
        .map(|relation| {
            RelationKind::parse(relation).ok_or_else(|| {
                RequestFailure::new("invalid_relation", format!("unknown relation '{relation}'"))
            })
        })
        .transpose()
}

fn parse_reason(reason: Option<&str>) -> Result<Option<UnresolvedReason>, RequestFailure> {
    reason
        .map(|reason| {
            UnresolvedReason::parse(reason).ok_or_else(|| {
                RequestFailure::new(
                    "invalid_unresolved_reason",
                    format!("unknown unresolved reason '{reason}'"),
                )
            })
        })
        .transpose()
}
