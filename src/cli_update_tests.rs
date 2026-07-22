use std::fs;
use std::path::PathBuf;
use std::sync::atomic::{AtomicUsize, Ordering};

use arcana::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, RepositorySnapshot,
    SourceSpan,
};
use arcana::synthetic::NodeId;

use super::cli::{ImportFactsCommand, UpdateFactsCommand};
use super::cli_commands::run_import_facts;
use super::cli_update::run_update_facts;

#[test]
fn update_command_creates_verified_cumulative_overlay() {
    let directory = TestDirectory::new();
    let base_facts_path = directory.path.join("base.tsv");
    let replacement_path = directory.path.join("replacement.tsv");
    let base_output = directory.path.join("base");
    let updated_output = directory.path.join("updated");
    fs::write(&base_facts_path, facts(false).encode()).unwrap();
    fs::write(&replacement_path, facts(true).encode()).unwrap();
    run_import_facts(&ImportFactsCommand {
        facts: base_facts_path,
        output: base_output.clone(),
        adapter_name: "test-go".to_owned(),
        adapter_version: "7".to_owned(),
    })
    .unwrap();

    let summary = run_update_facts(&UpdateFactsCommand {
        base: base_output.join("repository.manifest"),
        facts: replacement_path,
        changed: vec!["src/lib.go".to_owned()],
        output: updated_output.clone(),
    })
    .unwrap();
    assert!(summary.contains("added_edges=1") && summary.contains("removed_edges=1"));
    assert!(updated_output.join("overlay.arcana").is_file());

    let snapshot = RepositorySnapshot::open(updated_output.join("repository.manifest")).unwrap();
    assert_eq!(snapshot.manifest().adapter_name, "test-go");
    let neighbors = snapshot.graph().forward_neighbors(NodeId(2)).unwrap();
    assert_eq!(neighbors.len(), 1);
    assert_eq!(neighbors[0].node, NodeId(1));
}

fn facts(reverse: bool) -> RepositoryFacts {
    let first = NodeKey::from_u64(2);
    let second = NodeKey::from_u64(3);
    let (source, target) = if reverse {
        (second, first)
    } else {
        (first, second)
    };
    RepositoryFacts {
        nodes: vec![
            node(NodeKey::from_u64(1), NodeKind::Repository, "repo", "repo"),
            node(first, NodeKind::Function, "src/lib.go", "first"),
            node(second, NodeKind::Function, "src/lib.go", "second"),
        ],
        edges: vec![EdgeFact {
            source,
            target,
            relation: RelationKind::Calls,
            span: Some(SourceSpan::new("src/lib.go", 10, 1, 10, 8).unwrap()),
        }],
        unresolved: vec![],
    }
}

fn node(key: NodeKey, kind: NodeKind, path: &str, name: &str) -> NodeFact {
    NodeFact {
        key,
        kind,
        path: path.to_owned(),
        name: name.to_owned(),
        content_id: None,
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
            "arcana-update-test-{}-{}",
            std::process::id(),
            SEQUENCE.fetch_add(1, Ordering::Relaxed)
        ));
        let _ = fs::remove_dir_all(&path);
        fs::create_dir(&path).unwrap();
        Self { path }
    }
}
impl Drop for TestDirectory {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.path);
    }
}
