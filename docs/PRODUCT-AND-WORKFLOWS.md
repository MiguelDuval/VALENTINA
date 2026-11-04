# Product operating model and workflow design

**Status:** Initial operating contract for the single-user VALENTINA MVP  
**Date:** 2026-11-04

## 1. Product definition

VALENTINA is a local, single-operator research system that discovers publicly available commercial opportunities relevant to an owner-defined business profile and geographic scope, verifies source evidence, removes duplicates, assesses possible fit and presents results for human review.

The product outcome is not "many search results". It is a useful queue of traceable opportunities where the operator can answer: what is the opportunity, why might it matter, who published it, which facts are confirmed or unknown, when was it published and checked, and what decision is pending?

It is not an autonomous sales agent, a mass-contacting system, a generic web crawler or a multi-tenant SaaS platform.

## 2. Business profile is a required input

The target business profile determines what counts as relevant. Do not hard-code a sector, services, opportunity categories, minimum value, radius or qualification threshold before the owner defines it.

The public repository contains only a placeholder at templates/owner-profile.template.yaml. Copy it locally to private/owner-profile.yaml and fill it in there. The private/ directory is git-ignored.

The profile should describe:
- Business offerings, capabilities, target customer types and sector keywords/synonyms.
- Geographic coverage: Valencia municipality, selected municipalities, province, Comunitat Valenciana, Spain or other expressly approved areas.
- Ordered geographic priority bands and travel-accessibility preferences; bands are ranking preferences unless the owner explicitly marks an area as a hard exclusion. Unknown route information stays unknown.
- Opportunity categories to include and exclude.
- Minimum/maximum commercial values or project sizes, if applicable.
- Hard disqualifiers, capacity limits and competitor exclusions.
- Languages to search and preferred summary language.
- Qualification criteria, relative importance and required evidence.
- Source categories or publishers expressly approved or excluded.

Unknown fields stay unknown. Do not invent a profile from generic assumptions.

## 3. User-facing workflow

1. **Define scope:** select the profile, geography, categories, time window and optional source restrictions.
2. **Discover:** search approved public sources and collect candidate URLs. A search hit is not a verified opportunity.
3. **Classify demand evidence:** distinguish an explicit request, public tender, subcontracting request and early public demand signal. Keep indirect signals in a separate visible class; never present one as a confirmed client request or awarded job.
4. **Fetch:** retrieve the original publisher page through the least fragile permitted method.
5. **Extract:** map relevant facts into the versioned VALENTINA data contract; preserve original language and source links.
6. **Verify:** inspect the original source; verify publisher, dates, location, eligibility, value and deadline. Missing facts remain unknown.
7. **Normalize and deduplicate:** deterministic normalization and matching first; AI can suggest ambiguous duplicate pairs for human review.
8. **Assess fit:** compare verified facts to the owner profile and explain both matches and missing information.
9. **Review:** present evidence, fit assessment, freshness and duplicate links. The owner decides what to qualify, reject, archive or follow up.
10. **Learn:** record why an item was accepted/rejected and use those labels to improve search terms and synthetic evaluation fixtures.

The first implementation should be manual-triggered. Enable recurring schedules only after the source process proves reliable and its costs and failure behavior are understood.

## 4. Agent roles

Huginn already supports custom per-agent prompts, models, Skills, local tool allowlists and external connection allowlists. Start with three small roles; do not add more agents without a distinct measurable need.

### VALENTINA Scout

Purpose: discover candidate public pages within an explicit approved scope.

Allowed local tools: web_search (only when Brave Search is configured) and fetch_url.

Forbidden: shell, file writes, Git mutation, outbound messages, browser-profile access or arbitrary MCP tools. A search snippet is never proof.

Output: candidate URL, publisher, title, short inclusion reason, visible date/deadline hints and facts requiring verification.

### VALENTINA Verifier

Purpose: check facts against the exact original publisher pages supplied by Scout.

Allowed local tool: fetch_url only.

Forbidden: widening the search, writing to files/databases, outbound contacts or arbitrary shell execution.

