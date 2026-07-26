use crate::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, UnresolvedReason,
    UnresolvedReferenceFact,
};

use super::{graph_documents, semantic_eligible};

#[test]
fn semantic_eligibility_policy_covers_allowed_and_excluded_kinds() {
    let allowed = [
        NodeKind::File,
        NodeKind::Module,
        NodeKind::Namespace,
        NodeKind::Symbol,
        NodeKind::Type,
        NodeKind::Interface,
        NodeKind::Trait,
        NodeKind::Function,
        NodeKind::Method,
        NodeKind::Constructor,
        NodeKind::Constant,
        NodeKind::Signal,
        NodeKind::Test,
        NodeKind::HttpEndpoint,
        NodeKind::MessageChannel,
        NodeKind::ConfigKey,
    ];
    let excluded = [
        NodeKind::Repository,
        NodeKind::Directory,
        NodeKind::Field,
        NodeKind::Variable,
        NodeKind::Parameter,
        NodeKind::Import,
        NodeKind::Export,
    ];

    for kind in allowed {
        assert!(semantic_eligible(&kind), "expected {kind:?} to be eligible");
    }
    for kind in excluded {
        assert!(
            !semantic_eligible(&kind),
            "expected {kind:?} to be excluded"
        );
    }
}

#[test]
fn filters_ineligible_nodes_and_orders_documents_by_node_key() {
    let facts = RepositoryFacts {
        nodes: vec![
            node_with_kind(NodeKey::from_u64(3), NodeKind::Variable, "scratch"),
            node_with_kind(NodeKey::from_u64(2), NodeKind::Method, "run"),
            node_with_kind(NodeKey::from_u64(1), NodeKind::File, "main.rs"),
        ],
        edges: Vec::new(),
        unresolved: Vec::new(),
    };

    let documents = graph_documents(&facts);
    assert_eq!(
        documents
            .iter()
            .map(|document| (document.node_key, document.name.as_str()))
            .collect::<Vec<_>>(),
        vec![(1, "main.rs"), (2, "run")]
    );
    assert_eq!(
        facts.nodes.len(),
        3,
        "document filtering changed graph facts"
    );
}

#[test]
fn renders_node_and_immediate_graph_neighborhood() {
    let caller = NodeKey::from_u64(1);
    let callee = NodeKey::from_u64(2);
    let facts = RepositoryFacts {
        nodes: vec![
            node(caller, "create_profile", "src/profile.rs"),
            node(callee, "insert_profile", "src/repository.rs"),
        ],
        edges: vec![EdgeFact {
            source: caller,
            target: callee,
            relation: RelationKind::Calls,
            span: None,
        }],
        unresolved: vec![UnresolvedReferenceFact {
            source: caller,
            relation: RelationKind::Writes,
            expression: "profiles".to_owned(),
            candidate_namespace: None,
            candidate_name: Some("profiles".to_owned()),
            reason: UnresolvedReason::UnsupportedForm,
            span: None,
        }],
    };

    let documents = graph_documents(&facts);
    assert_eq!(documents.len(), 2);
    assert!(
        documents[0]
            .text
            .contains("outgoing calls function insert_profile")
    );
    assert!(documents[0].text.contains("unresolved writes profiles"));
    assert!(
        documents[1]
            .text
            .contains("incoming calls function create_profile")
    );
}

fn node(key: NodeKey, name: &str, path: &str) -> NodeFact {
    NodeFact {
        key,
        external_identity: None,
        kind: NodeKind::Function,
        path: path.to_owned(),
        name: name.to_owned(),
        content_id: None,
        span: None,
    }
}

fn node_with_kind(key: NodeKey, kind: NodeKind, name: &str) -> NodeFact {
    NodeFact {
        key,
        external_identity: None,
        kind,
        path: "src/main.rs".to_owned(),
        name: name.to_owned(),
        content_id: None,
        span: None,
    }
}
