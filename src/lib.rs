//! Reusable library boundary for ArcanaGraph.
//!
//! Synthetic workloads and future graph-storage implementations are exposed
//! from this crate rather than coupled to the `arcana` command-line binary.

pub mod benchmark;
pub mod snapshot;
pub mod storage;
pub mod synthetic;

/// Product name presented by the ArcanaGraph library and CLI.
pub const PROJECT_NAME: &str = "ArcanaGraph";

/// Package version supplied by Cargo.
pub const PROJECT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Returns the short project description used by integrations.
pub const fn about() -> &'static str {
    "independent repository-graph foundation"
}

#[cfg(test)]
mod tests {
    use super::{PROJECT_NAME, PROJECT_VERSION, about};

    #[test]
    fn exposes_stable_project_metadata() {
        assert_eq!(PROJECT_NAME, "ArcanaGraph");
        assert!(!PROJECT_VERSION.is_empty());
        assert_eq!(about(), "independent repository-graph foundation");
    }
}
