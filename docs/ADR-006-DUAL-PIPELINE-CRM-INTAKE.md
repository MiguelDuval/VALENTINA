# ADR-006 — Dual-Pipeline CRM: Cold Discovery + Hot Inbound Intake

**Status:** Owner-approved decision (task scope clarification)  
**Date:** 2026-11-06  
**Scope:** VALENTINA business domain — renovation, repair, construction, interior finishing  
**Applies to:** Domain model, source policy, API surface, Vue UI routes, permissions, AI usage

---

## 1. Decision

VALENTINA will support two explicitly separated acquisition pipelines:

- **Cold pipeline:** public opportunity discovery from published sources.
- **Hot pipeline:** inbound client enquiries received through the owner's business channels.

Both pipelines share company references, business categories, geography and reporting infrastructure, but maintain separate provenance, lifecycle rules, review queues and decision histories. AI may assist with extraction, classification and summarization in both pipelines, but it never autonomously qualifies, rejects, converts, contacts or acts externally.

This ADR supersedes the P0-only simplification framing and establishes VALENTINA's dual-pipeline CRM product direction.

---

## 2. Rationale

The Phase 0 audit confirmed VALENTINA reuses the pinned scrypster/huginn Go application as a single local runtime with an embedded Vue frontend. It does not provide renovation-opportunity CRM functionality, source-registry enforcement server-side, or SQLite-backed business persistence. The owner has now clarified that VALENTINA must become a dual-pipeline CRM:

---

## 3. Cold pipeline — public opportunity discovery

### 3.1 Purpose

Discover work published by others before the owner has direct contact with the publisher.

### 3.2 Processing chain (server-side, deterministic + AI-assisted)

Search → Candidate → Approved Source Check → Original Page Fetch → Extraction → Validation → Normalization → Deduplication → Assessment → Persistence → Human Review

### 3.3 Deterministic responsibilities (code, not LLM)

- Approved source IDs and host/path allowlists
- HTTP/HTTPS scheme validation
- Redirect checks and destination IP validation
- SSRF protection, timeouts and response-size limits
- Rate limits and request budgets
- Canonical URL normalization
- Date and currency normalization
- JSON Schema validation
- Stable identifiers
- Exact and deterministic duplicate detection
- Evidence persistence
- Expiry and status bookkeeping

### 3.4 AI responsibilities (may assist)

- Expand search terms
- Understand Spanish/Valencian/configured source languages
- Extract facts into the versioned JSON schema contract
- Summarize long descriptions
- Distinguish `direct_request`, `public_tender`, `subcontracting_request`, `early_demand_signal`, `unknown`
- Identify service categories
- Explain relevance to the owner profile

---

## 4. Hot pipeline — inbound client enquiries

### 4.1 Purpose

Receive requests from people/companies contacting the owner's business (website forms, messages, etc.).

### 4.2 Processing chain

Receive → Validate → Deduplicate → Preserve Original → Extract → Classify → Identify Missing Information → Assess Fit → Human Review → Convert or Decline

### 4.3 AI may assist (suggest only)

- Service category, approximate project scope, location
- Budget if explicitly stated, timing and urgency
- Missing details, possible company match, suggested response draft

### 4.4 AI must not

- Send a response, promise a price, make an appointment, submit a proposal or contact a customer
- Independently make irreversible business decisions

### 4.5 External channels

The server remains bound to loopback (127.0.0.1). No public webhook, reverse proxy, ngrok tunnel or unauthenticated ingress is introduced in this phase. Future channels (owned website forms, email intake, WhatsApp Business, Telegram) require official APIs and safe local pull/poll approaches, investigated under a separate security review.

### 4.6 Contact-data boundary (narrowly scoped exception)

The original v0.1 cold-opportunity contract excludes contact-person data. The hot pipeline permits a narrowly scoped, purpose-limited exception: minimal contact information supplied directly by an incoming customer and necessary to process that enquiry. This exception must be documented explicitly; no personal contact data is harvested from cold public websites merely because it is visible.

- Suggest likely duplicate pairs for human review
- Identify missing and contradictory details

### 3.5 AI must not

- Invent budgets, dates, locations, deadlines, company identity or contact details
- Upgrade an early signal to a confirmed request without source evidence

---

## 5. Domain entities

| Entity | Purpose | Pipeline |
|--------|---------|----------|
| Company | Public publisher / client / contractor reference | Shared |
| Opportunity | Publicly discovered commercial opportunity | Cold |
| InboundLead / ClientInquiry | Customer-initiated enquiry | Hot |
| Source | Approved publisher/endpoint + collection rules | Shared |
| SourceEvidence | One observation of one page/document at one time | Cold |
| ResearchRun | Manual research action + history | Cold |
| ReviewDecision | Human qualification/rejection/conversion decision | Both |

A qualified inbound enquiry may later be converted into or linked to an Opportunity/project. The original enquiry and its history are preserved.

### 5.1 Cold opportunity lifecycle

candidate → verified → needs_review → qualified → rejected → expired → archived

### 5.2 Hot inbound lifecycle

new → triage → needs_information → qualified → converted → declined → spam → archived

A model may recommend a status; it must not perform irreversible transitions or silently discard records.

## 6. AI provider / compute

All provider options supported by the pinned Huginn runtime are preserved (Ollama, external, custom/OpenAI-compatible including NVIDIA NIM, OpenAI, DeepSeek, Z.AI, Anthropic, OpenRouter, Google AI, Vertex AI, Managed/local llama.cpp). NVIDIA NIM is preferred initial provider, never exclusive. Per-agent provider/model selection is preserved.

## 7. Security and permissions

- Discovery agents are read-only, least-privilege; no shell, file-write, Git mutations or external actions
- `--dangerously-skip-permissions` is not used
- Source registry allowlists are enforced server-side by a VALENTINA source-policy MCP adapter before unattended collection
- Raw `fetch_url` does not enforce VALENTINA host/path allowlists; unattended agents do not use it as a replacement
- MCP `approval_gate` remains enabled
- Local SQLite persistence; no remote/cloud sync

## 8. Deferred integrations and prerequisites

- One approved real source (source profile approved, terms reviewed, adapter implemented and tested)
- Live NVIDIA NIM integration test (credentials available, real request verified)
- Official website/messaging channel integrations (secure local pull/poll only, separate security review)
- Playwright MCP (only if a permitted source genuinely requires JavaScript rendering)
- NSIS Windows installer (later milestone)

## 9. First usable implementation scope

Stage C–G implementation: domain persistence (SQLite), minimal API, Vue CRM UI (Overview, Opportunities, Incoming Leads, Companies, Sources, Research), source-policy foundation, manual inbound entry, AI-assisted extraction through provider abstraction, human review workflow. One approved source or synthetic fixtures — no real source scraping without owner approval.

- Silently merge or delete records
- Override source permissions
- Execute shell, Git, file-write or external-action tools
- Contact publishers or apply for work

### 3.6 Evidence requirement

Every review-ready cold opportunity must contain an original publisher reference and supporting SourceEvidence record: canonical URL, publisher, observation time, content hash when available, extractor ID/version, and short factual excerpts or locators.


1. **Cold:** find public construction/renovation opportunities, verify them, preserve evidence, assess fit and place them in a review queue.
2. **Hot:** collect customer enquiries arriving through the owner's business channels, preserve the request, extract meaning and place them in a separate incoming-client queue.

A clear separation avoids conflating publicly discovered opportunities (unverified, publisher-originated) with inbound enquiries (customer-originated, potentially containing contact information supplied directly by the customer).
