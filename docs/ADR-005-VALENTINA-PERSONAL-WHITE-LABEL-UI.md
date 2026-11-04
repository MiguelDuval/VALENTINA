# ADR-005 — VALENTINA personal white-label web application

**Status:** Owner-approved decision  
**Date:** 2026-11-04  
**Scope:** One private, local, single-operator Windows application  
**Supersedes:** Earlier proposals for a separate VALENTINA web frontend / Go companion as the primary UI

## 1. Decision

VALENTINA will use the pinned `scrypster/huginn` Go application and its existing embedded Vue web client as the single application runtime and browser UI foundation.

We will **rebrand and tailor that existing UI into VALENTINA**, not build a competing second frontend and not run a separate CRM website beside it. VALENTINA's business-specific screens and the smallest required domain APIs should be added to the existing application where practical.

The product is private and local to one operator. It is not intended to be marketed, sold, hosted, shared with customers, or distributed as a third-party service.

## 2. Verified upstream foundation

At the pinned upstream commit `f01c1d84c08b1de972d8cc268da25602598adff8`:

- The browser client lives in `upstream/huginn/web/`.
- It uses Vue 3, TypeScript, Vite and Tailwind CSS.
- Its existing shell, router and navigation are implemented in `web/src/App.vue`, `web/src/router/index.ts` and `web/src/utils/navLayout.ts`.
- The Go server serves the built frontend locally; the `embed_frontend` build tag embeds `internal/server/dist` into the executable.
- The current UI already has working management views for agents, models, connections, skills, workflows, settings, chat/sessions, memory, inbox, logs and statistics.
- The documented default browser endpoint is `http://127.0.0.1:8421` / `http://localhost:8421`.

This is a real reusable application shell, not just a screenshot or static theme. It does not, by itself, provide VALENTINA's renovation-opportunity CRM, evidence review or source registry.

## 3. Single application / no competing UI

The intended runtime is one local Go application process serving one embedded VALENTINA-branded browser interface.

- Do not create a second standalone web frontend, separate React/FastAPI stack, or separate Go companion solely to render the VALENTINA UI.
- Add VALENTINA routes, Vue views and small server handlers to the existing application where possible.
- If the domain workflow requires structured records, use a narrowly scoped SQLite-backed module in the same application process where practical. Reuse available upstream capabilities and dependencies; do not assume an existing table already matches the VALENTINA schema.
- Add scoped MCP tools/source-policy enforcement to the existing runtime if they can be implemented safely there.
- A second process is permitted only after a demonstrated technical blocker and a new ADR explaining why an in-process implementation is not viable.

## 4. Rebranding requirements

The normal user experience must present **VALENTINA** as the product identity.

Replace or remove upstream product branding from ordinary application surfaces, including:

- App name, browser title, favicon and application logo/mark.
- Main navigation, first-run/welcome experience and default landing page.
- Headers, footers, tooltips, profile popovers, notifications, restart/update banners and user-facing error messages.
- Product-specific help text and any prominent references in the normal settings flow.

The application must not display the upstream name or logo as its normal product identity. Internal identifiers, source paths, package metadata, API routes, diagnostics and developer documentation may retain upstream names where changing them would create unnecessary risk.

The legal license text is the explicit exception. Keep the original required license and copyright notices intact and make the license accessible from a clearly named **About → Licenses** screen. Do not remove, rewrite or conceal the BSL text to achieve white-label styling.

## 5. VALENTINA navigation and screens

Use the existing navigation shell and adapt it to a task-first layout. The default route should open the VALENTINA Overview / Review Queue, not a generic coding chat.

Primary product navigation:

1. **Overview** — new candidates, items to review, qualified/rejected/expired counts, latest research run and source failures.
2. **Opportunities** — searchable list with service type, demand-evidence class, municipality, deadline, location priority, evidence freshness and review status.
3. **Opportunity detail** — source-backed facts, original URLs, evidence excerpts/locators, missing/contradictory information, fit assessment and owner decision history.
4. **Companies** — public organization records connected to their opportunities.
5. **Research** — manual run controls, run history, progress and failures.
6. **Sources** — approved/paused source registry and source health.
7. **AI & Automation** — existing agent, model, connection, skill and workflow management screens, grouped coherently.
8. **Settings** — NVIDIA NIM configuration, local profile, privacy, application behavior and backup/restore.
9. **Advanced / Diagnostics** — logs, operational statistics and lower-frequency technical controls.

The first screen should be useful without requiring the user to understand agents, models, MCP, tokens or workflow internals.

## 6. Preserve valuable engine capabilities; hide irrelevant product surfaces

Keep the existing, working engine features available through VALENTINA's own navigation, especially:

- Agent creation, personas, tool permissions and memory controls.
- Model/provider configuration.
- Existing custom-backend/API credential controls where they prove compatible with NVIDIA NIM.
- Skills used by VALENTINA agents.
- Workflows, manual execution, run history, inbox and diagnostics.

