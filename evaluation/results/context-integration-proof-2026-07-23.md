# Context integration proof — 2026-07-23

## Integrated work

- Shared provider-neutral evidence descriptors.
- Reciprocal-rank fusion for lexical and vector providers.
- Lexicon and Arcana source-group linking.
- Bounded structural source anchoring during adaptive assembly.
- Retrieval-intent decomposition contract for query planning.

## Verification

All checks passed on `codex/context-integration` after merging current `main`:

```text
go test ./...
go vet ./...
go test -race ./...
```

Prepared-state proof used index format version 3 with generated-content exclusion and bounded chunk sizing:

```text
files: 234
generated files skipped: 3
prepared chunks: 3940
unique vectors: 3799
```

Lexicon and Arcana used the same immutable snapshot:

```text
sha256:e8afcfe67911aa16bf1b24fe27af0a4c78c22f02565c3c96b5fd0114ed9a4691
```

## Paired lexical + Lexicon corpus

Twelve Grimoire retrieval cases, adaptive assembly, identical prepared and Lexicon state.

| Build | Required source recall | Irrelevant source selections | Median latency |
| --- | ---: | ---: | ---: |
| Current main | 2.2% | 93.5% | 2187.6 ms |
| Integrated | 2.2% | 93.5% | 1829.1 ms |

Result: no quality regression. Median latency was 16.4% lower in this paired run.

## Paired hybrid + Lexicon + Arcana corpus

Twelve Grimoire retrieval cases, adaptive assembly, identical prepared, vector, Lexicon, and Arcana state.

| Build | Required source recall | Irrelevant source selections | Median latency |
| --- | ---: | ---: | ---: |
| Current main | 0.0% | 98.5% | 1969.8 ms |
| Integrated | 8.9% | 89.2% | 2050.7 ms |

Result: source recall increased by 8.9 percentage points and irrelevant source selection fell by 9.3 percentage points. Median latency increased by 80.9 ms.

Integrated recall by category:

| Category | Required source recall |
| --- | ---: |
| Architecture ownership | 12.5% |
| Call-chain investigation | 16.7% |
| Long mixed query | 7.7% |
| Direct location | 0.0% |
| Mechanism explanation | 0.0% |

## Defects found during integration

The initial package-fitting implementation serialized every represented evidence-group ID into package metadata. That consumed enough token budget to evict required source evidence. It was replaced with a bounded count.

The initial group-priority implementation recursively activated groups from every selected Lexicon source candidate. That reordered unrelated candidates and regressed corpus quality. It was replaced with bounded structural anchoring:

- Preserve the first eight curated candidates.
- Use only retained structural evidence to activate groups.
- Promote at most one source anchor per group.
- Promote at most eight groups.
- Do not recursively activate groups from source candidates.

## Remaining limitations

- Structural required-evidence recall remains 0% on this corpus, and retained structural facts were scored as irrelevant by the current structural rubric. Source anchoring improved source evidence despite that weakness.
- Retrieval intents are emitted in the policy contract but are not yet consumed by retrieval. That portion is infrastructure, not a measured quality improvement.
- Hybrid retrieval remains weak before assembly: required Recall@10 and Recall@20 were both 0% in this corpus. The measured gain came from structural source anchoring during assembly.
- Current `main` repeatedly crashed during the hybrid-only baseline with Windows access violation `0xc0000005`; the integrated build completed. This was treated as a stability observation, not as a quality metric.
