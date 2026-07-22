use super::{
    ContentId, EdgeFact, FactFileError, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts,
    SourceSpan, UnresolvedReason, UnresolvedReferenceFact, normalize_repository_path,
};

const HEADER_V1: &str = "version\t1";
pub const FACT_SCHEMA_VERSION: u64 = 2;
const HEADER_V2: &str = "version\t2";

/// Encodes repository facts as canonical tab-separated UTF-8 lines.
pub fn encode_facts(facts: &RepositoryFacts) -> String {
    let mut nodes = facts.nodes.clone();
    let mut edges = facts.edges.clone();
    let mut unresolved = facts.unresolved.clone();
    nodes.sort_unstable();
    edges.sort_unstable();
    unresolved.sort_unstable();

    let mut output = String::from(HEADER_V2);
    output.push('\n');
    for node in nodes {
        output.push_str("N\t");
        push_field(&mut output, &format_id(node.key.0));
        output.push('\t');
        push_field(&mut output, node.kind.as_str());
        output.push('\t');
        push_field(&mut output, &node.path);
        output.push('\t');
        push_field(&mut output, &node.name);
        output.push('\t');
        let content_id = node
            .content_id
            .map_or_else(|| "-".to_owned(), |id| format_id(id.0));
        push_field(&mut output, &content_id);
        push_span(&mut output, node.span.as_ref());
        output.push('\n');
    }
    for edge in edges {
        output.push_str("E\t");
        push_field(&mut output, &format_id(edge.source.0));
        output.push('\t');
        push_field(&mut output, &format_id(edge.target.0));
        output.push('\t');
        push_field(&mut output, edge.relation.as_str());
        push_span(&mut output, edge.span.as_ref());
        output.push('\n');
    }
    for reference in unresolved {
        output.push_str("U\t");
        push_field(&mut output, &format_id(reference.source.0));
        output.push('\t');
        push_field(&mut output, reference.relation.as_str());
        output.push('\t');
        push_field(&mut output, reference.reason.as_str());
        output.push('\t');
        push_field(&mut output, &reference.expression);
        output.push('\t');
        push_optional_field(&mut output, reference.candidate_namespace.as_deref());
        output.push('\t');
        push_optional_field(&mut output, reference.candidate_name.as_deref());
        push_span(&mut output, reference.span.as_ref());
        output.push('\n');
    }
    output
}

/// Parses the canonical tab-separated repository fact format.
pub fn parse_facts(input: &str) -> Result<RepositoryFacts, FactFileError> {
    let mut lines = input.lines();
    let version = match lines.next() {
        Some(HEADER_V1) => 1,
        Some(HEADER_V2) => 2,
        _ => return Err(FactFileError::InvalidHeader),
    };

    let mut facts = RepositoryFacts::default();
    for (index, line) in lines.enumerate() {
        let line_number = index + 2;
        if line.is_empty() {
            return Err(FactFileError::MalformedLine { line: line_number });
        }
        let fields = line
            .split('\t')
            .map(|field| unescape(field, line_number))
            .collect::<Result<Vec<_>, _>>()?;
        match fields.first().map(String::as_str) {
            Some("N") => facts.nodes.push(parse_node(&fields, line_number)?),
            Some("E") => facts.edges.push(parse_edge(&fields, line_number)?),
            Some("U") if version >= 2 => facts
                .unresolved
                .push(parse_unresolved(&fields, line_number)?),
            _ => return Err(FactFileError::UnknownRecord { line: line_number }),
        }
    }
    Ok(facts)
}

fn parse_node(fields: &[String], line: usize) -> Result<NodeFact, FactFileError> {
    if fields.len() != 11 {
        return Err(FactFileError::MalformedLine { line });
    }
    let path =
        normalize_repository_path(&fields[3]).map_err(|_| FactFileError::MalformedLine { line })?;
    Ok(NodeFact {
        key: parse_id(&fields[1], line).map(NodeKey::from_u64)?,
        kind: NodeKind::parse(&fields[2]).ok_or(FactFileError::InvalidKind { line })?,
        path,
        name: fields[4].clone(),
        content_id: parse_optional_id(&fields[5], line)?.map(ContentId::from_u64),
        span: parse_span(&fields[6..11], line)?,
    })
}

