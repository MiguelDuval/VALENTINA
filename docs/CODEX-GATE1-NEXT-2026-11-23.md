# VALENTINA — Codex Gate 1 implementation pass, 2026-11-23

Repository: MiguelDuval/VALENTINA
WORKTREE: C:\Users\migue\Projects\VALENTINA-CODEX-GATE1-20261123
BRANCH: feature/valentina-gate1-security-ui-20261123
BASE SHA: 0fbdafd24b6b58b75d8000eee8a608133ab67cb9
PINNED HUGINN: f01c1d84c08b1de972d8cc268da25602598adff8

**Preflight:** Check Git root/origin/branch/HEAD/status/pin before changing anything. Read AGENTS.md, docs/AGENT-HANDOFF.md (first 20 lines), docs/WORK-QUEUE.md, docs/MVP-DELIVERY-ROADMAP.md (MVP/gates only), docs/SECURITY-AND-SECRETS.md, docs/ADR-006-DUAL-PIPELINE-CRM-INTAKE.md and applicable .clinerules. ADR-007 owner-approved in separate PR #5 worktree; Gate 1 means a working local cold research -> evidence -> human review, and authenticated hot intake -> persistence -> triage. Read relevant technical files just once per task segment; avoid huge outputs and repeated identical calls. Windows terminal is PowerShell: NO 'ls -la', '&&' (PS 5), '/dev/null'. Use native PowerShell or cmd /c. Cline old sessions are closed/idle; never retry or create Cline tasks.

**Implement in coherent complete blocks, not just a plan.** Two priorities:

A. Hot-intake security: replace body-only signature as primary with a documented versioned HMAC-v2 contract that authenticates timestamp/version/body. Reject stale/future timestamps (bounded clock skew), wrong/missing HMAC/v2, invalid or ambiguous timestamps, and key reuse with changed body. Preserve existing schema data and SQLite v11/v12/v13 compatibility. Legacy body-only v1 must be disabled by default and permitted ONLY through explicit opt-in local environment config with clear deprecation docs/tests; do not silently break existing integrations without a migration guide. Enforce signed request idempotency across process restart, exercise concurrency/replay/retention with synthetic data, sanitized API 4xx errors, no PII/token logs. Keep loopback binding. Do NOT enable public webhooks.

B. Research UI + API integration: inspect existing tracked Vue patch and Go research endpoints. Make the Research/Sources UI manual-run action genuinely available only for server-side approved and terms-reviewed RSS/Atom sources; show pending/disabled source reason and loading/error/result counts, evidence provenance and human-review semantics. If this is already implemented, improve tests/end-to-end runtime contract instead of duplicating it. No live fetches unless an owner-approved source already exists; use mocked fixtures and deterministic offline tests. Preserve single Huginn Go process + patched Vue UI and all native provider settings.

C. Acceptance automation: add bounded smoke tests for signed synthetic intake -> persisted lead -> review decision and manual approved fixture RSS -> evidence-backed candidate -> review. Verify restart/reopen where supported, no false 'verified' claims. Go tests in a disposable clean archive of EXACT pinned Huginn + first-party code, not dirty upstream; Vue targeted tests, typecheck, Windows embedded build if feasible. Do not run expensive full suites blindly; if npm EACCES recur, diagnose once and clearly report. Review diff and scan for secrets or real customer data.

**Deliverables:** code, tests, exact checks, short docs/AGENT-HANDOFF.md and WORK-QUEUE.md update; note remaining Gate 1 blockers. No unnecessary rewrites or new dependencies, no network exposure or unapproved external sources. No changes in canonical C:\Users\migue\Projects\VALENTINA, old CI worktree or PR #5. No commits/push until code checked, relevant tests green, diff clean; then commit scoped files and normal push of this branch, never main/reset/stash/rebase/force push/merge/release. If blocked, stop with concrete cause and leave changes for review, do not assume success.
