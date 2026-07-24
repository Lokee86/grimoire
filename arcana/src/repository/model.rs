use crate::storage::StableHasher;

use super::path::{RepositoryPathError, normalize_repository_path};

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
    Interface,
    Trait,
    Function,
    Method,
    Constructor,
    Field,
    Variable,
    Constant,
    Signal,
    Parameter,
    Import,
    Export,
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
            Self::Interface => "interface",
            Self::Trait => "trait",
            Self::Function => "function",
            Self::Method => "method",
            Self::Constructor => "constructor",
            Self::Field => "field",
            Self::Variable => "variable",
            Self::Constant => "constant",
            Self::Signal => "signal",
            Self::Parameter => "parameter",
            Self::Import => "import",
            Self::Export => "export",
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
            "interface" => Self::Interface,
            "trait" => Self::Trait,
            "function" => Self::Function,
            "method" => Self::Method,
            "constructor" => Self::Constructor,
            "field" => Self::Field,
            "variable" => Self::Variable,
            "constant" => Self::Constant,
            "signal" => Self::Signal,
            "parameter" => Self::Parameter,
            "import" => Self::Import,
            "export" => Self::Export,
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
    PossibleCalls,
    ConvertsTo,
    Implements,
    Extends,
    UsesTrait,
    Overrides,
    Reads,
    Writes,
    Annotates,
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
            Self::PossibleCalls => "possible-calls",
            Self::ConvertsTo => "converts-to",
            Self::Implements => "implements",
            Self::Extends => "extends",
            Self::UsesTrait => "uses-trait",
            Self::Overrides => "overrides",
            Self::Reads => "reads",
            Self::Writes => "writes",
            Self::Annotates => "annotates",
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
            "possible-calls" => Self::PossibleCalls,
            "converts-to" => Self::ConvertsTo,
            "implements" => Self::Implements,
            "extends" => Self::Extends,
            "uses-trait" => Self::UsesTrait,
            "overrides" => Self::Overrides,
            "reads" => Self::Reads,
            "writes" => Self::Writes,
            "annotates" => Self::Annotates,
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
    /// Durable identity supplied by an external fact producer such as Lexicon.
    pub external_identity: Option<String>,
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
