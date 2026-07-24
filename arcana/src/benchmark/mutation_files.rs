use std::fs;
use std::path::{Path, PathBuf};

pub(super) fn mutation_path(work_dir: &Path, run_id: u64, suffix: &str) -> PathBuf {
    work_dir.join(format!(
        "arcana-mutation-{}-{run_id}-{suffix}",
        std::process::id()
    ))
}

pub(super) struct GeneratedFiles {
    paths: Vec<PathBuf>,
    keep_files: bool,
}

impl GeneratedFiles {
    pub fn new(keep_files: bool) -> Self {
        Self {
            paths: Vec::new(),
            keep_files,
        }
    }

    pub fn push(&mut self, path: PathBuf) {
        self.paths.push(path);
    }

    pub fn extend<const N: usize>(&mut self, paths: [PathBuf; N]) {
        self.paths.extend(paths);
    }
}

impl Drop for GeneratedFiles {
    fn drop(&mut self) {
        if self.keep_files {
            return;
        }
        for path in &self.paths {
            let _ = fs::remove_file(path);
        }
    }
}
