package lexicon.java;

import com.sun.source.tree.ClassTree;
import com.sun.source.tree.MethodTree;
import com.sun.source.tree.VariableTree;
import com.sun.source.util.TreePathScanner;
import javax.lang.model.element.Element;

final class DeclarationIndexer extends TreePathScanner<Void, Void> {
    private final AnalysisContext context;

    DeclarationIndexer(AnalysisContext context) {
        this.context = context;
    }

    @Override
    public Void visitClass(ClassTree node, Void unused) {
        remember();
        return super.visitClass(node, unused);
    }

    @Override
    public Void visitMethod(MethodTree node, Void unused) {
        remember();
        return super.visitMethod(node, unused);
    }

    @Override
    public Void visitVariable(VariableTree node, Void unused) {
        remember();
        return super.visitVariable(node, unused);
    }

    private void remember() {
        Element element = context.trees.getElement(getCurrentPath());
        Ref ref = context.identities.ref(element);
        if (element != null && ref != null) {
            context.knownRefs.put(element, ref);
        }
    }
}
