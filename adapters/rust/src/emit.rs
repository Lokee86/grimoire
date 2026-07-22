use crate::contract::JsonMap;
use crate::model::Context;
use anyhow::{Context as AnyhowContext, Result};
use serde_json::Value;

pub(crate) fn render(context: &Context, repository: &str) -> Result<String> {
    let mut header = JsonMap::new();
    header.insert("adapter_version".into(), Value::String("0.1.0".into()));
    header.insert("language".into(), Value::String("rust".into()));
    header.insert("record".into(), Value::String("lexicon".into()));
    header.insert("repository".into(), Value::String(repository.into()));
    header.insert("schema_version".into(), Value::Number(1.into()));
    let mut values = vec![Value::Object(header)];
    let mut facts: Vec<Value> = context
        .facts
        .nodes
        .values()
        .chain(context.facts.edges.values())
        .chain(context.facts.unresolved.values())
        .cloned()
        .collect();
    facts.sort_by_key(fact_sort_key);
    values.extend(facts);
    values
        .into_iter()
        .map(|value| serde_json::to_string(&value).context("cannot serialize fact"))
        .collect::<Result<Vec<_>>>()
        .map(|lines| format!("{}\n", lines.join("\n")))
}

pub(crate) fn fact_sort_key(value: &Value) -> (u8, String, String, String, String, Vec<String>) {
    let object = value.as_object().expect("fact records are objects");
    match object.get("record").and_then(Value::as_str).unwrap_or("") {
        "node" => (
            0,
            field(object, "id"),
            field(object, "kind"),
            field(object, "path"),
            field(object, "qualified_name"),
            Vec::new(),
        ),
        "edge" => (
            1,
            field(object, "source"),
            field(object, "target"),
            field(object, "relation"),
            String::new(),
            span_sort_key(object.get("span")),
        ),
        _ => (
            2,
            field(object, "source"),
            field(object, "relation"),
            field(object, "expression"),
            field(object, "reason"),
            span_sort_key(object.get("span")),
        ),
    }
}

fn field(object: &JsonMap, name: &str) -> String {
    object
        .get(name)
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string()
}

fn span_sort_key(value: Option<&Value>) -> Vec<String> {
    let Some(span) = value.and_then(Value::as_object) else {
        return vec![String::new(); 5];
    };
    [
        "path",
        "start_line",
        "start_column",
        "end_line",
        "end_column",
    ]
    .iter()
    .map(|key| match span.get(*key) {
        Some(Value::String(value)) => value.clone(),
        Some(value) => value.to_string(),
        None => String::new(),
    })
    .collect()
}
