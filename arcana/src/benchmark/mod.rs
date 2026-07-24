//! Deterministic graph mutation benchmarks and reporting.

mod cli;
#[cfg(test)]
mod cli_tests;
mod common;
mod error;
mod mutation_files;
mod mutation_plan;
mod mutation_query;
mod mutation_runner;
mod report;
mod workload;

pub use crate::synthetic::{GraphSpec, ScaleTier, Topology};
pub use cli::{BenchmarkCommand, BenchmarkParseError, benchmark_usage, topology_preset};
pub use common::BenchmarkConfig;
pub use error::BenchmarkError;
pub use mutation_runner::run_mutation_benchmark;
pub use report::{Backend, BenchmarkMetric, BenchmarkReport, BenchmarkSample};
pub use workload::{
    QueryDirection, QueryPattern, QueryWorkload, QueryWorkloadError, generate_workload,
};

#[cfg(test)]
mod mutation_runner_tests;

#[cfg(test)]
mod tests {
    use std::time::Duration;

    use super::{Backend, BenchmarkMetric, BenchmarkReport, BenchmarkSample};

    fn sample(backend: Backend, sample: u32, duration_ms: u64) -> BenchmarkSample {
        BenchmarkSample::new(
            "graph-a",
            backend,
            BenchmarkMetric::Query,
            "neighbors",
            sample,
            Duration::from_millis(duration_ms),
            100,
            64,
            2_048,
            0x2a,
        )
    }

    #[test]
    fn csv_contains_header_and_escaped_content() {
        let mut report = BenchmarkReport::new();
        report.push(BenchmarkSample::new(
            "graph,a",
            Backend::Overlay,
            BenchmarkMetric::Query,
            "forward\"neighbors",
            2,
            Duration::from_nanos(1_250),
            3,
            4,
            5,
            42,
        ));
        assert_eq!(
            report.to_csv(),
            "graph,backend,metric,workload,sample,duration_ns,operations,items,file_size,fingerprint\n\"graph,a\",overlay,query,\"forward\"\"neighbors\",2,1250,3,4,5,42\n"
        );
    }

    #[test]
    fn human_summary_uses_medians_and_query_throughput() {
        let mut report = BenchmarkReport::new();
        report.push(sample(Backend::Overlay, 0, 10));
        report.push(sample(Backend::Overlay, 1, 30));
        report.push(sample(Backend::RebuiltPacked, 0, 20));
        report.push(sample(Backend::RebuiltPacked, 1, 40));
        let summary = report.human_summary();
        assert!(summary.contains("overlay median 20.000ms"));
        assert!(summary.contains("rebuilt-packed median 30.000ms"));
        assert!(summary.contains("speedup 1.50x"));
        assert!(summary.contains("overlay throughput=6666.7 ops/s"));
        assert!(summary.contains("rebuilt-packed throughput=3750.0 ops/s"));
        assert!(summary.contains("file_size overlay=2048B rebuilt-packed=2048B"));
    }
}
