//! Optional semantic indexing over deterministic Arcana graph facts.

mod build;
mod client;
mod documents;
mod error;
mod http;
mod index;
mod search;

pub use build::{BuildSummary, build_current_index};
pub use client::{
    DEFAULT_DIMENSIONS, DEFAULT_ENDPOINT, DEFAULT_IDENTITY, DEFAULT_MODEL, Embedder,
    EmbeddingClient,
};
pub use documents::{GraphDocument, graph_documents};
pub use error::EmbeddingError;
pub use http::HttpError;
pub use index::{IndexManifest, SearchHit, VectorIndexError, current_index_directory};
pub use search::search_current_index;

#[cfg(test)]
#[path = "index_tests.rs"]
mod tests;
