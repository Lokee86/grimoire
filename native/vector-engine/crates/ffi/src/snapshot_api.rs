use std::{ptr, slice};

use grimoire_vector_core::{Snapshot, search};

use crate::{
    GV_BUFFER_TOO_SMALL, GV_INVALID_HANDLE, GV_OK, GvSearchResult,
    boundary::{boundary, path, write_bytes},
    state,
};

/// Open and register an immutable packed snapshot.
///
/// # Safety
/// The path must be readable for `path_len`; `out_handle` must be writable.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_open_snapshot(
    path_ptr: *const u8,
    path_len: usize,
    out_handle: *mut u64,
) -> i32 {
    boundary(|| {
        if out_handle.is_null() {
            return Err("out_handle is null".into());
        }
        let snapshot = Snapshot::open(unsafe { path(path_ptr, path_len, "snapshot") }?)
            .map_err(|error| error.to_string())?;
        unsafe { ptr::write(out_handle, state::insert(snapshot)) };
        Ok(GV_OK)
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn gv_close_snapshot(handle: u64) -> i32 {
    boundary(|| {
        if state::remove(handle) {
            Ok(GV_OK)
        } else {
            Ok(GV_INVALID_HANDLE)
        }
    })
}

/// Read metadata for an open snapshot handle.
///
/// # Safety
/// Output pointers must be writable and the model buffer valid for its capacity.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_snapshot_info(
    handle: u64,
    out_dimensions: *mut u32,
    out_count: *mut u64,
    model_buffer: *mut u8,
    model_capacity: usize,
    out_model_len: *mut usize,
) -> i32 {
    boundary(|| {
        if out_dimensions.is_null() || out_count.is_null() || out_model_len.is_null() {
            return Err("snapshot info output is null".into());
        }
        let snapshot = state::get(handle).ok_or_else(|| "invalid snapshot handle".to_string())?;
        let info = snapshot.info();
        unsafe {
            ptr::write(out_dimensions, info.dimensions as u32);
            ptr::write(out_count, info.count as u64);
            ptr::write(out_model_len, info.model.len());
        }
        if model_capacity < info.model.len() {
            return Ok(GV_BUFFER_TOO_SMALL);
        }
        unsafe { write_bytes(model_buffer, model_capacity, info.model.as_bytes()) }?;
        Ok(GV_OK)
    })
}

/// Search an open snapshot and write exact top-K results into caller buffers.
///
/// # Safety
/// The query must be readable and every output pointer valid for its declared capacity.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_search(
    handle: u64,
    query_ptr: *const f32,
    query_len: usize,
    top_k: usize,
    results_ptr: *mut GvSearchResult,
    results_capacity: usize,
    ids_ptr: *mut u8,
    ids_capacity: usize,
    out_count: *mut usize,
    out_ids_len: *mut usize,
) -> i32 {
    boundary(|| {
        if query_ptr.is_null() || out_count.is_null() || out_ids_len.is_null() {
            return Err("search input or output is null".into());
        }
        let snapshot = state::get(handle).ok_or_else(|| "invalid snapshot handle".to_string())?;
        let query = unsafe { slice::from_raw_parts(query_ptr, query_len) };
        let hits = search(&snapshot, query, top_k).map_err(|error| error.to_string())?;
        let ids_len: usize = hits.iter().map(|hit| hit.id.len()).sum();
        unsafe {
            ptr::write(out_count, hits.len());
            ptr::write(out_ids_len, ids_len);
        }
        if results_capacity < hits.len() || ids_capacity < ids_len {
            return Ok(GV_BUFFER_TOO_SMALL);
        }
        if !hits.is_empty() && (results_ptr.is_null() || ids_ptr.is_null()) {
            return Err("search result buffers are null".into());
        }
        let mut id_offset = 0_usize;
        for (position, hit) in hits.iter().enumerate() {
            unsafe {
                write_bytes(
                    ids_ptr.add(id_offset),
                    ids_capacity - id_offset,
                    hit.id.as_bytes(),
                )
            }?;
            unsafe {
                ptr::write(
                    results_ptr.add(position),
                    GvSearchResult {
                        id_offset: id_offset as u64,
                        id_len: hit.id.len() as u32,
                        score: hit.score,
                        index: hit.index as u64,
                        ..Default::default()
                    },
                )
            };
            id_offset += hit.id.len();
        }
        Ok(GV_OK)
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use grimoire_vector_core::{ObjectStore, RecordRef, materialize};

    #[test]
    fn close_invalidates_handle_without_freeing_active_arc() {
        let temp = tempfile::tempdir().unwrap();
        let store = ObjectStore::new(temp.path().join("store"));
        store.put("m", "a", &[1.0, 0.0]).unwrap();
        let path = temp.path().join("snapshot.gvs");
        materialize(
            &store,
            "m",
            &[RecordRef {
                id: "a".into(),
                source: "a".into(),
            }],
            &path,
        )
        .unwrap();
        let handle = state::insert(Snapshot::open(path).unwrap());
        let active = state::get(handle).unwrap();
        assert_eq!(gv_close_snapshot(handle), GV_OK);
        assert!(state::get(handle).is_none());
        assert_eq!(active.info().count, 1);
    }
}
