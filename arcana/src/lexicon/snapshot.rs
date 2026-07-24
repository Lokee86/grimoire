use std::collections::BTreeMap;
use std::fmt::Write;
use std::fs;
use std::path::{Path, PathBuf};

use sha2::{Digest, Sha256};

use super::binary::{is_binary_object, parse_binary_object};
use super::format::{LanguageEntry, Manifest};
use super::object::{FactObject, FactRecord, parse_json_object};
use super::records::build_repository_facts;
use super::{
    FACT_SCHEMA_VERSION, LexiconSnapshot, LexiconSnapshotError, OBJECT_VERSION, SNAPSHOT_VERSION,
};
use crate::repository::normalize_repository_path;

const SNAPSHOT_DOMAIN: &str = "lexicon:snapshot:v1\0";
const OBJECT_DOMAIN: &str = "lexicon:fact-object:v1\0";

pub fn current(root: impl AsRef<Path>) -> Result<LexiconSnapshot, LexiconSnapshotError> {
    let storage = storage_root(root.as_ref());
    let current = fs::read(storage.join("CURRENT"))?;
    let text = std::str::from_utf8(&current).map_err(|_| LexiconSnapshotError::InvalidCurrent)?;
    let id = text
        .strip_suffix('\n')
        .filter(|value| !value.is_empty() && !value.chars().any(char::is_whitespace))
        .ok_or(LexiconSnapshotError::InvalidCurrent)?;
    load(root, id)
}

pub fn load(root: impl AsRef<Path>, id: &str) -> Result<LexiconSnapshot, LexiconSnapshotError> {
    validate_id(id)?;
    let storage = storage_root(root.as_ref());
    let manifest_bytes = read_verified_json(
        &storage
            .join("snapshots")
            .join(format!("{}.json", hex_id(id))),
        id,
        SNAPSHOT_DOMAIN,
        "snapshot manifest",
    )?;
    let manifest: Manifest = serde_json::from_slice(&manifest_bytes)?;
    if manifest.version != SNAPSHOT_VERSION {
        return Err(LexiconSnapshotError::UnsupportedSnapshotVersion(
            manifest.version,
        ));
    }
    if manifest.state_commit.is_empty() {
        return Err(LexiconSnapshotError::Malformed("manifest metadata"));
    }

    let mut files = BTreeMap::new();
    let mut shared_objects = BTreeMap::new();
    let mut all_records = Vec::<FactRecord>::new();
    let mut previous_language = None;
    for language in &manifest.languages {
        validate_language(language)?;
        if previous_language
            .as_deref()
            .is_some_and(|previous| previous >= language.language.as_str())
        {
            return Err(LexiconSnapshotError::Malformed("language ordering"));
        }
        previous_language = Some(language.language.clone());
        shared_objects.insert(language.language.clone(), language.shared_object_id.clone());
        if let Some(object_id) = &language.shared_object_id {
            let object = read_object(&storage, object_id)?;
            validate_object(&object, language, None, None)?;
            all_records.extend(object.records);
        }
        let mut previous_path = None;
        for file in &language.files {
            if previous_path
                .as_deref()
                .is_some_and(|previous| previous >= file.path.as_str())
            {
                return Err(LexiconSnapshotError::Malformed("file ordering"));
            }
            previous_path = Some(file.path.clone());
            let path = normalize_path("file", &file.path)?;
            if file.language != language.language
                || files
                    .insert(
                        (language.language.clone(), path.clone()),
                        file.object_id.clone(),
                    )
                    .is_some()
            {
                return Err(LexiconSnapshotError::MetadataMismatch("file entry"));
            }
            validate_id(&file.content_id)?;
            let object = read_object(&storage, &file.object_id)?;
            validate_object(&object, language, Some(&path), Some(&file.content_id))?;
            all_records.extend(object.records);
        }
    }

    let facts = build_repository_facts(all_records)?;
    Ok(LexiconSnapshot {
        id: id.to_owned(),
        facts,
        files,
        shared_objects,
    })
}

