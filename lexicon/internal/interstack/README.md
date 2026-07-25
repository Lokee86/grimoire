# Interstack resolver

This package derives repository-wide contract facts after ordinary Lexicon adapters have produced a complete candidate snapshot.

It owns:

- HTTP request-to-route and route-to-handler resolution;
- packet/message producer and consumer channels;
- shared environment/configuration key nodes;
- conservative unresolved records when a target is missing or ambiguous; and
- deterministic facts-v1 output for the synthetic `interstack` library.

Language adapters continue to own language syntax and symbol facts. Arcana owns graph storage and traversal. Grimoire consumes the resulting relationships for bounded cross-stack retrieval.

The first supported framework patterns are Rails routes, Go HTTP registrations, common JavaScript/Python HTTP clients, GDScript URL builders and request calls through those helpers, packet type fields and switch/registration handlers, and environment reads in Go, Ruby, TypeScript, GDScript, and Python.

Contract nodes use `@interstack/...` paths so they remain shared graph nodes rather than pretending to be repository source files. Definite relationships retain confidence, source evidence, and spans. String similarity alone is not enough to emit a definite edge.
