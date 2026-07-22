# Go adapter real-repository validation

Validated on July 21, 2026 against two existing Go modules without modifying
either source repository.

## Results

| Repository | Nodes | Edges | Go files | Packages | Call expressions | Resolved direct calls | Unresolved calls | Packed graph | Catalogue | Import time |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Demon Docs | 2,871 | 5,964 | 353 | 33 | 14,727 | 3,333 | 11,394 | 117,648 B | 350,220 B | 0.233 s |
| Space Rocks game server | 4,091 | 7,651 | 528 | 53 | 15,370 | 3,497 | 11,873 | 157,424 B | 591,426 B | 0.263 s |

The first-pass resolver covered approximately 22.6% of call expressions in
Demon Docs and 22.8% in the Space Rocks game server. This is a deliberately
narrow baseline: only unambiguous, unqualified, same-package function calls are
resolved.

## Query checks

Demon Docs `ManagedRootTitle` resolved to its source declaration, outgoing calls
to `FirstHeadingTitle` and `TitleFromName`, and the reverse call from
`TestTitlesAndRootTitleFallbacks`.

Space Rocks `ProjectEventLane` resolved to its source declaration, its outgoing
call to `sequenceBackedBatchID`, and reverse calls from `BuildEventBatchPacket`
and four tests.

Both fact files were regenerated and compared byte-for-byte. The repeated
outputs were identical.

## Current omissions

The adapter does not yet resolve:

- selector and method calls;
- interface dispatch;
- internal cross-package calls;
- built-ins and calls through variables;
- calls nested in function literals;
- recursive self-calls, because the current packed graph rejects self-edges;
- build-tag and generated-code semantics beyond ordinary parsing.

These omissions are counted as unresolved calls rather than represented as
possibly incorrect edges.
