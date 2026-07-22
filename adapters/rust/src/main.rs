mod cli;
mod contract;
mod discovery;
mod emit;
mod extractor;
mod items;
mod model;
mod orchestrator;
mod parser;
mod paths;
mod relationships;
mod resolve;

fn main() -> anyhow::Result<()> {
    cli::run()
}
