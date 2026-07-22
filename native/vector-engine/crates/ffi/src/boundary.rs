use std::{
    panic::{AssertUnwindSafe, catch_unwind},
    path::PathBuf,
    ptr, slice,
};

use crate::{GV_ERROR, state};

pub fn boundary(operation: impl FnOnce() -> Result<i32, String>) -> i32 {
    match catch_unwind(AssertUnwindSafe(operation)) {
        Ok(Ok(status)) => status,
        Ok(Err(message)) => {
            state::set_error(message);
            GV_ERROR
        }
        Err(_) => {
            state::set_error("panic contained at vector ABI boundary");
            GV_ERROR
        }
    }
}

pub unsafe fn text<'a>(pointer: *const u8, length: usize, name: &str) -> Result<&'a str, String> {
    if pointer.is_null() || length == 0 {
        return Err(format!("{name} is empty"));
    }
    // SAFETY: the caller guarantees pointer validity for length bytes for this call.
    let bytes = unsafe { slice::from_raw_parts(pointer, length) };
    std::str::from_utf8(bytes).map_err(|_| format!("{name} is not UTF-8"))
}

pub unsafe fn path(pointer: *const u8, length: usize, name: &str) -> Result<PathBuf, String> {
    // SAFETY: forwarded from the same ABI call and not retained.
    unsafe { text(pointer, length, name) }.map(PathBuf::from)
}

pub unsafe fn write_bytes(pointer: *mut u8, capacity: usize, bytes: &[u8]) -> Result<(), String> {
    if bytes.len() > capacity || (!bytes.is_empty() && pointer.is_null()) {
        return Err("output buffer is too small or null".into());
    }
    if !bytes.is_empty() {
        // SAFETY: the caller guarantees a writable buffer of capacity bytes.
        unsafe { ptr::copy_nonoverlapping(bytes.as_ptr(), pointer, bytes.len()) };
    }
    Ok(())
}
