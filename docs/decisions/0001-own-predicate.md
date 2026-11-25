# ADR 0001: Define our own predicate type instead of extending SCAI or Runtime Traces

Status: accepted
Date: 2026-11-25

## Context

agent-signoff records that a git commit came out of a session with an AI
coding agent: agent, model, instruction-file digests, policy-file digest and a
transcript digest. It wants an in-toto Statement v1 that stock `cosign` can
verify, so the record needs a predicate type.

The in-toto predicate layer says users "choose an existing predicate type that
fits their needs, or develop a new one if no existing one satisfies", and that
vetting by the maintainers is optional
([predicate.md](https://github.com/in-toto/attestation/blob/main/spec/v1/predicate.md)).

The [vetted predicate index](https://github.com/in-toto/attestation/blob/main/spec/predicates/README.md),
read on 2026-11-25, lists CycloneDX, Link, Reference, Release, Runtime Traces,
SCAI Report, SLSA Provenance, SLSA Verification Summary, SPDX2, SPDX3, Simple
Verification Result, Test Result and VULNS. None mentions an AI agent, an LLM
or a model. The two closest are SCAI and Runtime Traces, so both were read.

### SCAI Report (`https://in-toto.io/attestation/scai/v0.3`)

[scai.md](https://github.com/in-toto/attestation/blob/main/spec/predicates/scai.md)
defines `predicate.attributes[*]` with a required free-form `attribute`
string and optional `target`, `conditions` (any object) and `evidence`
(ResourceDescriptor), plus an optional `producer`. The spec says it "does not
dictate the types of attributes that can be attested", and that the formats of
`attribute` and `conditions` are "up to the producer and consumer".

We could encode our record as SCAI, for example
`{"attribute": "agent-work", "conditions": {"agent": {...}, "session": {...}}}`.
But then every field we care about lives inside `conditions`, whose shape SCAI
leaves to private agreement. A consumer could not find the agent name or the
policy digest from the SCAI type alone; it would still need our document. The
`predicateType` would no longer tell a verifier what the fields mean, which is
the job of a predicate type.

### Runtime Traces (`https://in-toto.io/attestation/runtime-trace/v0.2`)

[runtime-trace.md](https://github.com/in-toto/attestation/blob/main/spec/predicates/runtime-trace.md)
requires `monitor` (with a `type` URI), `monitoredProcess` (`hostID`, `type`,
`event`) and `monitorLog` with at least one of `process`, `network` or
`fileAccess`. Its model is an independent monitor observing a process, for
example an eBPF tracer watching a build.

Our record is the opposite: the claims are self-reported by whoever signs,
not observed by a monitor. Filling `monitor` with our collector would suggest
an observation guarantee we do not have, and the required `monitorLog` has no
natural content (we deliberately store no tool-call log). The agent, model,
instruction-file and policy fields would again be extensions with no meaning
defined by the Runtime Traces type.

## Decision

Define our own type,
`https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0`,
specified in [spec/agent-work-v0.md](../../spec/agent-work-v0.md). It follows
the in-toto parsing rules. It follows the in-toto versioning rules from `v1`
on; while in draft, every 0.x version shares the mutable `/v0` URI, a
deviation the spec states; the required `specVersion` field names the
exact draft so consumers can reject drafts they do not support. It uses the upstream field types
(Timestamp, DigestSet), and states its evidence level (`self-reported`) inside
the predicate.

## Consequences

- Verifiers need our spec to interpret the predicate. This is equally true for
  a SCAI or Runtime Traces encoding, so nothing is lost.
- The Type URI must stay resolvable. It is served from this repository's
  GitHub Pages, which is not live yet.
- Community predicates for agent provenance already exist outside the upstream
  index, for example agentattest's `https://agentattest.dev/predicate/v0`. We
  do not claim to be first. If upstream vets an agent predicate, or the
  community converges on one, we migrate to it and record that in the
  changelog.
- If SCAI later gains a defined vocabulary for agent work, this decision is
  revisited.
