use std::{
    fs,
    path::{Path, PathBuf},
    sync::atomic::{AtomicU64, Ordering},
};

use crate::{
    Error, ObjectStore, RecordRef, Result, Snapshot, SnapshotInfo,
    materialize_format::{layout, write_snapshot},
};

static TEMP_SEQUENCE: AtomicU64 = AtomicU64::new(1);

pub fn materialize(
    store: &ObjectStore,
    model: &str,
    records: &[RecordRef],
    path: impl AsRef<Path>,
) -> Result<SnapshotInfo> {
    let records = validate_records(model, records)?;
    let dimensions = store.read(model, &records[0].source)?.vector.len();
    let layout = layout(model, &records, dimensions)?;
    let target = path.as_ref();
    fs::create_dir_all(target.parent().unwrap_or_else(|| Path::new(".")))?;
    let temp = temporary_path(target);
    if let Err(error) = write_snapshot(&temp, store, model, &records, dimensions, layout) {
        let _ = fs::remove_file(&temp);
        return Err(error);
    }
    publish(&temp, target)?;
    Snapshot::open(target).map(|snapshot| snapshot.info().clone())
}

fn validate_records(model: &str, records: &[RecordRef]) -> Result<Vec<RecordRef>> {
    if model.trim().is_empty() || records.is_empty() {
        return Err(Error::InvalidInput("model and records are required".into()));
    }
    let mut records = records.to_vec();
    records.sort_by(|left, right| left.id.cmp(&right.id));
    for record in &records {
        if record.id.is_empty() || record.source.is_empty() {
            return Err(Error::InvalidInput(
                "record id and source are required".into(),
            ));
        }
        u32::try_from(record.id.len())
            .map_err(|_| Error::InvalidInput("record id is too long".into()))?;
    }
    if records.windows(2).any(|pair| pair[0].id == pair[1].id) {
        return Err(Error::InvalidInput("record ids must be unique".into()));
    }
    Ok(records)
}

fn temporary_path(target: &Path) -> PathBuf {
    let mut path = target.as_os_str().to_owned();
    path.push(format!(
        ".tmp-{}-{}",
        std::process::id(),
        TEMP_SEQUENCE.fetch_add(1, Ordering::Relaxed)
    ));
    PathBuf::from(path)
}

fn publish(temp: &Path, target: &Path) -> Result<()> {
    match fs::rename(temp, target) {
        Ok(()) => Ok(()),
        Err(_) if target.exists() => {
            fs::remove_file(target)?;
            fs::rename(temp, target)?;
            Ok(())
        }
        Err(error) => Err(error.into()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn snapshot_is_deterministic_and_sorted() {
        let temp = tempfile::tempdir().unwrap();
        let store = ObjectStore::new(temp.path().join("store"));
        store.put("m", "a", &[1.0, 0.0]).unwrap();
        store.put("m", "b", &[0.0, 1.0]).unwrap();
        let records = vec![
            RecordRef {
                id: "b".into(),
                source: "b".into(),
            },
            RecordRef {
                id: "a".into(),
                source: "a".into(),
            },
        ];
        let first = materialize(&store, "m", &records, temp.path().join("first.gvs")).unwrap();
        let second = materialize(&store, "m", &records, temp.path().join("second.gvs")).unwrap();
        assert_eq!(first.identity, second.identity);
        let snapshot = Snapshot::open(temp.path().join("first.gvs")).unwrap();
        assert_eq!(snapshot.id(0), "a");
        assert_eq!(snapshot.info().dimensions, 2);
    }
}
