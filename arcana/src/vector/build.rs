use std::fs::{self, File};
use std::io::{BufWriter, Write};
use std::path::{Path, PathBuf};

use super::Embedder;
use super::documents::graph_documents;
use super::index::{
    INDEX_VERSION, IndexManifest, IndexRecord, MANIFEST_FILE, RECORDS_FILE, VECTORS_FILE,
    VectorIndexError, current_index_directory, current_snapshot_directory, manifest_matches,
    read_manifest, validate_files,
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
    let (_, snapshot_directory) = current_snapshot_directory(state)?;
    let snapshot = RepositorySnapshot::open(snapshot_directory.join(REPOSITORY_MANIFEST_FILE))?;
    let target = current_index_directory(state, embedder.identity())?;

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
    let parent = target.parent().ok_or_else(|| {
        VectorIndexError::InvalidState("vector index has no parent directory".to_owned())
    })?;
    fs::create_dir_all(parent)?;
    let temp = parent.join(format!(
        ".{}.tmp-{}",
        embedder.identity(),
        std::process::id()
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
    if target.try_exists()? {
        fs::remove_dir_all(&target)?;
    }
    fs::rename(&temp, &target)?;
    Ok(BuildSummary {
        directory: target,
        item_count: documents.len(),
        dimensions: embedder.dimensions(),
        mode: "built",
    })
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
        vectors_file: VECTORS_FILE.to_owned(),
    };
    let mut bytes = serde_json::to_vec_pretty(&manifest)?;
    bytes.push(b'\n');
    fs::write(directory.join(MANIFEST_FILE), bytes)?;
    validate_files(directory, &manifest)
}
