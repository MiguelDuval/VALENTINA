# VALENTINA — Phase 0 Cleanup Plan

**Task ID:** VAL-P0-001
**Repository:** MiguelDuval/VALENTINA
**Branch:** `audit/valentina-phase0-feature-inventory-20261106`
**Base branch:** `chore/agent-collaboration-ci-20261106`
**Base SHA:** `a2d67f8901c949f4d1012634b573c1e08044817c`
**Huginn submodule:** `f01c1d84c08b1de972d8cc268da25602598adff8`
**Status:** Documentation only — no application code changed
**Evidence level:** SOURCE VERIFIED (based on FEATURE-INVENTORY.md)

---

## Governing Principle

**Simplify VALENTINA's user experience without stripping down the AI platform underneath it.**

Preserve the full existing choice of AI providers, models, endpoints, credentials and compute backends. Preserve useful agent, MCP, Skills, Workflow, scheduling, API and diagnostic capabilities. Organize them intelligently, but do not silently remove or restrict existing functionality.

This plan describes **safe, staged simplification recommendations only**. No implementation changes are authorized by this task. Stages 1-3 describe navigation, settings, and permission reorganization. Stages 4-5 (selective disablement and source deletion) require explicit owner approval and separate ADRs before any implementation.


## Stage 0 — Inventory and Decisions

### Objective
Document the completed evidence gathering, architecture findings, risks and unresolved owner decisions from the Phase 0 audit.

### Scope
- All findings from docs/FEATURE-INVENTORY.md
- All findings from docs/SECURITY-PERMISSION-MATRIX.md
- Repository state at branch `audit/valentina-phase0-feature-inventory-20261106`

### Exact files or source areas
- `docs/FEATURE-INVENTORY.md` — complete feature inventory
- `docs/SECURITY-PERMISSION-MATRIX.md` — security and permission matrix
- `docs/AGENT-HANDOFF.md` — task handoff record
- `docs/WORK-QUEUE.md` — work queue status updates

### Dependencies
- FEATURE-INVENTORY.md must be complete and reviewed
- SECURITY-PERMISSION-MATRIX.md must be complete and reviewed

### Risks
- None at this stage (documentation only)

### Required tests
- Documentation review (internal consistency, cross-document references)

### Rollback strategy
- N/A (documentation only)

### Acceptance criteria
- All three deliverable documents exist and are internally consistent
- Handoff and work queue are updated
- Every verified provider/backend is explicitly classified KEEP

### Owner approvals
- Review of FEATURE-INVENTORY.md findings
- Review of SECURITY-PERMISSION-MATRIX.md findings


## Stage 1 — Navigation and Visual Simplification

### Objective
Organize the main application around the actual VALENTINA workflow (Overview / Review Queue as default route). Preserve advanced capabilities and existing runtime behavior.

### Scope
- Frontend navigation reorganization in `upstream/huginn/web/src/router/index.ts`
- Frontend navigation groups in `upstream/huginn/web/src/utils/navLayout.ts`
- Frontend shell in `upstream/huginn/web/src/App.vue`
- VALENTINA UI patch in `patches/upstream-ui/001-valentina-ui-foundation.patch`

### Exact files or source areas

| File | Current state | Recommended change | Decision |
|------|--------------|-------------------|----------|
| `web/src/router/index.ts` | Root redirects to `/chat` | Change root redirect to Overview/Review Queue (new VALENTINA route) | KEEP_BUT_MOVE |
| `web/src/utils/navLayout.ts` | NAV_GROUPS: chat, admin, logs | Reorganize: VALENTINA workflow first, AI & Automation / Advanced second | KEEP_BUT_MOVE |
| `web/src/App.vue` | Upstream branding (Huginn colors, "H" logo) | VALENTINA branding via patch | KEEP_BUT_MOVE |
| `patches/upstream-ui/001-valentina-ui-foundation.patch` | UI rebranding patch | Extend to cover navigation reordering | KEEP_BUT_MOVE |

### Dependencies
- New VALENTINA routes (Overview, Opportunities, Companies, Sources, Research) must be implemented in a later stage (P1/P2)
- Until then, Overview can use existing live APIs (health, agents, workflows, notifications) with honest empty state

