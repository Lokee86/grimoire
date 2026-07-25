package agentquery

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) inspect(ctx context.Context, request Request, response *Response) error {
	values := append([]string(nil), request.Handles...)
	if request.Anchor != "" {
		values = append(values, request.Anchor)
	}
	for _, value := range unique(values) {
		if len(response.Inspections) >= request.Limit {
			response.Truncated = true
			break
		}
		if strings.HasPrefix(value, "grimoire:v1:") {
			handle, err := parseHandle(value)
			if err != nil {
				return err
			}
			if err := engine.validateSnapshot(handle); err != nil {
				return err
			}
			inspection, err := engine.inspectHandle(ctx, handle, request.Adjacent)
			if err != nil {
				return err
			}
			response.Inspections = append(response.Inspections, inspection)
			continue
		}
		resolved, err := engine.resolveAnchors(ctx, value, "", request.Limit-len(response.Inspections))
		if err != nil {
			return err
		}
		for _, node := range resolved.lexicon {
			handle := nodeHandle("lexicon", engine.lexiconSnapshot, node)
			inspection, inspectErr := engine.inspectHandle(ctx, handle, request.Adjacent)
			if inspectErr != nil {
				return inspectErr
			}
			response.Inspections = append(response.Inspections, inspection)
		}
		if len(resolved.lexicon) == 0 {
			for _, node := range resolved.arcana {
				handle := nodeHandle("arcana", engine.arcanaSnapshotID, node)
				inspection, inspectErr := engine.inspectHandle(ctx, handle, request.Adjacent)
				if inspectErr != nil {
					return inspectErr
				}
				response.Inspections = append(response.Inspections, inspection)
			}
		}
	}
	if len(response.Inspections) == 0 {
		return fmt.Errorf("no inspectable source matched")
	}
	return nil
}

func (engine *Engine) inspectHandle(
	ctx context.Context,
	handle Handle,
	adjacent int,
) (Inspection, error) {
	switch handle.Provider {
	case "source":
		span := structure.Span{
			Path: handle.Path, StartLine: handle.StartLine, EndLine: handle.EndLine,
		}
		return engine.sourceInspection(handle, nil, span, adjacent)
	case "lexicon":
		nodes := engine.lexicon.Resolve(handle.NodeIdentity, 1)
		if len(nodes) != 1 {
			return Inspection{}, fmt.Errorf("Lexicon handle node %q is unavailable", handle.NodeIdentity)
		}
		node := nodes[0]
		if node.Span == nil {
			return Inspection{}, fmt.Errorf("Lexicon node %q has no source span", handle.NodeIdentity)
		}
		outputNode := engine.node("lexicon", engine.lexiconSnapshot, node)
		return engine.sourceInspection(handle, &outputNode, *node.Span, adjacent)
	case "arcana":
		if handle.NodeID == nil {
			return Inspection{}, fmt.Errorf("Arcana handle has no node ID")
		}
		node, err := engine.arcana.Inspect(ctx, engine.arcanaSnapshot, *handle.NodeID)
		if err != nil {
			return Inspection{}, err
		}
		if node.Identity != handle.NodeIdentity && handle.NodeIdentity != "" {
			return Inspection{}, fmt.Errorf("Arcana handle identity does not match node %d", *handle.NodeID)
		}
		if node.Span == nil {
			return Inspection{}, fmt.Errorf("Arcana node %d has no source span", *handle.NodeID)
		}
		outputNode := engine.node("arcana", engine.arcanaSnapshotID, node)
		return engine.sourceInspection(handle, &outputNode, *node.Span, adjacent)
	default:
		return Inspection{}, fmt.Errorf("unsupported handle provider %q", handle.Provider)
	}
}

func (engine *Engine) sourceInspection(
	handle Handle,
	node *Node,
	declaration structure.Span,
	adjacent int,
) (Inspection, error) {
	source, containing, err := sourceRange(engine.source, declaration.Path, declaration.StartLine, declaration.EndLine, adjacent)
	if err != nil {
		return Inspection{}, err
	}
	result := Inspection{
		Handle: handle, Node: node, Source: source,
		ContainingSpan: Range{
			Path: normalizePath(declaration.Path), StartLine: containing.StartLine,
			EndLine: containing.EndLine,
			Handle:  sourceHandle(engine.source.Identity(), declaration.Path, containing.StartLine, containing.EndLine),
		},
	}
	declarationRange := rangeFromStructure(declaration, engine.source.Identity())
	result.Declaration = &declarationRange
	return result, nil
}

func sourceRange(
	snapshot index.Snapshot,
	path string,
	start, end, adjacent int,
) (string, structure.Span, error) {
	path = normalizePath(path)
	var chunks []index.Chunk
	for _, file := range snapshot.Files {
		if normalizePath(file.Path) == path {
			chunks = append(chunks, file.Chunks...)
			break
		}
	}
	if len(chunks) == 0 {
		return "", structure.Span{}, fmt.Errorf("source path %q is not in prepared snapshot", path)
	}
	sort.SliceStable(chunks, func(i, j int) bool {
		if chunks[i].StartLine != chunks[j].StartLine {
			return chunks[i].StartLine < chunks[j].StartLine
		}
		return chunks[i].ID < chunks[j].ID
	})
	minLine, maxLine := chunks[0].StartLine, chunks[0].EndLine
	lines := make(map[int]string)
	for _, chunk := range chunks {
		maxLine = max(maxLine, chunk.EndLine)
		parts := strings.Split(strings.ReplaceAll(chunk.Text, "\r\n", "\n"), "\n")
		for index, part := range parts {
			line := chunk.StartLine + index
			if line > chunk.EndLine {
				break
			}
			if prior, exists := lines[line]; exists && chunk.StartLine == chunk.EndLine {
				lines[line] = prior + part
			} else {
				lines[line] = part
			}
		}
	}
	if start <= 0 {
		start = minLine
	}
	if end < start {
		end = start
	}
	if start < minLine || end > maxLine {
		return "", structure.Span{}, fmt.Errorf(
			"source range %s:%d-%d is outside prepared snapshot bounds %d-%d",
			path, start, end, minLine, maxLine,
		)
	}
	containingStart := max(minLine, start-adjacent)
	containingEnd := min(maxLine, end+adjacent)
	values := make([]string, 0, containingEnd-containingStart+1)
	for line := containingStart; line <= containingEnd; line++ {
		values = append(values, lines[line])
	}
	return strings.Join(values, "\n"), structure.Span{
		Path: path, StartLine: containingStart, EndLine: containingEnd,
	}, nil
}
