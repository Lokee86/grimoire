package demo.nested;

class Outer {
    @NestedAnnotation
    class Child extends NestedBase {
    }

    class NestedBase {
    }

    @interface NestedAnnotation {
    }
}
