use std::fmt;
use std::fs;
use std::fs::File;
use std::io::{BufRead, BufReader, Read};
use std::path::{Path, PathBuf};

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

use super::{Embedder, EmbeddingError, SEMANTIC_ELIGIBILITY_POLICY_VERSION};
use crate::repository::RepositorySnapshotError;

pub(crate) const INDEX_VERSION: u64 = 3;
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
    #[serde(default)]
    pub eligibility_policy_version: u64,
    pub dimensions: usize,
    pub item_count: usize,
    #[serde(default)]
    pub unique_vectors: usize,
    pub records_file: String,
    pub records_sha256: String,
    pub vectors_file: String,
    pub vectors_sha256: String,
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
    Ok(index_directory(state.as_ref(), &digest, identity))
}

pub(crate) fn index_directory(state: &Path, digest: &str, identity: &str) -> PathBuf {
    state.join("vectors").join(digest).join(identity)
}

pub(crate) fn validate_embedder(embedder: &dyn Embedder) -> Result<(), VectorIndexError> {
    validate_identity(embedder.identity())?;
    if embedder.model().trim().is_empty() {
        return Err(VectorIndexError::InvalidState(
            "embedding model is empty".to_owned(),
        ));
    }
    if embedder.dimensions() == 0 {
        return Err(VectorIndexError::InvalidState(
            "embedding dimensions must be greater than zero".to_owned(),
        ));
    }
    Ok(())
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
    if manifest.eligibility_policy_version != SEMANTIC_ELIGIBILITY_POLICY_VERSION {
        return Err(VectorIndexError::CorruptIndex(format!(
            "unsupported semantic eligibility policy version {}",
            manifest.eligibility_policy_version
        )));
    }
    validate_identity(&manifest.identity)?;
    if manifest.model.trim().is_empty() {
        return Err(VectorIndexError::CorruptIndex(
            "vector index model is empty".to_owned(),
        ));
    }
    if manifest.dimensions == 0 {
        return Err(VectorIndexError::CorruptIndex(
            "vector index dimensions must be greater than zero".to_owned(),
        ));
    }
    if manifest.unique_vectors > manifest.item_count {
        return Err(VectorIndexError::CorruptIndex(
            "vector index unique-vector count exceeds its item count".to_owned(),
        ));
    }
    if manifest.records_file != RECORDS_FILE || manifest.vectors_file != VECTORS_FILE {
        return Err(VectorIndexError::CorruptIndex(
            "vector index uses unsupported data filenames".to_owned(),
        ));
    }
    if !valid_sha256(&manifest.records_sha256) || !valid_sha256(&manifest.vectors_sha256) {
        return Err(VectorIndexError::CorruptIndex(
            "vector index has invalid data checksums".to_owned(),
        ));
    }
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
    let mut vectors = BufReader::new(File::open(directory.join(&manifest.vectors_file))?);
    let mut bytes = [0_u8; 4];
    for index in 0..expected / 4 {
        vectors.read_exact(&mut bytes)?;
        if !f32::from_le_bytes(bytes).is_finite() {
            return Err(VectorIndexError::CorruptIndex(format!(
                "vector index contains a non-finite value at position {index}"
            )));
        }
    }
    let records_path = directory.join(&manifest.records_file);
    if !records_path.is_file() {
        return Err(VectorIndexError::CorruptIndex(
            "vector node records are missing".to_owned(),
        ));
    }
    let records = BufReader::new(File::open(records_path)?);
    let mut count = 0_usize;
    for line in records.lines() {
        let record: IndexRecord = serde_json::from_str(&line?).map_err(|error| {
            VectorIndexError::CorruptIndex(format!("invalid vector node record {count}: {error}"))
        })?;
        validate_record(&record, count)?;
        count += 1;
    }
    if count != manifest.item_count {
        return Err(VectorIndexError::CorruptIndex(format!(
            "vector index has {count} node records; expected {}",
            manifest.item_count
        )));
    }
    for (file, expected, label) in [
        (
            &manifest.records_file,
            &manifest.records_sha256,
            "node records",
        ),
        (&manifest.vectors_file, &manifest.vectors_sha256, "vectors"),
    ] {
        let actual = file_sha256(&directory.join(file))?;
        if &actual != expected {
            return Err(VectorIndexError::CorruptIndex(format!(
                "vector index {label} checksum does not match its manifest"
            )));
        }
    }
    Ok(())
}

pub(crate) fn validate_open_files(
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

pub(crate) fn file_sha256(path: &Path) -> Result<String, VectorIndexError> {
    let mut file = File::open(path)?;
    let mut hasher = Sha256::new();
    let mut buffer = [0_u8; 64 * 1024];
    loop {
        let count = file.read(&mut buffer)?;
        if count == 0 {
            break;
        }
        hasher.update(&buffer[..count]);
    }
    Ok(format!("{:x}", hasher.finalize()))
}

fn valid_sha256(value: &str) -> bool {
    value.len() == 64
        && value
            .bytes()
            .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
}

pub(crate) fn validate_record(record: &IndexRecord, index: usize) -> Result<(), VectorIndexError> {
    if record.node_key.len() != 16
        || !record
            .node_key
            .bytes()
            .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
        || record.kind.trim().is_empty()
        || record.name.trim().is_empty()
    {
        return Err(VectorIndexError::CorruptIndex(format!(
            "invalid vector node record {index}"
        )));
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
        && manifest.identity == semantic_index_identity(embedder.identity())
        && manifest.eligibility_policy_version == SEMANTIC_ELIGIBILITY_POLICY_VERSION
        && manifest.dimensions == embedder.dimensions()
}

pub(crate) fn semantic_index_identity(embedding_identity: &str) -> String {
    format!("{embedding_identity}-arcana-semantic-v{SEMANTIC_ELIGIBILITY_POLICY_VERSION}")
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
