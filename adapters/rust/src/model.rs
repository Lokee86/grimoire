use crate::contract::Facts;
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
    pub(crate) processed: HashSet<(PathBuf, String)>,
}
