gooo-jev-runtime

A public Go and .gooo companion for bounded JEV calibration, evidence-linked provenance, and safe self-improvement experiments.

Scope

- typed choice sets and observed outcomes
- deterministic calibration measurements
- source and observation digests for provenance
- explicit UNKNOWN and DEFERRED states when evidence is incomplete or incomparable
- a .gooo contract that mirrors the Go boundary

It is not an agent runtime, a policy engine, or an authorization system. It does not execute generated code, grant permissions, or treat cache presence as semantic evidence.

Initial architecture

- jevcal/: dependency-free Go value objects and calibration comparison
- contracts/jev_runtime.gooo: language-side declaration of the same boundary
- .github/workflows/ci.yml: GitHub-hosted validation only

The first primitive is a bounded calibration window. A window is comparable only when its ordered choice set and evidence lineage are stable. Otherwise the result remains DEFERRED or UNKNOWN; it is never silently classified as an improvement.

Direction

1. Keep the Go and .gooo contracts structurally aligned.
2. Add outcome-backed Brier/log-loss measurements without inventing labels.
3. Project decisions into LSP diagnostics/code actions while preserving digests.
4. Add reverse observation so every projection can be checked against its source evidence.
5. Keep all execution, authorization, and security policy outside this read-only measurement layer.

See LICENSE for the Apache-2.0 terms.