fn validate_language(language: &LanguageEntry) -> Result<(), LexiconSnapshotError> {
    if language.language.is_empty()
        || language.adapter_version.is_empty()
        || language.repository.is_empty()
        || language.analysis_config_id.is_empty()
    {
        return Err(LexiconSnapshotError::Malformed("language metadata"));
    }
    if language.schema_version != FACT_SCHEMA_VERSION {
        return Err(LexiconSnapshotError::UnsupportedSchemaVersion(
            language.schema_version,
        ));
    }
    if let Some(fingerprint) = &language.adapter_fingerprint {
        validate_id(fingerprint)?;
    }
    validate_id(&language.analysis_config_id)
}

fn read_object(storage: &Path, id: &str) -> Result<FactObject, LexiconSnapshotError> {
    validate_id(id)?;
    let path = storage
        .join("objects")
        .join(&hex_id(id)[..2])
        .join(&hex_id(id)[2..]);
    let bytes = fs::read(path)?;
    let canonical = if is_binary_object(&bytes) {
        bytes.as_slice()
    } else {
        bytes.trim_ascii()
    };
    verify_content(canonical, id, OBJECT_DOMAIN, "fact object")?;
    if is_binary_object(canonical) {
        parse_binary_object(canonical)
    } else {
        parse_json_object(canonical)
    }
}

fn validate_object(
    object: &FactObject,
    language: &LanguageEntry,
    owner: Option<&str>,
    content_id: Option<&str>,
) -> Result<(), LexiconSnapshotError> {
    if object.version != OBJECT_VERSION {
        return Err(LexiconSnapshotError::UnsupportedObjectVersion(
            object.version,
        ));
    }
    if object.language != language.language
        || object.adapter_version != language.adapter_version
        || object.schema_version != language.schema_version
        || object.analysis_config_id != language.analysis_config_id
    {
        return Err(LexiconSnapshotError::MetadataMismatch("fact object"));
    }
    match (owner, content_id) {
        (Some(owner), Some(content_id)) => {
            if object.owner.as_deref() != Some(owner)
                || object.source_content_id.as_deref() != Some(content_id)
            {
                return Err(LexiconSnapshotError::MetadataMismatch("file fact object"));
            }
        }
        (None, None) if object.owner.is_none() && object.source_content_id.is_none() => {}
        _ => {
            return Err(LexiconSnapshotError::MetadataMismatch("shared fact object"));
        }
    }
    Ok(())
}

fn read_verified_json(
    path: &Path,
    expected: &str,
    domain: &str,
    kind: &'static str,
) -> Result<Vec<u8>, LexiconSnapshotError> {
    let bytes = fs::read(path)?;
    let canonical = bytes.trim_ascii();
    verify_content(canonical, expected, domain, kind)?;
    Ok(canonical.to_vec())
}

fn verify_content(
    bytes: &[u8],
    expected: &str,
    domain: &str,
    kind: &'static str,
) -> Result<(), LexiconSnapshotError> {
    let actual = digest(domain, bytes);
    if actual != expected {
        return Err(LexiconSnapshotError::ContentHashMismatch {
            kind,
            expected: expected.to_owned(),
            actual,
        });
    }
    Ok(())
}

fn normalize_path(field: &'static str, path: &str) -> Result<String, LexiconSnapshotError> {
    normalize_repository_path(path).map_err(|_| LexiconSnapshotError::InvalidPath {
        field,
        path: path.to_owned(),
    })
}

fn storage_root(root: &Path) -> PathBuf {
    if root.file_name().is_some_and(|name| name == ".lexicon") {
        root.to_owned()
    } else {
        root.join(".lexicon")
    }
}

fn validate_id(id: &str) -> Result<(), LexiconSnapshotError> {
    let Some(hex) = id.strip_prefix("sha256:") else {
        return Err(LexiconSnapshotError::InvalidId(id.to_owned()));
    };
    if hex.len() != 64
        || !hex
            .bytes()
            .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
    {
        return Err(LexiconSnapshotError::InvalidId(id.to_owned()));
    }
    Ok(())
}

fn hex_id(id: &str) -> &str {
    id.strip_prefix("sha256:").expect("validated Lexicon ID")
}

fn digest(domain: &str, bytes: &[u8]) -> String {
    let mut hasher = Sha256::new();
    hasher.update(domain.as_bytes());
    hasher.update(bytes);
    let mut output = String::from("sha256:");
    for byte in hasher.finalize() {
        write!(output, "{byte:02x}").expect("writing to String cannot fail");
    }
    output
}
