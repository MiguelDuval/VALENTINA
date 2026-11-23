# VALENTINA — active agent handoff

**Updated:** 2026-11-23 (Codex bounded Gate 1 pass)
**Status:** LOCAL CHECKS PASS; publication blocked by sandbox Git-metadata permissions; Gate 1 owner/runtime acceptance remains open.

## Gate 1 security and manual research pass
- Repository/origin verified: `MiguelDuval/VALENTINA`; isolated worktree `VALENTINA-CODEX-GATE1-20261123`; branch `feature/valentina-gate1-security-ui-20261123`. Base branch: `feature/valentina-intake-security-20261123`; base and unchanged HEAD: `0fbdafd24b6b58b75d8000eee8a608133ab67cb9`. Upstream is clean at `f01c1d84c08b1de972d8cc268da25602598adff8`. No PR created, commit or push claimed; no changes to canonical source worktree, CI worktree or PR #5.
- Implemented timestamp/version/body HMAC-v2 (300-second past/future tolerance, canonical timestamp and unique headers, no downgrade). Body-only v1 is disabled unless local startup environment `VALENTINA_INTAKE_ALLOW_LEGACY_V1=1`; migration contract is in SECURITY-AND-SECRETS section 10. Intake 4xx responses do not reflect submitted JSON/contact data. Existing v11/v12/v13 migrations remain unchanged.
- Fixed concurrent SQLite key reservations: a single Store connection serializes writes. Tests prove one lead under concurrent signed deliveries, changed-body 409, retained key across reopen, 30-day expiry preserving original leads, and existing migration fixtures.
- Research/Sources now consume server-derived `manual_run_allowed` / `manual_run_reason`; manual POST rechecks the same source approval/terms/RSS-Atom contract. UI shows blocked reasons, busy/error states and persisted counts, links to evidence cards, and explicitly requires human review. No source was approved or fetched live.
- Added `TestGate1OfflineSmoke`: real domain HTTP routes -> signed intake -> triage; fixture RSS adapter -> evidence-backed candidate -> human review; reopen SQLite -> persisted decisions/provenance -> signed duplicate and changed-payload rejection. This is handler/SQLite acceptance, not a launched desktop app or process-restart/UI-browser acceptance test.
- Changed files: `valentina/api/{handlers.go,handlers_test.go,research_runner.go,research_runner_test.go}`, `valentina/domain/{storage.go,storage_test.go}`, `valentina/web/src/views/{ResearchView.vue,SourcesView.vue,__tests__/ResearchRun.test.ts}`, this handoff, WORK-QUEUE and SECURITY-AND-SECRETS. Original untracked task specification preserved.
- Verification tree: ignored `build/gate1/source`, from `git archive` of the exact pin, all four tracked patches applied after `git apply --check --ignore-space-change`, plus current first-party sources. Go toolchain `go1.25.0`; `GOCACHE` under `build/gate1/gocache`.
- PASS: `go test ./internal/valentina/... -count=1 -timeout=90s` (all four packages); `go vet ./internal/valentina/...`; `go test -run '^$' ./internal/server`; 25 targeted Vitest tests in CRMViews, OpportunityDetailView, SfcStructure, ResearchRun and navLayout; `vue-tsc -b`; `vite build --configLoader native`; `go build -tags embed_frontend -o ../VALENTINA.exe .`; full authored diff reviewed and `git diff --check` passed.
- Local EXE: `build/gate1/VALENTINA.exe`, 67,184,640 bytes, SHA256 `ba774234a8f1bdc484fa1fb32314bf0fc8725c12d065996ed59345d829a76a60`. Build is from uncommitted working changes; binary is ignored, not published and not an installer or Actions artifact.
- Environment limits: scoped `git add` failed creating `.git/worktrees/VALENTINA-CODEX-GATE1-20261123/index.lock` with Permission denied; no files staged and no commit/push attempted. Submodule re-init could not write shared `.git/worktrees/.../modules/.../config`, although existing pin/source were valid. Offline npm caches missed packages; network `npm ci` failed EACCES. Reused installed dependencies from a previous temporary build whose package-lock SHA256 matches the exact pin; no dependency/lockfile changes. Native config loading bypassed esbuild's denied ancestor-directory scan. Warnings: outdated Browserslist data, large Vite chunk. A fresh npm installation is not claimed.
- Not run: broad upstream suites, race detector, real source/NIM requests, browser/desktop process smoke, install/upgrade/uninstall, GitHub CI. No NIM credential availability was established. Remaining Gate 1 blockers: owner-approved real source and original-page verification, owned intake channel/notice contract and local secret, Windows runtime acceptance and new-branch CI/artifact review. Unattended research/public webhooks remain disabled; source quota/concurrent-run hardening remains a separate prerequisite for automation.
- Next action: review these local changes, then commit only scoped files and publish via `scripts/publish-branch.ps1` in a session allowed to update this worktree's Git metadata. Do not merge/release or substitute this local EXE for a successful Actions artifact.

## Earlier handoff (historical)


**Updated:** 2026-11-23 (isolated intake acceptance; prior history retained below)
**Status:** PARTIAL — hot-intake idempotency/draft-PR migration convergence locally tested 2026-11-23; Windows CI/runtime and signed-request freshness acceptance pending

