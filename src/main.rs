use std::{env, process::ExitCode};

use arcana_graph::PROJECT_NAME;

fn main() -> ExitCode {
    let mut args = env::args().skip(1);

    match args.next().as_deref() {
        None | Some("--help" | "-h") => {
            print_help();
            ExitCode::SUCCESS
        }
        Some("--version" | "-V") => {
            println!("arcana {}", env!("CARGO_PKG_VERSION"));
            ExitCode::SUCCESS
        }
        Some(command) => {
            eprintln!("unknown command: {command}");
            eprintln!("run `arcana --help` for usage");
            ExitCode::FAILURE
        }
    }
}

fn print_help() {
    println!("{PROJECT_NAME} repository graph engine");
    println!();
    println!("Usage: arcana [OPTIONS]");
    println!();
    println!("Options:");
    println!("  -h, --help       Print help");
    println!("  -V, --version    Print version");
}
