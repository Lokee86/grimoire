use anyhow::{Context as AnyhowContext, Result};
use clap::Parser;
use std::fs;
use std::io::Write;
use std::path::PathBuf;

#[derive(Debug, Parser)]
#[command(
    name = "lexicon-rust-adapter",
    about = "Emit Lexicon facts v1 for a Rust repository"
)]
pub(crate) struct Args {
    #[arg(long)]
    repo: PathBuf,
    #[arg(long)]
    output: PathBuf,
    #[arg(long = "changed-file")]
    changed_files: Option<Vec<PathBuf>>,
    #[arg(long = "removed-file")]
    removed_files: Option<Vec<PathBuf>>,
}

pub(crate) fn run() -> Result<()> {
    let args = Args::parse();
    let repo = args
        .repo
        .canonicalize()
        .with_context(|| format!("cannot resolve repository {}", args.repo.display()))?;
    let output = if args.output.is_absolute() {
        args.output
    } else {
        std::env::current_dir()?.join(args.output)
    };
    let changed_files = normalize_paths(args.changed_files);
    let removed_files = normalize_paths(args.removed_files);
    let jsonl =
        crate::orchestrator::generate(&repo, changed_files.as_deref(), removed_files.as_deref())?;
    if let Some(parent) = output.parent() {
        fs::create_dir_all(parent)?;
    }
    fs::File::create(&output)?.write_all(jsonl.as_bytes())?;
    Ok(())
}

fn normalize_paths(paths: Option<Vec<PathBuf>>) -> Option<Vec<String>> {
    paths.map(|values| {
        values
            .into_iter()
            .map(|path| path.to_string_lossy().replace('\\', "/"))
            .collect()
    })
}
