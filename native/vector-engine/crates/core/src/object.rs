use std::{
    fs::{self, OpenOptions},
    io::Write,
    path::PathBuf,
    sync::atomic::{AtomicU64, Ordering},
};

use crate::{
    Error, Result,
    object_format::{address, decode, encode, hex, validate_identity, validate_vector},
};

static TEMP_SEQUENCE: AtomicU64 = AtomicU64::new(1);

pub struct ObjectStore {
    root: PathBuf,
}

pub(crate) struct ObjectData {
    pub hash: [u8; 32],
    pub vector: Vec<f32>,
}

impl ObjectStore {
    pub fn new(path: impl Into<PathBuf>) -> Self {
        Self { root: path.into() }
    }

    pub fn contains(&self, model: &str, source: &str) -> bool {
        self.object_path(&address(model, source)).is_file()
    }

    pub fn put(&self, model: &str, source: &str, vector: &[f32]) -> Result<[u8; 32]> {
        validate_identity(model, source)?;
        validate_vector(vector)?;
        let hash = address(model, source);
        let path = self.object_path(&hash);
        if path.is_file() {
            let existing = self.read(model, source)?;
            if existing.vector.as_slice() != vector {
                return Err(Error::InvalidInput(
                    "model/source address already contains different vector data".into(),
                ));
            }
            return Ok(hash);
        }

        let parent = path.parent().expect("object path has parent");
        fs::create_dir_all(parent)?;
        let temp = parent.join(format!(
            ".{}.tmp-{}-{}",
            hex(&hash),
            std::process::id(),
            TEMP_SEQUENCE.fetch_add(1, Ordering::Relaxed)
        ));
        let mut file = OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&temp)?;
        file.write_all(&encode(model, source, vector)?)?;
        file.sync_all()?;
        drop(file);
        match fs::rename(&temp, &path) {
            Ok(()) => Ok(hash),
            Err(error) if path.is_file() => {
                let _ = fs::remove_file(temp);
                if self.read(model, source)?.vector.as_slice() == vector {
                    Ok(hash)
                } else {
                    Err(error.into())
                }
            }
            Err(error) => {
                let _ = fs::remove_file(temp);
                Err(error.into())
            }
        }
    }

    pub(crate) fn read(&self, model: &str, source: &str) -> Result<ObjectData> {
        validate_identity(model, source)?;
        let hash = address(model, source);
        let path = self.object_path(&hash);
        if !path.is_file() {
            return Err(Error::MissingObject(source.into()));
        }
        let vector = decode(&fs::read(path)?, model, source)?;
        Ok(ObjectData { hash, vector })
    }

    fn object_path(&self, hash: &[u8; 32]) -> PathBuf {
        let encoded = hex(hash);
        self.root
            .join("objects")
            .join(&encoded[..2])
            .join(&encoded[2..])
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn reuses_identical_object_and_rejects_conflict() {
        let temp = tempfile::tempdir().unwrap();
        let store = ObjectStore::new(temp.path());
        let first = store.put("model", "source", &[0.5, -0.5]).unwrap();
        let second = store.put("model", "source", &[0.5, -0.5]).unwrap();
        assert_eq!(first, second);
        assert!(store.put("model", "source", &[1.0, 0.0]).is_err());
    }
}
