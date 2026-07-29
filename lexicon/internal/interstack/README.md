# Interstack resolver

This package derives repository-wide contract facts after ordinary Lexicon adapters have produced a complete candidate snapshot.

It owns:

- HTTP request-to-route and route-to-handler resolution;
- packet/message producer and consumer channels;
- executable and CLI-command ownership across repository components;
- JSONL protocol producer/consumer contracts;
- immutable state paths and snapshot handoffs;
- shared environment/configuration key nodes;
- conservative unresolved records when a target is missing or ambiguous; and
- deterministic facts-v1 output for the synthetic `interstack` library.

Language adapters continue to own language syntax and symbol facts. Arcana owns graph storage and traversal. Grimoire consumes the resulting relationships for bounded cross-stack retrieval.

The first supported framework patterns are Rails routes, Go HTTP registrations, common JavaScript/Python HTTP clients, GDScript URL builders and request calls through those helpers, packet type fields and switch/registration handlers, and environment reads in Go, Ruby, TypeScript, GDScript, and Python.

Repository boundary resolution also recognizes Go `exec.Command` and `exec.CommandContext` calls to Lexicon and Arcana, Go and Rust CLI dispatch ownership, the `arcana.query.v1` JSONL protocol, `.lexicon` and `.arcana` snapshot state, and the environment variables that configure provider discovery and Lexicon state. Rust is scanned by this post-pass only for these repository-wide contracts; Rust syntax and local semantics remain owned by the Rust adapter.

Packet channels require both dispatch context and a message-like semantic suffix such as `request`, `response`, `event`, `command`, `message`, or `packet`. Ordinary parser constants and grammar labels are not promoted into channels.

Contract nodes use `@interstack/...` paths so they remain shared graph nodes rather than pretending to be repository source files. Definite relationships retain confidence, source evidence, and spans. String similarity alone is not enough to emit a definite edge.
