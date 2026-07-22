use std::{collections::HashSet, fs::File, path::Path};

use bytemuck::try_cast_slice;
use memmap2::{Mmap, MmapOptions};

use crate::{
    Error, Result,
    format::{ENTRY_SIZE, HEADER_SIZE, VECTOR_ALIGNMENT, decode_header, get_u32, get_u64},
};

#[derive(Clone, Debug)]
pub struct RecordRef {
    pub id: String,
    pub source: String,
}

#[derive(Clone, Debug)]
pub struct SnapshotInfo {
    pub model: String,
    pub dimensions: usize,
    pub count: usize,
    pub identity: String,
}

pub struct Snapshot {
    mmap: Mmap,
    info: SnapshotInfo,
    ids: Vec<(usize, usize)>,
    vectors_offset: usize,
}

impl Snapshot {
    pub fn open(path: impl AsRef<Path>) -> Result<Self> {
        let file = File::open(path)?;
        // SAFETY: snapshots are immutable after publication. Rebuilds write a new file
        // and publish it only after completion, so mapped bytes are not mutated in place.
        let mmap = unsafe { MmapOptions::new().map(&file)? };
        let header = decode_header(&mmap)?;
        let count = usize::try_from(header.count)
            .map_err(|_| Error::InvalidFormat("record count is too large".into()))?;
        let dimensions = header.dimensions as usize;
        if count == 0 || dimensions == 0 {
            return Err(Error::InvalidFormat("snapshot is empty".into()));
        }

        let model_end = HEADER_SIZE
            .checked_add(header.model_len as usize)
            .ok_or_else(|| Error::InvalidFormat("model length overflow".into()))?;
        let entries_offset = header.entries_offset as usize;
        let ids_offset = header.ids_offset as usize;
        let vectors_offset = header.vectors_offset as usize;
        let entries_end = entries_offset
            .checked_add(
                count
                    .checked_mul(ENTRY_SIZE)
                    .ok_or_else(|| Error::InvalidFormat("entry table overflow".into()))?,
            )
            .ok_or_else(|| Error::InvalidFormat("entry table overflow".into()))?;
        let vector_bytes = count
            .checked_mul(dimensions)
            .and_then(|value| value.checked_mul(4))
            .ok_or_else(|| Error::InvalidFormat("vector matrix overflow".into()))?;
        if model_end > entries_offset
            || entries_end > ids_offset
            || ids_offset > vectors_offset
            || !vectors_offset.is_multiple_of(VECTOR_ALIGNMENT)
            || vectors_offset.checked_add(vector_bytes) != Some(mmap.len())
        {
            return Err(Error::InvalidFormat(
                "snapshot sections overlap or exceed file bounds".into(),
            ));
        }

        let model = std::str::from_utf8(&mmap[HEADER_SIZE..model_end])
            .map_err(|_| Error::InvalidFormat("model identity is not UTF-8".into()))?
            .to_owned();
        if model.is_empty() {
            return Err(Error::InvalidFormat("model identity is empty".into()));
        }
        let ids = validate_ids(&mmap, count, entries_offset, ids_offset, vectors_offset)?;
        let vectors = try_cast_slice::<u8, f32>(&mmap[vectors_offset..])
            .map_err(|_| Error::InvalidFormat("vector matrix is not aligned".into()))?;
        if vectors.iter().any(|value| !value.is_finite()) {
            return Err(Error::InvalidFormat(
                "vector matrix contains non-finite data".into(),
            ));
        }
        let identity = blake3::hash(&mmap).to_hex().to_string();
        Ok(Self {
            mmap,
            info: SnapshotInfo {
                model,
                dimensions,
                count,
                identity,
            },
            ids,
            vectors_offset,
        })
    }

    pub fn info(&self) -> &SnapshotInfo {
        &self.info
    }

    pub fn id(&self, index: usize) -> &str {
        let (start, length) = self.ids[index];
        std::str::from_utf8(&self.mmap[start..start + length]).expect("validated snapshot id")
    }

    pub(crate) fn vectors(&self) -> &[f32] {
        try_cast_slice(&self.mmap[self.vectors_offset..]).expect("validated vector matrix")
    }
}

fn validate_ids(
    bytes: &[u8],
    count: usize,
    entries_offset: usize,
    ids_offset: usize,
    vectors_offset: usize,
) -> Result<Vec<(usize, usize)>> {
    let mut ids = Vec::with_capacity(count);
    let mut seen = HashSet::with_capacity(count);
    for index in 0..count {
        let entry = entries_offset + index * ENTRY_SIZE;
        let relative = usize::try_from(get_u64(bytes, entry)?)
            .map_err(|_| Error::InvalidFormat("id offset is too large".into()))?;
        let length = get_u32(bytes, entry + 8)? as usize;
        let start = ids_offset
            .checked_add(relative)
            .ok_or_else(|| Error::InvalidFormat("id offset overflow".into()))?;
        let end = start
            .checked_add(length)
            .ok_or_else(|| Error::InvalidFormat("id length overflow".into()))?;
        if length == 0 || end > vectors_offset {
            return Err(Error::InvalidFormat("id exceeds id table".into()));
        }
        let id = std::str::from_utf8(&bytes[start..end])
            .map_err(|_| Error::InvalidFormat("record id is not UTF-8".into()))?;
        if !seen.insert(id.to_owned()) {
            return Err(Error::InvalidFormat("duplicate record id".into()));
        }
        ids.push((start, length));
    }
    Ok(ids)
}
