use std::fmt::{self, Write as FmtWrite};
use std::fs;
use std::io;
use std::path::Path;

use arcana_graph::repository::{
    CatalogueError, CompiledRepository, FactFileError, NodeFact, RelationKind, RepositoryCatalogue,
    RepositoryCompileError, RepositoryFacts, edge_kind_to_relation,
};
use arcana_graph::storage::{PackedError, PackedGraph, QueryError};
use arcana_graph::synthetic::NodeId;

use crate::cli::{ImportFactsCommand, QueryCommand};

#[derive(Debug)]
pub enum CliCommandError {
    Io(io::Error),
    Facts(FactFileError),
    Compile(RepositoryCompileError),
    Packed(PackedError),
    Query(QueryError),
    Catalogue(CatalogueError),
    UnknownEdgeKind(u16),
    MissingCatalogueNode(NodeId),
}

impl fmt::Display for CliCommandError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::Facts(error) => error.fmt(formatter),
            Self::Compile(error) => error.fmt(formatter),
            Self::Packed(error) => error.fmt(formatter),
            Self::Query(error) => error.fmt(formatter),
            Self::Catalogue(error) => error.fmt(formatter),
            Self::UnknownEdgeKind(kind) => {
                write!(formatter, "graph contains unknown edge kind {kind}")
            }
            Self::MissingCatalogueNode(node) => write!(
                formatter,
                "catalogue has no metadata for graph node {}",
                node.0
            ),
        }
    }
}

impl std::error::Error for CliCommandError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            Self::Facts(error) => Some(error),
            Self::Compile(error) => Some(error),
            Self::Packed(error) => Some(error),
            Self::Query(error) => Some(error),
            Self::Catalogue(error) => Some(error),
            Self::UnknownEdgeKind(_) | Self::MissingCatalogueNode(_) => None,
        }
    }
}

impl From<io::Error> for CliCommandError {
    fn from(error: io::Error) -> Self {
        Self::Io(error)
    }
}
impl From<FactFileError> for CliCommandError {
    fn from(error: FactFileError) -> Self {
        Self::Facts(error)
    }
}
impl From<RepositoryCompileError> for CliCommandError {
    fn from(error: RepositoryCompileError) -> Self {
        Self::Compile(error)
    }
}
impl From<PackedError> for CliCommandError {
    fn from(error: PackedError) -> Self {
        Self::Packed(error)
    }
}
impl From<QueryError> for CliCommandError {
    fn from(error: QueryError) -> Self {
        Self::Query(error)
    }
}
impl From<CatalogueError> for CliCommandError {
    fn from(error: CatalogueError) -> Self {
        Self::Catalogue(error)
    }
}

pub fn run_import_facts(command: &ImportFactsCommand) -> Result<String, CliCommandError> {
    if command.output.try_exists()? {
        return Err(io::Error::new(
            io::ErrorKind::AlreadyExists,
            format!(
                "output directory already exists: {}",
                command.output.display()
            ),
        )
        .into());
    }
    let text = fs::read_to_string(&command.facts)?;
    let facts = RepositoryFacts::parse(&text)?;
    let compiled = arcana_graph::repository::compile_repository_facts(&facts)?;
    fs::create_dir(&command.output)?;
    write_compiled(&command.output, &compiled)
}

fn write_compiled(output: &Path, compiled: &CompiledRepository) -> Result<String, CliCommandError> {
    let graph_path = output.join("graph.arcana");
    let catalogue_path = output.join("catalogue.tsv");
    arcana_graph::storage::write_packed(&graph_path, &compiled.dataset)?;
    arcana_graph::repository::write_catalogue(&catalogue_path, &compiled.catalogue)?;
    let graph_size = fs::metadata(&graph_path)?.len();
    let catalogue_size = fs::metadata(&catalogue_path)?.len();
    Ok(format!(
        "imported facts: nodes={} edges={} graph.arcana={} bytes catalogue.tsv={} bytes total={} bytes\n",
        compiled.dataset.node_count,
        compiled.dataset.edges.len(),
        graph_size,
        catalogue_size,
        graph_size + catalogue_size
    ))
}

pub fn run_query(command: &QueryCommand) -> Result<String, CliCommandError> {
    let graph = PackedGraph::open(&command.graph)?;
    let catalogue = RepositoryCatalogue::read(&command.catalogue)?;
    let matches = catalogue.lookup_by_name(&command.name);
    if matches.is_empty() {
        return Ok(format!("no exact-name matches for {:?}\n", command.name));
    }

    let mut output = String::new();
    writeln!(output, "exact-name matches: {}", matches.len()).expect("String writing cannot fail");
    for entry in matches {
        write_node(&mut output, "node", &entry.fact, entry.node_id);
        let neighbors = if command.reverse {
            graph.reverse_neighbors(entry.node_id)?
        } else {
            graph.forward_neighbors(entry.node_id)?
        };
        for neighbor in neighbors {
            let relation = edge_kind_to_relation(neighbor.kind)
                .ok_or(CliCommandError::UnknownEdgeKind(neighbor.kind.0))?;
            if command
                .relation
                .as_ref()
                .is_some_and(|wanted| wanted != &relation)
            {
                continue;
            }
            let target = catalogue
                .entries()
                .iter()
                .find(|candidate| candidate.node_id == neighbor.node)
                .ok_or(CliCommandError::MissingCatalogueNode(neighbor.node))?;
            write!(output, "  relation={} ", relation_name(&relation))
                .expect("String writing cannot fail");
            write_node(&mut output, "neighbor", &target.fact, target.node_id);
        }
    }
    Ok(output)
}

fn write_node(output: &mut String, label: &str, fact: &NodeFact, node_id: NodeId) {
    write!(
        output,
        "{label} node_id={} key={:016x} kind={:?} path={:?} name={:?} content_id={}",
        node_id.0,
        fact.key.0,
        fact.kind,
        fact.path,
        fact.name,
        fact.content_id
            .map_or_else(|| "-".to_owned(), |id| format!("{:016x}", id.0))
    )
    .expect("String writing cannot fail");
    if let Some(span) = &fact.span {
        write!(
            output,
            " span={:?}:{}:{}-{}:{}",
            span.path, span.start_line, span.start_column, span.end_line, span.end_column
        )
        .expect("String writing cannot fail");
    }
    output.push('\n');
}

fn relation_name(relation: &RelationKind) -> &'static str {
    match relation {
        RelationKind::Contains => "contains",
        RelationKind::Defines => "defines",
        RelationKind::References => "references",
        RelationKind::Imports => "imports",
        RelationKind::Calls => "calls",
        RelationKind::Implements => "implements",
        RelationKind::Extends => "extends",
        RelationKind::Includes => "includes",
        RelationKind::DependsOn => "depends-on",
        RelationKind::Tests => "tests",
        RelationKind::Documents => "documents",
        RelationKind::Generates => "generates",
    }
}
