use std::fs::File;
use std::io::{BufRead, BufReader, Read};
use std::path::Path;

use fs2::FileExt;

use super::Embedder;
use super::index::{
    IndexRecord, SearchHit, VectorIndexError, current_snapshot_directory, index_directory,
    manifest_matches, read_manifest, validate_embedder, validate_open_files, validate_record,
};
use crate::repository::{REPOSITORY_MANIFEST_FILE, RepositorySnapshot};

pub fn search_current_index(
    state: impl AsRef<Path>,
    embedder: &dyn Embedder,
    query: &str,
    limit: usize,
) -> Result<Vec<SearchHit>, VectorIndexError> {
    search_index(state.as_ref(), None, embedder, query, limit)
}

pub fn search_expected_index(
    state: impl AsRef<Path>,
    expected_snapshot: &str,
    embedder: &dyn Embedder,
    query: &str,
    limit: usize,
) -> Result<Vec<SearchHit>, VectorIndexError> {
    search_index(
        state.as_ref(),
        Some(expected_snapshot),
        embedder,
        query,
        limit,
    )
}

fn search_index(
    state: &Path,
    expected_snapshot: Option<&str>,
    embedder: &dyn Embedder,
    query: &str,
    limit: usize,
) -> Result<Vec<SearchHit>, VectorIndexError> {
    if query.trim().is_empty() {
        return Err(VectorIndexError::InvalidState(
            "semantic query is empty".to_owned(),
        ));
    }
    if limit == 0 {
        return Err(VectorIndexError::InvalidState(
            "semantic query limit must be greater than zero".to_owned(),
        ));
    }
    validate_embedder(embedder)?;
    let (digest, snapshot_directory) = current_snapshot_directory(state)?;
    if let Some(expected) = expected_snapshot
        && expected != format!("sha256:{digest}")
    {
        return Err(VectorIndexError::InvalidState(format!(
            "Arcana CURRENT does not match expected snapshot {expected}"
        )));
    }
    let directory = index_directory(state, &digest, embedder.identity());
    let parent = directory.parent().ok_or_else(|| {
        VectorIndexError::InvalidState("vector index has no parent directory".to_owned())
    })?;
    let lock = File::options()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(parent.join(format!(".{}.lock", embedder.identity())))?;
    FileExt::lock_shared(&lock)?;
    let manifest = read_manifest(&directory)?;
    // Full checksums and whole-file finite-value scans are build/status work. Query open
    // validates fixed metadata, then validates records and values in the scoring pass.
    validate_open_files(&directory, &manifest)?;
    let snapshot = RepositorySnapshot::open(snapshot_directory.join(REPOSITORY_MANIFEST_FILE))?;
    if !manifest_matches(
        &manifest,
        snapshot.manifest().snapshot_id,
        snapshot.manifest().graph_snapshot_id,
        embedder,
    ) {
        return Err(VectorIndexError::InvalidState(
            "Arcana vector index is stale for the current graph or embedding model".to_owned(),
        ));
    }
    let query_vector = embedder.embed_query(query)?;
    let (current_digest, _) = current_snapshot_directory(state)?;
    if current_digest != digest {
        return Err(VectorIndexError::InvalidState(
            "Arcana CURRENT changed during semantic query".to_owned(),
        ));
    }
    if query_vector.len() != manifest.dimensions {
        return Err(VectorIndexError::CorruptIndex(format!(
            "query embedding has {} dimensions; index requires {}",
            query_vector.len(),
            manifest.dimensions
        )));
    }
    if query_vector.iter().any(|value| !value.is_finite()) {
        return Err(VectorIndexError::CorruptIndex(
            "query embedding contains a non-finite value".to_owned(),
        ));
    }

    let records = BufReader::new(File::open(directory.join(&manifest.records_file))?);
    let mut vectors = BufReader::new(File::open(directory.join(&manifest.vectors_file))?);
    let mut buffer = vec![0_u8; manifest.dimensions * 4];
    let mut hits = Vec::with_capacity(manifest.item_count);

    for (index, line) in records.lines().enumerate() {
        if index >= manifest.item_count {
            return Err(VectorIndexError::CorruptIndex(
                "vector index has extra node records".to_owned(),
            ));
        }
        let record: IndexRecord = serde_json::from_str(&line?)?;
        validate_record(&record, index)?;
        vectors.read_exact(&mut buffer).map_err(|error| {
            VectorIndexError::CorruptIndex(format!("cannot read vector {index}: {error}"))
        })?;
        let mut score = 0.0_f32;
        for (position, (bytes, query)) in buffer.chunks_exact(4).zip(&query_vector).enumerate() {
            let value = f32::from_le_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]);
            if !value.is_finite() {
                return Err(VectorIndexError::CorruptIndex(format!(
                    "vector index contains a non-finite value at vector {index}, position {position}"
                )));
            }
            score += value * query;
        }
        hits.push(SearchHit {
            score,
            node_key: record.node_key,
            kind: record.kind,
            path: record.path,
            name: record.name,
        });
    }
    if hits.len() != manifest.item_count {
        return Err(VectorIndexError::CorruptIndex(format!(
            "vector index has {} node records; expected {}",
            hits.len(),
            manifest.item_count
        )));
    }
    let mut extra = [0_u8; 1];
    if vectors.read(&mut extra)? != 0 {
        return Err(VectorIndexError::CorruptIndex(
            "vector index has trailing vector bytes".to_owned(),
        ));
    }

    hits.sort_unstable_by(|left, right| {
        right
            .score
            .total_cmp(&left.score)
            .then_with(|| left.node_key.cmp(&right.node_key))
    });
    let (current_digest, _) = current_snapshot_directory(state)?;
    if current_digest != digest {
        return Err(VectorIndexError::InvalidState(
            "Arcana CURRENT changed during semantic query".to_owned(),
        ));
    }
    hits.truncate(limit);
    Ok(hits)
}
