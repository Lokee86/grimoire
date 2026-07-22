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
        processed: HashSet::new(),
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
            r#"{"adapter_version":"0.1.0","language":"rust","record":"lexicon","repository":"lexicon_fixture","schema_version":1}"#
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
}
