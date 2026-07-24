use std::fmt::Write as FmtWrite;
use std::path::{Component, Path, PathBuf};

use super::RepositorySnapshotError;

pub const REPOSITORY_MANIFEST_VERSION: u64 = 1;

const FIELDS: [&str; 19] = [
    "version",
    "snapshot_id",
    "created_unix_seconds",
    "repository_id",
    "adapter_name",
    "adapter_version",
    "fact_schema_version",
    "node_count",
    "edge_count",
    "unresolved_count",
    "graph_snapshot_id",
    "graph_manifest_checksum",
    "catalogue_checksum",
    "unresolved_checksum",
    "facts_checksum",
    "graph_manifest_file",
    "catalogue_file",
    "unresolved_file",
    "facts_file",
];

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct RepositorySnapshotManifest {
    pub snapshot_id: u64,
    pub created_unix_seconds: u64,
    pub repository_id: u64,
    pub adapter_name: String,
    pub adapter_version: String,
    pub fact_schema_version: u64,
    pub node_count: u32,
    pub edge_count: u64,
    pub unresolved_count: u64,
    pub graph_snapshot_id: u64,
    pub graph_manifest_checksum: u64,
    pub catalogue_checksum: u64,
    pub unresolved_checksum: u64,
    pub facts_checksum: u64,
    pub graph_manifest_file: PathBuf,
    pub catalogue_file: PathBuf,
    pub unresolved_file: PathBuf,
    pub facts_file: PathBuf,
}

impl RepositorySnapshotManifest {
    pub fn encode(&self) -> Result<String, RepositorySnapshotError> {
        validate_text("adapter_name", &self.adapter_name)?;
        validate_text("adapter_version", &self.adapter_version)?;
        let paths = [
            ("graph_manifest_file", &self.graph_manifest_file),
            ("catalogue_file", &self.catalogue_file),
            ("unresolved_file", &self.unresolved_file),
            ("facts_file", &self.facts_file),
        ];
        for (field, path) in paths {
            validate_path(field, path)?;
        }
        let mut output = String::new();
        macro_rules! field {
            ($name:literal, $value:expr) => {
                writeln!(output, concat!($name, "={}"), $value)
                    .expect("writing to String cannot fail")
            };
        }
        field!("version", REPOSITORY_MANIFEST_VERSION);
        field!("snapshot_id", format_args!("{:016x}", self.snapshot_id));
        field!("created_unix_seconds", self.created_unix_seconds);
        field!("repository_id", format_args!("{:016x}", self.repository_id));
        field!("adapter_name", self.adapter_name);
        field!("adapter_version", self.adapter_version);
        field!("fact_schema_version", self.fact_schema_version);
        field!("node_count", self.node_count);
        field!("edge_count", self.edge_count);
        field!("unresolved_count", self.unresolved_count);
        field!(
            "graph_snapshot_id",
            format_args!("{:016x}", self.graph_snapshot_id)
        );
        field!(
            "graph_manifest_checksum",
            format_args!("{:016x}", self.graph_manifest_checksum)
        );
        field!(
            "catalogue_checksum",
            format_args!("{:016x}", self.catalogue_checksum)
        );
        field!(
            "unresolved_checksum",
            format_args!("{:016x}", self.unresolved_checksum)
        );
        field!(
            "facts_checksum",
            format_args!("{:016x}", self.facts_checksum)
        );
        field!("graph_manifest_file", self.graph_manifest_file.display());
        field!("catalogue_file", self.catalogue_file.display());
        field!("unresolved_file", self.unresolved_file.display());
        field!("facts_file", self.facts_file.display());
        Ok(output)
    }

