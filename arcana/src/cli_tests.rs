use std::fs;
use std::path::PathBuf;
use std::sync::atomic::{AtomicUsize, Ordering};

use arcana::repository::{
    EdgeFact, NodeFact, NodeKey, NodeKind, RelationKind, RepositoryFacts, UnresolvedReason,
    UnresolvedReferenceFact, encode_facts,
};

use super::cli::{self, CliParseError, Command};
use super::cli_commands::run_import_facts;
use super::cli_query::run_query;

#[test]
fn parses_import_and_query_arguments() {
    let command = cli::parse([
        "import-facts".to_owned(),
        "--facts=facts.tsv".to_owned(),
        "--output".to_owned(),
        "out".to_owned(),
    ])
    .expect("import arguments should parse");
    let Command::ImportFacts(command) = command else {
        panic!("wrong command")
    };
    assert_eq!(command.facts, PathBuf::from("facts.tsv"));
    assert_eq!(command.output, PathBuf::from("out"));
    assert_eq!(command.adapter_name, "manual");
    assert_eq!(command.adapter_version, "1");

    let command = cli::parse([
        "update-facts".to_owned(),
        "--base=old/repository.manifest".to_owned(),
        "--facts".to_owned(),
        "replacement.tsv".to_owned(),
        "--changed".to_owned(),
        "src/lib.go".to_owned(),
        "--output".to_owned(),
        "next".to_owned(),
    ])
    .expect("update arguments should parse");
    let Command::UpdateFacts(command) = command else {
        panic!("wrong command")
    };
    assert_eq!(command.changed, vec!["src/lib.go"]);
    assert_eq!(command.output, PathBuf::from("next"));

    let command = cli::parse([
        "sync".to_owned(),
        "--lexicon".to_owned(),
        "facts".to_owned(),
        "--state=graph-state".to_owned(),
        "--register".to_owned(),
    ])
    .expect("sync arguments should parse");
    let Command::Sync(command) = command else {
        panic!("wrong command")
    };
    assert_eq!(command.lexicon, PathBuf::from("facts"));
    assert_eq!(command.state, PathBuf::from("graph-state"));
    assert!(command.register);

    let command = cli::parse([
        "query".to_owned(),
        "--graph".to_owned(),
        "graph.arcana".to_owned(),
        "--catalogue=catalogue.tsv".to_owned(),
        "--name".to_owned(),
        "caller".to_owned(),
        "--reverse".to_owned(),
        "--relation=calls".to_owned(),
    ])
    .expect("query arguments should parse");
    let Command::Query(command) = command else {
        panic!("wrong command")
    };
    assert!(command.reverse);
    assert_eq!(command.name, "caller");
    assert_eq!(command.relation, Some(RelationKind::Calls));

    let command = cli::parse([
        "protocol".to_owned(),
        "--snapshot".to_owned(),
        "target/repository-index".to_owned(),
    ])
    .expect("protocol arguments should parse");
    let Command::Protocol(command) = command else {
        panic!("wrong command")
    };
    assert_eq!(command.snapshot, PathBuf::from("target/repository-index"));
}

#[test]
fn reports_cli_argument_errors() {
    assert!(matches!(
        cli::parse(["import-facts".to_owned(), "--facts".to_owned()]),
        Err(CliParseError::MissingValue(option)) if option == "--facts"
    ));
    assert!(matches!(
        cli::parse(["query".to_owned(), "--graph".to_owned(), "graph".to_owned()]),
        Err(CliParseError::MissingRequired("--catalogue"))
    ));
    assert!(matches!(
        cli::parse(["query".to_owned(), "--graph".to_owned(), "g".to_owned(), "--catalogue".to_owned(), "c".to_owned(), "--name".to_owned(), "x".to_owned(), "--relation".to_owned(), "nope".to_owned()]),
        Err(CliParseError::InvalidRelation(relation)) if relation == "nope"
    ));
}

#[test]
fn import_and_query_round_trip() {
    let directory = TestDirectory::new();
    let facts_path = directory.path.join("facts.tsv");
    let output = directory.path.join("graph-output");
    let caller = NodeKey::from_u64(1);
    let callee = NodeKey::from_u64(2);
    let facts = RepositoryFacts {
        nodes: vec![
            NodeFact {
                key: caller,
                external_identity: None,
                kind: NodeKind::Function,
                path: "src/lib.go".to_owned(),
                name: "caller".to_owned(),
                content_id: None,
                span: None,
            },
            NodeFact {
                key: callee,
                external_identity: None,
                kind: NodeKind::Function,
                path: "src/lib.go".to_owned(),
                name: "callee".to_owned(),
                content_id: None,
                span: None,
            },
        ],
        edges: vec![EdgeFact {
            source: caller,
            target: callee,
            relation: RelationKind::Calls,
            span: None,
        }],
        unresolved: vec![UnresolvedReferenceFact {
            source: caller,
            relation: RelationKind::Calls,
            expression: "pkg.Call".to_owned(),
            candidate_namespace: Some("pkg".to_owned()),
            candidate_name: Some("Call".to_owned()),
            reason: UnresolvedReason::UnsupportedForm,
            span: None,
        }],
    };
    fs::write(&facts_path, encode_facts(&facts)).unwrap();

    let summary = run_import_facts(&super::cli::ImportFactsCommand {
        facts: facts_path,
        output: output.clone(),
        adapter_name: "test".to_owned(),
        adapter_version: "1".to_owned(),
    })
    .unwrap();
    assert!(
        summary.contains("nodes=2")
            && summary.contains("edges=1")
            && summary.contains("unresolved=1")
    );
    assert!(output.join("graph.arcana").is_file());
    assert!(output.join("catalogue.tsv").is_file());
    let unresolved_path = output.join("unresolved.tsv");
    assert!(unresolved_path.is_file());
    let persisted = RepositoryFacts::parse(&fs::read_to_string(unresolved_path).unwrap()).unwrap();
    assert_eq!(persisted.unresolved, facts.unresolved);
    assert!(
        run_import_facts(&super::cli::ImportFactsCommand {
            facts: directory.path.join("facts.tsv"),
            output: output.clone(),
            adapter_name: "test".to_owned(),
            adapter_version: "1".to_owned()
        })
        .is_err()
    );

    let result = run_query(&super::cli::QueryCommand {
        graph: output.join("graph.arcana"),
        catalogue: output.join("catalogue.tsv"),
        name: "caller".to_owned(),
        reverse: false,
        relation: Some(RelationKind::Calls),
    })
    .unwrap();
    assert!(
        result.contains("node_id=0")
            && result.contains("relation=calls")
            && result.contains("callee")
    );
    let not_found = run_query(&super::cli::QueryCommand {
        graph: output.join("graph.arcana"),
        catalogue: output.join("catalogue.tsv"),
        name: "missing".to_owned(),
        reverse: false,
        relation: None,
    })
    .unwrap();
    assert!(not_found.contains("no exact-name matches"));
}

struct TestDirectory {
    path: PathBuf,
}

impl TestDirectory {
    fn new() -> Self {
        static SEQUENCE: AtomicUsize = AtomicUsize::new(0);
        let path = std::env::temp_dir().join(format!(
            "arcana-cli-test-{}-{}",
            std::process::id(),
            SEQUENCE.fetch_add(1, Ordering::Relaxed)
        ));
        fs::create_dir(&path).unwrap();
        Self { path }
    }
}

impl Drop for TestDirectory {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.path);
    }
}
