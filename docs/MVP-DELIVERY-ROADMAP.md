# VALENTINA — MVP delivery roadmap

**Decision:** ADR-004, updated by owner-approved ADR-005  
**Status:** Execution baseline for a private, local, single-operator application  
**Date:** 2026-11-04  
**Primary outcome:** A polished VALENTINA-branded browser application and one simple Windows installer.

## 1. The actual MVP outcome

The first meaningful release is:

- One Windows 11 x64 `VALENTINA-Setup-x64.exe` installer.
- One local Go application process with its Vue browser client embedded.
- A user-facing VALENTINA identity: app logo, name, browser title, navigation, first-run and normal screens.
- A first screen focused on the renovation opportunity review queue, not a generic agent chat.
- Real local records for companies, opportunities, source evidence, research runs and human review decisions.
- Working native agent/model/connection/skill/workflow functions retained behind clear VALENTINA navigation.
- One small, real research pipeline that can add at least one source-verified opportunity to the review queue.
- Local-only operation, protected credentials and data-preserving upgrade/uninstall.

The user sees one VALENTINA app. The upstream runtime is the technical foundation, not a second product UI exposed alongside VALENTINA.

## 2. Architecture: reuse the existing application shell

The pinned upstream Go application already serves a full browser frontend. At the pinned commit:

- The frontend source is in `upstream/huginn/web/`.
- It uses Vue 3, TypeScript, Vite and Tailwind CSS.
- The shell and navigation are in `web/src/App.vue`, `web/src/router/index.ts` and `web/src/utils/navLayout.ts`.
- The frontend build outputs to `internal/server/dist`.
- The Go executable embeds those assets when built with `-tags embed_frontend`.
- Existing pages already cover chat/sessions, agents, models, connections, skills, workflows, settings, memory, inbox, logs and statistics.

**Decision under ADR-005:** adapt this existing web client into VALENTINA. Do not build a second React/HTMX/standalone CRM website and do not create a separate Go companion solely to serve a second UI.

The implementation should:

1. Keep the upstream submodule pinned and unchanged.
2. Store VALENTINA-owned patches/overrides and build scripts outside the submodule.
3. Apply the tracked patches to a clean temporary copy/worktree of the exact upstream revision.
4. Build the Vue frontend and then the Go binary with the embedded frontend.
5. Add the smallest required VALENTINA business routes and API handlers to this same application.
6. Add SQLite-backed domain records and scoped source/MCP tools in-process where practical.

A separate process requires a new ADR and a demonstrated blocker. The tracked patch process must fail closed if the expected upstream SHA changes or a patch no longer applies cleanly.

## 3. VALENTINA visual identity and navigation

The ordinary application experience must show VALENTINA, not upstream product branding.

Rebrand the app name, icon/logo, browser title, favicon, main navigation, tooltips, welcome/first-run screens, profile popover, notifications, status messages and normal settings chrome. Keep mandatory license notices in a clearly reachable **About → Licenses** screen; preserve the complete upstream LICENSE.

The root/default route must open **Overview / Review Queue**.

Primary navigation:

1. **Overview:** new candidates, needs-review count, qualified/rejected/expired counts, explicit-request versus early-signal totals, last successful research run and source failures.
2. **Opportunities:** filterable/sortable list with title, publisher/company, service type, demand-evidence class, municipality, location priority, deadline, fit status, evidence freshness and review state.
3. **Opportunity detail:** source-backed facts, original URLs, short evidence excerpts/locators, missing or contradictory facts, fit assessment and owner decision history.
4. **Companies:** public organizations linked to their separate opportunities and evidence.
5. **Research:** manual run controls, progress, run history, partial/failure states and deduplication summary.
6. **Sources:** approved/paused source registry, scope, last check and failures.
7. **AI & Automation:** existing agents, model/provider, connections, skills and workflow management.
8. **Settings:** local owner profile, NVIDIA NIM, privacy, local application behavior and backup/restore.
9. **Advanced / Diagnostics:** logs, usage statistics and technical controls not needed for the everyday task.

The main screens should use plain Russian labels and explain missing information without exposing the operator to internal architecture jargon.

## 4. Keep useful upstream capabilities; remove irrelevant surfaces from the normal UI

Preserve and expose the existing working features the project needs:

- AI agent creation, personas, tool permissions and memory controls.
- Model/provider configuration and connection status.
- MCP connections and relevant integrations.
- Skills used by VALENTINA agents.
- Workflows, manual execution, run history and inbox notifications.
- Logs and diagnostics for failure investigation.

Hide from the normal navigation, and disable where technically appropriate:

- Cloud relay, remote access and account connection flows.
- Multi-user, team, DM, channel and collaboration surfaces.
- Generic coding/developer workspaces unrelated to opportunity research.
- Community skill marketplace browsing and unrelated sample/demo content.

Do not delete backend packages, dependencies or shared functionality merely because a menu item is hidden. First prove they are unused and safe to remove. This is a single-user local application; it does not need roles, invites, collaboration or remote access.

## 5. NVIDIA NIM first-run configuration

