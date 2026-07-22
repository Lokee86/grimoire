use std::fmt;
use std::fs::{self, OpenOptions};
use std::io::{self, Write};
use std::path::Path;

use crate::synthetic::NodeId;

use super::{CatalogueEntry, RepositoryCatalogue};
use crate::repository::{
    ContentId, NodeFact, NodeKey, NodeKind, RepositoryPathError, SourceSpan,
    normalize_repository_path,
};

const HEADER: &str = "version\t1";

pub(super) fn encode(catalogue: &RepositoryCatalogue) -> Result<String, CatalogueError> {
    let validated = RepositoryCatalogue::new(catalogue.entries().to_vec())?;
    let mut output = String::from(HEADER);
    output.push('\n');
    for entry in validated.entries() {
        output.push_str("N\t");
        push_field(&mut output, &entry.node_id.0.to_string());
        output.push('\t');
        push_field(&mut output, &format_id(entry.fact.key.0));
        output.push('\t');
        push_field(&mut output, entry.fact.kind.as_str());
        output.push('\t');
        push_field(&mut output, &entry.fact.path);
        output.push('\t');
        push_field(&mut output, &entry.fact.name);
        output.push('\t');
        let content_id = entry
            .fact
            .content_id
            .map_or_else(|| "-".to_owned(), |id| format_id(id.0));
        push_field(&mut output, &content_id);
        push_span(&mut output, entry.fact.span.as_ref());
        output.push('\n');
    }
    Ok(output)
}

pub(super) fn decode(input: &str) -> Result<RepositoryCatalogue, CatalogueError> {
    if !input.ends_with('\n') {
        return Err(CatalogueError::MissingFinalNewline);
    }
    let mut lines = input[..input.len() - 1].split('\n');
    if lines.next() != Some(HEADER) {
        return Err(CatalogueError::InvalidHeader);
    }

    let mut entries = Vec::new();
    let mut previous_id = None;
    for (index, line) in lines.enumerate() {
        let line_number = index + 2;
        let fields = line
            .split('\t')
            .map(|field| unescape(field, line_number))
            .collect::<Result<Vec<_>, _>>()?;
        if fields.len() != 12 || fields.first().map(String::as_str) != Some("N") {
            return Err(CatalogueError::MalformedLine { line: line_number });
        }
        let node_id = parse_decimal(&fields[1], line_number)?;
        if previous_id.is_some_and(|previous| node_id <= previous) {
            return Err(CatalogueError::InvalidOrder { line: line_number });
        }
        previous_id = Some(node_id);
        let fact = NodeFact {
            key: NodeKey::from_u64(parse_hex(&fields[2], line_number)?),
            kind: NodeKind::parse(&fields[3])
                .ok_or(CatalogueError::InvalidKind { line: line_number })?,
            path: parse_path(&fields[4], line_number)?,
            name: fields[5].clone(),
            content_id: parse_optional_id(&fields[6], line_number)?.map(ContentId::from_u64),
            span: parse_span(&fields[7..12], line_number)?,
        };
        entries.push(CatalogueEntry {
            node_id: NodeId(node_id),
            fact,
        });
    }
    RepositoryCatalogue::new(entries)
}

pub(super) fn write(
    path: impl AsRef<Path>,
    catalogue: &RepositoryCatalogue,
) -> Result<(), CatalogueError> {
    let encoded = encode(catalogue)?;
    let mut file = OpenOptions::new().write(true).create_new(true).open(path)?;
    file.write_all(encoded.as_bytes())?;
    file.sync_all()?;
    Ok(())
}

pub(super) fn read(path: impl AsRef<Path>) -> Result<RepositoryCatalogue, CatalogueError> {
    decode(&fs::read_to_string(path)?)
}

/// A catalogue file or metadata validation failure.
#[derive(Debug)]
pub enum CatalogueError {
    Io(io::Error),
    InvalidHeader,
    MissingFinalNewline,
    MalformedLine { line: usize },
    InvalidOrder { line: usize },
    InvalidNumber { line: usize },
    InvalidKind { line: usize },
    InvalidEscape { line: usize },
    InvalidSpan { line: usize },
    InvalidPath(RepositoryPathError),
    NonCanonicalPath { line: usize },
    DuplicateNodeKey { key: NodeKey },
    NonDenseNodeId { expected: u32, found: u32 },
    NodeIdOverflow,
}

