use std::collections::BTreeMap;

use super::LexiconSnapshotError;
use super::object::{EdgeRecord, FactRecord, NodeRecord, SpanRecord, UnresolvedRecord};
use crate::repository::{
    ContentId, EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, SourceSpan,
    UnresolvedReason, UnresolvedReferenceFact, normalize_repository_path,
};

pub(super) fn build_repository_facts(
    records: Vec<FactRecord>,
) -> Result<(RepositoryFacts, Vec<String>), LexiconSnapshotError> {
    let mut nodes = BTreeMap::<String, NodeRecord>::new();
    let mut edges = Vec::new();
    let mut unresolved = Vec::new();
    let mut compatibility = BTreeMap::<String, usize>::new();
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
        let kind = NodeKind::parse(&record.kind).unwrap_or_else(|| {
            *compatibility
                .entry(format!(
                    "unrecognized Lexicon node kind {:?}; treating as symbol",
                    record.kind
                ))
                .or_default() += 1;
            NodeKind::Symbol
        });
        if record.qualified_name.is_empty() {
            return Err(LexiconSnapshotError::Malformed("node qualified name"));
        }
        facts.nodes.push(NodeFact {
            key,
            external_identity: Some(record.id),
            kind,
            path,
            name: record.name,
            qualified_name: record.qualified_name,
            content_id,
            span: convert_span(record.span)?,
        });
    }
    for record in edges {
        if let Some(edge) = convert_edge(&external_ids, record, &mut compatibility)? {
            facts.edges.push(edge);
        }
    }
    for record in unresolved {
        if let Some(reference) = convert_unresolved(&external_ids, record, &mut compatibility)? {
            facts.unresolved.push(reference);
        }
    }
    facts.nodes.sort_unstable();
    facts.nodes.dedup();
    facts.edges.sort_unstable();
    facts.edges.dedup();
    facts.unresolved.sort_unstable();
    facts.unresolved.dedup();
    let warnings = compatibility
        .into_iter()
        .map(|(message, count)| format!("{message} ({count} record(s))"))
        .collect();
    Ok((facts, warnings))
}

fn convert_edge(
    ids: &BTreeMap<String, NodeKey>,
    record: EdgeRecord,
    compatibility: &mut BTreeMap<String, usize>,
) -> Result<Option<EdgeFact>, LexiconSnapshotError> {
    validate_owner(record.owner.as_deref())?;
    let source = lookup_id(ids, &record.source)?;
    let target = lookup_id(ids, &record.target)?;
    let Some(relation) = RelationKind::parse(&record.relation) else {
        *compatibility
            .entry(format!(
                "unrecognized Lexicon edge relation {:?}; skipping edge",
                record.relation
            ))
            .or_default() += 1;
        return Ok(None);
    };
    Ok(Some(EdgeFact {
        source,
        target,
        relation,
        span: convert_span(record.span)?,
    }))
}

fn convert_unresolved(
    ids: &BTreeMap<String, NodeKey>,
    record: UnresolvedRecord,
    compatibility: &mut BTreeMap<String, usize>,
) -> Result<Option<UnresolvedReferenceFact>, LexiconSnapshotError> {
    validate_owner(record.owner.as_deref())?;
    let source = lookup_id(ids, &record.source)?;
    let Some(relation) = RelationKind::parse(&record.relation) else {
        *compatibility
            .entry(format!(
                "unrecognized Lexicon unresolved relation {:?}; skipping record",
                record.relation
            ))
            .or_default() += 1;
        return Ok(None);
    };
    let reason = UnresolvedReason::parse(&record.reason)
        .ok_or(LexiconSnapshotError::Malformed("empty unresolved reason"))?;
    if reason.is_unknown() {
        *compatibility
            .entry(format!(
                "unrecognized Lexicon unresolved reason {:?}; preserving label",
                record.reason
            ))
            .or_default() += 1;
    }
    Ok(Some(UnresolvedReferenceFact {
        source,
        relation,
        expression: record.expression,
        candidate_namespace: record.candidate_namespace,
        candidate_name: record.candidate_name,
        reason,
        span: convert_span(record.span)?,
    }))
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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::repository::{NodeKind, UnresolvedReason};

    fn id(value: char) -> String {
        format!("sha256:{}", value.to_string().repeat(64))
    }

    fn node(identity: String, kind: &str, name: &str) -> FactRecord {
        FactRecord::Node(NodeRecord {
            attributes: None,
            content_id: None,
            id: identity,
            kind: kind.to_owned(),
            name: name.to_owned(),
            owner: Some("source.cc".to_owned()),
            path: "source.cc".to_owned(),
            qualified_name: format!("demo::{name}"),
            span: None,
        })
    }

    #[test]
    fn accepts_unknown_labels_with_explicit_degradation_warnings() {
        let source = id('1');
        let target = id('2');
        let records = vec![
            node(source.clone(), "future-node-kind", "source"),
            node(target.clone(), "function", "target"),
            FactRecord::Edge(EdgeRecord {
                attributes: None,
                owner: Some("source.cc".to_owned()),
                relation: "future-edge-relation".to_owned(),
                source: source.clone(),
                span: None,
                target,
            }),
            FactRecord::Unresolved(UnresolvedRecord {
                attributes: None,
                candidate_name: None,
                candidate_namespace: None,
                expression: "future()".to_owned(),
                owner: Some("source.cc".to_owned()),
                reason: "future-unresolved-reason".to_owned(),
                relation: "calls".to_owned(),
                source: source.clone(),
                span: None,
            }),
            FactRecord::Unresolved(UnresolvedRecord {
                attributes: None,
                candidate_name: None,
                candidate_namespace: None,
                expression: "ignored()".to_owned(),
                owner: Some("source.cc".to_owned()),
                reason: "missing-target".to_owned(),
                relation: "future-unresolved-relation".to_owned(),
                source,
                span: None,
            }),
        ];

        let (facts, warnings) = build_repository_facts(records).unwrap();
        assert_eq!(facts.nodes.len(), 2);
        assert!(facts.nodes.iter().any(|node| node.kind == NodeKind::Symbol));
        assert!(facts.edges.is_empty());
        assert_eq!(facts.unresolved.len(), 1);
        assert_eq!(
            facts.unresolved[0].reason,
            UnresolvedReason::Unknown("future-unresolved-reason".to_owned())
        );
        for expected in [
            "future-node-kind",
            "future-edge-relation",
            "future-unresolved-reason",
            "future-unresolved-relation",
        ] {
            assert!(warnings.iter().any(|warning| warning.contains(expected)));
        }
    }
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
