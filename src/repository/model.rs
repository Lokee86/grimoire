use std::fmt;

use crate::storage::StableHasher;

/// A stable identity for a repository node.
#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct NodeKey(pub u64);

impl NodeKey {
    /// Hashes a repository-relative path after normalizing its separators.
    pub fn from_path(path: &str) -> Result<Self, RepositoryPathError> {
        Ok(Self::from_identity(normalize_repository_path(path)?))
    }

    /// Hashes an arbitrary stable node identity.
    pub fn from_identity(identity: impl AsRef<[u8]>) -> Self {
        Self(stable_hash(identity.as_ref()))
    }

    pub const fn from_u64(value: u64) -> Self {
        Self(value)
    }

    pub const fn as_u64(self) -> u64 {
        self.0
    }
}

/// A stable identity for source content.
#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct ContentId(pub u64);

impl ContentId {
    pub fn from_bytes(bytes: &[u8]) -> Self {
        Self(stable_hash(bytes))
    }

    pub const fn from_u64(value: u64) -> Self {
        Self(value)
    }

    pub const fn as_u64(self) -> u64 {
        self.0
    }
}

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

/// A language-neutral kind of repository node.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub enum NodeKind {
    Repository,
    Directory,
    File,
    Module,
    Namespace,
    Symbol,
    Type,
    Function,
    Method,
    Field,
    Variable,
    Constant,
    Import,
    Test,
}

impl NodeKind {
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            Self::Repository => "repository",
            Self::Directory => "directory",
            Self::File => "file",
            Self::Module => "module",
            Self::Namespace => "namespace",
            Self::Symbol => "symbol",
            Self::Type => "type",
            Self::Function => "function",
            Self::Method => "method",
            Self::Field => "field",
            Self::Variable => "variable",
            Self::Constant => "constant",
            Self::Import => "import",
            Self::Test => "test",
        }
    }

    pub(crate) fn parse(value: &str) -> Option<Self> {
        Some(match value {
            "repository" => Self::Repository,
            "directory" => Self::Directory,
            "file" => Self::File,
            "module" => Self::Module,
            "namespace" => Self::Namespace,
            "symbol" => Self::Symbol,
            "type" => Self::Type,
            "function" => Self::Function,
            "method" => Self::Method,
            "field" => Self::Field,
            "variable" => Self::Variable,
            "constant" => Self::Constant,
            "import" => Self::Import,
            "test" => Self::Test,
            _ => return None,
        })
    }
}

/// A language-neutral kind of relationship between repository nodes.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub enum RelationKind {
    Contains,
    Defines,
    References,
    Imports,
    Calls,
    Implements,
    Extends,
    Includes,
    DependsOn,
    Tests,
    Documents,
    Generates,
}

impl RelationKind {
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            Self::Contains => "contains",
            Self::Defines => "defines",
            Self::References => "references",
            Self::Imports => "imports",
            Self::Calls => "calls",
            Self::Implements => "implements",
            Self::Extends => "extends",
            Self::Includes => "includes",
            Self::DependsOn => "depends-on",
            Self::Tests => "tests",
            Self::Documents => "documents",
            Self::Generates => "generates",
        }
    }

    pub(crate) fn parse(value: &str) -> Option<Self> {
        Some(match value {
            "contains" => Self::Contains,
            "defines" => Self::Defines,
            "references" => Self::References,
            "imports" => Self::Imports,
            "calls" => Self::Calls,
            "implements" => Self::Implements,
            "extends" => Self::Extends,
            "includes" => Self::Includes,
            "depends-on" => Self::DependsOn,
            "tests" => Self::Tests,
            "documents" => Self::Documents,
            "generates" => Self::Generates,
            _ => return None,
        })
    }
}

/// A source location associated with a repository fact.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct SourceSpan {
    pub path: String,
    pub start_line: u32,
    pub start_column: u32,
    pub end_line: u32,
    pub end_column: u32,
}

impl SourceSpan {
    pub fn new(
        path: impl AsRef<str>,
        start_line: u32,
        start_column: u32,
        end_line: u32,
        end_column: u32,
    ) -> Result<Self, RepositoryPathError> {
        Ok(Self {
            path: normalize_repository_path(path.as_ref())?,
            start_line,
            start_column,
            end_line,
            end_column,
        })
    }
}

/// A fact describing one repository node.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct NodeFact {
    pub key: NodeKey,
    pub kind: NodeKind,
    pub path: String,
    pub name: String,
    pub content_id: Option<ContentId>,
    pub span: Option<SourceSpan>,
}

/// A fact describing one directed relationship.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct EdgeFact {
    pub source: NodeKey,
    pub target: NodeKey,
    pub relation: RelationKind,
    pub span: Option<SourceSpan>,
}

fn stable_hash(bytes: &[u8]) -> u64 {
    let mut hasher = StableHasher::new();
    hasher.update(bytes);
    hasher.finish()
}

fn has_drive_prefix(path: &str) -> bool {
    let bytes = path.as_bytes();
    bytes.len() >= 2 && bytes[0].is_ascii_alphabetic() && bytes[1] == b':'
}
