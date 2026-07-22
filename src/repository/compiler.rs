use std::collections::BTreeMap;
use std::fmt;

use crate::synthetic::{Edge, EdgeKind, GraphDataset, NodeId};

use super::catalogue::{CatalogueEntry, CatalogueError, RepositoryCatalogue};
use super::{NodeFact, NodeKey, RelationKind, RepositoryFacts};

/// The compiled graph and its metadata catalogue.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CompiledRepository {
    pub dataset: GraphDataset,
    pub node_ids: BTreeMap<NodeKey, NodeId>,
    pub catalogue: RepositoryCatalogue,
}

/// A repository fact set that cannot be compiled into a dense graph.
#[derive(Debug)]
pub enum RepositoryCompileError {
    DuplicateConflictingNode {
        key: NodeKey,
    },
    MissingEdgeEndpoint {
        key: NodeKey,
    },
    SelfEdge {
        key: NodeKey,
    },
    DuplicateEdge {
        source: NodeKey,
        target: NodeKey,
        relation: RelationKind,
    },
    NodeIdOverflow,
    Catalogue(CatalogueError),
}

impl fmt::Display for RepositoryCompileError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::DuplicateConflictingNode { key } => {
                write!(formatter, "node key {key:?} has conflicting facts")
            }
            Self::MissingEdgeEndpoint { key } => {
                write!(formatter, "edge references missing node key {key:?}")
            }
            Self::SelfEdge { key } => write!(formatter, "node key {key:?} has a self-edge"),
            Self::DuplicateEdge {
                source,
                target,
                relation,
            } => write!(
                formatter,
                "duplicate {relation:?} edge {source:?} -> {target:?}"
            ),
            Self::NodeIdOverflow => {
                formatter.write_str("repository node count exceeds u32 capacity")
            }
            Self::Catalogue(error) => error.fmt(formatter),
        }
    }
}

impl std::error::Error for RepositoryCompileError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Catalogue(error) => Some(error),
            _ => None,
        }
    }
}

/// Compiles facts into a deterministic dense graph and catalogue.
pub fn compile_repository_facts(
    facts: &RepositoryFacts,
) -> Result<CompiledRepository, RepositoryCompileError> {
    let nodes = unique_nodes(&facts.nodes)?;
    let node_count =
        u32::try_from(nodes.len()).map_err(|_| RepositoryCompileError::NodeIdOverflow)?;
    let node_ids = nodes
        .keys()
        .enumerate()
        .map(|(index, key)| {
            u32::try_from(index)
                .map(|value| (*key, NodeId(value)))
                .map_err(|_| RepositoryCompileError::NodeIdOverflow)
        })
        .collect::<Result<BTreeMap<_, _>, _>>()?;

    let mut graph_edges = Vec::with_capacity(facts.edges.len());
    for edge in &facts.edges {
        let source = *node_ids
            .get(&edge.source)
            .ok_or(RepositoryCompileError::MissingEdgeEndpoint { key: edge.source })?;
        let target = *node_ids
            .get(&edge.target)
            .ok_or(RepositoryCompileError::MissingEdgeEndpoint { key: edge.target })?;
        if source == target {
            return Err(RepositoryCompileError::SelfEdge { key: edge.source });
        }
        graph_edges.push(Edge {
            source,
            target,
            kind: relation_to_edge_kind(&edge.relation),
        });
    }
    graph_edges.sort_unstable();
    if graph_edges.windows(2).any(|pair| pair[0] == pair[1]) {
        let duplicate = graph_edges
            .windows(2)
            .find(|pair| pair[0] == pair[1])
            .expect("duplicate edge was found")[0];
        return Err(RepositoryCompileError::DuplicateEdge {
            source: node_key_for_id(&node_ids, duplicate.source),
            target: node_key_for_id(&node_ids, duplicate.target),
            relation: edge_kind_to_relation(duplicate.kind).expect("compiler emitted known kind"),
        });
    }

    let entries = nodes
        .into_iter()
        .map(|(key, fact)| CatalogueEntry {
            node_id: node_ids[&key],
            fact,
        })
        .collect();
    let catalogue = RepositoryCatalogue::new(entries).map_err(RepositoryCompileError::Catalogue)?;
    Ok(CompiledRepository {
        dataset: GraphDataset {
            node_count,
            edges: graph_edges,
        },
        node_ids,
        catalogue,
    })
}

/// Short alias for compiling a repository fact set.
pub fn compile_facts(
    facts: &RepositoryFacts,
) -> Result<CompiledRepository, RepositoryCompileError> {
    compile_repository_facts(facts)
}

/// Maps every repository relation to its stable nonzero graph edge code.
pub fn relation_to_edge_kind(relation: &RelationKind) -> EdgeKind {
    EdgeKind(match relation {
        RelationKind::Contains => 1,
        RelationKind::Defines => 2,
        RelationKind::References => 3,
        RelationKind::Imports => 4,
        RelationKind::Calls => 5,
        RelationKind::Implements => 6,
        RelationKind::Extends => 7,
        RelationKind::Includes => 8,
        RelationKind::DependsOn => 9,
        RelationKind::Tests => 10,
        RelationKind::Documents => 11,
        RelationKind::Generates => 12,
    })
}

/// Converts a stable graph edge code back to its repository relation.
pub fn edge_kind_to_relation(kind: EdgeKind) -> Option<RelationKind> {
    Some(match kind.0 {
        1 => RelationKind::Contains,
        2 => RelationKind::Defines,
        3 => RelationKind::References,
        4 => RelationKind::Imports,
        5 => RelationKind::Calls,
        6 => RelationKind::Implements,
        7 => RelationKind::Extends,
        8 => RelationKind::Includes,
        9 => RelationKind::DependsOn,
        10 => RelationKind::Tests,
        11 => RelationKind::Documents,
        12 => RelationKind::Generates,
        _ => return None,
    })
}

fn unique_nodes(nodes: &[NodeFact]) -> Result<BTreeMap<NodeKey, NodeFact>, RepositoryCompileError> {
    let mut unique = BTreeMap::new();
    for node in nodes {
        if let Some(previous) = unique.get(&node.key)
            && previous != node
        {
            return Err(RepositoryCompileError::DuplicateConflictingNode { key: node.key });
        }
        unique.entry(node.key).or_insert_with(|| node.clone());
    }
    Ok(unique)
}

fn node_key_for_id(node_ids: &BTreeMap<NodeKey, NodeId>, node_id: NodeId) -> NodeKey {
    node_ids
        .iter()
        .find_map(|(key, value)| (*value == node_id).then_some(*key))
        .expect("edge endpoint was assigned by compiler")
}
