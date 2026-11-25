# Predicate specifications

| Predicate | Type URI | Status |
|---|---|---|
| [Agent Work](agent-work-v0.md) | `https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0` | draft |

A second type, `agent-review/v0` (a signed review of agent work), is planned
and not yet written.

From `v1` on, versioning follows the in-toto
[versioning rules](https://github.com/in-toto/attestation/blob/main/spec/versioning.md):
the major version is in the Type URI and minor versions only add optional
fields. Consumers always ignore fields they do not recognise. `v0` is this
project's own deviation from those rules: upstream gives each 0.X version its
own Type URI, while here every 0.x draft shares the `/v0` URI and may change
in incompatible ways under it. The required `predicate.specVersion` field
names the exact draft, and consumers reject drafts they do not support.

Why this project defines its own predicate instead of extending an upstream
one: [ADR 0001](../docs/decisions/0001-own-predicate.md).

`spec_test.go` checks that each spec's example is a complete Statement of its
type, that every field in the example is documented, and that relative links
resolve. It runs as part of `scripts/check.sh`.

The prose in this directory is licensed under
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); code is Apache-2.0.
