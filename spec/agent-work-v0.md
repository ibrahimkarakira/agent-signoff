# Predicate type: Agent Work

Type URI: https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0

Version: 0.1.0 (draft). Producers write this value in `predicate.specVersion`.

Predicate Name: Agent Work

Status: draft. The `/v0` Type URI is a mutable draft URI: every 0.x version
of this predicate uses it, and a 0.x version may change in incompatible ways
without a new Type URI. This deviates from upstream in-toto versioning (see
[Parsing rules](#parsing-rules)). The URI does not resolve yet; it is planned
to be served from GitHub Pages.

This document follows the in-toto
[predicate template](https://github.com/in-toto/attestation/blob/main/spec/predicates/template/template.md).
The key words MUST, MUST NOT, SHOULD and MAY are used as described in
[RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

## Purpose

Record, in a signed and standard form, that a git commit was produced in a
session with an AI coding agent: which agent and model, which instruction
files and policy file were in force, and a digest of the session transcript.
The record is a claim made by whoever signs it. It does not prove that the
agent wrote the code (see [What this predicate does not prove](#what-this-predicate-does-not-prove)).

## Use cases

1. A pull request check verifies that every commit marked as agent-assisted
   carries an attestation, signed by an allowed signer, naming an allowed
   agent, and recording the digest of the policy file at the PR base.
2. An auditor verifies the signature offline with stock `cosign`, then reads
   the agent, model and instruction-file digests from the predicate.

No vetted in-toto predicate carries agent, model or instruction-file fields.
Why SCAI and Runtime Traces were not extended is recorded in
[ADR 0001](../docs/decisions/0001-own-predicate.md).

## Prerequisites

The [in-toto Attestation Framework](https://github.com/in-toto/attestation/tree/main/spec/v1),
Statement v1. Attestations are carried in a DSSE envelope with payload type
`application/vnd.in-toto+json` inside a
[Sigstore bundle](https://docs.sigstore.dev/about/bundle/).

## Model

The functionary is the process that runs `signoff attest` (or
`cosign attest-blob`) after an agent session ends: a developer machine or a
CI job. The step is "commit produced with agent assistance". The signer's
identity is not part of the predicate. It comes from the signing key or
certificate and is checked by the consumer against its own policy.

## Schema

```jsonc
{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [{ "name": "<git URL>@<commit>", "digest": { "sha256": "...", "gitCommit": "..." } }],
  "predicateType": "https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0",
  "predicate": {
    "specVersion": "0.1.0",
    "agent":    { "name": "<string>", "version": "<string>", "model": "<string>", "invocation": "<string>" },
    "session":  { "id": "<string>", "startedAt": "<Timestamp>", "endedAt": "<Timestamp>",
                  "transcriptDigest": { /* DigestSet */ }, "transcriptStored": false },
    "inputs":   { "baseCommit": "<string>",
                  "instructionFiles": [{ "path": "<string>", "digest": { /* DigestSet */ } }] },
    "policy":   { "path": "<string>", "digest": { /* DigestSet */ } },
    "evidence": { "level": "self-reported", "collector": { "name": "<string>", "version": "<string>" } }
  }
}
```

### Parsing rules

This predicate follows the in-toto
[parsing rules](https://github.com/in-toto/attestation/blob/main/spec/v1/README.md#parsing-rules):

- Consumers MUST ignore unrecognised fields.
- Producers MAY add extension fields with names unlikely to collide with other
  producers. An extension field MUST NOT change the meaning of any field
  defined here.
- Optional fields MAY be absent or null; both are treated the same.

Versioning deviates from upstream while the predicate is a draft. The in-toto
[versioning rules](https://github.com/in-toto/attestation/blob/main/spec/versioning.md)
treat each 0.X version as a major version with its own Type URI (upstream
uses, for example, `scai/v0.3` and `runtime-trace/v0.2`). This project does
not: all 0.x drafts share the `/v0` URI, and a breaking change between 0.x
drafts does not change it. So every predicate carries `specVersion`, the
exact draft it was produced against. Consumers MUST reject an attestation
whose `specVersion` is missing or is a version they do not support, and MUST
NOT read a predicate under the rules of a draft other than the one its
`specVersion` names. From `v1` on, the upstream rules apply
unchanged: the Type URI carries the major version and changes on any
backwards incompatible change, and minor versions only add optional fields
whose absence has no meaning.

Rules specific to this predicate:

- Consumers MUST rely on `predicateType` in the Statement, not on the DSSE
  payload type.
- Producers MUST NOT put raw prompts, tool output or transcript text in the
  predicate. Only digests are recorded.
- Consumers MUST treat every field as a claim by the signer, not as an
  observation by a trusted monitor (`evidence.level` says so).
- Consumers MUST reject a v0 attestation whose `predicate.evidence` is
  missing or whose `predicate.evidence.level` is anything other than
  `self-reported`. The signer controls this field, so a consumer never reads
  a stronger level from it.
- Every `path` field (`predicate.inputs.instructionFiles[*].path` and
  `predicate.policy.path`) MUST be a clean relative POSIX path: non-empty,
  segments separated by `/`, no leading `/`, no empty, `.` or `..` segment,
  no backslash, no drive letter or other `:`, and no NUL or other control
  character. Consumers MUST reject an attestation with any path that breaks
  this rule before touching the filesystem.
- A clean path can still name a symlink that points outside the repository
  (`CLAUDE.md -> ../../home/user/.ssh/id_ed25519`). A consumer that checks a
  digest against a path MUST read the file from a git tree, never from a
  working tree, so no link is ever followed. The tree is fixed per field:
  - `predicate.inputs.instructionFiles[*]`: the tree of
    `predicate.inputs.baseCommit`, the commit the session started from.
    Without `baseCommit`, instruction-file digests MUST NOT be reported as
    verified.
  - `predicate.policy`: the tree of the PR base, at the path from the
    consumer's own configuration (see the next rule).
- To read a path, a consumer MUST first list its tree entry with
  `git ls-tree <commit> -- <path>`, which prints the mode, type and object
  ID. It MUST reject the attestation unless exactly one entry is listed with
  mode `100644` or `100755` and type `blob`: a symlink (`120000`), a
  submodule (`160000`), a directory or a missing entry is rejected. A symlinked
  directory anywhere in the path is one `120000` entry in its parent tree, so
  the path below it lists as missing. The consumer MUST then hash the raw
  bytes of that blob, read by the object ID from the same listing (for example
  `git cat-file blob <object ID>`), with no checkout filters (`autocrlf`,
  `.gitattributes`, LFS) applied. `git cat-file -p <commit>:<path>` alone is
  not enough: it prints a symlink's target text and no mode.
- Consumers SHOULD set `GIT_NO_REPLACE_OBJECTS=1` while reading, so
  `refs/replace` cannot substitute the commit, tree or blob.
- Producers MUST record the digest of the raw git blob bytes, and MUST NOT
  record a path whose tree entry is not a regular file.
- A consumer that checks the policy digest MUST locate the policy file from
  its own configuration, never from `predicate.policy.path`. It MUST reject
  the attestation when `predicate.policy.path` differs from the configured
  path. Otherwise `predicate.policy.path` is informational only.
- A consumer whose configuration requires a policy digest MUST reject an
  attestation that has no `predicate.policy`.

### Statement-layer constraints

- `subject` MUST contain exactly one entry, the commit.
- `subject[0].digest.sha256` MUST be the SHA-256 of the raw commit object as
  printed by `git cat-file commit <commit>`, so the digest is reproducible with
  git and `sha256sum` alone.
- `subject[0].digest.gitCommit` SHOULD be the commit ID. `gitCommit` is a
  [DigestSet key defined upstream](https://github.com/in-toto/attestation/blob/main/spec/v1/digest_set.md).
  Whether stock `cosign verify-blob-attestation` accepts a subject with two
  digest keys is `[UNCLEAR]`; if it does not, producers will emit `sha256`
  only.
- `subject[0].name` SHOULD be `git+<repository URL>@<commit>`.
- Consumers MUST bind the attestation to a commit by `sha256` alone,
  recomputed from `git cat-file commit <commit>`. They MUST NOT accept a match
  on `gitCommit` alone, so the upstream rule that two DigestSets match on any
  shared algorithm does not apply to this subject.
- When `gitCommit` is present, consumers MUST check that it equals the ID of
  the commit whose object hashes to `sha256`, and MUST reject the attestation
  on mismatch.

### Fields

Types: Timestamp and DigestSet are the in-toto
[field types](https://github.com/in-toto/attestation/blob/main/spec/v1/field_types.md)
(a Timestamp is RFC 3339 in UTC, ending in `Z`). A digest value is lowercase hex.

| Field | Type | Required | Example | Meaning |
|---|---|---|---|---|
| `predicate.specVersion` | string | required | `0.1.0` | The draft of this document the producer followed. Consumers reject versions they do not support; see [Parsing rules](#parsing-rules). |
| `predicate.agent` | object | required | see below | The agent that worked in the session. |
| `predicate.agent.name` | string | required | `claude-code` | Agent identifier, lowercase. Policies match on this value. |
| `predicate.agent.version` | string | optional | `9.9.9` | Agent version as the agent reports it. |
| `predicate.agent.model` | string | optional | `example-model-id` | Model identifier as the agent reports it. Absent when the agent does not expose it. |
| `predicate.agent.invocation` | string | optional | `interactive` | `interactive` (a person drove the session) or `headless` (non-interactive run). Consumers MUST NOT reject other values. |
| `predicate.session` | object | required | see below | The work session. |
| `predicate.session.id` | string | required | `7d2f0c1e-0000-4000-8000-000000000001` | Session identifier from the agent, or a random UUID. |
| `predicate.session.startedAt` | Timestamp | optional | `2026-11-05T09:12:03Z` | When the session started. |
| `predicate.session.endedAt` | Timestamp | optional | `2026-11-05T09:47:55Z` | When the session ended. MUST NOT be earlier than `startedAt` when both are set. |
| `predicate.session.transcriptDigest` | DigestSet | optional | `{"sha256": "e3b0..."}` | Digest of the session transcript file as the collector read it. |
| `predicate.session.transcriptStored` | boolean | optional | `false` | Whether the transcript was stored anywhere reachable by the consumer. Absent means `false`. |
| `predicate.inputs` | object | optional | see below | What the agent was given. |
| `predicate.inputs.baseCommit` | string | optional | `1a2b3c4d...` | Commit ID the session started from. |
| `predicate.inputs.instructionFiles` | array of objects | optional | see below | Instruction files in force, for example `CLAUDE.md` or `AGENTS.md`. |
| `predicate.inputs.instructionFiles[*].path` | string | required | `CLAUDE.md` | Path relative to the repository root; see the path rule in [Parsing rules](#parsing-rules). |
| `predicate.inputs.instructionFiles[*].digest` | DigestSet | required | `{"sha256": "0f34..."}` | Digest of the raw git blob bytes at `baseCommit`; see the parsing rules. |
| `predicate.policy` | object | optional | see below | The agent-signoff policy file in force. |
| `predicate.policy.path` | string | required | `signoff.yml` | Path relative to the repository root; see the path rule in [Parsing rules](#parsing-rules). Informational: consumers never open the policy file through this value. |
| `predicate.policy.digest` | DigestSet | required | `{"sha256": "a1b2..."}` | Digest of the policy file. A consumer that requires a policy digest compares it with the policy file at the PR base, at the path from the consumer's own configuration. |
| `predicate.evidence` | object | required | see below | How strong the claims are. |
| `predicate.evidence.level` | string | required | `self-reported` | In v0 this MUST be `self-reported`: the signer asserts the claims, nothing observed them independently. Consumers MUST reject any other value. |
| `predicate.evidence.collector` | object | required | see below | The tool that gathered the claims. |
| `predicate.evidence.collector.name` | string | required | `signoff` | Collector name. |
| `predicate.evidence.collector.version` | string | optional | `0.1.0` | Collector version. |

"Required" inside an optional object applies only when that object is present.

## Example

All values are synthetic.

```json
{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [
    {
      "name": "git+https://github.com/example-org/example-repo@4f9c2d1e8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d",
      "digest": {
        "sha256": "9b74c9897bac770ffc029102a200c5de0a4d5e3a1f0b7c8d9e6f5a4b3c2d1e0f",
        "gitCommit": "4f9c2d1e8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3d"
      }
    }
  ],
  "predicateType": "https://ibrahimkarakira.github.io/agent-signoff/predicate/agent-work/v0",
  "predicate": {
    "specVersion": "0.1.0",
    "agent": {
      "name": "claude-code",
      "version": "9.9.9",
      "model": "example-model-id",
      "invocation": "interactive"
    },
    "session": {
      "id": "7d2f0c1e-0000-4000-8000-000000000001",
      "startedAt": "2026-11-05T09:12:03Z",
      "endedAt": "2026-11-05T09:47:55Z",
      "transcriptDigest": { "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" },
      "transcriptStored": false
    },
    "inputs": {
      "baseCommit": "1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d",
      "instructionFiles": [
        { "path": "CLAUDE.md", "digest": { "sha256": "0f343b0931126a20f133d67c2b018a3b4d1e5f6a7b8c9d0e1f2a3b4c5d6e7f80" } }
      ]
    },
    "policy": {
      "path": "signoff.yml",
      "digest": { "sha256": "a1b2c3d4e5f60718293a4b5c6d7e8f9001122334455667788990aabbccddeeff" }
    },
    "evidence": {
      "level": "self-reported",
      "collector": { "name": "signoff", "version": "0.1.0" }
    }
  }
}
```

## What this predicate does not prove

- That the named agent or model wrote the code. The signer makes the claim.
- That the transcript says what happened. The digest only shows the signer
  had a file with that hash.
- Who reviewed the change. Review happens after the work, so reviewer data is
  deliberately absent here; a separate `agent-review` predicate is planned.

## Non-goals

- Signer identity inside the predicate (it belongs to the signature layer).
- Per-line authorship, prompt content, tool-call logs or model behaviour scores.
- A policy language. The policy file is referenced by digest only.

## Changelog and migrations

- 0.1.0 (draft): first version.
