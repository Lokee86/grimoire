use crate::extractor;
use crate::model::{Context, CrateContext, SourceFile};
use crate::paths::span_value;
use crate::relationships;
use proc_macro2::Span;
use quote::ToTokens;
use serde_json::Value;
use std::collections::BTreeMap;
use syn::spanned::Spanned;
use syn::{Item, TraitItem};

pub(crate) fn process_items(
    context: &mut Context,
    items: &[Item],
    owner_id: &str,
    module_qn: &str,
    source: &SourceFile,
    crate_context: &CrateContext,
) {
    for item in items {
        match item {
            Item::Mod(item_mod) => {
                let name = item_mod.ident.to_string();
                let child_qn = format!("{module_qn}::{name}");
                let module_id = add_decl_node(
                    context,
                    "module",
                    &child_qn,
                    &name,
                    source,
                    item_mod.span(),
                    attrs([("language_kind", "module")]),
                );
                context.modules.insert(child_qn.clone(), module_id.clone());
                relationships::define_and_contain(
                    context,
                    owner_id,
                    &module_id,
                    item_mod.span(),
                    &source.relative,
                );
                if let Some((_, nested_items)) = &item_mod.content {
                    process_items(
                        context,
                        nested_items,
                        &module_id,
                        &child_qn,
                        source,
                        crate_context,
                    );
                } else if let Some(child_path) =
                    crate::paths::resolve_module_file(&context.sources, &source.path, &name)
                {
                    extractor::process_file(
                        context,
                        &child_path,
                        &module_id,
                        &child_qn,
                        crate_context,
                    );
                } else {
                    context.facts.add_unresolved(
                        owner_id,
                        "contains",
                        &format!("mod {name}"),
                        "missing-target",
                        span_value(item_mod.span(), &source.relative),
                    );
                }
            }
            Item::Struct(item_struct) => {
                let name = item_struct.ident.to_string();
                let qn = format!("{module_qn}::{name}");
                let id = add_decl_node(
                    context,
                    "type",
                    &qn,
                    &name,
                    source,
                    item_struct.span(),
                    attrs([("language_kind", "struct")]),
                );
                context.symbols.insert(qn.clone(), id.clone());
                context.types.insert(qn, id.clone());
                relationships::define_and_contain(
                    context,
                    owner_id,
                    &id,
                    item_struct.span(),
                    &source.relative,
                );
            }
            Item::Enum(item_enum) => {
                let name = item_enum.ident.to_string();
                let qn = format!("{module_qn}::{name}");
                let id = add_decl_node(
                    context,
                    "type",
                    &qn,
                    &name,
                    source,
                    item_enum.span(),
                    attrs([("language_kind", "enum")]),
                );
                context.symbols.insert(qn.clone(), id.clone());
                context.types.insert(qn, id.clone());
                relationships::define_and_contain(
                    context,
                    owner_id,
                    &id,
                    item_enum.span(),
                    &source.relative,
                );
            }
            Item::Trait(item_trait) => {
                let name = item_trait.ident.to_string();
                let qn = format!("{module_qn}::{name}");
                let id = add_decl_node(
                    context,
                    "trait",
                    &qn,
                    &name,
                    source,
                    item_trait.span(),
                    attrs([("language_kind", "trait")]),
                );
                context.symbols.insert(qn.clone(), id.clone());
                context.traits.insert(qn.clone(), id.clone());
                relationships::define_and_contain(
                    context,
                    owner_id,
                    &id,
                    item_trait.span(),
                    &source.relative,
                );
                for trait_item in &item_trait.items {
                    if let TraitItem::Fn(function) = trait_item {
                        let method_name = function.sig.ident.to_string();
                        let method_qn = format!("{qn}::{method_name}");
                        let method_id = add_decl_node(
                            context,
                            "method",
                            &method_qn,
                            &method_name,
                            source,
                            function.span(),
                            attrs([("language_kind", "trait-method")]),
                        );
                        context.symbols.insert(method_qn, method_id.clone());
                        relationships::define_and_contain(
                            context,
                            &id,
                            &method_id,
                            function.span(),
                            &source.relative,
                        );
                    }
                }
            }
            Item::Fn(function) => {
                let name = function.sig.ident.to_string();
                let qn = format!("{module_qn}::{name}");
                let id = add_decl_node(
                    context,
                    "function",
                    &qn,
                    &name,
                    source,
                    function.span(),
                    attrs([("language_kind", "function")]),
                );
                context.symbols.insert(qn, id.clone());
                relationships::define_and_contain(
                    context,
                    owner_id,
                    &id,
                    function.span(),
                    &source.relative,
                );
                relationships::collect_calls(
                    context,
                    &function.block,
                    &id,
                    module_qn,
                    &crate_context.qn,
                    &source.relative,
                );
            }
            Item::Impl(item_impl) => relationships::process_impl(
                context,
                item_impl,
                owner_id,
                module_qn,
                source,
                crate_context,
            ),
            Item::Use(item_use) => {
                relationships::process_use(context, item_use, owner_id, module_qn, source)
            }
            Item::Macro(item_macro) => context.facts.add_unresolved(
                owner_id,
                "defines",
                &item_macro.to_token_stream().to_string(),
                "generated-target",
                span_value(item_macro.span(), &source.relative),
            ),
            _ => {}
        }
    }
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
