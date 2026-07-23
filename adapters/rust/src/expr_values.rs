use crate::call_model::CallEvent;
use crate::call_resolution;
use crate::flow::Analyzer;
use crate::model::ValueSet;
use crate::paths::{span_start, span_value};
use crate::resolve;
use crate::syntax::normalized_tokens;
use quote::ToTokens;
use syn::spanned::Spanned;

pub(crate) fn path_value(analyzer: &Analyzer<'_>, value: &syn::ExprPath) -> ValueSet {
    if value.qself.is_none() && value.path.segments.len() == 1 {
        let name = value.path.segments[0].ident.to_string();
        if let Some(found) = analyzer.env.get(&name) {
            let mut found = found.clone();
            if found.callables.is_empty()
                && found.types.is_empty()
                && found.traits.is_empty()
                && found.unknown
            {
                found.dynamic_callable = true;
            }
            return found;
        }
    }
    let text = normalized_tokens(&value.path);
    if text == "None" {
        return ValueSet {
            builtin: true,
            ..ValueSet::default()
        };
    }
    let callable = call_resolution::function_item(analyzer.context, analyzer.function, &text);
    if !callable.callables.is_empty() {
        return callable;
    }
    let mut result = ValueSet::default();
    let resolved_qns = resolve::resolve_qns(analyzer.context, &text, analyzer.function);
    let mut found_value = false;
    for qn in &resolved_qns {
        if let Some(value_type) = analyzer.context.value_types.get(qn) {
            result.merge(&resolve::value_from_type(
                analyzer.context,
                value_type,
                analyzer.function,
            ));
            found_value = true;
        }
    }
    if found_value {
        return result;
    }
    result.types.extend(resolve::resolve_type_ids(
        analyzer.context,
        &text,
        analyzer.function,
    ));
    if result.types.is_empty() {
        let qns = resolve::resolve_qns(analyzer.context, &text, analyzer.function);
        for qn in qns {
            if let Some(constructor) = analyzer.context.constructors.get(&qn) {
                if let Some(type_qn) = analyzer.context.constructor_types.get(constructor) {
                    if let Some(type_id) = analyzer.context.types.get(type_qn) {
                        result.types.insert(type_id.clone());
                    }
                }
            }
        }
    }
    if result.types.is_empty() {
        result.builtin = resolve::is_builtin_path(analyzer.context, &text, analyzer.function);
        result.external = !result.builtin
            && resolve::is_external_path(analyzer.context, &text, analyzer.function);
    }
    result.unknown = result.types.is_empty() && !result.builtin && !result.external;
    result
}

pub(crate) fn call(analyzer: &mut Analyzer<'_>, value: &syn::ExprCall) -> ValueSet {
    let callee = analyzer.eval_expr(&value.func);
    let arguments: Vec<_> = value
        .args
        .iter()
        .map(|arg| analyzer.eval_expr(arg))
        .collect();
    let mut resolution = match value.func.as_ref() {
        syn::Expr::Path(path) => {
            call_resolution::path_call(analyzer.context, analyzer.function, path, &callee)
        }
        _ if !callee.callables.is_empty() => crate::call_model::CallResolution {
            possible: callee.callables.len() > 1 || callee.dynamic_callable,
            return_value: call_resolution::returns_for_targets(analyzer.context, &callee.callables),
            targets: callee.callables.clone(),
            ..Default::default()
        },
        _ => crate::call_model::CallResolution {
            reason: Some("dynamic-target"),
            ..Default::default()
        },
    };
    if let syn::Expr::Path(path) = value.func.as_ref() {
        let name = normalized_tokens(&path.path);
        if matches!(
            name.as_str(),
            "Some"
                | "Ok"
                | "Err"
                | "Box::new"
                | "Rc::new"
                | "Arc::new"
                | "RefCell::new"
                | "Mutex::new"
                | "RwLock::new"
        ) {
            let mut wrapped = ValueSet {
                builtin: true,
                ..ValueSet::default()
            };
            for argument in &arguments {
                wrapped
                    .contained_types
                    .extend(argument.types.iter().cloned());
                wrapped
                    .contained_types
                    .extend(argument.contained_types.iter().cloned());
                wrapped.callables.extend(argument.callables.iter().cloned());
                wrapped.contained_values.push(argument.clone());
            }
            resolution.return_value.merge(&wrapped);
        }
    }
    call_resolution::propagate_arguments(
        analyzer.context,
        &resolution.targets,
        &arguments,
        &mut analyzer.result.parameter_updates,
    );
    record(analyzer, value, resolution.clone());
    resolution.return_value
}

