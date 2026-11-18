# VALENTINA — work queue

This queue is the product-level source of task priority. Status changes require repository evidence.

## P0 — audit baseline

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P0-001 | Phase 0 feature inventory and safe cleanup audit | DONE — owner accepted | Inventory, cleanup plan and permission matrix committed; owner authorized P2 |
| VAL-P0-002 | Reconcile PR #1 / #2 dependency sequence | BLOCKED | Owner decides merge sequence; no agent merge/retarget without approval |
| VAL-P0-003 | Verify local workspace, Git authentication and project rules | READY_FOR_REVIEW | Repository, submodule and handoff verified |

## P1 — product simplification

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P1-001 | Task-first navigation | BLOCKED — separate owner approval required | Overview, CRM, AI, Settings and Diagnostics routes tested |
| VAL-P1-002 | Hide unrelated developer/team/cloud surfaces | BLOCKED | UI-only change, preserved routes and regression tests |
| VAL-P1-003 | Enforce local-only, single-operator behavior | BLOCKED | Runtime/security tests prove local-only operation |
| VAL-P1-004 | Least-privilege agent permissions | BLOCKED | Permission matrix and approval tests |

## P2 — renovation opportunity CRM (owner authorized)

| ID | Task | Status | Verified progress / remaining exit condition |
|---|---|---|---|
| VAL-P2-001 | Domain contracts and persistence design | IN_PROGRESS | Domain models, SQLite schema and storage tests exist; backup/restore and deduplication review still required |
| VAL-P2-002 | In-process API and SQLite persistence | IN_PROGRESS | CRUD, scoped evidence/history reads, transactional review decisions, non-empty reasons, 1 MiB body limit and restart persistence tested; CI confirmation pending |
| VAL-P2-003 | CRM list/detail/review UI | IN_PROGRESS | Opportunities list/detail, evidence, review history, status filter/pagination, manual lead creation and five CRM sections exist; inbound-lead detail and remaining UX polish pending |

### Latest P2 checkpoint — 2026-11-18

- Commit: `22b0b190702729741ffc5139ef37dc7aaef2daa6` (local feature branch; push pending at handoff update).
- Full Windows build: PASS; 2,002 frontend tests, Vue typecheck, Vite production build, Go domain/API tests and upstream server compile passed.
- Windows EXE: 66,955,776 bytes; SHA-256 `2E1135B318863F6F419DDC0781DA5142B923DBC662EC0566354593C3512EA447`.
- Isolated runtime smoke: health 200, unauthenticated CRM 401, opportunity list 200, source evidence 200, missing review filters 400, missing opportunity 404.
- No GitHub Actions result exists for this unpushed checkpoint yet.

## P3 — one permitted source end to end

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P3-001 | Select and approve one public source | BLOCKED | Terms/access review, host/path scope, fields, rate limits and owner approval |
| VAL-P3-002 | Source-scoped adapter and evidence extraction | BLOCKED — depends on P2 | Allowlist, redirect/DNS/IP/SSRF, timeouts and rate limits tested |
| VAL-P3-003 | Manual research vertical slice | BLOCKED — depends on P3-002 | Discover → fetch → extract → validate → deduplicate → persist → human review |

## P4 — AI and controlled automation

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P4-001 | Verify NVIDIA NIM compatibility | BLOCKED | Real request, tool call and streaming tested without exposing secrets |
| VAL-P4-002 | Connect scoped source/database MCP tools | BLOCKED — depends on P2 | Narrow schemas, read-only defaults and explicit side-effect approval |
| VAL-P4-003 | Enable recurring research | BLOCKED | Manual workflow reliability, source compliance and idempotency proven |

## Execution rules

- One active implementation task at a time unless the owner explicitly authorizes parallel branches.
- P2 is owner-authorized. P1, P3 and P4 remain blocked until separate owner decisions.
- Cline/ChatGPT must update evidence when a task changes state; the owner retains merge authority.
- Never stage all files, delete user artifacts, force-push or rewrite history.
- A local build does not count as GitHub CI success; record the actual Actions run and artifact after push.
