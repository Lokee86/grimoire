using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp.Syntax;

internal sealed partial class Analysis
{
    private static readonly SymbolDisplayFormat CanonicalFormat = new(
        globalNamespaceStyle: SymbolDisplayGlobalNamespaceStyle.Omitted,
        typeQualificationStyle: SymbolDisplayTypeQualificationStyle.NameAndContainingTypesAndNamespaces,
        genericsOptions: SymbolDisplayGenericsOptions.IncludeTypeParameters,
        memberOptions: SymbolDisplayMemberOptions.IncludeContainingType |
                       SymbolDisplayMemberOptions.IncludeParameters |
                       SymbolDisplayMemberOptions.IncludeType |
                       SymbolDisplayMemberOptions.IncludeExplicitInterface,
        parameterOptions: SymbolDisplayParameterOptions.IncludeType |
                          SymbolDisplayParameterOptions.IncludeName |
                          SymbolDisplayParameterOptions.IncludeParamsRefOut |
                          SymbolDisplayParameterOptions.IncludeDefaultValue,
        miscellaneousOptions: SymbolDisplayMiscellaneousOptions.UseSpecialTypes |
                              SymbolDisplayMiscellaneousOptions.EscapeKeywordIdentifiers |
                              SymbolDisplayMiscellaneousOptions.IncludeNullableReferenceTypeModifier);

    private readonly Dictionary<ISymbol, string> symbolIds = new(SymbolEqualityComparer.Default);
    private readonly Dictionary<ISymbol, string> qualifiedNames = new(SymbolEqualityComparer.Default);

    private void EmitDeclarations()
    {
        foreach (var document in model.Documents)
        {
            var semantic = document.Compilation.GetSemanticModel(document.SyntaxTree, ignoreAccessibility: true);
            foreach (var node in document.SyntaxTree.GetRoot().DescendantNodesAndSelf())
            {
                if (!IsDeclaration(node))
                {
                    continue;
                }

                var symbol = semantic.GetDeclaredSymbol(node);
                if (symbol is not null && ShouldEmit(symbol))
                {
                    EnsureSymbol(symbol, document.RelativePath);
                }
            }
        }
    }

    private static bool IsDeclaration(SyntaxNode node)
    {
        return node is BaseNamespaceDeclarationSyntax or BaseTypeDeclarationSyntax or DelegateDeclarationSyntax or
            BaseMethodDeclarationSyntax or LocalFunctionStatementSyntax or PropertyDeclarationSyntax or
            IndexerDeclarationSyntax or EventDeclarationSyntax or VariableDeclaratorSyntax or
            ParameterSyntax or AccessorDeclarationSyntax or EnumMemberDeclarationSyntax;
    }

    private static bool ShouldEmit(ISymbol symbol)
    {
        return symbol is INamespaceSymbol or INamedTypeSymbol or IMethodSymbol or IPropertySymbol or
            IFieldSymbol or IEventSymbol or IParameterSymbol or ILocalSymbol;
    }

    private string EnsureSymbol(ISymbol input, string fallbackPath)
    {
        var symbol = NormalizeSymbol(input);
        if (symbol is INamespaceSymbol { IsGlobalNamespace: true })
        {
            return repositoryId;
        }
        if (symbolIds.TryGetValue(symbol, out var known))
        {
            return known;
        }

        var location = SourceLocation(symbol);
        var owner = location?.SourceTree is null ? null : PathFor(location.SourceTree);
        var external = owner is null;
        var path = owner ?? "external";
        var qualifiedName = QualifiedName(symbol);
        var name = string.IsNullOrWhiteSpace(symbol.Name) ? qualifiedName : symbol.Name;
        var attributes = new SortedDictionary<string, object?>(StringComparer.Ordinal)
        {
            ["accessibility"] = symbol.DeclaredAccessibility.ToString().ToLowerInvariant(),
            ["external"] = external,
            ["static"] = symbol.IsStatic,
            ["symbol_kind"] = symbol.Kind.ToString().ToLowerInvariant(),
        };
        var id = facts.AddNode(
            NodeKind(symbol),
            name,
            path,
            qualifiedName,
            identity: SymbolIdentity(symbol, qualifiedName, location, owner),
            span: location is null ? null : Span(path, location),
            attributes: attributes,
            owner: owner);
        symbolIds[symbol] = id;

        if (!external)
        {
            var parent = ParentSymbol(symbol);
            if (parent is not null)
            {
                AddEdge(EnsureSymbol(parent, owner!), id, "contains", owner!);
            }
            else if (location?.SourceTree is not null && fileIds.TryGetValue(location.SourceTree, out var fileId))
            {
                AddEdge(fileId, id, "defines", owner!);
            }
        }
        return id;
    }

    private static ISymbol NormalizeSymbol(ISymbol symbol)
    {
        return symbol switch
        {
            IMethodSymbol method => (method.ReducedFrom ?? method).OriginalDefinition,
            INamedTypeSymbol type => type.OriginalDefinition,
            _ => symbol.OriginalDefinition,
        };
    }

    private static ISymbol? ParentSymbol(ISymbol symbol)
    {
        var parent = symbol.ContainingSymbol;
        return parent switch
        {
            null or IAssemblySymbol or IModuleSymbol => null,
            INamespaceSymbol { IsGlobalNamespace: true } => null,
            _ => parent,
        };
    }

    private static Location? SourceLocation(ISymbol symbol)
    {
        return symbol.Locations
            .Where(location => location.IsInSource && location.SourceTree is not null)
            .OrderBy(location => Facts.NormalizePath(location.SourceTree!.FilePath), StringComparer.Ordinal)
            .ThenBy(location => location.SourceSpan.Start)
            .FirstOrDefault();
    }

    private string QualifiedName(ISymbol symbol)
    {
        if (qualifiedNames.TryGetValue(symbol, out var known))
        {
            return known;
        }
        var value = symbol.ToDisplayString(CanonicalFormat);
        var qualifiedName = string.IsNullOrWhiteSpace(value) ? symbol.MetadataName : value;
        qualifiedNames[symbol] = qualifiedName;
        return qualifiedName;
    }

    private string SymbolIdentity(
        ISymbol symbol,
        string qualifiedName,
        Location? location,
        string? sourceOwner)
    {
        var identity = qualifiedName;
        if (symbol is ILocalSymbol or IParameterSymbol)
        {
            var owner = symbol.ContainingSymbol is null ? "" : QualifiedName(symbol.ContainingSymbol);
            var where = location is null || string.IsNullOrEmpty(sourceOwner)
                ? "external"
                : $"{sourceOwner}:{location.SourceSpan.Start}";
            identity = $"{owner}::{symbol.Kind}:{identity}@{where}";
        }
        if (location is not null && symbol.ContainingAssembly is { Name.Length: > 0 } assembly)
        {
            identity = $"{assembly.Name}::{identity}";
        }
        return identity;
    }

    private static string NodeKind(ISymbol symbol)
    {
        return symbol switch
        {
            INamespaceSymbol => "namespace",
            INamedTypeSymbol { TypeKind: TypeKind.Interface } => "interface",
            INamedTypeSymbol => "type",
            IMethodSymbol { MethodKind: MethodKind.Constructor or MethodKind.StaticConstructor } => "constructor",
            IMethodSymbol { MethodKind: MethodKind.LocalFunction } => "function",
            IMethodSymbol => "method",
            IFieldSymbol or IPropertySymbol or IEventSymbol => "field",
            IParameterSymbol => "parameter",
            ILocalSymbol => "variable",
            _ => "symbol",
        };
    }
}
