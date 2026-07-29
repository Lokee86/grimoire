use std::fs;
use std::sync::atomic::{AtomicUsize, Ordering};

use crate::snapshot::publish_snapshot;
use crate::storage::write_packed;

use super::*;

#[test]
fn binds_graph_catalogue_unresolved_and_source_facts() {
    let directory = test_directory();
    let facts = sample_facts();
    let compiled = compile_repository_facts(&facts).unwrap();
    write_packed(directory.join("graph.arcana"), &compiled.dataset).unwrap();
    publish_snapshot(directory.join("graph.manifest"), "graph.arcana", None, 7).unwrap();
    write_catalogue(directory.join("catalogue.tsv"), &compiled.catalogue).unwrap();
    fs::write(
        directory.join("unresolved.tsv"),
        RepositoryFacts::with_unresolved(vec![], vec![], compiled.unresolved.clone()).encode(),
    )
    .unwrap();
    fs::write(directory.join("facts.tsv"), facts.encode()).unwrap();
    publish_repository_snapshot(
        directory.join(REPOSITORY_MANIFEST_FILE),
        PublishRepositorySnapshot {
            graph_manifest_file: std::path::Path::new("graph.manifest"),
            catalogue_file: std::path::Path::new("catalogue.tsv"),
            unresolved_file: std::path::Path::new("unresolved.tsv"),
            facts_file: std::path::Path::new("facts.tsv"),
            adapter_name: "test",
            adapter_version: "1",
            created_unix_seconds: 7,
        },
    )
    .unwrap();
    let snapshot = RepositorySnapshot::open(directory.join(REPOSITORY_MANIFEST_FILE)).unwrap();
    assert_eq!(snapshot.catalogue().len(), 3);
    assert_eq!(snapshot.graph().edge_count(), 1);
    fs::write(directory.join("catalogue.tsv"), "tampered\n").unwrap();
    assert!(RepositorySnapshot::open(directory.join(REPOSITORY_MANIFEST_FILE)).is_err());
    fs::remove_dir_all(directory).unwrap();
}

#[test]
fn precompiled_publication_matches_standalone_validation_without_recompiling() {
    let legacy = test_directory();
    let optimized = test_directory();
    let facts = sample_facts();
    compiler::reset_compile_invocation_count();
    let compiled = compile_repository_facts(&facts).unwrap();
    assert_eq!(compiler::compile_invocation_count(), 1);

    let legacy_checksums = write_compiled_artifacts(&legacy, &compiled, &facts);
    let optimized_checksums = write_compiled_artifacts(&optimized, &compiled, &facts);
    assert_eq!(optimized_checksums, legacy_checksums);

    let request = || PublishRepositorySnapshot {
        graph_manifest_file: std::path::Path::new("graph.manifest"),
        catalogue_file: std::path::Path::new("catalogue.tsv"),
        unresolved_file: std::path::Path::new("unresolved.tsv"),
        facts_file: std::path::Path::new("facts.tsv"),
        adapter_name: "test",
        adapter_version: "1",
        created_unix_seconds: 7,
    };
    let optimized_manifest = publish_precompiled_repository_snapshot(
        optimized.join(REPOSITORY_MANIFEST_FILE),
        request(),
        &compiled,
        &facts,
        optimized_checksums,
    )
    .unwrap();
    assert_eq!(
        compiler::compile_invocation_count(),
        1,
        "precompiled publication must reuse the caller's compilation"
    );

    let legacy_manifest =
        publish_repository_snapshot(legacy.join(REPOSITORY_MANIFEST_FILE), request()).unwrap();
    assert_eq!(compiler::compile_invocation_count(), 2);
    assert_eq!(optimized_manifest, legacy_manifest);
    for file in [
        "graph.arcana",
        "graph.manifest",
        "catalogue.tsv",
        "unresolved.tsv",
        "facts.tsv",
        REPOSITORY_MANIFEST_FILE,
    ] {
        assert_eq!(
            fs::read(optimized.join(file)).unwrap(),
            fs::read(legacy.join(file)).unwrap(),
            "{file} differs"
        );
    }

    RepositorySnapshot::open(optimized.join(REPOSITORY_MANIFEST_FILE)).unwrap();
    fs::remove_dir_all(legacy).unwrap();
    fs::remove_dir_all(optimized).unwrap();
}

