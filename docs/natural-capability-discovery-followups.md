# Natural Capability Discovery Follow-ups

`jevcal.SuggestCapabilityDiscoveryFollowUps` turns a capability discovery observation into a small deterministic set of read-only questions.

- `BOUND` asks which discovered capability or declared operation should be inspected next.
- `DEFERRED` asks for the explicit external boundary required before execution or authorization.
- `UNKNOWN` asks for the missing source identity, contract identity, query, declaration, or catalog term.

The follow-up does not select a capability, execute a command, grant permission, or infer semantic completeness. Its `first_boundary` preserves the earliest unresolved stage so later conversation cannot relabel missing evidence as success.
