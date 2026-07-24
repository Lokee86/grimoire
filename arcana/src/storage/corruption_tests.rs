use std::fs;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};

use crate::storage::{PackedError, PackedGraph, QueryError, write_packed};
use crate::synthetic::{GraphDataset, GraphSpec, NodeId, Topology, generate};

static PATH_SEQUENCE: AtomicU64 = AtomicU64::new(0);

struct TempPath(PathBuf);

impl TempPath {
    fn new(label: &str) -> Self {
        let sequence = PATH_SEQUENCE.fetch_add(1, Ordering::Relaxed);
        Self(std::env::temp_dir().join(format!(
            "arcana-corrupt-{label}-{}-{sequence}.pack",
            std::process::id()
        )))
    }

    fn as_path(&self) -> &Path {
        &self.0
    }
}

impl Drop for TempPath {
    fn drop(&mut self) {
        let _ = fs::remove_file(&self.0);
    }
}

#[test]
fn reader_rejects_corrupt_headers_and_payloads() {
    let dataset = generate(&GraphSpec {
        topology: Topology::Entangled {
            cluster_count: 8,
            hub_count: 4,
        },
        node_count: 64,
        edge_count: 300,
        seed: 42,
    })
    .expect("valid synthetic graph");
    let valid = TempPath::new("source");
    write_packed(valid.as_path(), &dataset).unwrap();
    let original = fs::read(valid.as_path()).unwrap();

    type MutateBytes = fn(&mut Vec<u8>);
    type ErrorPredicate = fn(PackedError) -> bool;
    let cases: Vec<(&str, MutateBytes, ErrorPredicate)> = vec![
        (
            "magic",
            |bytes| bytes[0] ^= 0xff,
            |error| matches!(error, PackedError::InvalidMagic),
        ),
        (
            "version",
            |bytes| bytes[8..10].copy_from_slice(&2_u16.to_le_bytes()),
            |error| matches!(error, PackedError::UnsupportedVersion { found: 2 }),
        ),
        (
            "layout",
            |bytes| bytes[48..56].copy_from_slice(&136_u64.to_le_bytes()),
            |error| matches!(error, PackedError::LayoutMismatch { .. }),
        ),
        (
            "dataset-checksum",
            |bytes| bytes[32] ^= 1,
            |error| matches!(error, PackedError::DatasetChecksumMismatch { .. }),
        ),
        (
            "payload",
            |bytes| *bytes.last_mut().unwrap() ^= 1,
            |error| matches!(error, PackedError::PayloadChecksumMismatch { .. }),
        ),
        (
            "truncated",
            |bytes| bytes.truncate(64),
            |error| matches!(error, PackedError::FileTooShort { .. }),
        ),
    ];

    for (label, mutate, expected) in cases {
        let path = TempPath::new(label);
        let mut bytes = original.clone();
        mutate(&mut bytes);
        fs::write(path.as_path(), bytes).unwrap();
        let error = PackedGraph::open(path.as_path()).expect_err("corruption must fail");
        assert!(expected(error), "unexpected error for {label}");
    }
}

#[test]
fn invalid_node_queries_are_explicit() {
    let dataset = GraphDataset {
        node_count: 2,
        edges: Vec::new(),
    };
    let path = TempPath::new("invalid-query");
    write_packed(path.as_path(), &dataset).unwrap();
    let packed = PackedGraph::open(path.as_path()).unwrap();

    assert_eq!(
        packed.forward_neighbors(NodeId(2)),
        Err(QueryError::InvalidNode {
            node: NodeId(2),
            node_count: 2,
        })
    );
}
