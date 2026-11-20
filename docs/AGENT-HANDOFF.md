# VALENTINA вЂ” active agent handoff

**Updated:** 2026-11-20
**Status:** IN_PROGRESS вЂ” hot-intake privacy_notice_version hardening pass
**Task:** VAL-INTAKE-001 вЂ” secure hot-lead intake vertical slice (privacy notice version + security audit)
**Owner authorization:** P2 hot inbound intake explicitly authorized 2026-11-18. Does NOT authorize bypassing site terms, CAPTCHA/login restrictions, unrestricted crawling, external outreach, form submission, or public exposure of the local app.

## Repository checkpoint

- Repository: MiguelDuval/VALENTINA
- Branch: `feature/valentina-manual-rss-runner-20261120`
- Base HEAD: `ff9af74ffbe3f288878f08285c4059f530326370`
- Huginn submodule: `f01c1d84c08b1de972d8cc268da25602598adff8`
- Origin: `https://github.com/MiguelDuval/VALENTINA.git`
- PR: not created; no merge/release.
- Working tree has pre-existing untracked user files/directories (scripts, logs, ZIPs, temp builds). Preserve all; never stage all, clean, delete, move, or overwrite them.
- Canonical workspace: `C:\Users\migue\Projects\VALENTINA`.

## Required reading

Read `AGENTS.md`, `.clinerules/00-valentina-project.md`, `docs/AGENT-COLLABORATION.md`, `docs/ARCHITECTURE.md`, `docs/MVP-DELIVERY-ROADMAP.md`, `docs/PRODUCT-AND-WORKFLOWS.md`, `docs/DATA-AND-EVIDENCE.md`, `docs/SOURCE-REGISTRY.md`, `docs/SECURITY-AND-SECRETS.md`, `docs/SECURITY-PERMISSION-MATRIX.md`, `docs/WORK-QUEUE.md`, `docs/UI-IMPLEMENTATION-STATUS.md` and inspect the existing CRM domain/API/UI before editing. Verify current HEAD and branch again.

## Product outcome

Deliver a useful local-first system that has two distinct pipelines:

1. **Cold opportunity discovery:** identify public signals and explicit renovation/restoration/fit-out/repair requests across Valencia and owner-configured service areas; retrieve original evidence; classify intent; score fit/confidence; deduplicate; place candidates in the human review queue.
2. **Hot inbound intake:** capture people who explicitly ask for a quote/contact through a website or other configured channel; preserve their own contact details and consent; acknowledge/triage only when explicitly configured; link to a company/opportunity; track response status and follow-up.

Never mix a speculative public signal with a hot lead. Never label an indirect signal as a confirmed order. The app must make evidence, contact permission, source, freshness, missing facts and confidence visible.

## Implementation priorities вЂ” one large coherent pass

### A. Source policy gateway (security-critical)
- Implement VALENTINA-owned, in-process Go source registry/policy and a single narrow fetch boundary. Every fetch/search result must be associated with a source profile and run ID.
- Fail closed unless source is approved and enabled. Enforce exact scheme/hostname/path scope, safe URL canonicalization, redirect revalidation, DNS/IP SSRF protections (loopback/private/link-local/reserved/metadata), timeouts, concurrency/rate/daily caps, response size, content type, and sanitized errors.
- Do not allow the Huginn generic `fetch_url` tool to bypass this boundary for unattended discovery. Keep raw unrestricted fetch/search out of any autonomous agent allowlist.
- Store approval metadata, reviewed terms/access guidance, permitted fields/purpose, cadence, adapter type and last error. Pending/paused sources are never fetched.
- Add deterministic tests for allowed/disallowed host, path escape, redirect, private IP, malformed URL, timeout, size, content type, paused/pending source, rate limits and safe error redaction.
- No general-purpose crawler, proxy rotation, CAPTCHA bypass, login-wall bypass, scraping behind access controls or arbitrary URL fetch API.

