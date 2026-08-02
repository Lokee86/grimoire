package lexicon.java;

import com.sun.source.tree.CompilationUnitTree;
import com.sun.source.tree.Tree;
import com.sun.source.util.SourcePositions;
import com.sun.source.util.TreePath;
import com.sun.source.util.Trees;
import java.net.URI;
import java.nio.file.Path;
import java.util.Collections;
import java.util.HashMap;
import java.util.IdentityHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import javax.lang.model.element.Element;
import javax.lang.model.type.ArrayType;
import javax.lang.model.type.DeclaredType;
import javax.lang.model.type.TypeKind;
import javax.lang.model.type.TypeMirror;
import javax.lang.model.type.TypeVariable;
import javax.lang.model.type.WildcardType;
import javax.lang.model.util.Elements;
import javax.lang.model.util.Types;

final class AnalysisContext {
    final Path repository;
    final Trees trees;
    final Elements elements;
    final Types types;
    final Emitter emitter;
    final Identities identities;
    final Map<Element, Ref> knownRefs = new HashMap<>();
    private final SourcePositions positions;
    private final Set<String> errorLines = new LinkedHashSet<>();

    AnalysisContext(Path repository, Trees trees, Elements elements, Types types, Emitter emitter, List<ErrorLocation> diagnostics) {
        this.repository = repository;
        this.trees = trees;
        this.elements = elements;
        this.types = types;
        this.emitter = emitter;
        this.positions = trees.getSourcePositions();
        this.identities = new Identities(this);
        for (ErrorLocation diagnostic : diagnostics) {
            errorLines.add(pathFor(diagnostic.source()) + ":" + diagnostic.line());
        }
    }

    String pathFor(CompilationUnitTree unit) {
        return pathFor(unit.getSourceFile().toUri());
    }

    private String pathFor(URI uri) {
        try {
            Path absolute = Path.of(uri).toAbsolutePath().normalize();
            return repository.relativize(absolute).toString().replace('\\', '/');
        } catch (RuntimeException error) {
            return uri.toString();
        }
    }

    boolean hasError(TreePath path) {
        Span value = span(path);
        if (value == null) {
            return false;
        }
        for (int line = value.startLine(); line <= value.endLine(); line++) {
            if (errorLines.contains(value.path() + ":" + line)) {
                return true;
            }
        }
        return false;
    }

    Span span(TreePath path) {
        if (path == null) {
            return null;
        }
        CompilationUnitTree unit = path.getCompilationUnit();
        Tree leaf = path.getLeaf();
        long start = positions.getStartPosition(unit, leaf);
        long end = positions.getEndPosition(unit, leaf);
        if (start < 0) {
            return null;
        }
        if (end < start) {
            end = start;
        }
        return new Span(
            pathFor(unit),
            Math.toIntExact(unit.getLineMap().getLineNumber(start)),
            Math.toIntExact(unit.getLineMap().getColumnNumber(start)),
            Math.toIntExact(unit.getLineMap().getLineNumber(end)),
            Math.toIntExact(unit.getLineMap().getColumnNumber(end))
        );
    }

    void emitTypeReferences(Ref source, TypeMirror mirror, Span span) {
        if (source == null || mirror == null) {
            return;
        }
        Set<Ref> targets = new LinkedHashSet<>();
        Set<TypeMirror> visited = Collections.newSetFromMap(new IdentityHashMap<>());
        collectTypes(mirror, targets, visited);
        for (Ref target : targets) {
            emitter.edge(source, target, "references", span, "javac");
        }
    }

    private void collectTypes(TypeMirror mirror, Set<Ref> result, Set<TypeMirror> visited) {
        if (mirror == null || !visited.add(mirror)) {
            return;
        }
        TypeKind kind = mirror.getKind();
        if (kind == TypeKind.ARRAY) {
            collectTypes(((ArrayType) mirror).getComponentType(), result, visited);
            return;
        }
        if (kind == TypeKind.DECLARED) {
            DeclaredType declared = (DeclaredType) mirror;
            Element element = declared.asElement();
            Ref target = identities.ref(element);
            if (target != null) {
                result.add(target);
            }
            for (TypeMirror argument : declared.getTypeArguments()) {
                collectTypes(argument, result, visited);
            }
            return;
        }
        if (kind == TypeKind.TYPEVAR) {
            TypeVariable variable = (TypeVariable) mirror;
            collectTypes(variable.getUpperBound(), result, visited);
            collectTypes(variable.getLowerBound(), result, visited);
            return;
        }
        if (kind == TypeKind.WILDCARD) {
            WildcardType wildcard = (WildcardType) mirror;
            collectTypes(wildcard.getExtendsBound(), result, visited);
            collectTypes(wildcard.getSuperBound(), result, visited);
        }
    }
}

record Span(String path, int startLine, int startColumn, int endLine, int endColumn) {}
record Ref(String kind, String identity) {}
