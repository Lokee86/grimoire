use std::env;
use std::process::ExitCode;

use arcana_graph::{PROJECT_NAME, PROJECT_VERSION, about};

const USAGE: &str = "Usage: arcana [OPTIONS]\n\nOptions:\n    -h, --help       Print this help message\n    -V, --version    Print version information";

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
        Some(argument) => {
            eprintln!("arcana: unexpected argument '{argument}'\n\n{USAGE}");
            ExitCode::from(2)
        }
    }
}
