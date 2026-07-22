use std::ptr;

use grimoire_vector_core::{ObjectStore, ingest_jsonl, materialize_jsonl};

use crate::{
    GV_BUFFER_TOO_SMALL, GV_OK,
    boundary::{boundary, path, text, write_bytes},
};

/// Test whether an immutable vector object exists.
///
/// # Safety
/// Input pointers must be readable for their lengths; `out_exists` must be writable.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_object_exists(
    store_ptr: *const u8,
    store_len: usize,
    model_ptr: *const u8,
    model_len: usize,
    source_ptr: *const u8,
    source_len: usize,
    out_exists: *mut u8,
) -> i32 {
    boundary(|| {
        if out_exists.is_null() {
            return Err("out_exists is null".into());
        }
        let store = unsafe { path(store_ptr, store_len, "store") }?;
        let model = unsafe { text(model_ptr, model_len, "model") }?;
        let source = unsafe { text(source_ptr, source_len, "source") }?;
        unsafe {
            ptr::write(
                out_exists,
                ObjectStore::new(store).contains(model, source) as u8,
            )
        };
        Ok(GV_OK)
    })
}

/// Ingest vector records from a JSONL file.
///
/// # Safety
/// Input pointers must be readable for their lengths; `out_count` must be writable.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_ingest_jsonl(
    store_ptr: *const u8,
    store_len: usize,
    model_ptr: *const u8,
    model_len: usize,
    input_ptr: *const u8,
    input_len: usize,
    out_count: *mut u64,
) -> i32 {
    boundary(|| {
        if out_count.is_null() {
            return Err("out_count is null".into());
        }
        let store = ObjectStore::new(unsafe { path(store_ptr, store_len, "store") }?);
        let model = unsafe { text(model_ptr, model_len, "model") }?;
        let input = unsafe { path(input_ptr, input_len, "input") }?;
        let count = ingest_jsonl(&store, model, input).map_err(|error| error.to_string())?;
        unsafe { ptr::write(out_count, count as u64) };
        Ok(GV_OK)
    })
}

/// Materialize a packed snapshot from a JSONL manifest.
///
/// # Safety
/// Inputs must be readable; output pointers must be valid for their declared capacities.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_materialize_jsonl(
    store_ptr: *const u8,
    store_len: usize,
    model_ptr: *const u8,
    model_len: usize,
    manifest_ptr: *const u8,
    manifest_len: usize,
    snapshot_ptr: *const u8,
    snapshot_len: usize,
    identity_buffer: *mut u8,
    identity_capacity: usize,
    out_identity_len: *mut usize,
) -> i32 {
    boundary(|| {
        if out_identity_len.is_null() {
            return Err("out_identity_len is null".into());
        }
        let store = ObjectStore::new(unsafe { path(store_ptr, store_len, "store") }?);
        let model = unsafe { text(model_ptr, model_len, "model") }?;
        let manifest = unsafe { path(manifest_ptr, manifest_len, "manifest") }?;
        let snapshot = unsafe { path(snapshot_ptr, snapshot_len, "snapshot") }?;
        let info = materialize_jsonl(&store, model, manifest, snapshot)
            .map_err(|error| error.to_string())?;
        unsafe { ptr::write(out_identity_len, info.identity.len()) };
        if identity_capacity < info.identity.len() {
            return Ok(GV_BUFFER_TOO_SMALL);
        }
        unsafe { write_bytes(identity_buffer, identity_capacity, info.identity.as_bytes()) }?;
        Ok(GV_OK)
    })
}
