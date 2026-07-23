use crate::contract::Facts;
use serde_json::Value;
use std::collections::{BTreeMap, BTreeSet, HashSet};
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
    pub(crate) external_crates: BTreeSet<String>,
}

#[derive(Clone, Default, PartialEq, Eq)]
pub(crate) struct ValueSet {
    pub(crate) types: BTreeSet<String>,
    pub(crate) contained_types: BTreeSet<String>,
    pub(crate) traits: BTreeSet<String>,
    pub(crate) callables: BTreeSet<String>,
    pub(crate) tuple_elements: Vec<ValueSet>,
    pub(crate) contained_values: Vec<ValueSet>,
    pub(crate) external: bool,
    pub(crate) unknown: bool,
    pub(crate) dynamic_callable: bool,
}

impl ValueSet {
    pub(crate) fn merge(&mut self, other: &Self) -> bool {
        let before = self.clone();
        self.types.extend(other.types.iter().cloned());
        self.contained_types
            .extend(other.contained_types.iter().cloned());
        self.traits.extend(other.traits.iter().cloned());
        self.callables.extend(other.callables.iter().cloned());
        merge_value_lists(&mut self.tuple_elements, &other.tuple_elements);
        merge_value_lists(&mut self.contained_values, &other.contained_values);
        self.external |= other.external;
        self.unknown |= other.unknown;
        self.dynamic_callable |= other.dynamic_callable;
        *self != before
    }

    pub(crate) fn callable(id: String, dynamic: bool) -> Self {
        Self {
            callables: [id].into_iter().collect(),
            dynamic_callable: dynamic,
            ..Self::default()
        }
    }
}

fn merge_value_lists(target: &mut Vec<ValueSet>, source: &[ValueSet]) {
    if target.len() < source.len() {
        target.resize_with(source.len(), ValueSet::default);
    }
    for (index, value) in source.iter().enumerate() {
        target[index].merge(value);
    }
}

#[derive(Clone)]
pub(crate) enum FunctionBody {
    Block(syn::Block),
    Expr(syn::Expr),
}

#[derive(Clone)]
pub(crate) struct ParameterInfo {
    pub(crate) name: String,
    pub(crate) type_text: Option<String>,
    pub(crate) callable_bound: bool,
}

#[derive(Clone)]
pub(crate) struct FunctionInfo {
    pub(crate) id: String,
    pub(crate) qn: String,
    pub(crate) module_qn: String,
    pub(crate) crate_qn: String,
    pub(crate) source_path: String,
    pub(crate) body: FunctionBody,
    pub(crate) parameters: Vec<ParameterInfo>,
    pub(crate) return_type: Option<String>,
    pub(crate) self_type: Option<String>,
    pub(crate) trait_path: Option<String>,
    pub(crate) generic_bounds: BTreeMap<String, Vec<String>>,
}

#[derive(Clone)]
pub(crate) struct MethodInfo {
    pub(crate) id: String,
    pub(crate) self_type: String,
    pub(crate) trait_path: Option<String>,
    pub(crate) name: String,
    pub(crate) module_qn: String,
    pub(crate) crate_qn: String,
}

#[derive(Clone)]
pub(crate) struct PendingImpl {
    pub(crate) owner_id: String,
    pub(crate) self_type: String,
    pub(crate) trait_path: String,
    pub(crate) module_qn: String,
    pub(crate) crate_qn: String,
    pub(crate) expression: String,
    pub(crate) span: Option<Value>,
}

#[derive(Clone)]
pub(crate) struct PendingImport {
    pub(crate) owner_id: String,
    pub(crate) module_qn: String,
    pub(crate) crate_qn: String,
    pub(crate) item: syn::ItemUse,
    pub(crate) expression: String,
    pub(crate) span: Option<Value>,
}

#[derive(Clone, Default)]
pub(crate) struct ImportScope {
    pub(crate) bindings: BTreeMap<String, Vec<String>>,
    pub(crate) glob_modules: BTreeSet<String>,
    pub(crate) external_aliases: BTreeSet<String>,
}

#[derive(Clone)]
pub(crate) struct FieldInfo {
    pub(crate) type_text: String,
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
    pub(crate) macros: BTreeMap<String, String>,
    pub(crate) constructors: BTreeMap<String, String>,
    pub(crate) constructor_types: BTreeMap<String, String>,
    pub(crate) type_aliases: BTreeMap<String, String>,
    pub(crate) type_qn_by_id: BTreeMap<String, String>,
    pub(crate) trait_qn_by_id: BTreeMap<String, String>,
    pub(crate) function_qn_by_id: BTreeMap<String, String>,
    pub(crate) functions: BTreeMap<String, FunctionInfo>,
    pub(crate) methods: Vec<MethodInfo>,
    pub(crate) method_index: BTreeMap<(String, String), Vec<String>>,
    pub(crate) trait_method_index: BTreeMap<(String, String), Vec<String>>,
    pub(crate) type_traits: BTreeMap<String, BTreeSet<String>>,
    pub(crate) fields: BTreeMap<(String, String), FieldInfo>,
    pub(crate) pending_impls: Vec<PendingImpl>,
    pub(crate) pending_imports: Vec<PendingImport>,
    pub(crate) imports: BTreeMap<String, ImportScope>,
    pub(crate) closure_ids: BTreeMap<(String, usize, usize), String>,
    pub(crate) propagated_parameters: BTreeMap<(String, usize), ValueSet>,
    pub(crate) return_values: BTreeMap<String, ValueSet>,
    pub(crate) processed: HashSet<(PathBuf, String)>,
}
