package demo.receivers;

class JsonWriterLike {
    void value(String value) {
    }

    void value(boolean value) {
    }

    void value(Boolean value) {
    }

    void value(char value) {
    }

    void value(float value) {
    }

    void value(double value) {
    }

    void value(long value) {
    }

    void value(Number value) {
    }

    void value(ReceiverTarget value) {
    }
}

class PairTarget {
    void pair(int left, long right) {
    }

    void pair(long left, int right) {
    }
}

class OverloadCalls {
    void nullLiteral(JsonWriterLike writer) {
        writer.value(null);
    }

    void booleanLiteral(JsonWriterLike writer) {
        writer.value(true);
    }

    void stringLiteral(JsonWriterLike writer) {
        writer.value("value");
    }

    void charLiteral(JsonWriterLike writer) {
        writer.value('v');
    }

    void integralLiteral(JsonWriterLike writer) {
        writer.value(1);
    }

    void floatingLiteral(JsonWriterLike writer) {
        writer.value(1.0);
    }

    void floatLiteral(JsonWriterLike writer) {
        writer.value(1.0f);
    }

    void typedString(JsonWriterLike writer, String value) {
        writer.value(value);
    }

    void typedBoolean(JsonWriterLike writer, Boolean value) {
        writer.value(value);
    }

    void typedRepositoryType(JsonWriterLike writer, ReceiverTarget value) {
        writer.value(value);
    }

    void typedLocal(JsonWriterLike writer) {
        long value = 1L;
        writer.value(value);
    }

    void unknownExpression(JsonWriterLike writer) {
        writer.value(dynamic());
    }

    void ambiguousArguments(PairTarget target, int left, int right) {
        target.pair(left, right);
    }

    Object dynamic() {
        return null;
    }
}
