use crate::function_index::{self, Registration};
use crate::model::{Context, FieldInfo, FunctionBody, MethodInfo, SourceFile};
use crate::paths::span_value;
use crate::syntax::{normalized_tokens, type_tokens};
use proc_macro2::Span;
use quote::ToTokens;
use serde_json::Value;
use std::collections::BTreeMap;
use syn::spanned::Spanned;

pub(crate) fn structure(
    context: &mut Context,
    item: &syn::ItemStruct,
    owner: &str,
    module: &str,
    source: &SourceFile,
) {
    let name = item.ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(context, "type", &qn, &name, source, item.span(), "struct");
    register_type(context, &qn, &id);
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
    for (index, field) in item.fields.iter().enumerate() {
        let field_name = field
            .ident
            .as_ref()
            .map(ToString::to_string)
            .unwrap_or_else(|| index.to_string());
        context.fields.insert(
            (qn.clone(), field_name.clone()),
            FieldInfo {
                type_text: type_tokens(&field.ty),
            },
        );
        let field_qn = format!("{qn}::{field_name}");
        let field_id = add_node(
            context,
            "field",
            &field_qn,
            &field_name,
            source,
            field.span(),
            "field",
        );
        context
            .field_ids
            .insert((qn.clone(), field_name), field_id.clone());
        crate::relationships::define_and_contain(
            context,
            &id,
            &field_id,
            field.span(),
            &source.relative,
        );
    }
    if matches!(item.fields, syn::Fields::Unnamed(_)) {
        let constructor_qn = format!("{qn}::constructor");
        let constructor = add_node(
            context,
            "function",
            &constructor_qn,
            &name,
            source,
            item.span(),
            "tuple-struct-constructor",
        );
        context.constructors.insert(qn.clone(), constructor.clone());
        context.constructor_types.insert(constructor, qn);
    }
}

pub(crate) fn enumeration(
    context: &mut Context,
    item: &syn::ItemEnum,
    owner: &str,
    module: &str,
    source: &SourceFile,
) {
    let name = item.ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(context, "type", &qn, &name, source, item.span(), "enum");
    register_type(context, &qn, &id);
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
    for variant in &item.variants {
        let variant_name = variant.ident.to_string();
        let variant_qn = format!("{qn}::{variant_name}");
        let variant_id = add_node(
            context,
            "function",
            &variant_qn,
            &variant_name,
            source,
            variant.span(),
            "enum-variant",
        );
        context.constructors.insert(variant_qn, variant_id.clone());
        context
            .constructor_types
            .insert(variant_id.clone(), qn.clone());
        crate::relationships::define_and_contain(
            context,
            &id,
            &variant_id,
            variant.span(),
            &source.relative,
        );
    }
}

pub(crate) fn value_type(
    context: &mut Context,
    name: &syn::Ident,
    value_type: &syn::Type,
    owner: &str,
    module: &str,
    source: &SourceFile,
) {
    let name_text = name.to_string();
    let qn = format!("{module}::{name_text}");
    let id = add_node(
        context,
        "constant",
        &qn,
        &name_text,
        source,
        name.span(),
        "constant",
    );
    context.symbols.insert(qn.clone(), id.clone());
    crate::relationships::define_and_contain(context, owner, &id, name.span(), &source.relative);
    context.value_types.insert(qn, type_tokens(value_type));
}

pub(crate) fn alias(
    context: &mut Context,
    item: &syn::ItemType,
    owner: &str,
    module: &str,
    source: &SourceFile,
) {
    let name = item.ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(
        context,
        "type",
        &qn,
        &name,
        source,
        item.span(),
        "type-alias",
    );
    context.symbols.insert(qn.clone(), id.clone());
    context.type_aliases.insert(qn, type_tokens(&item.ty));
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
}

