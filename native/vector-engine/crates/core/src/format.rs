use crate::{Error, Result};

pub const SNAPSHOT_MAGIC: &[u8; 8] = b"GRMSNP01";
pub const SNAPSHOT_VERSION: u16 = 1;
pub const HEADER_SIZE: usize = 64;
pub const ENTRY_SIZE: usize = 48;
pub const VECTOR_ALIGNMENT: usize = 64;

#[derive(Clone, Copy, Debug)]
pub struct Header {
    pub dimensions: u32,
    pub count: u64,
    pub model_len: u32,
    pub entries_offset: u64,
    pub ids_offset: u64,
    pub vectors_offset: u64,
}

pub fn encode_header(header: Header) -> [u8; HEADER_SIZE] {
    let mut bytes = [0_u8; HEADER_SIZE];
    bytes[..8].copy_from_slice(SNAPSHOT_MAGIC);
    put_u16(&mut bytes, 8, SNAPSHOT_VERSION);
    put_u16(&mut bytes, 10, HEADER_SIZE as u16);
    put_u32(&mut bytes, 12, header.dimensions);
    put_u64(&mut bytes, 16, header.count);
    put_u32(&mut bytes, 24, header.model_len);
    put_u32(&mut bytes, 28, ENTRY_SIZE as u32);
    put_u64(&mut bytes, 32, HEADER_SIZE as u64);
    put_u64(&mut bytes, 40, header.entries_offset);
    put_u64(&mut bytes, 48, header.ids_offset);
    put_u64(&mut bytes, 56, header.vectors_offset);
    bytes
}

pub fn decode_header(bytes: &[u8]) -> Result<Header> {
    if bytes.len() < HEADER_SIZE || &bytes[..8] != SNAPSHOT_MAGIC {
        return Err(Error::InvalidFormat("snapshot header is missing".into()));
    }
    if get_u16(bytes, 8)? != SNAPSHOT_VERSION || get_u16(bytes, 10)? as usize != HEADER_SIZE {
        return Err(Error::InvalidFormat("unsupported snapshot version".into()));
    }
    if get_u32(bytes, 28)? as usize != ENTRY_SIZE || get_u64(bytes, 32)? as usize != HEADER_SIZE {
        return Err(Error::InvalidFormat("unsupported snapshot layout".into()));
    }
    Ok(Header {
        dimensions: get_u32(bytes, 12)?,
        count: get_u64(bytes, 16)?,
        model_len: get_u32(bytes, 24)?,
        entries_offset: get_u64(bytes, 40)?,
        ids_offset: get_u64(bytes, 48)?,
        vectors_offset: get_u64(bytes, 56)?,
    })
}

pub fn align(value: usize, alignment: usize) -> usize {
    (value + alignment - 1) & !(alignment - 1)
}

pub fn put_u16(bytes: &mut [u8], offset: usize, value: u16) {
    bytes[offset..offset + 2].copy_from_slice(&value.to_le_bytes());
}

pub fn put_u32(bytes: &mut [u8], offset: usize, value: u32) {
    bytes[offset..offset + 4].copy_from_slice(&value.to_le_bytes());
}

pub fn put_u64(bytes: &mut [u8], offset: usize, value: u64) {
    bytes[offset..offset + 8].copy_from_slice(&value.to_le_bytes());
}

pub fn get_u16(bytes: &[u8], offset: usize) -> Result<u16> {
    Ok(u16::from_le_bytes(read(bytes, offset)?))
}

pub fn get_u32(bytes: &[u8], offset: usize) -> Result<u32> {
    Ok(u32::from_le_bytes(read(bytes, offset)?))
}

pub fn get_u64(bytes: &[u8], offset: usize) -> Result<u64> {
    Ok(u64::from_le_bytes(read(bytes, offset)?))
}

fn read<const N: usize>(bytes: &[u8], offset: usize) -> Result<[u8; N]> {
    bytes
        .get(offset..offset + N)
        .ok_or_else(|| Error::InvalidFormat("integer exceeds file bounds".into()))?
        .try_into()
        .map_err(|_| Error::InvalidFormat("invalid integer width".into()))
}
