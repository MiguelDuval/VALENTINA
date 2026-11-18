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
- VAL-P2-002 In-process API/SQLite — DONE; opportunity evidence/history complete; inbound lead detail and secure intake implemented.
- VAL-P2-003 CRM UI — DONE; opportunity detail complete; inbound lead detail/review implemented.
- VAL-INTAKE-001 Secure hot-lead intake vertical slice — DONE; idempotent webhook contract, consent, spam controls, CSV import deferred, triage UI/API implemented; never expose loopback app.

## P3 — internet opportunity discovery (owner explicitly authorized 2026-11-18)
- VAL-P3-001 Source registry and technical policy gateway — DONE; source-specific allowlist, redirects/DNS/IP SSRF checks, quotas and fail-closed approval.
- VAL-P3-002 Search/provider and public-source adapters — DONE; official APIs/RSS/datasets first, bounded HTML only where allowed.
- VAL-P3-003 Manual research pipeline — READY_FOR_CLINE; discover → retrieve → extract → classify → validate → deduplicate → persist evidence → human review.
- VAL-P3-004 Research operations UI — READY_FOR_CLINE; source status, query packs, geography, run progress/cancel, errors and evidence.

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
