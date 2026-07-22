use crate::{Error, Result};

const MAGIC: &[u8; 8] = b"GRMVEC01";
const HEADER_SIZE: usize = 24;

pub(crate) fn address(model: &str, source: &str) -> [u8; 32] {
    let mut hasher = blake3::Hasher::new();
    hasher.update(b"grimoire-vector-object-v1\0");
    hasher.update(model.as_bytes());
    hasher.update(&[0]);
    hasher.update(source.as_bytes());
    *hasher.finalize().as_bytes()
}

pub(crate) fn encode(model: &str, source: &str, vector: &[f32]) -> Result<Vec<u8>> {
    let model_len = u32::try_from(model.len())
        .map_err(|_| Error::InvalidInput("model identity is too long".into()))?;
    let source_len = u32::try_from(source.len())
        .map_err(|_| Error::InvalidInput("source identity is too long".into()))?;
    let dimensions = u32::try_from(vector.len())
        .map_err(|_| Error::InvalidInput("vector is too large".into()))?;
    let mut bytes = Vec::with_capacity(HEADER_SIZE + model.len() + source.len() + vector.len() * 4);
    bytes.extend_from_slice(MAGIC);
    bytes.extend_from_slice(&1_u16.to_le_bytes());
    bytes.extend_from_slice(&0_u16.to_le_bytes());
    bytes.extend_from_slice(&dimensions.to_le_bytes());
    bytes.extend_from_slice(&model_len.to_le_bytes());
    bytes.extend_from_slice(&source_len.to_le_bytes());
    bytes.extend_from_slice(model.as_bytes());
    bytes.extend_from_slice(source.as_bytes());
    for value in vector {
        bytes.extend_from_slice(&value.to_le_bytes());
    }
    Ok(bytes)
}

pub(crate) fn decode(bytes: &[u8], model: &str, source: &str) -> Result<Vec<f32>> {
    if bytes.len() < HEADER_SIZE || &bytes[..8] != MAGIC {
        return Err(Error::InvalidFormat("object header is missing".into()));
    }
    let dimensions = u32::from_le_bytes(bytes[12..16].try_into().unwrap()) as usize;
    let model_len = u32::from_le_bytes(bytes[16..20].try_into().unwrap()) as usize;
    let source_len = u32::from_le_bytes(bytes[20..24].try_into().unwrap()) as usize;
    let text_end = HEADER_SIZE
        .checked_add(model_len)
        .and_then(|value| value.checked_add(source_len))
        .ok_or_else(|| Error::InvalidFormat("object lengths overflow".into()))?;
    let vector_bytes = dimensions
        .checked_mul(4)
        .ok_or_else(|| Error::InvalidFormat("vector length overflow".into()))?;
    let expected = text_end
        .checked_add(vector_bytes)
        .ok_or_else(|| Error::InvalidFormat("object length overflow".into()))?;
    if bytes.len() != expected
        || &bytes[HEADER_SIZE..HEADER_SIZE + model_len] != model.as_bytes()
        || &bytes[HEADER_SIZE + model_len..text_end] != source.as_bytes()
    {
        return Err(Error::InvalidFormat(
            "object identity or length does not match".into(),
        ));
    }
    let mut vector = Vec::with_capacity(dimensions);
    for chunk in bytes[text_end..].chunks_exact(4) {
        let value = f32::from_le_bytes(chunk.try_into().unwrap());
        if !value.is_finite() {
            return Err(Error::InvalidFormat(
                "object contains non-finite vector value".into(),
            ));
        }
        vector.push(value);
    }
    validate_vector(&vector)?;
    Ok(vector)
}

pub(crate) fn validate_identity(model: &str, source: &str) -> Result<()> {
    if model.trim().is_empty() || source.trim().is_empty() {
        return Err(Error::InvalidInput(
            "model and source identities are required".into(),
        ));
    }
    Ok(())
}

pub(crate) fn validate_vector(vector: &[f32]) -> Result<()> {
    if vector.is_empty() || vector.iter().any(|value| !value.is_finite()) {
        return Err(Error::InvalidInput(
            "vector must contain finite dimensions".into(),
        ));
    }
    Ok(())
}

pub(crate) fn hex(bytes: &[u8]) -> String {
    const DIGITS: &[u8; 16] = b"0123456789abcdef";
    let mut output = String::with_capacity(bytes.len() * 2);
    for byte in bytes {
        output.push(DIGITS[(byte >> 4) as usize] as char);
        output.push(DIGITS[(byte & 0x0f) as usize] as char);
    }
    output
}
