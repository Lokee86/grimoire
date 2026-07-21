use std::fmt;
use std::io;

/// An error while serializing a benchmark report.
#[derive(Debug)]
pub enum BenchmarkError {
    Io(io::Error),
}

impl fmt::Display for BenchmarkError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
        }
    }
}

impl std::error::Error for BenchmarkError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
        }
    }
}

impl From<io::Error> for BenchmarkError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}