pub(crate) fn method_call(analyzer: &mut Analyzer<'_>, value: &syn::ExprMethodCall) -> ValueSet {
    let receiver = analyzer.eval_expr(&value.receiver);
    let args: Vec<_> = value
        .args
        .iter()
        .map(|arg| analyzer.eval_expr(arg))
        .collect();
    let method = value.method.to_string();
    let mut resolution =
        call_resolution::method_call(analyzer.context, analyzer.function, &receiver, &method);
    propagate_known_callback(analyzer, &method, &receiver, &args);
    merge_callback_return(analyzer, &method, &args, &mut resolution.return_value);
    let mut propagated = Vec::with_capacity(args.len() + 1);
    propagated.push(receiver);
    propagated.extend(args);
    call_resolution::propagate_arguments(
        analyzer.context,
        &resolution.targets,
        &propagated,
        &mut analyzer.result.parameter_updates,
    );
    record(analyzer, value, resolution.clone());
    resolution.return_value
}

pub(crate) fn closure_value(analyzer: &mut Analyzer<'_>, value: &syn::ExprClosure) -> ValueSet {
    let start = span_start(value.span());
    let Some(id) = analyzer
        .context
        .closure_ids
        .get(&(analyzer.function.source_path.clone(), start.0, start.1))
        .cloned()
    else {
        return ValueSet {
            dynamic_callable: true,
            unknown: true,
            ..ValueSet::default()
        };
    };
    let parameter_names: std::collections::BTreeSet<_> = analyzer
        .context
        .functions
        .get(&id)
        .into_iter()
        .flat_map(|function| function.parameters.iter())
        .map(|parameter| parameter.name.as_str())
        .collect();
    for (name, captured) in &analyzer.env {
        if !parameter_names.contains(name.as_str()) {
            analyzer
                .result
                .capture_updates
                .entry((id.clone(), name.clone()))
                .or_default()
                .merge(captured);
        }
    }
    ValueSet::callable(id, false)
}

pub(crate) fn structure(analyzer: &mut Analyzer<'_>, value: &syn::ExprStruct) -> ValueSet {
    for field in &value.fields {
        analyzer.eval_expr(&field.expr);
    }
    if let Some(rest) = &value.rest {
        analyzer.eval_expr(rest);
    }
    let mut result = ValueSet::default();
    result.types.extend(resolve::resolve_type_ids(
        analyzer.context,
        &normalized_tokens(&value.path),
        analyzer.function,
    ));
    if result.types.is_empty() {
        let text = normalized_tokens(&value.path);
        result.builtin = resolve::is_builtin_path(analyzer.context, &text, analyzer.function);
        result.external = !result.builtin
            && resolve::is_external_path(analyzer.context, &text, analyzer.function);
    }
    result.unknown = result.types.is_empty() && !result.builtin && !result.external;
    result
}

