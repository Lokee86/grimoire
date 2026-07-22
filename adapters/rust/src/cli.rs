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
    let jsonl = crate::orchestrator::generate(&repo)?;
    if let Some(parent) = output.parent() {
        fs::create_dir_all(parent)?;
    }
    fs::File::create(&output)?.write_all(jsonl.as_bytes())?;
    Ok(())
}
