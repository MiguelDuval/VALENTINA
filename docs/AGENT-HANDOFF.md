# VALENTINA — active agent handoff

**Updated:** 2026-11-18
**Status:** READY_FOR_REVIEW — research + inbound-intake milestone implemented
**Task:** VAL-P3/INTAKE-001 — controlled multi-source discovery and hot-lead intake vertical slice
**Owner authorization:** The owner explicitly authorized implementation of internet-wide opportunity discovery and intake of hot leads. This authorizes building the P3 source-policy/search pipeline and P2 inbound-lead workflow. It does NOT authorize bypassing site terms, CAPTCHA/login restrictions, unrestricted crawling, external outreach, form submission, or public exposure of the local app.

## Repository checkpoint

- Repository: MiguelDuval/VALENTINA
- Branch: `feature/valentina-research-intake-engine-20261118`
- Base HEAD: `b1705293c6c3b823c1dec198366f62f445284fc1`
- Huginn submodule: `f01c1d84c08b1de972d8cc268da25602598adff8`
- Origin: `https://github.com/MiguelDuval/VALENTINA.git`
- PR: none for this task at assignment time.
- Working tree has 57 pre-existing untracked user files/directories (scripts, logs, ZIPs, temp builds). Preserve all; never stage all, clean, delete, move, or overwrite them.
- Canonical workspace: `C:\Users\migue\Projects\VALENTINA`.

## Required reading

Read `AGENTS.md`, `.clinerules/00-valentina-project.md`, `docs/AGENT-COLLABORATION.md`, `docs/ARCHITECTURE.md`, `docs/MVP-DELIVERY-ROADMAP.md`, `docs/PRODUCT-AND-WORKFLOWS.md`, `docs/DATA-AND-EVIDENCE.md`, `docs/SOURCE-REGISTRY.md`, `docs/SECURITY-AND-SECRETS.md`, `docs/SECURITY-PERMISSION-MATRIX.md`, `docs/WORK-QUEUE.md`, `docs/UI-IMPLEMENTATION-STATUS.md` and inspect the existing CRM domain/API/UI before editing. Verify current HEAD and branch again.

## Product outcome

Deliver a useful local-first system that has two distinct pipelines:

1. **Cold opportunity discovery:** identify public signals and explicit renovation/restoration/fit-out/repair requests across Valencia and owner-configured service areas; retrieve original evidence; classify intent; score fit/confidence; deduplicate; place candidates in the human review queue.
2. **Hot inbound intake:** capture people who explicitly ask for a quote/contact through a website or other configured channel; preserve their own contact details and consent; acknowledge/triage only when explicitly configured; link to a company/opportunity; track response status and follow-up.

Never mix a speculative public signal with a hot lead. Never label an indirect signal as a confirmed order. The app must make evidence, contact permission, source, freshness, missing facts and confidence visible.

## Implementation priorities — one large coherent pass

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
- Search-provider interface must support configurable provider credentials via protected local secret storage; Brave is a candidate because current Huginn web_search requires a separate Brave key. Do not assume NIM key is a search key. If credentials are absent, return actionable “not configured” state and keep offline tests green.
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
- No paid subscriptions assumed; prefer free official APIs, RSS, public datasets and user-configured free tiers. Do not promise that “all internet” can be exhaustively searched.
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

**`valentina/sourcepolicy/`** — Source registry + fail-closed policy gateway
- `registry.go`: SourceProfile, ApprovalStatus, Registry with CRUD and counts
- `gateway.go`: PolicyGateway with Evaluate/EvaluateRedirect, SSRF checks, canonical URL, sanitized errors
- `limits.go`: RateLimiter, ConcurrencyLimiter, DailyQuotaTracker
- `errors.go`: SanitizedError, Forbidden, Sanitize()
- Tests: registry_test.go, gateway_test.go (deterministic coverage)

**`valentina/sourceadapter/`** — Source adapters/provider interfaces
- `provider.go`: Provider interface, BaseProvider, ErrNotConfigured
- `official_api.go`: OfficialAPIAdapter with API key support
- `rss_adapter.go`: RSS/Atom feed parsing
- `sitemap_adapter.go`: Sitemap XML parsing
- `html_adapter.go`: Bounded HTML extraction with script/style stripping
- Tests: provider_test.go (mock HTTP client)

### Modified files

**`valentina/api/handlers.go`**
- Added `POST /api/v1/valentina/intake` route and handler
- HMAC-SHA256 verification (X-Valentina-Signature header, configurable secret)
- Idempotency key support
- Honeypot field (`website`) spam protection
- Consent + privacy notice required booleans
- 1 MB body limit, strict JSON schema validation, safe logging

**`valentina/api/handlers_test.go`**
- Added intake tests: success, missing consent, missing signature, invalid signature, honeypot
- Added `POST /api/v1/valentina/intake` to expected routes

**`valentina/web/src/views/LeadsView.vue`**
- Lead detail view with consent display, contact info, services, location, budget
- Status transitions: triage, qualify, decline, spam (with valid transition rules)
- Conversion from qualified lead to opportunity (explicit human action only)
- Manual creation form with consent checkbox

### Test results (local, via temporary Huginn module tree)
- `valentina/sourcepolicy`: 24 tests PASS
- `valentina/sourceadapter`: 9 tests PASS
- `valentina/domain`: PASS
- `valentina/api`: PASS (including intake tests)

### Pending
- Frontend typecheck, Vitest, Vite production build
- Full Windows build (scripts/build-valentina.ps1)
- Isolated runtime smoke test
- Diff review and staging
- Commit and push
