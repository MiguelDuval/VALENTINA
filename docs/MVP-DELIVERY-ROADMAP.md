# VALENTINA — MVP delivery roadmap

**Decision:** ADR-004 — UI-and-installer-first delivery objective  
**Status:** Proposed execution baseline; owner business profile still required for domain-specific behavior  
**Date:** 2026-11-04  
**Primary outcome:** A useful, local Windows application with a purpose-built VALENTINA opportunity-review interface and a simple installer.

## 1. The actual MVP outcome

The first meaningful product release is not a crawler benchmark or a collection of disconnected agents. It is:

- A Windows 11 x64 application installed through one \`VALENTINA-Setup-x64.exe\`.
- A branded VALENTINA entry point that opens its own local opportunity-management interface.
- A working local data store for companies, opportunities, source evidence, review decisions and source-check status.
- A small, real research pipeline that can add at least one verified opportunity to the review queue.
- The existing Huginn UI available separately as the advanced agent/workflow administration console.
- Local-only operation with secrets protected and business data preserved through upgrades and uninstall.

The UI and installer are the visible deliverables. Search, verification, storage and model integration are the minimum reliable foundation that makes those deliverables real rather than a static mock-up.

## 2. Architecture recommendation for the VALENTINA interface

Huginn supplies agent orchestration and its administrative web UI, but it is not a domain CRM. Do not attempt to force its generic agent screens to become the primary company/opportunity workspace.

**Recommended design to implement after a short ADR review:** one small VALENTINA-owned Go companion application/service with:
- A local web UI rendered from embedded static assets/templates. Prefer server-rendered HTML and small amounts of vanilla JavaScript/HTMX before adding React or another build system.
- A single local SQLite database and explicit schema migrations.
- A minimal HTTP API used by the UI.
- A scoped MCP server/interface used by Huginn agents and workflows.
- The source-allowlist enforcement adapter, so the same service can reject unapproved source IDs, hosts, paths, redirects and unsafe resolved IPs.
- A small launcher/supervisor or installer startup integration that starts both the VALENTINA service and Huginn predictably.

This consolidates the two concrete VALENTINA-specific needs—domain CRM/storage and server-enforced source policy—into one local Go component rather than one service per website or a separate backend for every concern.

The exact process boundary remains an implementation ADR gate. The installer may contain two executables (Huginn and VALENTINA) even though the user sees one product and installs one package. Do not promise a single process or single application binary until startup/shutdown behavior and upstream integration are verified.

Do not modify the Huginn submodule directly or start a long-lived fork solely for cosmetic UI changes. If extending the existing Huginn web UI proves materially simpler, compare that approach against the companion-app design in a written ADR, including upstream update/merge cost.

## 3. Minimal UI scope

Build the first useful VALENTINA interface around five simple areas:

1. **Overview / Review queue:** new candidates, items needing review, qualified, rejected and expired counts; separate counts for explicit requests and early demand signals; last successful research run; visible source failures.
2. **Opportunities:** searchable/sortable list with title, publisher/company, service type, demand-evidence type, geography, profile-based location priority/accessibility when reliable, deadline, fit status, evidence freshness and review state.
3. **Opportunity detail:** factual summary, explicit-request vs early-signal classification and its evidence, criteria-by-criteria fit assessment, missing/contradictory facts, original source links, evidence excerpts/locators, observation history and owner review controls.
4. **Companies:** organization records linked to their distinct opportunities and source evidence.
5. **Sources and Settings:** approved/paused sources, last check and failures, local business-profile settings, model connection status and safe first-run configuration.

Do not build contact management, sales pipeline automation, mass outreach, complex dashboards, team permissions, remote access or multi-tenant features in this MVP.

Use synthetic fixtures for the first visual/data-binding work. Replace fixtures with SQLite-backed records before calling the interface functional.

## 4. Minimum technical backbone (only what the UI and installer need)

### Required

- Build and run the pinned Huginn release on the target Windows machine; confirm its actual per-user data/config paths, tray behavior, local binding and shutdown lifecycle.
- Decide and implement the VALENTINA companion service boundary.
- Implement server-side source allowlist enforcement before any unattended collection.
- Implement the v0.1 SQLite schema, migrations, backup/restore and deterministic validation for the core entities.
- Implement a single source adapter and one end-to-end path: discover → retrieve original page → extract → validate → store evidence/opportunity → display in UI → human review.
- Configure the NIM backend and test one exact model for the model operations actually used. Keep model IDs configurable and keys outside Git.
- Package the two runtime components (if two are required) behind one Windows installer and one Start Menu entry.

### Not required before the first UI/installer release

- Broad source coverage or a general-purpose crawler.
- Perfect recall or a large golden evaluation set. Start with a small synthetic test set and a few owner-reviewed real examples.
- Automated recurring research. Manual "Run research" is enough until reliability is measured.
- Advanced model routing, vision, document OCR or multi-model debate.
- Playwright MCP or UFO² unless the one selected source actually requires it.
- Contact-person records, email/phone harvesting, outreach or CRM communication history.
- Remote sync, cloud access, multiple users, cross-platform packaging or a hosted service.
- Advanced analytics, forecasting, charts, custom report designers or a mobile app.
- Auto-updater and public binary distribution.
- Paid code signing for the first private installation. An unsigned build may show Windows SmartScreen warnings; revisit signing if frictionless distribution becomes a requirement.

## 5. Execution stages

### Stage 0 — Owner inputs and product boundary

Obtain a minimum local owner profile:
- What the business does and what it offers.
- Which kinds of commercial opportunities are wanted.
- Which geography is in scope.
- Three to five inclusion/exclusion rules or examples, if known.

The legal business name, financials, customer list, private contracts and personal contact records are not required. A generic description is sufficient.

### Stage 1 — Runtime and integration baseline

Build/run the pinned Huginn release once on the target Windows system. Verify the startup path, embedded frontend, local bind, actual data paths, API behavior, OS credential availability and clean shutdown. Confirm the NIM custom backend with one configured model only if the first workflow needs model calls.

Do not begin broad source scraping or customize Huginn internals here.

### Stage 2 — Domain service and persistence

Implement the smallest VALENTINA Go service with:
- SQLite schema/migrations for Company, Opportunity, Source, SourceEvidence, ResearchRun and ReviewDecision.
- Strict input/schema validation, stable IDs, canonical URL normalization and deterministic deduplication.
- Source-ID-based fetch/search operations enforcing approved hosts/path scope, redirect checks, DNS/IP safety, per-source rate limits, timeouts and response caps.
- MCP tools for Huginn to submit candidates and retrieve approved source results.
- Local HTTP API for the VALENTINA UI.
- Backup/restore and diagnostic logging without secrets.

Do not store contact-person data in the MVP.

### Stage 3 — CRM UI (start with synthetic data, then connect SQLite)

Implement the five areas described in section 3. Use the provisional contracts and synthetic fixtures first. Connect the UI to the real local API/SQLite service before declaring it usable.

Keep a clear distinction between:
- Candidate discovered but not verified.
- Verified source facts.
- Model's assessment against owner criteria.
- Owner's final review decision.
- Explicit demand (direct request, tender or subcontracting request) versus an early public signal that may indicate future demand but is not itself a request or award.

### Stage 4 — One real source and one complete workflow

Select one approved, representative source after the owner profile exists. Prove the full path with a narrow source adapter. Keep the workflow manual. Add only the extraction and parsing needed for this source. A source error must create a visible run/source status, not silently disappear.

### Stage 5 — Windows installer

Build the pinned Huginn executable and VALENTINA component(s) in a controlled build environment. Use pinned NSIS tooling to create one per-user installer that:
- Installs application files separately from mutable data.
- Creates a VALENTINA Start Menu entry and predictable startup/shutdown behavior.
- Opens the VALENTINA interface; Huginn admin UI remains accessible as an advanced option.
- Preserves local configuration, secrets, workflows and SQLite data on upgrade.
- Preserves user data on uninstall by default.
- Does not install Go, Node.js, Python, Docker, WSL, a database server or local AI models.
- Does not put keys or real business data into the installer.
- Includes license notices and internal-use restrictions.

### Stage 6 — Acceptance on the owner's Windows machine

Verify a clean install, first run, local-only access, NIM setup, one source run, creation/review of an opportunity, restart persistence, backup/restore, upgrade and data-preserving uninstall. Distinguish built, unit-tested, integration-tested and manually verified behavior. No success claim without the corresponding evidence.

## 6. Owner information: minimum vs later

### Needed now

1. Business/sector and a plain-language description of the services or products offered.
2. The opportunity types that matter most (for example, tenders, contracts, projects, grants, partnerships, event work, supplier requests—only select what is relevant).
3. Geographic scope: Valencia city only, province, Comunitat Valenciana, all Spain, remote opportunities, or a defined combination.
4. A few hard exclusions or thresholds, if any (minimum value, type of customer, deadline, distance, eligibility, language).

A rough description is enough. Do not spend time writing a formal business plan.

### Can be decided later

- Detailed scoring weights and numeric priority formulas.
- Every possible sector keyword and synonym.
- A complete catalogue of dozens of sources.
- Fine-grained interface styling, logo, colors and animations.
- Exact recurring schedule and run frequency.
- Contact-person fields and outreach workflows.
- All advanced model IDs and vision capabilities.
- Multi-user roles, remote access and cloud synchronization.
- Installer auto-update and code signing.
- Import/export formats and advanced reports.

## 7. Defaults to use unless the owner corrects them

- One local operator on Windows 11 x64.
- UI and summaries in Russian; search sources may be Spanish, Valencian/Catalan and English.
- Currency EUR.
- No personal contact collection in v0.1.
- No automatic external messages or form submissions.
- Manual research launch; no unattended schedules initially.
- Huginn remains the advanced agent/admin console.
- Local SQLite; no remote database service.
- Public Git contains templates, synthetic tests and code only; real profile, source registry and business data remain local.

## 8. Definition of done

The MVP is done when the owner can:
1. Install VALENTINA through one setup executable without manually installing development runtimes.
2. Launch the branded local interface from the Start Menu.
3. Configure the local business profile and NIM connection safely.
4. Run the first permitted source workflow manually.
5. See an opportunity in the UI with original source, evidence, dates, extraction status and model assessment.
6. Review it, change its human decision status and see that decision persist after restart.
7. Back up the data, restore it, upgrade the app and uninstall it without losing user data by default.

The first release may support only one source and a narrow set of opportunity types. Broad search coverage is a later expansion, not a condition for delivering the first usable installed product.