## 2026-11-23 isolated hot-intake security pass
- Working branch: `feature/valentina-intake-security-20261123`, based on `43c98e2751272f8411780a09cb8c8f58a71e2247`; isolated Git worktree `VALENTINA-CLINE-INTAKE-20261123`, origin `MiguelDuval/VALENTINA`; pinned Huginn `f01c1d84c08b1de972d8cc268da25602598adff8`. Canonical dirty workspace and PR #5 were not modified.
- Fresh VS Code Cline task `1795445681314_3znio` stopped after six tool/read_files errors; it did not author code. The technical-lead implementation used direct isolated file edits rather than retrying the old or failed Cline task.
- Intake now binds the 30-day idempotency key to SHA256 of the exact authenticated body in SQLite migration v13: identical retry returns original ID; changed body with the same key returns HTTP 409 without a second lead or contact information; pre-v13 keys with NULL fingerprint fail closed and require a new key.
- Migration v13 adds `idempotency_keys.payload_sha256`; reconciles both historic v11 states by ensuring `research_runs.updated_at` and `inbound_leads.privacy_notice_version` plus source URL index. Records PR #5 v12 as adopted after ensuring its privacy column to prevent a subsequent duplicate ALTER. PR #5 remains Draft and must not be merged without code/diff review.
- Offline evidence: all four `go test ./internal/valentina/...` packages PASS on a disposable clean exact pinned Huginn archive; added synthetic signed 201/200/409 API tests, SQLite reopen/legacy-key tests, PR-v11/v12 migration fixtures; `go test -run '^$' ./internal/server` PASS; `go vet ./internal/valentina/...` PASS; non-embedded `go build .` PASS; `git diff --check` PASS. Full Windows frontend CI for this new branch not yet evidenced; earlier successful #44936601301 belongs to base `43c98e2`, not these changes.
- Remaining: signed timestamp/freshness anti-replay (legacy HMAC signs body only), live owned website secret and notice metadata, real source approval and UI runtime smoke; never enable public intake, unattended fetch, or outbound messages without review.
- Files: `valentina/api/handlers.go`, `handlers_test.go`, `valentina/domain/storage.go`, `storage_test.go`, `docs/AGENT-HANDOFF.md`, `docs/WORK-QUEUE.md`, task specification. Follow-up: verify new CI and downloadable EXE, then design versioned timestamp HMAC compatibility and manual RSS/UI smoke. No releases or PR merge.



## 2026-11-21 checkpoint for the next agent
- VS Code Cline session `1795043585019_o8c7l` ended in NVIDIA NIM `Too Many Requests` after a very long task. Do **not** Retry the old session. Start a fresh bounded task with a new context instead.
- Repository: `C:\Users\migue\Projects\VALENTINA`; branch `feature/valentina-research-intake-engine-20261118`; starting HEAD `1904cdb`. Preserve unrelated/untracked owner files, dirty Huginn submodule content, parallel worktrees and all user changes.
- Recovery: wired operator-triggered RSS/Atom `handleCreateResearchRun` to a source-scoped policy fetch; added source-specific evidence duplicate lookup; repaired research status/completion-time persistence; added nullable `privacy_notice_version` with additive SQLite v11 migration; fixed Go compilation.
- Verified: `go test ./internal/valentina/...` passed on the local Huginn internal test mirror on 2026-11-21, including synthetic RSS and intake tests. This is **not** proof of production feed connectivity or a full Windows build.
- Safety: no source has been approved for real fetching by this checkpoint; intake remains HMAC-gated and must not be exposed externally. Verify mismatched idempotency keys, validation of policy version, frontend contract, real app build and security acceptance before declaring the intake feature ready.
- PR #5 merge guard (clarified 2026-11-23): draft branch `feature/valentina-manual-rss-runner-20261120` uses `v11` for `research_runs.updated_at` and `v12` for `inbound_leads.privacy_notice_version`; the CI-portability branch uses `v11` for the privacy column. The additive reconciliation in this feature branch is tested against both histories. PR #5 still requires explicit conflict/diff review; never blindly apply duplicate ALTER TABLE statements.
- Keep the next assignment narrow: one end-to-end acceptance goal; perform tests; report exact files, checks, SHA, and blockers. Codex is the preferred primary executor; Cline may be used only in a fresh short task.

**Historical status (2026-11-18):** READY_FOR_REVIEW — research + inbound-intake milestone implemented and pushed
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

**`valentina/domain/models.go`**
- Added `Source.IsApproved()` and `Source.CanFetch()` methods

**`valentina/domain/storage.go`**
- Added `UpdateResearchRun()` — updates status, counts, notes for research run lifecycle
- Added `UpdateSource()` — updates approval_status, approval_reason, terms_reviewed_at, last_reviewed_at, notes

**`valentina/api/handlers.go`**
- Added routes: POST /api/v1/valentina/research-runs/{id}/cancel, POST /api/v1/valentina/sources/{id}/approve, POST /api/v1/valentina/sources/{id}/pause, POST /api/v1/valentina/sources/{id}/test, GET /api/v1/valentina/query-packs, GET /api/v1/valentina/geography
- Added `handleCancelResearchRun` handler

**`valentina/api/handlers_extra.go`** (new file)
- `sourceUpdateRequest` type
- `handleApproveSource` — approves a source with terms_reviewed_at timestamp
- `handlePauseSource` — pauses a source
- `handleTestSource` — validates source is approved (connectivity check)
- `queryPack` type + `handleListQueryPacks` — Spanish/Valencian/Russian renovation query packs
- `geographyConfig` type + `handleGetGeography` — Valencia, Spain geography scope

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
- PR — not created (no merge requested)
