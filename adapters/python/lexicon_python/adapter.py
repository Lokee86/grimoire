"""A deterministic, standard-library-only Python Lexicon adapter."""

from __future__ import annotations

import ast
import hashlib
import io
import json
import os
import sys
import tokenize
from dataclasses import dataclass, field
from pathlib import Path, PurePosixPath
from typing import Any, Iterable, TextIO

from . import __version__

LANGUAGE = "python"
SCHEMA_VERSION = 1

# These names cover the root contract's defaults plus the conventional Python,
# JavaScript, Rust, and general build/vendor directories that should not become
# repository facts.
EXCLUDED_DIRECTORIES = frozenset(
    {
        ".git",
        ".worktrees",
        ".workingtrees",
        ".warlock",
        ".bundle",
        ".eggs",
        ".mypy_cache",
        ".nox",
        ".pytest_cache",
        ".ruff_cache",
        ".tox",
        ".venv",
        "__pycache__",
        "build",
        "dist",
        "env",
        "node_modules",
        "site-packages",
        "target",
        "vendor",
        "venv",
    }
)
BUILTINS = frozenset(
    {
        "__import__",
        "abs",
        "all",
        "any",
        "bool",
        "bytes",
        "callable",
        "classmethod",
        "dict",
        "dir",
        "enumerate",
        "filter",
        "float",
        "getattr",
        "hasattr",
        "hash",
        "int",
        "isinstance",
        "issubclass",
        "iter",
        "len",
        "list",
        "map",
        "max",
        "min",
        "next",
        "object",
        "open",
        "print",
        "property",
        "range",
        "repr",
        "reversed",
        "set",
        "setattr",
        "staticmethod",
        "str",
        "sum",
        "super",
        "tuple",
        "type",
        "vars",
        "zip",
    }
)


def digest(value: str | bytes) -> str:
    payload = value.encode("utf-8") if isinstance(value, str) else value
    return f"sha256:{hashlib.sha256(payload).hexdigest()}"


def node_id(kind: str, identity: str) -> str:
    return digest(f"lexicon:v1\0{LANGUAGE}\0{kind}\0{identity}")


def content_id(data: bytes) -> str:
    return digest(data)


def expression_text(node: ast.AST, source: str) -> str:
    segment = ast.get_source_segment(source, node)
    if segment:
        return segment.strip()
    try:
        return ast.unparse(node)
    except (AttributeError, ValueError):
        return type(node).__name__


def _column(line: str, offset: int) -> int:
    """Convert AST's UTF-8 byte offset to a one-based character column."""
    prefix = line.encode("utf-8")[:offset]
    return len(prefix.decode("utf-8", errors="ignore")) + 1


def span(node: ast.AST, path: str, lines: list[str]) -> dict[str, Any] | None:
    lineno = getattr(node, "lineno", None)
    col_offset = getattr(node, "col_offset", None)
    end_lineno = getattr(node, "end_lineno", None)
    end_col_offset = getattr(node, "end_col_offset", None)
    if None in (lineno, col_offset, end_lineno, end_col_offset):
        return None
    if not (1 <= lineno <= len(lines) and 1 <= end_lineno <= len(lines)):
        return None
    return {
        "end_column": _column(lines[end_lineno - 1], end_col_offset),
        "end_line": end_lineno,
        "path": path,
        "start_column": _column(lines[lineno - 1], col_offset),
        "start_line": lineno,
    }


def _span_key(record: dict[str, Any]) -> tuple[Any, ...]:
    value = record.get("span") or {}
    return (
        value.get("path", ""),
        value.get("start_line", 0),
        value.get("start_column", 0),
        value.get("end_line", 0),
        value.get("end_column", 0),
    )


def _record_sort_key(record: dict[str, Any]) -> tuple[Any, ...]:
    kind = record["record"]
    if kind == "node":
        return (0, record["id"], record["kind"], record["path"], record["qualified_name"])
    if kind == "edge":
        return (1, record["source"], record["target"], record["relation"], *_span_key(record))
    return (
        2,
        record["source"],
        record["relation"],
        record["expression"],
        record["reason"],
        *_span_key(record),
    )


