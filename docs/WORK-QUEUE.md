# VALENTINA — work queue

This queue is the product-level source of task priority. Status changes require repository evidence.

## P0 — audit baseline
- VAL-P0-001 Phase 0 feature inventory and safe cleanup audit — DONE, owner accepted.
- VAL-P0-002 Reconcile PR #1 / #2 dependency sequence — BLOCKED, owner decision required.
- VAL-P0-003 Verify local workspace, Git authentication and project rules — READY_FOR_REVIEW.

## P1 — product simplification
- VAL-P1-001 Task-first navigation — BLOCKED pending separate approval.
- VAL-P1-002 Hide unrelated developer/team/cloud surfaces — BLOCKED.
- VAL-P1-003 Enforce local-only single-operator behavior — BLOCKED.
- VAL-P1-004 Least-privilege agent permissions — BLOCKED.

## P2 — renovation CRM and hot inbound intake (owner authorized)
- VAL-P2-001 Domain contracts/persistence — DONE; backup/restore and deduplication verified.
- VAL-P2-002 In-process API/SQLite — PARTIAL; opportunity evidence/history and inbound lead detail/review are implemented; external intake security acceptance remains open.
- VAL-P2-003 CRM UI — DONE for opportunity detail and inbound lead detail/review.
- VAL-INTAKE-001 Secure hot-lead intake vertical slice — IN_PROGRESS; HMAC fail-closed, durable SQLite idempotency (30-day key retention), per-IP rate limiting, consent + consent timestamp persisted. privacy_notice_version migration (v12) added to inbound_leads table, persisted through create/read/list, validated on intake payload, and shown in lead detail/create views. Offline tests added for version persistence, optional version, same-key-different-payload idempotency, and rate-limit/HMAC failure logging. Endpoint remains disabled until VALENTINA_INTAKE_HMAC_SECRET (>=32 bytes) is configured. Do not expose or advertise the webhook as production-ready.

## P3 — internet opportunity discovery (owner explicitly authorized 2026-11-18)
- VAL-P3-001 Source registry and technical policy gateway — PARTIAL; source approval gate (approved status + terms review timestamp + human reason), scheme/host/path/standard-port policy, redirect hop revalidation, all-address DNS checks, direct IP-pinned dialing, no environment proxy, bounded response bodies, and registry deep-copy isolation have regression tests. Per-source minimum-interval, daily and concurrency caps are implemented in the validated client; the manual runner also uses durable run-history budgets. Global quotas spanning future connector clients need hardening. No source is pre-approved.
- VAL-P3-002 Search/provider and public-source adapters — PARTIAL; RSS/Atom and official API adapters exist, parse Atom alternate links, reject non-2xx responses, and fail closed without a policy HTTP client. RSS/Atom has been connected to the manual research-run API with strict source preflight and candidate persistence. Official API/HTML/sitemap adapters are not wired to this runner. No real external source has been approved or verified end-to-end.
- VAL-P3-003 Manual research pipeline — PARTIAL: approved RSS/Atom-only manual runner, explicit feed-observed (unverified) candidates, provenance, conservative keyword filter, SQLite deduplication, completion/failure history and offline integration tests implemented. Requires an actually approved source, production artifact and end-to-end runtime acceptance. No unattended scheduling.
- VAL-P3-004 Research operations UI — PARTIAL; manually starts eligible approved RSS/Atom sources and shows runs. Cancellation/progress is not yet asynchronous; other source adapters remain deferred.

## P4 — AI enrichment and controlled automation
- VAL-P4-001 NVIDIA NIM compatibility — BLOCKED until a local key is configured and live test approved.
- VAL-P4-002 Scoped MCP tools — BLOCKED until source gateway is enforced.
- VAL-P4-003 Recurring research — BLOCKED until manual pipeline, source compliance, idempotency and recovery pass acceptance.

## P5 — project execution and document workspace (ADR-007, owner authorized 2026-11-20)
- VAL-P5-001 Client/project lifecycle, milestones, site visits and human-confirmed conversion of hot lead or qualified cold opportunity — READY_FOR_CLINE after P3 runner/intake acceptance.
- VAL-P5-002 Local documents and attachments: document registry, project/client links, versions, secure storage and backup/restore — PLANNED.
- VAL-P5-003 Versioned estimates/quotes: units, materials/labor, taxes, amendments, PDF export and explicit owner approval — PLANNED.
- VAL-P5-004 Job scheduling, checklists, site measurements/photos, handover and warranty tracking — PLANNED.
- VAL-P5-005 Expenses, payments due, project margin and business funnel metrics — PLANNED; legal/tax compliance validation required.

## P6 — messenger interface (ADR-007, owner authorized product direction)
- VAL-P6-001 Security ADR for paired operator identity, remote access architecture, scopes, command nonces/expiration, event audit and privacy — PLANNED.
- VAL-P6-002 Opt-in Telegram read-only notifications and status enquiries using official bot capabilities — PLANNED after security ADR and integration testing.
- VAL-P6-003 Opt-in WhatsApp Business channel where official API, provider onboarding and costs permit — PLANNED.
- VAL-P6-004 Confirmed remote commands limited to reviewed application actions, never shell/tool access or unsupervised customer messages — PLANNED.

## Cross-cutting fixes from audit 2026-11-20
- VAL-AUDIT-007 Reproducible clean tracked-source Windows full build; no untracked upstream/local build debris copied into product — BLOCKED pending validation.
- VAL-AUDIT-008 Correct disabled-intake API 501/503 test discrepancy — local Go tests passed after correcting assertion; tracked commit and CI remain pending.
- VAL-AUDIT-009 Update stale handoff and reconcile source quotas/API integration with actual current HEAD — READY_FOR_CLINE.
- VAL-P3-005 Source expansion and optional UFO²/browser capture for permitted social/messenger surfaces — PLANNED, not a claim of universal coverage.

## Execution rules
- Owner explicitly authorizes P2 hot inbound intake and P3 internet opportunity discovery as of 2026-11-18.
- This authorization does not approve any particular source, terms, crawling scope, credential, public endpoint, recurring schedule or external write action.
- One active task branch; do not use main.
- Never stage all files, delete user artifacts, force-push, reset, rebase, clean or rewrite history.
- A local build is not GitHub CI success. Record actual Actions evidence after push.
- Do not claim all-internet coverage. Prefer source diversity and verified contactability, not raw result count.
