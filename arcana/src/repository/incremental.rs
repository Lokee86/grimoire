use std::collections::BTreeSet;
use std::fmt;

use crate::snapshot::OverlayChanges;
use crate::synthetic::{Edge, GraphDataset};

use super::{
    CompiledRepository, FactOwnershipError, NodeKey, RepositoryCompileError, RepositoryFacts,
    compile_repository_facts, replace_changed_files,
};

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct IncrementalUpdate {
    pub facts: RepositoryFacts,
    pub compiled: CompiledRepository,
    pub changes: OverlayChanges,
    changed_file_count: usize,
}

impl IncrementalUpdate {
    pub const fn changed_file_count(&self) -> usize {
        self.changed_file_count
    }
}

#[derive(Debug)]
pub enum IncrementalError {
    Ownership(FactOwnershipError),
    Compile(RepositoryCompileError),
    NodeSetChanged {
        added: Vec<NodeKey>,
        removed: Vec<NodeKey>,
    },
    BaseNodeCountMismatch {
        expected: u32,
        actual: u32,
    },
}

impl fmt::Display for IncrementalError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Ownership(error) => error.fmt(formatter),
            Self::Compile(error) => error.fmt(formatter),
            Self::NodeSetChanged { added, removed } => write!(
                formatter,
                "incremental graph update requires rebuild: {} node(s) added, {} removed",
                added.len(),
                removed.len()
            ),
            Self::BaseNodeCountMismatch { expected, actual } => write!(
                formatter,
                "packed base has {actual} nodes but updated repository requires {expected}"
            ),
        }
    }
}

impl std::error::Error for IncrementalError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Ownership(error) => Some(error),
            Self::Compile(error) => Some(error),
            _ => None,
        }
    }
}

impl From<FactOwnershipError> for IncrementalError {
    fn from(error: FactOwnershipError) -> Self {
        Self::Ownership(error)
    }
}

impl From<RepositoryCompileError> for IncrementalError {
    fn from(error: RepositoryCompileError) -> Self {
        Self::Compile(error)
    }
}

pub fn plan_file_update(
    current_facts: &RepositoryFacts,
    replacement_facts: &RepositoryFacts,
    changed_paths: &[String],
    packed_base: &GraphDataset,
) -> Result<IncrementalUpdate, IncrementalError> {
    let facts = replace_changed_files(current_facts, replacement_facts, changed_paths)?;
    let current = compile_repository_facts(current_facts)?;
    let compiled = compile_repository_facts(&facts)?;
    if current.node_ids != compiled.node_ids {
        let current_keys = current.node_ids.keys().copied().collect::<BTreeSet<_>>();
        let updated_keys = compiled.node_ids.keys().copied().collect::<BTreeSet<_>>();
        return Err(IncrementalError::NodeSetChanged {
            added: updated_keys.difference(&current_keys).copied().collect(),
            removed: current_keys.difference(&updated_keys).copied().collect(),
        });
    }
    if packed_base.node_count != compiled.dataset.node_count {
        return Err(IncrementalError::BaseNodeCountMismatch {
            expected: compiled.dataset.node_count,
            actual: packed_base.node_count,
        });
    }

    let base = packed_base
        .edges
        .iter()
        .copied()
        .collect::<BTreeSet<Edge>>();
    let visible = compiled
        .dataset
        .edges
        .iter()
        .copied()
        .collect::<BTreeSet<Edge>>();
    let changes = OverlayChanges {
        added: visible.difference(&base).copied().collect(),
        removed: base.difference(&visible).copied().collect(),
    };
    Ok(IncrementalUpdate {
        facts,
        compiled,
        changes,
        changed_file_count: changed_paths.len(),
    })
}
