use anyhow::{Context as AnyhowContext, Result};
use cargo_metadata::{Metadata, MetadataCommand};
use clap::Parser;
use proc_macro2::Span;
use quote::ToTokens;
use serde_json::{Map, Value};
use sha2::{Digest, Sha256};
use std::collections::{BTreeMap, BTreeSet, HashSet};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use syn::spanned::Spanned;
use syn::{ImplItem, Item, TraitItem, UseTree};

type JsonMap = Map<String, Value>;

#[derive(Debug, Parser)]
#[command(
    name = "lexicon-rust-adapter",
    about = "Emit Lexicon facts v1 for a Rust repository"
)]
struct Args {
    #[arg(long)]
    repo: PathBuf,
    #[arg(long)]
    output: PathBuf,
}

#[derive(Clone)]
struct SourceFile {
    path: PathBuf,
    relative: String,
    content: Vec<u8>,
    syntax: syn::File,
}

#[derive(Clone)]
struct CrateContext {
    qn: String,
    node_id: String,
    root: PathBuf,
    package_root: PathBuf,
}

struct Facts {
    nodes: BTreeMap<String, Value>,
    edges: BTreeMap<String, Value>,
    unresolved: BTreeMap<String, Value>,
}

impl Facts {
    fn new() -> Self {
        Self {
            nodes: BTreeMap::new(),
            edges: BTreeMap::new(),
            unresolved: BTreeMap::new(),
        }
    }

    fn add_node(
        &mut self,
        language: &str,
        kind: &str,
        canonical: &str,
        name: &str,
        path: &str,
        qualified_name: &str,
        content_id: Option<String>,
        span: Option<Value>,
        attributes: BTreeMap<String, Value>,
    ) -> String {
        let id = stable_id(language, kind, canonical);
        let mut node = JsonMap::new();
        node.insert("attributes".into(), object_from_btree(attributes));
        if let Some(content_id) = content_id {
            node.insert("content_id".into(), Value::String(content_id));
        }
        node.insert("id".into(), Value::String(id.clone()));
        node.insert("kind".into(), Value::String(kind.into()));
        node.insert("name".into(), Value::String(name.into()));
        node.insert("path".into(), Value::String(path.into()));
        node.insert(
            "qualified_name".into(),
            Value::String(qualified_name.into()),
        );
        node.insert("record".into(), Value::String("node".into()));
        if let Some(span) = span {
            node.insert("span".into(), span);
        }
        self.nodes.entry(id.clone()).or_insert(Value::Object(node));
        id
    }

    fn add_edge(&mut self, source: &str, target: &str, relation: &str, span: Option<Value>) {
        let mut edge = JsonMap::new();
        edge.insert("record".into(), Value::String("edge".into()));
        edge.insert("relation".into(), Value::String(relation.into()));
        edge.insert("source".into(), Value::String(source.into()));
        if let Some(span) = span.clone() {
            edge.insert("span".into(), span);
        }
        edge.insert("target".into(), Value::String(target.into()));
        let key = format!("{source}\0{target}\0{relation}\0{}", span_key(&span));
        self.edges.entry(key).or_insert(Value::Object(edge));
    }

    fn add_unresolved(
        &mut self,
        source: &str,
        relation: &str,
        expression: &str,
        reason: &str,
        span: Option<Value>,
    ) {
        let mut record = JsonMap::new();
        record.insert("expression".into(), Value::String(expression.into()));
        record.insert("reason".into(), Value::String(reason.into()));
        record.insert("record".into(), Value::String("unresolved".into()));
        record.insert("relation".into(), Value::String(relation.into()));
        record.insert("source".into(), Value::String(source.into()));
        if let Some(span) = span.clone() {
            record.insert("span".into(), span);
        }
        let key = format!(
            "{source}\0{relation}\0{expression}\0{reason}\0{}",
            span_key(&span)
        );
        self.unresolved.entry(key).or_insert(Value::Object(record));
    }
}

