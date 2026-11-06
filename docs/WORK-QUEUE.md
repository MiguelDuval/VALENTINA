# VALENTINA — work queue

This queue is the product-level source of task priority. Update status only when repository evidence supports the change.

## P0 — establish a reliable audit baseline

| ID | Task | Status | Owner / executor | Exit condition |
|---|---|---|---|---|
| VAL-P0-001 | Complete full Phase 0 feature inventory and safe cleanup audit | READY_FOR_REVIEW | ChatGPT directs; Cline audits | Inventory, cleanup plan and permission matrix committed; no application code changed; reviewed by ChatGPT |
| VAL-P0-002 | Reconcile existing PR #1 and PR #2 dependency/merge sequence | BLOCKED | Owner + ChatGPT | Owner decides merge sequence; neither PR is retargeted or merged by an agent without authorization |
| VAL-P0-003 | Verify Cline's local workspace, Git authentication and project-rule loading | READY_FOR_CLINE | Cline | Correct repository/branch/submodule verified; project rules visible in Cline |

## P1 — simplify the user-facing product (requires owner approval after P0)

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P1-001 | Implement VALENTINA task-first navigation | BLOCKED — P0 review and owner approval required | Overview, Opportunities, Companies, Research, Sources, AI & Automation, Settings and Diagnostics routes are real and tested |
| VAL-P1-002 | Hide unrelated developer/team/cloud surfaces from ordinary navigation | BLOCKED — feature inventory required | UI-only change, preserved routes/back-end, regression tests and screenshots |
| VAL-P1-003 | Disable inappropriate remote/cloud/collaboration capabilities safely | BLOCKED — configuration/security review required | Runtime behavior proves local-only and single-operator operation; no secrets or data loss |
| VAL-P1-004 | Establish least-privilege agent permission defaults | BLOCKED — permission matrix required | Scout/Verifier/Assessor permissions tested; no arbitrary shell, file-write, Git mutation or external actions for unattended research |

## P2 — build the renovation opportunity CRM (after scope and architecture review)

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P2-001 | Finalize domain contracts and persistence design | BLOCKED | Company, Opportunity, SourceEvidence, Source, ResearchRun and ReviewDecision contracts plus migration/backup design reviewed |
| VAL-P2-002 | Implement minimum in-process domain API and SQLite persistence | BLOCKED | Validated records, migration tests, deduplication and restart persistence verified |
| VAL-P2-003 | Implement real CRM list/detail/review screens | BLOCKED | Screens backed by real API/SQLite records, never mock-only claims |

## P3 — prove one permitted source end to end

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P3-001 | Select and approve one representative public source | BLOCKED | Terms/access review, exact host/path scope, allowed fields, rate limits and approval recorded |
| VAL-P3-002 | Implement source-scoped adapter and evidence extraction | BLOCKED | Server-side allowlist, redirect/DNS/IP/SSRF controls, timeouts and rate limits tested |
| VAL-P3-003 | Complete manual research vertical slice | BLOCKED | Discover → fetch → extract → validate → deduplicate → persist → human review demonstrated with real source evidence |

## P4 — AI integration and controlled automation

| ID | Task | Status | Exit condition |
|---|---|---|---|
| VAL-P4-001 | Verify NVIDIA NIM compatibility | BLOCKED | Real request, tool/function call and streaming tested; secret never exposed; model-specific capabilities documented |
| VAL-P4-002 | Connect VALENTINA source/database MCP tools | BLOCKED | Narrow schemas and read-only defaults; approval required for every side effect |
| VAL-P4-003 | Enable recurring research | BLOCKED | Only after manual workflow reliability, source compliance, idempotency and failure handling are measured |

## Execution rules

- Work one active task at a time unless the owner explicitly approves parallel tasks with separate branches and handoff files.
- A task cannot move from BLOCKED to READY without its dependency being satisfied.
- P0 audit deliverables are documentation only. Do not treat the queue as permission to begin P1 or P2 implementation.
- Cline updates the status and evidence when a task changes state. ChatGPT independently verifies the update.
- The owner retains merge authority unless explicitly delegated.
