use std::collections::BTreeMap;
use std::fs::{self, File};
use std::io::{BufWriter, Write};
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};
use std::thread;
use std::time::{Duration, Instant};

use fs2::FileExt;

use super::cache::{
    ObjectKey, append_object_vector, cache_directory, object_key, persist_object, validate_object,
};
use super::documents::graph_documents;
use super::index::{
    INDEX_VERSION, IndexManifest, IndexRecord, MANIFEST_FILE, RECORDS_FILE, VECTORS_FILE,
    VectorIndexError, current_snapshot_directory, file_sha256, index_directory, manifest_matches,
    read_manifest, semantic_index_identity, validate_embedder, validate_files,
};
use super::{Embedder, SEMANTIC_ELIGIBILITY_POLICY_VERSION};
use crate::repository::{REPOSITORY_MANIFEST_FILE, RepositorySnapshot};

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct BuildOptions {
    pub batch_size: usize,
    pub batch_concurrency: usize,
}

impl Default for BuildOptions {
    fn default() -> Self {
        Self {
            batch_size: 32,
            batch_concurrency: 1,
        }
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct BuildSummary {
    pub directory: PathBuf,
    pub item_count: usize,
    pub unique_vectors: usize,
    pub dimensions: usize,
    pub embedded_vectors: usize,
    pub reused_vectors: usize,
    pub cached_snapshot: bool,
    pub request_count: usize,
    pub snapshot_bytes: u64,
    pub duration: Duration,
    pub mode: &'static str,
}

pub fn build_current_index(
    state: impl AsRef<Path>,
    embedder: &dyn Embedder,
    batch_size: usize,
) -> Result<BuildSummary, VectorIndexError> {
    build_current_index_with_options(
        state,
        embedder,
        BuildOptions {
            batch_size,
            batch_concurrency: 1,
        },
    )
}

pub fn build_current_index_with_options(
    state: impl AsRef<Path>,
    embedder: &dyn Embedder,
    options: BuildOptions,
) -> Result<BuildSummary, VectorIndexError> {
    let started = Instant::now();
    if options.batch_size == 0 || options.batch_concurrency == 0 {
        return Err(VectorIndexError::InvalidState(
            "vector batch size and batch concurrency must be greater than zero".to_owned(),
        ));
    }
    let state = state.as_ref();
    validate_embedder(embedder)?;
    let (digest, snapshot_directory) = current_snapshot_directory(state)?;
    let snapshot = RepositorySnapshot::open(snapshot_directory.join(REPOSITORY_MANIFEST_FILE))?;

    let cache = cache_directory(state, embedder.identity());
    fs::create_dir_all(&cache)?;
    let cache_lock = File::options()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(cache.join(".build.lock"))?;
    FileExt::lock_exclusive(&cache_lock)?;

    let target = index_directory(state, &digest, embedder.identity());
    let parent = target.parent().ok_or_else(|| {
        VectorIndexError::InvalidState("vector index has no parent directory".to_owned())
    })?;
    fs::create_dir_all(parent)?;
    let index_lock = File::options()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(parent.join(format!(".{}.lock", embedder.identity())))?;
    FileExt::lock_exclusive(&index_lock)?;

    if let Ok(manifest) = read_manifest(&target)
        && manifest_matches(
            &manifest,
            snapshot.manifest().snapshot_id,
            snapshot.manifest().graph_snapshot_id,
            embedder,
        )
        && validate_files(&target, &manifest).is_ok()
    {
        let unique_vectors = if manifest.unique_vectors == 0 {
            manifest.item_count
        } else {
            manifest.unique_vectors
        };
        return Ok(BuildSummary {
            directory: target.clone(),
            item_count: manifest.item_count,
            unique_vectors,
            dimensions: manifest.dimensions,
            embedded_vectors: 0,
            reused_vectors: unique_vectors,
            cached_snapshot: true,
            request_count: 0,
            snapshot_bytes: snapshot_bytes(&target)?,
            duration: started.elapsed(),
            mode: "existing",
        });
    }

    let documents = graph_documents(snapshot.facts());
    let document_keys = documents
        .iter()
        .map(|document| object_key(&document.text, embedder))
        .collect::<Vec<_>>();
    let mut unique = BTreeMap::<ObjectKey, &str>::new();
    for (document, key) in documents.iter().zip(&document_keys) {
        unique.entry(key.clone()).or_insert(&document.text);
    }
    let missing = unique
        .iter()
        .filter(|(key, _)| validate_object(state, embedder, key).is_err())
        .map(|(key, text)| MissingObject {
            key: key.clone(),
            text: (*text).to_owned(),
        })
        .collect::<Vec<_>>();
    let request_count = missing.len().div_ceil(options.batch_size);
    embed_missing(
        state,
        embedder,
        &missing,
        options.batch_size,
        options.batch_concurrency,
    )?;

    static TEMP_SEQUENCE: AtomicU64 = AtomicU64::new(0);
    let temp = parent.join(format!(
        ".{}.tmp-{}-{}",
        embedder.identity(),
        std::process::id(),
        TEMP_SEQUENCE.fetch_add(1, Ordering::Relaxed)
    ));
    if temp.try_exists()? {
        fs::remove_dir_all(&temp)?;
    }
    fs::create_dir(&temp)?;

    let result = write_index(
        &temp,
        state,
        &documents,
        &document_keys,
        unique.len(),
        &snapshot,
        embedder,
    );
    if let Err(error) = result {
        let _ = fs::remove_dir_all(&temp);
        return Err(error);
    }
    let bytes = snapshot_bytes(&temp)?;
    let (current_digest, _) = current_snapshot_directory(state)?;
    if current_digest != digest {
        fs::remove_dir_all(&temp)?;
        return Err(VectorIndexError::InvalidState(
            "Arcana CURRENT changed while the vector index was being built; retry vectorize"
                .to_owned(),
        ));
    }
    if let Err(error) = replace_index(&temp, &target) {
        let _ = fs::remove_dir_all(&temp);
        return Err(error);
    }
    let (published_digest, _) = current_snapshot_directory(state)?;
    if published_digest != digest {
        return Err(VectorIndexError::InvalidState(
            "Arcana CURRENT changed while the vector index was being published; retry vectorize"
                .to_owned(),
        ));
    }
    Ok(BuildSummary {
        directory: target,
        item_count: documents.len(),
        unique_vectors: unique.len(),
        dimensions: embedder.dimensions(),
        embedded_vectors: missing.len(),
        reused_vectors: unique.len() - missing.len(),
        cached_snapshot: false,
        request_count,
        snapshot_bytes: bytes,
        duration: started.elapsed(),
        mode: "built",
    })
}

#[derive(Clone, Debug)]
struct MissingObject {
    key: ObjectKey,
    text: String,
}

fn embed_missing(
    state: &Path,
    embedder: &dyn Embedder,
    missing: &[MissingObject],
    batch_size: usize,
    concurrency: usize,
) -> Result<(), VectorIndexError> {
    let batches = missing.chunks(batch_size).collect::<Vec<_>>();
    thread::scope(|scope| {
        for wave in batches.chunks(concurrency) {
            let handles = wave
                .iter()
                .map(|batch| scope.spawn(move || embed_batch(state, embedder, batch)))
                .collect::<Vec<_>>();
            let mut first_error = None;
            for handle in handles {
                match handle.join() {
                    Ok(Ok(())) => {}
                    Ok(Err(error)) if first_error.is_none() => first_error = Some(error),
                    Ok(Err(_)) => {}
                    Err(_) if first_error.is_none() => {
                        first_error = Some(VectorIndexError::InvalidState(
                            "graph-document embedding worker panicked".to_owned(),
                        ));
                    }
                    Err(_) => {}
                }
            }
            if let Some(error) = first_error {
                return Err(error);
            }
        }
        Ok(())
    })
}

fn embed_batch(
    state: &Path,
    embedder: &dyn Embedder,
    batch: &[MissingObject],
) -> Result<(), VectorIndexError> {
    let inputs = batch
        .iter()
        .map(|object| object.text.clone())
        .collect::<Vec<_>>();
    let embedded = embedder.embed_documents(&inputs)?;
    if embedded.len() != batch.len() {
        return Err(VectorIndexError::CorruptIndex(format!(
            "embedder returned {} vectors for {} graph documents",
            embedded.len(),
            batch.len()
        )));
    }
    for (object, vector) in batch.iter().zip(&embedded) {
        persist_object(state, embedder, &object.key, vector)?;
    }
    Ok(())
}

pub(super) fn replace_index(temp: &Path, target: &Path) -> Result<(), VectorIndexError> {
    if !target.try_exists()? {
        fs::rename(temp, target)?;
        return Ok(());
    }
    let backup = target.with_extension(format!("old-{}", std::process::id()));
    if backup.try_exists()? {
        fs::remove_dir_all(&backup)?;
    }
    fs::rename(target, &backup)?;
    if let Err(error) = fs::rename(temp, target) {
        let rollback = fs::rename(&backup, target);
        return match rollback {
            Ok(()) => Err(error.into()),
            Err(rollback) => Err(VectorIndexError::InvalidState(format!(
                "cannot publish vector index ({error}) and cannot restore the previous index ({rollback})"
            ))),
        };
    }
    fs::remove_dir_all(backup)?;
    Ok(())
}

fn write_index(
    directory: &Path,
    state: &Path,
    documents: &[super::GraphDocument],
    document_keys: &[ObjectKey],
    unique_vectors: usize,
    snapshot: &RepositorySnapshot,
    embedder: &dyn Embedder,
) -> Result<(), VectorIndexError> {
    let records_file = File::create(directory.join(RECORDS_FILE))?;
    let vectors_file = File::create(directory.join(VECTORS_FILE))?;
    let mut records = BufWriter::new(records_file);
    let mut vectors = BufWriter::new(vectors_file);

    for (document, key) in documents.iter().zip(document_keys) {
        let record = IndexRecord {
            node_key: format!("{:016x}", document.node_key),
            kind: document.kind.clone(),
            path: document.path.clone(),
            name: document.name.clone(),
        };
        serde_json::to_writer(&mut records, &record)?;
        records.write_all(b"\n")?;
        append_object_vector(&mut vectors, state, embedder, key)?;
    }
    records.flush()?;
    vectors.flush()?;
    records.get_ref().sync_all()?;
    vectors.get_ref().sync_all()?;

    let manifest = IndexManifest {
        version: INDEX_VERSION,
        repository_snapshot_id: format!("{:016x}", snapshot.manifest().snapshot_id),
        graph_snapshot_id: format!("{:016x}", snapshot.manifest().graph_snapshot_id),
        model: embedder.model().to_owned(),
        identity: semantic_index_identity(embedder.identity()),
        eligibility_policy_version: SEMANTIC_ELIGIBILITY_POLICY_VERSION,
        dimensions: embedder.dimensions(),
        item_count: documents.len(),
        unique_vectors,
        records_file: RECORDS_FILE.to_owned(),
        records_sha256: file_sha256(&directory.join(RECORDS_FILE))?,
        vectors_file: VECTORS_FILE.to_owned(),
        vectors_sha256: file_sha256(&directory.join(VECTORS_FILE))?,
    };
    let mut bytes = serde_json::to_vec_pretty(&manifest)?;
    bytes.push(b'\n');
    let mut file = File::create(directory.join(MANIFEST_FILE))?;
    file.write_all(&bytes)?;
    file.sync_all()?;
    validate_files(directory, &manifest)
}

fn snapshot_bytes(directory: &Path) -> Result<u64, VectorIndexError> {
    [MANIFEST_FILE, RECORDS_FILE, VECTORS_FILE]
        .into_iter()
        .try_fold(0_u64, |total, file| {
            fs::metadata(directory.join(file))?
                .len()
                .checked_add(total)
                .ok_or_else(|| {
                    VectorIndexError::InvalidState("vector snapshot byte size overflow".to_owned())
                })
        })
}
