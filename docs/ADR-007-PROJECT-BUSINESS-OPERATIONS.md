# ADR-007 — Local projects, estimates and document register

**Status:** Proposed implementation record for review; not a previously accepted ADR.
**Date:** 2026-11-24
**Authorization:** Owner's Gate 2 task authorizes this business block. This record documents implementation choices within ADR-005/006, without changing runtime, licensing or distribution scope.

Use the existing in-process Go domain/API, local SQLite and patched Vue application. No new service, framework, dependency, provider restriction or external action is introduced. The existing Huginn bearer wrapper protects every new route.

Projects are created manually, optionally linked to one active, human-qualified inbound lead. Qualification requires both current qualified status and a persisted human review decision. A unique source-lead key and transaction prevent duplicate conversion. An identical creation request returns the existing project; different data for an already linked lead conflicts. The original lead, status and review history remain unchanged. Creation identity remains stable after subsequent project edits. The UI resolves existing links before offering a new conversion.

Projects have editable descriptive fields and optional existing company/date/location references. Version checks prevent lost updates. Status changes use a separate validated endpoint: planned → active/cancelled; active → on_hold/completed/cancelled; on_hold → active/cancelled. Completed/cancelled are terminal. Cancellation replaces destructive deletion; there is no project/estimate deletion API.

Migration 14 appends projects, estimate snapshots, separate approvals and document metadata tables. Versions 1–13 remain unchanged, including v13 reconciliation of both historical v11 schemas and PR #5's v12. Existing Store ownership serializes transactions on its single SQLite connection. New reference existence and cross-project scope checks are explicit inside transactions, independent of legacy connection foreign-key settings.

Saved estimates are immutable revision snapshots; SQLite triggers reject updates/deletes. Every save creates a draft. The caller must supply the current latest revision number; stale or concurrent saves conflict. Any historical revision may be copied as the starting point of a new latest revision. Approval is a separate idempotent human action and never sends anything. Previous approved/draft snapshots remain readable. No invoice generation or tax-compliance claim is made.

Money is EUR cents, quantity is integer thousandths, and IVA is integer basis points entered by the operator. No tax rate or price is inferred. Round each line's net amount half up; group rounded line nets by VAT rate; round each group's VAT half up; sum net plus VAT. Arbitrary-precision intermediate integers prevent multiplication overflow. Inputs/totals are capped at the browser's exact integer limit (9,007,199,254,740,991), then returned as JSON integers. Quantity must be positive; zero-priced lines are allowed; negatives, empty estimates and invalid rates are rejected. At most 200 lines, rates 0–10,000 basis points.

Documents are explicitly **metadata only**: type, name, annotation, timestamp and optional estimate revision from the same project. No file upload, download, file bytes, filesystem path field or URL fetch exists. Names reject path separators and colons. Safe attachments require a separate scoped implementation and review.

Alternative considered: mutable quote rows or filesystem attachments in this pass. Immutable snapshots better preserve operator history; metadata provides useful organization without introducing an unverified file-storage boundary. All data stays in the existing private runtime database and its backups, outside Git. Tests use synthetic records only.

See [PROJECTS-AND-ESTIMATES.md](PROJECTS-AND-ESTIMATES.md) for operator/API behavior and [AGENT-HANDOFF.md](AGENT-HANDOFF.md) for verified results and remaining acceptance gates.
