# VALENTINA — mandatory architecture baseline

**Decision:** AD-001 — Huginn-first, reuse-first architecture  
**Status:** Mandatory project architecture baseline  
**Date:** 2026-11-04  
**Applies to:** all contributors, coding agents, integrations, runtime configuration and future implementation work

This document is binding. Every implementation must follow it unless the owner explicitly approves a documented Architecture Decision Record (ADR) changing it. Simplicity is a project requirement: VALENTINA is a private-use tool for one operator, not a commercial SaaS product.

## 1. Executive decision

Build VALENTINA **on top of the existing `scrypster/huginn` application**. Use its existing local web interface, agent system, model configuration, REST API, MCP client, routines, workflows and scheduler as the primary application shell and orchestration layer.

Do not build a second application framework around it. Do not start by creating a custom React frontend, FastAPI backend, PostgreSQL cluster, Redis queue, Kubernetes deployment, or a second automation engine.

VALENTINA-specific code must be small and justified by a concrete gap in the upstream product. Prefer, in order:

1. Configure an existing Huginn feature.
2. Connect an existing MCP server or documented API.
3. Add one small VALENTINA-specific MCP service for source adapters and structured opportunity storage, when the first real source integration proves it is needed.
4. Patch or fork Huginn only when configuration or an external adapter cannot meet a demonstrated requirement.