### Risks
- Removing `/chat` as default route changes muscle memory for existing users
- Mitigation: Keep `/chat` accessible from navigation, just not as default

### Required tests
- Frontend unit tests (`web/src/views/__tests__/App.test.ts`, etc.)
- Visual review of navigation

### Rollback strategy
- Revert patch; router index.ts reverts to upstream default

### Acceptance criteria
- Default route is Overview / Review Queue
- VALENTINA branding visible on all ordinary product surfaces
- No upstream product name/logo in normal navigation
- Existing agent/model/connection/skill/workflow functions still accessible

### Owner approvals
- Owner approval required for navigation reorganization (ADR-005 governs UI changes)


## Stage 2 — Settings and Advanced Organization

### Objective
Improve the discoverability and organization of settings, including all existing AI providers, API credentials, custom endpoints, model selection, local inference, provider-specific capabilities, agent configurations, MCP, Skills, Workflows, Routines, and Diagnostics.

### Scope
- Settings view (`web/src/views/SettingsView.vue`, `web/src/views/settings/`)
- Models view (`web/src/views/ModelsView.vue`, `web/src/views/models/`)
- Connections view (`web/src/views/ConnectionsView.vue`)
- Skills view (`web/src/views/SkillsView.vue`)
- Workflows view (`web/src/views/WorkflowsView.vue`)
- Backend config (`upstream/huginn/internal/config/config.go`)
- Secrets management (`upstream/huginn/internal/secrets/`)
- MCP config (`upstream/huginn/internal/mcp/config.go`)

### Exact files or source areas

| Area | Current state | Recommended organization | Decision |
|------|--------------|-------------------------|----------|
| AI providers | All in Settings → Models | Group under "AI & Automation / Advanced" | KEEP_BUT_MOVE |
| API credentials | OS keychain (primary) + `~/.huginn/secrets.json` plaintext JSON with 0600 permissions (fallback, not application-encrypted) | Preserve all 11 slots; no consolidation | KEEP |
| Custom endpoints | Provider-specific | Keep per-provider endpoint configuration | KEEP |
| Model selection | Per-agent model ID | Preserve per-agent provider/model choice | KEEP |
| Local inference | Ollama/managed backends | Keep accessible, not hidden | KEEP |
| MCP | Settings → MCP connections | Move to Advanced area with approval gate info | KEEP_BUT_MOVE |
| Skills | Settings → Skills | Move to Advanced area | KEEP_BUT_MOVE |
| Workflows/Routines | Settings → Workflows | Move to Advanced area | KEEP_BUT_MOVE |
| Diagnostics | Logs, Stats views | Keep accessible; group under Advanced | KEEP_BUT_MOVE |

### Dependencies
- No provider may be removed, hidden, or made NIM-exclusive
- All 11 secret slots must remain independently configurable
- NVIDIA NIM via `custom` provider must remain configurable

### Risks
- Accidental hiding of provider settings could make them unusable
- Mitigation: "HIDE" means remove from ordinary navigation, not delete code or block access

### Required tests
- Frontend unit tests for settings views
- Configuration round-trip tests (load/modify/save)

### Rollback strategy
- Revert patch; settings views revert to upstream default

### Acceptance criteria
- All existing AI providers remain configurable
- All 11 secret slots remain accessible
- NVIDIA NIM via `custom` provider remains configurable
- Provider settings are reorganized but not removed or restricted
- Advanced area clearly labeled and accessible

### Owner approvals
- Owner approval required for settings reorganization


## Stage 3 — Permission and Configuration Hardening

### Objective
Describe future security improvements and the technical controls required.

### Scope
- Permission gate (`upstream/huginn/internal/permissions/permissions.go`)
- MCP approval gate (`upstream/huginn/internal/mcp/config.go`)
- Source registry enforcement (gap identified in FEATURE-INVENTORY.md §11.4)
- fetch_url tool (`upstream/huginn/internal/tools/fetch_url.go`)
- Secrets management (`upstream/huginn/internal/secrets/`)

### Exact files or source areas

