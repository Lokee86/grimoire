use std::fmt::Write as FmtWrite;

use arcana::repository::{NodeFact, RelationKind, RepositoryCatalogue, edge_kind_to_relation};
use arcana::storage::PackedGraph;
use arcana::synthetic::NodeId;

use crate::cli::QueryCommand;
use crate::cli_commands::CliCommandError;

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
        RelationKind::PossibleCalls => "possible-calls",
        RelationKind::ConvertsTo => "converts-to",
        RelationKind::Implements => "implements",
        RelationKind::UsesTrait => "uses-trait",
        RelationKind::Overrides => "overrides",
        RelationKind::Reads => "reads",
        RelationKind::Writes => "writes",
        RelationKind::Annotates => "annotates",
        RelationKind::Extends => "extends",
        RelationKind::Includes => "includes",
        RelationKind::DependsOn => "depends-on",
        RelationKind::Tests => "tests",
        RelationKind::Documents => "documents",
        RelationKind::Generates => "generates",
    }
}