pub(crate) fn field(analyzer: &mut Analyzer<'_>, value: &syn::ExprField) -> ValueSet {
    let base = analyzer.eval_expr(&value.base);
    let member = match &value.member {
        syn::Member::Named(name) => name.to_string(),
        syn::Member::Unnamed(index) => index.index.to_string(),
    };
    let mut result = ValueSet::default();
    if let syn::Member::Unnamed(index) = &value.member {
        if let Some(element) = base.tuple_elements.get(index.index as usize) {
            result.merge(element);
        }
    }
    for type_id in &base.types {
        let Some(qn) = analyzer.context.type_qn_by_id.get(type_id) else {
            continue;
        };
        if let Some(field) = analyzer.context.fields.get(&(qn.clone(), member.clone())) {
            result.merge(&resolve::value_from_type(
                analyzer.context,
                &field.type_text,
                analyzer.function,
            ));
        }
    }
    if result.types.is_empty()
        && result.traits.is_empty()
        && result.callables.is_empty()
        && (base.builtin || base.external)
    {
        result.builtin = base.builtin;
        result.external = base.external;
    }
    result.unknown = result.types.is_empty()
        && result.traits.is_empty()
        && result.callables.is_empty()
        && !result.builtin
        && !result.external;
    result
}

pub(crate) fn assignment(analyzer: &mut Analyzer<'_>, value: &syn::ExprAssign) -> ValueSet {
    let right = analyzer.eval_expr(&value.right);
    if let syn::Expr::Path(path) = value.left.as_ref() {
        if path.path.segments.len() == 1 {
            analyzer.assign_name(&path.path.segments[0].ident.to_string(), &right);
        }
    } else {
        analyzer.eval_expr(&value.left);
    }
    right
}

pub(crate) fn conditional(analyzer: &mut Analyzer<'_>, value: &syn::ExprIf) -> ValueSet {
    analyzer.eval_expr(&value.cond);
    let mut result = analyzer.eval_block(&value.then_branch);
    if let Some((_, else_expr)) = &value.else_branch {
        result.merge(&analyzer.eval_expr(else_expr));
    }
    result
}

pub(crate) fn match_expr(analyzer: &mut Analyzer<'_>, value: &syn::ExprMatch) -> ValueSet {
    let input = analyzer.eval_expr(&value.expr);
    let mut result = ValueSet::default();
    for arm in &value.arms {
        analyzer.bind_pattern(&arm.pat, &input);
        if let Some((_, guard)) = &arm.guard {
            analyzer.eval_expr(guard);
        }
        result.merge(&analyzer.eval_expr(&arm.body));
    }
    result
}

pub(crate) fn macro_call(
    analyzer: &mut Analyzer<'_>,
    path: &syn::Path,
    expression: &syn::Expr,
) -> ValueSet {
    let resolution = call_resolution::macro_call(analyzer.context, analyzer.function, path);
    record(analyzer, expression, resolution);
    let name = normalized_tokens(path);
    ValueSet {
        builtin: matches!(
            name.split("::").last().unwrap_or_default(),
            "concat"
                | "env"
                | "format"
                | "format_args"
                | "include_bytes"
                | "include_str"
                | "stringify"
                | "vec"
                | "write"
                | "writeln"
        ),
        ..ValueSet::default()
    }
}

pub(crate) fn contained<'a>(
    analyzer: &mut Analyzer<'_>,
    values: impl Iterator<Item = &'a syn::Expr>,
) -> ValueSet {
    let mut result = ValueSet {
        builtin: true,
        ..ValueSet::default()
    };
    for value in values {
        let item = analyzer.eval_expr(value);
        result.contained_types.extend(item.types.iter().cloned());
        result
            .contained_types
            .extend(item.contained_types.iter().cloned());
        result.callables.extend(item.callables.iter().cloned());
        result.dynamic_callable |= item.dynamic_callable;
        result.contained_values.push(item);
    }
    result
}

