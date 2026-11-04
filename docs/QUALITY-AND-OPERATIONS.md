# VALENTINA quality, evaluation and operations

**Status:** Required quality contract for implementation and ongoing use  
**Date:** 2026-11-04

## 1. Quality before volume

Optimize for useful, verifiable opportunities rather than a large number of search results. A high-volume pipeline that produces stale, duplicate or unsupported records is a failure.

Do not hard-code quality thresholds before there is a representative owner-reviewed evaluation set. Record a baseline first, then define acceptance thresholds against actual use.

## 2. Golden evaluation set

Create a small golden dataset once the owner profile and first permitted source are selected. Public Git may contain only synthetic fixtures or sanitized examples. Actual leads, contacts, owner-specific rules and source payloads stay outside Git.

The golden set should include:
- A clearly relevant opportunity.
- An irrelevant item from a similar sector.
- A duplicate page with tracking parameters or a repost.
- Two different opportunities from the same company.
- An expired opportunity.
- Missing or contradictory deadline/value/location.
- A page that is unavailable, rate limited or too large.
- Prompt-injection text embedded in a page description.
- A search snippet that disagrees with the original source.

Check that unsupported facts are marked unknown instead of invented.

## 3. Metrics

| Metric | Meaning |
|---|---|
| Precision@K | Share of the first K surfaced results considered relevant by the owner |
| Golden-set recall | Share of known eligible examples found by the workflow |
| Duplicate rate | Share of surfaced records that repeat the same opportunity |
| Evidence coverage | Share of material facts linked to a source evidence item |
| Verification success | Share of candidate pages whose original source could be inspected |
| Freshness | Age of last observation and age of publisher publication/update date |
| Expired-result rate | Share of active results whose deadline has passed |
| Extraction validity | Share of outputs passing JSON Schema and business-rule validation |
| Cost per accepted opportunity | Search/model/fetch cost divided by owner-accepted results |
| Runtime/failure rate | Timeouts, blocked pages, parser failures and partial workflows |

Keep fact accuracy and business relevance separate. A record can be accurately extracted but commercially irrelevant.

## 4. Validation gates

Before an extraction is surfaced as verified:

1. Parse the model response as strict JSON when machine-readable output is required.
2. Validate it against the relevant schema.
3. Validate URL scheme/host/scope and reject unsafe destinations.
4. Check source evidence, timestamps, dates, currency and required fields.
5. Ensure material claims point to evidence.
6. Run deterministic duplicate and expiry checks.
7. Mark unknown/conflicting facts explicitly.
8. Retain raw model output only when operationally necessary and safe.

An LLM's self-reported confidence is not proof. Prefer observable evidence and deterministic validation.

## 5. Model/provider evaluation

Before a model is assigned to a workflow, evaluate the exact NIM model ID and account endpoint for:
- Ordinary chat completion.
- Structured JSON/schema adherence.
- Tool calling, if that workflow needs tools.
- Streaming, if the client relies on it.
- Vision/image or document input only when a source requires it.
- Spanish/Valencian source understanding where relevant to the selected profile.
- Repeatability, timeout behavior, cost and latency on the golden set.

Different NIM models may have different capabilities. Use mocks and synthetic responses for offline tests. A live NIM test is an integration test, not a prerequisite for every developer unit-test run.

## 6. Operations and scheduling

Huginn's scheduler runs inside the Huginn process. Scheduled routines/workflows run only while it is running; missed schedules are not automatically replayed. Workflow-level retries/run replay are not interchangeable with routine retry semantics. Headless scheduled agents must be read-only because write and shell tools are denied without a human approval context.

Policy:
- Start with manual workflows.
- Add a low-frequency schedule only after source reliability and request limits are known.
- Show partial and failed runs distinctly from successful runs.
- Never interpret a missing result as proof that no opportunity exists.
- Provide a visible last-successful-run and source failure status when the results UI is implemented.
- Pause schedules during maintenance or source-policy review.

## 7. Backup and recovery

Identify the actual Huginn data/config paths from the pinned source and Windows runtime before documenting a final backup script.

The eventual backup set includes Huginn config, agents, Skills, workflows/routines, user state, VALENTINA SQLite and WAL/SHM state if introduced, local source registry and owner profile, and any locally retained evidence files allowed by source terms.

Never store backups in public Git. Test restoration into a clean local profile before calling recovery verified. A backup never restored is not a proven recovery process.

## 8. Logs and diagnostics

Safe operational metadata may include run ID, workflow step, source ID, adapter version, timestamps, HTTP status, response size, duration, retry count, failure category, sanitized schema errors, model ID, prompt/Skill version, token counts and latency.

Never log API keys, bearer tokens, cookies, OAuth tokens, authorization headers, the full private owner profile, unnecessary personal data or entire copyrighted source documents.

## 9. Release gates

Before accepting a feature:
- Document behavior and scope.
- Test deterministic logic automatically.
- Evaluate AI behavior with fixtures/golden set where practical.
- Preserve least-privilege defaults.
- Make failure states visible and recoverable.
- Keep docs aligned with implementation.
- Review on a task branch and propose through a PR.

Installer releases must additionally meet DEPLOYMENT-AND-INSTALLER.md.
