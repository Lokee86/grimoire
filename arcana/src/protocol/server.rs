use std::io::{BufRead, Write};

use super::{ProtocolError, ProtocolSnapshot};

/// Serves one JSON response for every input line until EOF.
pub fn serve_jsonl(
    snapshot: &ProtocolSnapshot,
    mut input: impl BufRead,
    mut output: impl Write,
) -> Result<(), ProtocolError> {
    let mut line = String::new();
    loop {
        line.clear();
        if input.read_line(&mut line)? == 0 {
            return Ok(());
        }
        let request = line.trim_end_matches(['\r', '\n']);
        let response = snapshot.handle_line(request);
        serde_json::to_writer(&mut output, &response)?;
        output.write_all(b"\n")?;
        output.flush()?;
    }
}
