# VALENTINA — Security and Permission Matrix

**Task ID:** VAL-P0-001
**Repository:** MiguelDuval/VALENTINA
**Branch:** `audit/valentina-phase0-feature-inventory-20261106`
**Base branch:** `chore/agent-collaboration-ci-20261106`
**Base SHA:** `a2d67f8901c949f4d1012634b573c1e08044817c`
**Huginn submodule:** `f01c1d84c08b1de972d8cc268da25602598adff8`
**Status:** Documentation only — no application code changed
**Evidence level:** SOURCE VERIFIED (based on FEATURE-INVENTORY.md)

---

## Critical Distinction: Current vs Desired

This matrix distinguishes **current observed permissions** (what the code actually enforces) from **desired target policy** (what should be enforced per project policy documents). A reader must not mistake desired future controls for currently enforced technical controls.

| Control | Current State | Target State | Gap |
|---------|--------------|--------------|-----|
| Source registry allowlist | Policy only — NOT technically enforced | Server-side enforcement via VALENTINA MCP adapter | IMPLEMENTATION GAP |
| Unattended fetch_url | Available to agents | Restricted to source-scoped adapter | POLICY GAP |
| `--dangerously-skip-permissions` | CLI flag exists | Not available in production builds | CONFIGURATION GAP |
| MCP approval gate | Default true (fail-closed) | Keep default true | None — already correct |
| Secret storage | OS keychain (primary) + `~/.huginn/secrets.json` plaintext JSON map with 0600 file permissions (fallback; not application-encrypted) | Keep both; no plaintext literal keys in config.json | None — already correct |
| Loopback binding | `127.0.0.1` default | Keep loopback-only | None — already correct |
| SSRF protection | `fetch_url` blocks private/internal IPs | Keep and extend to source adapter | None — already correct |

## Roles and Permissions

### 1. Owner / Human Operator

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | All local data via UI and API; config, secrets status, logs, sessions, artifacts | Same |
| Search/fetch access | Web search (Brave, requires key), fetch_url (any http/https URL) | Search via approved sources only; fetch via source-scoped adapter |
| File-system access | Read/write within sandbox root (~/.huginn, workspace) | Same |
| Shell access | bash tool available (requires permission prompt) | Same, but not for unattended agents |
| Git operations | git tools available (status, diff, log, commit, push, etc.) | Same, but not for unattended agents |
| API credentials | All 11 secret slots configurable via UI | Same |
| MCP capabilities | All MCP servers configurable; approval gate default true | Same |
| External mutations | Possible via connections tools (GitHub, Slack, etc.) | Require human approval for each |
| Data access | All local data | Same |
| Human approval requirements | Permission prompt per tool call (TUI/WS) | Same |
| Risks | Owner can inadvertently grant broad permissions | Mitigated by permission prompts and session allow-lists |
| Mitigations | Permission gate, audit log; `ResolveAPIKey` avoids including resolved secret values in its own error messages (application-wide log redaction NOT VERIFIED) | Same |



### 2. Interactive Assistant

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | All data accessible via agent tools | Same |
| Search/fetch access | web_search, fetch_url tools | Same (fetch_url is NOT technically limited to approved sources — prompt-level instruction only; source registry is policy-only, not technically enforced) |
| File-system access | Sandbox root only | Same |
| Shell access | bash tool (requires permission prompt per call) | Same |
| Git operations | git tools (requires permission prompt) | Same |
| API credentials | No direct access; resolved via keyring at request time | Same |
| MCP capabilities | MCP tools available (requires approval gate) | Same |
| External mutations | Possible via connections tools (requires approval) | Require human approval |
| Data access | All local data | Same |
| Human approval requirements | Permission prompt per tool call | Same |
| Risks | Prompt injection from fetched pages could override instructions | Source integrity skill mitigates |
| Mitigations | Permission gate, source integrity skill, prompt injection resistance | Same |

### 3. Unattended Scout

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | Read-only via web_search, fetch_url | Read-only via source-scoped MCP adapter |
| Search/fetch access | web_search, fetch_url (any URL) | Only approved source IDs via adapter; no arbitrary URLs |
| File-system access | NONE | NONE |
| Shell access | NONE | NONE |
| Git operations | NONE | NONE |
| API credentials | No direct access | No direct access |
| MCP capabilities | Read-only MCP tools only | Read-only source adapter tools only |
| External mutations | NONE | NONE |
| Data access | Public web data only | Approved source data only |
| Human approval requirements | Required for any mutation | Required for any mutation (read-only by design) |
| Risks | Unattended fetch to unapproved sources; SSRF; prompt injection | Mitigated by source adapter + allowlist |
| Mitigations | Starter agent templates define read-only tools; workflow is disabled/manual | Implement VALENTINA source-policy MCP adapter |


