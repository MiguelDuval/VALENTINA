# ADR-007 — VALENTINA as a renovation business operations hub

**Status:** Owner-approved product direction; integration designs pending security/technical validation
**Date:** 2026-11-20
**Scope:** Owner-operated renovation, finishing, repair and small-construction business

## Decision and hierarchy

VALENTINA is a private, local-first single-operator business workspace, not only an internet lead scraper. Four required pillars:

1. **Discovery (cold):** broad, source-compliant discovery of relevant local jobs, renovation, finishing and small building projects, across approved public listings, tender datasets, contractor calls, websites and permitted feeds. Multilingual query expansion, provenance, geographic ranking, confidence, source health, deduplication and human review. Exhaustive coverage cannot be guaranteed. Future UFO²/Playwright access to social networks or messenger UIs is optional, opt-in, permissioned and never a way to evade access controls.
2. **Direct enquiries (hot):** intake from the business's owned website, contact forms, email, official social/messenger integrations and manually entered communications. Preserve original message, source/channel, consent and privacy-notice version, contact history and attachments metadata. Authenticated, consent-aware and idempotent adapters; never expose the local loopback API as an unauthenticated public webhook.
3. **Documents and jobs:** an internal document workspace connected to client, lead, quote, project and site; create and version estimates/quotes, contracts, work orders, site notes, invoices, receipts and handover certificates, with attachment references, export and backups. The document module must manage client data locally, with retention, safe file validation and recovery. Legal/tax-compliant invoice formats for the owner's jurisdiction require separate validation; templates are not a guarantee of regulatory compliance.
4. **Messenger command center:** optional Telegram bot and WhatsApp Business integration for receiving notifications, querying VALENTINA state and submitting limited commands remotely. Implement only via official supported APIs, authenticated user/channel pairing, least-privilege actions, audit logs, expiring confirmation for sensitive commands, rate limits and explicit opt-in. Remote gateway design is a separate security ADR: keep the primary app bound to loopback, never expose its SQLite/administrative API or allow direct remote shell, file mutation, agent-permission change or unsupervised external outreach.

These pillars share business entities, but cold public signals MUST NOT be conflated with hot client requests.

## Additional essential business capabilities

- **Customer and project lifecycle:** enquiry → site visit → qualification → estimate → negotiation → accepted job → scheduling → progress → handover → warranty/follow-up; preserve status history and linked evidence.
- **Estimates and margins:** line items, units, labor/materials, tax assumptions, versioned quotes, acceptance and change orders; mark calculations as estimates until owner confirmation.
- **Calendar and capacity:** planned site visits, deadlines, crew/subcontractor availability, reminders, overdue follow-up, work schedule, without introducing multi-user remote access.
- **Field execution:** tasks/checklists, photos before/during/after, measurements, defects, punch lists, approvals and handover documentation.
- **Financial oversight:** costs, purchase receipts, expected/actual project margin, deposits, payment status, overdue invoices; no autonomous payments or promises.
- **Operational analytics:** source-to-qualified-lead conversion, response latency, quote win rate, backlog, project profitability, source cost/limits, failures and data freshness.
- **Optional communications:** response drafts, reminders and customer-facing message previews, sent only after explicit user approval and channel consent.

## Architecture constraints

Preserve pinned scrypster/huginn, existing Go process, SQLite, Vue client, approved provider options and white-label strategy. Use domain services and versioned schemas in VALENTINA-owned code with tracked integration patches; no parallel CRM/frontend, cloud CRM or new persistence service by default. Add migrations that preserve existing records. All operational data, client identities, documents and credentials remain outside the public Git repository; encrypt or protect sensitive files at rest appropriately and protect backups.

UFO² is not a baseline dependency. Prefer official APIs and scoped connectors; a future isolated desktop agent needs per-source compliance controls, least privilege, a dedicated approved environment and explicit confirmation for any side effects.

## Security, consent and remote actions

Distinguish operational contact consent from marketing consent. Prevent cross-channel identity confusion, replay, prompt injection, arbitrary tool invocation and exfiltration. Define RBAC-style permission scopes even though this is single-operator: read-only reports, proposed actions, local-only sensitive actions. Each accepted remote command records channel, verified operator identity, nonce, correlation ID, timestamp, action, decision and outcome with redacted PII. Telegram/WhatsApp integration costs, availability and business terms must be checked before activation.

## Delivery gates

**Gate 0:** source-controlled documentation and verified code baseline; protect dirty workspace.
**Gate 1:** repair reproducible clean build/test and manual source-run → evidence → review; secure intake → hot lead → triage.
**Gate 2:** project/customer workspace, document registry, estimate versioning, follow-up and scheduling.
**Gate 3:** approved messenger query/notification adapter in read-only mode; secure command approval and audit before any write operations.
**Gate 4:** expand approved discovery sources, optional browser automation/UFO², measurement-driven search quality and recurring execution after safety acceptance.

Acceptance requires demonstrated UI/API/persistence integration, relevant tests, restore testing and an actual Windows artifact, not screenshots or placeholder UI. No GitHub Releases; executable and installer only as successful GitHub Actions run artifacts.

## Relationship to existing decisions

This ADR extends ADR-006 (dual pipeline) rather than replacing it. Earlier MVP restrictions deferring documents, advanced analytics or messenger integration remain appropriate sequencing for the immediate vertical slice, but are no longer permanent product non-goals. Remote messenger access must be isolated and separately reviewed before changing the local-only network boundary.
