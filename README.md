# VALENTINA

Personal-use information discovery and lead-research system for finding, validating and organizing publicly available commercial opportunities relevant to a local business in Valencia.

## Purpose

VALENTINA collects permitted public information, preserves evidence and source provenance, reduces duplicate or irrelevant opportunities, and presents results for human review. It is an internal working tool, not a product or service for third parties.

## Architecture decision

**Huginn-first / reuse-first.** VALENTINA is built on the existing [`scrypster/huginn`](https://github.com/scrypster/huginn) local AI-agent platform. Huginn supplies the initial UI, agents, model configuration, MCP connectivity, routines, workflows and scheduling. VALENTINA adds only the smallest source-specific connectors and, when required by a real source proof-of-concept, one local MCP service with SQLite for normalized opportunities and evidence.

Do not confuse this with the unrelated Ruby project [`huginn/huginn`](https://github.com/huginn/huginn).

The binding architecture is in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md). The upstream audit and implementation gate are in [`docs/UPSTREAM-CANDIDATES.md`](docs/UPSTREAM-CANDIDATES.md).

## Project status

**Documentation and proof-of-concept stage.** The architecture baseline is selected, but Huginn has not yet been installed and validated on the target machine. NVIDIA NIM, one permitted source connector, and structured storage still require explicit runtime tests before implementation is considered complete.

## Mandatory rules

- [`AGENTS.md`](AGENTS.md) — mandatory instructions for every coding agent.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — binding architecture and delivery sequence.
- [`docs/SECURITY-AND-SECRETS.md`](docs/SECURITY-AND-SECRETS.md) — mandatory credential, privacy, local-runtime and CI policy.
- [`docs/UPSTREAM-CANDIDATES.md`](docs/UPSTREAM-CANDIDATES.md) — upstream audit, licensing and adoption decision.

## Working principles

- Prefer existing Huginn features, documented APIs and MCP integrations over new frameworks.
- Preserve source URLs, timestamps and evidence for every discovered opportunity.
- Separate companies, contacts and individual opportunities.
- Deduplicate without collapsing genuinely different opportunities from the same company.
- Keep a human in the loop before any commercial outreach or external action.
- Respect source terms, privacy, and applicable Spanish and EU law.
- Keep credentials, operational databases, real leads and private runtime configuration outside this public Git repository.
- Keep model providers replaceable; NVIDIA NIM is the preferred initial provider, not a permanent lock-in.

## Repository visibility

This repository is intentionally **public**. All committed code, documentation, workflow definitions, issues, pull requests and Git history may be read by anyone. Do not commit secrets, actual commercial leads, personal/contact records, local databases, exports, logs, backups or confidential operational data. Do not change repository visibility without an explicit owner instruction.

## Initial implementation sequence

1. Install and validate a pinned Huginn release locally.
2. Configure and test NVIDIA NIM through Huginn's generic OpenAI-compatible backend.
3. Prove one real, permitted information source end-to-end.
4. Add one small VALENTINA MCP + SQLite extension only if the proof-of-concept requires structured deduplication and review history.
5. Add further sources one at a time. Evaluate UFO² only for a proven GUI-only Windows workflow.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for mandatory details and acceptance criteria.
