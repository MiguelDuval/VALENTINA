# VALENTINA

Personal-use information discovery and lead-research system for finding, validating and organizing publicly available commercial opportunities relevant to a local business in Valencia.

## Purpose

VALENTINA collects permitted public information, preserves evidence and source provenance, reduces duplicate or irrelevant opportunities, and presents results for human review. It is an internal working tool, not a product or service for third parties.

## Architecture decision

**Huginn-first / reuse-first.** VALENTINA is built on the existing [`scrypster/huginn`](https://github.com/scrypster/huginn) local AI-agent platform. Huginn supplies the initial UI, agents, model configuration, MCP connectivity, routines, workflows and scheduling. VALENTINA adds only the smallest source-specific connectors and, when required by a real source proof-of-concept, one local MCP service with SQLite for normalized opportunities and evidence.

Do not confuse this with the unrelated Ruby project [`huginn/huginn`](https://github.com/huginn/huginn).

The binding architecture is in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md). The upstream audit and implementation gate are in [`docs/UPSTREAM-CANDIDATES.md`](docs/UPSTREAM-CANDIDATES.md).

## Project status

**Source-base stage.** The complete Huginn source is linked as a pinned Git submodule at `upstream/huginn` (release `v0.5.0`). Initialize it with `.\scripts\bootstrap.ps1` or `git submodule update --init --recursive`. The local runtime, NIM connection and source adapters are not yet configured.

## Mandatory rules

- [`AGENTS.md`](AGENTS.md) — mandatory instructions for every coding agent.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — binding architecture and delivery sequence.
- [`docs/SECURITY-AND-SECRETS.md`](docs/SECURITY-AND-SECRETS.md) — mandatory credential, privacy, local-runtime and CI policy.
- [`docs/UPSTREAM-CANDIDATES.md`](docs/UPSTREAM-CANDIDATES.md) — upstream audit, licensing and adoption decision.
- [`docs/UPSTREAM-INTEGRATION.md`](docs/UPSTREAM-INTEGRATION.md) — Huginn source submodule, pin, bootstrap and build instructions.
- [`docs/OPTIONAL-INTEGRATIONS.md`](docs/OPTIONAL-INTEGRATIONS.md) — NVIDIA NIM, Playwright MCP, UFO² and optional Brave Search.
- [`docs/DEPLOYMENT-AND-INSTALLER.md`](docs/DEPLOYMENT-AND-INSTALLER.md) — Windows-first standalone installer, runtime boundaries, administration UI and update/uninstall safety.
- [`docs/PRODUCT-AND-WORKFLOWS.md`](docs/PRODUCT-AND-WORKFLOWS.md) — user workflow, agent roles and deterministic-vs-AI boundaries.
- [`docs/DATA-AND-EVIDENCE.md`](docs/DATA-AND-EVIDENCE.md) — provisional company/opportunity/evidence contracts and deduplication.
- [`docs/SOURCE-REGISTRY.md`](docs/SOURCE-REGISTRY.md) — source approval, scope, collection limits and source onboarding.
- [`docs/QUALITY-AND-OPERATIONS.md`](docs/QUALITY-AND-OPERATIONS.md) — evaluation, quality metrics, failure handling and recovery.

## Starter kit and data contracts

The repository includes provisional JSON Schemas under `schemas/`, a private owner-profile placeholder, and safe Huginn agent/Skill/workflow templates under `templates/`. They are not activated automatically. See [`templates/huginn/README.md`](templates/huginn/README.md). The completed owner profile, live source registry, actual leads and local database must remain outside Git.

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

## Installation direction

The intended first user-facing package is a Windows x64 `VALENTINA-Setup.exe` installer. It will install the Go application with its web interface embedded; normal users should not need to install Go, Node.js, Python, Docker or a database server. Huginn supplies the initial local application/admin UI. Dedicated VALENTINA CRM screens are a separate domain feature and are not claimed to exist yet.

See [`docs/DEPLOYMENT-AND-INSTALLER.md`](docs/DEPLOYMENT-AND-INSTALLER.md) for the binding packaging decision.

## Initial implementation sequence

1. Initialize the pinned Huginn source submodule with the bootstrap script.
2. Build or install the pinned Huginn release using the upstream instructions.
3. Configure NVIDIA NIM through Huginn's generic OpenAI-compatible backend.
4. Add Playwright MCP only if an approved dynamic source requires browser access.
5. Add one small VALENTINA MCP + SQLite extension only if structured deduplication and review history require it.
6. Evaluate UFO² only for a proven GUI-only Windows desktop workflow.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for mandatory details and acceptance criteria.