def _posix_relative(root: Path, path: Path) -> str:
    relative = path.relative_to(root).as_posix()
    return relative or "."


def _module_name(root_name: str, relative_path: str) -> str:
    parts = list(PurePosixPath(relative_path).parts)
    leaf = parts.pop()
    stem = leaf[:-3] if leaf.endswith(".py") else leaf
    if stem == "__init__":
        if parts:
            return ".".join(parts)
        return root_name
    parts.append(stem)
    return ".".join(parts)


def _name_from_dotted(value: str) -> str:
    return value.rsplit(".", 1)[-1]


def _dotted(node: ast.AST) -> str | None:
    if isinstance(node, ast.Name):
        return node.id
    if isinstance(node, ast.Attribute):
        parent = _dotted(node.value)
        return f"{parent}.{node.attr}" if parent else None
    return None


@dataclass
class ImportInfo:
    module_name: str
    owner_id: str
    node_id: str
    statement: ast.AST
    expression: str
    binding: str | None
    target_module: str | None = None
    target_name: str | None = None
    relative_level: int = 0
    star: bool = False
    is_package: bool = False


@dataclass
class InheritanceInfo:
    source_id: str
    module_name: str
    class_qname: str
    base: ast.expr
    source: str
    path: str
    lines: list[str]


@dataclass
class CallInfo:
    module_name: str
    owner_id: str
    class_qname: str | None
    expression_node: ast.Call
    callee: ast.AST


@dataclass
class FileContext:
    root: Path
    path: Path
    relative_path: str
    module_name: str
    source: str
    lines: list[str]
    tree: ast.AST | None
    file_id: str
    module_id: str


@dataclass
class Facts:
    repository: str
    nodes: dict[str, dict[str, Any]] = field(default_factory=dict)
    edges: dict[str, dict[str, Any]] = field(default_factory=dict)
    unresolved: dict[str, dict[str, Any]] = field(default_factory=dict)
    modules: dict[str, str] = field(default_factory=dict)
    symbols: dict[str, str] = field(default_factory=dict)
    symbol_kinds: dict[str, str] = field(default_factory=dict)
    imports: list[ImportInfo] = field(default_factory=list)
    inheritances: list[InheritanceInfo] = field(default_factory=list)
    calls: list[CallInfo] = field(default_factory=list)
    module_bindings: dict[tuple[str, str], tuple[str | None, str]] = field(default_factory=dict)

    def add_node(
        self,
        kind: str,
        name: str,
        path: str,
        qualified_name: str,
        *,
        identity: str | None = None,
        record_span: dict[str, Any] | None = None,
        attributes: dict[str, Any] | None = None,
        file_content_id: str | None = None,
    ) -> str:
        identifier = node_id(kind, identity if identity is not None else qualified_name)
        record: dict[str, Any] = {
            "record": "node",
            "id": identifier,
            "kind": kind,
            "name": name,
            "path": path,
            "qualified_name": qualified_name,
        }
        if file_content_id is not None:
            record["content_id"] = file_content_id
        if attributes:
            record["attributes"] = attributes
        if record_span is not None:
            record["span"] = record_span
        self.nodes[identifier] = record
        return identifier

    def add_edge(
        self,
        source: str,
        target: str,
        relation: str,
        *,
        record_span: dict[str, Any] | None = None,
        attributes: dict[str, Any] | None = None,
    ) -> None:
        record: dict[str, Any] = {
            "record": "edge",
            "relation": relation,
            "source": source,
            "target": target,
        }
        if attributes:
            record["attributes"] = attributes
        if record_span is not None:
            record["span"] = record_span
        key = json.dumps(record, sort_keys=True, separators=(",", ":"))
        self.edges[key] = record

    def add_unresolved(
        self,
        source: str,
        relation: str,
        expression: str,
        reason: str,
        *,
        record_span: dict[str, Any] | None = None,
        candidate_name: str | None = None,
    ) -> None:
        record: dict[str, Any] = {
            "record": "unresolved",
            "relation": relation,
            "source": source,
            "expression": expression,
            "reason": reason,
        }
        if candidate_name:
            record["candidate_name"] = candidate_name
        if record_span is not None:
            record["span"] = record_span
        key = json.dumps(record, sort_keys=True, separators=(",", ":"))
        self.unresolved[key] = record