Provide a clear VALENTINA settings flow using the existing custom/OpenAI-compatible backend where supported:

- Provider label: NVIDIA NIM.
- Default endpoint: `https://integrate.api.nvidia.com/v1`.
- Model ID: configurable and limited to a model available to the owner's account.
- API key: password-style input, saved only through the protected local credential store.
- Actions: Test connection, Save, and a non-secret connected/error status.
- Useful sanitized errors without tokens or authorization headers.

The existing generic custom-backend controls are a starting point, not evidence of a completed NIM integration. Verify a real normal request, tool/function calling and streaming with the exact selected model. Test vision only if a concrete workflow requires it. Keep mocks for offline tests.

## 6. Minimum technical backbone

Required for a usable vertical slice:

- Reproducible build of the pinned upstream Go application and its embedded Vue frontend.
- A tracked, ordered and tested VALENTINA patch set for UI branding and custom routes.
- Local browser UI bound to loopback only.
- A minimal domain API and SQLite persistence for Company, Opportunity, SourceEvidence, ResearchRun and ReviewDecision. Prefer an in-process module in the existing Go app.
- Strict schema/business-rule validation, stable IDs, canonical URL normalization and deterministic deduplication.
- Server-side source scope enforcement for approved source IDs, hosts, paths, redirects and resolved IPs.
- Scoped MCP tools for agents, with read-only discovery permissions.
- One approved source and one manual workflow: discover → retrieve original page → extract → validate → persist evidence/opportunity → show in UI → human review.
- Backup/restore and diagnostic logging without secrets.
- One per-user Windows installer.

A search result is not a verified opportunity. Distinguish explicit requests, public tenders, subcontracting requests and early public demand signals. An indirect signal must never be shown as a confirmed order.

## 7. What can wait

Do not block the first branded UI and installer on:

- Broad source coverage or a general-purpose crawler.
- Automated recurring research; manual Run Research is sufficient initially.
- Playwright, UFO² or browser automation unless the first approved source proves it is necessary.
- Advanced vision, OCR, model routing or multi-model debate.
- Contact-person CRM, phone/email harvesting, outreach or automatic form submission.
- Multi-user roles, remote access, cloud sync or collaboration.
- Advanced analytics, forecasting, report builders or mobile app.
- Auto-updater, code signing and public binary distribution.
- Perfect scoring weights, dozens of keywords or a huge golden evaluation set.

## 8. Execution stages

### Stage 0 — owner profile

Use the local profile under git-ignored `private/owner-profile.yaml`. It defines renovation services, demand-evidence categories, geography and ranking preferences. Do not commit the actual profile or use it in synthetic public fixtures.

### Stage 1 — upstream runtime baseline

Initialize and verify the pinned submodule. Build and run the unmodified Go application and frontend on Windows 11. Confirm local bind, port, startup/shutdown, data/config paths, tray/browser behavior and credential-store availability.

### Stage 2 — VALENTINA white-label shell

Create the tracked patch/build pipeline. Rebrand ordinary UI surfaces, make Overview the landing page, introduce the VALENTINA navigation and group existing engine views under AI & Automation / Advanced. Run frontend build, unit tests and Playwright smoke tests.

### Stage 3 — real business workspace

Add Opportunities, Opportunity Detail, Companies, Sources and Research routes to the same Vue app. Start with synthetic fixtures; connect each page to the real local API and SQLite-backed records before calling it functional.

### Stage 4 — one real source

Implement one approved, representative source and the server-enforced source-policy boundary. Keep research manual. Show source success/failure, evidence, freshness and review state. Do not enable unattended collection while raw unrestricted fetch access bypasses the source registry.

### Stage 5 — NVIDIA NIM

Implement first-run configuration through the existing settings/backend where possible. Verify actual request, tool-call and streaming behavior with a chosen account-available model. Never claim live success from mocked tests.

### Stage 6 — Windows installer

Build the single Go executable with the VALENTINA frontend embedded and package it using pinned NSIS tooling. Create one VALENTINA Start Menu entry. Preserve all mutable user data across upgrade/uninstall by default. Include complete required license texts and third-party notices.

### Stage 7 — owner-machine acceptance

Verify clean install, first launch, local-only access, NIM configuration, one real source run, opportunity review, restart persistence, backup/restore, upgrade and data-preserving uninstall on the owner's Windows machine.

## 9. Definition of done

The first usable MVP is done when the owner can:

1. Install VALENTINA using one Setup executable without installing Go, Node.js, Python, Docker or a database server.
2. Open a branded VALENTINA browser interface from the Start Menu.
3. Configure NVIDIA NIM without exposing the key.
4. Use existing agent/model/workflow features from VALENTINA's own navigation.
5. Run one approved source workflow manually.
6. See a real opportunity with its original source, evidence, dates, demand-evidence type, location priority and assessment.
7. Review and persist a human decision across restarts.
8. Back up and restore the local data.
9. Upgrade or uninstall without losing user data by default.

The first release may support one real source and a narrow opportunity set. Broad source coverage is a later expansion, not an excuse to delay the actual VALENTINA UI.