| Control | Current state | Target state | Decision |
|---------|--------------|--------------|----------|
| Source registry enforcement | Policy only, not technically enforced | Server-side VALENTINA MCP adapter enforcing allowed_hosts/allowed_path_prefixes | OWNER_DECISION (new component) |
| Unattended fetch_url | Available to agents | Restricted to source-scoped adapter only | DISABLE_RECOMMENDED |
| `--dangerously-skip-permissions` | CLI flag available | Not available in production builds | DISABLE_RECOMMENDED |
| MCP approval gate | Default true (fail-closed) | Keep default true | KEEP |
| Secret storage | OS keychain + secrets.json fallback | Keep both; no plaintext in config | KEEP |
| Cloud/satellite connection | Available, disabled by default | Keep disabled in local product | KEEP_BUT_MOVE |

### Dependencies
- VALENTINA source-policy MCP adapter must be implemented before unattended collection
- Adapter must enforce host, path, redirect, DNS and destination IP validation

### Risks
- Raw fetch_url bypasses source registry allowlist
- Mitigation: Keep workflow manual/disabled until adapter exists

### Required tests
- Adapter unit tests for allowed/disallowed host, path escape, redirect validation, private/reserved IP, DNS rebinding, timeout, content size, paused/unapproved source
- Permission gate tests

### Rollback strategy
- N/A (new controls are additive)

### Acceptance criteria
- Source registry is technically enforced, not just policy
- Unattended agents cannot use raw fetch_url
- All permission defaults are least-privilege

### Owner approvals
- Owner approval required for any permission restriction
- New MCP adapter requires separate ADR


## Stage 4 — Selective Disablement

### Objective
Only where objectively justified and after explicit owner approval, disable specific runtime capabilities that are inappropriate for the VALENTINA product.

### Scope
No specific disablements recommended at Phase 0. Candidates for future owner review:

| Candidate | Reason | Required before disablement |
|-----------|--------|---------------------------|
| Cloud/satellite connection | Not needed for local single-operator app | Owner approval, security review |
| OAuth integrations (GitHub, Google, Slack, Jira, Bitbucket) | Not needed for renovation CRM | Owner approval |
| `--dangerously-skip-permissions` | Unsafe for production | Owner approval, build flag |
| Developer tools (bash, git, file write) | Not part of ordinary VALENTINA workflow | Owner approval, move to Advanced |

### Decision
All candidates are **OWNER_DECISION** — no disablement is recommended without explicit owner approval and separate ADR.

## Stage 5 — Source Deletion

### Objective
Only after separate dependency analysis, data/migration planning and regression testing, delete code that is no longer needed.

### Scope
No source deletion recommended at Phase 0.

### Decision
**No code deletion is recommended at this stage.** All existing functionality is preserved. Any future deletion requires:

1. Separate dependency analysis
2. Data/migration planning
3. Regression testing
4. Explicit owner approval
5. Dedicated ADR

## Decision Definitions

| Decision | Meaning | Authorized in this task? |
|----------|---------|-------------------------|
| KEEP | Preserve as-is, no change | Yes (documentation only) |
| KEEP_BUT_MOVE | Preserve functionality, reorganize UI location | Yes (recommendation only) |
| HIDE_RECOMMENDED | Remove from ordinary navigation, preserve code and access | Yes (recommendation only) |
| DISABLE_RECOMMENDED | Block or switch off runtime functionality | No (requires owner approval) |
| REMOVE_LATER_CANDIDATE | Delete code, integration, dependency or API | No (requires owner approval + ADR) |
| OWNER_DECISION | Requires explicit owner decision before any action | Yes (documentation only) |

## Cross-References

- [FEATURE-INVENTORY.md](FEATURE-INVENTORY.md) — complete feature inventory with all source paths
- [SECURITY-PERMISSION-MATRIX.md](SECURITY-PERMISSION-MATRIX.md) — security and permission matrix by role
- [ARCHITECTURE.md](ARCHITECTURE.md) — mandatory architecture baseline
- [ADR-005](ADR-005-VALENTINA-PERSONAL-WHITE-LABEL-UI.md) — UI rebranding decision
- [DEPLOYMENT-AND-INSTALLER.md](DEPLOYMENT-AND-INSTALLER.md) — deployment and installer policy
- [MVP-DELIVERY-ROADMAP.md](MVP-DELIVERY-ROADMAP.md) — MVP delivery stages
