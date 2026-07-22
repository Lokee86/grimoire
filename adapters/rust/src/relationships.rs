use crate::model::{CallForm, Context, CrateContext, PendingCall, SourceFile};
use crate::paths::{span_start, span_value};
use crate::resolve;
use proc_macro2::Span;
use quote::ToTokens;
use serde_json::Value;
use std::collections::BTreeMap;
use syn::spanned::Spanned;
use syn::visit::{self, Visit};
use syn::{Expr, ExprPath, ImplItem};

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
            if trait_text.is_none() {
                if let LocalLookup::Found(type_id) =
                    lookup_local(&context.types, &self_text, module_qn, &crate_context.qn)
                {
                    context
                        .inherent_methods
                        .entry(format!("{type_id}::{name}"))
                        .or_default()
                        .push(id.clone());
                }
            }
            define_and_contain(
                context,
                method_owner,
                &id,
                function.span(),
                &source.relative,
            );
            collect_calls(
                context,
                &function.block,
                &id,
                module_qn,
                &crate_context.qn,
                &source.relative,
            );
        }
    }
}

pub(crate) fn collect_calls(
    context: &mut Context,
    block: &syn::Block,
    owner_id: &str,
    module_qn: &str,
    crate_qn: &str,
    source_path: &str,
) {
    let mut visitor = CallVisitor { calls: Vec::new() };
    visitor.visit_block(block);
    context
        .pending_calls
        .extend(visitor.calls.into_iter().map(|call| PendingCall {
            owner_id: owner_id.into(),
            module_qn: module_qn.into(),
            crate_qn: crate_qn.into(),
            form: call.form,
            path: call.path,
            expression: call.expression,
            span: span_value(call.span, source_path),
        }));
}

pub(crate) fn resolve_calls(context: &mut Context) {
    let calls = std::mem::take(&mut context.pending_calls);
    for call in calls {
        match call.form {
            CallForm::Path => match resolve_free_function(
                context,
                call.path.as_deref().unwrap_or_default(),
                &call.module_qn,
                &call.crate_qn,
            ) {
                Ok(target) => context
                    .facts
                    .add_edge(&call.owner_id, &target, "calls", call.span),
                Err(reason) => context.facts.add_unresolved(
                    &call.owner_id,
                    "calls",
                    &call.expression,
                    reason,
                    call.span,
                ),
            },
            CallForm::Associated => context.facts.add_unresolved(
                &call.owner_id,
                "calls",
                &call.expression,
                "associated-target",
                call.span,
            ),
            CallForm::Method => context.facts.add_unresolved(
                &call.owner_id,
                "calls",
                &call.expression,
                "method-call",
                call.span,
            ),
            CallForm::Macro => context.facts.add_unresolved(
                &call.owner_id,
                "calls",
                &call.expression,
                "macro-call",
                call.span,
            ),
            CallForm::Unsupported => context.facts.add_unresolved(
                &call.owner_id,
                "calls",
                &call.expression,
                "unsupported-form",
                call.span,
            ),
        }
    }
}

struct CollectedCall {
    form: CallForm,
    path: Option<String>,
    expression: String,
    span: Span,
}

struct CallVisitor {
    calls: Vec<CollectedCall>,
}

impl<'ast> Visit<'ast> for CallVisitor {
    fn visit_expr_call(&mut self, expression: &'ast syn::ExprCall) {
        let (form, path) = match expression.func.as_ref() {
            Expr::Path(ExprPath {
                qself: None, path, ..
            }) => (CallForm::Path, Some(resolve::normalized_tokens(path))),
            Expr::Path(_) => (CallForm::Associated, None),
            _ => (CallForm::Unsupported, None),
        };
        self.calls.push(CollectedCall {
            form,
            path,
            expression: expression.to_token_stream().to_string(),
            span: expression.span(),
        });
        visit::visit_expr_call(self, expression);
    }

    fn visit_expr_method_call(&mut self, expression: &'ast syn::ExprMethodCall) {
        self.calls.push(CollectedCall {
            form: CallForm::Method,
            path: None,
            expression: expression.to_token_stream().to_string(),
            span: expression.span(),
        });
        visit::visit_expr_method_call(self, expression);
    }

    fn visit_expr_macro(&mut self, expression: &'ast syn::ExprMacro) {
        self.calls.push(CollectedCall {
            form: CallForm::Macro,
            path: None,
            expression: expression.to_token_stream().to_string(),
            span: expression.span(),
        });
        visit::visit_expr_macro(self, expression);
    }

