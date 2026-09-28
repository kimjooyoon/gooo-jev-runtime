# Natural Capability Discovery Flow

This flow lets a `.gooo`-based system answer what it may be able to do without executing an operation or granting permission.

## Stages

1. `contracts/jev_capability_discovery_query.gooo` records the original question, normalized terms, candidate capability IDs, unresolved terms, and a read-only next question.
2. `contracts/jev_capability_graph_shortlist.gooo` keeps discovery deterministic: declared candidates are collected first, authorization evidence is checked before ranking, and ranking never grants authorization.
3. `contracts/jev_domain_capability_measurement.gooo` measures observed domain coverage and unresolved capability boundaries. Its signal is an investment or observation prompt, not a correctness score.
4. The runtime returns `BOUND`, `DEFERRED`, `NEEDS_INPUT`, or `UNKNOWN` when discovery is incomplete. Missing terms and missing evidence remain explicit.
5. Any later execution must use a separate, explicit authorization and execution boundary.

## Example question

See `examples/capability-discovery-query.json`. The Korean question is only an input example; normalization and candidate binding must remain deterministic and provenance-linked.

## Provenance boundary

Each response binds the query digest and provenance digest to the source catalog and contract identity. A candidate is not treated as available merely because its name resembles a natural-language term. An unresolved term must either remain unresolved or produce a read-only clarification question.

## Non-goals

This flow does not execute tools, mutate data, issue credentials, infer semantic completeness, or convert a metric into an authorization decision.