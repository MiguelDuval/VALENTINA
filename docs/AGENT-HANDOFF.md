# VALENTINA — active agent handoff

**Current status:** READY_FOR_REVIEW
**Active task:** VAL-P0-001 — complete Phase 0 application feature inventory and cleanup audit  
**Updated:** 2026-11-06

## 1. Repository snapshot from ChatGPT's GitHub inspection

This is a remote GitHub snapshot, not a claim about the owner's local working tree. Cline must independently verify all values before acting.

| Item | Observed value |
|---|---|
| Repository | MiguelDuval/VALENTINA |
| Default branch | main |
| Infrastructure preparation branch | chore/agent-collaboration-ci-20261106 |
| Parent implementation branch | feature/valentina-ui-foundation-20261104 |
| Parent branch HEAD at preparation start | b7a563da4ae282016c01dcad134a14761c1fee5f |
| Huginn upstream pin | f01c1d84c08b1de972d8cc268da25602598adff8 |
| Existing PR #1 | Open — architecture/security baseline, targets main |
| Existing PR #2 | Open — UI foundation, based on PR #1 branch |
| Latest observed UI Windows build | Run 16, success, for b7a563da4ae282016c01dcad134a14761c1fee5f |

PR #1 and PR #2 were both open and unmerged at the time of this snapshot. PR #2 is stacked on the PR #1 branch. Do not retarget, merge, close or rewrite either PR autonomously.

## Local development workspace — required before task execution

The owner does not yet have a confirmed local VALENTINA clone. The canonical Windows location to create/reuse is:

%USERPROFILE%\Projects\VALENTINA

Cline must derive the actual user profile path at runtime. Do not assume the current editor workspace is this repository. If Cline opens in Mozart or any other repository, leave it untouched and prepare the VALENTINA clone separately. Clone the chore/agent-collaboration-ci-20261106 branch with recursive submodules, verify the exact origin and upstream pin, and verify the actual Cline editor workspace root. If the client cannot safely switch workspaces, stop after clone preparation and ask the owner to reopen this folder as a new Cline workspace.

## 2. Active assignment — VAL-P0-001

Complete the full Phase 0 Product Simplification Audit from the owner-provided task.

### Scope

Inspect and inventory:
- All frontend routes, navigation entries, views and shared components.
- Backend packages, API handlers and runtime capabilities.
- Built-in tools and permission tiers.
- Agent tool allowlists and templates.
- MCP client/server configuration and connections.
- All model providers/backends, model discovery and selection, provider-specific endpoints, independent API-key/credential settings, local/hosted inference options, per-agent provider/model assignment, and provider-specific capabilities (streaming, tool/function calling, vision, etc. where supported). Treat all currently implemented options as KEEP; do not hide, disable or remove any provider path.
- Skills, Skills Registry and workflow templates.
- Workflows, routines and scheduler.
- Chat, sessions, spaces, channels and collaboration features.
- Code Intelligence and workspace indexing.
- Cloud relay, remote access and external mutation capability.
- Logs, inbox, artifacts, persistence, configuration and diagnostics.
- Existing tests, CI coverage and Windows build path.

### Deliverables

Create:
1. docs/FEATURE-INVENTORY.md — feature-by-feature KEEP / KEEP BUT MOVE / HIDE / DISABLE / REMOVE LATER inventory, with exact paths, dependencies, risk and verification.
2. docs/CLEANUP-PLAN.md — phased UI, permissions and later-removal plan. Explicitly separate hiding, disabling and source deletion.
3. docs/SECURITY-PERMISSION-MATRIX.md — capabilities by role: owner, interactive assistant, unattended Scout, Verifier, Assessor, source adapter and diagnostics.
4. Update this handoff with factual results and unresolved items.

### Hard restrictions

- No application-code changes.
- No deletion, hiding or disabling of runtime features.
- No changes to the Huginn submodule or pinned SHA.
- No edits to credentials, real databases or local configuration.
- No reset, rebase, force-push, history rewrite or direct main work.
- Do not treat a route or component as safe to delete merely because it is developer-oriented.

Documentation and test-inventory files are the only in-scope repository changes for this task.

## 3. Required Cline preflight

Before work, record:
- Local repository path and git remote -v.
- Current branch and HEAD.
- git status --short including untracked files.
- Current intended base branch and SHA.
- git submodule status and git -C upstream/huginn rev-parse HEAD.
- Whether the working tree contains owner changes that must be preserved.

If the local repository is not MiguelDuval/VALENTINA, stop. If the working tree is dirty, do not stash, reset or overwrite it. Report the blocker and preserve the work.

Read AGENTS.md, docs/AGENT-COLLABORATION.md, docs/ARCHITECTURE.md, docs/ADR-005-VALENTINA-PERSONAL-WHITE-LABEL-UI.md, docs/UPSTREAM-CANDIDATES.md, docs/SECURITY-AND-SECRETS.md, docs/PRODUCT-AND-WORKFLOWS.md, docs/DATA-AND-EVIDENCE.md, docs/SOURCE-REGISTRY.md, docs/QUALITY-AND-OPERATIONS.md, docs/MVP-DELIVERY-ROADMAP.md, docs/UPSTREAM-INTEGRATION.md and docs/OPTIONAL-INTEGRATIONS.md.

## 4. Task completion summary — VAL-P0-001

- **Task status:** READY_FOR_REVIEW
- **Local repository:** C:\Users\migue\Projects\VALENTINA
- **Branch:** audit/valentina-phase0-feature-inventory-20261106
- **Base branch / SHA:** chore/agent-collaboration-ci-20261106 / a2d67f8901c949f4d1012634b573c1e08044817c
- **HEAD SHA:** d72049e7351907145ebd68e20efce0b3d0fb6482
- **Huginn submodule SHA:** f01c1d84c08b1de972d8cc268da25602598adff8
- **PR URL / number:** (pending creation)
- **Files changed:** docs/FEATURE-INVENTORY.md, docs/CLEANUP-PLAN.md, docs/SECURITY-PERMISSION-MATRIX.md, docs/AGENT-HANDOFF.md, docs/WORK-QUEUE.md
- **Audit summary:** Complete source-level inventory of VALENTINA + Huginn upstream. 10 verified AI provider backends (ALL KEEP), 13 frontend routes, 13 views, ~48 backend packages, MCP/skills/workflows/scheduler inventoried. 6 VALENTINA domain entities in JSON schemas only (not implemented). Source registry enforcement gap identified.
- **Tests/checks run:** Workspace identity verification, remote ancestry check, submodule pin verification. No test execution (documentation-only task).
- **Checks not run and why:** go test, npm test, go build, Windows build, live NIM test — not run because no application code was changed and no credentials are available.
- **CI run URL / result:** NOT AVAILABLE (CI not triggered yet; will run on push/PR)
- **Known risks / unresolved questions:** Source adapter not implemented (gap identified); VALENTINA domain entities not implemented (schemas only); no runtime verification of any feature.
- **Next requested ChatGPT action:** Review the PR and all three deliverable documents. Confirm provider preservation (all KEEP), security matrix accuracy (current vs desired), cleanup plan scoping. Do not begin Stage 1 or any application-code modifications.

Do not include secrets, private owner-profile values, actual leads, contact details or local database contents.

## 5. ChatGPT review

**Status:** Pending.

ChatGPT will compare the deliverables with the actual source tree and architecture, then record acceptance or required corrections in the implementation PR/review thread. The task must remain documentation-only until the owner approves a cleanup phase.
