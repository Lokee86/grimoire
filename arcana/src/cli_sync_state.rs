use std::fs::{self, File, OpenOptions};
use std::io::{self, Write};
use std::path::Path;

use atomicwrites::replace_atomic;
use fs2::FileExt;

pub struct SyncLock {
    file: File,
}

impl SyncLock {
    pub fn acquire(state: &Path) -> io::Result<Self> {
        let file = OpenOptions::new()
            .read(true)
            .write(true)
            .create(true)
            .truncate(false)
            .open(state.join("LOCK"))?;
        file.try_lock_exclusive().map_err(|error| {
            io::Error::new(error.kind(), format!("Arcana state is busy: {error}"))
        })?;
        Ok(Self { file })
    }
}

impl Drop for SyncLock {
    fn drop(&mut self) {
        let _ = FileExt::unlock(&self.file);
    }
}

pub fn replace_file(path: &Path, bytes: &[u8]) -> io::Result<()> {
    let temp = path.with_extension(format!("tmp-{}", std::process::id()));
    let result = (|| {
        let mut file = OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&temp)?;
        file.write_all(bytes)?;
        file.sync_all()?;
        drop(file);
        replace_atomic(&temp, path)
    })();
    if result.is_err() {
        let _ = fs::remove_file(&temp);
    }
    result
}

#[cfg(test)]
mod tests {
    use std::fs;
    use std::path::PathBuf;
    use std::sync::atomic::{AtomicUsize, Ordering};

    use super::{SyncLock, replace_file};

    #[test]
    fn serializes_state_writers() {
        let directory = TestDirectory::new();
        let first = SyncLock::acquire(&directory.path).unwrap();
        assert!(SyncLock::acquire(&directory.path).is_err());
        drop(first);
        SyncLock::acquire(&directory.path).unwrap();
    }

    #[test]
    fn atomically_replaces_existing_file() {
        let directory = TestDirectory::new();
        let path = directory.path.join("CURRENT");
        fs::write(&path, b"old\n").unwrap();
        replace_file(&path, b"new\n").unwrap();
        assert_eq!(fs::read(&path).unwrap(), b"new\n");
    }

    struct TestDirectory {
        path: PathBuf,
    }

    impl TestDirectory {
        fn new() -> Self {
            static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
            let path = std::env::temp_dir().join(format!(
                "arcana-sync-state-test-{}-{}",
                std::process::id(),
                SEQUENCE.fetch_add(1, Ordering::Relaxed)
            ));
            fs::create_dir(&path).unwrap();
            Self { path }
        }
    }

    impl Drop for TestDirectory {
        fn drop(&mut self) {
            let _ = fs::remove_dir_all(&self.path);
        }
    }
}