pub(crate) fn function(
    context: &mut Context,
    item: &syn::ItemFn,
    owner: &str,
    module: &str,
    crate_qn: &str,
    source: &SourceFile,
) {
    let name = item.sig.ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(
        context,
        "function",
        &qn,
        &name,
        source,
        item.span(),
        "function",
    );
    context.symbols.insert(qn.clone(), id.clone());
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
    function_index::register(
        context,
        Registration {
            id,
            qn,
            module_qn: module,
            crate_qn,
            source_path: &source.relative,
            signature: &item.sig,
            body: FunctionBody::Block((*item.block).clone()),
            self_type: None,
            trait_path: None,
        },
    );
}

pub(crate) fn trait_decl(
    context: &mut Context,
    item: &syn::ItemTrait,
    owner: &str,
    module: &str,
    crate_qn: &str,
    source: &SourceFile,
) {
    let name = item.ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(context, "trait", &qn, &name, source, item.span(), "trait");
    context.symbols.insert(qn.clone(), id.clone());
    context.traits.insert(qn.clone(), id.clone());
    context.trait_qn_by_id.insert(id.clone(), qn.clone());
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
    for trait_item in &item.items {
        if let syn::TraitItem::Fn(method) = trait_item {
            let method_name = method.sig.ident.to_string();
            let method_qn = format!("{qn}::{method_name}");
            let method_id = add_node(
                context,
                "method",
                &method_qn,
                &method_name,
                source,
                method.span(),
                "trait-method",
            );
            context.symbols.insert(method_qn.clone(), method_id.clone());
            context.trait_method_ids.insert(method_id.clone());
            context
                .trait_method_index
                .entry((id.clone(), method_name.clone()))
                .or_default()
                .push(method_id.clone());
            context
                .function_qn_by_id
                .insert(method_id.clone(), method_qn.clone());
            crate::relationships::define_and_contain(
                context,
                &id,
                &method_id,
                method.span(),
                &source.relative,
            );
            context.methods.push(MethodInfo {
                id: method_id.clone(),
                self_type: "Self".into(),
                trait_path: Some(qn.clone()),
                name: method_name,
                module_qn: module.into(),
                crate_qn: crate_qn.into(),
            });
            if let Some(block) = &method.default {
                function_index::register(
                    context,
                    Registration {
                        id: method_id,
                        qn: method_qn,
                        module_qn: module,
                        crate_qn,
                        source_path: &source.relative,
                        signature: &method.sig,
                        body: FunctionBody::Block(block.clone()),
                        self_type: Some("Self".into()),
                        trait_path: Some(qn.clone()),
                    },
                );
            }
        }
    }
}

pub(crate) fn macro_decl(
    context: &mut Context,
    item: &syn::ItemMacro,
    owner: &str,
    module: &str,
    source: &SourceFile,
) {
    let Some(ident) = &item.ident else {
        if normalized_tokens(&item.mac.path) == "thread_local" {
            if let Ok(file) = syn::parse2::<syn::File>(item.mac.tokens.clone()) {
                for nested in file.items {
                    if let syn::Item::Static(value) = nested {
                        value_type(context, &value.ident, &value.ty, owner, module, source);
                    }
                }
            }
        }
        context.facts.add_unresolved(
            owner,
            "defines",
            &item.to_token_stream().to_string(),
            "generated-target",
            span_value(item.span(), &source.relative),
        );
        return;
    };
    let name = ident.to_string();
    let qn = format!("{module}::{name}");
    let id = add_node(
        context,
        "function",
        &qn,
        &name,
        source,
        item.span(),
        "macro",
    );
    context.macros.insert(qn, id.clone());
    crate::relationships::define_and_contain(context, owner, &id, item.span(), &source.relative);
}

fn register_type(context: &mut Context, qn: &str, id: &str) {
    context.symbols.insert(qn.into(), id.into());
    context.types.insert(qn.into(), id.into());
    context.type_qn_by_id.insert(id.into(), qn.into());
}

pub(crate) fn add_node(
    context: &mut Context,
    kind: &str,
    qn: &str,
    name: &str,
    source: &SourceFile,
    span: Span,
    language_kind: &str,
) -> String {
    let mut attributes = BTreeMap::new();
    attributes.insert("language_kind".into(), Value::String(language_kind.into()));
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
