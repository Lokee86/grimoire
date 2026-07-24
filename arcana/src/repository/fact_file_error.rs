use std::fmt;

/// An error in the deterministic repository fact file.
#[derive(Clone, Debug, Eq, PartialEq)]
pub enum FactFileError {
    InvalidHeader,
    MalformedLine { line: usize },
    UnknownRecord { line: usize },
    InvalidNumber { line: usize },
    InvalidKind { line: usize },
    InvalidRelation { line: usize },
    InvalidReason { line: usize },
    InvalidEscape { line: usize },
    InvalidSpan { line: usize },
}

impl fmt::Display for FactFileError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::InvalidHeader => formatter.write_str("repository fact header is invalid"),
            Self::MalformedLine { line } => {
                write!(formatter, "repository fact line {line} is malformed")
            }
            Self::UnknownRecord { line } => write!(
                formatter,
                "repository fact line {line} has an unknown record"
            ),
            Self::InvalidNumber { line } => write!(
                formatter,
                "repository fact line {line} has an invalid number"
            ),
            Self::InvalidKind { line } => write!(
                formatter,
                "repository fact line {line} has an invalid node kind"
            ),
            Self::InvalidRelation { line } => write!(
                formatter,
                "repository fact line {line} has an invalid relation"
            ),
            Self::InvalidReason { line } => write!(
                formatter,
                "repository fact line {line} has an invalid unresolved reason"
            ),
            Self::InvalidEscape { line } => write!(
                formatter,
                "repository fact line {line} has an invalid escape"
            ),
            Self::InvalidSpan { line } => write!(
                formatter,
                "repository fact line {line} has an invalid source span"
            ),
        }
    }
}

impl std::error::Error for FactFileError {}
