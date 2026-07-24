use std::fs;
use std::io;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU64, Ordering};

use crate::storage::{
    DatasetError, InMemoryGraph, Neighbor, PackedError, PackedGraph, write_packed,
};
use crate::synthetic::{
    Edge, EdgeKind, GraphDataset, GraphSpec, NodeId, ScaleTier, Topology, generate,
};

static PATH_SEQUENCE: AtomicU64 = AtomicU64::new(0);

struct TempPath(PathBuf);

impl TempPath {
    fn new(label: &str) -> Self {
        let sequence = PATH_SEQUENCE.fetch_add(1, Ordering::Relaxed);
        Self(std::env::temp_dir().join(format!(
            "arcana-{label}-{}-{sequence}.pack",
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

fn topology_specs() -> Vec<GraphSpec> {
    vec![
        GraphSpec {
            topology: Topology::Modular {
                cluster_count: 8,
                cross_cluster_ratio: 2_500,
            },
            node_count: 64,
            edge_count: 300,
            seed: 41,
        },
        GraphSpec {
            topology: Topology::Entangled {
                cluster_count: 8,
                hub_count: 4,
            },
            node_count: 64,
            edge_count: 300,
            seed: 42,
        },
        GraphSpec {
            topology: Topology::HubHeavy { hub_count: 4 },
            node_count: 64,
            edge_count: 300,
            seed: 43,
        },
        GraphSpec {
            topology: Topology::Layered { layer_count: 8 },
            node_count: 64,
            edge_count: 300,
            seed: 44,
        },
        GraphSpec {
            topology: Topology::DenseSubsystem {
                dense_node_count: 16,
            },
            node_count: 64,
            edge_count: 300,
            seed: 45,
        },
    ]
}

fn assert_round_trip(dataset: &GraphDataset, label: &str) {
    let path = TempPath::new(label);
    let oracle = InMemoryGraph::new(dataset).expect("valid reference graph");
    let summary = write_packed(path.as_path(), dataset).expect("packed write succeeds");
    let packed = PackedGraph::open(path.as_path()).expect("packed graph opens");

    assert_eq!(packed.node_count(), oracle.node_count());
    assert_eq!(packed.edge_count(), oracle.edge_count());
    assert_eq!(summary.dataset_checksum, packed.dataset_checksum());
    assert_eq!(
        summary.file_len,
        fs::metadata(path.as_path()).unwrap().len()
    );

    for node in 0..dataset.node_count {
        let node = NodeId(node);
        assert_eq!(
            packed.forward_neighbors(node).unwrap(),
            oracle.forward_neighbors(node).unwrap()
        );
        assert_eq!(
            packed.reverse_neighbors(node).unwrap(),
            oracle.reverse_neighbors(node).unwrap()
        );
    }
}

#[test]
fn every_synthetic_topology_round_trips() {
    for (index, spec) in topology_specs().iter().enumerate() {
        let dataset = generate(spec).expect("valid synthetic graph");
        assert_round_trip(&dataset, &format!("topology-{index}"));
    }
}

#[test]
fn empty_adjacency_and_parallel_kinds_round_trip() {
    assert_round_trip(
        &GraphDataset {
            node_count: 5,
            edges: Vec::new(),
        },
        "empty",
    );

    let dataset = GraphDataset {
        node_count: 4,
        edges: vec![
            Edge {
                source: NodeId(0),
                target: NodeId(1),
                kind: EdgeKind(2),
            },
            Edge {
                source: NodeId(2),
                target: NodeId(0),
                kind: EdgeKind(3),
            },
            Edge {
                source: NodeId(0),
                target: NodeId(1),
                kind: EdgeKind(1),
            },
        ],
    };
    assert_round_trip(&dataset, "parallel-kinds");
}

#[test]
fn borrowed_neighbor_iterator_matches_owned_api() {
    let dataset = GraphDataset {
        node_count: 4,
        edges: vec![
            Edge {
                source: NodeId(0),
                target: NodeId(1),
                kind: EdgeKind(2),
            },
            Edge {
                source: NodeId(0),
                target: NodeId(2),
                kind: EdgeKind(3),
            },
            Edge {
                source: NodeId(3),
                target: NodeId(0),
                kind: EdgeKind(4),
            },
        ],
    };
    let path = TempPath::new("borrowed-neighbors");
    write_packed(path.as_path(), &dataset).unwrap();
    let packed = PackedGraph::open(path.as_path()).unwrap();

    let forward: Vec<_> = packed.forward_neighbors_iter(NodeId(0)).unwrap().collect();
    let reverse: Vec<_> = packed.reverse_neighbors_iter(NodeId(0)).unwrap().collect();
    assert_eq!(forward, packed.forward_neighbors(NodeId(0)).unwrap());
    assert_eq!(reverse, packed.reverse_neighbors(NodeId(0)).unwrap());
}

#[test]
fn logical_edge_order_does_not_change_packed_bytes() {
    let dataset = generate(&topology_specs()[0]).expect("valid synthetic graph");
    let mut reordered = dataset.clone();
    reordered.edges.reverse();
    let first = TempPath::new("deterministic-a");
    let second = TempPath::new("deterministic-b");

    write_packed(first.as_path(), &dataset).unwrap();
    write_packed(second.as_path(), &reordered).unwrap();
    assert_eq!(
        fs::read(first.as_path()).unwrap(),
        fs::read(second.as_path()).unwrap()
    );
}

#[test]
fn self_edges_round_trip() {
    let dataset = GraphDataset {
        node_count: 2,
        edges: vec![Edge {
            source: NodeId(1),
            target: NodeId(1),
            kind: EdgeKind(7),
        }],
    };

    assert_round_trip(&dataset, "self-edge");
}

#[test]
fn writer_rejects_invalid_datasets() {
    let invalid = [
        (
            GraphDataset {
                node_count: 2,
                edges: vec![Edge {
                    source: NodeId(0),
                    target: NodeId(2),
                    kind: EdgeKind(0),
                }],
            },
            "range",
        ),
        (
            GraphDataset {
                node_count: 2,
                edges: vec![
                    Edge {
                        source: NodeId(0),
                        target: NodeId(1),
                        kind: EdgeKind(0),
                    },
                    Edge {
                        source: NodeId(0),
                        target: NodeId(1),
                        kind: EdgeKind(0),
                    },
                ],
            },
            "duplicate",
        ),
    ];

    for (dataset, label) in invalid {
        let path = TempPath::new(label);
        assert!(matches!(
            write_packed(path.as_path(), &dataset),
            Err(PackedError::Dataset(
                DatasetError::EndpointOutOfRange { .. } | DatasetError::DuplicateEdge { .. }
            ))
        ));
        assert!(!path.as_path().exists());
    }
}

#[test]
fn writer_refuses_to_replace_an_existing_snapshot() {
    let path = TempPath::new("existing");
    fs::write(path.as_path(), b"owned").unwrap();
    let dataset = GraphDataset {
        node_count: 1,
        edges: Vec::new(),
    };

    assert!(matches!(
        write_packed(path.as_path(), &dataset),
        Err(PackedError::Io(error)) if error.kind() == io::ErrorKind::AlreadyExists
    ));
    assert_eq!(fs::read(path.as_path()).unwrap(), b"owned");
}

#[test]
#[ignore = "medium-scale storage smoke"]
fn medium_scale_packed_smoke() {
    let dataset = generate(&GraphSpec::for_tier(
        Topology::Modular {
            cluster_count: 1_000,
            cross_cluster_ratio: 2_500,
        },
        ScaleTier::Medium,
        77,
    ))
    .expect("valid medium graph");
    let path = TempPath::new("medium-scale");
    let summary = write_packed(path.as_path(), &dataset).expect("medium packed write");
    let packed = PackedGraph::open(path.as_path()).expect("medium packed open");

    assert_eq!(summary.node_count, 100_000);
    assert_eq!(summary.edge_count, 1_000_000);
    for node in [0, 1, 999, 50_000, 99_999] {
        let mut forward: Vec<_> = dataset
            .edges
            .iter()
            .filter(|edge| edge.source == NodeId(node))
            .map(|edge| Neighbor {
                node: edge.target,
                kind: edge.kind,
            })
            .collect();
        let mut reverse: Vec<_> = dataset
            .edges
            .iter()
            .filter(|edge| edge.target == NodeId(node))
            .map(|edge| Neighbor {
                node: edge.source,
                kind: edge.kind,
            })
            .collect();
        forward.sort_unstable();
        reverse.sort_unstable();
        assert_eq!(packed.forward_neighbors(NodeId(node)).unwrap(), forward);
        assert_eq!(packed.reverse_neighbors(NodeId(node)).unwrap(), reverse);
    }
}