#[test]
fn protocol_parts_transfer_owned_components() {
    let directory = test_directory();
    let facts = sample_facts();
    let compiled = compile_repository_facts(&facts).unwrap();
    write_packed(directory.join("graph.arcana"), &compiled.dataset).unwrap();
    publish_snapshot(directory.join("graph.manifest"), "graph.arcana", None, 7).unwrap();
    write_catalogue(directory.join("catalogue.tsv"), &compiled.catalogue).unwrap();
    fs::write(
        directory.join("unresolved.tsv"),
        RepositoryFacts::with_unresolved(vec![], vec![], compiled.unresolved.clone()).encode(),
    )
    .unwrap();
    fs::write(directory.join("facts.tsv"), facts.encode()).unwrap();
    publish_repository_snapshot(
        directory.join(REPOSITORY_MANIFEST_FILE),
        PublishRepositorySnapshot {
            graph_manifest_file: std::path::Path::new("graph.manifest"),
            catalogue_file: std::path::Path::new("catalogue.tsv"),
            unresolved_file: std::path::Path::new("unresolved.tsv"),
            facts_file: std::path::Path::new("facts.tsv"),
            adapter_name: "test",
            adapter_version: "1",
            created_unix_seconds: 7,
        },
    )
    .unwrap();

    let snapshot = RepositorySnapshot::open(directory.join(REPOSITORY_MANIFEST_FILE)).unwrap();
    let (graph, catalogue, unresolved) = snapshot.into_protocol_parts();
    assert_eq!(graph.edge_count(), 1);
    assert_eq!(catalogue.len(), 3);
    assert_eq!(unresolved.unresolved, compiled.unresolved);
    fs::remove_dir_all(directory).unwrap();
}

fn write_compiled_artifacts(
    directory: &std::path::Path,
    compiled: &CompiledRepository,
    facts: &RepositoryFacts,
) -> RepositoryArtifactChecksums {
    write_packed(directory.join("graph.arcana"), &compiled.dataset).unwrap();
    publish_snapshot(directory.join("graph.manifest"), "graph.arcana", None, 7).unwrap();

    let catalogue = compiled.catalogue.encode().unwrap();
    let catalogue_checksum = repository_artifact_checksum(catalogue.as_bytes());
    fs::write(directory.join("catalogue.tsv"), catalogue).unwrap();

    let unresolved =
        RepositoryFacts::with_unresolved(vec![], vec![], compiled.unresolved.clone()).encode();
    let unresolved_checksum = repository_artifact_checksum(unresolved.as_bytes());
    fs::write(directory.join("unresolved.tsv"), unresolved).unwrap();

    let encoded_facts = facts.encode();
    let facts_checksum = repository_artifact_checksum(encoded_facts.as_bytes());
    fs::write(directory.join("facts.tsv"), encoded_facts).unwrap();

    RepositoryArtifactChecksums {
        catalogue: catalogue_checksum,
        unresolved: unresolved_checksum,
        facts: facts_checksum,
    }
}

fn sample_facts() -> RepositoryFacts {
    RepositoryFacts {
        nodes: vec![
            NodeFact {
                key: NodeKey::from_u64(1),
                external_identity: None,
                kind: NodeKind::Repository,
                path: "repo".to_owned(),
                name: "repo".to_owned(),
                qualified_name: "repo".to_owned(),
                content_id: None,
                span: None,
            },
            NodeFact {
                key: NodeKey::from_u64(2),
                external_identity: None,
                kind: NodeKind::Function,
                path: "a.go".to_owned(),
                name: "a".to_owned(),
                qualified_name: "a".to_owned(),
                content_id: None,
                span: None,
            },
            NodeFact {
                key: NodeKey::from_u64(3),
                external_identity: None,
                kind: NodeKind::Function,
                path: "b.go".to_owned(),
                name: "b".to_owned(),
                qualified_name: "b".to_owned(),
                content_id: None,
                span: None,
            },
        ],
        edges: vec![EdgeFact {
            source: NodeKey::from_u64(2),
            target: NodeKey::from_u64(3),
            relation: RelationKind::Calls,
            span: None,
        }],
        unresolved: vec![],
    }
}

fn test_directory() -> std::path::PathBuf {
    static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
    let path = std::env::temp_dir().join(format!(
        "arcana-repository-snapshot-{}-{}",
        std::process::id(),
        SEQUENCE.fetch_add(1, Ordering::Relaxed)
    ));
    let _ = fs::remove_dir_all(&path);
    fs::create_dir(&path).unwrap();
    path
}
