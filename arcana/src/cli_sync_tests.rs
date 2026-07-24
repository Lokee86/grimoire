use std::fs;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicUsize, Ordering};

use arcana::repository::{NodeKind, RelationKind, RepositorySnapshot};
use serde_json::{Value, json};
use sha2::{Digest, Sha256};

use super::cli::SyncCommand;
use super::cli_sync::run_sync;

#[test]
fn sync_builds_reuses_and_registers_a_lexicon_snapshot() {
    let directory = TestDirectory::new();
    let lexicon = directory.path.join(".lexicon");
    let state = directory.path.join(".arcana");
    fs::create_dir_all(lexicon.join("objects")).unwrap();
    fs::create_dir_all(lexicon.join("snapshots")).unwrap();

    let repository_id = sha_id("repository-node");
    let object_id = write_object(
        &lexicon,
        vec![json!({
            "record": "node",
            "id": repository_id,
            "kind": "repository",
            "path": ".lexicon-repository",
            "name": "example/repository",
            "qualified_name": "example/repository"
        })],
    );
    let manifest = json!({
        "version": 1,
        "state_commit": "state",
        "languages": [{
            "language": "go",
            "adapter_version": "1",
            "schema_version": 1,
            "repository": "example/repository",
            "analysis_config_id": sha_id("config"),
            "shared_object_id": object_id,
            "files": []
        }]
    });
    let snapshot_id = write_snapshot(&lexicon, &manifest);
    fs::write(lexicon.join("CURRENT"), format!("{snapshot_id}\n")).unwrap();

    let summary = run_sync(&SyncCommand {
        lexicon: lexicon.clone(),
        state: state.clone(),
        register: true,
    })
    .unwrap();
    assert!(summary.contains("mode=rebuild") && summary.contains("registered=true"));
    assert_eq!(
        fs::read_to_string(state.join("CURRENT")).unwrap().trim(),
        snapshot_id
    );

    let digest = snapshot_id.strip_prefix("sha256:").unwrap();
    let output = state.join("snapshots").join(digest);
    let snapshot = RepositorySnapshot::open(output.join("repository.manifest")).unwrap();
    assert_eq!(snapshot.facts().nodes.len(), 1);
    assert_eq!(
        fs::read_to_string(output.join("lexicon.snapshot"))
            .unwrap()
            .trim(),
        snapshot_id
    );

    let registration: Value =
        serde_json::from_slice(&fs::read(lexicon.join("consumers").join("arcana.json")).unwrap())
            .unwrap();
    assert_eq!(registration["version"], 1);
    assert_eq!(registration["args"][0], "sync");

    let summary = run_sync(&SyncCommand {
        lexicon,
        state,
        register: false,
    })
    .unwrap();
    assert!(summary.contains("mode=existing"));
}

fn write_object(root: &Path, records: Vec<Value>) -> String {
    let object = json!({
        "version": 1,
        "language": "go",
        "owner": null,
        "source_content_id": null,
        "adapter_version": "1",
        "schema_version": 1,
        "analysis_config_id": sha_id("config"),
        "records": records
    });
    let bytes = serde_json::to_vec(&object).unwrap();
    let id = domain_id("lexicon:fact-object:v1\0", &bytes);
    let digest = id.strip_prefix("sha256:").unwrap();
    let directory = root.join("objects").join(&digest[..2]);
    fs::create_dir_all(&directory).unwrap();
    fs::write(directory.join(&digest[2..]), bytes).unwrap();
    id
}

fn write_snapshot(root: &Path, manifest: &Value) -> String {
    let bytes = serde_json::to_vec(manifest).unwrap();
    let id = domain_id("lexicon:snapshot:v1\0", &bytes);
    let digest = id.strip_prefix("sha256:").unwrap();
    fs::write(root.join("snapshots").join(format!("{digest}.json")), bytes).unwrap();
    id
}

fn domain_id(domain: &str, bytes: &[u8]) -> String {
    let mut hasher = Sha256::new();
    hasher.update(domain.as_bytes());
    hasher.update(bytes);
    format!("sha256:{:x}", hasher.finalize())
}

fn sha_id(value: &str) -> String {
    format!("sha256:{:x}", Sha256::digest(value.as_bytes()))
}

struct TestDirectory {
    path: PathBuf,
}

