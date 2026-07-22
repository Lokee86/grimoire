use std::{fs::OpenOptions, io::Write, path::Path};

use crate::{
    Error, ObjectStore, RecordRef, Result,
    format::{
        ENTRY_SIZE, HEADER_SIZE, Header, VECTOR_ALIGNMENT, align, encode_header, put_u32, put_u64,
    },
};

#[derive(Clone, Copy)]
pub(crate) struct Layout {
    header: Header,
    entries_offset: usize,
    ids_offset: usize,
    ids_len: usize,
    vectors_offset: usize,
}

pub(crate) fn layout(model: &str, records: &[RecordRef], dimensions: usize) -> Result<Layout> {
    let model_len = u32::try_from(model.len())
        .map_err(|_| Error::InvalidInput("model identity is too long".into()))?;
    let dimensions = u32::try_from(dimensions)
        .map_err(|_| Error::InvalidInput("vector has too many dimensions".into()))?;
    let entries_offset = align(
        HEADER_SIZE
            .checked_add(model.len())
            .ok_or_else(|| Error::InvalidInput("model layout overflow".into()))?,
        8,
    );
    let ids_offset = entries_offset
        .checked_add(
            records
                .len()
                .checked_mul(ENTRY_SIZE)
                .ok_or_else(|| Error::InvalidInput("entry table overflow".into()))?,
        )
        .ok_or_else(|| Error::InvalidInput("entry table overflow".into()))?;
    let ids_len = records
        .iter()
        .try_fold(0_usize, |total, record| total.checked_add(record.id.len()))
        .ok_or_else(|| Error::InvalidInput("id table overflow".into()))?;
    let vectors_offset = align(
        ids_offset
            .checked_add(ids_len)
            .ok_or_else(|| Error::InvalidInput("snapshot layout overflow".into()))?,
        VECTOR_ALIGNMENT,
    );
    Ok(Layout {
        header: Header {
            dimensions,
            count: records.len() as u64,
            model_len,
            entries_offset: entries_offset as u64,
            ids_offset: ids_offset as u64,
            vectors_offset: vectors_offset as u64,
        },
        entries_offset,
        ids_offset,
        ids_len,
        vectors_offset,
    })
}

pub(crate) fn write_snapshot(
    path: &Path,
    store: &ObjectStore,
    model: &str,
    records: &[RecordRef],
    dimensions: usize,
    layout: Layout,
) -> Result<()> {
    let mut file = OpenOptions::new().write(true).create_new(true).open(path)?;
    file.write_all(&encode_header(layout.header))?;
    file.write_all(model.as_bytes())?;
    write_padding(&mut file, layout.entries_offset - HEADER_SIZE - model.len())?;

    let mut id_offset = 0_usize;
    for record in records {
        let object = store.read(model, &record.source)?;
        if object.vector.len() != dimensions {
            return Err(Error::InvalidInput(
                "all vectors must have equal dimensions".into(),
            ));
        }
        let mut entry = [0_u8; ENTRY_SIZE];
        put_u64(&mut entry, 0, id_offset as u64);
        put_u32(&mut entry, 8, record.id.len() as u32);
        entry[16..48].copy_from_slice(&object.hash);
        file.write_all(&entry)?;
        id_offset += record.id.len();
    }
    for record in records {
        file.write_all(record.id.as_bytes())?;
    }
    write_padding(
        &mut file,
        layout.vectors_offset - layout.ids_offset - layout.ids_len,
    )?;
    for record in records {
        for value in store.read(model, &record.source)?.vector {
            file.write_all(&value.to_le_bytes())?;
        }
    }
    file.sync_all()?;
    Ok(())
}

fn write_padding(file: &mut impl Write, count: usize) -> Result<()> {
    if count > 0 {
        file.write_all(&vec![0_u8; count])?;
    }
    Ok(())
}
