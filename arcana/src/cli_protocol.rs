use std::io;

use arcana::protocol::{ProtocolError, ProtocolSnapshot, serve_jsonl};

use crate::cli::ProtocolCommand;

pub fn run_protocol(command: &ProtocolCommand) -> Result<(), ProtocolError> {
    let snapshot = ProtocolSnapshot::open(&command.snapshot)?;
    let stdin = io::stdin();
    let stdout = io::stdout();
    serve_jsonl(&snapshot, stdin.lock(), stdout.lock())
}
