use std::fmt;
use std::path::PathBuf;

use arcana_graph::repository::RelationKind;

pub const USAGE: &str = "Usage: arcana [OPTIONS] [COMMAND]\n\nOptions:\n    -h, --help       Print this help message\n    -V, --version    Print version information\n\nCommands:\n    benchmark        Compare overlays with packed snapshot rebuilds\n    import-facts     Compile a fact TSV into a packed graph and catalogue\n    query            Query exact node names from a packed graph\n\nImport facts:\n    arcana import-facts --facts <FILE> --output <NEW-DIRECTORY>\n\nQuery:\n    arcana query --graph <FILE> --catalogue <FILE> --name <EXACT-NAME> [--reverse] [--relation <RELATION>]";

#[derive(Debug)]
pub enum Command {
    Help,
    Version,
    Benchmark(Vec<String>),
    ImportFacts(ImportFactsCommand),
    Query(QueryCommand),
}

#[derive(Debug)]
pub struct ImportFactsCommand {
    pub facts: PathBuf,
    pub output: PathBuf,
}

#[derive(Debug)]
pub struct QueryCommand {
    pub graph: PathBuf,
    pub catalogue: PathBuf,
    pub name: String,
    pub reverse: bool,
    pub relation: Option<RelationKind>,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum CliParseError {
    MissingValue(String),
    MissingRequired(&'static str),
    UnknownFlag(String),
    UnexpectedArgument(String),
    InvalidRelation(String),
}

impl fmt::Display for CliParseError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::MissingValue(option) => write!(formatter, "missing value for {option}"),
            Self::MissingRequired(option) => write!(formatter, "missing required option {option}"),
            Self::UnknownFlag(flag) => write!(formatter, "unknown flag '{flag}'"),
            Self::UnexpectedArgument(argument) => {
                write!(formatter, "unexpected argument '{argument}'")
            }
            Self::InvalidRelation(relation) => write!(formatter, "unknown relation '{relation}'"),
        }
    }
}

impl std::error::Error for CliParseError {}

pub fn parse(arguments: impl IntoIterator<Item = String>) -> Result<Command, CliParseError> {
    let mut arguments = arguments.into_iter();
    let Some(command) = arguments.next() else {
        return Ok(Command::Help);
    };
    let rest = arguments.collect::<Vec<_>>();
    match command.as_str() {
        "-h" | "--help" if rest.is_empty() => Ok(Command::Help),
        "-V" | "--version" if rest.is_empty() => Ok(Command::Version),
        "benchmark" => Ok(Command::Benchmark(rest)),
        "import-facts" => Ok(Command::ImportFacts(parse_import_facts(rest)?)),
        "query" => Ok(Command::Query(parse_query(rest)?)),
        argument => Err(CliParseError::UnexpectedArgument(argument.to_owned())),
    }
}

fn parse_import_facts(arguments: Vec<String>) -> Result<ImportFactsCommand, CliParseError> {
    let mut facts = None;
    let mut output = None;
    parse_options(arguments, |option, value| match option {
        "--facts" => set_path(&mut facts, value.as_deref(), "--facts"),
        "--output" => set_path(&mut output, value.as_deref(), "--output"),
        option => Err(CliParseError::UnknownFlag(option.to_owned())),
    })?;
    Ok(ImportFactsCommand {
        facts: facts.ok_or(CliParseError::MissingRequired("--facts"))?,
        output: output.ok_or(CliParseError::MissingRequired("--output"))?,
    })
}

fn parse_query(arguments: Vec<String>) -> Result<QueryCommand, CliParseError> {
    let mut graph = None;
    let mut catalogue = None;
    let mut name = None;
    let mut reverse = false;
    let mut relation = None;
    parse_options(arguments, |option, value| match option {
        "--graph" => set_path(&mut graph, value.as_deref(), "--graph"),
        "--catalogue" => set_path(&mut catalogue, value.as_deref(), "--catalogue"),
        "--name" => set_string(&mut name, value.as_deref(), "--name"),
        "--relation" => {
            let value = value
                .as_deref()
                .ok_or(CliParseError::MissingValue("--relation".to_owned()))?;
            relation = Some(parse_relation(value)?);
            Ok(())
        }
        "--reverse" if value.is_none() => {
            reverse = true;
            Ok(())
        }
        option => Err(CliParseError::UnknownFlag(option.to_owned())),
    })?;
    Ok(QueryCommand {
        graph: graph.ok_or(CliParseError::MissingRequired("--graph"))?,
        catalogue: catalogue.ok_or(CliParseError::MissingRequired("--catalogue"))?,
        name: name.ok_or(CliParseError::MissingRequired("--name"))?,
        reverse,
        relation,
    })
}

fn parse_options<F>(arguments: Vec<String>, mut parse: F) -> Result<(), CliParseError>
where
    F: FnMut(&str, Option<String>) -> Result<(), CliParseError>,
{
    let mut arguments = arguments.into_iter();
    while let Some(argument) = arguments.next() {
        if !argument.starts_with('-') {
            return Err(CliParseError::UnexpectedArgument(argument));
        }
        let (option, inline) = argument
            .split_once('=')
            .map_or((argument.as_str(), None), |(option, value)| {
                (option, Some(value))
            });
        if option == "--reverse" {
            parse(option, inline.map(str::to_owned))?;
            continue;
        }
        let value = match inline {
            Some(value) if !value.is_empty() => Some(value.to_owned()),
            Some(_) => return Err(CliParseError::MissingValue(option.to_owned())),
            None => match arguments.next() {
                Some(value) if !value.starts_with('-') => Some(value),
                Some(value) => return Err(CliParseError::UnexpectedArgument(value)),
                None => None,
            },
        };
        parse(option, value)?;
    }
    Ok(())
}

fn set_path(
    destination: &mut Option<PathBuf>,
    value: Option<&str>,
    option: &'static str,
) -> Result<(), CliParseError> {
    let value = value.ok_or(CliParseError::MissingValue(option.to_owned()))?;
    if destination.replace(PathBuf::from(value)).is_some() {
        return Err(CliParseError::UnexpectedArgument(format!(
            "duplicate {option}"
        )));
    }
    Ok(())
}

fn set_string(
    destination: &mut Option<String>,
    value: Option<&str>,
    option: &'static str,
) -> Result<(), CliParseError> {
    let value = value.ok_or(CliParseError::MissingValue(option.to_owned()))?;
    if destination.replace(value.to_owned()).is_some() {
        return Err(CliParseError::UnexpectedArgument(format!(
            "duplicate {option}"
        )));
    }
    Ok(())
}

fn parse_relation(value: &str) -> Result<RelationKind, CliParseError> {
    match value {
        "contains" => Ok(RelationKind::Contains),
        "defines" => Ok(RelationKind::Defines),
        "references" => Ok(RelationKind::References),
        "imports" => Ok(RelationKind::Imports),
        "calls" => Ok(RelationKind::Calls),
        "implements" => Ok(RelationKind::Implements),
        "extends" => Ok(RelationKind::Extends),
        "includes" => Ok(RelationKind::Includes),
        "depends-on" => Ok(RelationKind::DependsOn),
        "tests" => Ok(RelationKind::Tests),
        "documents" => Ok(RelationKind::Documents),
        "generates" => Ok(RelationKind::Generates),
        value => Err(CliParseError::InvalidRelation(value.to_owned())),
    }
}
