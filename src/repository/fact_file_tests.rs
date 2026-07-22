use super::fact_file::{encode_facts, parse_facts};
use super::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, SourceSpan,
};

#[test]
fn fact_file_round_trips_and_escapes_fields() {
    let facts = RepositoryFacts {
        nodes: vec![NodeFact {
            key: NodeKey::from_identity("node\t1"),
            kind: NodeKind::Function,
            path: "src/main.rs".to_owned(),
            name: "line\t\nname\\".to_owned(),
            content_id: Some(ContentId::from_bytes(b"body")),
            span: Some(SourceSpan::new("src\\main.rs", 2, 3, 4, 5).unwrap()),
        }],
        edges: vec![EdgeFact {
            source: NodeKey::from_identity("source"),
            target: NodeKey::from_identity("target"),
            relation: RelationKind::References,
            span: None,
        }],
    };

    let encoded = encode_facts(&facts);
    assert!(encoded.contains("\\t") && encoded.contains("\\n") && encoded.contains("\\\\"));
    assert_eq!(parse_facts(&encoded).unwrap(), facts);
}

#[test]
fn encoding_sorts_records_deterministically() {
    let mut facts = RepositoryFacts::default();
    facts.nodes.push(NodeFact {
        key: NodeKey::from_u64(2),
        kind: NodeKind::File,
        path: "b".to_owned(),
        name: String::new(),
        content_id: None,
        span: None,
    });
    facts.nodes.push(NodeFact {
        key: NodeKey::from_u64(1),
        kind: NodeKind::File,
        path: "a".to_owned(),
        name: String::new(),
        content_id: None,
        span: None,
    });
    assert_eq!(
        encode_facts(&facts),
        encode_facts(&RepositoryFacts {
            nodes: facts.nodes.iter().rev().cloned().collect(),
            edges: Vec::new(),
        })
    );
}
