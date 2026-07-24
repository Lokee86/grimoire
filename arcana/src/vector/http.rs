use std::fmt;
use std::io::{Read, Write};
use std::net::TcpStream;

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct HttpEndpoint {
    host: String,
    port: u16,
    path: String,
    authority: String,
}

impl HttpEndpoint {
    pub fn parse(url: &str) -> Result<Self, HttpError> {
        let remainder = url
            .strip_prefix("http://")
            .ok_or_else(|| HttpError::InvalidUrl(url.to_owned()))?;
        let (authority, path) = remainder
            .split_once('/')
            .map_or((remainder, "/"), |(authority, path)| {
                (authority, if path.is_empty() { "/" } else { path })
            });
        if authority.is_empty() {
            return Err(HttpError::InvalidUrl(url.to_owned()));
        }
        let (host, port) = match authority.rsplit_once(':') {
            Some((host, port)) if !host.is_empty() => (
                host.to_owned(),
                port.parse()
                    .map_err(|_| HttpError::InvalidUrl(url.to_owned()))?,
            ),
            _ => (authority.to_owned(), 80),
        };
        Ok(Self {
            host,
            port,
            path: format!("/{}", path.trim_start_matches('/')),
            authority: authority.to_owned(),
        })
    }

    pub fn post_json(&self, body: &[u8]) -> Result<Vec<u8>, HttpError> {
        let mut stream = TcpStream::connect((self.host.as_str(), self.port))?;
        write!(
            stream,
            "POST {} HTTP/1.1\r\nHost: {}\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
            self.path,
            self.authority,
            body.len()
        )?;
        stream.write_all(body)?;
        stream.flush()?;

        let mut response = Vec::new();
        stream.read_to_end(&mut response)?;
        let separator = response
            .windows(4)
            .position(|window| window == b"\r\n\r\n")
            .ok_or(HttpError::MalformedResponse)?;
        let headers = std::str::from_utf8(&response[..separator])
            .map_err(|_| HttpError::MalformedResponse)?;
        let status = headers
            .lines()
            .next()
            .and_then(|line| line.split_whitespace().nth(1))
            .and_then(|value| value.parse::<u16>().ok())
            .ok_or(HttpError::MalformedResponse)?;
        let body = response[separator + 4..].to_vec();
        if !(200..300).contains(&status) {
            return Err(HttpError::Status { status, body });
        }
        Ok(body)
    }
}

#[derive(Debug)]
pub enum HttpError {
    Io(std::io::Error),
    InvalidUrl(String),
    MalformedResponse,
    Status { status: u16, body: Vec<u8> },
}

impl fmt::Display for HttpError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Io(error) => error.fmt(formatter),
            Self::InvalidUrl(url) => write!(
                formatter,
                "unsupported embedding endpoint {url:?}; Arcana currently requires plain HTTP"
            ),
            Self::MalformedResponse => {
                formatter.write_str("embedding service returned a malformed HTTP response")
            }
            Self::Status { status, body } => write!(
                formatter,
                "embedding service returned HTTP {status}: {}",
                String::from_utf8_lossy(body)
            ),
        }
    }
}

impl std::error::Error for HttpError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(error) => Some(error),
            _ => None,
        }
    }
}

impl From<std::io::Error> for HttpError {
    fn from(error: std::io::Error) -> Self {
        Self::Io(error)
    }
}
