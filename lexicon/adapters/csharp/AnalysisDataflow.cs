using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

internal sealed partial class Analysis
{
    private void EmitDataflow()
    {
        foreach (var document in model.Documents)
        {
            var semantic = document.Compilation.GetSemanticModel(document.SyntaxTree, ignoreAccessibility: true);
            foreach (var identifier in document.SyntaxTree.GetRoot().DescendantNodes().OfType<IdentifierNameSyntax>())
            {
                var symbol = semantic.GetSymbolInfo(identifier).Symbol;
                if (symbol is not IFieldSymbol and not IPropertySymbol and not IEventSymbol and
                    not ILocalSymbol and not IParameterSymbol)
                {
                    continue;
                }
                var sourceId = SourceFor(identifier, semantic, document);
                var targetId = EnsureSymbol(symbol, document.RelativePath);
                var access = Access(identifier);
                if (access.Read)
                {
                    AddEdge(sourceId, targetId, "reads", document.RelativePath, identifier.GetLocation());
                }
                if (access.Write)
                {
                    AddEdge(sourceId, targetId, "writes", document.RelativePath, identifier.GetLocation());
                }
            }
        }
    }

    private static (bool Read, bool Write) Access(IdentifierNameSyntax identifier)
    {
        for (SyntaxNode? node = identifier; node is not null; node = node.Parent)
        {
            if (node.Parent is AssignmentExpressionSyntax assignment && IsAssignedTarget(identifier, assignment.Left))
            {
                return assignment.IsKind(SyntaxKind.SimpleAssignmentExpression)
                    ? (false, true)
                    : (true, true);
            }
            if (node.Parent is PrefixUnaryExpressionSyntax prefix && IsIncrement(prefix.Kind()))
            {
                return (true, true);
            }
            if (node.Parent is PostfixUnaryExpressionSyntax postfix && IsIncrement(postfix.Kind()))
            {
                return (true, true);
            }
            if (node.Parent is ArgumentSyntax argument && argument.Expression.Span.Contains(identifier.Span))
            {
                if (argument.RefKindKeyword.IsKind(SyntaxKind.OutKeyword))
                {
                    return (false, true);
                }
                if (argument.RefKindKeyword.IsKind(SyntaxKind.RefKeyword))
                {
                    return (true, true);
                }
                return (true, false);
            }
            if (node is StatementSyntax or MemberDeclarationSyntax)
            {
                break;
            }
        }
        return (true, false);
    }

    private static bool IsAssignedTarget(IdentifierNameSyntax identifier, ExpressionSyntax left)
    {
        if (left == identifier)
        {
            return true;
        }
        return left is MemberAccessExpressionSyntax member && member.Name.Span.Contains(identifier.Span);
    }

    private static bool IsIncrement(SyntaxKind kind)
    {
        return kind is SyntaxKind.PreIncrementExpression or SyntaxKind.PreDecrementExpression or
            SyntaxKind.PostIncrementExpression or SyntaxKind.PostDecrementExpression;
    }
}
