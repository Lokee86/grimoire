use crate::model::{Context, CrateContext, SourceFile};
use crate::paths::{span_start, span_value};
use crate::resolve;
use proc_macro2::Span;
use quote::ToTokens;
use serde_json::Value;
use std::collections::BTreeMap;
use syn::spanned::Spanned;
use syn::ImplItem;

pub(crate) fn process_impl(
    context: &mut Context,
    item_impl: &syn::ItemImpl,
    owner_id: &str,
    module_qn: &str,
    source: &SourceFile,
    crate_context: &CrateContext,
) {
    let self_text = resolve::normalized_tokens(&item_impl.self_ty);
    let self_id = resolve::resolve_type(context, &self_text, module_qn, &crate_context.qn);
    let trait_text = item_impl
        .trait_
        .as_ref()
        .map(|(_, path, _)| resolve::normalized_tokens(path));
    if let Some(trait_text) = &trait_text {
        let trait_id = resolve::resolve_trait(context, trait_text, module_qn, &crate_context.qn);
        match (self_id.clone(), trait_id) {
            (Some(self_id), Some(trait_id)) => context.facts.add_edge(
                &self_id,
                &trait_id,
                "implements",
                span_value(item_impl.span(), &source.relative),
            ),
            _ => context.facts.add_unresolved(
                owner_id,
                "implements",
                &format!("impl {trait_text} for {self_text}"),
                if trait_text.starts_with("std::") || trait_text.starts_with("core::") {
                    "external-target"
                } else {
                    "missing-target"
                },
                span_value(item_impl.span(), &source.relative),
            ),
        }
    }
    let method_owner = self_id.as_deref().unwrap_or(owner_id);
    let type_name = self_text.split("::").last().unwrap_or(self_text.as_str());
    let impl_suffix = trait_text
        .as_deref()
        .map(|name| format!("::{name}"))
        .unwrap_or_default();
    for impl_item in &item_impl.items {
        if let ImplItem::Fn(function) = impl_item {
            let name = function.sig.ident.to_string();
            let qn = format!("{module_qn}::{type_name}{impl_suffix}::{name}");
            let id = add_decl_node(
                context,
                "method",
                &qn,
                &name,
                source,
                function.span(),
                attrs([("language_kind", "impl-method")]),
            );
            context.symbols.insert(qn, id.clone());
            define_and_contain(
                context,
                method_owner,
                &id,
                function.span(),
                &source.relative,
            );
        }
    }
}

pub(crate) fn process_use(
    context: &mut Context,
    item_use: &syn::ItemUse,
    owner_id: &str,
    module_qn: &str,
    source: &SourceFile,
) {
    let expression = item_use.to_token_stream().to_string();
    let name = expression
        .strip_prefix("use ")
        .unwrap_or(&expression)
        .trim_end_matches(';')
        .trim()
        .to_string();
    let start = span_start(item_use.span());
    let qn = format!("{module_qn}::use:{name}@{}:{}", start.0, start.1);
    let import_id = add_decl_node(
        context,
        "import",
        &qn,
        &name,
        source,
        item_use.span(),
        attrs([("language_kind", "use")]),
    );
    define_and_contain(
        context,
        owner_id,
        &import_id,
        item_use.span(),
        &source.relative,
    );
    if let Some(path) = resolve::simple_use_path(&item_use.tree) {
        let crate_qn = module_qn.split("::").take(2).collect::<Vec<_>>().join("::");
        if let Some(target) = resolve::resolve_symbol(context, &path, module_qn, &crate_qn) {
            context.facts.add_edge(
                owner_id,
                &target,
                "imports",
                span_value(item_use.span(), &source.relative),
            );
        } else {
            context.facts.add_unresolved(
                owner_id,
                "imports",
                &path,
                if path.starts_with("std::") || path.starts_with("core::") {
                    "external-target"
                } else {
                    "missing-target"
                },
                span_value(item_use.span(), &source.relative),
            );
        }
    } else {
        context.facts.add_unresolved(
            owner_id,
            "imports",
            &name,
            "unsupported-form",
            span_value(item_use.span(), &source.relative),
        );
    }
}

pub(crate) fn define_and_contain(
    context: &mut Context,
    owner: &str,
    target: &str,
    span: Span,
    path: &str,
) {
    let span = span_value(span, path);
    context
        .facts
        .add_edge(owner, target, "contains", span.clone());
    context.facts.add_edge(owner, target, "defines", span);
}

fn add_decl_node(
    context: &mut Context,
    kind: &str,
    qn: &str,
    name: &str,
    source: &SourceFile,
    span: Span,
    attributes: BTreeMap<String, Value>,
) -> String {
    context.facts.add_node(
        "rust",
        kind,
        qn,
        name,
        &source.relative,
        qn,
        None,
        span_value(span, &source.relative),
        attributes,
    )
}

fn attrs<const N: usize>(values: [(&str, &str); N]) -> BTreeMap<String, Value> {
    values
        .into_iter()
        .map(|(key, value)| (key.into(), Value::String(value.into())))
        .collect()
}
