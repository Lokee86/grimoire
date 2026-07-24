use std::fmt;
use std::fs;
use std::io;
use std::path::Path;
use std::time::{SystemTime, UNIX_EPOCH};

use arcana::repository::{
    CatalogueError, CompiledRepository, FactFileError, IncrementalError, PublishRepositorySnapshot,
    RepositoryCompileError, RepositoryFacts, RepositorySnapshotError, publish_repository_snapshot,
};
use arcana::snapshot::{OverlayError, SnapshotError, publish_snapshot};
use arcana::storage::{PackedError, QueryError};
use arcana::synthetic::NodeId;

use crate::cli::ImportFactsCommand;

#[derive(Debug)]
pub enum CliCommandError {
    Io(io::Error),
    Facts(FactFileError),
    Compile(RepositoryCompileError),
    Packed(PackedError),
    Query(QueryError),
    Catalogue(CatalogueError),
    RepositorySnapshot(RepositorySnapshotError),
    Incremental(IncrementalError),
    Overlay(OverlayError),
    Snapshot(SnapshotError),
    UnknownEdgeKind(u16),
    MissingCatalogueNode(NodeId),
}

impl fmt::Display for CliCommandError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::Facts(error) => error.fmt(formatter),
            Self::Compile(error) => error.fmt(formatter),
            Self::Packed(error) => error.fmt(formatter),
            Self::Query(error) => error.fmt(formatter),
            Self::Catalogue(error) => error.fmt(formatter),
            Self::RepositorySnapshot(error) => error.fmt(formatter),
            Self::Incremental(error) => error.fmt(formatter),
            Self::Overlay(error) => error.fmt(formatter),
            Self::Snapshot(error) => error.fmt(formatter),
            Self::UnknownEdgeKind(kind) => {
                write!(formatter, "graph contains unknown edge kind {kind}")
            }
            Self::MissingCatalogueNode(node) => write!(
                formatter,
                "catalogue has no metadata for graph node {}",
                node.0
            ),
        }
    }
}

impl std::error::Error for CliCommandError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            Self::Facts(error) => Some(error),
            Self::Compile(error) => Some(error),
            Self::Packed(error) => Some(error),
            Self::Query(error) => Some(error),
            Self::Catalogue(error) => Some(error),
            Self::RepositorySnapshot(error) => Some(error),
            Self::Incremental(error) => Some(error),
            Self::Overlay(error) => Some(error),
            Self::Snapshot(error) => Some(error),
            Self::UnknownEdgeKind(_) | Self::MissingCatalogueNode(_) => None,
        }
    }
}

impl From<io::Error> for CliCommandError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}
impl From<FactFileError> for CliCommandError {
    fn from(error: FactFileError) -> Self {
        Self::Facts(error)
    }
}
impl From<RepositoryCompileError> for CliCommandError {
    fn from(error: RepositoryCompileError) -> Self {
        Self::Compile(error)
    }
}
impl From<PackedError> for CliCommandError {
    fn from(error: PackedError) -> Self {
        Self::Packed(error)
    }
}
impl From<QueryError> for CliCommandError {
    fn from(error: QueryError) -> Self {
        Self::Query(error)
    }
}
impl From<CatalogueError> for CliCommandError {
    fn from(error: CatalogueError) -> Self {
        Self::Catalogue(error)
    }
}
impl From<RepositorySnapshotError> for CliCommandError {
    fn from(error: RepositorySnapshotError) -> Self {
        Self::RepositorySnapshot(error)
    }
}
impl From<IncrementalError> for CliCommandError {
    fn from(error: IncrementalError) -> Self {
        Self::Incremental(error)
    }
}
impl From<OverlayError> for CliCommandError {
    fn from(error: OverlayError) -> Self {
        Self::Overlay(error)
    }
}
impl From<SnapshotError> for CliCommandError {
    fn from(error: SnapshotError) -> Self {
        Self::Snapshot(error)
    }
}

pub fn run_import_facts(command: &ImportFactsCommand) -> Result<String, CliCommandError> {
    if command.output.try_exists()? {
        return Err(io::Error::new(
            io::ErrorKind::AlreadyExists,
            format!(
                "output directory already exists: {}",
                command.output.display()
            ),
        )
        .into());
    }
    let text = fs::read_to_string(&command.facts)?;
    let facts = RepositoryFacts::parse(&text)?;
    let compiled = arcana::repository::compile_repository_facts(&facts)?;
    fs::create_dir(&command.output)?;
    write_compiled(
        &command.output,
        &compiled,
        &facts,
        &command.adapter_name,
        &command.adapter_version,
    )
}

pub(crate) fn write_compiled(
    output: &Path,
    compiled: &CompiledRepository,
    facts: &RepositoryFacts,
    adapter_name: &str,
    adapter_version: &str,
) -> Result<String, CliCommandError> {
    let graph_path = output.join("graph.arcana");
    arcana::storage::write_packed(&graph_path, &compiled.dataset)?;
    publish_snapshot(
        output.join("graph.manifest"),
        "graph.arcana",
        None,
        timestamp()?,
    )?;
    write_repository_metadata(output, compiled, facts, adapter_name, adapter_version)?;
    let graph_size = fs::metadata(&graph_path)?.len();
    let metadata_size = [
        "catalogue.tsv",
        "unresolved.tsv",
        "facts.tsv",
        "graph.manifest",
        "repository.manifest",
    ]
    .iter()
    .map(|file| fs::metadata(output.join(file)).map(|metadata| metadata.len()))
    .collect::<Result<Vec<_>, _>>()?
    .into_iter()
    .sum::<u64>();
    Ok(format!(
        "imported facts: nodes={} edges={} unresolved={} graph.arcana={} bytes metadata={} bytes total={} bytes\n",
        compiled.dataset.node_count,
        compiled.dataset.edges.len(),
        compiled.unresolved.len(),
        graph_size,
        metadata_size,
        graph_size + metadata_size
    ))
}

pub(crate) fn write_repository_metadata(
    output: &Path,
    compiled: &CompiledRepository,
    facts: &RepositoryFacts,
    adapter_name: &str,
    adapter_version: &str,
) -> Result<(), CliCommandError> {
    arcana::repository::write_catalogue(output.join("catalogue.tsv"), &compiled.catalogue)?;
    let unresolved =
        RepositoryFacts::with_unresolved(Vec::new(), Vec::new(), compiled.unresolved.clone());
    fs::write(output.join("unresolved.tsv"), unresolved.encode())?;
    fs::write(output.join("facts.tsv"), facts.canonicalized().encode())?;
    publish_repository_snapshot(
        output.join("repository.manifest"),
        PublishRepositorySnapshot {
            graph_manifest_file: Path::new("graph.manifest"),
            catalogue_file: Path::new("catalogue.tsv"),
            unresolved_file: Path::new("unresolved.tsv"),
            facts_file: Path::new("facts.tsv"),
            adapter_name,
            adapter_version,
            created_unix_seconds: timestamp()?,
        },
    )?;
    Ok(())
}

pub(crate) fn timestamp() -> Result<u64, CliCommandError> {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|duration| duration.as_secs())
        .map_err(|error| io::Error::other(error).into())
}
