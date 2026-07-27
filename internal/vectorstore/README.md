# Vector store boundary

This package preserves Grimoire's internal import path while delegating vector persistence and exact search to Lodestone's shared Go binding.

Lodestone owns:

- content-addressed immutable vector objects;
- deterministic packed snapshots;
- memory-mapped validation;
- exact parallel top-K search; and
- the stable C ABI and Go loader.

Grimoire owns chunking, source identities, embedding requests, snapshot orchestration, hybrid ranking, and presentation.

`LODESTONE_LIBRARY` selects an explicit Lodestone dynamic library. `GRIMOIRE_VECTOR_ENGINE` and the legacy `grimoire_vector_ffi` ABI remain accepted during migration.