impl TestDirectory {
    fn new() -> Self {
        static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
        let path = std::env::temp_dir().join(format!(
            "arcana-sync-test-{}-{}",
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

#[test]
fn sync_builds_graph_from_typescript_exports_and_gdscript_signals() {
    let directory = TestDirectory::new();
    let lexicon = directory.path.join(".lexicon");
    let state = directory.path.join(".arcana");
    fs::create_dir_all(lexicon.join("objects")).unwrap();
    fs::create_dir_all(lexicon.join("snapshots")).unwrap();

    let repository_id = sha_id("mixed-repository");
    let export_id = sha_id("typescript-export");
    let signal_id = sha_id("gdscript-signal");
    let caller_id = sha_id("typescript-caller");
    let callee_id = sha_id("gdscript-callee");
    let repository = json!({
        "record": "node",
        "id": repository_id,
        "kind": "repository",
        "path": ".lexicon-repository",
        "name": "example/mixed",
        "qualified_name": "example/mixed"
    });
    let gdscript_object = write_language_object(
        &lexicon,
        "gdscript",
        vec![
            repository.clone(),
            json!({
                "record": "node",
                "id": signal_id,
                "kind": "signal",
                "path": "src/player.gd",
                "name": "finished",
                "qualified_name": "Player.finished"
            }),
            json!({
                "record": "node",
                "id": callee_id,
                "kind": "method",
                "path": "src/player.gd",
                "name": "finish",
                "qualified_name": "Player.finish"
            }),
        ],
    );
    let typescript_object = write_language_object(
        &lexicon,
        "typescript",
        vec![
            repository,
            json!({
                "record": "node",
                "id": export_id,
                "kind": "export",
                "path": "src/api.ts",
                "name": "run",
                "qualified_name": "src/api.ts::run"
            }),
            json!({
                "record": "node",
                "id": caller_id,
                "kind": "function",
                "path": "src/api.ts",
                "name": "run",
                "qualified_name": "src/api.ts::run_impl"
            }),
            json!({
                "record": "edge",
                "relation": "calls",
                "source": caller_id,
                "target": callee_id
            }),
        ],
    );
    let manifest = json!({
        "version": 1,
        "state_commit": "state",
        "languages": [
            language_entry("gdscript", gdscript_object),
            language_entry("typescript", typescript_object)
        ]
    });
    let snapshot_id = write_snapshot(&lexicon, &manifest);
    fs::write(lexicon.join("CURRENT"), format!("{snapshot_id}\n")).unwrap();

    run_sync(&SyncCommand {
        lexicon,
        state: state.clone(),
        register: false,
    })
    .unwrap();

    let digest = snapshot_id.strip_prefix("sha256:").unwrap();
    let snapshot = RepositorySnapshot::open(
        state
            .join("snapshots")
            .join(digest)
            .join("repository.manifest"),
    )
    .unwrap();
    assert!(
        snapshot
            .facts()
            .nodes
            .iter()
            .any(|node| node.kind == NodeKind::Export)
    );
    assert!(
        snapshot
            .facts()
            .nodes
            .iter()
            .any(|node| node.kind == NodeKind::Signal)
    );
    assert!(
        snapshot
            .facts()
            .edges
            .iter()
            .any(|edge| edge.relation == RelationKind::Calls)
    );
}

fn write_language_object(root: &Path, language: &str, records: Vec<Value>) -> String {
    let object = json!({
        "version": 1,
        "language": language,
        "owner": null,
        "source_content_id": null,
        "adapter_version": "1",
        "schema_version": 1,
        "analysis_config_id": sha_id("config"),
        "records": records
    });
    let bytes = serde_json::to_vec(&object).unwrap();
    let id = domain_id("lexicon:fact-object:v1\0", &bytes);
    let digest = id.strip_prefix("sha256:").unwrap();
    let directory = root.join("objects").join(&digest[..2]);
    fs::create_dir_all(&directory).unwrap();
    fs::write(directory.join(&digest[2..]), bytes).unwrap();
    id
}

fn language_entry(language: &str, object_id: String) -> Value {
    json!({
        "language": language,
        "adapter_version": "1",
        "schema_version": 1,
        "repository": "example/mixed",
        "analysis_config_id": sha_id("config"),
        "shared_object_id": object_id,
        "files": []
    })
}
