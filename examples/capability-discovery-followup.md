# Capability discovery follow-up example

A natural question such as “What can this `.gooo` declaration do?” first produces a read-only capability observation. This example describes the next conversation boundary:

- `BOUND` asks which declared capability or operation to inspect.
- `DEFERRED` asks for the explicit external boundary required before execution or authorization.
- `UNKNOWN` asks for missing source identity, contract identity, query, declaration, or catalog evidence.

The follow-up is not an execution plan and never grants permission. The earliest unresolved boundary remains explicit so a later answer cannot relabel missing evidence as success.