### 4. Verifier

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | Read-only via fetch_url (original publisher pages) | Read-only via source-scoped MCP adapter |
| Search/fetch access | fetch_url only (no web_search expansion) | fetch via source-scoped adapter only |
| File-system access | NONE | NONE |
| Shell access | NONE | NONE |
| Git operations | NONE | NONE |
| API credentials | No direct access | No direct access |
| MCP capabilities | Read-only fetch tools only | Read-only source adapter tools only |
| External mutations | NONE | NONE |
| Data access | Original publisher pages only | Approved source data only |
| Human approval requirements | Required for any mutation | Required for any mutation |
| Risks | Verification of compromised or misleading pages | Source integrity skill; original URL preservation |
| Mitigations | Starter agent: fetch_url only, no external tools, no memory | Implement source adapter |

### 5. Assessor

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | Read-only via provided facts (no browsing) | Same |
| Search/fetch access | NONE (no local tools) | NONE |
| File-system access | NONE | NONE |
| Shell access | NONE | NONE |
| Git operations | NONE | NONE |
| API credentials | No direct access | No direct access |
| MCP capabilities | NONE | NONE |
| External mutations | NONE | NONE |
| Data access | Source-verified facts only | Same |
| Human approval requirements | N/A (read-only assessment) | N/A |
| Risks | Assessment based on incomplete or incorrect facts | Fit assessment explains missing/uncertain facts |
| Mitigations | Starter agent: no tools, recommendation only | Same |

### 6. Source Adapter (future component)

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | NOT IMPLEMENTED | Approved source data only |
| Search/fetch access | NOT IMPLEMENTED | Only via source ID; enforced host/path allowlist |
| File-system access | NOT IMPLEMENTED | Evidence cache only (read/write in designated dir) |
| Shell access | NOT IMPLEMENTED | NONE |
| Git operations | NOT IMPLEMENTED | NONE |
| API credentials | NOT IMPLEMENTED | Credential reference only (no direct key access) |
| MCP capabilities | NOT IMPLEMENTED | Read-only fetch/search via scoped tools |
| External mutations | NOT IMPLEMENTED | NONE |
| Data access | NOT IMPLEMENTED | Approved source data + evidence records |
| Human approval requirements | NOT IMPLEMENTED | Required for any mutation |
| Risks | N/A (not implemented) | SSRF, redirect abuse, DNS rebinding, rate limit bypass |
| Mitigations | N/A | Fail-closed adapter; redirect validation; DNS/IP checks; rate limits |

### 7. Diagnostics / Maintenance

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | Logs, stats, health endpoint, session data | Same |
| Search/fetch access | NONE | NONE |
| File-system access | Read logs, config, status files | Same |
| Shell access | NONE | NONE |
| Git operations | NONE | NONE |
| API credentials | Secret status only (not values) | Same |
| MCP capabilities | MCP server health/status | Same |
| External mutations | NONE | NONE |
| Data access | Operational metadata only | Same |
| Human approval requirements | N/A (read-only diagnostics) | N/A |
| Risks | Logs could contain secrets if misconfigured | Credential redaction in logs |
| Mitigations | Log level control, secret redaction policy | Same |

### 8. CI / Build Workflow

| Attribute | Current Observed | Desired Target |
|-----------|-----------------|----------------|
| Read access | Repository source code (public) | Same |
| Search/fetch access | `npm ci` and Go module setup access package registries (dependency-fetch network only) | Same — not web research, scraping or external business actions |
| File-system access | Read repository; write to dist/ | Same |
| Shell access | Build commands only (go build, npm ci, npm test) | Same |
| Git operations | Checkout submodule with `persist-credentials: false`; no branch push or PR merge by CI | Same |
| API credentials | No `NVIDIA_NIM_API_KEY` or any `secrets.*` referenced — no AI provider API credential is used by CI | Same |
| MCP capabilities | NONE | NONE |
| External mutations | None — CI does not push branches or merge PRs | Same |
| Data access | No customer/lead data | No customer/lead data |
| Human approval requirements | Owner reviews PR | Same |
| Risks | Dependency-fetch network access to package registries | Mitigated by pinned action versions (tags, not commit SHAs) |
| Mitigations | `permissions: contents: read`; `persist-credentials: false`; actions use version tags (not commit-SHA pins); no secret in artifacts | Same |

