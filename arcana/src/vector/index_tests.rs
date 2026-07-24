use std::fs;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicUsize, Ordering};

use crate::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, PublishRepositorySnapshot, RelationKind,
    RepositoryFacts, compile_repository_facts, publish_repository_snapshot, write_catalogue,
};
use crate::snapshot::publish_snapshot;
use crate::storage::write_packed;

use super::{
    Embedder, EmbeddingError, IndexManifest, build_current_index, current_index_directory,
    search_current_index,
};

#[test]
fn builds_reuses_and_searches_current_graph_index() {
    let directory = TestDirectory::new();
    let state = directory.path.join(".arcana");
    let digest = "a".repeat(64);
    let snapshot_directory = state.join("snapshots").join(&digest);
    fs::create_dir_all(&snapshot_directory).unwrap();

    let create = NodeKey::from_u64(1);
    let insert = NodeKey::from_u64(2);
    let facts = RepositoryFacts {
        nodes: vec![
            node(create, "create_profile", "src/profile.rs"),
            node(insert, "insert_profile", "src/repository.rs"),
        ],
        edges: vec![EdgeFact {
            source: create,
            target: insert,
            relation: RelationKind::Calls,
            span: None,
        }],
        unresolved: Vec::new(),
    };
    publish_test_snapshot(&snapshot_directory, &facts);
    fs::write(state.join("CURRENT"), format!("sha256:{digest}\n")).unwrap();

    let embedder = FakeEmbedder;
    let built = build_current_index(&state, &embedder, 1).unwrap();
    assert_eq!(built.mode, "built");
    assert_eq!(built.item_count, 2);
    assert_eq!(built.dimensions, 4);
    assert_eq!(
        built.directory,
        current_index_directory(&state, "fake-4d").unwrap()
    );

    let reused = build_current_index(&state, &embedder, 2).unwrap();
    assert_eq!(reused.mode, "existing");

    let hits = search_current_index(&state, &embedder, "create profile", 1).unwrap();
    assert_eq!(hits.len(), 1);
    assert_eq!(hits[0].name, "create_profile");
    assert_eq!(hits[0].score, 1.0);

    let manifest_path = built.directory.join("manifest.json");
    let mut manifest: IndexManifest =
        serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    manifest.repository_snapshot_id = "0000000000000000".to_owned();
    fs::write(
        &manifest_path,
        serde_json::to_vec_pretty(&manifest).unwrap(),
    )
    .unwrap();
    let error = search_current_index(&state, &embedder, "create profile", 1).unwrap_err();
    assert!(error.to_string().contains("stale"));
}

fn publish_test_snapshot(directory: &Path, facts: &RepositoryFacts) {
    let compiled = compile_repository_facts(facts).unwrap();
    write_packed(directory.join("graph.arcana"), &compiled.dataset).unwrap();
    publish_snapshot(directory.join("graph.manifest"), "graph.arcana", None, 1).unwrap();
    write_catalogue(directory.join("catalogue.tsv"), &compiled.catalogue).unwrap();
    let unresolved =
        RepositoryFacts::with_unresolved(Vec::new(), Vec::new(), compiled.unresolved.clone());
    fs::write(directory.join("unresolved.tsv"), unresolved.encode()).unwrap();
    fs::write(directory.join("facts.tsv"), facts.canonicalized().encode()).unwrap();
    publish_repository_snapshot(
        directory.join("repository.manifest"),
        PublishRepositorySnapshot {
            graph_manifest_file: Path::new("graph.manifest"),
            catalogue_file: Path::new("catalogue.tsv"),
            unresolved_file: Path::new("unresolved.tsv"),
            facts_file: Path::new("facts.tsv"),
            adapter_name: "test",
            adapter_version: "1",
            created_unix_seconds: 1,
        },
    )
    .unwrap();
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

struct FakeEmbedder;

impl Embedder for FakeEmbedder {
    fn model(&self) -> &str {
        "fake-model"
    }

    fn identity(&self) -> &str {
        "fake-4d"
    }

    fn dimensions(&self) -> usize {
        4
    }

    fn embed_documents(&self, documents: &[String]) -> Result<Vec<Vec<f32>>, EmbeddingError> {
        Ok(documents
            .iter()
            .map(|document| {
                if document.contains("name: create_profile\n") {
                    vec![1.0, 0.0, 0.0, 0.0]
                } else {
                    vec![0.0, 1.0, 0.0, 0.0]
                }
            })
            .collect())
    }

    fn embed_query(&self, _query: &str) -> Result<Vec<f32>, EmbeddingError> {
        Ok(vec![1.0, 0.0, 0.0, 0.0])
    }
}

struct TestDirectory {
    path: PathBuf,
}

impl TestDirectory {
    fn new() -> Self {
        static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
        let path = std::env::temp_dir().join(format!(
            "arcana-vector-index-test-{}-{}",
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