fn parse_edge(fields: &[String], line: usize) -> Result<EdgeFact, FactFileError> {
    if fields.len() != 9 {
        return Err(FactFileError::MalformedLine { line });
    }
    Ok(EdgeFact {
        source: NodeKey::from_u64(parse_id(&fields[1], line)?),
        target: NodeKey::from_u64(parse_id(&fields[2], line)?),
        relation: RelationKind::parse(&fields[3]).ok_or(FactFileError::InvalidRelation { line })?,
        span: parse_span(&fields[4..9], line)?,
    })
}

fn parse_unresolved(
    fields: &[String],
    line: usize,
) -> Result<UnresolvedReferenceFact, FactFileError> {
    if fields.len() != 12 {
        return Err(FactFileError::MalformedLine { line });
    }
    Ok(UnresolvedReferenceFact {
        source: NodeKey::from_u64(parse_id(&fields[1], line)?),
        relation: RelationKind::parse(&fields[2]).ok_or(FactFileError::InvalidRelation { line })?,
        reason: UnresolvedReason::parse(&fields[3]).ok_or(FactFileError::InvalidReason { line })?,
        expression: fields[4].clone(),
        candidate_namespace: parse_optional_field(&fields[5]),
        candidate_name: parse_optional_field(&fields[6]),
        span: parse_span(&fields[7..12], line)?,
    })
}

fn parse_span(fields: &[String], line: usize) -> Result<Option<SourceSpan>, FactFileError> {
    if fields.len() != 5 {
        return Err(FactFileError::InvalidSpan { line });
    }
    let absent = fields.iter().all(|field| field == "-");
    if absent {
        return Ok(None);
    }
    if fields.iter().any(|field| field == "-") {
        return Err(FactFileError::InvalidSpan { line });
    }
    let path =
        normalize_repository_path(&fields[0]).map_err(|_| FactFileError::InvalidSpan { line })?;
    Ok(Some(SourceSpan {
        path,
        start_line: parse_u32(&fields[1], line)?,
        start_column: parse_u32(&fields[2], line)?,
        end_line: parse_u32(&fields[3], line)?,
        end_column: parse_u32(&fields[4], line)?,
    }))
}

fn parse_optional_id(value: &str, line: usize) -> Result<Option<u64>, FactFileError> {
    if value == "-" {
        Ok(None)
    } else {
        parse_id(value, line).map(Some)
    }
}

fn parse_id(value: &str, line: usize) -> Result<u64, FactFileError> {
    if value.len() != 16 {
        return Err(FactFileError::InvalidNumber { line });
    }
    u64::from_str_radix(value, 16).map_err(|_| FactFileError::InvalidNumber { line })
}

fn parse_u32(value: &str, line: usize) -> Result<u32, FactFileError> {
    value
        .parse()
        .map_err(|_| FactFileError::InvalidNumber { line })
}

fn push_span(output: &mut String, span: Option<&SourceSpan>) {
    if let Some(span) = span {
        for field in [
            span.path.as_str(),
            &span.start_line.to_string(),
            &span.start_column.to_string(),
            &span.end_line.to_string(),
            &span.end_column.to_string(),
        ] {
            output.push('\t');
            push_field(output, field);
        }
    } else {
        output.push_str("\t-\t-\t-\t-\t-");
    }
}

fn push_optional_field(output: &mut String, value: Option<&str>) {
    push_field(output, value.unwrap_or("-"));
}

fn parse_optional_field(value: &str) -> Option<String> {
    (value != "-").then(|| value.to_owned())
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

fn unescape(value: &str, line: usize) -> Result<String, FactFileError> {
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
                _ => return Err(FactFileError::InvalidEscape { line }),
            });
            escaped = false;
        } else if character == '\\' {
            escaped = true;
        } else {
            output.push(character);
        }
    }
    if escaped {
        return Err(FactFileError::InvalidEscape { line });
    }
    Ok(output)
}

fn format_id(value: u64) -> String {
    format!("{value:016x}")
}
