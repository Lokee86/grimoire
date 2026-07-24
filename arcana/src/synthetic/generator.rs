use super::{GraphDataset, GraphSpec, GraphSpecError, Topology};

/// Generates a graph described by `spec`.
pub fn generate(spec: &GraphSpec) -> Result<GraphDataset, GraphSpecError> {
    spec.validate()?;

    let dataset = match spec.topology {
        Topology::Modular {
            cluster_count,
            cross_cluster_ratio,
        } => super::modular::generate(spec, cluster_count, cross_cluster_ratio),
        Topology::Entangled {
            cluster_count,
            hub_count,
        } => super::entangled::generate(spec, cluster_count, hub_count),
        Topology::HubHeavy { hub_count } => super::hub_heavy::generate(spec, hub_count),
        Topology::Layered { layer_count } => super::layered::generate(spec, layer_count),
        Topology::DenseSubsystem { dense_node_count } => {
            super::dense::generate(spec, dense_node_count)
        }
    };

    Ok(dataset)
}
