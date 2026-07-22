//! Machine-readable JSON Lines query protocol for repository snapshots.

mod diff;
mod error;
mod queries;
mod request;
mod response;
mod server;
mod session;
mod stats;

#[cfg(test)]
mod tests;

pub use error::ProtocolError;
pub use server::serve_jsonl;
pub use session::ProtocolSnapshot;

/// Stable protocol identifier emitted in every response.
pub const PROTOCOL_ID: &str = "arcana.query.v1";
