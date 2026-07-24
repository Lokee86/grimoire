use std::fs;
use std::path::{Path, PathBuf};

use crate::snapshot::GraphSnapshot;
use crate::storage::{QueryError, StableHasher};
use crate::synthetic::GraphDataset;

use super::fact_file::FACT_SCHEMA_VERSION;
use super::repository_snapshot_validation::{
    checksum, read_verified, repository_identity, text, validate_components, write_immutable,
};
use super::{
    RepositoryCatalogue, RepositoryFacts, RepositorySnapshotError, RepositorySnapshotManifest,
};

pub const REPOSITORY_MANIFEST_FILE: &str = "repository.manifest";

#[derive(Clone, Debug)]
pub struct RepositorySnapshot {
    root: PathBuf,
    manifest: RepositorySnapshotManifest,
    graph: GraphSnapshot,
    catalogue: RepositoryCatalogue,
    facts: RepositoryFacts,
    unresolved: RepositoryFacts,
}

impl RepositorySnapshot {
    pub fn open(path: impl AsRef<Path>) -> Result<Self, RepositorySnapshotError> {
        let path = path.as_ref();
        let root = path
            .parent()
            .filter(|path| !path.as_os_str().is_empty())
            .unwrap_or_else(|| Path::new("."))
            .to_path_buf();
        let manifest = RepositorySnapshotManifest::decode(&fs::read_to_string(path)?)?;
        if !(2..=FACT_SCHEMA_VERSION).contains(&manifest.fact_schema_version) {
            return Err(RepositorySnapshotError::UnsupportedFactSchema(
                manifest.fact_schema_version,
            ));
        }
        read_verified(
            &root,
            &manifest.graph_manifest_file,
            manifest.graph_manifest_checksum,
            "graph_manifest_checksum",
        )?;
        let catalogue_bytes = read_verified(
            &root,
            &manifest.catalogue_file,
            manifest.catalogue_checksum,
            "catalogue_checksum",
        )?;
        let unresolved_bytes = read_verified(
            &root,
            &manifest.unresolved_file,
            manifest.unresolved_checksum,
            "unresolved_checksum",
        )?;
        let facts_bytes = read_verified(
            &root,
            &manifest.facts_file,
            manifest.facts_checksum,
            "facts_checksum",
        )?;
        let graph = GraphSnapshot::open(root.join(&manifest.graph_manifest_file))?;
        let catalogue = RepositoryCatalogue::decode(text(&catalogue_bytes)?)?;
        let unresolved = RepositoryFacts::parse(text(&unresolved_bytes)?)?;
        let facts = RepositoryFacts::parse(text(&facts_bytes)?)?;
        validate_components(&manifest, &graph, &catalogue, &facts, &unresolved)?;
        Ok(Self {
            root,
            manifest,
            graph,
            catalogue,
            facts,
            unresolved,
        })
    }

    pub const fn manifest(&self) -> &RepositorySnapshotManifest {
        &self.manifest
    }
    pub const fn graph(&self) -> &GraphSnapshot {
        &self.graph
    }
    pub const fn catalogue(&self) -> &RepositoryCatalogue {
        &self.catalogue
    }
    pub const fn facts(&self) -> &RepositoryFacts {
        &self.facts
    }
    pub const fn unresolved(&self) -> &RepositoryFacts {
        &self.unresolved
    }

    /// Transfers the components needed by the query protocol without cloning them.
    pub fn into_protocol_parts(self) -> (GraphSnapshot, RepositoryCatalogue, RepositoryFacts) {
        (self.graph, self.catalogue, self.unresolved)
    }

    pub fn root(&self) -> &Path {
        &self.root
    }
    pub fn base_graph_path(&self) -> PathBuf {
        self.root.join(&self.graph.manifest().base_file)
    }
    pub fn materialize_base_dataset(&self) -> Result<GraphDataset, QueryError> {
        self.graph.materialize_base_dataset()
    }
}

pub struct PublishRepositorySnapshot<'a> {
    pub graph_manifest_file: &'a Path,
    pub catalogue_file: &'a Path,
    pub unresolved_file: &'a Path,
    pub facts_file: &'a Path,
    pub adapter_name: &'a str,
    pub adapter_version: &'a str,
    pub created_unix_seconds: u64,
}

pub fn publish_repository_snapshot(
    manifest_path: impl AsRef<Path>,
    request: PublishRepositorySnapshot<'_>,
) -> Result<RepositorySnapshotManifest, RepositorySnapshotError> {
    let manifest_path = manifest_path.as_ref();
    let root = manifest_path
        .parent()
        .filter(|path| !path.as_os_str().is_empty())
        .unwrap_or_else(|| Path::new("."));
    let graph = GraphSnapshot::open(root.join(request.graph_manifest_file))?;
    let graph_manifest_checksum = checksum(&fs::read(root.join(request.graph_manifest_file))?);
    let catalogue_bytes = fs::read(root.join(request.catalogue_file))?;
    let unresolved_bytes = fs::read(root.join(request.unresolved_file))?;
    let facts_bytes = fs::read(root.join(request.facts_file))?;
    let catalogue = RepositoryCatalogue::decode(text(&catalogue_bytes)?)?;
    let unresolved = RepositoryFacts::parse(text(&unresolved_bytes)?)?;
    let facts = RepositoryFacts::parse(text(&facts_bytes)?)?;
    let repository_id = repository_identity(&facts);
    let catalogue_checksum = checksum(&catalogue_bytes);
    let unresolved_checksum = checksum(&unresolved_bytes);
    let facts_checksum = checksum(&facts_bytes);
    let snapshot_id = derive_repository_snapshot_id(
        repository_id,
        graph.snapshot_id(),
        catalogue_checksum,
        unresolved_checksum,
        facts_checksum,
        request.adapter_name,
        request.adapter_version,
    );
    let manifest = RepositorySnapshotManifest {
        snapshot_id,
        created_unix_seconds: request.created_unix_seconds,
        repository_id,
        adapter_name: request.adapter_name.to_owned(),
        adapter_version: request.adapter_version.to_owned(),
        fact_schema_version: FACT_SCHEMA_VERSION,
        node_count: graph.node_count(),
        edge_count: graph.edge_count(),
        unresolved_count: unresolved.unresolved.len() as u64,
        graph_snapshot_id: graph.snapshot_id(),
        graph_manifest_checksum,
        catalogue_checksum,
        unresolved_checksum,
        facts_checksum,
        graph_manifest_file: request.graph_manifest_file.to_path_buf(),
        catalogue_file: request.catalogue_file.to_path_buf(),
        unresolved_file: request.unresolved_file.to_path_buf(),
        facts_file: request.facts_file.to_path_buf(),
    };
    validate_components(&manifest, &graph, &catalogue, &facts, &unresolved)?;
    write_immutable(manifest_path, manifest.encode()?.as_bytes())?;
    Ok(manifest)
}

pub fn derive_repository_snapshot_id(
    repository_id: u64,
    graph_snapshot_id: u64,
    catalogue_checksum: u64,
    unresolved_checksum: u64,
    facts_checksum: u64,
    adapter_name: &str,
    adapter_version: &str,
) -> u64 {
    let mut hasher = StableHasher::new();
    hasher.update(b"arcana-repository-snapshot-v1");
    for value in [
        repository_id,
        graph_snapshot_id,
        catalogue_checksum,
        unresolved_checksum,
        facts_checksum,
    ] {
        hasher.update(&value.to_le_bytes());
    }
    hasher.update(adapter_name.as_bytes());
    hasher.update(&[0]);
    hasher.update(adapter_version.as_bytes());
    hasher.finish()
}
