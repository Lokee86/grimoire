use std::fmt::Write as FmtWrite;

use serde::Serialize;

use arcana::vector::{
    BuildOptions, EmbeddingClient, VectorIndexError, build_current_index_with_options,
    search_current_index, search_expected_index,
};

use crate::cli::{SemanticQueryCommand, VectorizeCommand};

pub fn run_vectorize(command: &VectorizeCommand) -> Result<String, VectorIndexError> {
    let client = EmbeddingClient::new(&command.endpoint);
    let summary = build_current_index_with_options(
        &command.state,
        &client,
        BuildOptions {
            batch_size: command.batch_size,
            batch_concurrency: command.batch_concurrency,
        },
    )?;
    Ok(format!(
        "Arcana vectors: mode={} documents={} unique_vectors={} dimensions={} embedded_vectors={} reused_vectors={} cached_snapshot={} request_count={} snapshot_bytes={} duration_ms={:.3} directory={}\n",
        summary.mode,
        summary.item_count,
        summary.unique_vectors,
        summary.dimensions,
        summary.embedded_vectors,
        summary.reused_vectors,
        summary.cached_snapshot,
        summary.request_count,
        summary.snapshot_bytes,
        summary.duration.as_secs_f64() * 1000.0,
        summary.directory.display()
    ))
}

pub fn run_semantic_query(command: &SemanticQueryCommand) -> Result<String, VectorIndexError> {
    let client = EmbeddingClient::new(&command.endpoint);
    let hits = match command.expected_snapshot.as_deref() {
        Some(expected) => search_expected_index(
            &command.state,
            expected,
            &client,
            &command.query,
            command.limit,
        )?,
        None => search_current_index(&command.state, &client, &command.query, command.limit)?,
    };
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
