package lexicon.java;

import java.io.PrintWriter;
import java.util.Map;
import java.util.concurrent.ConcurrentSkipListMap;

final class Emitter {
    private final PrintWriter output;
    private final Map<String, String> records = new ConcurrentSkipListMap<>();

    Emitter(PrintWriter output) {
        this.output = output;
    }

    void edge(Ref source, Ref target, String relation, Span span, String engine) {
        if (source == null || target == null || span == null) {
            return;
        }
        String json = "{"
            + field("record", "edge") + ","
            + field("source_kind", source.kind()) + ","
            + field("source_identity", source.identity()) + ","
            + field("target_kind", target.kind()) + ","
            + field("target_identity", target.identity()) + ","
            + field("relation", relation) + ","
            + field("path", span.path()) + ","
            + number("start_line", span.startLine()) + ","
            + number("start_column", span.startColumn()) + ","
            + number("end_line", span.endLine()) + ","
            + number("end_column", span.endColumn()) + ","
            + field("engine", engine)
            + "}";
        records.put(json, json);
    }

    void failure(String path, String reason) {
        String json = "{"
            + field("record", "failure") + ","
            + field("path", path) + ","
            + field("reason", reason)
            + "}";
        records.put(json, json);
    }

    void suppression(Ref source, String relation, Span span, String engine, String reason) {
        if (source == null || span == null) {
            return;
        }
        String json = "{"
            + field("record", "suppression") + ","
            + field("source_kind", source.kind()) + ","
            + field("source_identity", source.identity()) + ","
            + field("relation", relation) + ","
            + field("path", span.path()) + ","
            + number("start_line", span.startLine()) + ","
            + number("start_column", span.startColumn()) + ","
            + number("end_line", span.endLine()) + ","
            + number("end_column", span.endColumn()) + ","
            + field("engine", engine) + ","
            + field("reason", reason)
            + "}";
        records.put(json, json);
    }

    void flush() {
        for (String record : records.values()) {
            output.println(record);
        }
        output.flush();
    }

    private static String field(String name, String value) {
        return quote(name) + ":" + quote(value);
    }

    private static String number(String name, int value) {
        return quote(name) + ":" + value;
    }

    private static String quote(String value) {
        StringBuilder result = new StringBuilder(value.length() + 2);
        result.append('"');
        for (int index = 0; index < value.length(); index++) {
            char current = value.charAt(index);
            switch (current) {
                case '"' -> result.append("\\\"");
                case '\\' -> result.append("\\\\");
                case '\b' -> result.append("\\b");
                case '\f' -> result.append("\\f");
                case '\n' -> result.append("\\n");
                case '\r' -> result.append("\\r");
                case '\t' -> result.append("\\t");
                default -> {
                    if (current < 0x20) {
                        result.append(String.format("\\u%04x", (int) current));
                    } else {
                        result.append(current);
                    }
                }
            }
        }
        return result.append('"').toString();
    }
}
