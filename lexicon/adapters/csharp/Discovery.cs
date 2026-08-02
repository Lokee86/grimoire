using System.Security.Cryptography;
using System.Text;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.Text;

internal sealed record SourceDocument(
    string AbsolutePath,
    byte[] Content,
    CSharpCompilation Compilation,
    string RelativePath,
    SyntaxTree SyntaxTree);

internal sealed record RepositoryModel(
    IReadOnlyList<SourceDocument> Documents,
    int CompilationCount,
    int FallbackFileCount,
    string Mode,
    int ProjectCount,
    string Root,
    string? TargetFramework,
    IReadOnlyList<string> WorkspaceDiagnostics);

internal static class Discovery
{
    private const int MaxFallbackCompilationFiles = 512;

    private static readonly HashSet<string> IgnoredDirectories = new(StringComparer.OrdinalIgnoreCase)
    {
        ".git", ".lexicon", ".worktrees", ".workingtrees", ".vs", ".idea",
        "bin", "obj", "packages", "node_modules", "vendor", "generated",
        "artifacts", "build", "dist", "target", "coverage", "tmp", "temp",
    };

    internal static async Task<RepositoryModel> LoadAsync(
        string repositoryRoot,
        ProjectLoadingMode loadingMode,
        CancellationToken cancellationToken = default)
    {
        var root = Path.GetFullPath(repositoryRoot);
        if (loadingMode != ProjectLoadingMode.Files)
        {
            try
            {
                var projectModel = await MsBuildDiscovery.TryLoadAsync(root, cancellationToken);
                if (projectModel is not null)
                {
                    return CompleteProjectModel(root, projectModel);
                }
                if (loadingMode == ProjectLoadingMode.MsBuild)
                {
                    throw new InvalidOperationException(
                        "MSBuild project loading was requested but no project graph could be loaded");
                }
            }
            catch (Exception error) when (
                loadingMode == ProjectLoadingMode.Auto &&
                error is not OperationCanceledException)
            {
                // Project evaluation is optional in auto mode. Preserve a deterministic reason for
                // the fallback without allowing SDK/MSBuild compatibility failures to abort analysis.
                return LoadFiles(root, new[]
                {
                    $"MSBuild project loading failed: {error.GetType().FullName}",
                });
            }
        }

        return LoadFiles(root);
    }

    internal static bool IsIgnoredPath(string root, string path)
    {
        var relative = Path.GetRelativePath(root, path);
        return relative.Split(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar)
            .Any(IgnoredDirectories.Contains);
    }

    private static RepositoryModel LoadFiles(
        string root,
        IReadOnlyList<string>? workspaceDiagnostics = null,
        IReadOnlySet<string>? excludedRelativePaths = null)
    {
        var projectDirectories = ProjectFiles.Discover(root)
            .Select(Path.GetDirectoryName)
            .Where(directory => !string.IsNullOrWhiteSpace(directory))
            .Select(directory => Path.GetFullPath(directory!))
            .ToHashSet(StringComparer.OrdinalIgnoreCase);
        var references = TrustedPlatformReferences();
        var options = new CSharpCompilationOptions(
            OutputKind.DynamicallyLinkedLibrary,
            allowUnsafe: true,
            nullableContextOptions: NullableContextOptions.Enable);
        var documents = new List<SourceDocument>();
        var compilationCount = 0;

        var groups = DiscoverFiles(root)
            .Select(path => LoadSyntax(root, path))
            .Where(document => excludedRelativePaths is null ||
                !excludedRelativePaths.Contains(document.RelativePath))
            .GroupBy(
                document => FallbackCompilationGroup(root, document.AbsolutePath, projectDirectories),
                StringComparer.OrdinalIgnoreCase)
            .OrderBy(group => group.Key, StringComparer.OrdinalIgnoreCase);
        foreach (var group in groups)
        {
            var ordered = group
                .OrderBy(document => document.RelativePath, StringComparer.Ordinal)
                .ToArray();
            for (var offset = 0; offset < ordered.Length; offset += MaxFallbackCompilationFiles)
            {
                var shard = offset / MaxFallbackCompilationFiles;
                var provisional = ordered
                    .Skip(offset)
                    .Take(MaxFallbackCompilationFiles)
                    .ToArray();
                var compilation = CSharpCompilation.Create(
                    assemblyName: FallbackAssemblyName(group.Key, shard),
                    syntaxTrees: provisional.Select(document => document.SyntaxTree),
                    references: references,
                    options: options);
                documents.AddRange(provisional.Select(document => document with { Compilation = compilation }));
                compilationCount++;
            }
        }

        return new RepositoryModel(
            documents.OrderBy(document => document.RelativePath, StringComparer.Ordinal).ToArray(),
            compilationCount,
            0,
            "files",
            0,
            root,
            null,
            workspaceDiagnostics ?? Array.Empty<string>());
    }

