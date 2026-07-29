# Supplemental LevelDB Grimoire/Hermes rerun

Date: July 29, 2026  
Purpose: determine whether the primary 18-call, 1.20M-token Grimoire result was a one-off model outlier  
Included in primary benchmark totals: **no**

## Fixed conditions

- Task: `leveldb-background-compaction-pause`
- LevelDB revision: `99b3c03b3284f5886f9ef9a4ef703d57373e61be`
- Model: `gpt-5.6-sol`
- Reasoning effort: `high`
- Service tier: Fast, audited as `priority`
- Runner: Hermes one-shot
- Condition: Grimoire only
- Prompt: same read-only benchmark prompt and task wording as the primary run
- Checkout and Hermes profile: fresh and isolated

## Result

| Metric | Primary Grimoire | Supplemental rerun | Change |
|---|---:|---:|---:|
| API calls | 18 | **19** | +1 |
| Fresh input | 100,796 | **151,660** | +50,864 |
| Cache reads | 1,084,928 | **1,536,000** | +451,072 |
| Output | 17,509 | 15,611 | -1,898 |
| Reasoning | 11,204 | 7,775 | -3,429 |
| **Total tokens** | **1,203,233** | **1,703,271** | **+500,038 / +41.6%** |
| Grimoire calls | 2 | **3** | +1 |
| Grimoire response bytes | 28,917 | **53,024** | +24,107 |
| Grimoire tool errors | 0 | 0 | none |

The second run did not regress to the 12-to-13-call trajectory seen in Plain and CBM. It repeated the long trajectory and consumed substantially more context than the first Grimoire run. This indicates that the LevelDB result was not merely one isolated stochastic failure, although model variance affected its exact magnitude.

## Answer-capture limitation

Hermes completed the model run and wrote a valid usage report:

- `completed: true`
- `failed: false`
- `service_tier: priority`
- session: `20260729_143928_9c23ce`

After completion, the wrapper attempted to decode captured output using the Windows CP1252 default and encountered one undecodable byte. The resulting reader failure left the final answer unavailable. The usage and Grimoire audit records remain valid, but the answer cannot be rescored for quality or grounding. The rerun is therefore evidence about token use, retrieval behavior, and stopping trajectory only.

## Interpretation

The primary LevelDB run had already shown that Grimoire retrieved the correct ownership seam with the least fresh input but Hermes continued for 18 calls. The supplemental run made 19 calls and used 1.70M tokens. Together they support the following interpretation:

- Grimoire is not necessarily helpful when a task is narrow and direct inspection is already cheap.
- Structured evidence can encourage the model to continue exploring rather than conclude.
- Retrieval succeeded; end-to-end stopping behavior failed.
- A compact narrow-task response mode and stronger stopping guidance are plausible optimization targets.
- More narrow tasks and repeated seeds are required before generalizing the exact magnitude.

## Artifacts

- `grimoire.usage.json`: accepted audited usage.
- `grimoire.mcp-audit.jsonl`: three successful Grimoire calls.
- `grimoire-prewarm.stdout.txt`: fresh-state preparation details and timings.
- `grimoire-prewarm.stderr.txt`: preparation stderr.
- `recovered-result.json`: primary-versus-rerun usage comparison and cleanup verification.

## Cleanup and restoration

- Temporary checkout removed.
- Temporary Hermes profile removed.
- Temporary benchmark scripts removed.
- Hermes `oneshot.py` restored byte-for-byte.
- Restored SHA-256: `3c8b2233b803aa873538049e911d21bd12ca9a75c453744e54fff57a57e1244e`.