De-emphasize or remove from the ordinary navigation where not required by this single-user product:

- Cloud relay, remote access and cloud-account connection flows.
- Team, multi-user, channel and collaboration surfaces.
- Generic coding/developer workspaces and code-intelligence screens that are unrelated to renovation opportunity research.
- Community skill marketplace browsing and unrelated demo/template experiences.

Hiding a screen is not the same as deleting its backend implementation. Do not rip out upstream code, dependencies, API handlers or shared components merely because a menu item is hidden. First prove that a component is unused and safe to remove. Disable remote/cloud capabilities for the local product rather than merely hiding their buttons if the runtime supports that safely.

## 7. NVIDIA NIM configuration

VALENTINA's settings must provide a clear, first-class model connection flow:

- Provider: NVIDIA NIM / OpenAI-compatible custom endpoint.
- Endpoint default: `https://integrate.api.nvidia.com/v1`.
- Model ID: user-configurable and validated against models available to the account.
- API key: password-style input; saved only through the protected local credential mechanism / OS credential store.
- Actions: Test connection, save configuration, show non-secret connection status and a useful sanitized error.
- No key in frontend bundles, localStorage, URL, logs, command-line arguments, Git, screenshots or installer resources.

The upstream custom-provider UI and API are a foundation, not proof that NIM is already integrated. Test the exact selected model for a normal request, tool/function calling and streaming before claiming compatibility. Test vision only if a workflow actually needs it.

## 8. Visual and interaction direction

The UI should feel like a polished, friendly, professional personal workbench rather than a generic developer-agent console.

- Russian is the primary UI language for the first usable release.
- Clear labels, readable typography, consistent spacing, high contrast and helpful empty/error/loading states.
- Make the most common task — reviewing a new opportunity — immediately obvious.
- Avoid decorative dashboard clutter, unexplained metrics and jargon in primary screens.
- Keep desktop-browser use on Windows as the first target; preserve reasonable responsive behavior.
- Apply a coherent VALENTINA design token system for colors, typography, surfaces, controls and status states. Do not scatter one-off hard-coded colors across components.
- Final palette/logo can be refined with visual screenshots and owner review; do not block structural UI work on branding perfection.

## 9. Reproducible upstream customization

Do not commit ad-hoc edits into the Git submodule or change its pinned SHA for routine UI customization.

- Store VALENTINA-owned patches, overrides and build scripts outside `upstream/huginn` in tracked first-party paths (for example, `patches/upstream-ui/` and `scripts/build-valentina.ps1`).
- Build from the exact pinned upstream commit in a clean temporary worktree/copy, apply the tracked patch set, build the Vue frontend, then build the Go executable with `embed_frontend`.
- Fail closed if the submodule SHA differs from the pin, the patch does not apply cleanly, or the frontend build fails.
- Keep patches small, ordered, documented and covered by tests. Re-check them whenever the upstream pin changes.
- Do not maintain a permanent independent fork unless a future demonstrated requirement justifies it in a new ADR.

## 10. License and permitted use

The pinned upstream LICENSE is BSL 1.1. Its Additional Use Grant explicitly permits production use for an individual/hobbyist/researcher/open-source project, an eligible small organization, **or use that is strictly internal and not offered as a product or service to third parties in any form**. The owner's declared scope is private, local, single-operator use and no third-party distribution or hosted service.

The license also requires conspicuous display of the license on each original or modified copy. Therefore:

- Do not use upstream logos or product branding as VALENTINA's identity.
- Preserve the complete pinned upstream LICENSE and required notices in the source/build/install.
- Provide a clearly reachable About → Licenses screen containing the complete license text.
- Do not publicly distribute the installed binary or offer access to it as a service without a fresh license review and any required separate permission.
- Re-check the exact license and use grant if the audience, distribution, hosting or business model changes.

This is an engineering record of the reviewed license text, not legal advice.

## 11. Acceptance criteria

The white-label shell is accepted only when:

- The application opens at the local browser URL and identifies itself as VALENTINA on all ordinary product surfaces.
- The default route is Overview / Review Queue.
- No upstream product name/logo is visible in normal navigation, page chrome, notifications, browser title or first-run flow; the mandatory license remains accessible in About → Licenses.
- Existing agent, model, connection, skill and workflow functions still work after customization.
- The user can configure the NIM endpoint/model/key without exposing the secret and can run the real compatibility tests.
- Irrelevant cloud/collaboration surfaces are absent from the normal UI and remote access remains disabled.
- New VALENTINA CRM screens are real routes backed by validated local data, not decorative mock-ups once marked functional.
- Frontend unit tests, production frontend build, relevant Go tests and Playwright smoke tests pass.
- A clean Windows 11 build/install/launch test is performed before any installer success claim.
