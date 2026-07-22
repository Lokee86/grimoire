use std::collections::BTreeMap;
use std::fmt::Write;
use std::fs;
use std::path::{Path, PathBuf};

use serde_json::Value;
use sha2::{Digest, Sha256};

use super::format::{FactObject, LanguageEntry, Manifest};
use super::{
    FACT_SCHEMA_VERSION, LexiconSnapshot, LexiconSnapshotError, OBJECT_VERSION, SNAPSHOT_VERSION,
};
use crate::repository::{RepositoryFacts, normalize_repository_path};

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
    let manifest_bytes = read_verified(
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
    let mut all_records = Vec::new();
    let mut previous_language = None;
    let mut header = None;
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
        if header.is_none() {
            header = Some((
                language.adapter_version.clone(),
                language.language.clone(),
                language.repository.clone(),
            ));
        }
        if let Some(object_id) = &language.shared_object_id {
            let object = read_object(&storage, object_id)?;
            validate_object(&object, language, None, None)?;
            append_records(&mut all_records, object.records)?;
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
            append_records(&mut all_records, object.records)?;
        }
    }

    let facts = if all_records.is_empty() {
        RepositoryFacts::default()
    } else {
        let (adapter, language, repository) = header.expect("manifest has a language");
        let mut nodes = BTreeMap::<String, Value>::new();
        let mut edges = Vec::new();
        let mut unresolved = Vec::new();
        for record in all_records {
            match record.get("record").and_then(Value::as_str) {
                Some("node") => {
                    let id = record
                        .get("id")
                        .and_then(Value::as_str)
                        .ok_or(LexiconSnapshotError::Malformed("node identity"))?
                        .to_owned();
                    match nodes.get(&id) {
                        Some(existing) if existing != &record => {
                            return Err(LexiconSnapshotError::ConflictingNode(id));
                        }
                        Some(_) => {}
                        None => {
                            nodes.insert(id, record);
                        }
                    }
                }
                Some("edge") => edges.push(record),
                Some("unresolved") => unresolved.push(record),
                _ => return Err(LexiconSnapshotError::Malformed("fact object record")),
            }
        }
        let mut input = format!(
            "{{\"adapter_version\":{},\"language\":{},\"mode\":\"full\",\"record\":\"lexicon\",\"repository\":{},\"schema_version\":1}}\n",
            json_string(&adapter)?,
            json_string(&language)?,
            json_string(&repository)?,
        );
        for record in nodes.into_values().chain(edges).chain(unresolved) {
            input.push_str(&serde_json::to_string(&record)?);
            input.push('\n');
        }
        let mut facts = RepositoryFacts::parse(&input)?;
        facts.nodes.sort_unstable();
        facts.nodes.dedup();
        facts.edges.sort_unstable();
        facts.edges.dedup();
        facts.unresolved.sort_unstable();
        facts.unresolved.dedup();
        facts
    };

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
    validate_id(&language.analysis_config_id)
}

fn read_object(storage: &Path, id: &str) -> Result<FactObject, LexiconSnapshotError> {
    validate_id(id)?;
    let path = storage
        .join("objects")
        .join(&hex_id(id)[..2])
        .join(&hex_id(id)[2..]);
    let bytes = read_verified(&path, id, OBJECT_DOMAIN, "fact object")?;
    Ok(serde_json::from_slice(&bytes)?)
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
        _ => return Err(LexiconSnapshotError::MetadataMismatch("shared fact object")),
    }
    Ok(())
}

fn append_records(
    target: &mut Vec<Value>,
    records: Vec<Value>,
) -> Result<(), LexiconSnapshotError> {
    if records.iter().any(|record| !record.is_object()) {
        return Err(LexiconSnapshotError::Malformed("fact object records"));
    }
    target.extend(records);
    Ok(())
}

fn read_verified(
    path: &Path,
    expected: &str,
    domain: &str,
    kind: &'static str,
) -> Result<Vec<u8>, LexiconSnapshotError> {
    let bytes = fs::read(path)?;
    let canonical = bytes.trim_ascii();
    let actual = digest(domain, canonical);
    if actual != expected {
        return Err(LexiconSnapshotError::ContentHashMismatch {
            kind,
            expected: expected.to_owned(),
            actual,
        });
    }
    Ok(canonical.to_vec())
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

fn json_string(value: &str) -> Result<String, LexiconSnapshotError> {
    Ok(serde_json::to_string(value)?)
}
