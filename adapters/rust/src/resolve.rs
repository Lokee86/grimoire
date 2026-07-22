use crate::model::Context;
use quote::ToTokens;
use std::collections::BTreeMap;
use syn::UseTree;

pub(crate) fn resolve_type(
    context: &Context,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Option<String> {
    resolve_from_map(&context.types, path, module_qn, crate_qn)
}

pub(crate) fn resolve_trait(
    context: &Context,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Option<String> {
    resolve_from_map(&context.traits, path, module_qn, crate_qn)
}

pub(crate) fn resolve_symbol(
    context: &Context,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Option<String> {
    resolve_from_map(&context.symbols, path, module_qn, crate_qn)
}

fn resolve_from_map(
    map: &BTreeMap<String, String>,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Option<String> {
    let path = path.trim_start_matches("::");
    if path.is_empty() || path.contains('{') || path.contains('*') {
        return None;
    }
    if let Some(rest) = path.strip_prefix("crate::") {
        return map.get(&format!("{crate_qn}::{rest}")).cloned();
    }
    if let Some(rest) = path.strip_prefix("self::") {
        return map.get(&format!("{module_qn}::{rest}")).cloned();
    }
    if let Some(rest) = path.strip_prefix("super::") {
        let parent = parent_module(module_qn, crate_qn)?;
        return resolve_from_map(map, rest, &parent, crate_qn);
    }
    let mut base = module_qn.to_string();
    loop {
        if let Some(value) = map.get(&format!("{base}::{path}")) {
            return Some(value.clone());
        }
        if base == crate_qn {
            break;
        }
        base = parent_module(&base, crate_qn)?;
    }
    if !path.contains("::") {
        map.iter()
            .find(|(candidate, _)| candidate.ends_with(&format!("::{path}")))
            .map(|(_, value)| value.clone())
    } else {
        None
    }
}

fn parent_module(module_qn: &str, crate_qn: &str) -> Option<String> {
    if module_qn == crate_qn {
        return None;
    }
    let parent = module_qn.rsplit_once("::")?.0.to_string();
    if parent.len() < crate_qn.len() || !parent.starts_with(crate_qn) {
        None
    } else {
        Some(parent)
    }
}

pub(crate) fn use_paths(tree: &UseTree) -> Option<Vec<String>> {
    let mut paths = Vec::new();
    collect_use_paths(tree, &[], &mut paths)?;
    Some(paths)
}

fn collect_use_paths(tree: &UseTree, prefix: &[String], paths: &mut Vec<String>) -> Option<()> {
    match tree {
        UseTree::Path(path) => {
            let mut next = prefix.to_vec();
            next.push(path.ident.to_string());
            collect_use_paths(&path.tree, &next, paths)
        }
        UseTree::Name(name) => {
            let mut path = prefix.to_vec();
            path.push(name.ident.to_string());
            paths.push(path.join("::"));
            Some(())
        }
        UseTree::Group(group) => {
            for item in &group.items {
                collect_use_paths(item, prefix, paths)?;
            }
            Some(())
        }
        UseTree::Glob(_) | UseTree::Rename(_) => None,
    }
}

pub(crate) fn normalized_tokens<T: ToTokens>(value: &T) -> String {
    value
        .to_token_stream()
        .to_string()
        .split_whitespace()
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn expands_grouped_use_paths_and_rejects_unsupported_forms() {
        let item: syn::ItemUse =
            syn::parse_str("use crate::storage::{PackedGraph, write_packed};").unwrap();
        assert_eq!(
            use_paths(&item.tree),
            Some(vec![
                "crate::storage::PackedGraph".into(),
                "crate::storage::write_packed".into(),
            ])
        );

        let glob: syn::ItemUse = syn::parse_str("use super::*;").unwrap();
        assert_eq!(use_paths(&glob.tree), None);
    }
}
