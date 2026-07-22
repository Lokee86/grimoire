use std::{fmt, io};

#[derive(Debug)]
pub enum Error {
    Io(io::Error),
    InvalidFormat(String),
    InvalidInput(String),
    MissingObject(String),
}

pub type Result<T> = std::result::Result<T, Error>;

impl fmt::Display for Error {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => write!(formatter, "{error}"),
            Self::InvalidFormat(message) => write!(formatter, "invalid vector data: {message}"),
            Self::InvalidInput(message) => write!(formatter, "invalid vector input: {message}"),
            Self::MissingObject(source) => {
                write!(formatter, "missing vector object for source {source}")
            }
        }
    }
}

impl std::error::Error for Error {}

impl From<io::Error> for Error {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}

impl From<serde_json::Error> for Error {
    fn from(error: serde_json::Error) -> Self {
        Self::InvalidInput(error.to_string())
    }
}