impl fmt::Display for CatalogueError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::InvalidHeader => formatter.write_str("catalogue header is invalid"),
            Self::MissingFinalNewline => {
                formatter.write_str("catalogue is missing its final newline")
            }
            Self::MalformedLine { line } => write!(formatter, "catalogue line {line} is malformed"),
            Self::InvalidOrder { line } => {
                write!(formatter, "catalogue line {line} is out of order")
            }
            Self::InvalidNumber { line } => {
                write!(formatter, "catalogue line {line} has an invalid number")
            }
            Self::InvalidKind { line } => {
                write!(formatter, "catalogue line {line} has an invalid node kind")
            }
            Self::InvalidEscape { line } => {
                write!(formatter, "catalogue line {line} has an invalid escape")
            }
            Self::InvalidSpan { line } => {
                write!(formatter, "catalogue line {line} has an invalid span")
            }
            Self::InvalidPath(error) => error.fmt(formatter),
            Self::NonCanonicalPath { line } => {
                write!(formatter, "catalogue line {line} has a non-canonical path")
            }
            Self::DuplicateNodeKey { key } => {
                write!(formatter, "catalogue contains duplicate node key {key:?}")
            }
            Self::NonDenseNodeId { expected, found } => write!(
                formatter,
                "catalogue expected node id {expected} but found {found}"
            ),
            Self::NodeIdOverflow => {
                formatter.write_str("catalogue node count exceeds u32 capacity")
            }
        }
    }
}

impl std::error::Error for CatalogueError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            Self::InvalidPath(error) => Some(error),
            _ => None,
        }
    }
}

impl From<io::Error> for CatalogueError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}

fn parse_path(value: &str, line: usize) -> Result<String, CatalogueError> {
    let normalized = normalize_repository_path(value).map_err(CatalogueError::InvalidPath)?;
    if normalized != value {
        return Err(CatalogueError::NonCanonicalPath { line });
    }
    Ok(normalized)
}

fn parse_span(fields: &[String], line: usize) -> Result<Option<SourceSpan>, CatalogueError> {
    if fields.iter().all(|field| field == "-") {
        return Ok(None);
    }
    if fields.len() != 5 || fields.iter().any(|field| field == "-") {
        return Err(CatalogueError::InvalidSpan { line });
    }
    Ok(Some(SourceSpan {
        path: parse_path(&fields[0], line)?,
        start_line: parse_u32(&fields[1], line)?,
        start_column: parse_u32(&fields[2], line)?,
        end_line: parse_u32(&fields[3], line)?,
        end_column: parse_u32(&fields[4], line)?,
    }))
}

fn parse_optional_id(value: &str, line: usize) -> Result<Option<u64>, CatalogueError> {
    if value == "-" {
        Ok(None)
    } else {
        parse_hex(value, line).map(Some)
    }
}

fn parse_decimal(value: &str, line: usize) -> Result<u32, CatalogueError> {
    if value.is_empty()
        || (value.len() > 1 && value.starts_with('0'))
        || !value.bytes().all(|b| b.is_ascii_digit())
    {
        return Err(CatalogueError::InvalidNumber { line });
    }
    value
        .parse()
        .map_err(|_| CatalogueError::InvalidNumber { line })
}

fn parse_u32(value: &str, line: usize) -> Result<u32, CatalogueError> {
    value
        .parse()
        .map_err(|_| CatalogueError::InvalidNumber { line })
}

fn parse_hex(value: &str, line: usize) -> Result<u64, CatalogueError> {
    if value.len() != 16
        || !value
            .bytes()
            .all(|b| b.is_ascii_hexdigit() && !b.is_ascii_uppercase())
    {
        return Err(CatalogueError::InvalidNumber { line });
    }
    u64::from_str_radix(value, 16).map_err(|_| CatalogueError::InvalidNumber { line })
}

fn push_span(output: &mut String, span: Option<&SourceSpan>) {
    if let Some(span) = span {
        for field in [
            span.path.clone(),
            span.start_line.to_string(),
            span.start_column.to_string(),
            span.end_line.to_string(),
            span.end_column.to_string(),
        ] {
            output.push('\t');
            push_field(output, &field);
        }
    } else {
        output.push_str("\t-\t-\t-\t-\t-");
    }
}

fn push_field(output: &mut String, value: &str) {
    for character in value.chars() {
        match character {
            '\\' => output.push_str("\\\\"),
            '\t' => output.push_str("\\t"),
            '\n' => output.push_str("\\n"),
            '\r' => output.push_str("\\r"),
            '\0' => output.push_str("\\0"),
            character => output.push(character),
        }
    }
}

fn unescape(value: &str, line: usize) -> Result<String, CatalogueError> {
    let mut output = String::with_capacity(value.len());
    let mut escaped = false;
    for character in value.chars() {
        if escaped {
            output.push(match character {
                '\\' => '\\',
                't' => '\t',
                'n' => '\n',
                'r' => '\r',
                '0' => '\0',
                _ => return Err(CatalogueError::InvalidEscape { line }),
            });
            escaped = false;
        } else if character == '\\' {
            escaped = true;
        } else {
            output.push(character);
        }
    }
    if escaped {
        return Err(CatalogueError::InvalidEscape { line });
    }
    Ok(output)
}

fn format_id(value: u64) -> String {
    format!("{value:016x}")
}
