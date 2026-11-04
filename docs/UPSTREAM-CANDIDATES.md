# Upstream component evaluation

**Date:** 2026-11-04  
**Repository:** [MiguelDuval/VALENTINA](https://github.com/MiguelDuval/VALENTINA)  
**Status:** Initial technical assessment; no upstream project has been forked or vendored.

## Executive decision

Build VALENTINA as the product-specific application and data system. Reuse upstream software as separately installed services or normal dependencies wherever possible. Do not copy several complete applications into this repository.

Initial direction:

1. Use **scrypster/huginn** as an optional agent orchestration service, connected to VALENTINA through its supported API and/or MCP. Do not fork it for the MVP.
2. Use **NVIDIA NIM** as a model inference endpoint. NIM is an API/runtime service, not source code to fork.
3. Keep **Microsoft UFO²** as an optional Windows desktop executor. First test an external invocation/bridge; do not fork it unless a concrete required change cannot be made through configuration or an integration layer.
4. Choose one primary web collection library after the target business and actual source sites are known: **Scrapy** for a Python-centric crawler or **Crawlee** for a TypeScript-centric HTTP/browser crawler.
5. Keep VALENTINA's own business rules, provenance, deduplication, review process, data model, database, and branded interface in this repository.

## Repository identity: two different projects called Huginn

These repositories must not be confused:

| Repository | What it is | License | Relevance |
|---|---|---|---|
| [scrypster/huginn](https://github.com/scrypster/huginn) / [huginn.sh](https://huginn.sh) | Go-based local AI-agent and coding orchestration application | BSL 1.1 | This is the Huginn discussed in the VALENTINA architecture |
| [huginn/huginn](https://github.com/huginn/huginn) | Classic Ruby event-monitoring and automation system | MIT | Different product; not the AI orchestration UI |

Do not fork or integrate the classic Ruby project by mistake.

## Candidate assessment

### 1. scrypster/huginn — agent orchestration

- Upstream: https://github.com/scrypster/huginn
- Default branch observed: `develop`
- Observed release: `v0.5.0` (2026-08-30)
- License: **Business Source License 1.1 (BSL 1.1)**, not MIT and not currently an OSI open-source license.
- Relevant capabilities: multi-agent delegation, custom agents, per-agent model/provider/endpoint settings, workflow scheduling, MCP stdio connections, headless/server mode, and an HTTP API under `/api/v1/`.
- NVIDIA NIM integration: source code supports a generic Custom/OpenAI-compatible provider. Its endpoint builder appends `/v1/chat/completions` when the configured base URL does not already end in a version segment. A real NIM model request and streaming test is still required.
- License implication: the checked-in BSL text allows production use when the use is strictly internal and is not offered as a product or service to third parties. Any offering of Huginn or a derivative as a hosted service, SaaS, managed platform, API service, or commercial product/service requires a separate commercial license. Preserve the license and notices, and obtain legal review if the deployment model changes.
- Assessment: useful as a replaceable orchestration service, but a young project (v0.5.0 at time of review). Keep VALENTINA's database and core pipeline independent of it.

**Decision:** install and configure separately; integrate with API/MCP. Fork only for a necessary upstream fix or UI modification that cannot be isolated in a bridge. Do not make Huginn the system of record.

### 2. Microsoft UFO² — Windows desktop execution

- Upstream: https://github.com/microsoft/UFO
- Current main repository includes the newer UFO³ project and the UFO² implementation/documentation under `ufo/`; the UFO² README identifies version 2.0.0.
- License: MIT at the repository root; retain the copyright and license notice when reusing substantial source.
- Platform: Windows desktop automation; combines Windows UI Automation / native Windows APIs with visual GUI actions.
- Integration: UFO² has agent and MCP-related facilities, but a stable, general-purpose external API suitable for VALENTINA-to-UFO² orchestration has not yet been established by this repository review. Prototype a narrow local bridge (CLI or custom MCP server) before committing to an integration.
- Assessment: useful as a secondary executor for cases where a site or desktop program requires UI interaction. It should not be the primary high-volume web crawler.

**Decision:** install/run separately and test a small bridge. Fork only if we need a maintained patch to the UFO²-specific code. If forking, pin the intended UFO² version/path rather than casually adopting UFO³.

### 3. NVIDIA NIM — model inference

- Documentation: https://docs.nvidia.com/nim/
- Hosted API: https://integrate.api.nvidia.com
- The hosted LLM interface exposes OpenAI-compatible chat completions. Huginn's generic Custom/OpenAI-compatible provider is a plausible connection path.
- NIM is a model-serving/inference platform, not an application repository to fork.
- Hosted API usage and downloadable/self-hosted NIM containers have different availability, hardware, and entitlement requirements. Validate the chosen model and endpoint before choosing a deployment mode.

**Decision:** configure as an external model provider; never place credentials in Git. Start with the hosted endpoint if the desired model is available there.

### 4. Crawling and extraction — choose one first

| Candidate | License | Strength | Best initial use |
|---|---|---|---|
| [Scrapy](https://github.com/scrapy/scrapy) | BSD-3-Clause | Mature Python crawler, structured extraction, queues and retry handling | Python-first crawling of known sites and structured pages |
| [Crawlee](https://github.com/apify/crawlee) | Apache-2.0 | TypeScript crawler with HTTP and real-browser crawling abstractions, request queues and storage | TypeScript-first collection, particularly when JavaScript rendering is needed |
| [Playwright](https://github.com/microsoft/playwright) | Apache-2.0 | Browser automation and page control | Browser engine/test dependency, not a separate business-data system |

Start with permitted APIs, RSS/Atom feeds, sitemaps, and ordinary HTTP before browser automation. Choose Scrapy or Crawlee after inspecting the business's actual source sites. Do not adopt both for the MVP without a measured need.

### 5. Classic huginn/huginn — event monitoring

- Upstream: https://github.com/huginn/huginn
- License: MIT.
- It creates agents to monitor feeds, websites, and events and pass events between agents.
- It is not the Go AI agent platform described above.

**Decision:** not selected for the initial VALENTINA architecture. Reconsider only if its event-monitoring features fill a concrete gap more simply than our chosen crawler/scheduler.

## What belongs in VALENTINA

Own and test these parts here:

- Product requirements and target-business profile.
- Canonical company, contact, and opportunity data model.
- Source registry and source-specific collection adapters.
- Evidence/provenance (source URL, capture time, publication time when known, extracted facts).
- Validation, deduplication, relevance scoring, and explainable rejection reasons.
- Human review, contact history, and opt-out/objection handling.
- Branded VALENTINA UI and application API.
- Observability, backups, security, and retention rules.

Keep AI orchestration, desktop automation, model serving, and crawler engines replaceable behind small interfaces.

## Forking and reuse strategy

A GitHub **fork is a separate repository**, not a directory that should be dumped inside VALENTINA. Use this order of preference:

1. **Package/dependency** — for a library such as Scrapy, Crawlee, or Playwright.
2. **Separate service/container** — for whole applications such as Huginn or UFO².
3. **MCP/HTTP adapter in VALENTINA** — to expose only the capabilities the product needs.
4. **Git submodule or pinned upstream reference** — only when source pinning is useful and the team is comfortable maintaining that workflow.
5. **Separate fork under MiguelDuval** — only when we need to patch/brand upstream code. Pin its upstream base, retain license notices, and document the divergence.

Avoid copying entire upstream repositories into this project. It creates update, security, licensing, and testing burden and makes it difficult to tell which code is ours.

## Integration spike before implementation

Once the target business and first source sites are identified, run a small, measurable proof of concept:

1. Configure Huginn's Custom provider against a chosen NVIDIA NIM model and verify a real chat-completion response, tool-call behavior if needed, and streaming.
2. Connect one read-only MCP tool that fetches a permitted public source and returns structured evidence.
3. Pass the result through a minimal VALENTINA opportunity schema and persist it.
4. Test UFO² separately with one specific desktop workflow; only build the bridge if it adds real value over API/HTTP crawling.
5. Compare Scrapy and Crawlee against the same representative pages before selecting one.

Do not begin large-scale scraping or contact collection until the target business, permitted sources, and data-handling rules are defined.

## Security and visibility note

At the time of this review, **MiguelDuval/VALENTINA is public**. The repository currently contains only generic project documentation, but it should be changed to private before adding proprietary source code, credentials, customer/contact data, internal business information, or real lead records. Never commit API keys or production data. Keep secrets in environment variables or a proper secret store.

## Audit trail

The repository was initially empty. Read/write access was tested through the connected GitHub API by creating and reading back the initial README.

- Initial bootstrap commit: `4965d56066d17d3778afe4999d53edea6efcb664`
- This evaluation is being added on a separate documentation branch; it does not change `main`.
