package lexicon.java;

import com.sun.source.tree.CompilationUnitTree;
import com.sun.source.util.JavacTask;
import com.sun.source.util.Trees;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import javax.lang.model.util.Elements;
import javax.lang.model.util.Types;
import javax.tools.JavaCompiler;
import javax.tools.JavaFileObject;
import javax.tools.StandardJavaFileManager;

final class CompilerBatch {
    private final JavaCompiler compiler;
    private final Path repository;
    private final Emitter emitter;
    private final List<Path> sourceRoots;
    private final Map<Path, List<Path>> groups;

    CompilerBatch(JavaCompiler compiler, Path repository, List<Path> sources, Emitter emitter) {
        this.compiler = compiler;
        this.repository = repository;
        this.emitter = emitter;
        this.groups = groupSources(sources);
        this.sourceRoots = new ArrayList<>(groups.keySet());
        this.sourceRoots.sort(Comparator.comparing(Path::toString));
    }

    void run() throws Exception {
        if (sourceRoots.isEmpty()) {
            return;
        }
        ExecutorService executor = Executors.newFixedThreadPool(workerCount());
        List<Future<?>> tasks = new ArrayList<>();
        try {
            for (Path root : sourceRoots) {
                tasks.add(executor.submit(() -> {
                    analyzeResilient(root, groups.get(root));
                    return null;
                }));
            }
            for (Future<?> task : tasks) {
                await(task);
            }
        } finally {
            executor.shutdownNow();
        }
    }

    private static void await(Future<?> task) throws Exception {
        try {
            task.get();
        } catch (InterruptedException failure) {
            Thread.currentThread().interrupt();
            throw failure;
        } catch (ExecutionException failure) {
            Throwable cause = failure.getCause();
            if (cause instanceof Exception exception) {
                throw exception;
            }
            if (cause instanceof Error error) {
                throw error;
            }
            throw new RuntimeException(cause);
        }
    }

    private int workerCount() {
        int fallback = Math.min(4, Runtime.getRuntime().availableProcessors());
        String configured = System.getenv("LEXICON_JAVA_WORKERS");
        if (configured == null || configured.isBlank()) {
            return Math.max(1, Math.min(fallback, sourceRoots.size()));
        }
        try {
            int workers = Integer.parseInt(configured);
            return Math.max(1, Math.min(Math.min(workers, 32), sourceRoots.size()));
        } catch (NumberFormatException ignored) {
            return Math.max(1, Math.min(fallback, sourceRoots.size()));
        }
    }

    private void analyzeResilient(Path root, List<Path> sources) throws Exception {
        try {
            analyze(root, sources);
        } catch (VirtualMachineError failure) {
            throw failure;
        } catch (Throwable failure) {
            if (sources.size() == 1) {
                emitter.failure(relative(sources.get(0)), failure.getClass().getSimpleName());
                return;
            }
            int midpoint = sources.size() / 2;
            analyzeResilient(root, new ArrayList<>(sources.subList(0, midpoint)));
            analyzeResilient(root, new ArrayList<>(sources.subList(midpoint, sources.size())));
        }
    }

    private void analyze(Path root, List<Path> sources) throws Exception {
        ErrorCollector diagnostics = new ErrorCollector();
        try (StandardJavaFileManager files = compiler.getStandardFileManager(
            diagnostics, null, StandardCharsets.UTF_8
        )) {
            Iterable<? extends JavaFileObject> units = files.getJavaFileObjectsFromPaths(sources);
            JavacTask task = (JavacTask) compiler.getTask(
                null, files, diagnostics, options(), null, units
            );
            List<CompilationUnitTree> parsed = new ArrayList<>();
            task.parse().forEach(parsed::add);
            task.analyze();
            scan(task, diagnostics, parsed);
        }
    }

    private void scan(
        JavacTask task,
        ErrorCollector diagnostics,
        List<CompilationUnitTree> parsed
    ) {
        Trees trees = Trees.instance(task);
        Elements elements = task.getElements();
        Types types = task.getTypes();
        AnalysisContext context = new AnalysisContext(
            repository, trees, elements, types, emitter, diagnostics.errors()
        );
        parsed.sort(Comparator.comparing(context::pathFor));
        for (CompilationUnitTree unit : parsed) {
            new DeclarationIndexer(context).scan(unit, null);
        }
        for (CompilationUnitTree unit : parsed) {
            new SemanticScanner(context).scan(unit, null);
        }
    }

    private List<String> options() {
        return List.of(
            "-proc:none", "-implicit:none", "-Xlint:none", "-Xmaxerrs", "1000000"
        );
    }

    private Map<Path, List<Path>> groupSources(List<Path> sources) {
        Map<Path, List<Path>> result = new LinkedHashMap<>();
        for (Path source : sources) {
            result.computeIfAbsent(sourceRoot(source), ignored -> new ArrayList<>()).add(source);
        }
        for (List<Path> group : result.values()) {
            group.sort(Comparator.comparing(Path::toString));
        }
        return result;
    }

    private Path sourceRoot(Path source) {
        Path current = source.getParent();
        while (current != null && current.startsWith(repository)) {
            if (current.getFileName() != null && current.getFileName().toString().equals("java")) {
                return current;
            }
            current = current.getParent();
        }
        return repository;
    }

    private String relative(Path source) {
        return repository.relativize(source).toString().replace('\\', '/');
    }
}
