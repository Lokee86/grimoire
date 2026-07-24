use std::fmt;
use std::io;

use crate::repository::{CatalogueError, FactFileError};
use crate::storage::PackedError;

/// Fatal snapshot-opening or stream-serving failure.
#[derive(Debug)]
pub enum ProtocolError {
    Io(io::Error),
    Json(serde_json::Error),
    Packed(PackedError),
    Catalogue(CatalogueError),
    Facts(FactFileError),
    InvalidSnapshot(String),
}

impl fmt::Display for ProtocolError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::Json(error) => error.fmt(formatter),
            Self::Packed(error) => error.fmt(formatter),
            Self::Catalogue(error) => error.fmt(formatter),
            Self::Facts(error) => error.fmt(formatter),
            Self::InvalidSnapshot(message) => formatter.write_str(message),
        }
    }
}

impl std::error::Error for ProtocolError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            Self::Json(error) => Some(error),
            Self::Packed(error) => Some(error),
            Self::Catalogue(error) => Some(error),
            Self::Facts(error) => Some(error),
            Self::InvalidSnapshot(_) => None,
        }
    }
}

impl From<io::Error> for ProtocolError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}

impl From<serde_json::Error> for ProtocolError {
    fn from(error: serde_json::Error) -> Self {
        Self::Json(error)
    }
}

impl From<PackedError> for ProtocolError {
    fn from(error: PackedError) -> Self {
        Self::Packed(error)
    }
}

impl From<CatalogueError> for ProtocolError {
    fn from(error: CatalogueError) -> Self {
        Self::Catalogue(error)
    }
}

impl From<FactFileError> for ProtocolError {
    fn from(error: FactFileError) -> Self {
        Self::Facts(error)
    }
}
