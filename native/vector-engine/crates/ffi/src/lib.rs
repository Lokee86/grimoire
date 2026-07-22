mod boundary;
mod snapshot_api;
mod state;
mod storage_api;

use std::ptr;

pub const GV_OK: i32 = 0;
pub const GV_INVALID_ARGUMENT: i32 = 1;
pub const GV_BUFFER_TOO_SMALL: i32 = 2;
pub const GV_INVALID_HANDLE: i32 = 3;
pub const GV_ERROR: i32 = 4;

#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct GvSearchResult {
    pub id_offset: u64,
    pub id_len: u32,
    pub reserved0: u32,
    pub score: f32,
    pub reserved1: u32,
    pub index: u64,
}

#[unsafe(no_mangle)]
pub extern "C" fn gv_abi_version() -> u32 {
    1
}

/// Copy the current thread's last error into caller-owned storage.
///
/// # Safety
/// `buffer` must be null or writable for `capacity` bytes for this call.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn gv_last_error_message(buffer: *mut u8, capacity: usize) -> usize {
    let message = state::last_error();
    if !buffer.is_null() && capacity > 0 {
        let count = capacity.min(message.len());
        // SAFETY: caller promises a writable buffer of at least capacity bytes.
        unsafe { ptr::copy_nonoverlapping(message.as_ptr(), buffer, count) };
    }
    message.len()
}
