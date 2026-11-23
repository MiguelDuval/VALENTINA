# VALENTINA — fresh Cline bounded pass: secure hot intake

Task: VAL-INTAKE-001 security acceptance. Worktree: C:\Users\migue\Projects\VALENTINA-CLINE-INTAKE-20261123
Branch: feature/valentina-intake-security-20261123; baseline HEAD: 43c98e2751272f8411780a09cb8c8f58a71e2247.
Origin: https://github.com/MiguelDuval/VALENTINA.git. Pinned submodule f01c1d84c08b1de972d8cc268da25602598adff8.

Before edits, verify actual VS Code workspace, Git root, origin, branch, HEAD, status, pin. Read AGENTS.md, .clinerules, docs/AGENT-COLLABORATION.md, docs/AGENT-HANDOFF.md, docs/WORK-QUEUE.md, docs/SECURITY-AND-SECRETS.md, and the applicable ADR-006/ADR-007 from the separate existing PR #5 worktree if absent here. Inspect current API/domain/tests; do not copy unreviewed code from PR #5.

Objective: deliver one complete tested hardening slice in the existing Go + SQLite + Vue architecture.
1. For signed website intake, persist a safe request fingerprint with the idempotency key. Same key + same payload returns original lead (including after reopen); same key + different payload must return a clear 409, with no second lead and no disclosure of contact content. Handle concurrent duplicates, retention and existing idempotency rows safely. Tests must be deterministic and synthetic.
2. Review HMAC verification, replay/timestamp window, body size and consent/version validation, sanitized logs and fail-closed behavior. Fix definite vulnerabilities. Any signature change requires an explicit versioned compatibility path and tests; do not silently break installed clients. Privacy notice version is explicit metadata, never inferred marketing consent.
3. Investigate v11 migration collision: this branch has v11 privacy_notice_version; Draft PR #5 has a different v11 research_runs.updated_at schema (and later v12). Provide forward-only, safe migration logic and tests for clean DB, existing branch-v11 DB, PR-v11 DB where feasible, preserving data. Do not merge PR #5 or blindly run duplicate ALTER TABLE statements.
4. Verify synthetic signed request -> durable lead -> list/detail -> human triage plus persisted replay outcomes, and keep offline RSS/sourcepolicy tests green. Add targeted Vue tests only if relevant. Do not connect external feeds or messages.
5. Update docs/AGENT-HANDOFF.md and docs/WORK-QUEUE.md with exact evidence and residual blockers.

Acceptance: clean pinned Huginn archive tests go test ./internal/valentina/... and server compile; relevant frontend tests and typecheck, Windows Full build when possible (if npm EACCES persists report exact blocker). Check full diff and security-sensitive strings; stage only intended tracked first-party files and task document, commit only if verified; normal non-force push only after successful checks. CI success must be independently confirmed, not guessed.

Hard boundaries: no main, reset/rebase/stash/clean/force-push, no release, no subagents, no automatic outreach, no real data/secrets, no public binding, no permissions changes, no second runtime/app/database, no modification of upstream/huginn. Canonical C:\Users\migue\Projects\VALENTINA is dirty and MUST remain untouched. Other worktrees untouched. Do not retry old failed Cline session or duplicate Codex work. Implement in one bounded pass, not just a plan.
