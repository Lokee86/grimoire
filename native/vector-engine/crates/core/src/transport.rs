use std::{
    fs::File,
    io::{BufRead, BufReader},
    path::Path,
};

use serde::Deserialize;

use crate::{ObjectStore, RecordRef, Result, SnapshotInfo, materialize};

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct IngestRecord {
    source: String,
    vector: Vec<f32>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct ManifestRecord {
    id: String,
    source: String,
}

pub fn ingest_jsonl(store: &ObjectStore, model: &str, path: impl AsRef<Path>) -> Result<usize> {
    let mut count = 0;
    for line in lines(path)? {
        let line = line?;
        if line.trim().is_empty() {
            continue;
        }
        let record: IngestRecord = serde_json::from_str(&line)?;
        store.put(model, &record.source, &record.vector)?;
        count += 1;
    }
    Ok(count)
}

pub fn materialize_jsonl(
    store: &ObjectStore,
    model: &str,
    manifest: impl AsRef<Path>,
    snapshot: impl AsRef<Path>,
) -> Result<SnapshotInfo> {
    let mut records = Vec::new();
    for line in lines(manifest)? {
        let line = line?;
        if line.trim().is_empty() {
            continue;
        }
        let record: ManifestRecord = serde_json::from_str(&line)?;
        records.push(RecordRef {
            id: record.id,
            source: record.source,
        });
    }
    materialize(store, model, &records, snapshot)
}

fn lines(path: impl AsRef<Path>) -> Result<impl Iterator<Item = std::io::Result<String>>> {
    Ok(BufReader::new(File::open(path)?).lines())
}
