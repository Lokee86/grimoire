using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

internal sealed partial class Analysis
{
    private void EmitRelationships()
    {
        foreach (var document in model.Documents)
        {
            var semantic = document.Compilation.GetSemanticModel(document.SyntaxTree, ignoreAccessibility: true);
            var root = document.SyntaxTree.GetRoot();
            foreach (var directive in root.DescendantNodes().OfType<UsingDirectiveSyntax>())
            {
                if (directive.Name is null)
                {
                    continue;
                }
                var target = semantic.GetSymbolInfo(directive.Name).Symbol;
                var targetId = target is null
                    ? ExternalNamespace(directive.Name.ToString())
                    : EnsureSymbol(target, document.RelativePath);
                AddEdge(fileIds[document.SyntaxTree], targetId, "depends-on", document.RelativePath, directive.GetLocation());
            }

            foreach (var declaration in root.DescendantNodes().OfType<BaseTypeDeclarationSyntax>())
            {
                if (semantic.GetDeclaredSymbol(declaration) is not INamedTypeSymbol type)
                {
                    continue;
                }
                var sourceId = EnsureSymbol(type, document.RelativePath);
                if (type.BaseType is { SpecialType: not SpecialType.System_Object } baseType)
                {
                    AddEdge(sourceId, EnsureSymbol(baseType, document.RelativePath), "extends", document.RelativePath, declaration.GetLocation());
                }
                foreach (var implemented in type.Interfaces)
                {
                    AddEdge(sourceId, EnsureSymbol(implemented, document.RelativePath), "implements", document.RelativePath, declaration.GetLocation());
                }
            }

            foreach (var node in root.DescendantNodes())
            {
                ISymbol? symbol = node switch
                {
                    BaseMethodDeclarationSyntax method => semantic.GetDeclaredSymbol(method),
                    PropertyDeclarationSyntax property => semantic.GetDeclaredSymbol(property),
                    IndexerDeclarationSyntax indexer => semantic.GetDeclaredSymbol(indexer),
                    EventDeclarationSyntax eventDeclaration => semantic.GetDeclaredSymbol(eventDeclaration),
                    _ => null,
                };
                ISymbol? overridden = symbol switch
                {
                    IMethodSymbol method => method.OverriddenMethod,
                    IPropertySymbol property => property.OverriddenProperty,
                    IEventSymbol @event => @event.OverriddenEvent,
                    _ => null,
                };
                if (symbol is not null && overridden is not null)
                {
                    AddEdge(
                        EnsureSymbol(symbol, document.RelativePath),
                        EnsureSymbol(overridden, document.RelativePath),
                        "overrides",
                        document.RelativePath,
                        node.GetLocation());
                }
            }
        }
    }

    private void EmitCalls()
    {
        foreach (var document in model.Documents)
        {
            var semantic = document.Compilation.GetSemanticModel(document.SyntaxTree, ignoreAccessibility: true);
            foreach (var node in document.SyntaxTree.GetRoot().DescendantNodes())
            {
                if (node is not InvocationExpressionSyntax and not ObjectCreationExpressionSyntax and
                    not ImplicitObjectCreationExpressionSyntax and not ConstructorInitializerSyntax)
                {
                    continue;
                }
                var sourceId = SourceFor(node, semantic, document);
                var info = semantic.GetSymbolInfo(node);
                if (info.Symbol is not null)
                {
                    AddEdge(sourceId, EnsureSymbol(info.Symbol, document.RelativePath), "calls", document.RelativePath, node.GetLocation());
                    continue;
                }
                var candidates = info.CandidateSymbols
                    .Select(NormalizeSymbol)
                    .Distinct(SymbolEqualityComparer.Default)
                    .OrderBy(QualifiedName, StringComparer.Ordinal)
                    .ToArray();
                if (candidates.Length is > 0 and <= 4)
                {
                    foreach (var candidate in candidates)
                    {
                        AddEdge(sourceId, EnsureSymbol(candidate, document.RelativePath), "possible-calls", document.RelativePath, node.GetLocation());
                    }
                    continue;
                }
                var attributes = new SortedDictionary<string, object?>(StringComparer.Ordinal)
                {
                    ["candidate_count"] = candidates.Length,
                    ["candidate_reason"] = info.CandidateReason.ToString(),
                };
                if (candidates.Length > 0)
                {
                    attributes["candidate_sample"] = candidates.Take(4).Select(QualifiedName).ToArray();
                }
                facts.AddUnresolved(new UnresolvedRecord
                {
                    Attributes = attributes,
                    CandidateName = CandidateName(node),
                    Expression = node.ToString(),
                    Owner = document.RelativePath,
                    Reason = candidates.Length > 4 ? "ambiguous-target" : "unresolved-symbol",
                    Record = "unresolved",
                    Relation = "calls",
                    Source = sourceId,
                    Span = Span(document.RelativePath, node.GetLocation()),
                });
            }
        }
    }

    private string SourceFor(SyntaxNode node, SemanticModel semantic, SourceDocument document)
    {
        var enclosing = semantic.GetEnclosingSymbol(node.SpanStart);
        while (enclosing is not null && enclosing is IAssemblySymbol or IModuleSymbol)
        {
            enclosing = enclosing.ContainingSymbol;
        }
        return enclosing is null
            ? fileIds[document.SyntaxTree]
            : EnsureSymbol(enclosing, document.RelativePath);
    }

    private string ExternalNamespace(string name)
    {
        return facts.AddNode(
            "namespace",
            name.Split('.').LastOrDefault() ?? name,
            "external",
            name,
            identity: $"external-namespace:{name}",
            attributes: new SortedDictionary<string, object?>(StringComparer.Ordinal) { ["external"] = true });
    }

    private static string? CandidateName(SyntaxNode node)
    {
        return node switch
        {
            InvocationExpressionSyntax invocation => invocation.Expression.ToString(),
            ObjectCreationExpressionSyntax creation => creation.Type.ToString(),
            ImplicitObjectCreationExpressionSyntax => "new",
            ConstructorInitializerSyntax initializer => initializer.ThisOrBaseKeyword.Text,
            _ => null,
        };
    }
}
