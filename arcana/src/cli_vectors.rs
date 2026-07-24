use std::fmt::Write as FmtWrite;

use serde::Serialize;

use arcana::vector::{
    EmbeddingClient, VectorIndexError, build_current_index, search_current_index,
};

use crate::cli::{SemanticQueryCommand, VectorizeCommand};

pub fn run_vectorize(command: &VectorizeCommand) -> Result<String, VectorIndexError> {
    let client = EmbeddingClient::new(&command.endpoint);
    let summary = build_current_index(&command.state, &client, command.batch_size)?;
    Ok(format!(
        "Arcana vectors: mode={} nodes={} dimensions={} directory={}\n",
        summary.mode,
        summary.item_count,
        summary.dimensions,
        summary.directory.display()
    ))
}

pub fn run_semantic_query(command: &SemanticQueryCommand) -> Result<String, VectorIndexError> {
    let client = EmbeddingClient::new(&command.endpoint);
    let hits = search_current_index(&command.state, &client, &command.query, command.limit)?;
    if command.json {
        let mut output = serde_json::to_string(&SemanticMatches { matches: &hits })?;
        output.push('\n');
        return Ok(output);
    }
    let mut output = String::new();
    writeln!(output, "semantic matches: {}", hits.len()).unwrap();
    for hit in hits {
        writeln!(
            output,
            "score={:.6} key={} kind={} path={:?} name={:?}",
            hit.score, hit.node_key, hit.kind, hit.path, hit.name
        )
        .unwrap();
    }
    Ok(output)
}

#[derive(Serialize)]
struct SemanticMatches<'a> {
    matches: &'a [arcana::vector::SearchHit],
}
