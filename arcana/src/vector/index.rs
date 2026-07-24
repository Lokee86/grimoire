use std::fmt;
use std::fs;
use std::path::{Path, PathBuf};

use serde::{Deserialize, Serialize};

use super::{Embedder, EmbeddingError};
use crate::repository::RepositorySnapshotError;

pub(crate) const INDEX_VERSION: u64 = 1;
pub(crate) const MANIFEST_FILE: &str = "manifest.json";
pub(crate) const RECORDS_FILE: &str = "nodes.jsonl";
pub(crate) const VECTORS_FILE: &str = "vectors.f32";

#[derive(Clone, Debug, Eq, PartialEq, Serialize, Deserialize)]
pub struct IndexManifest {
    pub version: u64,
    pub repository_snapshot_id: String,
    pub graph_snapshot_id: String,
    pub model: String,
    pub identity: String,
    pub dimensions: usize,
    pub item_count: usize,
    pub records_file: String,
    pub vectors_file: String,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct SearchHit {
    pub score: f32,
    pub node_key: String,
    pub kind: String,
    pub path: String,
    pub name: String,
}

#[derive(Clone, Debug, Eq, PartialEq, Serialize, Deserialize)]
pub(crate) struct IndexRecord {
    pub node_key: String,
    pub kind: String,
    pub path: String,
    pub name: String,
}

pub fn current_index_directory(
    state: impl AsRef<Path>,
    identity: &str,
) -> Result<PathBuf, VectorIndexError> {
    validate_identity(identity)?;
    let (digest, _) = current_snapshot_directory(state.as_ref())?;
    Ok(state.as_ref().join("vectors").join(digest).join(identity))
}

pub(crate) fn current_snapshot_directory(
    state: &Path,
) -> Result<(String, PathBuf), VectorIndexError> {
    let current = fs::read_to_string(state.join("CURRENT"))?;
    let id = current.trim();
    let digest = id
        .strip_prefix("sha256:")
        .filter(|digest| {
            digest.len() == 64
                && digest
                    .bytes()
                    .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
        })
        .ok_or_else(|| {
            VectorIndexError::InvalidState(format!("invalid Arcana CURRENT value {id:?}"))
        })?;
    Ok((digest.to_owned(), state.join("snapshots").join(digest)))
}

pub(crate) fn read_manifest(directory: &Path) -> Result<IndexManifest, VectorIndexError> {
    let manifest: IndexManifest =
        serde_json::from_slice(&fs::read(directory.join(MANIFEST_FILE))?)?;
    if manifest.version != INDEX_VERSION {
        return Err(VectorIndexError::CorruptIndex(format!(
            "unsupported vector index version {}",
            manifest.version
        )));
    }
    validate_identity(&manifest.identity)?;
    Ok(manifest)
}

pub(crate) fn validate_files(
    directory: &Path,
    manifest: &IndexManifest,
) -> Result<(), VectorIndexError> {
    let expected = manifest
        .item_count
        .checked_mul(manifest.dimensions)
        .and_then(|values| values.checked_mul(4))
        .ok_or_else(|| VectorIndexError::CorruptIndex("vector file size overflow".to_owned()))?;
    let actual = fs::metadata(directory.join(&manifest.vectors_file))?.len();
    if actual != expected as u64 {
        return Err(VectorIndexError::CorruptIndex(format!(
            "vector file is {actual} bytes; expected {expected}"
        )));
    }
    if !directory.join(&manifest.records_file).is_file() {
        return Err(VectorIndexError::CorruptIndex(
            "vector node records are missing".to_owned(),
        ));
    }
    Ok(())
}

pub(crate) fn manifest_matches(
    manifest: &IndexManifest,
    repository_snapshot_id: u64,
    graph_snapshot_id: u64,
    embedder: &dyn Embedder,
) -> bool {
    manifest.repository_snapshot_id == format!("{repository_snapshot_id:016x}")
        && manifest.graph_snapshot_id == format!("{graph_snapshot_id:016x}")
        && manifest.model == embedder.model()
        && manifest.identity == embedder.identity()
        && manifest.dimensions == embedder.dimensions()
}

fn validate_identity(identity: &str) -> Result<(), VectorIndexError> {
    if identity.is_empty()
        || !identity.bytes().all(|byte| {
            byte.is_ascii_lowercase() || byte.is_ascii_digit() || matches!(byte, b'-' | b'_' | b'.')
        })
    {
        return Err(VectorIndexError::InvalidState(format!(
            "invalid vector index identity {identity:?}"
        )));
    }
    Ok(())
}

#[derive(Debug)]
pub enum VectorIndexError {
    Io(std::io::Error),
    Json(serde_json::Error),
    Repository(RepositorySnapshotError),
    Embedding(EmbeddingError),
    InvalidState(String),
    CorruptIndex(String),
}

impl fmt::Display for VectorIndexError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::Json(error) => error.fmt(formatter),
            Self::Repository(error) => error.fmt(formatter),
            Self::Embedding(error) => error.fmt(formatter),
            Self::InvalidState(message) | Self::CorruptIndex(message) => {
                formatter.write_str(message)
            }
        }
    }
}

impl std::error::Error for VectorIndexError {}
impl From<std::io::Error> for VectorIndexError {
    fn from(error: std::io::Error) -> Self {
        Self::Io(error)
    }
}
impl From<serde_json::Error> for VectorIndexError {
    fn from(error: serde_json::Error) -> Self {
        Self::Json(error)
    }
}
impl From<RepositorySnapshotError> for VectorIndexError {
    fn from(error: RepositorySnapshotError) -> Self {
        Self::Repository(error)
    }
}
impl From<EmbeddingError> for VectorIndexError {
    fn from(error: EmbeddingError) -> Self {
        Self::Embedding(error)
    }
}
