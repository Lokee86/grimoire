use std::fmt;

/// Repository-relative path normalization failures.
#[derive(Clone, Debug, Eq, PartialEq)]
pub enum RepositoryPathError {
    Empty,
    Absolute,
    ParentComponent,
    InvalidComponent,
}

impl fmt::Display for RepositoryPathError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(match self {
            Self::Empty => "repository path is empty",
            Self::Absolute => "repository path must be relative",
            Self::ParentComponent => "repository path cannot contain a parent component",
            Self::InvalidComponent => "repository path contains an invalid component",
        })
    }
}

impl std::error::Error for RepositoryPathError {}

/// Converts a repository path to a deterministic, slash-separated UTF-8 path.
pub fn normalize_repository_path(path: &str) -> Result<String, RepositoryPathError> {
    if path.is_empty() || path.contains('\0') {
        return Err(if path.is_empty() {
            RepositoryPathError::Empty
        } else {
            RepositoryPathError::InvalidComponent
        });
    }

    let path = path.replace('\\', "/");
    if path.starts_with('/') || path.starts_with("//") || has_drive_prefix(&path) {
        return Err(RepositoryPathError::Absolute);
    }

    let mut components = Vec::new();
    for component in path.split('/') {
        match component {
            "" | "." => {}
            ".." => return Err(RepositoryPathError::ParentComponent),
            component if component.chars().any(char::is_control) => {
                return Err(RepositoryPathError::InvalidComponent);
            }
            component => components.push(component),
        }
    }

    if components.is_empty() {
        return Err(RepositoryPathError::Empty);
    }
    Ok(components.join("/"))
}

fn has_drive_prefix(path: &str) -> bool {
    let bytes = path.as_bytes();
    bytes.len() >= 2 && bytes[0].is_ascii_alphabetic() && bytes[1] == b':'
}
