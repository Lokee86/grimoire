use super::*;

fn id(digit: char) -> String {
    format!("sha256:{}", digit.to_string().repeat(64))
}

#[test]
fn parses_lexicon_jsonl_into_repository_facts() {
    let repository = id('1');
    let caller = id('2');
    let callee = id('3');
    let stream = format!(
        concat!(
            "{{\"adapter_version\":\"0.1.0\",\"language\":\"go\",\"mode\":\"full\",\"record\":\"lexicon\",\"repository\":\"example.com/demo\",\"schema_version\":1}}\n",
            "{{\"id\":\"{repository}\",\"kind\":\"repository\",\"name\":\"example.com/demo\",\"path\":\".\",\"qualified_name\":\"example.com/demo\",\"record\":\"node\"}}\n",
            "{{\"id\":\"{caller}\",\"kind\":\"function\",\"name\":\"caller\",\"owner\":\"src/lib.go\",\"path\":\"src/lib.go\",\"qualified_name\":\"example.com/demo.caller\",\"record\":\"node\",\"span\":{{\"end_column\":2,\"end_line\":2,\"path\":\"src/lib.go\",\"start_column\":1,\"start_line\":1}}}}\n",
            "{{\"id\":\"{callee}\",\"kind\":\"function\",\"name\":\"callee\",\"owner\":\"src/lib.go\",\"path\":\"src/lib.go\",\"qualified_name\":\"example.com/demo.callee\",\"record\":\"node\"}}\n",
            "{{\"owner\":\"src/lib.go\",\"record\":\"edge\",\"relation\":\"possible-calls\",\"source\":\"{caller}\",\"target\":\"{callee}\"}}\n",
            "{{\"expression\":\"dynamic()\",\"owner\":\"src/lib.go\",\"reason\":\"dynamic-target\",\"record\":\"unresolved\",\"relation\":\"calls\",\"source\":\"{caller}\"}}\n"
        ),
        repository = repository,
        caller = caller,
        callee = callee,
    );

    let facts = parse_facts(&stream).unwrap();
    assert_eq!(facts.nodes.len(), 3);
    assert_eq!(
        facts.nodes[0].external_identity.as_deref(),
        Some(repository.as_str())
    );
    assert_eq!(
        facts.nodes[1].external_identity.as_deref(),
        Some(caller.as_str())
    );
    assert_eq!(facts.edges.len(), 1);
    assert_eq!(facts.edges[0].relation, RelationKind::PossibleCalls);
    assert_eq!(facts.unresolved.len(), 1);
    assert_eq!(facts.unresolved[0].reason, UnresolvedReason::DynamicTarget);
    assert_eq!(facts.nodes[1].span.as_ref().unwrap().path, "src/lib.go");
}

#[test]
fn rejects_scoped_incremental_stream_as_complete_input() {
    let stream = "{\"adapter_version\":\"0.1.0\",\"changed_files\":[\"src/lib.go\"],\"language\":\"go\",\"mode\":\"incremental\",\"record\":\"lexicon\",\"removed_files\":[],\"repository\":\"example.com/demo\",\"schema_version\":1}\n";
    assert_eq!(parse_facts(stream), Err(FactFileError::InvalidHeader));
}

#[test]
fn rejects_lexicon_edges_to_unknown_nodes() {
    let caller = id('2');
    let missing = id('3');
    let stream = format!(
        concat!(
            "{{\"adapter_version\":\"0.1.0\",\"language\":\"go\",\"record\":\"lexicon\",\"repository\":\"example.com/demo\",\"schema_version\":1}}\n",
            "{{\"id\":\"{caller}\",\"kind\":\"function\",\"name\":\"caller\",\"path\":\"src/lib.go\",\"qualified_name\":\"caller\",\"record\":\"node\"}}\n",
            "{{\"record\":\"edge\",\"relation\":\"calls\",\"source\":\"{caller}\",\"target\":\"{missing}\"}}\n"
        ),
        caller = caller,
        missing = missing,
    );
    assert_eq!(
        parse_facts(&stream),
        Err(FactFileError::MalformedLine { line: 3 })
    );
}

#[test]
fn parses_typescript_exports_and_gdscript_signals() {
    let repository = id('1');
    let export = id('2');
    let signal = id('3');
    let caller = id('4');
    let callee = id('5');
    let stream = format!(
        concat!(
            "{{\"adapter_version\":\"0.1.0\",\"language\":\"mixed\",\"mode\":\"full\",\"record\":\"lexicon\",\"repository\":\"example/demo\",\"schema_version\":1}}\n",
            "{{\"id\":\"{repository}\",\"kind\":\"repository\",\"name\":\"example/demo\",\"path\":\".lexicon-repository\",\"qualified_name\":\"example/demo\",\"record\":\"node\"}}\n",
            "{{\"id\":\"{export}\",\"kind\":\"export\",\"name\":\"run\",\"path\":\"src/api.ts\",\"qualified_name\":\"src/api.ts::run\",\"record\":\"node\"}}\n",
            "{{\"id\":\"{signal}\",\"kind\":\"signal\",\"name\":\"finished\",\"path\":\"src/player.gd\",\"qualified_name\":\"Player.finished\",\"record\":\"node\"}}\n",
            "{{\"id\":\"{caller}\",\"kind\":\"function\",\"name\":\"caller\",\"path\":\"src/api.ts\",\"qualified_name\":\"src/api.ts::caller\",\"record\":\"node\"}}\n",
            "{{\"id\":\"{callee}\",\"kind\":\"method\",\"name\":\"callee\",\"path\":\"src/player.gd\",\"qualified_name\":\"Player.callee\",\"record\":\"node\"}}\n",
            "{{\"record\":\"edge\",\"relation\":\"calls\",\"source\":\"{caller}\",\"target\":\"{callee}\"}}\n"
        ),
        repository = repository,
        export = export,
        signal = signal,
        caller = caller,
        callee = callee,
    );

    let facts = parse_facts(&stream).unwrap();
    assert!(facts.nodes.iter().any(|node| node.kind == NodeKind::Export));
    assert!(facts.nodes.iter().any(|node| node.kind == NodeKind::Signal));
    assert_eq!(facts.edges.len(), 1);
    assert_eq!(facts.edges[0].relation, RelationKind::Calls);
}
