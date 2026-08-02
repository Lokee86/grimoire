package lexicon.java;

import com.sun.source.tree.MethodTree;
import com.sun.source.tree.Tree;
import com.sun.source.tree.VariableTree;
import com.sun.source.util.TreePath;
import java.util.ArrayList;
import java.util.List;
import javax.lang.model.element.Element;
import javax.lang.model.element.ElementKind;
import javax.lang.model.element.ExecutableElement;
import javax.lang.model.element.NestingKind;
import javax.lang.model.element.TypeElement;
import javax.lang.model.element.VariableElement;

final class Identities {
    private final AnalysisContext context;

    Identities(AnalysisContext context) {
        this.context = context;
    }

    Ref ref(Element element) {
        if (element == null) {
            return null;
        }
        Ref known = context.knownRefs.get(element);
        if (known != null) {
            return known;
        }
        if (context.trees.getPath(element) == null) {
            return null;
        }
        Ref resolved = switch (element.getKind()) {
            case CLASS, ENUM, RECORD -> typeRef((TypeElement) element, "type");
            case INTERFACE, ANNOTATION_TYPE -> typeRef((TypeElement) element, "interface");
            case METHOD -> executableRef((ExecutableElement) element, false);
            case CONSTRUCTOR -> executableRef((ExecutableElement) element, true);
            case FIELD, ENUM_CONSTANT, RECORD_COMPONENT -> fieldRef((VariableElement) element);
            case PARAMETER -> parameterRef((VariableElement) element);
            default -> null;
        };
        if (resolved != null) {
            context.knownRefs.put(element, resolved);
        }
        return resolved;
    }

    private Ref typeRef(TypeElement element, String kind) {
        if (element.getNestingKind() == NestingKind.LOCAL || element.getNestingKind() == NestingKind.ANONYMOUS) {
            return null;
        }
        String name = element.getQualifiedName().toString();
        if (name.isEmpty()) {
            return null;
        }
        return new Ref(kind, name);
    }

    private Ref executableRef(ExecutableElement element, boolean constructor) {
        TypeElement owner = enclosingType(element);
        if (owner == null || owner.getQualifiedName().isEmpty()
            || owner.getNestingKind() == NestingKind.LOCAL
            || owner.getNestingKind() == NestingKind.ANONYMOUS) {
            return null;
        }
        TreePath path = context.trees.getPath(element);
        if (path == null || context.span(path) == null || !(path.getLeaf() instanceof MethodTree method)) {
            return null;
        }
        List<String> parameterTypes = new ArrayList<>();
        List<? extends VariableTree> parameters = method.getParameters();
        for (int index = 0; index < parameters.size(); index++) {
            String type = normalize(parameters.get(index).getType().toString());
            if (element.isVarArgs() && index == parameters.size() - 1) {
                type = asVarargs(type);
            }
            parameterTypes.add(type);
        }
        String name = constructor ? "<init>" : element.getSimpleName().toString();
        String identity = owner.getQualifiedName() + "." + name + "(" + String.join(",", parameterTypes) + ")";
        return new Ref(constructor ? "constructor" : "method", identity);
    }

    private Ref fieldRef(VariableElement element) {
        TypeElement owner = enclosingType(element);
        if (owner == null) {
            return null;
        }
        return new Ref("field", owner.getQualifiedName() + "." + element.getSimpleName());
    }

    private Ref parameterRef(VariableElement element) {
        if (!(element.getEnclosingElement() instanceof ExecutableElement executable)) {
            return null;
        }
        Ref callable = ref(executable);
        if (callable == null) {
            return null;
        }
        List<? extends VariableElement> parameters = executable.getParameters();
        for (int index = 0; index < parameters.size(); index++) {
            if (parameters.get(index).equals(element)) {
                return new Ref("parameter", callable.identity() + "#parameter:" + index + ":" + element.getSimpleName());
            }
        }
        return null;
    }

    TypeElement enclosingType(Element element) {
        Element current = element;
        while (current != null && !(current instanceof TypeElement)) {
            current = current.getEnclosingElement();
        }
        return (TypeElement) current;
    }

    private static String normalize(String value) {
        return value.replaceAll("\\s+", "");
    }

    private static String asVarargs(String value) {
        if (value.endsWith("[]")) {
            return value.substring(0, value.length() - 2) + "...";
        }
        return value.endsWith("...") ? value : value + "...";
    }
}
