use std::fs::{self, File};
use std::io::{BufWriter, Write};
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};

use fs2::FileExt;

use super::Embedder;
use super::documents::graph_documents;
use super::index::{
    INDEX_VERSION, IndexManifest, IndexRecord, MANIFEST_FILE, RECORDS_FILE, VECTORS_FILE,
    VectorIndexError, current_snapshot_directory, file_sha256, index_directory, manifest_matches,
    read_manifest, validate_embedder, validate_files,
};
use crate::repository::{REPOSITORY_MANIFEST_FILE, RepositorySnapshot};

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct BuildSummary {
    pub directory: PathBuf,
    pub item_count: usize,
    pub dimensions: usize,
    pub mode: &'static str,
}

pub fn build_current_index(
    state: impl AsRef<Path>,
    embedder: &dyn Embedder,
    batch_size: usize,
) -> Result<BuildSummary, VectorIndexError> {
    if batch_size == 0 {
        return Err(VectorIndexError::InvalidState(
            "vector batch size must be greater than zero".to_owned(),
        ));
    }
    let state = state.as_ref();
    validate_embedder(embedder)?;
    let (digest, snapshot_directory) = current_snapshot_directory(state)?;
    let snapshot = RepositorySnapshot::open(snapshot_directory.join(REPOSITORY_MANIFEST_FILE))?;
    let target = index_directory(state, &digest, embedder.identity());
    let parent = target.parent().ok_or_else(|| {
        VectorIndexError::InvalidState("vector index has no parent directory".to_owned())
    })?;
    fs::create_dir_all(parent)?;
    let lock = File::options()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(parent.join(format!(".{}.lock", embedder.identity())))?;
    FileExt::lock_exclusive(&lock)?;

    if let Ok(manifest) = read_manifest(&target)
        && manifest_matches(
            &manifest,
            snapshot.manifest().snapshot_id,
            snapshot.manifest().graph_snapshot_id,
            embedder,
        )
        && validate_files(&target, &manifest).is_ok()
    {
        return Ok(BuildSummary {
            directory: target,
            item_count: manifest.item_count,
            dimensions: manifest.dimensions,
            mode: "existing",
        });
    }

    let documents = graph_documents(snapshot.facts());
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

    let result = write_index(&temp, &documents, &snapshot, embedder, batch_size);
    if let Err(error) = result {
        let _ = fs::remove_dir_all(&temp);
        return Err(error);
    }
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
        dimensions: embedder.dimensions(),
        mode: "built",
    })
}

fn replace_index(temp: &Path, target: &Path) -> Result<(), VectorIndexError> {
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
    documents: &[super::GraphDocument],
    snapshot: &RepositorySnapshot,
    embedder: &dyn Embedder,
    batch_size: usize,
) -> Result<(), VectorIndexError> {
    let records_file = File::create(directory.join(RECORDS_FILE))?;
    let vectors_file = File::create(directory.join(VECTORS_FILE))?;
    let mut records = BufWriter::new(records_file);
    let mut vectors = BufWriter::new(vectors_file);

    for batch in documents.chunks(batch_size) {
        let inputs = batch
            .iter()
            .map(|document| document.text.clone())
            .collect::<Vec<_>>();
        let embedded = embedder.embed_documents(&inputs)?;
        if embedded.len() != batch.len() {
            return Err(VectorIndexError::CorruptIndex(format!(
                "embedder returned {} vectors for {} graph documents",
                embedded.len(),
                batch.len()
            )));
        }
        for (document, vector) in batch.iter().zip(embedded) {
            if vector.len() != embedder.dimensions() {
                return Err(VectorIndexError::CorruptIndex(format!(
                    "embedder returned {} dimensions; expected {}",
                    vector.len(),
                    embedder.dimensions()
                )));
            }
            let record = IndexRecord {
                node_key: format!("{:016x}", document.node_key),
                kind: document.kind.clone(),
                path: document.path.clone(),
                name: document.name.clone(),
            };
            serde_json::to_writer(&mut records, &record)?;
            records.write_all(b"\n")?;
            for value in vector {
                vectors.write_all(&value.to_le_bytes())?;
            }
        }
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
        identity: embedder.identity().to_owned(),
        dimensions: embedder.dimensions(),
        item_count: documents.len(),
        records_file: RECORDS_FILE.to_owned(),
        records_sha256: file_sha256(&directory.join(RECORDS_FILE))?,
        vectors_file: VECTORS_FILE.to_owned(),
        vectors_sha256: file_sha256(&directory.join(VECTORS_FILE))?,
    };
    let mut bytes = serde_json::to_vec_pretty(&manifest)?;
    bytes.push(b'\n');
    fs::write(directory.join(MANIFEST_FILE), bytes)?;
    validate_files(directory, &manifest)
}
