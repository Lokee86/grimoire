use std::env;
use std::fs::{self, File};
use std::process::ExitCode;

use arcana_graph::benchmark::{
    BenchmarkCommand, BenchmarkConfig, benchmark_usage, run_mutation_benchmark,
};
use arcana_graph::{PROJECT_NAME, PROJECT_VERSION, about};

const USAGE: &str = "Usage: arcana [OPTIONS] [COMMAND]\n\nOptions:\n    -h, --help       Print this help message\n    -V, --version    Print version information\n\nCommands:\n    benchmark        Compare overlays with packed snapshot rebuilds";

fn main() -> ExitCode {
    let mut arguments = env::args().skip(1);

    match arguments.next().as_deref() {
        None => {
            println!("{PROJECT_NAME} — {}", about());
            println!("{USAGE}");
            ExitCode::SUCCESS
        }
        Some("-h" | "--help") if arguments.next().is_none() => {
            println!("{USAGE}");
            ExitCode::SUCCESS
        }
        Some("-V" | "--version") if arguments.next().is_none() => {
            println!("{PROJECT_NAME} {PROJECT_VERSION}");
            ExitCode::SUCCESS
        }
        Some("benchmark") => run_benchmark_command(arguments.collect()),
        Some(argument) => {
            eprintln!("arcana: unexpected argument '{argument}'\n\n{USAGE}");
            ExitCode::from(2)
        }
    }
}

fn run_benchmark_command(arguments: Vec<String>) -> ExitCode {
    if matches!(arguments.as_slice(), [argument] if argument == "-h" || argument == "--help") {
        println!("{}", benchmark_usage());
        return ExitCode::SUCCESS;
    }

    let command = match BenchmarkCommand::parse(arguments.iter().map(String::as_str)) {
        Ok(command) => command,
        Err(error) => {
            eprintln!("arcana benchmark: {error}\n\n{}", benchmark_usage());
            return ExitCode::from(2);
        }
    };
    let config = BenchmarkConfig::new(
        command.graph,
        command.query_count,
        command.sample_count as usize,
        &command.work_dir,
        command.keep_files,
    );
    let report = match run_mutation_benchmark(&config) {
        Ok(report) => report,
        Err(error) => {
            eprintln!("arcana benchmark: {error}");
            return ExitCode::FAILURE;
        }
    };

    print!("{}", report.human_summary());
    if let Some(path) = command.csv_path {
        if let Err(error) = write_csv(&report, &path) {
            eprintln!("arcana benchmark: write {}: {error}", path.display());
            return ExitCode::FAILURE;
        }
        println!("raw samples: {}", path.display());
    }
    ExitCode::SUCCESS
}

fn write_csv(
    report: &arcana_graph::benchmark::BenchmarkReport,
    path: &std::path::Path,
) -> Result<(), arcana_graph::benchmark::BenchmarkError> {
    if let Some(parent) = path
        .parent()
        .filter(|parent| !parent.as_os_str().is_empty())
    {
        fs::create_dir_all(parent)?;
    }
    report.write_csv(File::create(path)?)
}