struct Context {
    repo: PathBuf,
    repository: String,
    sources: BTreeMap<PathBuf, SourceFile>,
    facts: Facts,
    crates: Vec<CrateContext>,
    modules: BTreeMap<String, String>,
    symbols: BTreeMap<String, String>,
    types: BTreeMap<String, String>,
    traits: BTreeMap<String, String>,
    processed: HashSet<(PathBuf, String)>,
}

fn main() -> Result<()> {
    let args = Args::parse();
    let repo = args
        .repo
        .canonicalize()
        .with_context(|| format!("cannot resolve repository {}", args.repo.display()))?;
    let output = if args.output.is_absolute() {
        args.output
    } else {
        std::env::current_dir()?.join(args.output)
    };
    let jsonl = generate(&repo)?;
    if let Some(parent) = output.parent() {
        fs::create_dir_all(parent)?;
    }
    fs::File::create(&output)?.write_all(jsonl.as_bytes())?;
    Ok(())
}

fn generate(repo: &Path) -> Result<String> {
    let manifest = repo.join("Cargo.toml");
    let metadata = MetadataCommand::new()
        .manifest_path(&manifest)
        .no_deps()
        .exec()
        .with_context(|| format!("cargo metadata failed for {}", manifest.display()))?;
    let repository = repository_identity(repo, &metadata);
    let sources = scan_sources(repo)?;
    let mut context = Context {
        repo: repo.to_path_buf(),
        repository: repository.clone(),
        sources,
        facts: Facts::new(),
        crates: Vec::new(),
        modules: BTreeMap::new(),
        symbols: BTreeMap::new(),
        types: BTreeMap::new(),
        traits: BTreeMap::new(),
        processed: HashSet::new(),
    };
    add_repository_and_files(&mut context);
    add_crates(&mut context, &metadata);

    let crates = context.crates.clone();
    for crate_context in &crates {
        if context.sources.contains_key(&crate_context.root) {
            process_file(
                &mut context,
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
                &mut context,
                &path,
                &crate_context.node_id,
                &crate_context.qn,
                crate_context,
            );
        } else {
            context.facts.add_unresolved(
                &repository_node(&repository),
                "contains",
                &path.display().to_string(),
                "unsupported-form",
                None,
            );
        }
    }

    render(&context, &repository)
}

fn repository_identity(repo: &Path, metadata: &Metadata) -> String {
    if metadata.packages.len() == 1 {
        return metadata.packages[0].name.clone();
    }
    repo.file_name()
        .and_then(|name| name.to_str())
        .filter(|name| !name.is_empty())
        .unwrap_or("repository")
        .to_string()
}

fn add_repository_and_files(context: &mut Context) {
    let repo_id = context.facts.add_node(
        "rust",
        "repository",
        &context.repository,
        &context.repository,
        ".",
        &context.repository,
        None,
        None,
        BTreeMap::new(),
    );
    let mut directories = BTreeSet::new();
    for source in context.sources.values() {
        let mut current = Path::new(source.relative.as_str()).parent();
        while let Some(path) = current {
            if path.as_os_str().is_empty() || path == Path::new(".") {
                break;
            }
            directories.insert(path.to_path_buf());
            current = path.parent();
        }
    }
    for directory in directories {
        let relative = normalize_path(&directory);
        let name = directory
            .file_name()
            .and_then(|value| value.to_str())
            .unwrap_or(&relative);
        let id = context.facts.add_node(
            "rust",
            "directory",
            &relative,
            name,
            &relative,
            &relative,
            None,
            None,
            BTreeMap::new(),
        );
        let parent = directory
            .parent()
            .filter(|path| !path.as_os_str().is_empty());
        let parent_id = parent
            .map(|path| stable_id("rust", "directory", &normalize_path(path)))
            .unwrap_or_else(|| repo_id.clone());
        context.facts.add_edge(&parent_id, &id, "contains", None);
    }
    for source in context.sources.values() {
        let file_id = context.facts.add_node(
            "rust",
            "file",
            &source.relative,
            Path::new(&source.relative)
                .file_name()
                .and_then(|value| value.to_str())
                .unwrap_or(&source.relative),
            &source.relative,
            &source.relative,
            Some(content_id(&source.content)),
            None,
            BTreeMap::new(),
        );
        let parent = Path::new(&source.relative).parent();
        let parent_id = parent
            .filter(|path| !path.as_os_str().is_empty())
            .map(|path| stable_id("rust", "directory", &normalize_path(path)))
            .unwrap_or(repo_id.clone());
        context
            .facts
            .add_edge(&parent_id, &file_id, "contains", None);
    }
}

