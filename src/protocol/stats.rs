use std::collections::BTreeMap;

use serde_json::{Value, json};

use crate::repository::{RelationKind, edge_kind_to_relation};
use crate::synthetic::NodeId;

use super::response::reason_name;
use super::session::{ProtocolSnapshot, RequestFailure};

impl ProtocolSnapshot {
    pub(crate) fn stats(&self) -> Result<Value, RequestFailure> {
        let mut node_kinds = BTreeMap::<String, u64>::new();
        for entry in self.catalogue.entries() {
            *node_kinds
                .entry(entry.fact.kind.as_str().to_owned())
                .or_default() += 1;
        }

        let mut relations = BTreeMap::<String, u64>::new();
        for node in 0..self.graph.node_count() {
            for neighbor in self
                .graph
                .forward_neighbors(NodeId(node))
                .map_err(|error| RequestFailure::new("query_failed", error.to_string()))?
            {
                let relation = edge_kind_to_relation(neighbor.kind).ok_or_else(|| {
                    RequestFailure::new(
                        "corrupt_graph",
                        format!("unknown edge kind {}", neighbor.kind.0),
                    )
                })?;
                *relations.entry(relation.as_str().to_owned()).or_default() += 1;
            }
        }

        let mut unresolved_reasons = BTreeMap::<String, u64>::new();
        let mut unresolved_calls = 0_u64;
        for reference in &self.unresolved {
            *unresolved_reasons
                .entry(reason_name(&reference.reason).to_owned())
                .or_default() += 1;
            if reference.relation == RelationKind::Calls {
                unresolved_calls += 1;
            }
        }
        let resolved_call_relationships = relations.get("calls").copied().unwrap_or(0);

        Ok(json!({
            "node_count": self.graph.node_count(),
            "edge_count": self.graph.edge_count(),
            "unresolved_count": self.unresolved.len(),
            "dataset_checksum": format!("{:016x}", self.graph.dataset_checksum()),
            "nodes_by_kind": node_kinds,
            "edges_by_relation": relations,
            "unresolved_by_reason": unresolved_reasons,
            "call_resolution": {
                "resolved_unique_relationships": resolved_call_relationships,
                "unresolved_references": unresolved_calls,
                "coverage_available": false,
                "coverage": Value::Null,
                "coverage_unavailable_reason":
                    "resolved call sites are deduplicated into graph relationships",
            },
        }))
    }
}
