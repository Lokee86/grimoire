use crate::call_model::CallEvent;
use crate::call_resolution;
use crate::flow::Analyzer;
use crate::model::ValueSet;
use crate::paths::span_value;
use crate::resolve;
use crate::syntax::type_tokens;
use quote::ToTokens;
use syn::spanned::Spanned;

pub(crate) fn evaluate_statement_macro(analyzer: &mut Analyzer<'_>, value: &syn::StmtMacro) {
    let resolution =
        call_resolution::macro_call(analyzer.context, analyzer.function, &value.mac.path);
    analyzer.result.calls.push(CallEvent {
        expression: value.to_token_stream().to_string(),
        span: span_value(value.span(), &analyzer.function.source_path),
        resolution,
    });
}

pub(crate) fn evaluate(analyzer: &mut Analyzer<'_>, expression: &syn::Expr) -> ValueSet {
    match expression {
        syn::Expr::Path(value) => crate::expr_values::path_value(analyzer, value),
        syn::Expr::Call(value) => crate::expr_values::call(analyzer, value),
        syn::Expr::MethodCall(value) => crate::expr_values::method_call(analyzer, value),
        syn::Expr::Closure(value) => crate::expr_values::closure_value(analyzer, value),
        syn::Expr::Block(value) => analyzer.eval_block(&value.block),
        syn::Expr::Async(value) => analyzer.eval_block(&value.block),
        syn::Expr::Unsafe(value) => analyzer.eval_block(&value.block),
        syn::Expr::Const(value) => analyzer.eval_block(&value.block),
        syn::Expr::Paren(value) => analyzer.eval_expr(&value.expr),
        syn::Expr::Group(value) => analyzer.eval_expr(&value.expr),
        syn::Expr::Reference(value) => analyzer.eval_expr(&value.expr),
        syn::Expr::Try(value) => analyzer.eval_expr(&value.expr),
        syn::Expr::Await(value) => analyzer.eval_expr(&value.base),
        syn::Expr::Cast(value) => {
            analyzer.eval_expr(&value.expr);
            resolve::value_from_type(analyzer.context, &type_tokens(&value.ty), analyzer.function)
        }
        syn::Expr::Struct(value) => crate::expr_values::structure(analyzer, value),
        syn::Expr::Field(value) => crate::expr_values::field(analyzer, value),
        syn::Expr::Assign(value) => crate::expr_values::assignment(analyzer, value),
        syn::Expr::Return(value) => {
            let result = value
                .expr
                .as_ref()
                .map(|expr| analyzer.eval_expr(expr))
                .unwrap_or_default();
            analyzer.result.return_value.merge(&result);
            result
        }
        syn::Expr::If(value) => crate::expr_values::conditional(analyzer, value),
        syn::Expr::Match(value) => crate::expr_values::match_expr(analyzer, value),
        syn::Expr::Loop(value) => analyzer.eval_block(&value.body),
        syn::Expr::While(value) => {
            analyzer.eval_expr(&value.cond);
            analyzer.eval_block(&value.body)
        }
        syn::Expr::ForLoop(value) => {
            let iter = analyzer.eval_expr(&value.expr);
            let item = iter
                .contained_values
                .first()
                .cloned()
                .unwrap_or_else(|| iter.clone());
            analyzer.bind_pattern(&value.pat, &item);
            analyzer.eval_block(&value.body)
        }
        syn::Expr::Macro(value) => {
            crate::expr_values::macro_call(analyzer, &value.mac.path, expression)
        }
        syn::Expr::Unary(value) => analyzer.eval_expr(&value.expr),
        syn::Expr::Binary(value) => {
            let mut result = analyzer.eval_expr(&value.left);
            result.merge(&analyzer.eval_expr(&value.right));
            if matches!(
                value.op,
                syn::BinOp::Eq(_)
                    | syn::BinOp::Lt(_)
                    | syn::BinOp::Le(_)
                    | syn::BinOp::Ne(_)
                    | syn::BinOp::Ge(_)
                    | syn::BinOp::Gt(_)
                    | syn::BinOp::And(_)
                    | syn::BinOp::Or(_)
            ) {
                return ValueSet {
                    builtin: true,
                    ..ValueSet::default()
                };
            }
            result
        }
        syn::Expr::Index(value) => {
            let base = analyzer.eval_expr(&value.expr);
            analyzer.eval_expr(&value.index);
            base.contained_values
                .first()
                .cloned()
                .unwrap_or_else(|| ValueSet {
                    types: base.contained_types.clone(),
                    builtin: base.builtin,
                    external: base.external,
                    unknown: base.contained_types.is_empty() && !base.builtin && !base.external,
                    ..ValueSet::default()
                })
        }
        syn::Expr::Array(value) => crate::expr_values::contained(analyzer, value.elems.iter()),
        syn::Expr::Tuple(value) => crate::expr_values::tuple(analyzer, value.elems.iter()),
        syn::Expr::Repeat(value) => {
            let item = analyzer.eval_expr(&value.expr);
            analyzer.eval_expr(&value.len);
            ValueSet {
                contained_types: item.types.union(&item.contained_types).cloned().collect(),
                contained_values: vec![item],
                builtin: true,
                ..ValueSet::default()
            }
        }
        syn::Expr::Let(value) => {
            let item = analyzer.eval_expr(&value.expr);
            analyzer.bind_pattern(&value.pat, &item);
            ValueSet {
                builtin: true,
                ..ValueSet::default()
            }
        }
        syn::Expr::Range(value) => {
            if let Some(start) = &value.start {
                analyzer.eval_expr(start);
            }
            if let Some(end) = &value.end {
                analyzer.eval_expr(end);
            }
            ValueSet {
                builtin: true,
                ..ValueSet::default()
            }
        }
        syn::Expr::Break(value) => value
            .expr
            .as_ref()
            .map(|expr| analyzer.eval_expr(expr))
            .unwrap_or_default(),
        syn::Expr::Yield(value) => value
            .expr
            .as_ref()
            .map(|expr| analyzer.eval_expr(expr))
            .unwrap_or_default(),
        syn::Expr::Lit(_) | syn::Expr::Infer(_) | syn::Expr::Continue(_) => ValueSet {
            builtin: true,
            ..ValueSet::default()
        },
        _ => ValueSet {
            unknown: true,
            ..ValueSet::default()
        },
    }
}
