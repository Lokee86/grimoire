using Microsoft.Build.Locator;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.MSBuild;

internal static class MsBuildDiscovery
{
    internal static async Task<RepositoryModel?> TryLoadAsync(
        string root,
        CancellationToken cancellationToken)
    {
        var projectFiles = ProjectFiles.Discover(root);
        if (projectFiles.Count == 0 || !TryRegisterMsBuild())
        {
            return null;
        }

        var diagnostics = new List<string>();
        var properties = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase)
        {
            ["AlwaysCompileMarkupFilesInSeparateDomain"] = "false",
            ["BuildProjectReferences"] = "false",
            ["DesignTimeBuild"] = "true",
            ["ProvideCommandLineArgs"] = "true",
            ["SkipCompilerExecution"] = "true",
        };
        var targetFramework = ProjectFiles.SelectTargetFramework(projectFiles);
        if (!string.IsNullOrWhiteSpace(targetFramework))
        {
            properties["TargetFramework"] = targetFramework;
        }

        using var workspace = MSBuildWorkspace.Create(properties);
        workspace.SkipUnrecognizedProjects = true;
        workspace.WorkspaceFailed += (_, eventArgs) => diagnostics.Add(eventArgs.Diagnostic.Message);

        foreach (var projectPath in projectFiles)
        {
            if (workspace.CurrentSolution.Projects.Any(project =>
                    string.Equals(project.FilePath, projectPath, StringComparison.OrdinalIgnoreCase)))
            {
                continue;
            }
            try
            {
                await workspace.OpenProjectAsync(projectPath, cancellationToken: cancellationToken);
            }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException)
            {
                diagnostics.Add($"{Facts.NormalizePath(Path.GetRelativePath(root, projectPath))}: {error.Message}");
            }
        }

        var documents = new Dictionary<string, SourceDocument>(StringComparer.Ordinal);
        var loadedProjects = 0;
        foreach (var project in workspace.CurrentSolution.Projects
                     .Where(project => project.Language == LanguageNames.CSharp)
                     .OrderBy(project => project.FilePath, StringComparer.OrdinalIgnoreCase)
                     .ThenBy(project => project.Name, StringComparer.Ordinal))
        {
            if (await project.GetCompilationAsync(cancellationToken) is not CSharpCompilation compilation)
            {
                diagnostics.Add($"{project.Name}: C# compilation was unavailable");
                continue;
            }
            loadedProjects++;
            foreach (var document in project.Documents.OrderBy(document => document.FilePath, StringComparer.OrdinalIgnoreCase))
            {
                var absolutePath = document.FilePath;
                if (string.IsNullOrWhiteSpace(absolutePath) ||
                    !File.Exists(absolutePath) ||
                    !IsInside(root, absolutePath) ||
                    Discovery.IsIgnoredPath(root, absolutePath))
                {
                    continue;
                }
                var tree = await document.GetSyntaxTreeAsync(cancellationToken);
                if (tree is null)
                {
                    continue;
                }
                var relativePath = Facts.NormalizePath(Path.GetRelativePath(root, absolutePath));
                var candidate = new SourceDocument(
                    absolutePath,
                    File.ReadAllBytes(absolutePath),
                    compilation,
                    relativePath,
                    tree);
                documents.TryAdd(relativePath, candidate);
            }
        }

        if (documents.Count == 0)
        {
            var detail = diagnostics.Count == 0
                ? "no workspace diagnostics were reported"
                : string.Join(" | ", diagnostics.Take(8));
            throw new InvalidOperationException($"MSBuild loaded no repository documents: {detail}");
        }
        var normalizedDiagnostics = diagnostics
            .Select(message => NormalizeDiagnostic(root, message))
            .OrderBy(message => message, StringComparer.Ordinal)
            .Distinct(StringComparer.Ordinal)
            .ToArray();
        return new RepositoryModel(
            documents.Values.OrderBy(document => document.RelativePath, StringComparer.Ordinal).ToArray(),
            0,
            "msbuild",
            loadedProjects,
            root,
            targetFramework,
            normalizedDiagnostics);
    }

    private static bool TryRegisterMsBuild()
    {
        if (MSBuildLocator.IsRegistered)
        {
            return true;
        }

        foreach (var root in DotnetRoots())
        {
            var sdkRoot = Path.Combine(root, "sdk");
            if (!Directory.Exists(sdkRoot))
            {
                continue;
            }
            var sdk = Directory.EnumerateDirectories(sdkRoot)
                .Where(path => File.Exists(Path.Combine(path, "MSBuild.dll")))
                .OrderByDescending(path => SdkVersion(path))
                .ThenByDescending(path => Path.GetFileName(path), StringComparer.OrdinalIgnoreCase)
                .FirstOrDefault();
            if (sdk is not null)
            {
                MSBuildLocator.RegisterMSBuildPath(sdk);
                return true;
            }
        }

        var instance = MSBuildLocator.QueryVisualStudioInstances()
            .OrderByDescending(candidate => candidate.Version)
            .FirstOrDefault();
        if (instance is null)
        {
            return false;
        }
        MSBuildLocator.RegisterInstance(instance);
        return true;
    }

    private static Version SdkVersion(string path)
    {
        return Version.TryParse(Path.GetFileName(path), out var version)
            ? version
            : new Version(0, 0);
    }

    private static string NormalizeDiagnostic(string root, string message)
    {
        var normalizedRoot = Path.GetFullPath(root)
            .TrimEnd(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar);
        var result = message
            .Replace(normalizedRoot + Path.DirectorySeparatorChar, string.Empty, StringComparison.OrdinalIgnoreCase)
            .Replace(normalizedRoot + Path.AltDirectorySeparatorChar, string.Empty, StringComparison.OrdinalIgnoreCase)
            .Replace(normalizedRoot, ".", StringComparison.OrdinalIgnoreCase);
        return Facts.NormalizePath(result);
    }

    private static IEnumerable<string> DotnetRoots()
    {
        var configured = Environment.GetEnvironmentVariable("DOTNET_ROOT");
        if (!string.IsNullOrWhiteSpace(configured))
        {
            yield return configured;
        }
        var executable = Environment.GetEnvironmentVariable("LEXICON_DOTNET");
        if (!string.IsNullOrWhiteSpace(executable))
        {
            var parent = Path.GetDirectoryName(executable);
            if (!string.IsNullOrWhiteSpace(parent))
            {
                yield return parent;
            }
        }
        yield return Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFiles), "dotnet");
        yield return Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".dotnet");
    }

    private static bool IsInside(string root, string path)
    {
        var relative = Path.GetRelativePath(root, path);
        return relative != ".." && !relative.StartsWith($"..{Path.DirectorySeparatorChar}", StringComparison.Ordinal);
    }
}
