use serde_json::{Map, Value};
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;

pub(crate) type JsonMap = Map<String, Value>;

pub(crate) struct Facts {
    pub(crate) nodes: BTreeMap<String, Value>,
    pub(crate) edges: BTreeMap<String, Value>,
    pub(crate) unresolved: BTreeMap<String, Value>,
    pub(crate) dataflow_edges: std::collections::BTreeSet<String>,
}

impl Facts {
    pub(crate) fn new() -> Self {
        Self {
            nodes: BTreeMap::new(),
            edges: BTreeMap::new(),
            unresolved: BTreeMap::new(),
            dataflow_edges: std::collections::BTreeSet::new(),
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub(crate) fn add_node(
        &mut self,
        language: &str,
        kind: &str,
        canonical: &str,
        name: &str,
        path: &str,
        qualified_name: &str,
        content_id: Option<String>,
        span: Option<Value>,
        attributes: BTreeMap<String, Value>,
    ) -> String {
        let id = stable_id(language, kind, canonical);
        let mut node = JsonMap::new();
        node.insert(
            "attributes".into(),
            attributes.into_iter().collect::<Map<_, _>>().into(),
        );
        if let Some(content_id) = content_id {
            node.insert("content_id".into(), Value::String(content_id));
        }
        node.insert("id".into(), Value::String(id.clone()));
        node.insert("kind".into(), Value::String(kind.into()));
        node.insert("name".into(), Value::String(name.into()));
        node.insert("path".into(), Value::String(path.into()));
        node.insert(
            "qualified_name".into(),
            Value::String(qualified_name.into()),
        );
        node.insert("record".into(), Value::String("node".into()));
        if let Some(span) = span {
            node.insert("span".into(), span);
        }
        self.nodes.entry(id.clone()).or_insert(Value::Object(node));
        id
    }

    pub(crate) fn add_edge(
        &mut self,
        source: &str,
        target: &str,
        relation: &str,
        span: Option<Value>,
    ) {
        let mut edge = JsonMap::new();
        edge.insert("record".into(), Value::String("edge".into()));
        edge.insert("relation".into(), Value::String(relation.into()));
        edge.insert("source".into(), Value::String(source.into()));
        if let Some(span) = span.clone() {
            edge.insert("span".into(), span);
        }
        edge.insert("target".into(), Value::String(target.into()));
        let key = format!("{source}\0{target}\0{relation}\0{}", span_key(&span));
        self.edges.entry(key).or_insert(Value::Object(edge));
    }

    pub(crate) fn add_edge_with_attributes(
        &mut self,
        source: &str,
        target: &str,
        relation: &str,
        span: Option<Value>,
        attributes: BTreeMap<String, Value>,
    ) {
        let mut edge = JsonMap::new();
        edge.insert(
            "attributes".into(),
            attributes.into_iter().collect::<Map<_, _>>().into(),
        );
        edge.insert("record".into(), Value::String("edge".into()));
        edge.insert("relation".into(), Value::String(relation.into()));
        edge.insert("source".into(), Value::String(source.into()));
        if let Some(span) = span.clone() {
            edge.insert("span".into(), span);
        }
        edge.insert("target".into(), Value::String(target.into()));
        let key = format!("{source}\0{target}\0{relation}\0{}", span_key(&span));
        self.edges.entry(key).or_insert(Value::Object(edge));
    }

    pub(crate) fn add_unresolved(
        &mut self,
        source: &str,
        relation: &str,
        expression: &str,
        reason: &str,
        span: Option<Value>,
    ) {
        let mut record = JsonMap::new();
        record.insert("expression".into(), Value::String(expression.into()));
        record.insert("reason".into(), Value::String(reason.into()));
        record.insert("record".into(), Value::String("unresolved".into()));
        record.insert("relation".into(), Value::String(relation.into()));
        record.insert("source".into(), Value::String(source.into()));
        if let Some(span) = span.clone() {
            record.insert("span".into(), span);
        }
        let key = format!(
            "{source}\0{relation}\0{expression}\0{reason}\0{}",
            span_key(&span)
        );
        self.unresolved.entry(key).or_insert(Value::Object(record));
    }

    pub(crate) fn add_dataflow_edge(
        &mut self,
        source: &str,
        target: &str,
        relation: &str,
        span: Option<Value>,
    ) {
        let key = format!("{source}\0{target}\0{relation}");
        if !self.dataflow_edges.insert(key) {
            return;
        }
        self.add_edge(source, target, relation, span);
    }
}

pub(crate) fn stable_id(language: &str, kind: &str, canonical: &str) -> String {
    let input = format!("lexicon:v1\0{language}\0{kind}\0{canonical}");
    let digest = Sha256::digest(input.as_bytes());
    format!("sha256:{digest:x}")
}

pub(crate) fn content_id(content: &[u8]) -> String {
    let digest = Sha256::digest(content);
    format!("sha256:{digest:x}")
}

fn span_key(value: &Option<Value>) -> String {
    value.as_ref().map(Value::to_string).unwrap_or_default()
}