### B. Multi-source discovery architecture
- Create provider/adapter interfaces with deterministic normalization into the existing Opportunity + SourceEvidence model.
- Support official APIs, RSS/Atom, sitemap/public dataset and bounded HTML/document adapters as separate capabilities. Prefer APIs/RSS/datasets over HTML.
- Search-provider interface must support configurable provider credentials via protected local secret storage; Brave is a candidate because current Huginn web_search requires a separate Brave key. Do not assume NIM key is a search key. If credentials are absent, return actionable вЂњnot configuredвЂќ state and keep offline tests green.
- Add source-aware query packs for Russian, Spanish and Valencian/Catalan renovation terms; use owner-configurable service categories, geography and exclusions. Keep actual owner profile and API keys outside Git. Do not fabricate profile facts; provide a first-run editable profile with safe defaults.
- Classify demand evidence explicitly: explicit request for quote/contractor, public tender/procurement, subcontractor/vendor call, planning/permit signal, project announcement, directory/company listing, general early signal. Use explainable rules/features; LLM enrichment must be optional and never overwrite source facts.
- Score fit and confidence separately; explain every score; freshness/expiry and deduplication must be deterministic.
- Research runs must support queued/running/partial/completed/failed/cancelled, progress, per-source counters/errors, cancellation and safe retry. Initial mode manual-only.
- Do not activate any external source by default unless its terms/scope are reviewed and approved. Create source candidates as pending_review, not silently approved.

### C. Hot inbound leads
- Complete InboundLead detail/create/review API and UI, with separate domain semantics from Opportunity.
- Add secure local intake endpoint for website form/webhook integrations: strict schema/size/rate limits, configurable high-entropy bearer/HMAC verification, timestamp/replay protection where feasible, idempotency key, honeypot field support and spam quarantine. Do not log secrets or full contact payloads.
- Preserve source channel, submitted_at, name/company, phone/email, requested service, location, free-text request, preferred contact method, consent/consent timestamp/privacy notice version, attachments metadata (never accept arbitrary executable uploads).
- Validate consent and privacy requirements. Do not infer marketing consent from a quote request.
- Provide CSV import with preview/field mapping/validation/deduplication if consistent with existing architecture.
- Add inbound pipeline states (new, needs_info, qualified, converted, declined, spam, closed) and human reasoned decisions; never auto-send messages.
- Provide an adapter interface for website, email and other channels, but do not claim email/site connection is live without credentials/endpoint and an end-to-end test. Do not expose loopback API to the internet.

### D. UI and workflow
- Research screen: sources, query packs, geography, run start/cancel, progress, partial failure, evidence counts, dedupe summary, manual approval.
- Sources screen: pending/approved/paused, scope and policy details, terms reviewed date, caps, last check and test connection.
- Opportunities screen: demand class, geography, fit/confidence, freshness, source count and review state.
- Inbound Leads screen: separate hot queue, detail, consent, contact fields, channel, response state and conversion link.
- Use existing Vue shell and in-process Go API; do not introduce a second web app or second database.
- Add helpful empty/error/loading states and realistic synthetic fixtures only.

### E. Documentation and evidence
- Update `docs/WORK-QUEUE.md`, `docs/AGENT-HANDOFF.md`, `docs/SOURCE-REGISTRY.md`, `docs/DATA-AND-EVIDENCE.md`, `docs/SECURITY-AND-SECRETS.md` and `docs/UI-IMPLEMENTATION-STATUS.md` to match actual behavior.
- Include a source onboarding matrix with exact status and no invented approvals.
- Document first-run setup, local website integration contract, how to configure a search provider, limits, privacy/consent, and how to add a source.
- Never include real owner/contact records, API keys, .env, databases, logs, screenshots with personal data or private profile.

## Non-goals / hard boundaries
- No paid subscriptions assumed; prefer free official APIs, RSS, public datasets and user-configured free tiers. Do not promise that вЂњall internetвЂќ can be exhaustively searched.
- No unrestricted crawler or mass scraping.
- No auto-contact, email/SMS/WhatsApp sending, quote submission, account creation, purchase, or any external write action.
- No public/LAN binding, tunnel, cloud relay or inbound internet listener for the local desktop app.
- No NIM live test unless the owner has configured a key; never ask the agent to print env vars or secrets.
- No automatic recurring research until the source policy boundary and manual workflow are proven.
- Do not modify the pinned upstream Huginn submodule.
- No PR merge, force push, reset, rebase, clean, history rewrite or direct main changes.

