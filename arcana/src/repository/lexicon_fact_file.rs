use std::collections::BTreeMap;

use serde_json::{Map, Value};

use super::super::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, SourceSpan,
    UnresolvedReason, UnresolvedReferenceFact, normalize_repository_path,
};
use super::FactFileError;

pub(super) fn parse_lexicon_facts(input: &str) -> Result<RepositoryFacts, FactFileError> {
    let mut lines = input.lines();
    let header_line = lines.next().ok_or(FactFileError::InvalidHeader)?;
    let header = parse_object(header_line, 1)?;
    if string_field(&header, "record", 1)? != "lexicon"
        || integer_field(&header, "schema_version", 1)? != 1
    {
        return Err(FactFileError::InvalidHeader);
    }
    match optional_string(&header, "mode", 1)? {
        None | Some("full") => {}
        Some(_) => return Err(FactFileError::InvalidHeader),
    }
    for field in ["adapter_version", "language", "repository"] {
        if string_field(&header, field, 1)?.is_empty() {
            return Err(FactFileError::InvalidHeader);
        }
    }

    let mut facts = RepositoryFacts::default();
    let mut external_ids = BTreeMap::<String, NodeKey>::new();
    let mut compact_ids = BTreeMap::<NodeKey, String>::new();
    let mut phase = 0_u8;

    for (index, line) in lines.enumerate() {
        let line_number = index + 2;
        if line.trim().is_empty() {
            return Err(FactFileError::MalformedLine { line: line_number });
        }
        let record = parse_object(line, line_number)?;
        match string_field(&record, "record", line_number)? {
            "node" if phase == 0 => {
                let external_id = string_field(&record, "id", line_number)?;
                validate_sha256_id(external_id, line_number)?;
                if external_ids.contains_key(external_id) {
                    return Err(FactFileError::MalformedLine { line: line_number });
                }
                let key = NodeKey::from_identity(external_id.as_bytes());
                if compact_ids
                    .insert(key, external_id.to_owned())
                    .is_some_and(|existing| existing != external_id)
                {
                    return Err(FactFileError::MalformedLine { line: line_number });
                }
                external_ids.insert(external_id.to_owned(), key);
                let path = normalized_field(&record, "path", line_number)?;
                let _qualified_name = string_field(&record, "qualified_name", line_number)?;
                validate_owner(&record, line_number)?;
                let content_id = optional_string(&record, "content_id", line_number)?
                    .map(|id| {
                        validate_sha256_id(id, line_number)?;
                        Ok(ContentId::from_bytes(id.as_bytes()))
                    })
                    .transpose()?;
                facts.nodes.push(NodeFact {
                    key,
                    external_identity: Some(external_id.to_owned()),
                    kind: NodeKind::parse(string_field(&record, "kind", line_number)?)
                        .ok_or(FactFileError::InvalidKind { line: line_number })?,
                    path,
                    name: string_field(&record, "name", line_number)?.to_owned(),
                    content_id,
                    span: parse_span(record.get("span"), line_number)?,
                });
            }
            "edge" if phase <= 1 => {
                phase = 1;
                validate_owner(&record, line_number)?;
                facts.edges.push(EdgeFact {
                    source: lookup_id(
                        &external_ids,
                        string_field(&record, "source", line_number)?,
                        line_number,
                    )?,
                    target: lookup_id(
                        &external_ids,
                        string_field(&record, "target", line_number)?,
                        line_number,
                    )?,
                    relation: RelationKind::parse(string_field(&record, "relation", line_number)?)
                        .ok_or(FactFileError::InvalidRelation { line: line_number })?,
                    span: parse_span(record.get("span"), line_number)?,
                });
            }
            "unresolved" if phase <= 2 => {
                phase = 2;
                validate_owner(&record, line_number)?;
                facts.unresolved.push(UnresolvedReferenceFact {
                    source: lookup_id(
                        &external_ids,
                        string_field(&record, "source", line_number)?,
                        line_number,
                    )?,
                    relation: RelationKind::parse(string_field(&record, "relation", line_number)?)
                        .ok_or(FactFileError::InvalidRelation { line: line_number })?,
                    reason: UnresolvedReason::parse(string_field(&record, "reason", line_number)?)
                        .ok_or(FactFileError::InvalidReason { line: line_number })?,
                    expression: string_field(&record, "expression", line_number)?.to_owned(),
                    candidate_namespace: optional_string(
                        &record,
                        "candidate_namespace",
                        line_number,
                    )?
                    .map(str::to_owned),
                    candidate_name: optional_string(&record, "candidate_name", line_number)?
                        .map(str::to_owned),
                    span: parse_span(record.get("span"), line_number)?,
                });
            }
            _ => return Err(FactFileError::UnknownRecord { line: line_number }),
        }
    }
    Ok(facts)
}

