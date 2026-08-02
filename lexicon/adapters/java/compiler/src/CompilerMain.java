package lexicon.java;

import java.io.IOException;
import java.io.PrintWriter;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import javax.tools.JavaCompiler;
import javax.tools.ToolProvider;

public final class CompilerMain {
    private CompilerMain() {}

    public static void main(String[] args) throws Exception {
        Arguments parsed = Arguments.parse(args);
        JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
        if (compiler == null) {
            throw new IllegalStateException("jdk.compiler is unavailable");
        }
        List<Path> sources = readSources(parsed.sourcesFile());
        Emitter emitter = new Emitter(new PrintWriter(System.out, false, StandardCharsets.UTF_8));
        new CompilerBatch(compiler, parsed.repository(), sources, emitter).run();
        emitter.flush();
    }

    private static List<Path> readSources(Path file) throws IOException {
        List<Path> result = new ArrayList<>();
        for (String line : Files.readAllLines(file, StandardCharsets.UTF_8)) {
            String value = line.trim();
            if (!value.isEmpty()) {
                result.add(Path.of(value).toAbsolutePath().normalize());
            }
        }
        result.sort(Comparator.comparing(Path::toString));
        return result;
    }

    private record Arguments(Path repository, Path sourcesFile) {
        private static Arguments parse(String[] args) {
            Path repository = null;
            Path sourcesFile = null;
            for (int index = 0; index < args.length; index++) {
                String argument = args[index];
                if (argument.equals("--repo") && index + 1 < args.length) {
                    repository = Path.of(args[++index]).toAbsolutePath().normalize();
                } else if (argument.equals("--sources-file") && index + 1 < args.length) {
                    sourcesFile = Path.of(args[++index]).toAbsolutePath().normalize();
                } else {
                    throw new IllegalArgumentException("unsupported argument: " + argument);
                }
            }
            if (repository == null || sourcesFile == null) {
                throw new IllegalArgumentException("--repo and --sources-file are required");
            }
            return new Arguments(repository, sourcesFile);
        }
    }
}
