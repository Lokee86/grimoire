use serde::{Deserialize, Serialize};

use super::error::EmbeddingError;
use super::http::HttpEndpoint;

pub const DEFAULT_ENDPOINT: &str = "http://127.0.0.1:9876/v1";
pub const DEFAULT_MODEL: &str = "Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0";
pub const DEFAULT_IDENTITY: &str = "qwen3-embedding-0.6b-q8_0-512d";
pub const DEFAULT_DIMENSIONS: usize = 512;
const QUERY_INSTRUCTION: &str = "Given a software development query, retrieve relevant source code and documentation from a repository";

pub trait Embedder {
    fn model(&self) -> &str;
    fn identity(&self) -> &str;
    fn dimensions(&self) -> usize;
    fn embed_documents(&self, documents: &[String]) -> Result<Vec<Vec<f32>>, EmbeddingError>;
    fn embed_query(&self, query: &str) -> Result<Vec<f32>, EmbeddingError>;
}

#[derive(Clone, Debug)]
pub struct EmbeddingClient {
    endpoint: String,
    model: String,
    identity: String,
    dimensions: usize,
}

impl EmbeddingClient {
    pub fn new(endpoint: impl Into<String>) -> Self {
        Self {
            endpoint: endpoint.into(),
            model: DEFAULT_MODEL.to_owned(),
            identity: DEFAULT_IDENTITY.to_owned(),
            dimensions: DEFAULT_DIMENSIONS,
        }
    }

    fn embeddings_endpoint(&self) -> Result<HttpEndpoint, EmbeddingError> {
        let endpoint = self.endpoint.trim_end_matches('/');
        let url = if endpoint.ends_with("/embeddings") {
            endpoint.to_owned()
        } else {
            format!("{endpoint}/embeddings")
        };
        Ok(HttpEndpoint::parse(&url)?)
    }

    fn embed(&self, input: &[String]) -> Result<Vec<Vec<f32>>, EmbeddingError> {
        if input.is_empty() || input.iter().any(|value| value.trim().is_empty()) {
            return Err(EmbeddingError::InvalidInput);
        }
        let request = EmbeddingsRequest {
            input,
            model: &self.model,
            encoding_format: "float",
        };
        let body = serde_json::to_vec(&request)?;
        let response = self.embeddings_endpoint()?.post_json(&body)?;
        let decoded: EmbeddingsResponse = serde_json::from_slice(&response)?;
        if let Some(error) = decoded.error {
            return Err(EmbeddingError::Service(error.message));
        }
        if decoded.data.len() != input.len() {
            return Err(EmbeddingError::WrongVectorCount {
                expected: input.len(),
                found: decoded.data.len(),
            });
        }
        let mut vectors = vec![None; input.len()];
        for item in decoded.data {
            if item.index >= vectors.len() || vectors[item.index].is_some() {
                return Err(EmbeddingError::InvalidIndex(item.index));
            }
            vectors[item.index] = Some(normalize_truncated(item.embedding, self.dimensions)?);
        }
        vectors
            .into_iter()
            .enumerate()
            .map(|(index, vector)| vector.ok_or(EmbeddingError::InvalidIndex(index)))
            .collect()
    }
}

impl Default for EmbeddingClient {
    fn default() -> Self {
        Self::new(DEFAULT_ENDPOINT)
    }
}

impl Embedder for EmbeddingClient {
    fn model(&self) -> &str {
        &self.model
    }

    fn identity(&self) -> &str {
        &self.identity
    }

    fn dimensions(&self) -> usize {
        self.dimensions
    }

    fn embed_documents(&self, documents: &[String]) -> Result<Vec<Vec<f32>>, EmbeddingError> {
        self.embed(documents)
    }

    fn embed_query(&self, query: &str) -> Result<Vec<f32>, EmbeddingError> {
        let input = format!("Instruct: {QUERY_INSTRUCTION}\nQuery:{}", query.trim());
        self.embed(&[input]).map(|mut vectors| vectors.remove(0))
    }
}

fn normalize_truncated(input: Vec<f64>, dimensions: usize) -> Result<Vec<f32>, EmbeddingError> {
    if input.len() < dimensions {
        return Err(EmbeddingError::WrongDimensions {
            expected: dimensions,
            found: input.len(),
        });
    }
    let norm_squared = input[..dimensions].iter().try_fold(0.0, |sum, value| {
        if value.is_finite() {
            Ok(sum + value * value)
        } else {
            Err(EmbeddingError::NonFiniteVector)
        }
    })?;
    if norm_squared == 0.0 {
        return Err(EmbeddingError::ZeroVector);
    }
    let norm = norm_squared.sqrt();
    Ok(input[..dimensions]
        .iter()
        .map(|value| (value / norm) as f32)
        .collect())
}

#[derive(Serialize)]
struct EmbeddingsRequest<'a> {
    input: &'a [String],
    model: &'a str,
    encoding_format: &'static str,
}

#[derive(Deserialize)]
struct EmbeddingsResponse {
    #[serde(default)]
    data: Vec<EmbeddingItem>,
    error: Option<ServiceError>,
}

#[derive(Deserialize)]
struct EmbeddingItem {
    index: usize,
    embedding: Vec<f64>,
}

#[derive(Deserialize)]
struct ServiceError {
    message: String,
}