fn parse_object(line: &str, line_number: usize) -> Result<Map<String, Value>, FactFileError> {
    serde_json::from_str::<Map<String, Value>>(line)
        .map_err(|_| FactFileError::MalformedLine { line: line_number })
}

fn string_field<'a>(
    object: &'a Map<String, Value>,
    field: &str,
    line: usize,
) -> Result<&'a str, FactFileError> {
    object
        .get(field)
        .and_then(Value::as_str)
        .ok_or(FactFileError::MalformedLine { line })
}

fn optional_string<'a>(
    object: &'a Map<String, Value>,
    field: &str,
    line: usize,
) -> Result<Option<&'a str>, FactFileError> {
    match object.get(field) {
        None | Some(Value::Null) => Ok(None),
        Some(Value::String(value)) => Ok(Some(value)),
        Some(_) => Err(FactFileError::MalformedLine { line }),
    }
}

fn integer_field(
    object: &Map<String, Value>,
    field: &str,
    line: usize,
) -> Result<u64, FactFileError> {
    object
        .get(field)
        .and_then(Value::as_u64)
        .ok_or(FactFileError::MalformedLine { line })
}

fn normalized_field(
    object: &Map<String, Value>,
    field: &str,
    line: usize,
) -> Result<String, FactFileError> {
    normalize_repository_path(string_field(object, field, line)?)
        .map_err(|_| FactFileError::MalformedLine { line })
}

fn validate_owner(object: &Map<String, Value>, line: usize) -> Result<(), FactFileError> {
    if let Some(owner) = optional_string(object, "owner", line)? {
        normalize_repository_path(owner).map_err(|_| FactFileError::MalformedLine { line })?;
    }
    Ok(())
}

fn lookup_id(
    ids: &BTreeMap<String, NodeKey>,
    external_id: &str,
    line: usize,
) -> Result<NodeKey, FactFileError> {
    validate_sha256_id(external_id, line)?;
    ids.get(external_id)
        .copied()
        .ok_or(FactFileError::MalformedLine { line })
}

fn validate_sha256_id(value: &str, line: usize) -> Result<(), FactFileError> {
    let Some(digest) = value.strip_prefix("sha256:") else {
        return Err(FactFileError::InvalidNumber { line });
    };
    if digest.len() != 64
        || !digest
            .bytes()
            .all(|byte| byte.is_ascii_digit() || (b'a'..=b'f').contains(&byte))
    {
        return Err(FactFileError::InvalidNumber { line });
    }
    Ok(())
}

fn parse_span(value: Option<&Value>, line: usize) -> Result<Option<SourceSpan>, FactFileError> {
    let Some(value) = value else {
        return Ok(None);
    };
    let object = value
        .as_object()
        .ok_or(FactFileError::InvalidSpan { line })?;
    let path = normalize_repository_path(string_field(object, "path", line)?)
        .map_err(|_| FactFileError::InvalidSpan { line })?;
    Ok(Some(SourceSpan {
        path,
        start_line: u32_field(object, "start_line", line)?,
        start_column: u32_field(object, "start_column", line)?,
        end_line: u32_field(object, "end_line", line)?,
        end_column: u32_field(object, "end_column", line)?,
    }))
}

fn u32_field(object: &Map<String, Value>, field: &str, line: usize) -> Result<u32, FactFileError> {
    let value = integer_field(object, field, line)?;
    u32::try_from(value).map_err(|_| FactFileError::InvalidNumber { line })
}