    private static RepositoryModel CompleteProjectModel(string root, RepositoryModel projectModel)
    {
        var documents = projectModel.Documents.ToDictionary(
            document => document.RelativePath,
            StringComparer.Ordinal);
        var fallback = LoadFiles(root, excludedRelativePaths: documents.Keys.ToHashSet(StringComparer.Ordinal));
        var added = 0;
        foreach (var document in fallback.Documents)
        {
            if (documents.TryAdd(document.RelativePath, document))
            {
                added++;
            }
        }
        return projectModel with
        {
            Documents = documents.Values.OrderBy(document => document.RelativePath, StringComparer.Ordinal).ToArray(),
            CompilationCount = projectModel.CompilationCount + fallback.CompilationCount,
            FallbackFileCount = added,
            Mode = added == 0 ? "msbuild" : "msbuild+files",
        };
    }

    private static string FallbackCompilationGroup(
        string root,
        string absolutePath,
        IReadOnlySet<string> projectDirectories)
    {
        var directory = Path.GetDirectoryName(absolutePath);
        while (!string.IsNullOrWhiteSpace(directory) && IsInside(root, directory))
        {
            if (projectDirectories.Contains(directory))
            {
                return Facts.NormalizePath(Path.GetRelativePath(root, directory));
            }
            var parent = Directory.GetParent(directory)?.FullName;
            if (string.Equals(parent, directory, StringComparison.OrdinalIgnoreCase))
            {
                break;
            }
            directory = parent;
        }

        var relativePath = Facts.NormalizePath(Path.GetRelativePath(root, absolutePath));
        var separator = relativePath.IndexOf('/');
        return separator < 0 ? "." : relativePath[..separator];
    }

    private static string FallbackAssemblyName(string group, int shard)
    {
        var identity = $"{group}\n{shard}";
        var hash = SHA256.HashData(Encoding.UTF8.GetBytes(identity));
        return $"Lexicon.CSharp.Analysis.{Convert.ToHexString(hash.AsSpan(0, 8))}";
    }

    private static bool IsInside(string root, string path)
    {
        var relative = Path.GetRelativePath(root, path);
        return relative != ".." &&
            !relative.StartsWith($"..{Path.DirectorySeparatorChar}", StringComparison.Ordinal);
    }

    private static IEnumerable<string> DiscoverFiles(string root)
    {
        var pending = new Stack<string>();
        pending.Push(root);
        while (pending.Count > 0)
        {
            var directory = pending.Pop();
            foreach (var child in Directory.EnumerateDirectories(directory)
                         .OrderByDescending(path => path, StringComparer.Ordinal))
            {
                if (!IgnoredDirectories.Contains(Path.GetFileName(child)))
                {
                    pending.Push(child);
                }
            }

            foreach (var file in Directory.EnumerateFiles(directory, "*.cs")
                         .OrderBy(path => path, StringComparer.Ordinal))
            {
                yield return file;
            }
        }
    }

    private static SourceDocument LoadSyntax(string root, string absolutePath)
    {
        var content = File.ReadAllBytes(absolutePath);
        var relativePath = Facts.NormalizePath(Path.GetRelativePath(root, absolutePath));
        var text = SourceText.From(content, content.Length, System.Text.Encoding.UTF8, canBeEmbedded: false);
        var tree = CSharpSyntaxTree.ParseText(
            text,
            new CSharpParseOptions(LanguageVersion.Preview, DocumentationMode.Parse),
            relativePath);
        return new SourceDocument(absolutePath, content, null!, relativePath, tree);
    }

    private static IReadOnlyList<MetadataReference> TrustedPlatformReferences()
    {
        var value = AppContext.GetData("TRUSTED_PLATFORM_ASSEMBLIES") as string;
        if (string.IsNullOrWhiteSpace(value))
        {
            throw new InvalidOperationException("TRUSTED_PLATFORM_ASSEMBLIES is unavailable");
        }

        return value.Split(Path.PathSeparator, StringSplitOptions.RemoveEmptyEntries)
            .Distinct(StringComparer.OrdinalIgnoreCase)
            .OrderBy(path => path, StringComparer.OrdinalIgnoreCase)
            .Select(path => MetadataReference.CreateFromFile(path))
            .ToArray();
    }
}
