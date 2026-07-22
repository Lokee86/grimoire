use std::{
    cell::RefCell,
    collections::HashMap,
    sync::{
        Arc, OnceLock, RwLock,
        atomic::{AtomicU64, Ordering},
    },
};

use grimoire_vector_core::Snapshot;

thread_local! {
    static LAST_ERROR: RefCell<Vec<u8>> = const { RefCell::new(Vec::new()) };
}

static NEXT_HANDLE: AtomicU64 = AtomicU64::new(1);
static SNAPSHOTS: OnceLock<RwLock<HashMap<u64, Arc<Snapshot>>>> = OnceLock::new();

pub fn set_error(message: impl Into<String>) {
    LAST_ERROR.with(|slot| *slot.borrow_mut() = message.into().into_bytes());
}

pub fn last_error() -> Vec<u8> {
    LAST_ERROR.with(|slot| slot.borrow().clone())
}

pub fn insert(snapshot: Snapshot) -> u64 {
    let handle = NEXT_HANDLE.fetch_add(1, Ordering::Relaxed);
    snapshots()
        .write()
        .expect("snapshot registry poisoned")
        .insert(handle, Arc::new(snapshot));
    handle
}

pub fn get(handle: u64) -> Option<Arc<Snapshot>> {
    snapshots()
        .read()
        .expect("snapshot registry poisoned")
        .get(&handle)
        .cloned()
}

pub fn remove(handle: u64) -> bool {
    snapshots()
        .write()
        .expect("snapshot registry poisoned")
        .remove(&handle)
        .is_some()
}

fn snapshots() -> &'static RwLock<HashMap<u64, Arc<Snapshot>>> {
    SNAPSHOTS.get_or_init(|| RwLock::new(HashMap::new()))
}
