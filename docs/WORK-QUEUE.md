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
- VAL-INTAKE-001 Secure hot-lead intake vertical slice — LOCAL CHECKS PASS 2026-11-23. HMAC-v2 authenticates version/timestamp/body with 300-second skew; legacy v1 requires explicit local opt-in. Signed concurrent retries, changed-body conflict, retention, SQLite reopen and v11/v12/v13 compatibility tested. Local Windows HTTP runtime smoke PASS after correcting the route wrapper: HMAC-v2 first 201, duplicate 200, changed body 409, invalid/stale/future 401, SQLite persistence after restart, other CRM routes 401, loopback-only. A new automatic Windows CI smoke gate now tests the compiled EXE before artifact upload. Production remains blocked on owned-channel privacy/secret configuration, newly corrected CI evidence and owner acceptance; never expose a public webhook.

## P3 — internet opportunity discovery (owner explicitly authorized 2026-11-18)
- VAL-P3-001 Source registry and technical policy gateway — PARTIAL; source approval gate (approved status + terms review timestamp + human reason), scheme/host/path/standard-port policy, redirect hop revalidation, all-address DNS checks, direct IP-pinned dialing, no environment proxy, bounded response bodies, and registry deep-copy isolation have regression tests. Manual RSS/Atom admission now enforces persisted minimum intervals, UTC-day attempt caps and one active run per source atomically; no source is pre-approved. Unattended tools remain outside this acceptance scope.
- VAL-P3-002 Search/provider and public-source adapters — PARTIAL; RSS/Atom synthetic fetch is connected to the manual research API through a source-scoped policy gateway and tested offline. Official API and other adapters remain unconnected. No public source approval or live retrieval has been verified.
- VAL-P3-003 Manual research pipeline — admission/recovery implemented 2026-11-24 on `feature/valentina-research-admission-recovery-20261124`: atomic all-source SQLite reservation, duplicate rejection, durable attempt quotas, startup interruption recovery and bounded cancellation persistence. Synthetic concurrent HTTP/store, quota, rollback, UTC midnight, reopen and cancellation tests added. Final verification evidence is in AGENT-HANDOFF; real approved external source and desktop acceptance remain open.
- VAL-P3-004 Research operations UI — LOCAL CHECKS PASS 2026-11-24. Research shows sanitized busy/quota errors, stored status/counts/completion time, manual refresh and cancellation-request acknowledgement. Sources preserves approval reasons and displays nested limits. Nine Research/Sources tests pass within the 30-test targeted Vue suite. Evidence-only candidates remain unverified; real-source and desktop acceptance remain open.

## P4 — AI enrichment and controlled automation
- VAL-P4-001 NVIDIA NIM compatibility — BLOCKED until a local key is configured and live test approved.
- VAL-P4-002 Scoped MCP tools — BLOCKED until source gateway is enforced.
- VAL-P4-003 Recurring research — BLOCKED until manual pipeline, source compliance, idempotency and recovery pass acceptance.

## Gate 2 - project business operations (owner authorized)
- VAL-GATE2-001 Projects and human-qualified lead conversion - LOCAL QA PASS 2026-11-25; migration v14, Go API, Vue UI, authenticated runtime and browser verified.
- VAL-GATE2-002 Versioned EUR-cent estimates - LOCAL QA PASS; immutable revisions, grouped VAT rounding, separate approval and history verified.
- VAL-GATE2-003 Document register - LOCAL QA PASS; metadata only, same-project estimate reference. Attachments outside scope.
- VAL-GATE2-004 Publication and acceptance - PENDING exact-commit Windows Actions runtime smoke and SHA-labelled EXE. Live RSS/NIM requires owner authorization; no production acceptance claimed.

## Execution rules
- Owner explicitly authorizes P2 hot inbound intake and P3 internet opportunity discovery as of 2026-11-18.
- This authorization does not approve any particular source, terms, crawling scope, credential, public endpoint, recurring schedule or external write action.
- One active task branch; do not use main.
- Never stage all files, delete user artifacts, force-push, reset, rebase, clean or rewrite history.
- A local build is not GitHub CI success. Record actual Actions evidence after push.
- Do not claim all-internet coverage. Prefer source diversity and verified contactability, not raw result count.
