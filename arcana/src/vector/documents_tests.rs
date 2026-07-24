use crate::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, UnresolvedReason,
    UnresolvedReferenceFact,
};

use super::graph_documents;

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
