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
- VAL-INTAKE-001 Secure hot-lead intake vertical slice — BLOCKED pending secret configuration and consent audit metadata. HMAC is fail-closed; durable SQLite idempotency (30-day key retention) and per-IP rate limiting are implemented. The endpoint remains disabled until a secret of at least 32 bytes is configured. Consent is validated and its server-side timestamp is persisted; the privacy-policy version is not yet persisted. Do not expose or advertise the webhook as production-ready.

## P3 — internet opportunity discovery (owner explicitly authorized 2026-11-18)
- VAL-P3-001 Source registry and technical policy gateway — PARTIAL; source approval gate (approved status + terms review timestamp + human reason), scheme/host/path/standard-port policy, redirect hop revalidation, all-address DNS checks, direct IP-pinned dialing, no environment proxy, bounded response bodies, and registry deep-copy isolation have regression tests. Per-source minimum-interval, daily and concurrency quotas are not yet enforced. No source is pre-approved.
- VAL-P3-002 Search/provider and public-source adapters — PARTIAL; RSS/Atom and official API adapters exist, parse Atom alternate links, reject non-2xx responses, and fail closed without a policy HTTP client. They are not connected to the research-run API; no public source is fetched.
- VAL-P3-003 Manual research pipeline — BLOCKED; API returns 501 until a real policy-gateway-backed runner retrieves, classifies, deduplicates and persists evidence. It must not create fake running records.
- VAL-P3-004 Research operations UI — PARTIAL; source approval, query packs and geography controls exist; run execution is visibly disabled until the runner is implemented.

## P4 — AI enrichment and controlled automation
- VAL-P4-001 NVIDIA NIM compatibility — BLOCKED until a local key is configured and live test approved.
- VAL-P4-002 Scoped MCP tools — BLOCKED until source gateway is enforced.
- VAL-P4-003 Recurring research — BLOCKED until manual pipeline, source compliance, idempotency and recovery pass acceptance.

## Execution rules
- Owner explicitly authorizes P2 hot inbound intake and P3 internet opportunity discovery as of 2026-11-18.
- This authorization does not approve any particular source, terms, crawling scope, credential, public endpoint, recurring schedule or external write action.
- One active task branch; do not use main.
- Never stage all files, delete user artifacts, force-push, reset, rebase, clean or rewrite history.
- A local build is not GitHub CI success. Record actual Actions evidence after push.
- Do not claim all-internet coverage. Prefer source diversity and verified contactability, not raw result count.
