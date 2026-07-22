use super::*;

#[test]
fn creates_cumulative_edge_overlay_for_changed_file() {
    let repository = NodeKey::from_u64(1);
    let first = NodeKey::from_u64(2);
    let second = NodeKey::from_u64(3);
    let base = facts(repository, first, second, first, second);
    let replacement = facts(repository, first, second, second, first);
    let compiled = compile_repository_facts(&base).unwrap();
    let update =
        plan_file_update(&base, &replacement, &["a.go".to_owned()], &compiled.dataset).unwrap();
    assert_eq!(update.changes.added.len(), 1);
    assert_eq!(update.changes.removed.len(), 1);
}

#[test]
fn requires_rebuild_when_node_keys_change() {
    let repository = NodeKey::from_u64(1);
    let first = NodeKey::from_u64(2);
    let second = NodeKey::from_u64(3);
    let base = facts(repository, first, second, first, second);
    let replacement = facts(
        repository,
        first,
        NodeKey::from_u64(4),
        first,
        NodeKey::from_u64(4),
    );
    let compiled = compile_repository_facts(&base).unwrap();
    assert!(matches!(
        plan_file_update(
            &base,
            &replacement,
            &["a.go".to_owned(), "b.go".to_owned()],
            &compiled.dataset
        ),
        Err(IncrementalError::NodeSetChanged { .. })
    ));
}

fn facts(
    repository: NodeKey,
    first: NodeKey,
    second: NodeKey,
    source: NodeKey,
    target: NodeKey,
) -> RepositoryFacts {
    RepositoryFacts {
        nodes: vec![
            NodeFact {
                key: repository,
                kind: NodeKind::Repository,
                path: "repo".to_owned(),
                name: "repo".to_owned(),
                content_id: None,
                span: None,
            },
            NodeFact {
                key: first,
                kind: NodeKind::Function,
                path: "a.go".to_owned(),
                name: "a".to_owned(),
                content_id: None,
                span: None,
            },
            NodeFact {
                key: second,
                kind: NodeKind::Function,
                path: "b.go".to_owned(),
                name: "b".to_owned(),
                content_id: None,
                span: None,
            },
        ],
        edges: vec![EdgeFact {
            source,
            target,
            relation: RelationKind::Calls,
            span: Some(SourceSpan::new("a.go", 1, 1, 1, 2).unwrap()),
        }],
        unresolved: vec![],
    }
}
