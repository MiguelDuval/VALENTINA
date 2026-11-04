# VALENTINA source registry and onboarding policy

**Status:** Mandatory source onboarding contract  
**Date:** 2026-11-04

## 1. Why a registry is required

A search engine is not a source policy. Before VALENTINA collects data from a website, feed, portal or API, that source needs a reviewed profile.

The registry defines which source is allowed, its purpose, collection method, permitted fields, request frequency, and failure behavior. The public repository contains only the schema and placeholder template. The active registry belongs in local protected storage.

## 2. Source selection hierarchy

Choose the least fragile method available:

1. Official API or documented public export.
2. RSS/Atom, sitemap or published dataset.
3. Ordinary HTTP fetch plus deterministic parsing.
4. One narrow source-specific MCP adapter.
5. Playwright MCP only when a permitted source genuinely requires JavaScript rendering.
6. UFO² only for a justified Windows GUI-only task that cannot be handled through API, HTTP or browser integration.

Do not add a general crawler framework because one page is difficult. Do not run several browser/crawler stacks in parallel by default.

## 3. Source onboarding checklist

Every source profile must record:
- Unique source ID, display name, publisher and official homepage.
- Source type: official API, RSS/Atom, sitemap, HTML, PDF/document, dynamic page, dataset or other.
- Exact allowed hostnames and URL/path scope; never unrestricted arbitrary fetching.
- Official terms of use, robots/access guidance and date reviewed.
- Whether authentication is required. Credentials are configured separately and never stored in the registry.
- Permitted collection purpose/fields, prohibited fields, geography and date scope.
- Preferred retrieval method and adapter ID/version.
- Minimum request interval, parallel limit, daily cap, timeout and response-size limit.
- Expected update frequency, expiry/staleness policy and known failure modes.
- Approval state and reviewer/date.

A source remains pending_review until access terms, allowed scope and collection behavior are understood. A blocked source must never be scheduled or fetched.

## 4. Huginn built-in web tools: known limits

The pinned Huginn fetch_url implementation is a bounded public page fetcher. At v0.5.0 it uses a 10-second timeout, a 512 KB response read limit, HTML-to-Markdown conversion and SSRF protections for private/internal destinations and related blocked ranges. It is not a browser, does not execute JavaScript and is not a general crawler.

Huginn web_search uses Brave Search and is unavailable until a separate Brave API key is configured.

Therefore:
- Use web_search to discover candidate pages, not as final evidence.
- Use fetch_url only when the page fits its documented limits.
- Escalate to a source-specific adapter or Playwright only after the source requires it.
- Preserve failures as explicit states: blocked, timeout, too large, unsupported content, not found, rate limited, authentication required, parse failure or unknown.

Pinned source references:
- https://github.com/scrypster/huginn/blob/f01c1d84c08b1de972d8cc268da25602598adff8/internal/tools/fetch_url.go
- https://github.com/scrypster/huginn/blob/f01c1d84c08b1de972d8cc268da25602598adff8/internal/tools/web_search.go
- https://github.com/scrypster/huginn/blob/v0.5.0/docs/features/local-tools.md

## 5. Important enforcement gap in the current starter kit

The source registry and agent prompts are policy/configuration documents; they are **not a hard technical hostname allowlist in the pinned Huginn release**.

Huginn's built-in `fetch_url` blocks private/internal destinations and applies response/time limits, but the tool does not receive a VALENTINA source-profile ID and does not enforce this registry's `allowed_hosts` or `allowed_path_prefixes` fields. A prompt telling an agent to use approved hosts is not an authorization boundary, and prompt injection can attempt to make an agent fetch unrelated public URLs.

Therefore:
- The current starter workflow is manual-only and disabled by default.
- Do not run discovery/verification unattended or describe the registry as technically enforced while agents call built-in `fetch_url` directly.
- Before enabling unattended operation or treating source scope as guaranteed, implement one narrow VALENTINA fetch/search MCP adapter that resolves a source ID, checks the exact host/path allowlist, re-validates every redirect and resolved IP, enforces request/response limits and records the source ID. It must fail closed when a source is missing, paused or not approved.
- Keep raw `fetch_url` out of the unattended agent tool allowlist once the enforcing adapter is available.
- The adapter itself must have deterministic unit tests for allowed host, disallowed host, path escape, redirect to a disallowed host, private/reserved IP, DNS rebinding/re-resolution, timeout, content size, unsupported type and paused/unapproved source.

This is the concrete point where a small VALENTINA-owned MCP component becomes justified: not because Huginn lacks workflow orchestration, but because the native fetch tool cannot enforce VALENTINA's source-specific policy. Its implementation should be one small Go service, not a fleet of per-source microservices.

## 6. Source vs evidence

A Source describes an approved publisher/endpoint and collection rules. A SourceEvidence record describes one actual observation of one page/document at a particular time.

A generic source URL is not proof that a specific opportunity exists. Every candidate must retain its exact original page URL and observation time.

## 7. Source discovery order

The target business profile has not yet been fully specified. Therefore, no fixed list of tenders, directories, marketplaces, job boards, event sites or commercial portals is declared approved here.

Once the local profile is completed, evaluate candidates in this order:
1. Official public-sector and municipal/regional portals relevant to the defined categories.
2. Official procurement, tender, grant or funding portals only if those categories are relevant.
3. Official associations, trade bodies and chambers relevant to the target business.
4. Public company/venue/event/marketplace announcements where access terms allow it.
5. Search-engine discovery as a supplemental path, followed by publisher verification.

This is a discovery plan, not blanket permission to scrape any named website.

## 8. Template and approval

Use templates/source-profile.template.yaml as the starting point. The logical schema is schemas/source-profile.schema.json.

A source profile must be reviewed before it is copied into an active local registry. Do not commit sensitive strategy URLs, cookies, access tokens, account details or secrets. An approved flag does not replace human review of source terms.

## 9. Change control

When a source changes structure, terms, access rules or rate limits:
- Pause the adapter if it may exceed the approved scope.
- Update the source profile and adapter version.
- Retain the previous extraction version in historical evidence.
- Re-run the source-specific fixture and quality checks.
- Record the change and reason.

Adding a source does not justify a new crawler framework or one service per website.
