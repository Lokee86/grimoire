package lexicon.java;

import com.sun.source.tree.AssignmentTree;
import com.sun.source.tree.ClassTree;
import com.sun.source.tree.CompoundAssignmentTree;
import com.sun.source.tree.IdentifierTree;
import com.sun.source.tree.MemberReferenceTree;
import com.sun.source.tree.LambdaExpressionTree;
import com.sun.source.tree.MemberSelectTree;
import com.sun.source.tree.MethodInvocationTree;
import com.sun.source.tree.MethodTree;
import com.sun.source.tree.NewClassTree;
import com.sun.source.tree.Tree;
import com.sun.source.tree.UnaryTree;
import com.sun.source.tree.VariableTree;
import com.sun.source.util.TreePath;
import com.sun.source.util.TreePathScanner;
import java.util.EnumSet;
import javax.lang.model.element.Element;
import javax.lang.model.element.ElementKind;
import javax.lang.model.element.ExecutableElement;
import javax.lang.model.element.TypeElement;
import javax.lang.model.element.VariableElement;

final class SemanticScanner extends TreePathScanner<Void, Void> {
    private static final EnumSet<Tree.Kind> MUTATING_UNARY = EnumSet.of(
        Tree.Kind.PREFIX_INCREMENT,
        Tree.Kind.PREFIX_DECREMENT,
        Tree.Kind.POSTFIX_INCREMENT,
        Tree.Kind.POSTFIX_DECREMENT
    );

    private final AnalysisContext context;
    private Ref currentSource;
    private TypeElement currentType;

    SemanticScanner(AnalysisContext context) {
        this.context = context;
    }

    @Override
    public Void visitClass(ClassTree node, Void unused) {
        Element element = context.trees.getElement(getCurrentPath());
        Ref previousSource = currentSource;
        TypeElement previousType = currentType;
        if (element instanceof TypeElement type) {
            Ref typeRef = context.identities.ref(type);
            if (typeRef == null) {
                return null;
            }
            currentSource = typeRef;
            currentType = type;
        }
        Void result = super.visitClass(node, unused);
        currentSource = previousSource;
        currentType = previousType;
        return result;
    }

    @Override
    public Void visitMethod(MethodTree node, Void unused) {
        Element element = context.trees.getElement(getCurrentPath());
        Ref previousSource = currentSource;
        if (element instanceof ExecutableElement method) {
            Ref methodRef = context.identities.ref(method);
            if (methodRef != null) {
                currentSource = methodRef;
                context.emitTypeReferences(methodRef, method.getReturnType(), context.span(getCurrentPath()));
                emitOverrides(method, methodRef);
            }
        }
        Void result = super.visitMethod(node, unused);
        currentSource = previousSource;
        return result;
    }

    @Override
    public Void visitVariable(VariableTree node, Void unused) {
        Element element = context.trees.getElement(getCurrentPath());
        Ref previousSource = currentSource;
        Ref declaration = context.identities.ref(element);
        if (declaration != null) {
            context.emitTypeReferences(declaration, element.asType(), context.span(getCurrentPath()));
            if (element.getKind() == ElementKind.FIELD
                || element.getKind() == ElementKind.ENUM_CONSTANT
                || element.getKind() == ElementKind.RECORD_COMPONENT) {
                currentSource = declaration;
            }
        } else if (currentSource != null && element != null) {
            context.emitTypeReferences(currentSource, element.asType(), context.span(getCurrentPath()));
        }
        Void result = super.visitVariable(node, unused);
        currentSource = previousSource;
        return result;
    }

    @Override
    public Void visitLambdaExpression(LambdaExpressionTree node, Void unused) {
        return null;
    }

    @Override
    public Void visitMethodInvocation(MethodInvocationTree node, Void unused) {
        emitTarget("calls", context.trees.getElement(getCurrentPath()));
        return super.visitMethodInvocation(node, unused);
    }

    @Override
    public Void visitNewClass(NewClassTree node, Void unused) {
        emitTarget("calls", context.trees.getElement(getCurrentPath()));
        return super.visitNewClass(node, unused);
    }

