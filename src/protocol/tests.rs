use std::fs;
use std::io::Cursor;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicUsize, Ordering};

use serde_json::Value;

use crate::repository::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts,
    UnresolvedReason, UnresolvedReferenceFact, compile_repository_facts, write_catalogue,
};
use crate::storage::write_packed;

use super::{ProtocolSnapshot, serve_jsonl};

#[test]
fn serves_repository_queries_and_snapshot_diffs() {
    let directory = TestDirectory::new();
    let current_path = directory.path.join("current");
    let other_path = directory.path.join("other");
    write_snapshot(&current_path, current_facts());
    write_snapshot(&other_path, other_facts());
    let snapshot = ProtocolSnapshot::open(&current_path).unwrap();

    let resolved = request(
        &snapshot,
        r#"{"id":"symbol","op":"resolve_symbol","name":"caller","kind":"function"}"#,
    );
    assert_eq!(resolved["id"], "symbol");
    assert_eq!(resolved["result"]["count"], 1);
    assert_eq!(resolved["result"]["nodes"][0]["node_id"], 1);

    let file = request(&snapshot, r#"{"op":"resolve_file","path":"src/lib.go"}"#);
    assert_eq!(file["result"]["count"], 1);
    assert_eq!(file["result"]["nodes"][0]["kind"], "file");

    let nodes = request(
        &snapshot,
        r#"{"op":"list_nodes","kind":"function","path_prefix":"src"}"#,
    );
    assert_eq!(nodes["result"]["count"], 2);

    let neighbors = request(
        &snapshot,
        r#"{"op":"neighbors","node_id":1,"direction":"outgoing","relation":"calls"}"#,
    );
    assert_eq!(neighbors["result"]["count"], 1);
    assert_eq!(
        neighbors["result"]["relationships"][0]["node"]["name"],
        "callee"
    );

    let unresolved = request(
        &snapshot,
        r#"{"op":"unresolved","node_id":1,"reason":"unsupported-form"}"#,
    );
    assert_eq!(unresolved["result"]["count"], 1);
    assert_eq!(
        unresolved["result"]["unresolved"][0]["expression"],
        "pkg.Call"
    );

    let stats = request(&snapshot, r#"{"op":"stats"}"#);
    assert_eq!(stats["result"]["node_count"], 3);
    assert_eq!(stats["result"]["edge_count"], 1);
    assert_eq!(stats["result"]["call_resolution"]["coverage"], 0.5);

    let diff = request(
        &snapshot,
        &format!(
            r#"{{"op":"diff","other_snapshot":{},"limit":10}}"#,
            serde_json::to_string(&other_path).unwrap()
        ),
    );
    assert_eq!(diff["result"]["counts"]["added"], 1);
    assert_eq!(diff["result"]["counts"]["removed"], 1);
    assert_eq!(diff["result"]["counts"]["metadata_changed"], 1);
    assert_eq!(diff["result"]["counts"]["relationship_changed"], 1);
    assert_eq!(diff["result"]["graph_changed"], true);
}

#[test]
fn jsonl_server_continues_after_request_errors() {
    let directory = TestDirectory::new();
    let snapshot_path = directory.path.join("snapshot");
    write_snapshot(&snapshot_path, current_facts());
    let snapshot = ProtocolSnapshot::open(snapshot_path).unwrap();
    let input = Cursor::new(
        b"{\"id\":1,\"op\":\"stats\"}\nnot-json\n{\"id\":3,\"op\":\"neighbors\",\"node_id\":99,\"direction\":\"incoming\"}\n",
    );
    let mut output = Vec::new();
    serve_jsonl(&snapshot, input, &mut output).unwrap();
    let responses = String::from_utf8(output)
        .unwrap()
        .lines()
        .map(|line| serde_json::from_str::<Value>(line).unwrap())
        .collect::<Vec<_>>();
    assert_eq!(responses.len(), 3);
    assert_eq!(responses[0]["ok"], true);
    assert_eq!(responses[1]["error"]["code"], "invalid_json");
    assert_eq!(responses[2]["id"], 3);
    assert_eq!(responses[2]["error"]["code"], "unknown_node");
}

fn request(snapshot: &ProtocolSnapshot, line: &str) -> Value {
    let response = snapshot.handle_line(line);
    assert_eq!(response["protocol"], "arcana.query.v1");
    assert_eq!(response["ok"], true, "response was {response}");
    response
}

fn write_snapshot(path: &Path, facts: RepositoryFacts) {
    fs::create_dir(path).unwrap();
    let compiled = compile_repository_facts(&facts).unwrap();
    write_packed(path.join("graph.arcana"), &compiled.dataset).unwrap();
    write_catalogue(path.join("catalogue.tsv"), &compiled.catalogue).unwrap();
    let unresolved = RepositoryFacts::with_unresolved(Vec::new(), Vec::new(), compiled.unresolved);
    fs::write(path.join("unresolved.tsv"), unresolved.encode()).unwrap();
}

fn current_facts() -> RepositoryFacts {
    RepositoryFacts::with_unresolved(
        vec![
            node(10, NodeKind::File, "src/lib.go", "lib.go", None),
            node(20, NodeKind::Function, "src/lib.go", "caller", Some(1)),
            node(30, NodeKind::Function, "src/lib.go", "callee", Some(2)),
        ],
        vec![EdgeFact {
            source: NodeKey::from_u64(20),
            target: NodeKey::from_u64(30),
            relation: RelationKind::Calls,
            span: None,
        }],
        vec![UnresolvedReferenceFact {
            source: NodeKey::from_u64(20),
            relation: RelationKind::Calls,
            expression: "pkg.Call".to_owned(),
            candidate_namespace: Some("pkg".to_owned()),
            candidate_name: Some("Call".to_owned()),
            reason: UnresolvedReason::UnsupportedForm,
            span: None,
        }],
    )
}

fn other_facts() -> RepositoryFacts {
    RepositoryFacts::new(
        vec![
            node(10, NodeKind::File, "src/lib.go", "lib.go", None),
            node(20, NodeKind::Function, "src/lib.go", "caller", Some(99)),
            node(40, NodeKind::Function, "src/lib.go", "replacement", Some(4)),
        ],
        vec![EdgeFact {
            source: NodeKey::from_u64(20),
            target: NodeKey::from_u64(40),
            relation: RelationKind::Calls,
            span: None,
        }],
    )
}

fn node(key: u64, kind: NodeKind, path: &str, name: &str, content: Option<u64>) -> NodeFact {
    NodeFact {
        key: NodeKey::from_u64(key),
        kind,
        path: path.to_owned(),
        name: name.to_owned(),
        content_id: content.map(ContentId::from_u64),
        span: None,
    }
}

struct TestDirectory {
    path: PathBuf,
}

impl TestDirectory {
    fn new() -> Self {
        static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
        let path = std::env::temp_dir().join(format!(
            "arcana-protocol-test-{}-{}",
            std::process::id(),
            SEQUENCE.fetch_add(1, Ordering::Relaxed)
        ));
        fs::create_dir(&path).unwrap();
        Self { path }
    }
}

impl Drop for TestDirectory {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.path);
    }
}
