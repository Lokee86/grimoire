use crate::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, UnresolvedReason,
    UnresolvedReferenceFact,
};

use super::{graph_documents, identifier_terms, semantic_eligible};

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
        NodeKind::Signal,
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
        NodeKind::Constant,
        NodeKind::Test,
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
fn filters_synthetic_and_anonymous_entry_points() {
    let facts = RepositoryFacts {
        nodes: vec![
            node(NodeKey::from_u64(1), "visible", "src/main.rs"),
            node(NodeKey::from_u64(2), "builtin", "@builtin/go"),
            node(NodeKey::from_u64(3), "closure@42:7", "src/main.rs"),
            node(NodeKey::from_u64(4), "<lambda@9:3>", "src/main.rs"),
        ],
        edges: Vec::new(),
        unresolved: Vec::new(),
    };

    let documents = graph_documents(&facts);
    assert_eq!(documents.len(), 1);
    assert_eq!(documents[0].name, "visible");
}

#[test]
fn bounds_generated_unresolved_expressions_to_the_embedding_contract() {
    let source = NodeKey::from_u64(1);
    let facts = RepositoryFacts {
        nodes: vec![node(source, "generated_contract", "src/generated.rb")],
        edges: Vec::new(),
        unresolved: vec![UnresolvedReferenceFact {
            source,
            relation: RelationKind::Calls,
            expression: "generated_payload ".repeat(10_000),
            candidate_namespace: None,
            candidate_name: Some("generated_payload".to_owned()),
            reason: UnresolvedReason::UnsupportedForm,
            span: None,
        }],
    };

    let documents = graph_documents(&facts);
    assert_eq!(documents.len(), 1);
    assert!(documents[0].text.len() <= 6_000);
    assert!(documents[0].text.ends_with("\ndocument truncated\n"));
    assert!(documents[0].text.contains("name: generated_contract"));
}

#[test]
fn renders_identifier_and_path_terms_for_conceptual_retrieval() {
    assert_eq!(identifier_terms("SemanticSeeds"), "semantic seeds");
    assert_eq!(identifier_terms("graph_documents"), "graph documents");
    assert_eq!(identifier_terms("HTTPServer2"), "http server2");
    assert_eq!(
        identifier_terms("internal/arcana_graph/vector-index.rs"),
        "internal arcana graph vector index rs"
    );

    let source = NodeKey::from_u64(1);
    let target = NodeKey::from_u64(2);
    let documents = graph_documents(&RepositoryFacts {
        nodes: vec![
            node(source, "SemanticSeeds", "internal/arcana_graph/semantic.go"),
            node(
                target,
                "semanticIndexExists",
                "internal/arcana_graph/semantic.go",
            ),
        ],
        edges: vec![EdgeFact {
            source,
            target,
            relation: RelationKind::Calls,
            span: None,
        }],
        unresolved: Vec::new(),
    });
    let document = documents
        .iter()
        .find(|document| document.name == "SemanticSeeds")
        .unwrap();

    assert!(document.text.contains("name terms: semantic seeds"));
    assert!(
        document
            .text
            .contains("path terms: internal arcana graph semantic go")
    );
    assert!(
        document
            .text
            .contains("outgoing calls function semanticIndexExists (semantic index exists)")
    );
}

#[test]
fn prioritizes_semantic_neighbors_over_low_level_and_reference_noise() {
    let source = NodeKey::from_u64(1);
    let callee = NodeKey::from_u64(2);
    let mut nodes = vec![
        node(source, "semantic_seeds", "src/semantic.rs"),
        node(callee, "semantic_index_exists", "src/semantic.rs"),
    ];
    let mut edges = vec![EdgeFact {
        source,
        target: callee,
        relation: RelationKind::Calls,
        span: None,
    }];

    for index in 0..16_u64 {
        let key = NodeKey::from_u64(10 + index);
        nodes.push(node_with_kind(
            key,
            NodeKind::Variable,
            &format!("scratch_{index}"),
        ));
        edges.push(EdgeFact {
            source,
            target: key,
            relation: RelationKind::Reads,
            span: None,
        });
    }
    for index in 0..16_u64 {
        let key = NodeKey::from_u64(100 + index);
        nodes.push(node_with_kind(
            key,
            NodeKind::File,
            &format!("reference_{index}.rs"),
        ));
        edges.push(EdgeFact {
            source,
            target: key,
            relation: RelationKind::References,
            span: None,
        });
    }

    let documents = graph_documents(&RepositoryFacts {
        nodes,
        edges,
        unresolved: Vec::new(),
    });
    let document = documents
        .iter()
        .find(|document| document.name == "semantic_seeds")
        .unwrap();

    assert!(
        document
            .text
            .contains("outgoing calls function semantic_index_exists")
    );
    assert!(!document.text.contains("scratch_"));
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
        qualified_name: name.to_owned(),
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
        qualified_name: name.to_owned(),
        content_id: None,
        span: None,
    }
}
