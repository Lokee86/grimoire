use std::fs;
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};

use crate::synthetic::{GraphSpec, Topology};

use super::{Backend, BenchmarkConfig, BenchmarkMetric, run_mutation_benchmark};

static PATH_SEQUENCE: AtomicU64 = AtomicU64::new(0);

struct TempDirectory(PathBuf);

impl TempDirectory {
    fn new() -> Self {
        let sequence = PATH_SEQUENCE.fetch_add(1, Ordering::Relaxed);
        let path = std::env::temp_dir().join(format!(
            "arcana-mutation-benchmark-test-{}-{sequence}",
            std::process::id()
        ));
        fs::create_dir(&path).unwrap();
        Self(path)
    }
}

impl Drop for TempDirectory {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[test]
fn all_mutation_patterns_match_rebuilt_packed_results_and_cleanup() {
    let directory = TempDirectory::new();
    let config = BenchmarkConfig::new(
        GraphSpec {
            topology: Topology::Modular {
                cluster_count: 8,
                cross_cluster_ratio: 2_500,
            },
            node_count: 64,
            edge_count: 300,
            seed: 71,
        },
        20,
        1,
        &directory.0,
        false,
    );
    let report = run_mutation_benchmark(&config).unwrap();
    assert_eq!(report.samples().len(), 80);
    assert_eq!(
        report
            .samples()
            .iter()
            .filter(|sample| sample.metric == BenchmarkMetric::Mutation)
            .count(),
        10
    );
    assert!(
        report
            .samples()
            .iter()
            .any(|sample| sample.backend == Backend::Overlay)
    );
    assert!(
        report
            .samples()
            .iter()
            .any(|sample| sample.backend == Backend::RebuiltPacked)
    );
    let summary = report.human_summary();
    assert!(summary.contains("overlay median"));
    assert!(summary.contains("rebuilt-packed median"));
    assert_eq!(fs::read_dir(&directory.0).unwrap().count(), 0);
}
