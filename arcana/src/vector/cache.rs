use std::fs::{self, File};
use std::io::{Read, Write};
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};

use sha2::{Digest, Sha256};

use super::{Embedder, VectorIndexError};

const OBJECT_MAGIC: &[u8; 8] = b"ARCAVEC1";
const OBJECT_HEADER_BYTES: usize = 8 + 8 + 32 + 32;

#[derive(Clone, Debug, Eq, Hash, Ord, PartialEq, PartialOrd)]
pub(super) struct ObjectKey {
    bytes: [u8; 32],
    hex: String,
}

pub(super) fn cache_directory(state: &Path, identity: &str) -> PathBuf {
    state.join("vector-cache").join(identity)
}

pub(super) fn object_key(document: &str, embedder: &dyn Embedder) -> ObjectKey {
    let mut hasher = Sha256::new();
    hasher.update(b"arcana-graph-document-vector-v1\0");
    hash_field(&mut hasher, embedder.identity().as_bytes());
    hash_field(&mut hasher, embedder.model().as_bytes());
    hasher.update((embedder.dimensions() as u64).to_le_bytes());
    hash_field(&mut hasher, document.as_bytes());
    let digest = hasher.finalize();
    let hex = format!("{digest:x}");
    let bytes: [u8; 32] = digest.into();
    ObjectKey { hex, bytes }
}

fn hash_field(hasher: &mut Sha256, value: &[u8]) {
    hasher.update((value.len() as u64).to_le_bytes());
    hasher.update(value);
}

pub(super) fn object_path(state: &Path, identity: &str, key: &ObjectKey) -> PathBuf {
    cache_directory(state, identity)
        .join("objects")
        .join(&key.hex[..2])
        .join(format!("{}.avec", key.hex))
}

pub(super) fn validate_object(
    state: &Path,
    embedder: &dyn Embedder,
    key: &ObjectKey,
) -> Result<(), VectorIndexError> {
    read_vector_bytes(
        &object_path(state, embedder.identity(), key),
        key,
        embedder.dimensions(),
    )
    .map(|_| ())
}

pub(super) fn append_object_vector(
    destination: &mut impl Write,
    state: &Path,
    embedder: &dyn Embedder,
    key: &ObjectKey,
) -> Result<(), VectorIndexError> {
    let vector = read_vector_bytes(
        &object_path(state, embedder.identity(), key),
        key,
        embedder.dimensions(),
    )?;
    destination.write_all(&vector)?;
    Ok(())
}

pub(super) fn persist_object(
    state: &Path,
    embedder: &dyn Embedder,
    key: &ObjectKey,
    vector: &[f32],
) -> Result<(), VectorIndexError> {
    if vector.len() != embedder.dimensions() {
        return Err(VectorIndexError::CorruptIndex(format!(
            "embedder returned {} dimensions; expected {}",
            vector.len(),
            embedder.dimensions()
        )));
    }
    if vector.iter().any(|value| !value.is_finite()) {
        return Err(VectorIndexError::CorruptIndex(
            "embedder returned a non-finite graph-document vector".to_owned(),
        ));
    }

    let path = object_path(state, embedder.identity(), key);
    if read_vector_bytes(&path, key, embedder.dimensions()).is_ok() {
        return Ok(());
    }
    let parent = path.parent().ok_or_else(|| {
        VectorIndexError::InvalidState("vector cache object has no parent directory".to_owned())
    })?;
    fs::create_dir_all(parent)?;

    let vector_bytes = vector
        .iter()
        .flat_map(|value| value.to_le_bytes())
        .collect::<Vec<_>>();
    let vector_sha256: [u8; 32] = Sha256::digest(&vector_bytes).into();
    let mut object = Vec::with_capacity(OBJECT_HEADER_BYTES + vector_bytes.len());
    object.extend_from_slice(OBJECT_MAGIC);
    object.extend_from_slice(&(embedder.dimensions() as u64).to_le_bytes());
    object.extend_from_slice(&key.bytes);
    object.extend_from_slice(&vector_sha256);
    object.extend_from_slice(&vector_bytes);

    static TEMP_SEQUENCE: AtomicU64 = AtomicU64::new(0);
    let temp = parent.join(format!(
        ".{}.tmp-{}-{}",
        key.hex,
        std::process::id(),
        TEMP_SEQUENCE.fetch_add(1, Ordering::Relaxed)
    ));
    let mut file = File::create(&temp)?;
    file.write_all(&object)?;
    file.sync_all()?;
    drop(file);

    if let Err(error) = replace_object(&temp, &path) {
        let _ = fs::remove_file(&temp);
        return Err(error);
    }
    read_vector_bytes(&path, key, embedder.dimensions()).map(|_| ())
}

fn replace_object(temp: &Path, target: &Path) -> Result<(), VectorIndexError> {
    if !target.try_exists()? {
        fs::rename(temp, target)?;
        return Ok(());
    }
    let backup = target.with_extension(format!("corrupt-{}", std::process::id()));
    if backup.try_exists()? {
        fs::remove_file(&backup)?;
    }
    fs::rename(target, &backup)?;
    if let Err(error) = fs::rename(temp, target) {
        let rollback = fs::rename(&backup, target);
        return match rollback {
            Ok(()) => Err(error.into()),
            Err(rollback) => Err(VectorIndexError::InvalidState(format!(
                "cannot repair vector cache object ({error}) and cannot restore the prior object ({rollback})"
            ))),
        };
    }
    fs::remove_file(backup)?;
    Ok(())
}

fn read_vector_bytes(
    path: &Path,
    key: &ObjectKey,
    dimensions: usize,
) -> Result<Vec<u8>, VectorIndexError> {
    let vector_bytes = dimensions.checked_mul(4).ok_or_else(|| {
        VectorIndexError::CorruptIndex("vector cache object size overflow".to_owned())
    })?;
    let expected = OBJECT_HEADER_BYTES
        .checked_add(vector_bytes)
        .ok_or_else(|| {
            VectorIndexError::CorruptIndex("vector cache object size overflow".to_owned())
        })?;
    let mut object = Vec::with_capacity(expected);
    File::open(path)?.read_to_end(&mut object)?;
    if object.len() != expected {
        return Err(VectorIndexError::CorruptIndex(format!(
            "vector cache object is {} bytes; expected {expected}",
            object.len()
        )));
    }
    if &object[..8] != OBJECT_MAGIC {
        return Err(VectorIndexError::CorruptIndex(
            "vector cache object has invalid magic".to_owned(),
        ));
    }
    let stored_dimensions = u64::from_le_bytes(object[8..16].try_into().unwrap());
    if stored_dimensions != dimensions as u64 {
        return Err(VectorIndexError::CorruptIndex(
            "vector cache object has the wrong dimensions".to_owned(),
        ));
    }
    if object[16..48] != key.bytes {
        return Err(VectorIndexError::CorruptIndex(
            "vector cache object content key does not match its path".to_owned(),
        ));
    }
    let vector = &object[OBJECT_HEADER_BYTES..];
    let actual_sha256: [u8; 32] = Sha256::digest(vector).into();
    if object[48..80] != actual_sha256 {
        return Err(VectorIndexError::CorruptIndex(
            "vector cache object checksum does not match its header".to_owned(),
        ));
    }
    for (index, bytes) in vector.chunks_exact(4).enumerate() {
        if !f32::from_le_bytes(bytes.try_into().unwrap()).is_finite() {
            return Err(VectorIndexError::CorruptIndex(format!(
                "vector cache object contains a non-finite value at position {index}"
            )));
        }
    }
    Ok(vector.to_vec())
}
