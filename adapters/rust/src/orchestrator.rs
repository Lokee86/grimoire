use crate::contract::Facts;
use crate::discovery;
use crate::emit;
use crate::extractor;
use crate::model::Context;
use crate::parser;
use anyhow::Result;
use std::collections::{BTreeMap, HashSet};
use std::path::Path;

pub(crate) fn generate(repo: &Path) -> Result<String> {
    let metadata = discovery::load_metadata(repo)?;
    let repository = discovery::repository_identity(repo, &metadata);
    let sources = parser::parse_sources(repo)?;
    let mut context = Context {
        repo: repo.to_path_buf(),
        repository: repository.clone(),
        sources,
        facts: Facts::new(),
        crates: Vec::new(),
        modules: BTreeMap::new(),
        symbols: BTreeMap::new(),
        types: BTreeMap::new(),
        traits: BTreeMap::new(),
        inherent_methods: BTreeMap::new(),
        processed: HashSet::new(),
        pending_calls: Vec::new(),
    };
    discovery::add_repository_and_files(&mut context);
    discovery::add_crates(&mut context, &metadata);
    extractor::extract(&mut context);
    emit::render(&context, &repository)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;
    use std::path::PathBuf;

    fn fixture() -> PathBuf {
        PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/sample")
    }

    #[test]
    fn emits_declarations_relationships_and_unresolved_macro() {
        let records: Vec<Value> = generate(&fixture())
            .unwrap()
            .lines()
            .map(|line| serde_json::from_str(line).unwrap())
            .collect();
        assert_eq!(records[0]["language"], "rust");
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "type"
            && record["name"] == "Service"));
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "trait"
            && record["name"] == "Runnable"));
        assert!(records
            .iter()
            .any(|record| record["record"] == "node" && record["kind"] == "import"));
        assert!(records
            .iter()
            .any(|record| record["record"] == "edge" && record["relation"] == "implements"));
        assert!(records.iter().any(
            |record| record["record"] == "unresolved" && record["reason"] == "generated-target"
        ));
        assert!(records
            .iter()
            .any(|record| record["record"] == "edge" && record["relation"] == "imports"));
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "module"
            && record["name"] == "child"));
    }

    #[test]
    fn emits_conservative_free_function_calls_and_call_unresolved_records() {
        let records: Vec<Value> = generate(&fixture())
            .unwrap()
            .lines()
            .map(|line| serde_json::from_str(line).unwrap())
            .collect();
        let node_id = |qualified_name: &str| {
            records
                .iter()
                .find(|record| {
                    record["record"] == "node" && record["qualified_name"] == qualified_name
                })
                .and_then(|record| record["id"].as_str())
                .unwrap()
        };
        let build_id = node_id("lexicon_fixture::lexicon_fixture::build");
        let helper_id = node_id("lexicon_fixture::lexicon_fixture::helper");
        let new_id = node_id("lexicon_fixture::lexicon_fixture::Service::new");
        let run_id = node_id("lexicon_fixture::lexicon_fixture::Service::Runnable::run");
        assert!(records.iter().any(|record| {
            record["record"] == "edge"
                && record["relation"] == "calls"
                && record["source"] == build_id
                && record["target"] == helper_id
        }));
        assert!(records.iter().any(|record| {
            record["record"] == "edge"
                && record["relation"] == "calls"
                && record["source"] == run_id
                && record["target"] == build_id
        }));
        assert!(records.iter().any(|record| {
            record["record"] == "edge"
                && record["relation"] == "calls"
                && record["source"] == build_id
                && record["target"] == new_id
        }));
        for reason in [
            "missing-target",
            "external-target",
            "ambiguous-target",
            "method-call",
            "macro-call",
            "unsupported-form",
        ] {
            assert!(
                records.iter().any(|record| {
                    record["record"] == "unresolved"
                        && record["relation"] == "calls"
                        && record["reason"] == reason
                }),
                "missing calls unresolved reason: {reason}"
            );
        }
    }

    #[test]
    fn repeat_runs_are_byte_identical_and_paths_are_relative() {
        let first = generate(&fixture()).unwrap();
        assert_eq!(first, generate(&fixture()).unwrap());
        for record in first.lines().skip(1) {
            let value: Value = serde_json::from_str(record).unwrap();
            if let Some(path) = value.get("path").and_then(Value::as_str) {
                assert!(!path.contains(":\\"));
                assert!(!path.starts_with('/'));
            }
            if let Some(span) = value.get("span") {
                assert!(!span["path"].as_str().unwrap_or("").contains(':'));
            }
        }
    }

    #[test]
    fn header_and_fact_order_are_canonical() {
        let output = generate(&fixture()).unwrap();
        assert_eq!(
            output.lines().next().unwrap(),
            r#"{"adapter_version":"0.2.0","language":"rust","record":"lexicon","repository":"lexicon_fixture","schema_version":1}"#
        );
        let records: Vec<Value> = output
            .lines()
            .skip(1)
            .map(|line| serde_json::from_str(line).unwrap())
            .collect();
        let mut previous = None;
        for record in records {
            let key = emit::fact_sort_key(&record);
            if let Some(previous) = &previous {
                assert!(previous <= &key);
            }
            previous = Some(key);
        }
    }

    #[test]
    fn numeric_call_span_order_places_line_98_before_line_102() {
        let mut records = vec![
            serde_json::json!({
                "record": "edge",
                "source": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "target": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
                "relation": "calls",
                "span": {
                    "path": "src/storage/format.rs",
                    "start_line": 102,
                    "start_column": 23,
                    "end_line": 102,
                    "end_column": 40
                }
            }),
            serde_json::json!({
                "record": "edge",
                "source": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "target": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
                "relation": "calls",
                "span": {
                    "path": "src/storage/format.rs",
                    "start_line": 98,
                    "start_column": 23,
                    "end_line": 98,
                    "end_column": 40
                }
            }),
        ];
        records.sort_by_key(emit::fact_sort_key);
        assert_eq!(records[0]["span"]["start_line"].as_u64(), Some(98));
    }
}