Output: source-backed facts, exact source URL, concise evidence and explicit missing/conflicting facts.

### VALENTINA Assessor

Purpose: compare verified records with the owner profile.

Allowed local tools: none by default. It receives the profile criteria and verified facts from earlier workflow steps.

Output: fit classification, matched/unmet criteria, missing facts, evidence quality and human-review priority. It recommends; it does not make irreversible decisions.

### Human operator

The owner remains the only authority for changing the profile/source allowlist, merging uncertain duplicates, accepting/rejecting candidates, editing business records, contacting organizations, submitting forms or performing any external action.

## 5. Use Huginn's existing workflow features

Prefer native Huginn features before writing code:
- Custom agents for role-specific prompts, model selection, Skills and permissions.
- Prompt-only Skills for reusable source, extraction and review rules.
- Workflows to sequence steps and pass outputs through {{prev.output}}, trigger inputs through {{inputs.x}}, and run scratch state through {{run.scratch.K}} / set_scratch.
- Per-step model overrides, conditional execution, sub-workflows, retries and run replay/diff where they fit.
- Inbox and run history for visibility.
- Artifacts for reviewable outputs, not as the permanent VALENTINA business database.

Keep the first workflow sequential, disabled and manual. Do not enable recurring schedules or write-capable agent templates by default. A workflow output is a handoff, not proof that the content is true.

Pinned upstream references:
- https://github.com/scrypster/huginn/blob/v0.5.0/docs/features/custom-agents.md
- https://github.com/scrypster/huginn/blob/v0.5.0/docs/features/skills.md
- https://github.com/scrypster/huginn/blob/v0.5.0/docs/features/workflows.md
- https://github.com/scrypster/huginn/blob/v0.5.0/docs/features/permissions.md

## 6. Deterministic code vs AI

Use deterministic code for URL/host allowlists, redirects and SSRF validation, response limits, date parsing, schema validation, canonical URLs, identifiers, exact-match deduplication, expiry, rate limits, retries, bookkeeping, export and backup. The pinned Huginn fetch_url does not enforce the VALENTINA source registry's host/path allowlists; the starter agent prompts are not a hard authorization boundary. The workflow must stay manual/disabled until a source-gating adapter enforces them server-side.

Use the model for search-query expansion, understanding public descriptions, mapping source text to a fixed schema, explaining fit against explicit owner criteria and suggesting likely duplicates.

Never let an LLM invent missing fields, silently alter evidence, delete/merge records, override source policy or perform external actions.

## 7. Model and cost strategy

NVIDIA NIM is the initial hosted model interface, not a model identity embedded in the data contracts.

- Keep model IDs configurable; use only models available to the account.
- Use a lower-cost capable model for broad classification/extraction when it passes the same evaluation set.
- Reserve stronger reasoning or vision-capable models for ambiguous cases and documents that genuinely require image/PDF understanding.
- Use step-level model overrides only after the exact model has passed structured-output/tool-calling tests.
- Cache source fetches and avoid repeating model analysis when the content hash has not changed.
- Record model ID, prompt/Skill version and latency for AI-assisted extraction.

Do not assume every NIM model supports the same vision, tool-calling, JSON or streaming behavior.

## 8. Explicit non-goals

The initial product does not send email/WhatsApp/social messages or applications automatically; bypass logins, CAPTCHAs, paywalls or rate limits; harvest personal profiles; mirror whole publisher websites; require a local LLM/GPU, Kubernetes, PostgreSQL or Redis; or become a general CRM for multiple organizations.

## 9. Definition of a review-ready result

A result must have a stable opportunity ID, factual title/summary, publisher/company reference where identifiable, at least one original source evidence record, first-seen/last-observed times, publication/deadline values when available, an owner-profile fit explanation, an explicit review state and notes for uncertainty or conflicting facts.

See DATA-AND-EVIDENCE.md, SOURCE-REGISTRY.md and QUALITY-AND-OPERATIONS.md. The manual workflow input shape is defined in schemas/workflow-input.schema.json and illustrated in templates/workflow-input.example.json.
