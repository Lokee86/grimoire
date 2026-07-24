use super::{GraphSpec, GraphSpecError, ScaleTier, Topology};

#[test]
fn scale_tiers_define_increasing_workloads() {
    assert_eq!(ScaleTier::Small.counts(), (10_000, 100_000));
    assert_eq!(ScaleTier::Medium.counts(), (100_000, 1_000_000));
    assert_eq!(ScaleTier::Large.counts(), (1_000_000, 10_000_000));
    assert_eq!(ScaleTier::Stress.counts(), (5_000_000, 50_000_000));
}

#[test]
fn rejects_invalid_global_bounds() {
    let spec = GraphSpec {
        topology: Topology::HubHeavy { hub_count: 1 },
        node_count: 0,
        edge_count: 0,
        seed: 1,
    };
    assert_eq!(spec.validate(), Err(GraphSpecError::ZeroNodes));

    let spec = GraphSpec {
        topology: Topology::HubHeavy { hub_count: 1 },
        node_count: 2,
        edge_count: 3,
        seed: 1,
    };
    assert!(matches!(
        spec.validate(),
        Err(GraphSpecError::EdgeCountExceedsCapacity { .. })
    ));
}

#[test]
fn rejects_invalid_topology_parameters() {
    let spec = GraphSpec {
        topology: Topology::Modular {
            cluster_count: 2,
            cross_cluster_ratio: 10_001,
        },
        node_count: 4,
        edge_count: 2,
        seed: 1,
    };
    assert!(matches!(
        spec.validate(),
        Err(GraphSpecError::InvalidBasisPointRatio { .. })
    ));

    let spec = GraphSpec {
        topology: Topology::Layered { layer_count: 1 },
        node_count: 4,
        edge_count: 2,
        seed: 1,
    };
    assert!(matches!(
        spec.validate(),
        Err(GraphSpecError::PartitionTooSmall { .. })
    ));
}