    fn visit_stmt_macro(&mut self, statement: &'ast syn::StmtMacro) {
        self.calls.push(CollectedCall {
            form: CallForm::Macro,
            path: None,
            expression: statement.to_token_stream().to_string(),
            span: statement.span(),
        });
        visit::visit_stmt_macro(self, statement);
    }
}

fn resolve_free_function(
    context: &Context,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Result<String, &'static str> {
    if path.is_empty() || path.contains('<') || path.contains('>') {
        return Err("unsupported-form");
    }
    let absolute = path.starts_with("::");
    let path = path.trim_start_matches("::");
    if path.is_empty() {
        return Err("unsupported-form");
    }
    if is_external_root(path) {
        return Err("external-target");
    }
    if absolute {
        return Err("unsupported-form");
    }
    if let Some(rest) = path.strip_prefix("crate::") {
        return resolve_rooted_function(context, rest, crate_qn, "crate", module_qn, crate_qn);
    }
    if let Some(rest) = path.strip_prefix("self::") {
        return resolve_rooted_function(context, rest, module_qn, "self", module_qn, crate_qn);
    }
    if let Some(result) = resolve_associated_call(context, path, module_qn, crate_qn) {
        return result;
    }
    if path.starts_with("super::") || path.contains("::") {
        return if is_associated_path(context, path, module_qn, crate_qn) {
            Err("associated-target")
        } else if is_local_module_path(context, path, module_qn, crate_qn) {
            Err("unsupported-form")
        } else {
            Err("external-target")
        };
    }

    let mut base = module_qn.to_string();
    loop {
        let candidate = format!("{base}::{path}");
        if let Some(target) = free_function_at(context, &candidate) {
            return Ok(target);
        }
        if base == crate_qn {
            break;
        }
        let Some(parent) = base.rsplit_once("::").map(|(parent, _)| parent) else {
            break;
        };
        base = parent.to_string();
    }

    let candidates: Vec<_> = context
        .symbols
        .iter()
        .filter(|(candidate, id)| {
            is_free_function(context, id) && candidate.ends_with(&format!("::{path}"))
        })
        .map(|(_, id)| id.clone())
        .collect();
    match candidates.as_slice() {
        [target] => Ok(target.clone()),
        [] => {
            if context
                .symbols
                .get(path)
                .is_some_and(|id| !is_free_function(context, id))
            {
                Err("associated-target")
            } else {
                Err("missing-target")
            }
        }
        _ => Err("ambiguous-target"),
    }
}

fn resolve_rooted_function(
    context: &Context,
    rest: &str,
    base: &str,
    root: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Result<String, &'static str> {
    if rest.is_empty() {
        return Err("unsupported-form");
    }
    let candidate = format!("{base}::{rest}");
    if let Some(target) = free_function_at(context, &candidate) {
        return Ok(target);
    }
    let path = format!("{root}::{rest}");
    if let Some(result) = resolve_associated_call(context, &path, module_qn, crate_qn) {
        result
    } else {
        Err("missing-target")
    }
}

enum LocalLookup {
    Found(String),
    Ambiguous,
    Missing,
}

fn resolve_associated_call(
    context: &Context,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> Option<Result<String, &'static str>> {
    let (type_path, method_name) = split_associated_path(path)?;
    match lookup_local(&context.types, type_path, module_qn, crate_qn) {
        LocalLookup::Found(type_id) => {
            let key = format!("{type_id}::{method_name}");
            match context.inherent_methods.get(&key).map(Vec::as_slice) {
                Some([target]) => Some(Ok(target.clone())),
                Some([]) => Some(Err("missing-target")),
                Some(_) => Some(Err("ambiguous-target")),
                None => Some(Err("missing-target")),
            }
        }
        LocalLookup::Ambiguous => Some(Err("ambiguous-target")),
        LocalLookup::Missing => match lookup_local(&context.traits, type_path, module_qn, crate_qn)
        {
            LocalLookup::Found(_) | LocalLookup::Ambiguous => Some(Err("associated-target")),
            LocalLookup::Missing => Some(Err(
                if type_path.starts_with("crate::") || type_path.starts_with("self::") {
                    "missing-target"
                } else {
                    "external-target"
                },
            )),
        },
    }
}

