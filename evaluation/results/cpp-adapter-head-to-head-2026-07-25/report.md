# C++ adapter head-to-head — July 25, 2026

Same checked-out scopes were indexed by Lexicon 0.5.0 and CBM at commit `97ce23f`.
CBM used `fast` mode, which its CLI documents as retaining per-file and cross-file type-aware LSP call/usage resolution while excluding similarity edges.

## Results

| Case | Split | Lexicon time | CBM time | Lexicon nodes/edges | CBM nodes/edges | Protected judgments |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| fixture | adversarial | 0.391s | 0.102s | 65/113 | 20/31 | — |
| leveldb | calibration | 3.484s | 0.654s | 9201/40025 | 2660/9613 | Lexicon 6/6; CBM 4/6 |
| fmt | calibration | 11.473s | 3.026s | 17056/78953 | 4710/16351 | Lexicon 5/5; CBM 4/5 |
| catch2 | validation | 2.380s | 0.548s | 9853/26299 | 3075/10159 | Lexicon 6/6; CBM 6/6 |
| nlohmann-json | holdout | 1.893s | 1.217s | 3484/17000 | 412/1695 | Lexicon 6/6; CBM 2/6 |

## Aggregate

- Real-corpus protected judgments: Lexicon **23/23**; CBM **16/23**.
- Adversarial fixture checks: Lexicon **10/10**; CBM **2/10**.
- Observed repeated-run indexing time: Lexicon median **2.380s**; CBM median **0.654s**. This is not a cold-cache benchmark and CBM's output graph is materially shallower, so timing is not used for the semantic verdict.

## Fixture details

| Check | Lexicon | CBM |
| --- | --- | --- |
| qualified overload by arity | PASS | FAIL |
| qualified namespace | PASS | FAIL |
| typed parameter receiver | PASS | FAIL |
| explicit local receiver | PASS | FAIL |
| auto reference receiver | PASS | PASS |
| override receiver | PASS | FAIL |
| inherited method | PASS | FAIL |
| template function | PASS | PASS |
| macro-mediated call | PASS | FAIL |
| function-pointer parameter remains explicitly unresolved | PASS | FAIL |

## CBM full-mode confirmation

The adversarial fixture was also indexed in CBM `full` mode. Its targeted call resolution was unchanged from `fast`: four CALLS edges were emitted, including the same high-confidence misroutes from `beta::pick` to the two-argument `alpha::pick`, and from `child.step` to `Derived::step`. Full mode did not recover the omitted overload, typed-receiver, override, macro, or indirect-call evidence.

## Interpretation boundary

This benchmark scores exact protected nodes/relationships and deliberately targeted C++ resolution cases. It does not treat raw edge count as accuracy, and it does not score CBM's unrelated retrieval/UI features. The real-corpus judgments originate in Lexicon's existing acceptance corpus, so they are useful exact checks rather than an independent neutral benchmark. The separate adversarial fixture was added specifically for this head-to-head. Lexicon's definite/possible distinction and provenance are retained; CBM's CALLS confidence/strategy fields are recorded in the JSON report.
