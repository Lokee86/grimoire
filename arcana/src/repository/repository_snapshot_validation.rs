use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::Path;

use crate::snapshot::GraphSnapshot;
use crate::storage::{StableHasher, dataset_checksum};

use super::{
    CompiledRepository, NodeKind, RepositoryCatalogue, RepositoryFacts, RepositorySnapshotError,
    RepositorySnapshotManifest, compile_repository_facts,
};

pub(super) fn validate_components(
    manifest: &RepositorySnapshotManifest,
    graph: &GraphSnapshot,
    catalogue: &RepositoryCatalogue,
    facts: &RepositoryFacts,
    unresolved: &RepositoryFacts,
) -> Result<(), RepositorySnapshotError> {
    if !unresolved.nodes.is_empty() || !unresolved.edges.is_empty() {
        return Err(RepositorySnapshotError::InvalidUnresolvedArtifact);
    }
    let compiled = compile_repository_facts(facts)?;
    validate_compiled_components(manifest, graph, &compiled, repository_identity(facts))?;
    if &compiled.catalogue != catalogue {
        return Err(RepositorySnapshotError::ArtifactMismatch {
            field: "catalogue_contents",
            expected: manifest.catalogue_checksum,
            actual: 0,
        });
    }
    if compiled.unresolved != unresolved.unresolved {
        return Err(RepositorySnapshotError::ArtifactMismatch {
            field: "unresolved_contents",
            expected: manifest.unresolved_checksum,
            actual: 0,
        });
    }
    Ok(())
}

pub(super) fn validate_precompiled_components(
    manifest: &RepositorySnapshotManifest,
    graph: &GraphSnapshot,
    compiled: &CompiledRepository,
    facts: &RepositoryFacts,
) -> Result<(), RepositorySnapshotError> {
    validate_compiled_components(
        manifest,
        graph,
        compiled,
        repository_identity_from_checksum(facts, manifest.facts_checksum),
    )
}

fn validate_compiled_components(
    manifest: &RepositorySnapshotManifest,
    graph: &GraphSnapshot,
    compiled: &CompiledRepository,
    actual_repository_id: u64,
) -> Result<(), RepositorySnapshotError> {
    compare(
        "graph_snapshot_id",
        manifest.graph_snapshot_id,
        graph.snapshot_id(),
    )?;
    compare(
        "node_count",
        u64::from(manifest.node_count),
        u64::from(graph.node_count()),
    )?;
    compare("edge_count", manifest.edge_count, graph.edge_count())?;
    compare(
        "compiled_checksum",
        dataset_checksum(compiled.dataset.node_count, &compiled.dataset.edges),
        graph.dataset_checksum(),
    )?;
    compare(
        "catalogue_length",
        compiled.catalogue.len() as u64,
        graph.node_count() as u64,
    )?;
    compare(
        "unresolved_count",
        manifest.unresolved_count,
        compiled.unresolved.len() as u64,
    )?;
    if manifest.repository_id != actual_repository_id {
        return Err(RepositorySnapshotError::RepositoryIdentityMismatch {
            expected: manifest.repository_id,
            actual: actual_repository_id,
        });
    }
    Ok(())
}

pub(super) fn repository_identity(facts: &RepositoryFacts) -> u64 {
    unique_repository_identity(facts).unwrap_or_else(|| checksum(facts.encode().as_bytes()))
}

pub(super) fn repository_identity_from_checksum(
    facts: &RepositoryFacts,
    facts_checksum: u64,
) -> u64 {
    unique_repository_identity(facts).unwrap_or(facts_checksum)
}

fn unique_repository_identity(facts: &RepositoryFacts) -> Option<u64> {
    let mut repositories = facts
        .nodes
        .iter()
        .filter(|node| node.kind == NodeKind::Repository);
    let first = repositories.next();
    if let Some(repository) = first
        && repositories.next().is_none()
    {
        Some(repository.key.0)
    } else {
        None
    }
}

pub(super) fn read_verified(
    root: &Path,
    file: &Path,
    expected: u64,
    field: &'static str,
) -> Result<Vec<u8>, RepositorySnapshotError> {
    let bytes = fs::read(root.join(file))?;
    compare(field, expected, checksum(&bytes))?;
    Ok(bytes)
}

pub(super) fn compare(
    field: &'static str,
    expected: u64,
    actual: u64,
) -> Result<(), RepositorySnapshotError> {
    if expected == actual {
        Ok(())
    } else {
        Err(RepositorySnapshotError::ArtifactMismatch {
            field,
            expected,
            actual,
        })
    }
}

pub(super) fn checksum(bytes: &[u8]) -> u64 {
    let mut hasher = StableHasher::new();
    hasher.update(bytes);
    hasher.finish()
}

pub(super) fn text(bytes: &[u8]) -> Result<&str, RepositorySnapshotError> {
    std::str::from_utf8(bytes)
        .map_err(|_| RepositorySnapshotError::MalformedManifest("artifact is not UTF-8"))
}

pub(super) fn write_immutable(path: &Path, bytes: &[u8]) -> Result<(), RepositorySnapshotError> {
    let mut file = OpenOptions::new().write(true).create_new(true).open(path)?;
    file.write_all(bytes)?;
    file.sync_all()?;
    Ok(())
}
