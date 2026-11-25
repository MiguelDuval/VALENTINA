# Projects, estimates and documents

Open **Проекты** in business navigation. Create a project manually, or open an inbound lead, qualify it through the human review controls, and choose **Создать проект из заявки**. Existing links show **Открыть проект**. The project links back to the original lead. Conversion never edits the lead's original content or review history and does not infer contact details or prices.

The project card supports title, description, scope, existing company, place and optional calendar dates. Empty optional fields mean unknown. Status buttons show only valid next states. The API also enforces transitions and a version number; on a conflict refresh before editing again. Completed and cancelled projects remain available for history; hard deletion is not implemented.

In the estimate form enter labor/material lines, quantities (three decimal places), unit prices in EUR (two decimal places), and operator-selected IVA percentages (two decimal places). Dot or comma is accepted, but exponent notation and negatives are not. Click **Рассчитать**, inspect totals, then **Сохранить ревизию**. Changing an input clears the preview. Every saved revision starts as a draft; use **Утвердить эту ревизию** to approve that specific snapshot. There is no sending operation. Expand history to see the exact original lines, totals, timestamps and approval; **Взять за основу** copies a revision into the editable form without altering it. A new revision never inherits approval.

Rounding contract: `line_net = half_up(quantity_milli * unit_cents / 1000)`. Sum rounded line nets by VAT basis-point rate, then `group_vat = half_up(group_net * rate / 10000)`. Sum groups for totals. Example: 1.500 × 101 cents at 2100 bps gives 152 cents net and 32 cents VAT; another 1.000 × 100 cents at 1000 bps gives 100 + 10; total 294 cents. Rates are operator inputs, not Spanish tax advice. Estimates are not legally validated invoices.

The document register stores **only metadata**. Choose a type, name, annotation and optional saved estimate revision. It shows creation times and revision links. Files cannot be attached or downloaded. Names cannot be machine paths. Do not put secrets or unnecessary personal data in notes. Actual records remain in the existing private runtime SQLite database, outside the application installation and outside public Git.

## API contract

All paths below are relative to `/api/v1/valentina`; all require Huginn bearer authentication. JSON rejects unknown fields, trailing values and bodies exceeding 1 MiB. Storage errors are sanitized. No route is public or offered as an agent write tool.

| Method / path | Behavior |
|---|---|
| GET /projects?offset=0 | Stable descending page, up to 100 projects |
| POST /projects | Create planned project; optional `source_lead_id` enables qualified conversion |
| GET /projects/{id} | Project card, version and stable provenance |
| PUT /projects/{id} | Replace editable fields with matching `version`; source lead cannot change |
| POST /projects/{id}/status | `{status, version}` validated transition |
| GET /leads/{id}/project | `{project: object or null}`; original lead must exist |
| POST /projects/{id}/estimates/preview | `{base_revision, lines}`; deterministic totals, no persistence |
| POST /projects/{id}/estimates | Save draft with matching latest `base_revision` (zero initially) |
| GET /projects/{id}/estimates | Immutable snapshots, newest first, with separate approval status |
| POST /projects/{id}/estimates/{estimate}/approve | Empty object; idempotent explicit approval |
| POST /projects/{id}/documents | `{type,name,annotation,estimate_id}`; optional same-project revision |
| GET /projects/{id}/documents | `{documents,storage:"metadata_only"}` |

Project input fields: `title`, `description`, `service_scope`, `company_id`, `source_lead_id`, `location_text`, `start_date`, `end_date`. Only title is mandatory for manual creation; dates use YYYY-MM-DD. A retry with the same linked-lead creation input returns the same project (201); changed input conflicts (409). Use GET to resolve an existing link after editing the project.

Line fields: `kind` (`labor`/`material`), `name`, `quantity_milli`, `unit_cents`, `vat_bps`. All numeric fields are integers. Positive quantities, nonnegative unit prices, 0–10000 bps, 1–200 lines. Money/quantity/totals exceeding the exact JSON/browser integer bound are rejected. A stale project version or estimate base returns 409, invalid input 400, missing records 404. Document types: `estimate`, `contract`, `plan`, `note`, `other`.

## Verification and boundaries

First-party Go tests cover migrations from historical schemas, preservation/reopen, concurrent unique conversion, human-qualification requirements, project editing and transitions, integer arithmetic, VAT grouping, revision conflicts, immutable history and document scope. Vue tests cover forms, navigation, preview invalidation, history copying, approval, existing-link resolution and errors. The existing Windows runtime smoke also exercises the connected block, rejects unauthenticated requests for all 12 new routes and checks persistence after restarting the embedded EXE.

Use the existing pinned archive + four-patch build workflow. New first-party files must be tracked before the standard build script's `Copy-TrackedTree` includes them. Local uncommitted verification must explicitly copy every new file and compare the tested snapshot. No dependency/lockfile change is needed. Local checks do not establish GitHub CI, installer acceptance or live NIM/RSS acceptance. See the current handoff for exact evidence.
