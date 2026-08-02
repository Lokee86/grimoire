package lexicon.java;

import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import javax.tools.Diagnostic;
import javax.tools.DiagnosticListener;
import javax.tools.JavaFileObject;

final class ErrorCollector implements DiagnosticListener<JavaFileObject> {
    private final List<ErrorLocation> errors = new ArrayList<>();

    @Override
    public void report(Diagnostic<? extends JavaFileObject> diagnostic) {
        if (diagnostic.getKind() != Diagnostic.Kind.ERROR
            || diagnostic.getSource() == null
            || diagnostic.getLineNumber() < 1) {
            return;
        }
        errors.add(new ErrorLocation(diagnostic.getSource().toUri(), diagnostic.getLineNumber()));
    }

    List<ErrorLocation> errors() {
        return errors;
    }
}

record ErrorLocation(URI source, long line) {}