fn add_crates(context: &mut Context, metadata: &Metadata) {
    let mut packages = metadata.packages.clone();
    packages.sort_by(|left, right| left.name.cmp(&right.name));
    for package in packages {
        let manifest_path = PathBuf::from(package.manifest_path.as_std_path());
        let package_root = fs::canonicalize(&manifest_path)
            .unwrap_or_else(|_| manifest_path.clone())
            .parent()
            .unwrap_or(Path::new("."))
            .to_path_buf();
        let mut targets = package.targets.clone();
        targets.sort_by(|left, right| left.name.cmp(&right.name));
        for target in targets {
            let supported = target.kind.iter().any(|kind| {
                matches!(
                    kind.to_string().as_str(),
                    "lib" | "bin" | "example" | "test" | "bench"
                )
            });
            if !supported {
                continue;
            }
            let root_path = PathBuf::from(target.src_path.as_std_path());
            let root = fs::canonicalize(&root_path).unwrap_or(root_path);
            let Some(root) = source_path_for(context, &root) else {
                continue;
            };
            let qn = format!("{}::{}", package.name, target.name);
            let path = relative_path(&context.repo, package_root.join("Cargo.toml"));
            let mut attributes = BTreeMap::new();
            attributes.insert("package".into(), Value::String(package.name.clone()));
            attributes.insert(
                "target_kind".into(),
                Value::String(
                    target
                        .kind
                        .first()
                        .map(|kind| kind.to_string().to_lowercase())
                        .unwrap_or_else(|| "crate".into()),
                ),
            );
            let node_id = context.facts.add_node(
                "rust",
                "module",
                &format!("crate:{qn}"),
                &target.name,
                &path,
                &qn,
                None,
                None,
                attributes,
            );
            context.modules.insert(qn.clone(), node_id.clone());
            context.crates.push(CrateContext {
                qn,
                node_id,
                root,
                package_root: package_root.clone(),
            });
        }
    }
}

fn process_file(
    context: &mut Context,
    path: &Path,
    owner_id: &str,
    module_qn: &str,
    crate_context: &CrateContext,
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
    process_items(context, &items, &file_id, module_qn, &source, crate_context);
}

fn process_items(
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
                define_and_contain(
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
                } else if let Some(child_path) = resolve_module_file(context, &source.path, &name) {
                    process_file(context, &child_path, &module_id, &child_qn, crate_context);
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
                define_and_contain(context, owner_id, &id, item_struct.span(), &source.relative);
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
                define_and_contain(context, owner_id, &id, item_enum.span(), &source.relative);
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
                define_and_contain(context, owner_id, &id, item_trait.span(), &source.relative);
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
                        define_and_contain(
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
                define_and_contain(context, owner_id, &id, function.span(), &source.relative);
            }
            Item::Impl(item_impl) => process_impl(
                context,
                item_impl,
                owner_id,
                module_qn,
                source,
                crate_context,
            ),
            Item::Use(item_use) => process_use(context, item_use, owner_id, module_qn, source),
            Item::Macro(item_macro) => {
                context.facts.add_unresolved(
                    owner_id,
                    "defines",
                    &item_macro.to_token_stream().to_string(),
                    "generated-target",
                    span_value(item_macro.span(), &source.relative),
                );
            }
            _ => {}
        }
    }
}