## Acceptance criteria
1. Source registry and policy enforced in code, not just prompts/docs.
2. All fetches are source-scoped, bounded, SSRF-resistant and test-covered.
3. At least one useful discovery adapter/provider works end-to-end with a permitted public source, OR the workflow clearly reports credentials/source approval as the only blocker; do not fabricate successful retrieval.
4. Search results are persisted as evidence-backed candidates and deduplicated.
5. Hot inbound lead API/UI supports detail, consent, idempotent intake and human triage.
6. Cold opportunities and hot inbound leads remain separate, linked only by explicit conversion.
7. Frontend tests, Go tests, typecheck, full Windows build and isolated runtime smoke pass.
8. Exact branch/HEAD, changed files, tests, known limits and artifact hash documented.
9. Review full diff; stage only intended files. Preserve all 57 pre-existing untracked artifacts.

## Agent execution
Work in large coherent blocks. Do not stop after writing a plan. Implement, test, fix, document, review and commit. Keep one task branch. Push only after tests and final diff review using normal non-force push. Do not merge PR. ChatGPT will independently review the result.

## Implementation summary (this task)

### New packages

**`valentina/sourcepolicy/`** вЂ” Source registry + fail-closed policy gateway
- `registry.go`: SourceProfile, ApprovalStatus, Registry with CRUD and counts
- `gateway.go`: PolicyGateway with Evaluate/EvaluateRedirect, SSRF checks, canonical URL, sanitized errors
- `limits.go`: RateLimiter, ConcurrencyLimiter, DailyQuotaTracker
- `errors.go`: SanitizedError, Forbidden, Sanitize()
- Tests: registry_test.go, gateway_test.go (deterministic coverage)

**`valentina/sourceadapter/`** вЂ” Source adapters/provider interfaces
- `provider.go`: Provider interface, BaseProvider, ErrNotConfigured
- `official_api.go`: OfficialAPIAdapter with API key support
- `rss_adapter.go`: RSS/Atom feed parsing
- `sitemap_adapter.go`: Sitemap XML parsing
- `html_adapter.go`: Bounded HTML extraction with script/style stripping
- Tests: provider_test.go (mock HTTP client)

### Modified files

**`valentina/domain/models.go`**
- Added `Source.IsApproved()` and `Source.CanFetch()` methods

**`valentina/domain/storage.go`**
- Added `UpdateResearchRun()` вЂ” updates status, counts, notes for research run lifecycle
- Added `UpdateSource()` вЂ” updates approval_status, approval_reason, terms_reviewed_at, last_reviewed_at, notes

**`valentina/api/handlers.go`**
- Added routes: POST /api/v1/valentina/research-runs/{id}/cancel, POST /api/v1/valentina/sources/{id}/approve, POST /api/v1/valentina/sources/{id}/pause, POST /api/v1/valentina/sources/{id}/test, GET /api/v1/valentina/query-packs, GET /api/v1/valentina/geography
- Added `handleCancelResearchRun` handler

**`valentina/api/handlers_extra.go`** (new file)
- `sourceUpdateRequest` type
- `handleApproveSource` вЂ” approves a source with terms_reviewed_at timestamp
- `handlePauseSource` вЂ” pauses a source
- `handleTestSource` вЂ” validates source is approved (connectivity check)
- `queryPack` type + `handleListQueryPacks` вЂ” Spanish/Valencian/Russian renovation query packs
- `geographyConfig` type + `handleGetGeography` вЂ” Valencia, Spain geography scope

**`valentina/api/handlers_test.go`**
- Added intake tests: success, missing consent, missing signature, invalid signature, honeypot
- Added `POST /api/v1/valentina/intake` to expected routes

**`valentina/web/src/views/LeadsView.vue`**
- Lead detail view with consent display, contact info, services, location, budget
- Status transitions: triage, qualify, decline, spam (with valid transition rules)
- Conversion from qualified lead to opportunity (explicit human action only)
- Manual creation form with consent checkbox

**`valentina/web/src/views/ResearchView.vue`**
- Toolbar with "Start research" button
- Run cards with status badges, stats (candidates/verified/duplicates/accepted/rejected)
- Cancel button for running research runs
- New run button for completed/failed runs
- startRun() and cancelRun() functions calling API endpoints

**`valentina/web/src/views/SourcesView.vue`**
- Policy details: terms_reviewed_at, max_parallel_requests, daily_request_limit, timeout_seconds
- Approve/Pause/Test buttons per source based on current approval status
- approveSource(), pauseSource(), testSource() functions calling API endpoints

### Test results (local, via temporary Huginn module tree)
- `valentina/sourcepolicy`: 24 tests PASS
- `valentina/sourceadapter`: 9 tests PASS
- `valentina/api`: PASS (including intake tests)
- `valentina/domain`: PASS
- Frontend SFC structure regression test: SfcStructure.test.ts (new)

