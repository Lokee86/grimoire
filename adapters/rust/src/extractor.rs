use crate::contract::stable_id;
use crate::items;
use crate::model::Context;
use crate::relationships;
use std::path::{Path, PathBuf};

pub(crate) fn extract(context: &mut Context) {
    let crates = context.crates.clone();
    for crate_context in &crates {
        if context.sources.contains_key(&crate_context.root) {
            process_file(
                context,
                &crate_context.root,
                &crate_context.node_id,
                &crate_context.qn,
                crate_context,
            );
        }
    }
    let remaining: Vec<PathBuf> = context
        .sources
        .keys()
        .filter(|path| !context.processed.iter().any(|(seen, _)| seen == *path))
        .cloned()
        .collect();
    for path in remaining {
        if let Some(crate_context) = crates
            .iter()
            .filter(|candidate| path.starts_with(&candidate.package_root))
            .max_by_key(|candidate| candidate.package_root.as_os_str().len())
        {
            process_file(
                context,
                &path,
                &crate_context.node_id,
                &crate_context.qn,
                crate_context,
            );
        } else {
            context.facts.add_unresolved(
                &stable_id("rust", "repository", &context.repository),
                "contains",
                &path.display().to_string(),
                "unsupported-form",
                None,
            );
        }
    }
    relationships::resolve_calls(context);
}

pub(crate) fn process_file(
    context: &mut Context,
    path: &Path,
    owner_id: &str,
    module_qn: &str,
    crate_context: &crate::model::CrateContext,
) {
    let key = (path.to_path_buf(), module_qn.to_string());
    if !context.processed.insert(key) {
        return;
    }
    let Some(source) = context.sources.get(path).cloned() else {
        context.facts.add_unresolved(
            owner_id,
            "contains",
            &path.display().to_string(),
            "missing-target",
            None,
        );
        return;
    };
    let file_id = stable_id("rust", "file", &source.relative);
    context.facts.add_edge(owner_id, &file_id, "contains", None);
    let items = source.syntax.items.clone();
    items::process_items(context, &items, &file_id, module_qn, &source, crate_context);
}
