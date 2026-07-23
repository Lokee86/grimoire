# Retrieval evaluation: Gum

Generated: 2026-07-23 06:55:11-07:00  
Variant: `ranking-metrics-baseline`  
Source: `https://github.com/charmbracelet/gum`  
Revision: `716d8b5d0221558f944b5a078dbbcca8572534fb`  
Scope: `.`  
Judged: `2026-07-23`  
Cases: 5  
Runs: 5  
Structural providers: ``

## Mode comparison

| Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 0.0% | 28.6% | 0.0% | 74.1% | 0.0% | 1749.1 ms | 3642.5 ms |

## Pre-curation source ranking

These metrics score the retrieved order before exact-result merging, curation, and package fitting.

| Mode | Queries | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 5 | 26.7% | 31.7% | 0.476 | 34.0% | 24.0% |

## Category comparison

| Category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 33.3% | 0.0% | 66.7% | 0.0% | 1749.1 ms | 1749.1 ms |
| call-chain-investigation | 0.0% | 33.3% | 0.0% | 75.0% | 0.0% | 3743.2 ms | 3743.2 ms |
| direct-location | 0.0% | 0.0% | 0.0% | 80.0% | 0.0% | 962.6 ms | 962.6 ms |
| long-mixed-query | 0.0% | 50.0% | 0.0% | 60.0% | 0.0% | 3239.5 ms | 3239.5 ms |
| mechanism-explanation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1741.7 ms | 1741.7 ms |

## Mode by category

| Mode/category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical/architecture-ownership | 0.0% | 33.3% | 0.0% | 66.7% | 0.0% | 1749.1 ms | 1749.1 ms |
| lexical/call-chain-investigation | 0.0% | 33.3% | 0.0% | 75.0% | 0.0% | 3743.2 ms | 3743.2 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 80.0% | 0.0% | 962.6 ms | 962.6 ms |
| lexical/long-mixed-query | 0.0% | 50.0% | 0.0% | 60.0% | 0.0% | 3239.5 ms | 3239.5 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1741.7 ms | 1741.7 ms |

## Per-case results

| Case | Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| gum-dl-01 | lexical | false | 0.0% | 0.0% | 80.0% | 0.0% | 2946 | 962.6 ms | budget-fitting loss |
| gum-me-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 5985 | 1741.7 ms | budget-fitting loss |
| gum-ao-01 | lexical | false | 33.3% | 0.0% | 66.7% | 0.0% | 4956 | 1749.1 ms | budget-fitting loss |
| gum-cc-01 | lexical | false | 33.3% | 0.0% | 75.0% | 0.0% | 7958 | 3743.2 ms | budget-fitting loss |
| gum-lm-01 | lexical | false | 50.0% | 0.0% | 60.0% | 0.0% | 8936 | 3239.5 ms | budget-fitting loss |

## Concrete failures

- `gum-dl-01` / `lexical`: budget-fitting loss
  - `gum.go` symbols `Gum`: budget-fitting loss
  - `main.go` symbols `main`: budget-fitting loss
- `gum-me-01` / `lexical`: budget-fitting loss
  - `filter/command.go` symbols `Run`, `checkSelected`: budget-fitting loss
  - `filter/filter.go` symbols `Update`, `exactMatches`, `matchAll`, `ToggleSelection`: budget-fitting loss
- `gum-ao-01` / `lexical`: budget-fitting loss
  - `internal/timeout/context.go` symbols `Context`: budget-fitting loss
  - `main.go` symbols `main`: budget-fitting loss
- `gum-cc-01` / `lexical`: budget-fitting loss
  - `spin/spin.go` symbols `Init`, `commandStart`, `Update`, `commandAbort`: budget-fitting loss
  - `spin/pty.go` symbols `openPty`: budget-fitting loss
- `gum-lm-01` / `lexical`: budget-fitting loss
  - `file/file.go` symbols `Update`, `View`, `helpView`: budget-fitting loss
  - `internal/timeout/context.go` symbols `Context`: budget-fitting loss
