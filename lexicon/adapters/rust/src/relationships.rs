use crate::model::Context;
use crate::paths::span_value;
use proc_macro2::Span;

pub(crate) fn finalize(context: &mut Context) {
    crate::imports::resolve_all(context);
    crate::implementations::finalize(context);
    crate::semantic::analyze(context);
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
