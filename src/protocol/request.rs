use std::path::PathBuf;

use serde::Deserialize;
use serde_json::Value;

#[derive(Debug, Deserialize)]
pub(crate) struct RequestEnvelope {
    #[serde(default)]
    pub id: Value,
    #[serde(flatten)]
    pub command: RequestCommand,
}

#[derive(Debug, Deserialize)]
#[serde(tag = "op", rename_all = "snake_case")]
pub(crate) enum RequestCommand {
    ResolveSymbol {
        name: String,
        kind: Option<String>,
        path: Option<String>,
        limit: Option<usize>,
    },
    ResolveFile {
        path: String,
        limit: Option<usize>,
    },
    ListNodes {
        kind: Option<String>,
        path_prefix: Option<String>,
        limit: Option<usize>,
    },
    Neighbors {
        node_id: u32,
        direction: QueryDirection,
        relation: Option<String>,
    },
    Unresolved {
        node_id: Option<u32>,
        path: Option<String>,
        reason: Option<String>,
        relation: Option<String>,
        limit: Option<usize>,
    },
    Stats,
    Diff {
        other_snapshot: PathBuf,
        limit: Option<usize>,
    },
}

#[derive(Clone, Copy, Debug, Deserialize)]
#[serde(rename_all = "lowercase")]
pub(crate) enum QueryDirection {
    Incoming,
    Outgoing,
}