pub(crate) fn tuple<'a>(
    analyzer: &mut Analyzer<'_>,
    values: impl Iterator<Item = &'a syn::Expr>,
) -> ValueSet {
    let tuple_elements: Vec<_> = values.map(|value| analyzer.eval_expr(value)).collect();
    let mut result = ValueSet {
        tuple_elements: tuple_elements.clone(),
        contained_values: tuple_elements,
        builtin: true,
        ..ValueSet::default()
    };
    for item in &result.tuple_elements {
        result.contained_types.extend(item.types.iter().cloned());
        result
            .contained_types
            .extend(item.contained_types.iter().cloned());
        result.callables.extend(item.callables.iter().cloned());
        result.dynamic_callable |= item.dynamic_callable;
    }
    result
}

fn propagate_known_callback(
    analyzer: &mut Analyzer<'_>,
    method: &str,
    receiver: &ValueSet,
    arguments: &[ValueSet],
) {
    let item = receiver
        .contained_values
        .first()
        .cloned()
        .unwrap_or_else(|| ValueSet {
            types: receiver.contained_types.clone(),
            builtin: receiver.builtin,
            external: receiver.external,
            unknown: receiver.contained_types.is_empty() && !receiver.builtin && !receiver.external,
            ..ValueSet::default()
        });
    let error = receiver
        .contained_values
        .get(1)
        .cloned()
        .unwrap_or_else(|| item.clone());
    if method == "map_or_else" {
        propagate_callback(analyzer, arguments.first(), &[]);
        propagate_callback(analyzer, arguments.get(1), std::slice::from_ref(&item));
        return;
    }
    let (callback_index, inputs): (usize, Vec<ValueSet>) = match method {
        "sort_by" => (0, vec![item.clone(), item]),
        "map" | "filter_map" | "and_then" | "filter" | "inspect" | "find" | "any" | "all"
        | "position" | "take_while" | "skip_while" | "sort_by_key" | "for_each" | "with"
        | "is_some_and" | "is_ok_and" | "is_err_and" => (0, vec![item]),
        "then" | "then_with" | "unwrap_or_else" => (0, Vec::new()),
        "map_err" | "or_else" => (0, vec![error]),
        "fold" | "try_fold" => {
            let Some(initial) = arguments.first() else {
                return;
            };
            (1, vec![initial.clone(), item])
        }
        _ => return,
    };
    propagate_callback(analyzer, arguments.get(callback_index), &inputs);
}

fn propagate_callback(
    analyzer: &mut Analyzer<'_>,
    callback: Option<&ValueSet>,
    inputs: &[ValueSet],
) {
    let Some(callback) = callback else {
        return;
    };
    for target in &callback.callables {
        for (index, input) in inputs.iter().enumerate() {
            analyzer
                .result
                .parameter_updates
                .entry((target.clone(), index))
                .or_default()
                .merge(input);
        }
    }
}

fn merge_callback_return(
    analyzer: &Analyzer<'_>,
    method: &str,
    arguments: &[ValueSet],
    output: &mut ValueSet,
) {
    let callback_index = match method {
        "map" | "filter_map" | "map_err" | "and_then" | "or_else" | "then" | "then_with"
        | "unwrap_or_else" => 0,
        "map_or_else" => 1,
        _ => return,
    };
    let Some(callback) = arguments.get(callback_index) else {
        return;
    };
    if callback.callables.is_empty() {
        return;
    }
    let returned = call_resolution::returns_for_targets(analyzer.context, &callback.callables);
    if returned == ValueSet::default() {
        return;
    }
    if matches!(method, "unwrap_or_else" | "map_or_else") {
        output.merge(&returned);
        return;
    }
    output.builtin = true;
    output
        .contained_types
        .extend(returned.types.iter().cloned());
    output
        .contained_types
        .extend(returned.contained_types.iter().cloned());
    output.contained_values.push(returned);
}

fn record<T: ToTokens + Spanned>(
    analyzer: &mut Analyzer<'_>,
    expression: &T,
    resolution: crate::call_model::CallResolution,
) {
    analyzer.result.calls.push(CallEvent {
        expression: expression.to_token_stream().to_string(),
        span: span_value(expression.span(), &analyzer.function.source_path),
        resolution,
    });
}