    pub fn decode(text: &str) -> Result<Self, RepositorySnapshotError> {
        if !text.ends_with('\n') {
            return Err(RepositorySnapshotError::MalformedManifest(
                "missing final newline",
            ));
        }
        let lines = text[..text.len() - 1].split('\n').collect::<Vec<_>>();
        if lines.len() != FIELDS.len() {
            return Err(RepositorySnapshotError::MalformedManifest(
                "wrong field count",
            ));
        }
        let mut values = Vec::with_capacity(FIELDS.len());
        for (index, line) in lines.iter().enumerate() {
            let (field, value) =
                line.split_once('=')
                    .ok_or(RepositorySnapshotError::MalformedManifest(
                        "malformed field",
                    ))?;
            if field != FIELDS[index] {
                return Err(RepositorySnapshotError::MalformedManifest(
                    "field order mismatch",
                ));
            }
            values.push(value);
        }
        let version = decimal(values[0])?;
        if version != REPOSITORY_MANIFEST_VERSION {
            return Err(RepositorySnapshotError::UnsupportedManifestVersion(version));
        }
        let manifest = Self {
            snapshot_id: hex(values[1])?,
            created_unix_seconds: decimal(values[2])?,
            repository_id: hex(values[3])?,
            adapter_name: values[4].to_owned(),
            adapter_version: values[5].to_owned(),
            fact_schema_version: decimal(values[6])?,
            node_count: u32::try_from(decimal(values[7])?)
                .map_err(|_| RepositorySnapshotError::MalformedManifest("node count overflow"))?,
            edge_count: decimal(values[8])?,
            unresolved_count: decimal(values[9])?,
            graph_snapshot_id: hex(values[10])?,
            graph_manifest_checksum: hex(values[11])?,
            catalogue_checksum: hex(values[12])?,
            unresolved_checksum: hex(values[13])?,
            facts_checksum: hex(values[14])?,
            graph_manifest_file: PathBuf::from(values[15]),
            catalogue_file: PathBuf::from(values[16]),
            unresolved_file: PathBuf::from(values[17]),
            facts_file: PathBuf::from(values[18]),
        };
        validate_text("adapter_name", &manifest.adapter_name)?;
        validate_text("adapter_version", &manifest.adapter_version)?;
        for (field, path) in [
            ("graph_manifest_file", &manifest.graph_manifest_file),
            ("catalogue_file", &manifest.catalogue_file),
            ("unresolved_file", &manifest.unresolved_file),
            ("facts_file", &manifest.facts_file),
        ] {
            validate_path(field, path)?;
        }
        Ok(manifest)
    }
}

fn decimal(value: &str) -> Result<u64, RepositorySnapshotError> {
    if value.is_empty() || (value.len() > 1 && value.starts_with('0')) {
        return Err(RepositorySnapshotError::MalformedManifest(
            "invalid decimal",
        ));
    }
    value
        .parse()
        .map_err(|_| RepositorySnapshotError::MalformedManifest("invalid decimal"))
}

fn hex(value: &str) -> Result<u64, RepositorySnapshotError> {
    if value.len() != 16
        || !value
            .bytes()
            .all(|byte| byte.is_ascii_hexdigit() && !byte.is_ascii_uppercase())
    {
        return Err(RepositorySnapshotError::MalformedManifest(
            "invalid checksum",
        ));
    }
    u64::from_str_radix(value, 16)
        .map_err(|_| RepositorySnapshotError::MalformedManifest("invalid checksum"))
}

fn validate_text(field: &'static str, value: &str) -> Result<(), RepositorySnapshotError> {
    if value.is_empty() || value.chars().any(char::is_control) || value.contains('=') {
        return Err(RepositorySnapshotError::InvalidTextField(field));
    }
    Ok(())
}

fn validate_path(field: &'static str, path: &Path) -> Result<(), RepositorySnapshotError> {
    if path.as_os_str().is_empty()
        || path.is_absolute()
        || path
            .components()
            .any(|component| matches!(component, Component::ParentDir))
        || path.to_str().is_none()
    {
        return Err(RepositorySnapshotError::InvalidComponentPath(field));
    }
    Ok(())
}
