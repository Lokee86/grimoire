internal static class Program
{
    private const string AdapterVersion = "0.2.0";
    private const string Language = "csharp";

    private static async Task<int> Main(string[] args)
    {
        try
        {
            var options = CommandLineOptions.Parse(args);
            var repositoryRoot = ResolveRepositoryRoot(options.Repository);
            var repository = ResolveRepositoryIdentity(repositoryRoot);
            var header = CreateHeader(repository, options);
            var facts = await Analysis.RunAsync(repositoryRoot, repository, options.ProjectLoading);
            WriteOutput(options.Output, facts.EmitJsonl(header));
            return 0;
        }
        catch (ArgumentException error)
        {
            Console.Error.WriteLine($"error: {error.Message}");
            return 2;
        }
        catch (Exception error) when (error is IOException or InvalidOperationException)
        {
            Console.Error.WriteLine($"error: {error.Message}");
            return 1;
        }
    }

    private static HeaderRecord CreateHeader(string repository, CommandLineOptions options)
    {
        if (!options.HasIncrementalScope)
        {
            return new HeaderRecord
            {
                AdapterVersion = AdapterVersion,
                Language = Language,
                Mode = "full",
                Record = "lexicon",
                Repository = repository,
                SchemaVersion = 1,
            };
        }

        var changedFiles = NormalizePaths(options.ChangedFiles);
        var removedFiles = NormalizePaths(options.RemovedFiles);
        if (changedFiles.Intersect(removedFiles, StringComparer.Ordinal).Any())
        {
            throw new ArgumentException("--changed-file and --removed-file must be disjoint");
        }

        return new HeaderRecord
        {
            AdapterVersion = AdapterVersion,
            ChangedFiles = changedFiles,
            Language = Language,
            Mode = "incremental",
            Record = "lexicon",
            RemovedFiles = removedFiles,
            Repository = repository,
            SchemaVersion = 1,
            SharedComplete = true,
        };
    }

    private static string ResolveRepositoryRoot(string repositoryPath)
    {
        var fullPath = Path.GetFullPath(repositoryPath);
        if (!Directory.Exists(fullPath))
        {
            throw new ArgumentException($"repository does not exist: {repositoryPath}");
        }
        return fullPath;
    }

    private static string ResolveRepositoryIdentity(string repositoryRoot)
    {
        var name = new DirectoryInfo(repositoryRoot).Name;
        if (string.IsNullOrEmpty(name))
        {
            throw new ArgumentException($"repository has no directory name: {repositoryRoot}");
        }
        return name;
    }

    private static IReadOnlyList<string> NormalizePaths(IEnumerable<string> paths)
    {
        return paths
            .Select(NormalizePath)
            .Distinct(StringComparer.Ordinal)
            .OrderBy(path => path, StringComparer.Ordinal)
            .ToArray();
    }

    private static string NormalizePath(string path)
    {
        var normalized = path.Replace('\\', '/');
        if (string.IsNullOrWhiteSpace(normalized) || Path.IsPathRooted(path) || normalized.StartsWith('/'))
        {
            throw new ArgumentException($"file path must be repository-relative: {path}");
        }

        var segments = normalized.Split('/');
        if (segments.Any(segment => segment is "" or "." or ".."))
        {
            throw new ArgumentException($"file path is not normalized: {path}");
        }

        return normalized;
    }

    private static void WriteOutput(string outputPath, string content)
    {
        if (outputPath == "-")
        {
            Console.Out.Write(content);
            return;
        }

        var fullPath = Path.GetFullPath(outputPath);
        var parent = Path.GetDirectoryName(fullPath);
        if (!string.IsNullOrEmpty(parent))
        {
            Directory.CreateDirectory(parent);
        }

        File.WriteAllText(fullPath, content, new System.Text.UTF8Encoding(encoderShouldEmitUTF8Identifier: false));
    }
}

internal enum ProjectLoadingMode
{
    Auto,
    MsBuild,
    Files,
}

internal sealed class CommandLineOptions
{
    private CommandLineOptions(
        string repository,
        string output,
        ProjectLoadingMode projectLoading,
        IReadOnlyList<string> changedFiles,
        IReadOnlyList<string> removedFiles,
        bool hasIncrementalScope)
    {
        Repository = repository;
        Output = output;
        ProjectLoading = projectLoading;
        ChangedFiles = changedFiles;
        RemovedFiles = removedFiles;
        HasIncrementalScope = hasIncrementalScope;
    }

    internal string Repository { get; }
    internal string Output { get; }
    internal ProjectLoadingMode ProjectLoading { get; }
    internal IReadOnlyList<string> ChangedFiles { get; }
    internal IReadOnlyList<string> RemovedFiles { get; }
    internal bool HasIncrementalScope { get; }

    internal static CommandLineOptions Parse(string[] args)
    {
        string? repository = null;
        string? output = null;
        var projectLoading = ProjectLoadingMode.Auto;
        var changedFiles = new List<string>();
        var removedFiles = new List<string>();
        var hasIncrementalScope = false;

        for (var index = 0; index < args.Length; index++)
        {
            var argument = args[index];
            switch (argument)
            {
                case "--repo":
                    repository = ReadValue(args, ref index, "--repo");
                    break;
                case "--output":
                    output = ReadValue(args, ref index, "--output");
                    break;
                case "--project-loading":
                    projectLoading = ParseProjectLoading(ReadValue(args, ref index, "--project-loading"));
                    break;
                case "--changed-file":
                    changedFiles.Add(ReadValue(args, ref index, "--changed-file"));
                    hasIncrementalScope = true;
                    break;
                case "--removed-file":
                    removedFiles.Add(ReadValue(args, ref index, "--removed-file"));
                    hasIncrementalScope = true;
                    break;
                default:
                    throw new ArgumentException($"unknown argument: {argument}");
            }
        }

        if (string.IsNullOrWhiteSpace(repository))
        {
            throw new ArgumentException("--repo is required");
        }

        if (string.IsNullOrWhiteSpace(output))
        {
            throw new ArgumentException("--output is required");
        }

        return new CommandLineOptions(
            repository,
            output,
            projectLoading,
            changedFiles,
            removedFiles,
            hasIncrementalScope);
    }

    private static ProjectLoadingMode ParseProjectLoading(string value)
    {
        return value.ToLowerInvariant() switch
        {
            "auto" => ProjectLoadingMode.Auto,
            "msbuild" => ProjectLoadingMode.MsBuild,
            "files" => ProjectLoadingMode.Files,
            _ => throw new ArgumentException("--project-loading must be auto, msbuild, or files"),
        };
    }

    private static string ReadValue(string[] args, ref int index, string option)
    {
        if (++index >= args.Length || args[index].StartsWith("--", StringComparison.Ordinal))
        {
            throw new ArgumentException($"{option} requires a value");
        }

        return args[index];
    }
}