    @Override
    public Void visitMemberReference(MemberReferenceTree node, Void unused) {
        emitTarget("references", context.trees.getElement(getCurrentPath()));
        return super.visitMemberReference(node, unused);
    }

    @Override
    public Void visitIdentifier(IdentifierTree node, Void unused) {
        emitVariableAccess(context.trees.getElement(getCurrentPath()));
        return super.visitIdentifier(node, unused);
    }

    @Override
    public Void visitMemberSelect(MemberSelectTree node, Void unused) {
        emitVariableAccess(context.trees.getElement(getCurrentPath()));
        return super.visitMemberSelect(node, unused);
    }

    private void emitTarget(String relation, Element element) {
        Span span = context.span(getCurrentPath());
        if (span == null || context.hasError(getCurrentPath())
            || (span.startLine() == span.endLine() && span.startColumn() == span.endColumn())) {
            return;
        }
        Ref target = context.identities.ref(element);
        if (currentSource != null && target != null) {
            context.emitter.edge(currentSource, target, relation, span, "javac");
        }
    }

    private void emitVariableAccess(Element element) {
        if (!(element instanceof VariableElement variable) || currentSource == null) {
            return;
        }
        Access access = accessKind();
        if (localVariable(variable.getKind())) {
            Span span = context.span(getCurrentPath());
            if (access.read()) {
                context.emitter.suppression(currentSource, "reads", span, "javac", "local-variable");
            }
            if (access.write()) {
                context.emitter.suppression(currentSource, "writes", span, "javac", "local-variable");
            }
            return;
        }
        if (!modeledVariable(variable.getKind())) {
            return;
        }
        Ref target = context.identities.ref(variable);
        if (target == null) {
            return;
        }
        Span span = context.span(getCurrentPath());
        if (access.read()) {
            context.emitter.edge(currentSource, target, "reads", span, "javac");
        }
        if (access.write()) {
            context.emitter.edge(currentSource, target, "writes", span, "javac");
        }
    }

    private Access accessKind() {
        TreePath parentPath = getCurrentPath().getParentPath();
        if (parentPath == null) {
            return Access.READ;
        }
        Tree current = getCurrentPath().getLeaf();
        Tree parent = parentPath.getLeaf();
        if (parent instanceof AssignmentTree assignment && assignment.getVariable() == current) {
            return Access.WRITE;
        }
        if (parent instanceof CompoundAssignmentTree assignment && assignment.getVariable() == current) {
            return Access.READ_WRITE;
        }
        if (parent instanceof UnaryTree unary && unary.getExpression() == current && MUTATING_UNARY.contains(unary.getKind())) {
            return Access.READ_WRITE;
        }
        return Access.READ;
    }

    private void emitOverrides(ExecutableElement method, Ref methodRef) {
        if (currentType == null || method.getKind() != ElementKind.METHOD) {
            return;
        }
        for (Element member : context.elements.getAllMembers(currentType)) {
            if (!(member instanceof ExecutableElement candidate) || candidate.equals(method)) {
                continue;
            }
            Ref target = context.identities.ref(candidate);
            if (target != null && context.elements.overrides(method, candidate, currentType)) {
                context.emitter.edge(methodRef, target, "overrides", context.span(getCurrentPath()), "javac");
            }
        }
    }

    private static boolean localVariable(ElementKind kind) {
        return kind == ElementKind.LOCAL_VARIABLE
            || kind == ElementKind.RESOURCE_VARIABLE
            || kind == ElementKind.EXCEPTION_PARAMETER
            || kind == ElementKind.BINDING_VARIABLE;
    }

    private static boolean modeledVariable(ElementKind kind) {
        return kind == ElementKind.FIELD
            || kind == ElementKind.ENUM_CONSTANT
            || kind == ElementKind.RECORD_COMPONENT
            || kind == ElementKind.PARAMETER;
    }

    private record Access(boolean read, boolean write) {
        static final Access READ = new Access(true, false);
        static final Access WRITE = new Access(false, true);
        static final Access READ_WRITE = new Access(true, true);
    }
}
