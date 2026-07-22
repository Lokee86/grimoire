use super::*;

#[test]
fn partitions_and_replaces_file_owned_facts() {
    let repository = NodeKey::from_u64(1);
    let first = NodeKey::from_u64(2);
    let second = NodeKey::from_u64(3);
    let base = RepositoryFacts {
        nodes: vec![
            node(repository, NodeKind::Repository, "repo"),
            node(first, NodeKind::Function, "a.go"),
            node(second, NodeKind::Function, "b.go"),
        ],
        edges: vec![edge(first, second, "a.go")],
        unresolved: vec![],
    };
    let replacement = RepositoryFacts {
        nodes: vec![
            node(repository, NodeKind::Repository, "repo"),
            node(first, NodeKind::Function, "a.go"),
            node(second, NodeKind::Function, "b.go"),
        ],
        edges: vec![edge(second, first, "a.go")],
        unresolved: vec![],
    };
    let partitions = partition_facts(&base).unwrap();
    assert_eq!(partitions.shared.nodes.len(), 1);
    assert_eq!(partitions.files.len(), 2);
    let merged = replace_changed_files(&base, &replacement, &["a.go".to_owned()]).unwrap();
    assert!(merged.edges.iter().any(|item| item.source == second));
    assert!(!merged.edges.iter().any(|item| item.source == first));
}

fn node(key: NodeKey, kind: NodeKind, path: &str) -> NodeFact {
    NodeFact {
        key,
        kind,
        path: path.to_owned(),
        name: path.to_owned(),
        content_id: None,
        span: None,
    }
}

fn edge(source: NodeKey, target: NodeKey, path: &str) -> EdgeFact {
    EdgeFact {
        source,
        target,
        relation: RelationKind::Calls,
        span: Some(SourceSpan::new(path, 1, 1, 1, 2).unwrap()),
    }
}
