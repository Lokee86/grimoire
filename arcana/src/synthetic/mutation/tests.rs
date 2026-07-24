use std::collections::HashSet;

use super::{GraphMutation, MutationError, MutationScenario, apply_mutation, plan_mutation};
use crate::synthetic::{Edge, GraphDataset, GraphSpec, NodeId, Topology, generate};

fn dataset() -> GraphDataset {
    generate(&GraphSpec {
        topology: Topology::HubHeavy { hub_count: 2 },
        node_count: 24,
        edge_count: 120,
        seed: 11,
    })
    .expect("valid graph")
}

#[test]
fn every_scenario_plans_and_applies_deterministically() {
    let dataset = dataset();
    let scenarios = [
        MutationScenario::SingleNode { node: NodeId(0) },
        MutationScenario::LocalRange {
            start: NodeId(2),
            node_count: 4,
            edge_count: 5,
        },
        MutationScenario::Scattered { edge_count: 7 },
        MutationScenario::Hub {
            hub_count: 2,
            edge_count: 6,
        },
        MutationScenario::Percentage { basis_points: 500 },
    ];

    for scenario in scenarios {
        let first = plan_mutation(&dataset, scenario, 19).expect("valid mutation");
        let second = plan_mutation(&dataset, scenario, 19).expect("valid mutation");
        assert_eq!(first, second);
        assert_eq!(first.removed.len(), first.added.len());

        let original_pairs: HashSet<_> = dataset
            .edges
            .iter()
            .map(|edge| (edge.source, edge.target))
            .collect();
        assert!(
            first
                .added
                .iter()
                .all(|edge| !original_pairs.contains(&(edge.source, edge.target)))
        );

        let updated = apply_mutation(&dataset, &first).expect("planned mutation applies");
        let unique: HashSet<Edge> = updated.edges.iter().copied().collect();
        assert_eq!(updated.node_count, dataset.node_count);
        assert_eq!(updated.edges.len(), dataset.edges.len());
        assert_eq!(unique.len(), updated.edges.len());
        assert!(updated.edges.windows(2).all(|pair| pair[0] < pair[1]));
    }
}

#[test]
fn removed_edge_kind_counts_are_preserved_by_additions() {
    let dataset = dataset();
    let mutation = plan_mutation(&dataset, MutationScenario::Scattered { edge_count: 20 }, 23)
        .expect("valid mutation");
    let mut removed: Vec<_> = mutation.removed.iter().map(|edge| edge.kind).collect();
    let mut added: Vec<_> = mutation.added.iter().map(|edge| edge.kind).collect();
    removed.sort_unstable();
    added.sort_unstable();
    assert_eq!(removed, added);
}

#[test]
fn invalid_requests_return_explicit_errors() {
    let dataset = dataset();
    assert!(matches!(
        plan_mutation(
            &dataset,
            MutationScenario::SingleNode { node: NodeId(24) },
            1
        ),
        Err(MutationError::InvalidNode { .. })
    ));
    assert!(matches!(
        plan_mutation(
            &dataset,
            MutationScenario::LocalRange {
                start: NodeId(23),
                node_count: 2,
                edge_count: 1,
            },
            1
        ),
        Err(MutationError::InvalidRange { .. })
    ));
    assert_eq!(
        plan_mutation(&dataset, MutationScenario::Scattered { edge_count: 0 }, 1),
        Err(MutationError::ZeroRequestedEdges)
    );
    assert!(matches!(
        plan_mutation(
            &dataset,
            MutationScenario::Percentage {
                basis_points: 10_001
            },
            1
        ),
        Err(MutationError::InvalidPercentage { .. })
    ));
}

#[test]
fn full_pair_graph_has_no_replacement_capacity() {
    let full = generate(&GraphSpec {
        topology: Topology::DenseSubsystem {
            dense_node_count: 3,
        },
        node_count: 3,
        edge_count: 6,
        seed: 2,
    })
    .expect("complete graph");

    assert!(matches!(
        plan_mutation(&full, MutationScenario::Scattered { edge_count: 1 }, 3),
        Err(MutationError::InsufficientReplacementCapacity { .. })
    ));
}

#[test]
fn apply_rejects_count_mismatch() {
    let dataset = dataset();
    let mutation = GraphMutation {
        removed: vec![dataset.edges[0]],
        added: Vec::new(),
    };
    assert!(matches!(
        apply_mutation(&dataset, &mutation),
        Err(MutationError::CountMismatch { .. })
    ));
}
