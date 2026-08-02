package demo.app;

import demo.api.Contract;
import demo.api.ImportedAnnotation;
import demo.base.ImportedBase;
import demo.wild.*;

@ImportedAnnotation
class Positive<T> extends ImportedBase<T> implements Contract {
    @ImportedAnnotation
    Positive() {
    }

    @ImportedAnnotation
    String value;

    @ImportedAnnotation
    void apply(@ImportedAnnotation String input) {
    }

    class WildChild extends WildBase {
    }

    @WildAnnotation
    class WildAnnotated {
    }
}

class SameChild extends SameBase {
}

class QualifiedChild extends demo.base.QualifiedBase {
}

record AnnotatedRecord(@ImportedAnnotation String value) {
}
