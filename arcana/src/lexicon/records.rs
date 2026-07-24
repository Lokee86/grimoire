use std::collections::BTreeMap;

use super::LexiconSnapshotError;
use super::object::{EdgeRecord, FactRecord, NodeRecord, SpanRecord, UnresolvedRecord};
use crate::repository::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, SourceSpan,
    UnresolvedReason, UnresolvedReferenceFact, normalize_repository_path,
};

pub(super) fn build_repository_facts(
    records: Vec<FactRecord>,
) -> Result<RepositoryFacts, LexiconSnapshotError> {
    let mut nodes = BTreeMap::<String, NodeRecord>::new();
    let mut edges = Vec::new();
    let mut unresolved = Vec::new();
    for record in records {
        match record {
            FactRecord::Node(record) => match nodes.get(&record.id) {
                Some(existing) if existing != &record => {
                    return Err(LexiconSnapshotError::ConflictingNode(record.id));
                }
                Some(_) => {}
                None => {
                    nodes.insert(record.id.clone(), record);
                }
            },
            FactRecord::Edge(record) => edges.push(record),
            FactRecord::Unresolved(record) => unresolved.push(record),
        }
    }

    let mut facts = RepositoryFacts::default();
    let mut external_ids = BTreeMap::<String, NodeKey>::new();
    let mut compact_ids = BTreeMap::<NodeKey, String>::new();
    for record in nodes.into_values() {
        validate_sha256_id(&record.id)?;
        validate_owner(record.owner.as_deref())?;
        let path = normalize_path(&record.path)?;
        let key = NodeKey::from_identity(record.id.as_bytes());
        if compact_ids
            .insert(key, record.id.clone())
            .is_some_and(|existing| existing != record.id)
        {
            return Err(LexiconSnapshotError::Malformed("node identity collision"));
        }
        external_ids.insert(record.id.clone(), key);
        let content_id = record
            .content_id
            .as_deref()
            .map(|id| -> Result<ContentId, LexiconSnapshotError> {
                validate_sha256_id(id)?;
                Ok(ContentId::from_bytes(id.as_bytes()))
            })
            .transpose()?;
        let kind =
            NodeKind::parse(&record.kind).ok_or(LexiconSnapshotError::Malformed("node kind"))?;
        if record.qualified_name.is_empty() {
            return Err(LexiconSnapshotError::Malformed("node qualified name"));
        }
        facts.nodes.push(NodeFact {
            key,
            external_identity: Some(record.id),
            kind,
            path,
            name: record.name,
            content_id,
            span: convert_span(record.span)?,
        });
    }
    for record in edges {
        facts.edges.push(convert_edge(&external_ids, record)?);
    }
    for record in unresolved {
        facts
            .unresolved
            .push(convert_unresolved(&external_ids, record)?);
    }
    facts.nodes.sort_unstable();
    facts.nodes.dedup();
    facts.edges.sort_unstable();
    facts.edges.dedup();
    facts.unresolved.sort_unstable();
    facts.unresolved.dedup();
    Ok(facts)
}

fn convert_edge(
    ids: &BTreeMap<String, NodeKey>,
    record: EdgeRecord,
) -> Result<EdgeFact, LexiconSnapshotError> {
    validate_owner(record.owner.as_deref())?;
    Ok(EdgeFact {
        source: lookup_id(ids, &record.source)?,
        target: lookup_id(ids, &record.target)?,
        relation: RelationKind::parse(&record.relation)
            .ok_or(LexiconSnapshotError::Malformed("edge relation"))?,
        span: convert_span(record.span)?,
    })
}

fn convert_unresolved(
    ids: &BTreeMap<String, NodeKey>,
    record: UnresolvedRecord,
) -> Result<UnresolvedReferenceFact, LexiconSnapshotError> {
    validate_owner(record.owner.as_deref())?;
    Ok(UnresolvedReferenceFact {
        source: lookup_id(ids, &record.source)?,
        relation: RelationKind::parse(&record.relation)
            .ok_or(LexiconSnapshotError::Malformed("unresolved relation"))?,
        expression: record.expression,
        candidate_namespace: record.candidate_namespace,
        candidate_name: record.candidate_name,
        reason: UnresolvedReason::parse(&record.reason)
            .ok_or(LexiconSnapshotError::Malformed("unresolved reason"))?,
        span: convert_span(record.span)?,
    })
}

fn lookup_id(
    ids: &BTreeMap<String, NodeKey>,
    external_id: &str,
) -> Result<NodeKey, LexiconSnapshotError> {
    validate_sha256_id(external_id)?;
    ids.get(external_id)
        .copied()
        .ok_or(LexiconSnapshotError::Malformed("unknown relationship node"))
}

fn validate_owner(owner: Option<&str>) -> Result<(), LexiconSnapshotError> {
    if let Some(owner) = owner {
        normalize_path(owner)?;
    }
    Ok(())
}

fn convert_span(span: Option<SpanRecord>) -> Result<Option<SourceSpan>, LexiconSnapshotError> {
    let Some(span) = span else {
        return Ok(None);
    };
    Ok(Some(SourceSpan {
        path: normalize_path(&span.path)?,
        start_line: u32_value(span.start_line, "span start line")?,
        start_column: u32_value(span.start_column, "span start column")?,
        end_line: u32_value(span.end_line, "span end line")?,
        end_column: u32_value(span.end_column, "span end column")?,
    }))
}

fn u32_value(value: u64, field: &'static str) -> Result<u32, LexiconSnapshotError> {
    u32::try_from(value).map_err(|_| LexiconSnapshotError::Malformed(field))
}

fn normalize_path(path: &str) -> Result<String, LexiconSnapshotError> {
    normalize_repository_path(path).map_err(|_| LexiconSnapshotError::InvalidPath {
        field: "fact",
        path: path.to_owned(),
    })
}

fn validate_sha256_id(value: &str) -> Result<(), LexiconSnapshotError> {
    let Some(digest) = value.strip_prefix("sha256:") else {
        return Err(LexiconSnapshotError::InvalidId(value.to_owned()));
    };
    if digest.len() != 64
        || !digest
            .bytes()
            .all(|byte| byte.is_ascii_digit() || (b'a'..=b'f').contains(&byte))
    {
        return Err(LexiconSnapshotError::InvalidId(value.to_owned()));
    }
    Ok(())
}