fn split_associated_path(path: &str) -> Option<(&str, &str)> {
    let (type_path, method_name) = path.rsplit_once("::")?;
    if type_path.is_empty() || method_name.is_empty() {
        return None;
    }
    let segments: Vec<_> = type_path.split("::").collect();
    let allowed = match segments.as_slice() {
        [_] => true,
        ["self", ..] | ["crate", ..] => segments.len() >= 2,
        _ => false,
    };
    allowed.then_some((type_path, method_name))
}

fn lookup_local(
    map: &BTreeMap<String, String>,
    path: &str,
    module_qn: &str,
    crate_qn: &str,
) -> LocalLookup {
    if let Some(rest) = path.strip_prefix("crate::") {
        return map
            .get(&format!("{crate_qn}::{rest}"))
            .cloned()
            .map(LocalLookup::Found)
            .unwrap_or(LocalLookup::Missing);
    }
    if let Some(rest) = path.strip_prefix("self::") {
        return map
            .get(&format!("{module_qn}::{rest}"))
            .cloned()
            .map(LocalLookup::Found)
            .unwrap_or(LocalLookup::Missing);
    }
    if path.contains("::") || path.is_empty() {
        return LocalLookup::Missing;
    }

    let mut base = module_qn.to_string();
    loop {
        if let Some(target) = map.get(&format!("{base}::{path}")) {
            return LocalLookup::Found(target.clone());
        }
        if base == crate_qn {
            break;
        }
        let Some(parent) = base.rsplit_once("::").map(|(parent, _)| parent) else {
            break;
        };
        base = parent.to_string();
    }

    let candidates: Vec<_> = map
        .iter()
        .filter(|(candidate, _)| candidate.ends_with(&format!("::{path}")))
        .map(|(_, target)| target.clone())
        .collect();
    match candidates.as_slice() {
        [target] => LocalLookup::Found(target.clone()),
        [] => LocalLookup::Missing,
        _ => LocalLookup::Ambiguous,
    }
}

fn free_function_at(context: &Context, qn: &str) -> Option<String> {
    context
        .symbols
        .get(qn)
        .filter(|id| is_free_function(context, id))
        .cloned()
}

fn is_free_function(context: &Context, id: &str) -> bool {
    context
        .facts
        .nodes
        .get(id)
        .and_then(|node| node.get("kind"))
        .and_then(Value::as_str)
        == Some("function")
}

fn is_associated_path(context: &Context, path: &str, module_qn: &str, crate_qn: &str) -> bool {
    let Some((type_path, _)) = path.rsplit_once("::") else {
        return path == "Self";
    };
    type_path == "Self" || resolve::resolve_type(context, type_path, module_qn, crate_qn).is_some()
}

fn is_local_module_path(context: &Context, path: &str, module_qn: &str, crate_qn: &str) -> bool {
    let mut base = module_qn.to_string();
    loop {
        if context.modules.contains_key(&format!("{base}::{path}")) {
            return true;
        }
        if base == crate_qn {
            return false;
        }
        let Some(parent) = base.rsplit_once("::").map(|(parent, _)| parent) else {
            return false;
        };
        base = parent.to_string();
    }
}

fn is_external_root(path: &str) -> bool {
    matches!(path.split("::").next(), Some("std" | "core" | "alloc"))
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
    if let Some(paths) = resolve::use_paths(&item_use.tree) {
        let crate_qn = module_qn.split("::").take(2).collect::<Vec<_>>().join("::");
        let targets: Vec<_> = paths
            .iter()
            .filter_map(|path| resolve::resolve_symbol(context, path, module_qn, &crate_qn))
            .collect();
        if targets.is_empty() {
            let (expression, reason) = if paths.len() == 1 {
                let path = &paths[0];
                (
                    path,
                    if path.starts_with("std::") || path.starts_with("core::") {
                        "external-target"
                    } else {
                        "missing-target"
                    },
                )
            } else {
                (&name, "unsupported-form")
            };
            context.facts.add_unresolved(
                owner_id,
                "imports",
                expression,
                reason,
                span_value(item_use.span(), &source.relative),
            );
        } else {
            for target in targets {
                context.facts.add_edge(
                    owner_id,
                    &target,
                    "imports",
                    span_value(item_use.span(), &source.relative),
                );
            }
            if paths
                .iter()
                .any(|path| resolve::resolve_symbol(context, path, module_qn, &crate_qn).is_none())
            {
                context.facts.add_unresolved(
                    owner_id,
                    "imports",
                    &name,
                    "missing-target",
                    span_value(item_use.span(), &source.relative),
                );
            }
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
