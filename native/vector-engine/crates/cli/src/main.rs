use std::{collections::HashMap, env, time::Instant};

use grimoire_vector_core::{ObjectStore, Snapshot, ingest_jsonl, materialize_jsonl, search};
use serde_json::json;

fn main() {
    if let Err(error) = run() {
        eprintln!("error: {error}");
        std::process::exit(1);
    }
}

fn run() -> Result<(), Box<dyn std::error::Error>> {
    let mut args = env::args().skip(1);
    let command = args
        .next()
        .ok_or("expected command: ingest, build, inspect, or search")?;
    let flags = flags(args.collect())?;
    match command.as_str() {
        "ingest" => {
            let store = ObjectStore::new(required(&flags, "store")?);
            let count = ingest_jsonl(
                &store,
                required(&flags, "model")?,
                required(&flags, "input")?,
            )?;
            println!("{}", json!({"ingested": count}));
        }
        "build" => {
            let store = ObjectStore::new(required(&flags, "store")?);
            let info = materialize_jsonl(
                &store,
                required(&flags, "model")?,
                required(&flags, "manifest")?,
                required(&flags, "snapshot")?,
            )?;
            println!(
                "{}",
                json!({"identity": info.identity, "model": info.model, "dimensions": info.dimensions, "count": info.count})
            );
        }
        "inspect" => {
            let snapshot = Snapshot::open(required(&flags, "snapshot")?)?;
            let info = snapshot.info();
            println!(
                "{}",
                json!({"identity": info.identity, "model": info.model, "dimensions": info.dimensions, "count": info.count})
            );
        }
        "search" => {
            let snapshot = Snapshot::open(required(&flags, "snapshot")?)?;
            let query = vector(required(&flags, "query")?)?;
            let top_k = flags
                .get("top-k")
                .map(|value| value.parse())
                .transpose()?
                .unwrap_or(10);
            let started = Instant::now();
            let hits = search(&snapshot, &query, top_k)?;
            println!(
                "{}",
                json!({"elapsed_micros": started.elapsed().as_micros(), "hits": hits.iter().map(|hit| json!({"id": hit.id, "score": hit.score, "index": hit.index})).collect::<Vec<_>>() })
            );
        }
        _ => return Err(format!("unknown command {command:?}").into()),
    }
    Ok(())
}

fn flags(args: Vec<String>) -> Result<HashMap<String, String>, Box<dyn std::error::Error>> {
    let mut result = HashMap::new();
    let mut index = 0;
    while index < args.len() {
        let name = args[index]
            .strip_prefix("--")
            .ok_or("flags must start with --")?;
        let value = args.get(index + 1).ok_or("flag value is missing")?;
        result.insert(name.to_owned(), value.to_owned());
        index += 2;
    }
    Ok(result)
}

fn required<'a>(
    flags: &'a HashMap<String, String>,
    name: &str,
) -> Result<&'a str, Box<dyn std::error::Error>> {
    flags
        .get(name)
        .map(String::as_str)
        .ok_or_else(|| format!("--{name} is required").into())
}

fn vector(value: &str) -> Result<Vec<f32>, Box<dyn std::error::Error>> {
    value
        .split(',')
        .map(|part| part.trim().parse::<f32>().map_err(Into::into))
        .collect()
}
