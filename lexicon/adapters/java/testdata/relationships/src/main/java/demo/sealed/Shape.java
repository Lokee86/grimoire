package demo.sealed;

sealed interface Shape permits Circle, missing.ExternalShape {
}

final class Circle implements Shape {
}
