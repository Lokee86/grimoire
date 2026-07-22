use crate::synthetic::{EdgeKind, NodeId};

use super::{
    CatalogueEntry, ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind,
    RepositoryCatalogue, RepositoryFacts, SourceSpan, compile_repository_facts,
    edge_kind_to_relation, read_catalogue, relation_to_edge_kind, write_catalogue,
};

#[test]
fn compiler_assigns_dense_ids_and_stable_relation_codes() {
    let first = NodeKey::from_identity("first");
    let second = NodeKey::from_identity("second");
    let facts = RepositoryFacts {
        nodes: vec![node(second, "b.rs"), node(first, "a.rs")],
        edges: vec![EdgeFact {
            source: first,
            target: second,
            relation: RelationKind::Calls,
            span: None,
        }],
    };
    let compiled = compile_repository_facts(&facts).unwrap();
    assert_eq!(compiled.dataset.node_count, 2);
    assert_eq!(
        compiled.dataset.edges,
        vec![crate::synthetic::Edge {
            source: NodeId(0),
            target: NodeId(1),
            kind: EdgeKind(5),
        }]
    );
    assert_eq!(compiled.node_ids[&first], NodeId(0));
    assert_eq!(compiled.node_ids[&second], NodeId(1));
    assert_eq!(
        edge_kind_to_relation(EdgeKind(5)),
        Some(RelationKind::Calls)
    );
    assert_eq!(relation_to_edge_kind(&RelationKind::Calls), EdgeKind(5));
}

#[test]
fn compiler_rejects_invalid_facts() {
    let key = NodeKey::from_identity("node");
    let base = node(key, "a.rs");
    let conflicting = NodeFact {
        name: "other".to_owned(),
        ..base.clone()
    };
    assert!(
        compile_repository_facts(&RepositoryFacts {
            nodes: vec![base.clone(), conflicting],
            edges: vec![]
        })
        .is_err()
    );
    assert!(
        compile_repository_facts(&RepositoryFacts {
            nodes: vec![base.clone()],
            edges: vec![edge(key, NodeKey::from_u64(99))]
        })
        .is_err()
    );
    assert!(
        compile_repository_facts(&RepositoryFacts {
            nodes: vec![base],
            edges: vec![edge(key, key)]
        })
        .is_err()
    );
}

#[test]
fn catalogue_round_trips_and_supports_exact_lookups() {
    let catalogue = sample_catalogue();
    let key = catalogue.entries()[0].fact.key;
    let encoded = catalogue.encode().unwrap();
    assert_eq!(RepositoryCatalogue::decode(&encoded).unwrap(), catalogue);
    assert_eq!(catalogue.lookup_by_key(key).unwrap().node_id, NodeId(0));
    assert_eq!(catalogue.lookup_by_path("src\\main.rs").unwrap().len(), 1);
    assert_eq!(catalogue.lookup_by_name("main").len(), 1);
}

#[test]
fn catalogue_file_is_immutable_and_validated() {
    let catalogue = sample_catalogue();
    let path = std::env::temp_dir().join(format!("arcana-catalogue-{}.txt", std::process::id()));
    let _ = std::fs::remove_file(&path);
    write_catalogue(&path, &catalogue).unwrap();
    assert_eq!(read_catalogue(&path).unwrap(), catalogue);
    assert!(write_catalogue(&path, &catalogue).is_err());
    let encoded = std::fs::read_to_string(&path).unwrap();
    assert!(RepositoryCatalogue::decode(encoded.trim_end_matches('\n')).is_err());
    std::fs::remove_file(path).unwrap();
}

fn sample_catalogue() -> RepositoryCatalogue {
    RepositoryCatalogue::new(vec![CatalogueEntry {
        node_id: NodeId(0),
        fact: NodeFact {
            key: NodeKey::from_identity("node"),
            kind: NodeKind::File,
            path: "src/main.rs".to_owned(),
            name: "main".to_owned(),
            content_id: Some(ContentId::from_bytes(b"content")),
            span: Some(SourceSpan::new("src/main.rs", 1, 2, 3, 4).unwrap()),
        },
    }])
    .unwrap()
}

fn node(key: NodeKey, path: &str) -> NodeFact {
    NodeFact {
        key,
        kind: NodeKind::File,
        path: path.to_owned(),
        name: String::new(),
        content_id: None,
        span: None,
    }
}

fn edge(source: NodeKey, target: NodeKey) -> EdgeFact {
    EdgeFact {
        source,
        target,
        relation: RelationKind::References,
        span: None,
    }
}
