mod call_model;
mod call_resolution;
mod call_support;
mod cli;
mod contract;
mod declarations;
mod discovery;
mod emit;
mod expr_eval;
mod expr_values;
mod extractor;
mod flow;
mod function_index;
mod implementations;
mod imports;
mod items;
mod model;
mod orchestrator;
mod parser;
mod paths;
mod relationships;
mod resolve;
mod semantic;
mod syntax;
mod type_resolution;

fn main() -> anyhow::Result<()> {
    cli::run()
}

#[cfg(test)]
mod tests;
