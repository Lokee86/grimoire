use crate::{emit, orchestrator};
use serde_json::Value;
use std::path::PathBuf;

fn fixture() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/sample")
}

fn records() -> Vec<Value> {
    orchestrator::generate(&fixture(), None, None)
        .unwrap()
        .lines()
        .map(|line| serde_json::from_str(line).unwrap())
        .collect()
}

fn node_id<'a>(records: &'a [Value], qualified_name: &str) -> &'a str {
    records
        .iter()
        .find(|record| record["record"] == "node" && record["qualified_name"] == qualified_name)
        .and_then(|record| record["id"].as_str())
        .unwrap_or_else(|| panic!("missing node {qualified_name}"))
}

fn has_edge(records: &[Value], source: &str, target: &str, relation: &str) -> bool {
    records.iter().any(|record| {
        record["record"] == "edge"
            && record["relation"] == relation
            && record["source"] == source
            && record["target"] == target
    })
}

#[test]
fn indexes_rust_declarations_imports_traits_and_local_macros() {
    let records = records();
    for (kind, name) in [
        ("type", "Service"),
        ("type", "Worker"),
        ("trait", "Runnable"),
        ("function", "generated"),
        ("function", "Data"),
        ("import", "use crate :: child"),
    ] {
        assert!(
            records.iter().any(|record| {
                record["record"] == "node"
                    && record["kind"] == kind
                    && record["name"]
                        .as_str()
                        .is_some_and(|value| value.contains(name))
            }),
            "missing {kind} {name}"
        );
    }
    assert!(records
        .iter()
        .any(|record| record["record"] == "edge" && record["relation"] == "implements"));
    assert!(records
        .iter()
        .any(|record| record["record"] == "edge" && record["relation"] == "imports"));
}

#[test]
fn resolves_inherent_field_alias_constructor_ufcs_and_macro_calls() {
    let records = records();
    let top = node_id(&records, "lexicon_fixture::lexicon_fixture::top");
    for target in [
        "lexicon_fixture::lexicon_fixture::Service::factory",
        "lexicon_fixture::lexicon_fixture::Service::run_local",
        "lexicon_fixture::lexicon_fixture::Service::Runnable::run",
        "lexicon_fixture::lexicon_fixture::child::Worker::new",
        "lexicon_fixture::lexicon_fixture::child::Worker::work",
        "lexicon_fixture::lexicon_fixture::child::helper",
        "lexicon_fixture::lexicon_fixture::generated",
    ] {
        let target = node_id(&records, target);
        assert!(
            has_edge(&records, top, target, "calls")
                || has_edge(&records, top, target, "possible-calls"),
            "missing top call to {target}"
        );
    }
    let run_local = node_id(
        &records,
        "lexicon_fixture::lexicon_fixture::Service::run_local",
    );
    let work = node_id(
        &records,
        "lexicon_fixture::lexicon_fixture::child::Worker::work",
    );
    assert!(has_edge(&records, run_local, work, "calls"));
    let factory = node_id(
        &records,
        "lexicon_fixture::lexicon_fixture::Service::factory",
    );
    let new = node_id(&records, "lexicon_fixture::lexicon_fixture::Service::new");
    assert!(has_edge(&records, factory, new, "calls"));
    let enum_build = node_id(&records, "lexicon_fixture::lexicon_fixture::enum_build");
    let variant = node_id(&records, "lexicon_fixture::lexicon_fixture::Kind::Data");
    assert!(has_edge(&records, enum_build, variant, "calls"));
}

#[test]
fn propagates_callbacks_and_generic_trait_dispatch_as_possible_calls() {
    let records = records();
    let invoke = node_id(&records, "lexicon_fixture::lexicon_fixture::invoke");
    let helper = node_id(&records, "lexicon_fixture::lexicon_fixture::helper");
    assert!(has_edge(&records, invoke, helper, "possible-calls"));
    assert!(records.iter().any(|record| {
        record["record"] == "edge"
            && record["source"] == invoke
            && record["relation"] == "possible-calls"
            && records.iter().any(|node| {
                node["record"] == "node"
                    && node["id"] == record["target"]
                    && node["name"]
                        .as_str()
                        .is_some_and(|name| name.starts_with("closure@"))
            })
    }));
    let runner = node_id(&records, "lexicon_fixture::lexicon_fixture::invoke_runner");
    let service_run = node_id(
        &records,
        "lexicon_fixture::lexicon_fixture::Service::Runnable::run",
    );
    assert!(has_edge(&records, runner, service_run, "possible-calls"));
}

#[test]
fn classifies_nested_bare_function_alias_calls_as_dynamic() {
    let records = records();
    let invoke_aliases = node_id(&records, "lexicon_fixture::lexicon_fixture::invoke_aliases");
    assert!(records.iter().any(|record| {
        record["record"] == "unresolved"
            && record["source"] == invoke_aliases
            && record["relation"] == "calls"
            && record["reason"] == "dynamic-target"
            && record["expression"] == "callback (& mut bytes)"
    }));
    assert!(!records.iter().any(|record| {
        record["record"] == "unresolved"
            && record["source"] == invoke_aliases
            && record["reason"] == "missing-target"
    }));
}

#[test]
fn propagates_callable_aliases_through_tuple_container_destructuring() {
    let records = records();
    let source = node_id(
        &records,
        "lexicon_fixture::lexicon_fixture::invoke_tuple_aliases",
    );
    let dynamic_calls: Vec<_> = records
        .iter()
        .filter(|record| {
            record["record"] == "unresolved"
                && record["source"] == source
                && record["relation"] == "calls"
                && record["reason"] == "dynamic-target"
        })
        .collect();
    assert_eq!(dynamic_calls.len(), 2);
    assert!(!records.iter().any(|record| {
        record["record"] == "unresolved"
            && record["source"] == source
            && record["reason"] == "missing-target"
    }));
}

#[test]
fn classifies_only_current_unresolved_reasons() {
    let records = records();
    let reasons: Vec<_> = records
        .iter()
        .filter(|record| record["record"] == "unresolved" && record["relation"] == "calls")
        .filter_map(|record| record["reason"].as_str())
        .collect();
    assert!(reasons.contains(&"missing-target"));
    assert!(reasons.contains(&"external-target"));
    assert!(reasons.contains(&"builtin-target"));
    for legacy in [
        "method-call",
        "associated-target",
        "macro-call",
        "generated-target",
    ] {
        assert!(!reasons.contains(&legacy));
    }
    let ambiguous = node_id(&records, "lexicon_fixture::lexicon_fixture::ambiguous");
    assert_eq!(
        records
            .iter()
            .filter(|record| record["record"] == "edge"
                && record["source"] == ambiguous
                && record["relation"] == "possible-calls")
            .count(),
        2
    );
}

#[test]
fn repeat_runs_are_byte_identical_and_paths_are_relative() {
    let first = orchestrator::generate(&fixture(), None, None).unwrap();
    assert_eq!(
        first,
        orchestrator::generate(&fixture(), None, None).unwrap()
    );
    for record in first.lines().skip(1) {
        let value: Value = serde_json::from_str(record).unwrap();
        if let Some(path) = value.get("path").and_then(Value::as_str) {
            assert!(!path.contains(":\\"));
            assert!(!path.starts_with('/'));
        }
    }
}

#[test]
fn header_and_fact_order_are_canonical() {
    let output = orchestrator::generate(&fixture(), None, None).unwrap();
    assert_eq!(
        output.lines().next().unwrap(),
        r#"{"adapter_version":"0.3.0","language":"rust","record":"lexicon","repository":"lexicon_fixture","schema_version":1}"#
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
