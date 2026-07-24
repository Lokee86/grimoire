use std::collections::BTreeMap;
use std::fmt::Write as FmtWrite;

use crate::repository::{EdgeFact, NodeFact, NodeKey, RepositoryFacts, UnresolvedReferenceFact};

const MAX_OUTGOING: usize = 32;
const MAX_INCOMING: usize = 32;
const MAX_UNRESOLVED: usize = 16;

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct GraphDocument {
    pub node_key: u64,
    pub kind: String,
    pub path: String,
    pub name: String,
    pub text: String,
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
    writeln!(output, "path: {}", node.path).unwrap();
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
    for edge in edges.iter().take(limit) {
        let other_key = if forward { edge.target } else { edge.source };
        let Some(other) = nodes.get(&other_key) else {
            continue;
        };
        writeln!(
            output,
            "{label} {} {} {} at {}",
            edge.relation.as_str(),
            other.kind.as_str(),
            other.name,
            other.path
        )
        .unwrap();
    }
    if edges.len() > limit {
        writeln!(output, "{label} omitted: {}", edges.len() - limit).unwrap();
    }
}

#[cfg(test)]
#[path = "documents_tests.rs"]
mod tests;
