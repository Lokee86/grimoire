use std::collections::BTreeMap;
use std::fmt::Write as FmtWrite;

use crate::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, UnresolvedReferenceFact,
};

const MAX_OUTGOING: usize = 12;
const MAX_INCOMING: usize = 12;
const MAX_UNRESOLVED: usize = 8;
const MAX_DOCUMENT_BYTES: usize = 6_000;
const DOCUMENT_TRUNCATED_SUFFIX: &str = "\ndocument truncated\n";
pub const SEMANTIC_ELIGIBILITY_POLICY_VERSION: u64 = 6;

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct GraphDocument {
    pub node_key: u64,
    pub kind: String,
    pub path: String,
    pub name: String,
    pub text: String,
}

/// Returns whether a complete graph node is a useful semantic retrieval entry point.
pub fn semantic_eligible(kind: &NodeKind) -> bool {
    match kind {
        NodeKind::Repository
        | NodeKind::Directory
        | NodeKind::Field
        | NodeKind::Variable
        | NodeKind::Parameter
        | NodeKind::Import
        | NodeKind::Export
        | NodeKind::Constant
        | NodeKind::Test => false,
        NodeKind::File
        | NodeKind::Module
        | NodeKind::Namespace
        | NodeKind::Symbol
        | NodeKind::Type
        | NodeKind::Interface
        | NodeKind::Trait
        | NodeKind::Function
        | NodeKind::Method
        | NodeKind::Constructor
        | NodeKind::Signal
        | NodeKind::HttpEndpoint
        | NodeKind::MessageChannel
        | NodeKind::ConfigKey
        | NodeKind::Process
        | NodeKind::CliCommand
        | NodeKind::Protocol
        | NodeKind::StatePath => true,
    }
}

fn semantic_node_eligible(node: &NodeFact) -> bool {
    if !semantic_eligible(&node.kind) || node.path.starts_with('@') {
        return false;
    }
    let name = node.name.to_ascii_lowercase();
    !name.starts_with("closure@") && !name.starts_with("<lambda@") && !name.starts_with("lambda@")
}

pub fn graph_documents(facts: &RepositoryFacts) -> Vec<GraphDocument> {
    let nodes = facts
        .nodes
        .iter()
        .map(|node| (node.key, node))
        .collect::<BTreeMap<_, _>>();
    let mut outgoing = BTreeMap::<NodeKey, Vec<&EdgeFact>>::new();
    let mut incoming = BTreeMap::<NodeKey, Vec<&EdgeFact>>::new();
    let mut unresolved = BTreeMap::<NodeKey, Vec<&UnresolvedReferenceFact>>::new();

    for edge in &facts.edges {
        outgoing.entry(edge.source).or_default().push(edge);
        incoming.entry(edge.target).or_default().push(edge);
    }
    for reference in &facts.unresolved {
        unresolved
            .entry(reference.source)
            .or_default()
            .push(reference);
    }
    for edges in outgoing.values_mut() {
        edges.sort_unstable_by_key(|edge| (edge.relation.clone(), edge.target));
    }
    for edges in incoming.values_mut() {
        edges.sort_unstable_by_key(|edge| (edge.relation.clone(), edge.source));
    }
    for references in unresolved.values_mut() {
        references.sort_unstable();
    }

    nodes
        .values()
        .filter(|node| semantic_node_eligible(node))
        .map(|node| GraphDocument {
            node_key: node.key.0,
            kind: node.kind.as_str().to_owned(),
            path: node.path.clone(),
            name: node.name.clone(),
            text: render_document(
                node,
                &nodes,
                outgoing.get(&node.key).map_or(&[], Vec::as_slice),
                incoming.get(&node.key).map_or(&[], Vec::as_slice),
                unresolved.get(&node.key).map_or(&[], Vec::as_slice),
            ),
        })
        .collect()
}

fn render_document(
    node: &NodeFact,
    nodes: &BTreeMap<NodeKey, &NodeFact>,
    outgoing: &[&EdgeFact],
    incoming: &[&EdgeFact],
    unresolved: &[&UnresolvedReferenceFact],
) -> String {
    let mut output = String::new();
    writeln!(output, "repository graph node").unwrap();
    writeln!(output, "kind: {}", node.kind.as_str()).unwrap();
    writeln!(output, "name: {}", node.name).unwrap();
    writeln!(output, "name terms: {}", identifier_terms(&node.name)).unwrap();
    writeln!(output, "path: {}", node.path).unwrap();
    writeln!(output, "path terms: {}", identifier_terms(&node.path)).unwrap();
    if let Some(span) = &node.span {
        writeln!(
            output,
            "source: {}:{}:{}-{}:{}",
            span.path, span.start_line, span.start_column, span.end_line, span.end_column
        )
        .unwrap();
    }
    render_edges(&mut output, "outgoing", outgoing, nodes, true, MAX_OUTGOING);
    render_edges(
        &mut output,
        "incoming",
        incoming,
        nodes,
        false,
        MAX_INCOMING,
    );
    for reference in unresolved.iter().take(MAX_UNRESOLVED) {
        writeln!(
            output,
            "unresolved {} {} candidate_namespace={} candidate_name={}",
            reference.relation.as_str(),
            reference.expression,
            reference.candidate_namespace.as_deref().unwrap_or("-"),
            reference.candidate_name.as_deref().unwrap_or("-")
        )
        .unwrap();
    }
    if unresolved.len() > MAX_UNRESOLVED {
        writeln!(
            output,
            "unresolved omitted: {}",
            unresolved.len() - MAX_UNRESOLVED
        )
        .unwrap();
    }
    bound_document(output)
}

