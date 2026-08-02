using System.Xml.Linq;

internal static class ProjectFiles
{
    internal static IReadOnlyList<string> Discover(string root)
    {
        var projects = new List<string>();
        var pending = new Stack<string>();
        pending.Push(root);
        while (pending.Count > 0)
        {
            var directory = pending.Pop();
            foreach (var child in Directory.EnumerateDirectories(directory)
                         .OrderByDescending(path => path, StringComparer.Ordinal))
            {
                if (!Discovery.IsIgnoredPath(root, child))
                {
                    pending.Push(child);
                }
            }
            projects.AddRange(Directory.EnumerateFiles(directory, "*.csproj"));
        }
        return projects
            .Select(Path.GetFullPath)
            .Distinct(StringComparer.OrdinalIgnoreCase)
            .OrderBy(path => path, StringComparer.OrdinalIgnoreCase)
            .ToArray();
    }

    internal static string? SelectTargetFramework(IReadOnlyList<string> projects)
    {
        var configured = Environment.GetEnvironmentVariable("LEXICON_CSHARP_TARGET_FRAMEWORK");
        if (!string.IsNullOrWhiteSpace(configured))
        {
            return configured.Trim();
        }

        var counts = new Dictionary<string, int>(StringComparer.OrdinalIgnoreCase);
        foreach (var project in projects)
        {
            try
            {
                var document = XDocument.Load(project, LoadOptions.None);
                foreach (var property in document.Descendants()
                             .Where(element => element.Name.LocalName is "TargetFramework" or "TargetFrameworks"))
                {
                    foreach (var framework in property.Value.Split(';', StringSplitOptions.RemoveEmptyEntries))
                    {
                        var value = framework.Trim();
                        if (value.Length == 0 || value.Contains("$(", StringComparison.Ordinal))
                        {
                            continue;
                        }
                        counts[value] = counts.GetValueOrDefault(value) + 1;
                    }
                }
            }
            catch (Exception error) when (error is IOException or System.Xml.XmlException)
            {
                // MSBuild will report project evaluation diagnostics later.
            }
        }
        return counts
            .OrderByDescending(entry => FrameworkRank(entry.Key))
            .ThenByDescending(entry => entry.Value)
            .ThenBy(entry => entry.Key, StringComparer.OrdinalIgnoreCase)
            .Select(entry => entry.Key)
            .FirstOrDefault();
    }

    private static int FrameworkRank(string framework)
    {
        var value = framework.ToLowerInvariant();
        if (value.StartsWith("netstandard", StringComparison.Ordinal))
        {
            return 100;
        }
        if (value.StartsWith("netcoreapp", StringComparison.Ordinal))
        {
            return 700 + ParsedVersionRank(value[10..]);
        }
        if (!value.StartsWith("net", StringComparison.Ordinal))
        {
            return 0;
        }

        var versionText = new string(value.Skip(3)
            .TakeWhile(character => char.IsDigit(character) || character == '.')
            .ToArray());
        if (!versionText.Contains('.', StringComparison.Ordinal))
        {
            return 300;
        }
        return 1000 + ParsedVersionRank(versionText);
    }

    private static int ParsedVersionRank(string value)
    {
        return Version.TryParse(value, out var version)
            ? version.Major * 100 + version.Minor
            : 1;
    }
}
