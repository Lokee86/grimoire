use std::fmt;

/// The topology family used to construct a synthetic graph.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum Topology {
    Modular {
        cluster_count: u32,
        /// Desired cross-cluster edge share in basis points.
        cross_cluster_ratio: u16,
    },
    Entangled {
        cluster_count: u32,
        hub_count: u32,
    },
    HubHeavy {
        hub_count: u32,
    },
    Layered {
        layer_count: u32,
    },
    DenseSubsystem {
        dense_node_count: u32,
    },
}

/// Standard graph sizes used by the storage benchmark suite.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum ScaleTier {
    Small,
    Medium,
    Large,
    Stress,
}

impl ScaleTier {
    pub const fn counts(self) -> (u32, u64) {
        match self {
            Self::Small => (10_000, 100_000),
            Self::Medium => (100_000, 1_000_000),
            Self::Large => (1_000_000, 10_000_000),
            Self::Stress => (5_000_000, 50_000_000),
        }
    }
}

/// A deterministic synthetic graph request.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct GraphSpec {
    pub topology: Topology,
    pub node_count: u32,
    pub edge_count: u64,
    pub seed: u64,
}

impl GraphSpec {
    pub const fn for_tier(topology: Topology, tier: ScaleTier, seed: u64) -> Self {
        let (node_count, edge_count) = tier.counts();
        Self {
            topology,
            node_count,
            edge_count,
            seed,
        }
    }

    /// Validates counts and topology parameters before allocating edge storage.
    pub fn validate(&self) -> Result<(), GraphSpecError> {
        if self.node_count == 0 {
            return Err(GraphSpecError::ZeroNodes);
        }

        if self.node_count == 1 && self.edge_count > 0 {
            return Err(GraphSpecError::SelfEdgesOnly {
                node_count: self.node_count,
                edge_count: self.edge_count,
            });
        }

        let capacity = u64::from(self.node_count) * u64::from(self.node_count - 1);
        if self.edge_count > capacity {
            return Err(GraphSpecError::EdgeCountExceedsCapacity {
                edge_count: self.edge_count,
                capacity,
            });
        }

        if usize::try_from(self.edge_count).is_err() {
            return Err(GraphSpecError::EdgeCountExceedsPlatform {
                edge_count: self.edge_count,
            });
        }

        match self.topology {
            Topology::Modular {
                cluster_count,
                cross_cluster_ratio,
            } => {
                validate_partition("cluster_count", cluster_count, self.node_count)?;
                if cross_cluster_ratio > 10_000 {
                    return Err(GraphSpecError::InvalidBasisPointRatio {
                        field: "cross_cluster_ratio",
                        ratio: cross_cluster_ratio,
                    });
                }
            }
            Topology::Entangled {
                cluster_count,
                hub_count,
            } => {
                validate_partition("cluster_count", cluster_count, self.node_count)?;
                if cluster_count < 2 {
                    return Err(GraphSpecError::PartitionTooSmall {
                        field: "cluster_count",
                        value: cluster_count,
                        minimum: 2,
                    });
                }
                validate_partition("hub_count", hub_count, self.node_count)?;
            }
            Topology::HubHeavy { hub_count } => {
                validate_partition("hub_count", hub_count, self.node_count)?;
            }
            Topology::Layered { layer_count } => {
                validate_partition("layer_count", layer_count, self.node_count)?;
                if layer_count < 2 {
                    return Err(GraphSpecError::PartitionTooSmall {
                        field: "layer_count",
                        value: layer_count,
                        minimum: 2,
                    });
                }
            }
            Topology::DenseSubsystem { dense_node_count } => {
                validate_partition("dense_node_count", dense_node_count, self.node_count)?;
                if dense_node_count < 2 {
                    return Err(GraphSpecError::PartitionTooSmall {
                        field: "dense_node_count",
                        value: dense_node_count,
                        minimum: 2,
                    });
                }
            }
        }

        Ok(())
    }
}

fn validate_partition(
    field: &'static str,
    value: u32,
    node_count: u32,
) -> Result<(), GraphSpecError> {
    if value == 0 {
        return Err(GraphSpecError::ZeroPartition { field });
    }
    if value > node_count {
        return Err(GraphSpecError::PartitionExceedsNodes {
            field,
            value,
            node_count,
        });
    }
    Ok(())
}

/// A graph specification that cannot be generated.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum GraphSpecError {
    ZeroNodes,
    ZeroPartition {
        field: &'static str,
    },
    PartitionTooSmall {
        field: &'static str,
        value: u32,
        minimum: u32,
    },
    PartitionExceedsNodes {
        field: &'static str,
        value: u32,
        node_count: u32,
    },
    InvalidBasisPointRatio {
        field: &'static str,
        ratio: u16,
    },
    SelfEdgesOnly {
        node_count: u32,
        edge_count: u64,
    },
    EdgeCountExceedsCapacity {
        edge_count: u64,
        capacity: u64,
    },
    EdgeCountExceedsPlatform {
        edge_count: u64,
    },
}

impl fmt::Display for GraphSpecError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::ZeroNodes => formatter.write_str("node_count must be greater than zero"),
            Self::ZeroPartition { field } => write!(formatter, "{field} must be greater than zero"),
            Self::PartitionTooSmall {
                field,
                value,
                minimum,
            } => write!(formatter, "{field} ({value}) must be at least {minimum}"),
            Self::PartitionExceedsNodes {
                field,
                value,
                node_count,
            } => write!(
                formatter,
                "{field} ({value}) cannot exceed node_count ({node_count})"
            ),
            Self::InvalidBasisPointRatio { field, ratio } => write!(
                formatter,
                "{field} ({ratio}) must be between 0 and 10,000 basis points"
            ),
            Self::SelfEdgesOnly {
                node_count,
                edge_count,
            } => write!(
                formatter,
                "node_count ({node_count}) has only self-edges, so edge_count ({edge_count}) must be zero"
            ),
            Self::EdgeCountExceedsCapacity {
                edge_count,
                capacity,
            } => write!(
                formatter,
                "edge_count ({edge_count}) exceeds the unique directed non-self edge capacity ({capacity})"
            ),
            Self::EdgeCountExceedsPlatform { edge_count } => write!(
                formatter,
                "edge_count ({edge_count}) cannot fit in this platform's address space"
            ),
        }
    }
}

impl std::error::Error for GraphSpecError {}