class DeclarationVisitor(ast.NodeVisitor):
    def __init__(self, facts: Facts, context: FileContext) -> None:
        self.facts = facts
        self.context = context
        self.class_stack: list[tuple[str, str]] = []
        self.function_stack: list[tuple[str, str]] = []
        self.owner_stack: list[str] = [context.module_id]
        self.import_index = 0

    @property
    def owner_id(self) -> str:
        return self.owner_stack[-1]

    @property
    def class_qname(self) -> str | None:
        return self.class_stack[-1][0] if self.class_stack else None

    @property
    def current_qname(self) -> str:
        if self.function_stack:
            return self.function_stack[-1][0]
        if self.class_stack:
            return self.class_stack[-1][0]
        return self.context.module_name

    def _attributes(self, node: ast.AST) -> dict[str, Any]:
        decorators = sorted(expression_text(item, self.context.source) for item in getattr(node, "decorator_list", []))
        attributes: dict[str, Any] = {}
        if decorators:
            attributes["decorators"] = decorators
        if isinstance(node, (ast.AsyncFunctionDef,)):
            attributes["async"] = True
        return attributes

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        nested_names = [name for _, name in self.class_stack + self.function_stack]
        nested_names.append(node.name)
        qname = f"{self.context.module_name}.{'.'.join(nested_names)}"
        identifier = self.facts.add_node(
            "type",
            node.name,
            self.context.relative_path,
            qname,
            record_span=span(node, self.context.relative_path, self.context.lines),
            attributes={
                **self._attributes(node),
                **({"bases": sorted(expression_text(base, self.context.source) for base in node.bases)} if node.bases else {}),
            },
        )
        self.facts.symbols[qname] = identifier
        self.facts.symbol_kinds[qname] = "type"
        self.facts.add_edge(
            self.owner_id,
            identifier,
            "defines",
            record_span=span(node, self.context.relative_path, self.context.lines),
        )
        self.class_stack.append((qname, node.name))
        self.owner_stack.append(identifier)
        for base in node.bases:
            self._inheritance(identifier, qname, base)
        for statement in node.body:
            self.visit(statement)
        self.owner_stack.pop()
        self.class_stack.pop()

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self._visit_function(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self._visit_function(node)

    def _visit_function(self, node: ast.FunctionDef | ast.AsyncFunctionDef) -> None:
        prefix = [self.context.module_name]
        prefix.extend(name for _, name in self.class_stack)
        prefix.extend(name for _, name in self.function_stack)
        prefix.append(node.name)
        qname = ".".join(prefix)
        kind = "method" if self.class_stack and not self.function_stack else "function"
        identifier = self.facts.add_node(
            kind,
            node.name,
            self.context.relative_path,
            qname,
            record_span=span(node, self.context.relative_path, self.context.lines),
            attributes=self._attributes(node),
        )
        self.facts.symbols[qname] = identifier
        self.facts.symbol_kinds[qname] = kind
        self.facts.add_edge(
            self.owner_id,
            identifier,
            "defines",
            record_span=span(node, self.context.relative_path, self.context.lines),
        )
        self.function_stack.append((qname, node.name))
        self.owner_stack.append(identifier)
        for statement in node.body:
            self.visit(statement)
        self.owner_stack.pop()
        self.function_stack.pop()

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            self._add_import(node, alias.name, alias.asname, alias.asname or alias.name.split(".")[0], 0, None, False)

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        module = node.module or ""
        for alias in node.names:
            binding = None if alias.name == "*" else (alias.asname or alias.name)
            self._add_import(node, module, alias.asname, binding, node.level, alias.name, alias.name == "*")

    def _add_import(
        self,
        statement: ast.AST,
        target_module: str,
        asname: str | None,
        binding: str | None,
        relative_level: int,
        target_name: str | None,
        star: bool,
    ) -> None:
        self.import_index += 1
        statement_span = span(statement, self.context.relative_path, self.context.lines)
        expression = expression_text(statement, self.context.source)
        if target_name and target_name != "*":
            expression = f"{expression} [{target_name}]"
        qualified_name = (
            f"{self.context.module_name}::import:{statement_span.get('start_line', 0) if statement_span else 0}:"
            f"{self.import_index}:{binding or target_module}"
        )
        identifier = self.facts.add_node(
            "import",
            binding or target_name or target_module,
            self.context.relative_path,
            qualified_name,
            identity=qualified_name,
            record_span=statement_span,
            attributes={"expression": expression},
        )
        self.facts.add_edge(self.owner_id, identifier, "defines", record_span=statement_span)
        info = ImportInfo(
            module_name=self.context.module_name,
            owner_id=self.owner_id,
            node_id=identifier,
            statement=statement,
            expression=expression,
            binding=binding,
            target_module=target_module,
            relative_level=relative_level,
            target_name=target_name,
            star=star,
            is_package=self.context.relative_path.endswith("/__init__.py") or self.context.relative_path == "__init__.py",
        )
        self.facts.imports.append(info)
        if self.owner_id == self.context.module_id and binding:
            self.facts.module_bindings[(self.context.module_name, binding)] = (None, "unresolved")

    def visit_Call(self, node: ast.Call) -> None:
        self.facts.calls.append(
            CallInfo(
                module_name=self.context.module_name,
                owner_id=self.owner_id,
                class_qname=self.class_qname,
                expression_node=node,
                callee=node.func,
            )
        )
        self.generic_visit(node)

    def _inheritance(self, source_id: str, class_qname: str, base: ast.expr) -> None:
        self.facts.inheritances.append(
            InheritanceInfo(
                source_id=source_id,
                module_name=self.context.module_name,
                class_qname=class_qname,
                base=base,
                source=self.context.source,
                path=self.context.relative_path,
                lines=self.context.lines,
            )
        )


def resolve_relative_module(info: ImportInfo) -> str:
    if info.relative_level == 0:
        return info.target_module or ""
    current_parts = info.module_name.split(".") if info.module_name else []
    # A module's package is itself only for __init__ modules. The root package
    # convention is represented by the repository name in _module_name.
    package_parts = current_parts if info.is_package else current_parts[:-1]
    remove = max(info.relative_level - 1, 0)
    if remove:
        package_parts = package_parts[:-remove] if remove <= len(package_parts) else []
    suffix = (info.target_module or "").split(".") if info.target_module else []
    return ".".join(package_parts + suffix)


def resolve_reference(
    facts: Facts,
    module_name: str,
    class_qname: str | None,
    reference: str | None,
) -> tuple[str | None, str]:
    if not reference:
        return None, "unsupported-form"
    if reference in BUILTINS:
        return None, "builtin-target"
    parts = reference.split(".")
    if class_qname and parts[0] in {"self", "cls"} and len(parts) == 2:
        candidate = f"{class_qname}.{parts[1]}"
        if candidate in facts.symbols:
            return facts.symbols[candidate], ""
    binding = facts.module_bindings.get((module_name, parts[0]))
    if binding and binding[0]:
        base_qname = next((qname for qname, value in facts.symbols.items() if value == binding[0]), None)
        if base_qname:
            candidate = ".".join([base_qname, *parts[1:]])
            if candidate in facts.symbols:
                return facts.symbols[candidate], ""
        if len(parts) == 1:
            return binding[0], ""
    candidates = [
        ".".join([module_name, *parts]),
        reference,
    ]
    if class_qname and len(parts) == 1:
        candidates.insert(0, f"{class_qname}.{reference}")
    for candidate in candidates:
        if candidate in facts.symbols:
            return facts.symbols[candidate], ""
        if candidate in facts.modules:
            return facts.modules[candidate], ""
    if binding and binding[1] == "external":
        return None, "external-target"
    return None, "missing-target"


def resolve_import(facts: Facts, info: ImportInfo) -> tuple[str | None, str]:
    if info.star:
        return None, "unsupported-form"
    module_name = resolve_relative_module(info)
    if info.target_name is None:
        candidates = [module_name]
        # `import package.module` binds the complete imported module.
        target = next((facts.modules[candidate] for candidate in candidates if candidate in facts.modules), None)
        return (target, "") if target else (None, "external-target")
    symbol = f"{module_name}.{info.target_name}" if module_name else info.target_name
    if symbol in facts.symbols:
        return facts.symbols[symbol], ""
    if symbol in facts.modules:
        return facts.modules[symbol], ""
    if module_name in facts.modules:
        # A known internal module with an unknown member is a missing target,
        # not evidence that another symbol should be guessed.
        return None, "missing-target"
    return None, "external-target"


def _call_resolution(facts: Facts, call: CallInfo) -> tuple[str | None, str]:
    reference = _dotted(call.callee)
    if reference is None:
        return None, "dynamic-target"
    if reference in {"importlib.import_module", "import_module", "__import__"}:
        return None, "dynamic-target"
    return resolve_reference(facts, call.module_name, call.class_qname, reference)


def _scan_paths(root: Path) -> tuple[list[Path], list[Path]]:
    directories: list[Path] = [root]
    python_files: list[Path] = []
    for current, dirnames, filenames in os.walk(root, topdown=True, followlinks=False):
        dirnames[:] = sorted(name for name in dirnames if name not in EXCLUDED_DIRECTORIES)
        current_path = Path(current)
        directories.extend(current_path / name for name in dirnames)
        python_files.extend(
            current_path / name for name in sorted(filenames) if name.lower().endswith(".py")
        )
    return sorted(set(directories)), sorted(set(python_files))


def _read_source(path: Path) -> tuple[str, bytes]:
    data = path.read_bytes()
    with tokenize.open(path) as handle:
        return handle.read(), data


def build_facts(repo: Path) -> list[dict[str, Any]]:
    root = repo.expanduser().resolve()
    if not root.is_dir():
        raise NotADirectoryError(f"repository is not a directory: {repo}")
    repository = root.name
    facts = Facts(repository=repository)
    repository_id = facts.add_node("repository", repository, ".", repository, identity=repository)

    directories, python_files = _scan_paths(root)
    directory_ids: dict[str, str] = {".": repository_id}
    for directory in directories:
        relative = _posix_relative(root, directory)
        if relative == ".":
            continue
        name = PurePosixPath(relative).name
        directory_ids[relative] = facts.add_node("directory", name, relative, relative, identity=relative)
    for relative, identifier in sorted(directory_ids.items()):
        if relative == ".":
            continue
        parent = PurePosixPath(relative).parent.as_posix()
        facts.add_edge(directory_ids.get(parent, repository_id), identifier, "contains")

    contexts: list[FileContext] = []
    for path in python_files:
        relative = _posix_relative(root, path)
        data = path.read_bytes()
        file_id = facts.add_node(
            "file",
            path.name,
            relative,
            relative,
            identity=relative,
            file_content_id=content_id(data),
        )
        parent = PurePosixPath(relative).parent.as_posix()
        facts.add_edge(directory_ids.get(parent, repository_id), file_id, "contains")
        module_name = _module_name(repository, relative)
        module_id = facts.add_node("module", _name_from_dotted(module_name), relative, module_name, identity=module_name)
        facts.modules[module_name] = module_id
        facts.add_edge(file_id, module_id, "contains")
        try:
            source, _ = _read_source(path)
            tree: ast.AST | None = ast.parse(source, filename=relative)
        except (OSError, SyntaxError, UnicodeDecodeError, tokenize.TokenError) as error:
            source = ""
            tree = None
            facts.add_unresolved(
                module_id,
                "parses",
                relative,
                "unsupported-form",
                candidate_name=type(error).__name__,
            )
        contexts.append(
            FileContext(
                root=root,
                path=path,
                relative_path=relative,
                module_name=module_name,
                source=source,
                lines=source.splitlines(),
                tree=tree,
                file_id=file_id,
                module_id=module_id,
            )
        )

    for context in contexts:
        if context.tree is not None:
            DeclarationVisitor(facts, context).visit(context.tree)

    # Resolve imports only after all modules and declarations are known.
    for info in facts.imports:
        target_id, reason = resolve_import(facts, info)
        if info.owner_id == facts.modules.get(info.module_name) and info.binding:
            facts.module_bindings[(info.module_name, info.binding)] = (
                target_id,
                "external" if target_id is None and reason == "external-target" else reason,
            )
        import_span = _import_span_for_node(facts, info.node_id)
        if target_id:
            facts.add_edge(info.owner_id, target_id, "imports", record_span=import_span)
        else:
            facts.add_unresolved(
                info.owner_id,
                "imports",
                info.expression,
                reason,
                record_span=import_span,
                candidate_name=resolve_relative_module(info) or info.target_module,
            )

    for inheritance in facts.inheritances:
        target_name = _dotted(inheritance.base)
        target_id, reason = resolve_reference(
            facts,
            inheritance.module_name,
            inheritance.class_qname,
            target_name,
        )
        base_span = span(inheritance.base, inheritance.path, inheritance.lines)
        if target_id:
            facts.add_edge(inheritance.source_id, target_id, "extends", record_span=base_span)
        else:
            facts.add_unresolved(
                inheritance.source_id,
                "extends",
                expression_text(inheritance.base, inheritance.source),
                reason,
                record_span=base_span,
                candidate_name=target_name,
            )

    for call in facts.calls:
        target_id, reason = _call_resolution(facts, call)
        call_span = _call_span(call, contexts)
        if target_id:
            facts.add_edge(call.owner_id, target_id, "calls", record_span=call_span)
        else:
            facts.add_unresolved(
                call.owner_id,
                "calls",
                expression_text(call.expression_node, _source_for_call(call, contexts)),
                reason,
                record_span=call_span,
                candidate_name=_dotted(call.callee),
            )

    header = {
        "record": "lexicon",
        "schema_version": SCHEMA_VERSION,
        "adapter_version": __version__,
        "language": LANGUAGE,
        "repository": repository,
    }
    records = [header, *sorted(facts.nodes.values(), key=_record_sort_key)]
    records.extend(sorted(facts.edges.values(), key=_record_sort_key))
    records.extend(sorted(facts.unresolved.values(), key=_record_sort_key))
    return records


def _node_path(facts: Facts, identifier: str) -> str:
    return facts.nodes.get(identifier, {}).get("path", "")


def _import_span_for_node(facts: Facts, identifier: str) -> dict[str, Any] | None:
    return facts.nodes.get(identifier, {}).get("span")


def _source_for_call(call: CallInfo, contexts: list[FileContext]) -> str:
    for context in contexts:
        if context.module_name == call.module_name:
            return context.source
    return ""


def _call_span(call: CallInfo, contexts: list[FileContext]) -> dict[str, Any] | None:
    for context in contexts:
        if context.module_name == call.module_name:
            return span(call.expression_node, context.relative_path, context.lines)
    return None


def write_facts(repo: Path, output: Path) -> None:
    records = build_facts(repo)
    lines = [json.dumps(record, ensure_ascii=False, sort_keys=True, separators=(",", ":")) for record in records]
    if str(output) == "-":
        sys.stdout.write("\n".join(lines) + "\n")
        return
    destination = output.expanduser()
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_text("\n".join(lines) + "\n", encoding="utf-8", newline="\n")
