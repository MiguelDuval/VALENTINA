# VALENTINA

Personal-use information discovery and lead-research system for finding, validating and organizing publicly available commercial opportunities relevant to a local business in Valencia.

## Purpose

VALENTINA collects permitted public information, preserves evidence and source provenance, reduces duplicate or irrelevant opportunities, and presents results for human review. It is an internal working tool, not a product or service for third parties.

## Architecture decision

**Reuse-first, VALENTINA-branded.** VALENTINA uses the pinned [`scrypster/huginn`](https://github.com/scrypster/huginn) Go runtime and its existing Vue 3 browser client as the single local application foundation. The browser UI is customized as VALENTINA, keeps the useful agent/model/connection/workflow functions, and adds renovation-opportunity review screens. The intended local product is one application process, not a second CRM website or companion app.

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
- [`docs/MVP-DELIVERY-ROADMAP.md`](docs/MVP-DELIVERY-ROADMAP.md) — prioritized path to the VALENTINA-branded browser UI and Windows installer.
- [`docs/ADR-005-VALENTINA-PERSONAL-WHITE-LABEL-UI.md`](docs/ADR-005-VALENTINA-PERSONAL-WHITE-LABEL-UI.md) — approved white-label and single-runtime decision.

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

The intended first user-facing package is a Windows x64 `VALENTINA-Setup-x64.exe` installer. It packages one local Go application with the customized Vue frontend embedded. The normal interface is VALENTINA; engine administration remains available inside its AI & Automation / Advanced areas. Normal users should not need to install Go, Node.js, Python, Docker or a database server.

See [`docs/DEPLOYMENT-AND-INSTALLER.md`](docs/DEPLOYMENT-AND-INSTALLER.md) for the binding packaging decision.

## Initial implementation sequence

1. Capture the minimum local owner business profile.
2. Build and verify the pinned Huginn runtime on Windows 11 x64.
3. Build the tracked VALENTINA white-label patch set against the pinned Vue client and make Overview the default route.
4. Add the renovation opportunity screens and the smallest in-process Go API/SQLite/MCP domain extension; then connect the UI to real records.
5. Implement one approved real source and one manual end-to-end research workflow.
6. Package Huginn and VALENTINA behind one per-user Windows NSIS installer.
7. Verify install, first-run, source workflow, persistence, backup/restore, upgrade and data-preserving uninstall.

See [`docs/MVP-DELIVERY-ROADMAP.md`](docs/MVP-DELIVERY-ROADMAP.md), which supersedes older UI-last/UFO²-first planning.


## Agent collaboration and delivery

- docs/AGENT-COLLABORATION.md — shared execution/review protocol for Cline and ChatGPT.
- docs/AGENT-HANDOFF.md — current task handoff and evidence record.
- docs/WORK-QUEUE.md — prioritized task queue and dependencies.
- .clinerules/00-valentina-project.md — persistent Cline workspace rules.
- scripts/publish-branch.ps1 — guarded task-completion push; never force-pushes or publishes main.

Eligible pushes automatically run the Windows test build. A successful run exposes a 30-day SHA-labelled artifact ZIP containing VALENTINA.exe. This workflow does not create GitHub Releases, build the NSIS installer, or automatically invoke ChatGPT; independent review is requested through the shared handoff and PR review process.
