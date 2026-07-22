use std::path::Path;

use crate::synthetic::NodeId;

use super::{NodeFact, NodeKey, RepositoryPathError, normalize_repository_path};

#[path = "catalogue_file.rs"]
mod catalogue_file;

pub use catalogue_file::CatalogueError;

/// One dense graph node and its repository metadata.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CatalogueEntry {
    pub node_id: NodeId,
    pub fact: NodeFact,
}

/// An immutable, validated index of compiled repository node metadata.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct RepositoryCatalogue {
    entries: Vec<CatalogueEntry>,
}

impl RepositoryCatalogue {
    pub fn new(mut entries: Vec<CatalogueEntry>) -> Result<Self, CatalogueError> {
        for entry in &entries {
            validate_fact(&entry.fact)?;
        }
        entries.sort_unstable_by_key(|entry| entry.node_id);
        for (index, entry) in entries.iter().enumerate() {
            let expected = u32::try_from(index).map_err(|_| CatalogueError::NodeIdOverflow)?;
            if entry.node_id.0 != expected {
                return Err(CatalogueError::NonDenseNodeId {
                    expected,
                    found: entry.node_id.0,
                });
            }
            if index > 0 && entries[index - 1].fact.key == entry.fact.key {
                return Err(CatalogueError::DuplicateNodeKey {
                    key: entry.fact.key,
                });
            }
        }
        Ok(Self { entries })
    }

    pub fn entries(&self) -> &[CatalogueEntry] {
        &self.entries
    }

    pub fn len(&self) -> usize {
        self.entries.len()
    }

    pub fn is_empty(&self) -> bool {
        self.entries.is_empty()
    }

    pub fn lookup_by_key(&self, key: NodeKey) -> Option<&CatalogueEntry> {
        self.entries.iter().find(|entry| entry.fact.key == key)
    }

    pub fn lookup_by_path(&self, path: &str) -> Result<Vec<&CatalogueEntry>, CatalogueError> {
        let path = normalize_repository_path(path).map_err(CatalogueError::InvalidPath)?;
        Ok(self
            .entries
            .iter()
            .filter(|entry| entry.fact.path == path)
            .collect())
    }

    pub fn lookup_by_name(&self, name: &str) -> Vec<&CatalogueEntry> {
        self.entries
            .iter()
            .filter(|entry| entry.fact.name == name)
            .collect()
    }

    pub fn encode(&self) -> Result<String, CatalogueError> {
        catalogue_file::encode(self)
    }

    pub fn decode(input: &str) -> Result<Self, CatalogueError> {
        catalogue_file::decode(input)
    }

    pub fn write(&self, path: impl AsRef<Path>) -> Result<(), CatalogueError> {
        catalogue_file::write(path, self)
    }

    pub fn read(path: impl AsRef<Path>) -> Result<Self, CatalogueError> {
        catalogue_file::read(path)
    }
}

pub fn write_catalogue(
    path: impl AsRef<Path>,
    catalogue: &RepositoryCatalogue,
) -> Result<(), CatalogueError> {
    catalogue.write(path)
}

pub fn read_catalogue(path: impl AsRef<Path>) -> Result<RepositoryCatalogue, CatalogueError> {
    RepositoryCatalogue::read(path)
}

fn validate_fact(fact: &NodeFact) -> Result<(), CatalogueError> {
    let normalized = normalize_repository_path(&fact.path).map_err(CatalogueError::InvalidPath)?;
    if normalized != fact.path {
        return Err(CatalogueError::InvalidPath(
            RepositoryPathError::InvalidComponent,
        ));
    }
    if let Some(span) = &fact.span {
        let normalized =
            normalize_repository_path(&span.path).map_err(CatalogueError::InvalidPath)?;
        if normalized != span.path {
            return Err(CatalogueError::InvalidPath(
                RepositoryPathError::InvalidComponent,
            ));
        }
    }
    Ok(())
}
