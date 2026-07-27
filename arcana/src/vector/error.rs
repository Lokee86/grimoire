use std::fmt;

use super::http::HttpError;

#[derive(Debug)]
pub enum EmbeddingError {
    Http(HttpError),
    Json(serde_json::Error),
    InvalidInput,
    Service(String),
    WrongVectorCount { expected: usize, found: usize },
    InvalidIndex(usize),
    WrongDimensions { expected: usize, found: usize },
    NonFiniteVector,
    ZeroVector,
}

impl EmbeddingError {
    pub(crate) fn is_context_limit(&self) -> bool {
        match self {
            Self::Http(HttpError::Status { status: 400, body }) => {
                context_limit_message(&String::from_utf8_lossy(body))
            }
            Self::Service(message) => context_limit_message(message),
            _ => false,
        }
    }
}

fn context_limit_message(message: &str) -> bool {
    let message = message.to_ascii_lowercase();
    message.contains("context")
        && (message.contains("exceed")
            || message.contains("too large")
            || message.contains("too long"))
}

impl fmt::Display for EmbeddingError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Http(error) => error.fmt(formatter),
            Self::Json(error) => error.fmt(formatter),
            Self::InvalidInput => formatter.write_str("embedding input is empty"),
            Self::Service(message) => write!(formatter, "embedding service error: {message}"),
            Self::WrongVectorCount { expected, found } => write!(
                formatter,
                "embedding service returned {found} vectors; expected {expected}"
            ),
            Self::InvalidIndex(index) => write!(
                formatter,
                "embedding service returned invalid index {index}"
            ),
            Self::WrongDimensions { expected, found } => write!(
                formatter,
                "embedding service returned {found} dimensions; need at least {expected}"
            ),
            Self::NonFiniteVector => formatter.write_str("embedding contains a non-finite value"),
            Self::ZeroVector => formatter.write_str("embedding has zero norm"),
        }
    }
}

impl std::error::Error for EmbeddingError {}

impl From<HttpError> for EmbeddingError {
    fn from(error: HttpError) -> Self {
        Self::Http(error)
    }
}

impl From<serde_json::Error> for EmbeddingError {
    fn from(error: serde_json::Error) -> Self {
        Self::Json(error)
    }
}
