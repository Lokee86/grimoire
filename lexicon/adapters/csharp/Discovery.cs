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
    int FallbackFileCount,
    string Mode,
    int ProjectCount,
    string Root,
    string? TargetFramework,
    IReadOnlyList<string> WorkspaceDiagnostics);

internal static class Discovery
{
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
            catch (InvalidOperationException) when (loadingMode == ProjectLoadingMode.Auto)
            {
                // Project evaluation is optional in auto mode; file loading remains the safe fallback.
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

    private static RepositoryModel LoadFiles(string root)
    {
        var provisional = DiscoverFiles(root)
            .Select(path => LoadSyntax(root, path))
            .OrderBy(document => document.RelativePath, StringComparer.Ordinal)
            .ToArray();
        var compilation = CSharpCompilation.Create(
            assemblyName: "Lexicon.CSharp.Analysis",
            syntaxTrees: provisional.Select(document => document.SyntaxTree),
            references: TrustedPlatformReferences(),
            options: new CSharpCompilationOptions(
                OutputKind.DynamicallyLinkedLibrary,
                allowUnsafe: true,
                nullableContextOptions: NullableContextOptions.Enable));
        var documents = provisional
            .Select(document => document with { Compilation = compilation })
            .ToArray();
        return new RepositoryModel(documents, 0, "files", 0, root, null, Array.Empty<string>());
    }

    private static RepositoryModel CompleteProjectModel(string root, RepositoryModel projectModel)
    {
        var fallback = LoadFiles(root);
        var documents = projectModel.Documents.ToDictionary(
            document => document.RelativePath,
            StringComparer.Ordinal);
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
            FallbackFileCount = added,
            Mode = added == 0 ? "msbuild" : "msbuild+files",
        };
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
