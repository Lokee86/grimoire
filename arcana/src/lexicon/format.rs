use serde::{Deserialize, Deserializer};
use serde_json::Value;

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub(super) struct Manifest {
    pub(super) version: u64,
    pub(super) state_commit: String,
    pub(super) languages: Vec<LanguageEntry>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub(super) struct LanguageEntry {
    pub(super) language: String,
    pub(super) adapter_version: String,
    #[serde(default)]
    pub(super) adapter_fingerprint: Option<String>,
    pub(super) schema_version: u64,
    pub(super) repository: String,
    pub(super) analysis_config_id: String,
    #[serde(default)]
    pub(super) shared_object_id: Option<String>,
    #[serde(default, deserialize_with = "null_default")]
    pub(super) files: Vec<FileEntry>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub(super) struct FileEntry {
    pub(super) path: String,
    pub(super) language: String,
    pub(super) content_id: String,
    pub(super) object_id: String,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub(super) struct JsonFactObject {
    pub(super) version: u64,
    pub(super) language: String,
    #[serde(default)]
    pub(super) owner: Option<String>,
    #[serde(default)]
    pub(super) source_content_id: Option<String>,
    pub(super) adapter_version: String,
    pub(super) schema_version: u64,
    pub(super) analysis_config_id: String,
    pub(super) records: Vec<Value>,
}

fn null_default<'de, D, T>(deserializer: D) -> Result<T, D::Error>
where
    D: Deserializer<'de>,
    T: Deserialize<'de> + Default,
{
    Ok(Option::<T>::deserialize(deserializer)?.unwrap_or_default())
}
