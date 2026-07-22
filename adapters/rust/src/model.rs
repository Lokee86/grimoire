use crate::contract::Facts;
use serde_json::Value;
use std::collections::{BTreeMap, HashSet};
use std::path::PathBuf;

#[derive(Clone)]
pub(crate) struct SourceFile {
    pub(crate) path: PathBuf,
    pub(crate) relative: String,
    pub(crate) content: Vec<u8>,
    pub(crate) syntax: syn::File,
}

#[derive(Clone)]
pub(crate) struct CrateContext {
    pub(crate) qn: String,
    pub(crate) node_id: String,
    pub(crate) root: PathBuf,
    pub(crate) package_root: PathBuf,
}

#[derive(Clone, Copy)]
pub(crate) enum CallForm {
    Path,
    Associated,
    Method,
    Macro,
    Unsupported,
}

pub(crate) struct PendingCall {
    pub(crate) owner_id: String,
    pub(crate) module_qn: String,
    pub(crate) crate_qn: String,
    pub(crate) form: CallForm,
    pub(crate) path: Option<String>,
    pub(crate) expression: String,
    pub(crate) span: Option<Value>,
}

pub(crate) struct Context {
    pub(crate) repo: PathBuf,
    pub(crate) repository: String,
    pub(crate) sources: BTreeMap<PathBuf, SourceFile>,
    pub(crate) facts: Facts,
    pub(crate) crates: Vec<CrateContext>,
    pub(crate) modules: BTreeMap<String, String>,
    pub(crate) symbols: BTreeMap<String, String>,
    pub(crate) types: BTreeMap<String, String>,
    pub(crate) traits: BTreeMap<String, String>,
    pub(crate) inherent_methods: BTreeMap<String, Vec<String>>,
    pub(crate) processed: HashSet<(PathBuf, String)>,
    pub(crate) pending_calls: Vec<PendingCall>,
}