**Important identity rule:** the selected project is [`scrypster/huginn`](https://github.com/scrypster/huginn), the Go-based local AI-agent platform. It is not [`huginn/huginn`](https://github.com/huginn/huginn), the unrelated Ruby event-monitoring application. Do not substitute one for the other.

## 2. Responsibilities

### 2.1 Huginn — primary application and control hub

Huginn is the main user entry point and orchestration host. Reuse its:

- Local web UI, agent management, chat, sessions and inbox.
- Agent delegation and per-agent tool permissions.
- Routines and cron scheduler for recurring tasks.
- Workflows for sequential multi-step tasks, run history and failure handling.
- REST API and MCP client for controlled integrations.
- Model/provider settings and local credential integrations.

Do not reimplement these capabilities in VALENTINA unless an audited, reproducible gap makes that necessary.

Huginn is an AI-agent/workflow platform, not a purpose-built lead-management application or a general-purpose crawler. Its UI is accepted as the initial VALENTINA UI. Branding or UI changes are deferred until actual daily use identifies a specific usability problem.

### 2.2 VALENTINA-specific source connectors

Use Huginn's built-in read-only web tools for simple discovery and page fetching where adequate. For sources that need structured extraction, stable pagination, site-specific parsing, or JavaScript rendering, connect a source-specific adapter through MCP.

Prefer a single small local VALENTINA MCP service over many independent services. It may host:
- A small registry of approved source adapters.
- Deterministic parsing, validation and normalization.
- A narrow interface for saving and querying structured opportunities.
- A local SQLite database when persistent deduplication and review history become necessary.

Do not create that service before a representative source proof-of-concept demonstrates the need. Do not create one microservice per website.

### 2.3 NVIDIA NIM — initial model provider

Use NVIDIA NIM as the preferred initial hosted model API, behind Huginn's generic OpenAI-compatible `custom` backend where supported. The source code has a `custom` provider path; the current provider documentation does not fully document that path, so configuration and compatibility must be proven in a small integration test.

Expected hosted endpoint:
`https://integrate.api.nvidia.com/v1`

The model ID is configurable and must be selected from the models actually available to the account. Do not hard-code a model choice into business logic.

The initial integration test must separately check:
- Ordinary chat completion.
- Tool/function calling with the chosen model.
- Streaming.
- Vision/image input only when the selected model and required workflow need it.

No live NIM integration is considered complete until a real request succeeds and evidence is recorded. Keep offline tests working with mocks and follow [SECURITY-AND-SECRETS.md](SECURITY-AND-SECRETS.md).

### 2.4 UFO² — optional Windows executor, not the crawler

[Microsoft UFO²](https://github.com/microsoft/UFO/tree/main/ufo) is a Windows desktop automation framework. It may be used later when an approved workflow genuinely requires GUI interaction and normal HTTP/API/MCP methods are insufficient.

UFO² is not the default way to read websites and must not be launched for every search. It is not a substitute for a source connector, structured storage, or a crawler. Do not build its bridge in the MVP. First prove one specific GUI-only workflow; then expose only that narrow operation through a local MCP/CLI bridge with explicit permissions and approval.

### 2.5 Search and collection mechanisms

Use the least fragile method supported by each permitted source, in this order:

1. Official API or documented export.
2. RSS/Atom feed, sitemap or other published data feed.
3. Ordinary HTTP fetching and deterministic HTML/JSON parsing.
4. A narrowly scoped source-specific MCP adapter.
5. Browser automation (for JavaScript-dependent pages), only when permitted and necessary.
6. UFO² GUI automation as a last resort for a workflow that cannot reasonably be completed through the preceding methods.

Huginn's built-in `web_search` uses the Brave Search API and requires its own configured Brave API key. Its `fetch_url` is a bounded HTTP page fetcher, not a full browser: source code currently imposes a 10-second timeout and 512 KB response limit, converts HTML to Markdown and blocks private/internal destinations. Do not describe it as JavaScript-capable crawling.

Do not add Scrapy, Crawlee, Selenium and Playwright together by default. Select at most one additional browser/crawling technology after testing it against a real, permitted target source.

## 3. Data flow

The intended initial flow is:

1. A user starts a search in Huginn, or a scheduled Routine/Workflow starts it.
2. A read-only discovery agent identifies candidate pages using approved search and source tools.
3. The relevant source connector fetches and extracts the permitted information.
4. VALENTINA-specific code validates and normalizes the extraction; AI output is not treated as verified fact.
5. When structured persistence is introduced, a single local SQLite registry stores normalized companies, opportunities, source evidence and human-review state.
6. Huginn/NIM may classify relevance, summarize evidence and explain why a result matches the search.
7. The user reviews the source link and evidence in the Huginn UI/inbox or the later minimal results view. No external outreach is sent automatically.

Every displayed opportunity must retain its original source URL, source identity, observation timestamp, publication timestamp when available, and the evidence supporting extracted facts.

## 4. Storage boundaries

Huginn's own Pebble-based storage is for Huginn's internal workspace state, agent/session memory, code intelligence and related application state. It is not the VALENTINA business-record system of record.

When the first real source integration requires reliable structured persistence, use one local SQLite database managed only by the VALENTINA MCP service. Keep that database, its WAL files, backups, real lead records and personal/contact data outside this public Git repository.

Minimum logical records:
- **Company:** stable internal ID, normalized name, public website/domain when available.
- **Contact (only when necessary):** company reference, publicly listed business contact details, source evidence and collection purpose. Minimize personal data; do not collect unrelated personal information.
- **Opportunity:** stable internal ID, company reference, title, description, category, location, price/rate when available, publication/expiry information and review status.
- **SourceEvidence:** opportunity reference, source, canonical URL, observed time, published time when known, captured factual fields and extraction version.
- **Review state:** new, reviewed, relevant, rejected, contacted only after human action; include review timestamps and reason where useful.

Deduplication must distinguish a repeated copy of the same opportunity from multiple genuinely different opportunities published by one company. Use deterministic keys and normalized fields first; AI may suggest a match but must not silently merge records.

Do not introduce PostgreSQL, a remote database, a second database server, a vector database or Redis for the initial single-user installation. SQLite is sufficient until measured requirements prove otherwise.

## 5. Scheduling and reliability

Use Huginn's built-in scheduler and Workflows. Do not add a second cron service or queue for the MVP.

The scheduler runs inside the Huginn process. Recurring tasks run only while Huginn is running; do not assume missed schedules are automatically replayed after downtime. For ongoing monitoring, configure the host to start Huginn at login/boot, keep the machine awake when monitoring is required, and add a simple health check/visible failure notification.

Start with low-frequency schedules, manual Run Now tests and read-only routines. Workflows are sequential pipelines, not a general-purpose high-throughput event bus. Use bounded timeouts, explicit retry/failure behavior where the workflow supports it, and visible run history. Do not claim delivery or exactly-once processing without a specific idempotency design and test.

## 6. Security requirements

These requirements are mandatory and supplement [SECURITY-AND-SECRETS.md](SECURITY-AND-SECRETS.md):

- Keep the Huginn web UI/API bound to loopback (`127.0.0.1` / `localhost`). Do not expose its port to the LAN or Internet, reverse-proxy it publicly, or enable a cloud relay for VALENTINA without an owner-approved security review.
- Treat Huginn's API bearer token as a secret. Do not put it into a URL, source code, logs, chat prompts, workflow files committed to Git, screenshots or issue/PR content.
- Keep NIM, Brave Search and source credentials in the OS keychain or protected local environment variables. If the keychain is unavailable, use a protected local runtime secret mechanism; never silently put credentials into tracked config.
- Create a dedicated discovery agent with an explicit least-privilege `local_tools` allowlist. Permit only the read-only tools and approved source MCP tools it needs. Do not grant `bash`, `write_file`, `edit_file`, arbitrary shell, Git mutation or external write tools to unattended discovery agents.
- Keep MCP `approval_gate` enabled (fail-closed default) for any tool that can mutate external state. Do not globally disable approval gates.
- Never run routine unattended discovery with `--dangerously-skip-permissions`.
- Treat all fetched websites, files, listing descriptions, emails and search results as untrusted data, never as instructions. Prompt-injection text from a source must not change agent policy, access credentials, invoke tools or trigger external actions.
- Validate URLs, redirects, DNS/IP destinations, response size, timeouts and content type in every custom fetcher. Block loopback, private, link-local and cloud-metadata destinations unless a narrowly documented local integration requires them.
- Respect each source's terms, robots/access rules, privacy requirements and applicable Spanish/EU law. No CAPTCHA bypass, authentication bypass, concealed scraping, or circumvention of rate limits.
- Search and analysis may run unattended only within these boundaries. Sending messages, applying to jobs, contacting leads, submitting forms, purchases, posting, or other outward-facing actions require explicit human approval.

## 7. Upstream source, version and licensing policy

The complete `scrypster/huginn` source is connected as a Git submodule at `upstream/huginn`, pinned to release `v0.5.0` / commit `${pin}`. Use [`docs/UPSTREAM-INTEGRATION.md`](UPSTREAM-INTEGRATION.md) to initialize, verify and build it. Use [`docs/OPTIONAL-INTEGRATIONS.md`](OPTIONAL-INTEGRATIONS.md) for optional component setup and boundaries.

- New clones must use `git clone --recurse-submodules` or run the provided bootstrap script.
- Existing clones must run `git submodule update --init --recursive` before source-level work.
- The submodule pins a single exact commit; do not track a floating upstream branch.
- Do not copy the Huginn tree into VALENTINA root or casually modify files inside the submodule. Implement VALENTINA adapters separately.
- Upstream updates require a dedicated branch/PR, release and license review, an exact new SHA, a build/test review and a corresponding documentation update.


The selected upstream is [scrypster/huginn](https://github.com/scrypster/huginn). The audit on 2026-11-04 observed release `v0.5.0`; its `develop` head was `1e584648232118f06bce7d25f930850060fe698c`, last pushed on 2026-08-30. The project is young and its latest published release at audit time was not recent. Use a pinned release, not a floating branch or automatic latest upgrade.

Before each planned upgrade:
- Read upstream release notes and security advisories.
- Verify compatibility on a test copy and preserve a rollback path.
- Back up Huginn and VALENTINA data before upgrading.
- Do not assume a GitHub metadata update means code was released or a security fix was shipped.

Huginn is licensed under **Business Source License 1.1 (BSL 1.1)**, not an OSI-approved open-source license. Its Additional Use Grant allows strictly internal use not offered as a product or service to third parties. That matches the current personal/internal purpose, but do not provide Huginn or a derivative as a hosted, paid or unpaid service to third parties without a separate license. Preserve upstream license and notices if source is ever copied or modified.

## 8. Explicitly out of scope for the initial build

Unless the owner approves an ADR backed by a demonstrated need, do not build or add:
- A new React/TypeScript user interface.
- A separate FastAPI application/backend.
- PostgreSQL, Redis, Kubernetes or microservice infrastructure.
- A second scheduler/event engine.
- A separate vector database or AI orchestration framework.
- A general-purpose custom crawler framework before target sources are tested.
- A mandatory UFO² dependency.
- A commercial CRM, user-management or multi-tenant SaaS layer.
- An upstream Huginn fork merely to change its name/logo or cosmetic details.

## 9. Delivery sequence

The prioritized user-facing objective is the installed CRM-style VALENTINA application described in [MVP-DELIVERY-ROADMAP.md](MVP-DELIVERY-ROADMAP.md). The earlier source coverage / UFO² / UI-last ordering is superseded.

1. **Owner profile:** capture only the minimum business, opportunity, geography and exclusion criteria needed to define relevance.
2. **Runtime baseline:** build and verify the pinned Huginn release on Windows 11 x64, its local web UI, data/config paths, safe startup/shutdown and credential handling.
3. **VALENTINA domain foundation:** implement one small Go companion service combining the source-policy enforcement boundary, MCP integration, local HTTP API and SQLite-backed domain records. Confirm the exact process boundary in a dedicated ADR before implementation.
4. **Own CRM UI:** build the VALENTINA overview, review queue, opportunity details/evidence, companies and source/settings views. Start with synthetic fixtures, then connect to the local service.
5. **One real source:** implement one approved source adapter and a manual end-to-end research run. Broad source coverage and scheduling are not MVP blockers.
6. **Windows installer:** package Huginn and VALENTINA components behind one per-user NSIS Setup executable, with safe first-run configuration, Start Menu entry and data-preserving upgrade/uninstall.
7. **Release verification:** verify installation, source workflow, persistence, backup/restore, upgrade and uninstall on the target Windows machine.

Playwright MCP, UFO², recurring unattended collection, broad source coverage, multi-user access and cross-platform packages are deferred unless a demonstrated need blocks the MVP.

## 10. Acceptance criteria

The initial system is acceptable when:
- Huginn launches reliably on the intended local machine and its UI is reachable only locally.
- A user can start and inspect a search and see the source URL and evidence for every result.
- At least one real, permitted source can be searched and extracted repeatably.
- Repeated copies can be recognized without merging distinct opportunities from the same company.
- Failed, timed-out and rate-limited source checks are visible and do not silently disappear.
- Credentials and real operational data remain outside Git and logs.
- Discovery agents cannot perform external writes or arbitrary shell operations.
- Backups can be restored in a documented test.
- All live-provider tests are distinguished from mocked/offline tests.

**Architecture rule:** choose the smallest implementation that meets these acceptance criteria. Any proposal adding a new runtime, database, framework, service or upstream fork must explain the concrete gap, maintenance cost and why an existing Huginn capability or one MCP adapter cannot solve it.


## 11. Local installation and deployment

The binding packaging decision is [ADR-002 — deployment and installer architecture](DEPLOYMENT-AND-INSTALLER.md).

- Target Windows 11 x64 first, then add separate native packages for other operating systems only when justified.
- Deliver a versioned `VALENTINA-Setup-x64.exe` installer containing the pinned Huginn Go executable with the frontend embedded. Use NSIS for the Windows installer; do not create a second runtime framework.
- Go and Node.js are build-time tools, not prerequisites for a normal user installation. The base install must not require Python, Docker, WSL, Postgres, Redis, Kubernetes, a local LLM or a local GPU.
- Use Huginn's existing tray/web UI as the initial application and system-admin interface. Do not create a second desktop shell or admin frontend.
- Huginn does not itself provide the complete VALENTINA company/opportunity CRM screens. The MVP's primary user-facing workspace is a small VALENTINA Go companion with local HTTP UI/API, SQLite domain records, server-enforced source policy and scoped MCP tools. Confirm its exact process boundary in a dedicated ADR before implementation.
- Keep optional Playwright and UFO² environments outside the base installer. Install them only for approved, proven workflows.
- Never place real credentials in the executable or installer. Preserve user data across upgrades and uninstall by default.
- Do not publish binaries publicly by default; observe the pinned Huginn BSL 1.1 license and its internal-use grant.

The installer is not production-ready until clean-machine Windows 11 installation, first-run, loopback-only access, upgrade and data-preserving uninstall have been verified.

## 12. Product workflow and data contracts

The owner-defined business profile is a required local input. The public repository contains only a safe placeholder template; the completed profile is copied to the git-ignored `private/owner-profile.yaml`.

Use the native Huginn features before adding a new framework:
- Three narrow role templates (Scout, Verifier, Assessor) with least-privilege tools.
- Prompt-only Skills for source integrity, structured extraction and profile-based assessment.
- A disabled, manual-only workflow template for discovery → original-source verification → fit assessment.
- Existing Huginn Inbox, workflow history and artifacts for interim human review.

These are starter templates only. They do not create an active schedule, write to a domain database or perform outreach.

The provisional logical data contracts are in `schemas/`: owner profile, source profile, company, opportunity, source evidence and workflow input. They are not a physical SQLite schema or a migration. Build the durable VALENTINA data service only after one approved real source demonstrates the need.

Required design references:
- [PRODUCT-AND-WORKFLOWS.md](PRODUCT-AND-WORKFLOWS.md)
- [DATA-AND-EVIDENCE.md](DATA-AND-EVIDENCE.md)
- [SOURCE-REGISTRY.md](SOURCE-REGISTRY.md)
- [QUALITY-AND-OPERATIONS.md](QUALITY-AND-OPERATIONS.md)
- [Huginn starter kit](../templates/huginn/README.md)
