# Post-mitigation LevelDB Grimoire/Hermes rerun

Date: July 29, 2026  
Included in primary benchmark totals: **no**

## Purpose

Measure whether the new narrow-task path changes the repeated 18-to-19-call LevelDB trajectory after adding:

- cheap-first routing guidance;
- combined narrow evidence budgeting;
- handle-only discovery;
- explicit discovery-versus-inspection separation;
- conservative evidence assessment;
- a two-call stopping contract.

## Fixed conditions

- Task: `leveldb-background-compaction-pause`
- LevelDB revision: `99b3c03b3284f5886f9ef9a4ef703d57373e61be`
- Prompt: unchanged read-only benchmark prompt and task wording
- Model: `gpt-5.6-sol`
- Reasoning effort: `high`
- Service tier: Fast, audited as `priority`
- Runner: Hermes one-shot
- Condition: Grimoire only
- Checkout, Grimoire build, Hermes profile, and repository state: fresh and isolated
- Grimoire source: committed head `748977977bb66c51abfb3c98ea69328c791b03d6` plus the recorded uncommitted mitigation diff

## Result

| Metric | Primary Grimoire | Second rerun | Post-mitigation rerun |
|---|---:|---:|---:|
| API calls | 18 | 19 | **13** |
| Fresh input | 100,796 | 151,660 | 126,997 |
| Cache reads | 1,084,928 | 1,536,000 | **625,152** |
| Output | 17,509 | 15,611 | 14,876 |
| Reasoning | 11,204 | 7,775 | 8,690 |
| **Total tokens** | **1,203,233** | **1,703,271** | **767,025** |
| Model time | 422.0s | capture-limited | **388.9s** |
| Grimoire calls | 2 | 3 | **2** |
| Grimoire response bytes | 28,917 | 53,024 | **9,722** |
| Grounding | valid | unscored | **valid** |
| Manual quality | 8/8 | unscored | **7/8** |

Relative to the primary Grimoire run, the post-mitigation rerun used:

- 5 fewer API calls, a **27.8% reduction**;
- 436,208 fewer total tokens, a **36.3% reduction**;
- 459,776 fewer cache-read tokens, a **42.4% reduction**;
- 19,195 fewer Grimoire response bytes, a **66.4% reduction**;
- 33.1 fewer model-runtime seconds, a **7.9% reduction**.

Relative to the second rerun, it used 6 fewer calls and 936,246 fewer total tokens, a **55.0% reduction**.

## Comparison with Plain and CBM

| Condition | API calls | Total tokens | Model time | Quality |
|---|---:|---:|---:|---:|
| Plain | 13 | 788,781 | 399.1s | 8/8 |
| CBM | 12 | 699,584 | 387.6s | 8/8 |
| Post-mitigation Grimoire | **13** | **767,025** | **388.9s** | **7/8** |

The revised Grimoire run matched Plain's 13-call trajectory, used 21,756 fewer total tokens than Plain, and finished 10.2 seconds faster. It remained 67,441 tokens above CBM and 1.4 seconds slower in model time.

Fresh Grimoire preparation took 8.9 seconds. Excluding the one-time product build, preparation plus model execution was 397.8 seconds: approximately 1.3 seconds below Plain's no-preparation runtime and 9.3 seconds above CBM including its 0.9-second preparation.

## Retrieval trajectory

The agent followed the intended workflow exactly:

1. One `search` call with `breadth: "narrow"` and `detail: "handles"`.
   - Four combined discovery items.
   - 3,155 response bytes.
   - Assessment: owner, control flow, and public boundary observed; tests missing.
   - Next action: inspect selected handles.
2. One `inspect` call over three selected handles.
   - 6,567 response bytes.
   - No third Grimoire call.

The complete audit contained two calls, 9,722 response bytes, four newly discovered nodes, six source ranges materialized during inspection, and zero tool errors.

## Grounding and quality

Automatic grounding passed:

- process completed;
- evidence JSON valid;
- 46 inline citations;
- 11 structured evidence items;
- 2 canonical Grimoire handle items;
- all required `db/db_impl`, `include/leveldb`, and `db` evidence prefixes present;
- no grounding findings.

Manual score: **7/8**.

The answer correctly covered:

- the mutex-protected per-`DBImpl` ownership seam;
- automatic-only admission and dispatch gating;
- preservation of immutable-memtable flushing and public manual compaction;
- L0 write-pressure and stop-threshold behavior;
- queued and in-flight callback races;
- Boolean-versus-token ownership semantics;
- deterministic pause, manual, resume, and race tests.

It lost one verification point because the proposed test plan did not explicitly include closing or destroying the database while automatic compaction remained paused. The earlier primary answer covered shutdown explicitly.

## Interpretation

The narrow-task mitigation materially improved the end-to-end trajectory in this run. It did not merely shrink the MCP payload: the agent stopped after 13 calls rather than repeating the earlier 18-to-19-call pattern, and total token use moved into the same range as Plain and CBM.

This is still one post-change seed. It demonstrates that the new path can solve the observed failure mode, not that every narrow task or every run will do so. Repeated seeds and additional narrow tasks remain necessary to estimate reliability.

## Rejected non-priority trial

Before this accepted run, the same mitigation produced a 16-call, 895,854-token result. Hermes reported `service_tier: null`, so that trial is retained under `rerun3/` but excluded from the controlled comparison. Its retrieval path was also one narrow search followed by one inspection.

## Artifacts

- `run.json`: full accepted-run provenance, timing, usage, grounding, and build identities.
- `grimoire.stdout.txt`: captured final answer.
- `grimoire.usage.json`: audited Hermes usage.
- `grimoire.grounding.json`: automatic grounding result.
- `grimoire.mcp-audit.jsonl`: exact two-call Grimoire trace.
- `grimoire-prewarm.stdout.txt`: fresh-state preparation output.
- `comparison.json`: machine-readable comparison with prior LevelDB runs.

The temporary Hermes bridge was restored byte-for-byte after the run. The isolated profile, checkout, and build were removed.