fn process_impl(
    context: &mut Context,
    item_impl: &syn::ItemImpl,
    owner_id: &str,
    module_qn: &str,
    source: &SourceFile,
    crate_context: &CrateContext,
) {
    let self_text = normalized_tokens(&item_impl.self_ty);
    let self_id = resolve_type(context, &self_text, module_qn, &crate_context.qn);
    let trait_text = item_impl
        .trait_
        .as_ref()
        .map(|(_, path, _)| normalized_tokens(path));
    if let Some(trait_text) = &trait_text {
        let trait_id = resolve_trait(context, trait_text, module_qn, &crate_context.qn);
        match (self_id.clone(), trait_id) {
            (Some(self_id), Some(trait_id)) => {
                context.facts.add_edge(
                    &self_id,
                    &trait_id,
                    "implements",
                    span_value(item_impl.span(), &source.relative),
                );
            }
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

fn process_use(
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
    if let Some(path) = simple_use_path(&item_use.tree) {
        if let Some(target) = resolve_symbol(
            context,
            &path,
            module_qn,
            &module_qn.split("::").take(2).collect::<Vec<_>>().join("::"),
        ) {
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

fn define_and_contain(context: &mut Context, owner: &str, target: &str, span: Span, path: &str) {
    let span = span_value(span, path);
    context
        .facts
        .add_edge(owner, target, "contains", span.clone());
    context.facts.add_edge(owner, target, "defines", span);
}

fn resolve_module_file(context: &Context, source: &Path, name: &str) -> Option<PathBuf> {
    let parent = source.parent()?;
    let stem = source.file_stem()?.to_str()?;
    let mut bases = vec![parent.to_path_buf()];
    if stem != "lib" && stem != "main" && stem != "mod" {
        bases.push(parent.join(stem));
    }
    for base in bases {
        for candidate in [
            base.join(format!("{name}.rs")),
            base.join(name).join("mod.rs"),
        ] {
            if context.sources.contains_key(&candidate) {
                return Some(candidate);
            }
        }
    }
    None
}

fn resolve_type(context: &Context, path: &str, module_qn: &str, crate_qn: &str) -> Option<String> {
    resolve_from_map(&context.types, path, module_qn, crate_qn)
}

fn resolve_trait(context: &Context, path: &str, module_qn: &str, crate_qn: &str) -> Option<String> {
    resolve_from_map(&context.traits, path, module_qn, crate_qn)
}

fn resolve_symbol(
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

fn simple_use_path(tree: &UseTree) -> Option<String> {
    let tokens = tree.to_token_stream().to_string();
    if tokens.contains('{') || tokens.contains('*') || tokens.contains(" as ") {
        return None;
    }
    Some(tokens.split_whitespace().collect::<String>())
}

fn normalized_tokens<T: ToTokens>(value: &T) -> String {
    value
        .to_token_stream()
        .to_string()
        .split_whitespace()
        .collect()
}

fn source_path_for(context: &Context, candidate: &Path) -> Option<PathBuf> {
    let candidate = comparable_path(candidate);
    context
        .sources
        .keys()
        .find(|path| comparable_path(path) == candidate)
        .cloned()
}

fn comparable_path(path: &Path) -> String {
    let value = path.to_string_lossy().replace('\\', "/");
    value
        .strip_prefix("//?/")
        .unwrap_or(&value)
        .to_ascii_lowercase()
}

fn scan_sources(repo: &Path) -> Result<BTreeMap<PathBuf, SourceFile>> {
    let mut paths = Vec::new();
    collect_rust_files(repo, repo, &mut paths)?;
    paths.sort();
    let mut sources = BTreeMap::new();
    for path in paths {
        let content = fs::read(&path)?;
        let syntax = syn::parse_file(
            std::str::from_utf8(&content)
                .with_context(|| format!("Rust source is not UTF-8: {}", path.display()))?,
        )
        .with_context(|| format!("cannot parse Rust source {}", path.display()))?;
        let relative = relative_path(repo, &path);
        sources.insert(
            path.clone(),
            SourceFile {
                path,
                relative,
                content,
                syntax,
            },
        );
    }
    Ok(sources)
}

fn collect_rust_files(root: &Path, directory: &Path, output: &mut Vec<PathBuf>) -> Result<()> {
    let mut entries: Vec<_> = fs::read_dir(directory)?.collect::<std::io::Result<Vec<_>>>()?;
    entries.sort_by_key(|entry| entry.file_name());
    for entry in entries {
        let path = entry.path();
        let file_type = entry.file_type()?;
        if file_type.is_dir() {
            if is_excluded(root, &path) {
                continue;
            }
            collect_rust_files(root, &path, output)?;
        } else if file_type.is_file() && path.extension().and_then(|ext| ext.to_str()) == Some("rs")
        {
            output.push(path);
        }
    }
    Ok(())
}

fn is_excluded(root: &Path, path: &Path) -> bool {
    let defaults = [
        ".git",
        ".worktrees",
        ".workingtrees",
        ".warlock",
        "target",
        "node_modules",
        "vendor",
        "build",
        "dist",
        "out",
    ];
    path.strip_prefix(root)
        .ok()
        .into_iter()
        .flat_map(Path::components)
        .any(|component| {
            let value = component.as_os_str().to_string_lossy();
            defaults.iter().any(|default| *default == value)
        })
}

fn render(context: &Context, repository: &str) -> Result<String> {
    let mut header = JsonMap::new();
    header.insert("adapter_version".into(), Value::String("0.1.0".into()));
    header.insert("language".into(), Value::String("rust".into()));
    header.insert("record".into(), Value::String("lexicon".into()));
    header.insert("repository".into(), Value::String(repository.into()));
    header.insert("schema_version".into(), Value::Number(1.into()));
    let mut values = vec![Value::Object(header)];
    let mut facts: Vec<Value> = context
        .facts
        .nodes
        .values()
        .chain(context.facts.edges.values())
        .chain(context.facts.unresolved.values())
        .cloned()
        .collect();
    facts.sort_by_key(fact_sort_key);
    values.extend(facts);
    values
        .into_iter()
        .map(|value| serde_json::to_string(&value).context("cannot serialize fact"))
        .collect::<Result<Vec<_>>>()
        .map(|lines| format!("{}\n", lines.join("\n")))
}

fn fact_sort_key(value: &Value) -> (u8, String, String, String, String, Vec<String>) {
    let object = value.as_object().expect("fact records are objects");
    let record = object.get("record").and_then(Value::as_str).unwrap_or("");
    match record {
        "node" => (
            0,
            string_field(object, "id"),
            string_field(object, "kind"),
            string_field(object, "path"),
            string_field(object, "qualified_name"),
            Vec::new(),
        ),
        "edge" => (
            1,
            string_field(object, "source"),
            string_field(object, "target"),
            string_field(object, "relation"),
            String::new(),
            span_sort_key(object.get("span")),
        ),
        _ => (
            2,
            string_field(object, "source"),
            string_field(object, "relation"),
            string_field(object, "expression"),
            string_field(object, "reason"),
            span_sort_key(object.get("span")),
        ),
    }
}

fn string_field(object: &JsonMap, name: &str) -> String {
    object
        .get(name)
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string()
}

fn span_sort_key(value: Option<&Value>) -> Vec<String> {
    let Some(span) = value.and_then(Value::as_object) else {
        return vec![String::new(); 5];
    };
    [
        "path",
        "start_line",
        "start_column",
        "end_line",
        "end_column",
    ]
    .iter()
    .map(|key| match span.get(*key) {
        Some(Value::String(value)) => value.clone(),
        Some(value) => value.to_string(),
        None => String::new(),
    })
    .collect()
}

fn object_from_btree(values: BTreeMap<String, Value>) -> Value {
    values.into_iter().collect::<Map<_, _>>().into()
}

fn attrs<const N: usize>(values: [(&str, &str); N]) -> BTreeMap<String, Value> {
    values
        .into_iter()
        .map(|(key, value)| (key.into(), Value::String(value.into())))
        .collect()
}

fn stable_id(language: &str, kind: &str, canonical: &str) -> String {
    let input = format!("lexicon:v1\0{language}\0{kind}\0{canonical}");
    let digest = Sha256::digest(input.as_bytes());
    format!("sha256:{digest:x}")
}

fn content_id(content: &[u8]) -> String {
    let digest = Sha256::digest(content);
    format!("sha256:{digest:x}")
}

fn span_value(span: Span, path: &str) -> Option<Value> {
    let start = span.start();
    let end = span.end();
    let mut value = JsonMap::new();
    value.insert("end_column".into(), Value::Number((end.column + 1).into()));
    value.insert("end_line".into(), Value::Number(end.line.into()));
    value.insert("path".into(), Value::String(path.into()));
    value.insert(
        "start_column".into(),
        Value::Number((start.column + 1).into()),
    );
    value.insert("start_line".into(), Value::Number(start.line.into()));
    Some(Value::Object(value))
}

fn span_start(span: Span) -> (usize, usize) {
    let start = span.start();
    (start.line, start.column + 1)
}

fn span_key(value: &Option<Value>) -> String {
    value
        .as_ref()
        .map(|value| value.to_string())
        .unwrap_or_default()
}

fn relative_path(repo: &Path, path: impl AsRef<Path>) -> String {
    let path = path.as_ref();
    if let Ok(relative) = path.strip_prefix(repo) {
        return normalize_path(relative);
    }
    let repo_key = comparable_path(repo).trim_end_matches('/').to_string();
    let path_key = comparable_path(path);
    path_key
        .strip_prefix(&(repo_key + "/"))
        .map(str::to_string)
        .unwrap_or_else(|| normalize_path(path))
}

fn normalize_path(path: &Path) -> String {
    let value = path.to_string_lossy().replace('\\', "/");
    if value.is_empty() {
        ".".into()
    } else {
        value
    }
}

fn repository_node(repository: &str) -> String {
    stable_id("rust", "repository", repository)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;

    fn fixture() -> PathBuf {
        PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/sample")
    }

    #[test]
    fn emits_declarations_relationships_and_unresolved_macro() {
        let output = generate(&fixture()).expect("generate facts");
        let records: Vec<Value> = output
            .lines()
            .map(|line| serde_json::from_str(line).unwrap())
            .collect();
        assert_eq!(records[0]["language"], "rust");
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "type"
            && record["name"] == "Service"));
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "trait"
            && record["name"] == "Runnable"));
        assert!(records
            .iter()
            .any(|record| record["record"] == "node" && record["kind"] == "import"));
        assert!(records
            .iter()
            .any(|record| record["record"] == "edge" && record["relation"] == "implements"));
        assert!(records.iter().any(
            |record| record["record"] == "unresolved" && record["reason"] == "generated-target"
        ));
        assert!(records
            .iter()
            .any(|record| record["record"] == "edge" && record["relation"] == "imports"));
        assert!(records.iter().any(|record| record["record"] == "node"
            && record["kind"] == "module"
            && record["name"] == "child"));
    }

    #[test]
    fn repeat_runs_are_byte_identical_and_paths_are_relative() {
        let first = generate(&fixture()).expect("first run");
        let second = generate(&fixture()).expect("second run");
        assert_eq!(first, second);
        for record in first.lines().skip(1) {
            let value: Value = serde_json::from_str(record).unwrap();
            if let Some(path) = value.get("path").and_then(Value::as_str) {
                assert!(!path.contains(":\\"));
                assert!(!path.starts_with('/'));
            }
            if let Some(span) = value.get("span") {
                assert!(!span["path"].as_str().unwrap_or("").contains(':'));
            }
        }
    }

    #[test]
    fn header_and_fact_order_are_canonical() {
        let output = generate(&fixture()).expect("generate facts");
        assert_eq!(
            output.lines().next().unwrap(),
            r#"{"adapter_version":"0.1.0","language":"rust","record":"lexicon","repository":"lexicon_fixture","schema_version":1}"#
        );
        let records: Vec<Value> = output
            .lines()
            .skip(1)
            .map(|line| serde_json::from_str(line).unwrap())
            .collect();
        let mut previous = None;
        for record in records {
            let key = fact_sort_key(&record);
            if let Some(previous) = &previous {
                assert!(previous <= &key);
            }
            previous = Some(key);
        }
    }
}
