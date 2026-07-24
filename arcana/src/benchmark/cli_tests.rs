use std::path::Path;

use super::{BenchmarkCommand, BenchmarkParseError, ScaleTier, topology_preset};

#[test]
fn parses_all_options_and_benchmark_prefix() {
    let command = BenchmarkCommand::parse([
        "benchmark",
        "--tier",
        "large",
        "--topology=dense-subsystem",
        "--queries",
        "2500",
        "--samples",
        "7",
        "--seed=99",
        "--csv",
        "reports/bench.csv",
        "--work-dir=tmp/bench",
        "--keep-files",
    ])
    .expect("options should parse");

    assert_eq!(command.graph.node_count, 1_000_000);
    assert_eq!(command.graph.edge_count, 10_000_000);
    assert_eq!(command.graph.seed, 99);
    assert_eq!(command.query_count, 2_500);
    assert_eq!(command.sample_count, 7);
    assert_eq!(
        command.csv_path.as_deref(),
        Some(Path::new("reports/bench.csv"))
    );
    assert_eq!(command.work_dir, Path::new("tmp/bench"));
    assert!(command.keep_files);
}
#[test]
fn reports_missing_unknown_and_invalid_arguments() {
    assert!(matches!(
        BenchmarkCommand::parse(["--csv"]),
        Err(BenchmarkParseError::MissingValue { option: "--csv" })
    ));
    assert!(matches!(
        BenchmarkCommand::parse(["--nope"]),
        Err(BenchmarkParseError::UnknownFlag(flag)) if flag == "--nope"
    ));
    assert!(matches!(
        BenchmarkCommand::parse(["--queries", "many"]),
        Err(BenchmarkParseError::InvalidNumber {
            option: "--queries",
            ..
        })
    ));
}
#[test]
fn reports_unsupported_tier_and_topology_names() {
    assert!(matches!(
        BenchmarkCommand::parse(["--tier", "huge"]),
        Err(BenchmarkParseError::UnsupportedTier(tier)) if tier == "huge"
    ));
    assert!(matches!(
        BenchmarkCommand::parse(["--topology", "random"]),
        Err(BenchmarkParseError::UnsupportedTopology(topology)) if topology == "random"
    ));
}
#[test]
fn every_tier_and_topology_preset_validates() {
    let tiers = [
        ScaleTier::Small,
        ScaleTier::Medium,
        ScaleTier::Large,
        ScaleTier::Stress,
    ];
    let topologies = [
        "modular",
        "entangled",
        "hub-heavy",
        "layered",
        "dense-subsystem",
    ];
    for tier in tiers {
        for topology in topologies {
            let spec = topology_preset(tier, topology, 42).expect("preset should parse");
            assert_eq!(spec.validate(), Ok(()));
        }
    }
}
