use serde_json::{Value, json};

use crate::repository::{
    NodeKind, RelationKind, UnresolvedReason, edge_kind_to_relation, normalize_repository_path,
};
use crate::synthetic::NodeId;

use super::request::QueryDirection;
use super::response::{node_value, relationship_value, unresolved_value};
use super::session::{ProtocolSnapshot, RequestFailure};
use super::traversal::parse_relations;

const DEFAULT_LIMIT: usize = 1_000;
const MAX_LIMIT: usize = 10_000;

impl ProtocolSnapshot {
    pub(crate) fn search_nodes(
        &self,
        query: &str,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let query = query.trim().to_ascii_lowercase();
        if query.is_empty() {
            return Ok(json!({
                "count": 0,
                "returned": 0,
                "truncated": false,
                "matches": [],
            }));
        }

        let normalized_query = query.replace('\\', "/");
        let mut matches = self
            .catalogue
            .entries()
            .iter()
            .filter_map(|entry| {
                let name = entry.fact.name.to_ascii_lowercase();
                let qualified_name = entry.fact.qualified_name.to_ascii_lowercase();
                let path = entry.fact.path.to_ascii_lowercase().replace('\\', "/");
                let mut matched_fields = Vec::new();
                let mut rank = usize::MAX;
                for (field, value, exact, prefix, contains) in [
                    ("name", name.as_str(), 0, 3, 6),
                    ("qualified_name", qualified_name.as_str(), 1, 4, 7),
                    ("path", path.as_str(), 2, 5, 8),
                ] {
                    if let Some(field_rank) =
                        text_match_rank(value, &normalized_query, exact, prefix, contains)
                    {
                        matched_fields.push(field);
                        rank = rank.min(field_rank);
                    }
                }
                (!matched_fields.is_empty()).then_some((rank, matched_fields, entry))
            })
            .collect::<Vec<_>>();
        matches.sort_by(|left, right| {
            left.0
                .cmp(&right.0)
                .then_with(|| left.2.fact.qualified_name.cmp(&right.2.fact.qualified_name))
                .then_with(|| left.2.fact.path.cmp(&right.2.fact.path))
                .then_with(|| left.2.node_id.cmp(&right.2.node_id))
        });

        let count = matches.len();
        let matches = matches
            .into_iter()
            .take(bounded_limit(limit))
            .map(|(rank, matched_fields, entry)| {
                json!({
                    "rank": rank,
                    "matched_fields": matched_fields,
                    "node": node_value(entry),
                })
            })
            .collect::<Vec<_>>();
        Ok(json!({
            "count": count,
            "returned": matches.len(),
            "truncated": count > matches.len(),
            "matches": matches,
        }))
    }

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
        offset: Option<usize>,
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
        Ok(node_list_page(self, matches, offset, limit))
    }

    pub(crate) fn neighbors(
        &self,
        node_id: u32,
        direction: QueryDirection,
        relation: Option<&str>,
        relations: Option<&[String]>,
        limit: Option<usize>,
    ) -> Result<Value, RequestFailure> {
        let node_id = NodeId(node_id);
        let source = self.entry(node_id).ok_or_else(|| {
            RequestFailure::new("unknown_node", format!("node {node_id:?} does not exist"))
        })?;
        let wanted = parse_relation(relation)?;
        let wanted_many = parse_relations(relations)?;
        let neighbors = match direction {
            QueryDirection::Outgoing => self.graph.forward_neighbors_iter(node_id),
            QueryDirection::Incoming => self.graph.reverse_neighbors_iter(node_id),
        }
        .map_err(|error| RequestFailure::new("query_failed", error.to_string()))?;

        let limit = bounded_limit(limit);
        let scan_limit = limit.saturating_mul(16).clamp(1_024, 16_384);
        let mut scanned = 0usize;
        let mut matched = 0usize;
        let mut truncated = false;
        let mut relationships = Vec::new();
        for neighbor in neighbors {
            if scanned >= scan_limit {
                truncated = true;
                break;
            }
            scanned += 1;
            let relation = edge_kind_to_relation(neighbor.kind).ok_or_else(|| {
                RequestFailure::new(
                    "corrupt_graph",
                    format!("unknown edge kind {}", neighbor.kind.0),
                )
            })?;
            if wanted_many.is_some_and(|wanted| !wanted.contains(&relation)) {
                continue;
            }
            if wanted_many.is_none() && wanted.as_ref().is_some_and(|wanted| wanted != &relation) {
                continue;
            }
            matched += 1;
            if relationships.len() >= limit {
                truncated = true;
                break;
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
            "count": matched,
            "returned": relationships.len(),
            "scanned": scanned,
            "truncated": truncated,
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
    node_list_page(snapshot, matches, None, limit)
}

fn node_list_page(
    snapshot: &ProtocolSnapshot,
    matches: Vec<NodeId>,
    offset: Option<usize>,
    limit: Option<usize>,
) -> Value {
    let total = matches.len();
    let offset = offset.unwrap_or(0).min(total);
    let limit = bounded_limit(limit);
    let nodes = matches
        .into_iter()
        .skip(offset)
        .take(limit)
        .map(|node_id| {
            node_value(
                snapshot
                    .entry(node_id)
                    .expect("catalogue index contains only valid node IDs"),
            )
        })
        .collect::<Vec<_>>();
    let next_offset = offset + nodes.len();
    let truncated = next_offset < total;
    json!({
        "count": total,
        "offset": offset,
        "returned": nodes.len(),
        "truncated": truncated,
        "next_offset": truncated.then_some(next_offset),
        "nodes": nodes,
    })
}

fn text_match_rank(
    value: &str,
    query: &str,
    exact_rank: usize,
    prefix_rank: usize,
    contains_rank: usize,
) -> Option<usize> {
    if value == query {
        Some(exact_rank)
    } else if value.starts_with(query) {
        Some(prefix_rank)
    } else if value.contains(query) {
        Some(contains_rank)
    } else {
        None
    }
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
