# Discovery quality

Grimoire's active quality target is successful progressive repository investigation, not the quality of a preassembled context package.

## Evaluate lanes independently

A search can fail in distinct ways:

- exact recovery misses a concrete literal;
- source BM25 misses implementation evidence;
- document retrieval misses intent or rationale;
- Lexicon misses or misidentifies a symbol;
- Arcana or Lexicon relationships omit a useful edge;
- a returned handle fails exact inspection;
- the agent chooses an irrelevant branch despite adequate discovery.

Reports should preserve this attribution rather than collapsing every failure into final recall.

## Required measurements

For each task record:

- repository revision and prepared-state identities;
- query and completion criteria;
- required source, document, symbol, and relationship evidence;
- evidence returned by each lane and rank within that lane;
- follow-up inspect, trace, and impact calls;
- irrelevant paths or branches opened;
- input tokens, output tokens, tool calls, and elapsed time;
- unsupported conclusions.

## Lane metrics

Within one lane, ordinary retrieval metrics remain useful:

- recall at k;
- reciprocal rank;
- precision at k;
- exact owner or symbol hit rate;
- relationship edge coverage;
- document freshness and citation correctness.

Do not compare raw scores across lanes. BM25, Lexicon, and Arcana values are provider-local signals.

## End-to-end metrics

The agent-discovery evaluator measures whether the investigation found the required evidence and respected the ownership boundary. Useful aggregate metrics include:

- task completion rate;
- required-evidence recall;
- structural-evidence recall;
- unsupported-conclusion rate;
- median discovery calls;
- median input and output tokens;
- median time to first required evidence;
- median time to complete evidence;
- irrelevant branch count.

## Assisted-agent comparisons

Use identical tasks, revisions, agent models, normal repository tools, skills, and completion criteria. Give each assisted condition exactly one optional discovery system. Do not provide a hidden prepared answer or exclude setup and refresh costs unless equivalent costs are excluded for every system.

Compare at least:

- answer correctness and evidence support;
- owner-file and symbol discovery;
- relationship/path discovery;
- tool-call count;
- source-open count;
- token use;
- latency;
- irrelevant exploration.

A lower call count is not automatically better if required evidence is missing. A higher recall is not automatically better if it floods the agent with irrelevant branches. Report both.

## Corpus design

Cases should be concrete and implementation-checkable. Each case needs:

- a task;
- an ownership boundary;
- required evidence paths and symbols;
- required structural evidence where applicable;
- forbidden unsupported conclusions;
- completion criteria;
- known relevant branches.

Include exact literals, local symbol ownership, cross-file call paths, configuration readers, documentation rationale, mixed source/document questions, cross-language generated contracts, architecture plans, and impact analysis. Report results by task class rather than averaging lookup and architecture work into one undifferentiated score.

## Historical package evaluation

The repository contains older retrieval and package-fitting corpora and reports. They remain historical calibration artifacts for the retired context pipeline. They must not be presented as current unified-discovery benchmarks.
