use std::fs::File;
use std::io::{BufRead, BufReader, Read};
use std::path::Path;

use super::Embedder;
use super::index::{
    IndexRecord, SearchHit, VectorIndexError, current_index_directory, current_snapshot_directory,
    manifest_matches, read_manifest, validate_files,
};
use crate::repository::{REPOSITORY_MANIFEST_FILE, RepositorySnapshot};

pub fn search_current_index(
    state: impl AsRef<Path>,
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
    let state = state.as_ref();
    let directory = current_index_directory(state, embedder.identity())?;
    let manifest = read_manifest(&directory)?;
    validate_files(&directory, &manifest)?;
    let (_, snapshot_directory) = current_snapshot_directory(state)?;
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
    if query_vector.len() != manifest.dimensions {
        return Err(VectorIndexError::CorruptIndex(format!(
            "query embedding has {} dimensions; index requires {}",
            query_vector.len(),
            manifest.dimensions
        )));
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
        vectors.read_exact(&mut buffer).map_err(|error| {
            VectorIndexError::CorruptIndex(format!("cannot read vector {index}: {error}"))
        })?;
        let score = buffer
            .chunks_exact(4)
            .zip(&query_vector)
            .map(|(bytes, query)| {
                f32::from_le_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]) * query
            })
            .sum();
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
    hits.truncate(limit);
    Ok(hits)
}