### Pending
- Go vet/build for handlers_extra.go and updated handlers.go (verified via copy to Huginn module tree)
- Frontend typecheck, Vitest, Vite production build
- Full Windows build (scripts/build-valentina.ps1)
- Isolated runtime smoke test
- PR вЂ” not created (no merge requested)


---

## 2026-11-20 technical-lead implementation checkpoint (supersedes the older task status above)

**Repository:** MiguelDuval/VALENTINA
**Active feature branch before publication:** `feature/valentina-manual-rss-runner-20261120`
**Base HEAD:** `1904cdb9f3d91ba87b4de1de84acd1ffc05a425c`
**Pinned Huginn submodule:** `f01c1d84c08b1de972d8cc268da25602598adff8`
**PR:** not created; no merge/release.
**State:** local feature implementation and tests passed; GitHub CI and owner-machine acceptance pending.

### Delivered in this implementation pass

- Replaced the stub ResearchRun endpoint with manual RSS/Atom-only approved-source fetch using the fail-closed policy gateway, strict terms-review/approval-reason preflight, local run budgeting and capped source/entry counts. Unreviewed and non-RSS adapters stay blocked.
- Extracted feed candidates safely, recorded original feed locator and external notice URL, persisted unverified candidates and SourceEvidence in SQLite with exact source+URL deduplication. No article was fetched or verified; demand class remains `unknown`, status `candidate`, never a confirmed order.
- Added deterministic success/failure run history, completed timestamps, SQLite schema migration for `updated_at`, stricter source creation/approval tests, and source-adapter truncation/non-2xx and concurrency release tests.
- Enabled explicit manual search action in the existing VALENTINA Vue Research UI.
- Changed Windows build script to extract only the pinned upstream tracked Git commit with `git archive`, rather than copy local untracked upstream submodule files.
- Updated ADR-007, product positioning, work queue, and audit documentation.
- **Hot intake privacy_notice_version:** Added SQLite migration v12, InboundLead model field, create/read/list persistence, intake payload validation, and LeadsView UI display.
- **Intake security audit:** Added slog.Warn failure logging for rate-limit exceeded and HMAC verification failure. Added offline tests for version persistence, optional version, same-key-different-payload idempotency, missing consent/signature fail-closed.

### Local verification

- `go test ./internal/valentina/...` in an isolated extraction of pinned Huginn with current VALENTINA Go sources: **4 packages passed** (api, domain, sourceadapter, sourcepolicy), including privacy_notice_version persistence, same-key-different-payload idempotency, rate-limit/HMAC failure logging, synthetic feed end-to-end, errors/dedup, approval gate and concurrency safety.
- `scripts/build-valentina.ps1 -Profile Full` completed successfully after changing the upstream copy step to a pinned Git archive. Frontend: **116 test files / 2004 tests passed**, Vue typecheck passed, Vite production build passed; CRM Go tests passed, Huginn server compilation passed, Windows EXE compiled.
- Clean-archive local test EXE: **67,193,344 bytes**, SHA256 `67dc0adf5c4447e194de1c06a026e6a1b7d6fe967f36c949c80559663aa0ea8e`. The local ZIP is in the machine's temporary build output and is **not** a GitHub Actions artifact. The archived build began before the final source-creation hardening edit; the final Go change was independently retested, and GitHub CI must rebuild the final commit.
- `git diff --check` was run during this implementation; check the final staged diff again before publication.
- No external production feed, live NIM, Messenger API, Windows installer, clean install or end-to-end runtime smoke verified.

### Mandatory next work

1. Verify the pushed feature branch CI build and download/run SHA-labelled EXE; do not claim a successful CI artifact without actual Actions evidence.
2. Obtain owner review of a specific public source's terms and URLs; test RSS feed в†’ evidence в†’ UI review with source content, not only synthetic tests.
3. Add secure privacy notice version to hot intake persistence and tests — DONE (migration v12, model, storage, handler, UI, offline tests).
4. Extend acquisition only through approved adapters/official APIs; add real cancellation, progress and resilient retry before unattended schedules.
5. Continue ADR-007 project/document workspace only after discovery/intake foundation acceptance.

### Preservation

Existing user temp scripts, ZIPs, work dirs, original submodule untracked `internal/valentina`, and unrelated dirty files were not deleted, stashed, reset or force-pushed. Publish only explicitly reviewed first-party files from the feature branch.