fn bound_document(mut output: String) -> String {
    if output.len() <= MAX_DOCUMENT_BYTES {
        return output;
    }
    let mut limit = MAX_DOCUMENT_BYTES - DOCUMENT_TRUNCATED_SUFFIX.len();
    while !output.is_char_boundary(limit) {
        limit -= 1;
    }
    output.truncate(limit);
    output.push_str(DOCUMENT_TRUNCATED_SUFFIX);
    output
}

fn render_edges(
    output: &mut String,
    label: &str,
    edges: &[&EdgeFact],
    nodes: &BTreeMap<NodeKey, &NodeFact>,
    forward: bool,
    limit: usize,
) {
    let mut semantic_edges = edges
        .iter()
        .filter_map(|edge| {
            let edge = *edge;
            let other_key = if forward { edge.target } else { edge.source };
            let other = nodes.get(&other_key).copied()?;
            semantic_node_eligible(other).then_some((edge, other))
        })
        .collect::<Vec<_>>();
    semantic_edges.sort_unstable_by(|(left_edge, left_node), (right_edge, right_node)| {
        relation_priority(&left_edge.relation)
            .cmp(&relation_priority(&right_edge.relation))
            .then_with(|| left_edge.relation.cmp(&right_edge.relation))
            .then_with(|| left_node.kind.cmp(&right_node.kind))
            .then_with(|| left_node.path.cmp(&right_node.path))
            .then_with(|| left_node.name.cmp(&right_node.name))
            .then_with(|| left_node.key.cmp(&right_node.key))
    });

    for (edge, other) in semantic_edges.iter().take(limit) {
        writeln!(
            output,
            "{label} {} {} {} ({}) at {}",
            edge.relation.as_str(),
            other.kind.as_str(),
            other.name,
            identifier_terms(&other.name),
            other.path
        )
        .unwrap();
    }
    if semantic_edges.len() > limit {
        writeln!(output, "{label} omitted: {}", semantic_edges.len() - limit).unwrap();
    }
}

fn identifier_terms(value: &str) -> String {
    let characters = value.chars().collect::<Vec<_>>();
    let mut output = String::with_capacity(value.len());

    for (index, current) in characters.iter().copied().enumerate() {
        if !current.is_alphanumeric() {
            push_term_separator(&mut output);
            continue;
        }

        let previous = index.checked_sub(1).map(|previous| characters[previous]);
        let next = characters.get(index + 1).copied();
        let starts_word = current.is_uppercase()
            && (previous.is_some_and(|value| value.is_lowercase() || value.is_numeric())
                || (previous.is_some_and(char::is_uppercase)
                    && next.is_some_and(char::is_lowercase)));
        if starts_word {
            push_term_separator(&mut output);
        }
        for lowercase in current.to_lowercase() {
            output.push(lowercase);
        }
    }

    output.trim().to_owned()
}

fn push_term_separator(output: &mut String) {
    if !output.is_empty() && !output.ends_with(' ') {
        output.push(' ');
    }
}

fn relation_priority(relation: &RelationKind) -> u8 {
    match relation {
        RelationKind::Calls
        | RelationKind::ObservedCalls
        | RelationKind::CallsEndpoint
        | RelationKind::RoutesTo
        | RelationKind::HandledBy
        | RelationKind::Publishes
        | RelationKind::Consumes
        | RelationKind::CommunicatesWith
        | RelationKind::InvokesProcess
        | RelationKind::ProducesMessage
        | RelationKind::ConsumesMessage => 0,
        RelationKind::PossibleCalls | RelationKind::PassesTo => 1,
        RelationKind::Implements
        | RelationKind::Extends
        | RelationKind::UsesTrait
        | RelationKind::Overrides => 2,
        RelationKind::Generates
        | RelationKind::DependsOn
        | RelationKind::Includes
        | RelationKind::ReadsConfig
        | RelationKind::ConvertsTo => 3,
        RelationKind::Defines
        | RelationKind::Contains
        | RelationKind::Documents
        | RelationKind::Tests => 4,
        RelationKind::References
        | RelationKind::Imports
        | RelationKind::Annotates
        | RelationKind::SimilarTo => 5,
        RelationKind::Reads | RelationKind::Writes => 6,
    }
}

#[cfg(test)]
#[path = "documents_tests.rs"]
mod tests;