## Capability Boundaries by Role

| Capability | Owner | Interactive Assistant | Unattended Scout | Verifier | Assessor | Source Adapter | Diagnostics | CI |
|------------|-------|----------------------|------------------|----------|----------|----------------|-------------|-----|
| Read local config | YES | YES | NO | NO | NO | NO | YES | NO |
| Read secrets | YES (via UI) | NO | NO | NO | NO | NO | NO | NO |
| Web search | YES | YES | YES (raw web_search tool; NOT technically limited to approved sources) | NO | NO | NO | NO | NO |
| fetch arbitrary URL | YES | YES | YES (raw fetch_url tool; NOT technically limited to approved sources) | YES (raw fetch_url tool) | NO | NO | NO | NO |
| fetch approved source | NO (via source-scoped adapter only — NOT IMPLEMENTED) | NO (via source-scoped adapter only — NOT IMPLEMENTED) | NO (source-scoped adapter capability NONE / NOT IMPLEMENTED) | NO (source-scoped adapter capability NONE / NOT IMPLEMENTED) | NO | NO (source-scoped adapter capability NONE / NOT IMPLEMENTED) | NO | NO |
| Shell (bash) | YES (with prompt) | YES (with prompt) | NO | NO | NO | NO | NO | Build only |
| File write | YES (with prompt) | YES (with prompt) | NO | NO | NO | Evidence cache only | NO | Build only |
| Git operations | YES (with prompt) | YES (with prompt) | NO | NO | NO | NO | NO | Repository checkout; pinned submodule checkout; no branch push; no PR merge |
| API credentials | YES (configure) | NO | NO | NO | NO | NO | NO | CI secret only |
| MCP servers | YES (configure) | YES (with approval) | NO | NO | NO | YES (scoped) | NO | NO |
| External mutations | YES (with approval) | YES (with approval) | NO | NO | NO | NO | NO | NO |
| Workflow execution | YES (manual) | YES (manual) | NO (disabled) | NO | NO | NO | NO | NO |
| Scheduler | YES (configure) | YES (view) | NO | NO | NO | NO | NO | NO |

### Scout / Verifier Tool Enforcement — Current vs Desired

**Current observed state (SOURCE VERIFIED):**

The VALENTINA-Scout agent template (`templates/huginn/agents/valentina-scout.json`) declares `local_tools: ["web_search", "fetch_url"]`. The VALENTINA-Verifier template (`templates/huginn/agents/valentina-verifier.json`) declares `local_tools: ["fetch_url"]`. These are **templates only** — they are not installed or activated by default. They must be copied to `~/.huginn/agents/` and enabled in the Huginn UI.

If provisioned with their current templates, Scout/Verifier would have the raw `web_search` and `fetch_url` tools. However:
- `fetch_url` can fetch arbitrary public HTTP(S) URLs subject only to upstream SSRF protections and bounded-fetch limits (10s timeout, 512KB max, HTML-to-Markdown conversion). It does **not** enforce the VALENTINA source registry's `allowed_hosts` or `allowed_path_prefixes`.
- `web_search` calls the Brave Search API and does **not** technically filter results to approved source registry entries.
- The approved-source scope in the Scout's system prompt is **prompt-level guidance only**, not a technical authorization boundary.

**Source allowlist enforcement is NOT IMPLEMENTED.** No VALENTINA source-policy MCP adapter exists yet.

**Desired future state (NOT IMPLEMENTED):**

Unattended agents receive only source-scoped adapter tools. The adapter enforces host/path/redirect/DNS/IP restrictions. This requires a separate implementation task and tested approval gates.

## Credential Boundaries

| Credential type | Storage location | Accessible by | Never stored in |
|----------------|-----------------|---------------|-----------------|
| NVIDIA NIM API key | OS keychain / secrets.json | Owner via UI; backend at request time | Git, logs, frontend bundle, process args |
| Brave Search API key | OS keychain / secrets.json | Owner via UI; backend at request time | Git, logs, frontend bundle |
| GitHub OAuth | OS keychain / secrets.json | Owner via UI | Git, logs |
| Google OAuth | OS keychain / secrets.json | Owner via UI | Git, logs |
| Slack OAuth | OS keychain / secrets.json | Owner via UI | Git, logs |
| Jira OAuth | OS keychain / secrets.json | Owner via UI | Git, logs |
| Bitbucket OAuth | OS keychain / secrets.json | Owner via UI | Git, logs |
| Server bearer token | Runtime memory only | Local clients via /api/v1/token | Logs, Git, config |
| HuginnCloud relay token | Runtime memory only | Local clients | Logs, Git, config |



