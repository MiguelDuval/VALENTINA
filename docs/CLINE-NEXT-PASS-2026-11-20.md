# Cline assignment — VALENTINA verified vertical-slice completion

**Assigned:** 2026-11-20; **State:** PARTIALLY_IMPLEMENTED_BY_TECHNICAL_LEAD — do not reimplement the manually tested source files in this feature branch; resume only after verifying new branch/HEAD and CI
**Repository:** MiguelDuval/VALENTINA
**Original base branch:** feature/valentina-research-intake-engine-20261118
**New implementation branch:** feature/valentina-manual-rss-runner-20261120
**HEAD:** 1904cdb9f3d91ba87b4de1de84acd1ffc05a425c
**Required upstream pin:** f01c1d84c08b1de972d8cc268da25602598adff8

## Read first and guard

Read AGENTS.md, .clinerules/00-valentina-project.md, docs/AGENT-COLLABORATION.md, docs/ADR-007-BUSINESS-OPERATIONS-HUB.md, docs/AUDIT-2026-11-20.md, docs/WORK-QUEUE.md, docs/AGENT-HANDOFF.md and security/architecture/source docs. Verify actual VS Code/Cline workspace root, git remote, branch, HEAD, dirty files, submodule pin before edits. Use the currently active feature/valentina-manual-rss-runner-20261120 branch after git identity checks; never main. The initial RSS/Atom runner and clean-archive build are implemented locally as of 2026-11-20. Do not reopen an independent concurrent editing session while ChatGPT is still publishing this branch. The next independent pass should verify remote Actions, improve hot-intake privacy notice version and acceptance, and extend source adapters only after reviewing specific approved source terms.

There are existing user edits in docs/AGENT-HANDOFF.md and valentina/sourceadapter/provider_test.go, and many untracked files/temporary outputs; a nested untracked upstream/huginn/internal/valentina directory exists. The technical lead has added ADR-007, audit, business documentation, queue and test assertion changes that are still uncommitted. Preserve all. Do NOT stash, reset, clean, delete, overwrite, stage all, force-push or alter the submodule pin. Before selectively staging, inspect each file/diff and distinguish lead's new edits from prior unrelated user edits. Do not use raw upstream directory as proof of a reproducible build.

## Principal outcome of this pass

Deliver ONE reproducible, useful, verified **local manual discovery + secure inbound intake** product slice. The owner has authorized the broader business hub (ADR-007), but do not spread implementation across unfinished messenger/documents/finance modules until the foundation works.

### 1. Build and test integrity

- Fix tracked-source discrepancy between valentina/ code, disposable integration copy and upstream local untracked tree. Reproduce tests/build using clean temporary copy of pinned upstream and tracked first-party patches, without copying untracked upstream/integration leftovers. Do not modify/delete original local workspace artifacts. Verify relevant build/profile commands and use output directories that cannot overwrite owner data.
- Existing regression identified and corrected in valentina/api/handlers_test.go: disabled HMAC intake returns 503 and should be asserted as StatusServiceUnavailable; retain safe disabled-by-default behavior.
- Run targeted Go tests, applicable frontend tests, production build and at least compilation of complete single-process Windows EXE. Record exact commands and results; if full compilation blocked, show the first root cause and make coherent fix.

### 2. Real manual discovery runner

- Replace current POST /api/v1/valentina/research-runs 501 placeholder with policy-gateway-backed manual runner only when the full path is implemented. Otherwise preserve fail-closed 501 and document blocker, never create fake successful runs.
- Select only explicitly approved source profiles (no pre-approved sources), use source-scoped API/RSS/Atom/sitemap/HTML adapters, supported parsing and real evidence (publisher URL, observation timestamp, original document location).
- Persist ResearchRun progress, partial/failure/cancellation, candidates/Opportunity + SourceEvidence, deduplication and human review state. Rigorously separate direct demand, tenders/subcontracting and weak early signals.
- Integrate API and existing Vue Research/Sources/Opportunities UI. No mock success cards.
- Enforce all URL/path/DNS/IP/redirect/scope/budget/concurrency controls and no raw fetch_url bypass. Add tests for source state/SSRF/rate/limits and end-to-end synthetic local server fixture.

### 3. Secure inbound enquiry slice

- Review fail-closed HMAC signature, replay protection, idempotency, payload validation, honeypot, privacy-policy version, retention and safe operational logging. Add missing durable privacy notice version metadata and tests if reasonably possible.
- Prove local intake flow: signed input -> validated stored InboundLead -> hot queue/detail -> human review and status. Ensure no secrets in logs or URLs.
- Build adapter boundaries for owned website and future official email/messaging polling but do not invent live connections, create a public tunnel, or auto-message customers.

### 4. Documentation, handoff and future phases

- Keep ADR-007's four pillars as approved strategic scope: broad compliant cold search, owned-channel hot intake, business documents/projects, and separately secured Telegram/WhatsApp command center. Follow queue P5/P6 after foundation acceptance; UFO² is optional and must not evade site restrictions.
- Reconcile WORK-QUEUE statuses with actual implementation: per-source limits exist in gateway at HEAD, but full runner is still missing; correct any stale statements based on tests, not assumptions.
- Update docs/AGENT-HANDOFF.md with changed files, actual test/CI evidence, current HEAD, submodule pin, risks and next tasks.
- Commit ONLY reviewed relevant files on feature branch, push normally only after tests and git-status/diff audit. No official GitHub Releases; only successful GitHub Actions SHA-named .exe artifacts. Do not claim CI success before inspecting actual run.

## Definition of acceptable result

A real record from a permitted manually approved local fixture/source is shown in UI with provenance; a secured inbound test lead is stored and can be reviewed; restart persistence works; tests pass from tracked clean source; Windows build has verified output. If an external account or approval is required, keep that connector disabled and document exactly what remains.

## Work style

Take one large coherent pass with focused testing; fix root causes together rather than iterating blind retries. Do not treat previous ChatGPT changes as Cline's own, do not overwrite concurrently modified files, and do not edit another repository. Report exact evidence; never invent a test, build, artifact, commit or push.