## Data Boundaries

| Data type | Storage | Git-ignored? | Contains PII? | Retention |
|-----------|---------|-------------|---------------|-----------|
| Owner profile | `private/owner-profile.yaml` | Yes | Business info (not personal contact in v0.1) | Indefinite |
| Source registry | Local protected storage | Yes | Source URLs, terms | Indefinite |
| SQLite database (future) | Local protected storage | Yes | Business records | Indefinite |
| Huginn config | `~/.huginn/config.json` | Yes | API key references | Indefinite |
| Huginn secrets | OS keychain (primary); `~/.huginn/secrets.json` plaintext JSON map with 0600 file permissions (fallback, not application-encrypted) | Yes | Actual API keys/tokens | Indefinite |
| Sessions | SQLite (`~/.huginn/store/`) | Yes | Chat content | Indefinite |
| Logs | Local file | Yes | Operational metadata (no secrets) | Configurable |
| Notifications | Pebble/SQLite | Yes | Activity data | Configurable with pruning |
| Synthetic fixtures | `testdata/fixtures/synthetic/` | No (public Git) | No (fictional, example.com) | Indefinite |

## External Side Effects

| Operation | Current enforcement | Required approval | Roles permitted |
|-----------|-------------------|-------------------|-----------------|
| Outbound HTTP fetch | SSRF protection only (fetch_url) | Human approval for unattended | Owner, Interactive Assistant |
| Git commit/push | Permission prompt | Human approval | Owner, Interactive Assistant |
| File write | Permission prompt | Human approval | Owner, Interactive Assistant |
| Shell command | Permission prompt | Human approval | Owner, Interactive Assistant |
| OAuth connection | OAuth flow | Human approval | Owner |
| MCP tool call | Approval gate (default true) | Human approval (per server) | Owner, Interactive Assistant |
| Web search | API key required | None (read-only) | Owner, Interactive Assistant, Scout |
| Form submission | No technical block | Human approval | None (not in v0.1) |
| Account creation | No technical block | Human approval | None (not in v0.1) |

## Security Risks and Mitigations

| Risk | Severity | Current Mitigation | Gap | Required Action |
|------|----------|-------------------|-----|-----------------|
| Raw fetch_url bypasses source registry | High | None (policy only) | Source registry not technically enforced | Implement VALENTINA source-policy MCP adapter |
| Unattended agents with arbitrary fetch | High | Workflow disabled/manual | No technical enforcement | Keep disabled until adapter exists |
| Prompt injection from fetched pages | Medium | Source integrity skill | Skill is template only (not active) | Activate skill via templates/huginn/ |
| Secret in logs | Medium | Credential redaction policy | Policy only (not code-enforced) | Code review of log output |
| Credential in config.json (literal) | Low | IsLiteralAPIKey flag | Literal keys possible | Encourage keyring storage |
| Cloud/satellite connection | Medium | Disabled by default | Available in UI | Keep disabled in local product |
| `--dangerously-skip-permissions` | High | CLI flag only | Could be used in production | Remove from production builds |

## Verification Requirements

| Control | Verification method | Status |
|---------|-------------------|--------|
| Loopback binding | Config default `127.0.0.1` | SOURCE VERIFIED |
| SSRF protection | `fetch_url.go` `isBlockedIP()` | SOURCE VERIFIED |
| fetch_url limits | 10s timeout, 512KB max | SOURCE VERIFIED |
| MCP approval gate | Default `true` in config.go | SOURCE VERIFIED |
| Permission gate | `permissions.go` Gate struct | SOURCE VERIFIED |
| Credential redaction | `ResolveAPIKey` avoids resolved values in its own error messages | NOT VERIFIED for application-wide log redaction |
| Source registry enforcement | NOT IMPLEMENTED | IMPLEMENTATION GAP |
| Unattended fetch restriction | NOT IMPLEMENTED | POLICY GAP |
| Live NIM test | No credentials available | NOT RUN |
